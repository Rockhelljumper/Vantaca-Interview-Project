package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"vantaca-interview-project/Demo/api/internal/domain"
)

type Repository struct {
	db     *sql.DB
	linkID string
}

func NewRepository(db *sql.DB, linkID string) *Repository {
	return &Repository{db: db, linkID: linkID}
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) UpsertAccount(ctx context.Context, tenantID string, account domain.Account) (bool, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return false, fmt.Errorf("begin account upsert: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var (
		internalID int64
		current    domain.Account
	)
	err = tx.QueryRowContext(ctx, `
SELECT
    a.linked_account_id,
    a.account_type,
    a.last_four,
    a.balance_minor,
    a.currency,
    a.status,
    a.data_version
FROM dbo.linked_accounts AS a WITH (UPDLOCK, HOLDLOCK)
INNER JOIN dbo.northwind_links AS l ON l.link_id = a.link_id
WHERE l.tenant_id = @tenant_id
  AND l.link_id = @link_id
  AND a.northwind_account_id = @account_id`,
		sql.Named("tenant_id", tenantID),
		sql.Named("link_id", r.linkID),
		sql.Named("account_id", account.ID),
	).Scan(
		&internalID,
		&current.Type,
		&current.LastFour,
		&current.Balance,
		&current.Currency,
		&current.Status,
		&current.Version,
	)

	changed := false
	switch {
	case errors.Is(err, sql.ErrNoRows):
		result := tx.QueryRowContext(ctx, `
INSERT INTO dbo.linked_accounts (
    link_id,
    northwind_account_id,
    account_type,
    last_four,
    balance_minor,
    currency,
    status,
    fetched_at,
    checked_at
)
OUTPUT INSERTED.linked_account_id
VALUES (
    @link_id,
    @account_id,
    @account_type,
    @last_four,
    @balance_minor,
    @currency,
    @status,
    @fetched_at,
    @checked_at
)`,
			sql.Named("link_id", r.linkID),
			sql.Named("account_id", account.ID),
			sql.Named("account_type", account.Type),
			sql.Named("last_four", account.LastFour),
			sql.Named("balance_minor", int64(account.Balance)),
			sql.Named("currency", account.Currency),
			sql.Named("status", account.Status),
			sql.Named("fetched_at", account.FetchedAt),
			sql.Named("checked_at", account.CheckedAt),
		)
		if err := result.Scan(&internalID); err != nil {
			return false, fmt.Errorf("insert account: %w", err)
		}
		account.Version = 1
		changed = true
	case err != nil:
		return false, fmt.Errorf("load account for upsert: %w", err)
	default:
		changed = current.Type != account.Type ||
			current.LastFour != account.LastFour ||
			current.Balance != account.Balance ||
			current.Currency != account.Currency ||
			current.Status != account.Status

		versionIncrement := 0
		if changed {
			versionIncrement = 1
		}
		if _, err := tx.ExecContext(ctx, `
UPDATE dbo.linked_accounts
SET
    account_type = @account_type,
    last_four = @last_four,
    balance_minor = @balance_minor,
    currency = @currency,
    status = @status,
    data_version = data_version + @version_increment,
    fetched_at = @fetched_at,
    checked_at = @checked_at,
    last_sync_error = NULL,
    updated_at = SYSUTCDATETIME()
WHERE linked_account_id = @internal_id`,
			sql.Named("account_type", account.Type),
			sql.Named("last_four", account.LastFour),
			sql.Named("balance_minor", int64(account.Balance)),
			sql.Named("currency", account.Currency),
			sql.Named("status", account.Status),
			sql.Named("version_increment", versionIncrement),
			sql.Named("fetched_at", account.FetchedAt),
			sql.Named("checked_at", account.CheckedAt),
			sql.Named("internal_id", internalID),
		); err != nil {
			return false, fmt.Errorf("update account: %w", err)
		}
		account.Version = current.Version + int64(versionIncrement)
	}

	if changed {
		eventID, err := newID()
		if err != nil {
			return false, err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO dbo.read_model_outbox (
    event_id,
    linked_account_id,
    aggregate_version,
    event_type
)
VALUES (@event_id, @account_id, @version, 'account.updated')`,
			sql.Named("event_id", eventID),
			sql.Named("account_id", internalID),
			sql.Named("version", account.Version),
		); err != nil {
			return false, fmt.Errorf("insert account outbox event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit account upsert: %w", err)
	}
	return changed, nil
}

func (r *Repository) ListAccounts(ctx context.Context, tenantID string) ([]domain.Account, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT
    northwind_account_id,
    account_type,
    last_four,
    balance_minor,
    currency,
    status,
    data_version,
    fetched_at,
    checked_at,
    COALESCE(last_sync_error, '')
FROM dbo.vw_customer_linked_accounts
WHERE tenant_id = @tenant_id
ORDER BY account_type, northwind_account_id`,
		sql.Named("tenant_id", tenantID),
	)
	if err != nil {
		return nil, fmt.Errorf("query accounts: %w", err)
	}
	defer rows.Close()

	accounts := make([]domain.Account, 0)
	for rows.Next() {
		var account domain.Account
		if err := rows.Scan(
			&account.ID,
			&account.Type,
			&account.LastFour,
			&account.Balance,
			&account.Currency,
			&account.Status,
			&account.Version,
			&account.FetchedAt,
			&account.CheckedAt,
			&account.LastSyncError,
		); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accounts: %w", err)
	}
	return accounts, nil
}

func (r *Repository) GetAccount(ctx context.Context, tenantID string, accountID string) (domain.Account, error) {
	var account domain.Account
	err := r.db.QueryRowContext(ctx, `
SELECT
    northwind_account_id,
    account_type,
    last_four,
    balance_minor,
    currency,
    status,
    data_version,
    fetched_at,
    checked_at,
    COALESCE(last_sync_error, '')
FROM dbo.vw_customer_linked_accounts
WHERE tenant_id = @tenant_id
  AND northwind_account_id = @account_id`,
		sql.Named("tenant_id", tenantID),
		sql.Named("account_id", accountID),
	).Scan(
		&account.ID,
		&account.Type,
		&account.LastFour,
		&account.Balance,
		&account.Currency,
		&account.Status,
		&account.Version,
		&account.FetchedAt,
		&account.CheckedAt,
		&account.LastSyncError,
	)
	if err != nil {
		return domain.Account{}, fmt.Errorf("get account: %w", err)
	}
	return account, nil
}

func (r *Repository) ListTransactions(ctx context.Context, tenantID string, accountID string) ([]domain.Transaction, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT
    northwind_transaction_id,
    amount_minor,
    currency,
    description,
    COALESCE(CONVERT(VARCHAR(4), merchant_category_code), ''),
    posted_at
FROM dbo.vw_customer_recent_transactions
WHERE tenant_id = @tenant_id
  AND northwind_account_id = @account_id
ORDER BY posted_at DESC, northwind_transaction_id DESC`,
		sql.Named("tenant_id", tenantID),
		sql.Named("account_id", accountID),
	)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]domain.Transaction, 0)
	for rows.Next() {
		var transaction domain.Transaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.Amount,
			&transaction.Currency,
			&transaction.Description,
			&transaction.MerchantCategoryCode,
			&transaction.PostedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transactions: %w", err)
	}
	return transactions, nil
}

func (r *Repository) ReplaceTransactions(
	ctx context.Context,
	tenantID string,
	accountID string,
	transactions []domain.Transaction,
	checkedAt time.Time,
) (bool, int64, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return false, 0, fmt.Errorf("begin transaction refresh: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var internalID int64
	var version int64
	if err := tx.QueryRowContext(ctx, `
SELECT a.linked_account_id, a.data_version
FROM dbo.linked_accounts AS a WITH (UPDLOCK, HOLDLOCK)
INNER JOIN dbo.northwind_links AS l ON l.link_id = a.link_id
WHERE l.tenant_id = @tenant_id
  AND a.northwind_account_id = @account_id`,
		sql.Named("tenant_id", tenantID),
		sql.Named("account_id", accountID),
	).Scan(&internalID, &version); err != nil {
		return false, 0, fmt.Errorf("lock account for transaction refresh: %w", err)
	}

	current, err := listTransactionsTx(ctx, tx, internalID)
	if err != nil {
		return false, 0, err
	}
	if transactionsEqual(current, transactions) {
		if _, err := tx.ExecContext(ctx, `
UPDATE dbo.linked_accounts
SET checked_at = @checked_at,
    fetched_at = @checked_at,
    last_sync_error = NULL,
    updated_at = SYSUTCDATETIME()
WHERE linked_account_id = @internal_id`,
			sql.Named("checked_at", checkedAt),
			sql.Named("internal_id", internalID),
		); err != nil {
			return false, 0, fmt.Errorf("mark matching transactions checked: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return false, 0, fmt.Errorf("commit matching transaction refresh: %w", err)
		}
		return false, version, nil
	}

	if _, err := tx.ExecContext(ctx,
		"DELETE FROM dbo.account_transactions WHERE linked_account_id = @internal_id",
		sql.Named("internal_id", internalID),
	); err != nil {
		return false, 0, fmt.Errorf("replace transactions delete: %w", err)
	}
	for _, transaction := range transactions {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO dbo.account_transactions (
    linked_account_id,
    northwind_transaction_id,
    amount_minor,
    currency,
    description,
    merchant_category_code,
    posted_at
)
VALUES (
    @internal_id,
    @transaction_id,
    @amount_minor,
    @currency,
    @description,
    @mcc,
    @posted_at
)`,
			sql.Named("internal_id", internalID),
			sql.Named("transaction_id", transaction.ID),
			sql.Named("amount_minor", int64(transaction.Amount)),
			sql.Named("currency", transaction.Currency),
			sql.Named("description", transaction.Description),
			sql.Named("mcc", nullableString(transaction.MerchantCategoryCode)),
			sql.Named("posted_at", transaction.PostedAt),
		); err != nil {
			return false, 0, fmt.Errorf("replace transactions insert: %w", err)
		}
	}

	if err := tx.QueryRowContext(ctx, `
UPDATE dbo.linked_accounts
SET data_version = data_version + 1,
    checked_at = @checked_at,
    fetched_at = @checked_at,
    last_sync_error = NULL,
    updated_at = SYSUTCDATETIME()
OUTPUT INSERTED.data_version
WHERE linked_account_id = @internal_id`,
		sql.Named("checked_at", checkedAt),
		sql.Named("internal_id", internalID),
	).Scan(&version); err != nil {
		return false, 0, fmt.Errorf("advance transaction version: %w", err)
	}

	eventID, err := newID()
	if err != nil {
		return false, 0, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO dbo.read_model_outbox (
    event_id,
    linked_account_id,
    aggregate_version,
    event_type
)
VALUES (@event_id, @internal_id, @version, 'recent_transactions.updated')`,
		sql.Named("event_id", eventID),
		sql.Named("internal_id", internalID),
		sql.Named("version", version),
	); err != nil {
		return false, 0, fmt.Errorf("insert transaction outbox event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, 0, fmt.Errorf("commit transaction replacement: %w", err)
	}
	return true, version, nil
}

func (r *Repository) MarkAccountSyncFailure(ctx context.Context, tenantID string, accountID string, category string, checkedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE a
SET a.checked_at = @checked_at,
    a.last_sync_error = @category,
    a.updated_at = SYSUTCDATETIME()
FROM dbo.linked_accounts AS a
INNER JOIN dbo.northwind_links AS l ON l.link_id = a.link_id
WHERE l.tenant_id = @tenant_id
  AND a.northwind_account_id = @account_id`,
		sql.Named("checked_at", checkedAt),
		sql.Named("category", category),
		sql.Named("tenant_id", tenantID),
		sql.Named("account_id", accountID),
	)
	if err != nil {
		return fmt.Errorf("mark account sync failure: %w", err)
	}
	return nil
}

func (r *Repository) MarkOutboxPublished(ctx context.Context, accountID string, version int64, eventType string) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE o
SET published_at = SYSUTCDATETIME(),
    publish_attempts = publish_attempts + 1
FROM dbo.read_model_outbox AS o
INNER JOIN dbo.linked_accounts AS a ON a.linked_account_id = o.linked_account_id
WHERE a.northwind_account_id = @account_id
  AND o.aggregate_version = @version
  AND o.event_type = @event_type
  AND o.published_at IS NULL`,
		sql.Named("account_id", accountID),
		sql.Named("version", version),
		sql.Named("event_type", eventType),
	)
	if err != nil {
		return fmt.Errorf("mark outbox published: %w", err)
	}
	return nil
}

func (r *Repository) CreateTransferIntent(ctx context.Context, transfer domain.Transfer) (domain.Transfer, bool, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return domain.Transfer{}, false, fmt.Errorf("begin transfer intent: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	existing, err := getTransferTx(ctx, tx, transfer.TenantID, "", transfer.RequestID)
	switch {
	case err == nil:
		if err := tx.Commit(); err != nil {
			return domain.Transfer{}, false, fmt.Errorf("commit existing transfer lookup: %w", err)
		}
		return existing, false, nil
	case !errors.Is(err, sql.ErrNoRows):
		return domain.Transfer{}, false, fmt.Errorf("find existing transfer request: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO dbo.transfers (
    internal_transfer_id,
    tenant_id,
    request_id,
    from_account_id,
    to_account_id,
    from_display,
    to_display,
    amount_minor,
    currency,
    status,
    created_at,
    updated_at
)
VALUES (
    @id,
    @tenant_id,
    @request_id,
    @from_account_id,
    @to_account_id,
    @from_display,
    @to_display,
    @amount_minor,
    @currency,
    @status,
    @created_at,
    @updated_at
)`,
		sql.Named("id", transfer.ID),
		sql.Named("tenant_id", transfer.TenantID),
		sql.Named("request_id", transfer.RequestID),
		sql.Named("from_account_id", transfer.FromAccountID),
		sql.Named("to_account_id", transfer.ToAccountID),
		sql.Named("from_display", transfer.FromDisplay),
		sql.Named("to_display", transfer.ToDisplay),
		sql.Named("amount_minor", int64(transfer.Amount)),
		sql.Named("currency", transfer.Currency),
		sql.Named("status", transfer.Status),
		sql.Named("created_at", transfer.CreatedAt),
		sql.Named("updated_at", transfer.UpdatedAt),
	); err != nil {
		return domain.Transfer{}, false, fmt.Errorf("insert transfer intent: %w", err)
	}
	if err := insertTransferHistory(ctx, tx, transfer.ID, transfer.Status, "customer_request", "intent durable"); err != nil {
		return domain.Transfer{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Transfer{}, false, fmt.Errorf("commit transfer intent: %w", err)
	}
	return transfer, true, nil
}

func (r *Repository) UpdateTransferStatus(
	ctx context.Context,
	tenantID string,
	internalID string,
	nextStatus string,
	partnerTransferID string,
	errorCategory string,
	source string,
) (domain.Transfer, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return domain.Transfer{}, fmt.Errorf("begin transfer status update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	current, err := getTransferTx(ctx, tx, tenantID, internalID, "")
	if err != nil {
		return domain.Transfer{}, fmt.Errorf("load transfer for status update: %w", err)
	}
	if !domain.CanTransitionTransfer(current.Status, nextStatus) {
		return domain.Transfer{}, fmt.Errorf("transition %s to %s: %w", current.Status, nextStatus, ErrInvalidTransition)
	}
	if current.Status == nextStatus &&
		(partnerTransferID == "" || partnerTransferID == current.PartnerTransferID) &&
		errorCategory == current.LastErrorCategory {
		if err := tx.Commit(); err != nil {
			return domain.Transfer{}, fmt.Errorf("commit transfer no-op: %w", err)
		}
		return current, nil
	}

	if partnerTransferID != "" {
		current.PartnerTransferID = partnerTransferID
	}
	current.Status = nextStatus
	current.LastErrorCategory = errorCategory
	current.UpdatedAt = time.Now().UTC()

	if _, err := tx.ExecContext(ctx, `
UPDATE dbo.transfers
SET status = @status,
    partner_transfer_id = @partner_transfer_id,
    last_error_category = @error_category,
    updated_at = @updated_at
WHERE tenant_id = @tenant_id
  AND internal_transfer_id = @id`,
		sql.Named("status", current.Status),
		sql.Named("partner_transfer_id", nullableString(current.PartnerTransferID)),
		sql.Named("error_category", nullableString(current.LastErrorCategory)),
		sql.Named("updated_at", current.UpdatedAt),
		sql.Named("tenant_id", tenantID),
		sql.Named("id", internalID),
	); err != nil {
		return domain.Transfer{}, fmt.Errorf("update transfer status: %w", err)
	}
	if err := insertTransferHistory(ctx, tx, internalID, nextStatus, source, errorCategory); err != nil {
		return domain.Transfer{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Transfer{}, fmt.Errorf("commit transfer status update: %w", err)
	}
	return current, nil
}

func (r *Repository) ListTransfers(ctx context.Context, tenantID string) ([]domain.Transfer, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT
    CONVERT(NVARCHAR(36), internal_transfer_id),
    tenant_id,
    request_id,
    from_account_id,
    to_account_id,
    from_display,
    to_display,
    amount_minor,
    currency,
    status,
    COALESCE(partner_transfer_id, ''),
    COALESCE(last_error_category, ''),
    created_at,
    updated_at
FROM dbo.transfers
WHERE tenant_id = @tenant_id
ORDER BY created_at DESC`,
		sql.Named("tenant_id", tenantID),
	)
	if err != nil {
		return nil, fmt.Errorf("query transfers: %w", err)
	}
	defer rows.Close()

	transfers := make([]domain.Transfer, 0)
	for rows.Next() {
		transfer, err := scanTransfer(rows)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transfers: %w", err)
	}
	return transfers, nil
}

func (r *Repository) GetTransfer(ctx context.Context, tenantID string, internalID string) (domain.Transfer, error) {
	row := r.db.QueryRowContext(ctx, transferSelect+`
WHERE tenant_id = @tenant_id
  AND internal_transfer_id = @id`,
		sql.Named("tenant_id", tenantID),
		sql.Named("id", internalID),
	)
	transfer, err := scanTransfer(row)
	if err != nil {
		return domain.Transfer{}, fmt.Errorf("get transfer: %w", err)
	}
	return transfer, nil
}

func (r *Repository) GetTransferByPartnerID(ctx context.Context, tenantID string, partnerID string) (domain.Transfer, error) {
	row := r.db.QueryRowContext(ctx, transferSelect+`
WHERE tenant_id = @tenant_id
  AND partner_transfer_id = @partner_id`,
		sql.Named("tenant_id", tenantID),
		sql.Named("partner_id", partnerID),
	)
	transfer, err := scanTransfer(row)
	if err != nil {
		return domain.Transfer{}, fmt.Errorf("get transfer by partner id: %w", err)
	}
	return transfer, nil
}

func (r *Repository) RecordWebhook(
	ctx context.Context,
	dedupeKey string,
	partnerTransferID string,
	reportedStatus string,
	payloadHash string,
	receivedAt time.Time,
) (string, bool, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return "", false, fmt.Errorf("begin webhook receipt: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var existing string
	err = tx.QueryRowContext(ctx, `
SELECT CONVERT(NVARCHAR(36), webhook_receipt_id)
FROM dbo.webhook_inbox WITH (UPDLOCK, HOLDLOCK)
WHERE dedupe_key = @dedupe_key`,
		sql.Named("dedupe_key", dedupeKey),
	).Scan(&existing)
	switch {
	case err == nil:
		if err := tx.Commit(); err != nil {
			return "", false, fmt.Errorf("commit duplicate webhook lookup: %w", err)
		}
		return existing, false, nil
	case !errors.Is(err, sql.ErrNoRows):
		return "", false, fmt.Errorf("find webhook receipt: %w", err)
	}

	receiptID, err := newID()
	if err != nil {
		return "", false, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO dbo.webhook_inbox (
    webhook_receipt_id,
    dedupe_key,
    partner_transfer_id,
    reported_status,
    payload_sha256,
    received_at
)
VALUES (
    @receipt_id,
    @dedupe_key,
    @partner_id,
    @status,
    @payload_hash,
    @received_at
)`,
		sql.Named("receipt_id", receiptID),
		sql.Named("dedupe_key", dedupeKey),
		sql.Named("partner_id", partnerTransferID),
		sql.Named("status", reportedStatus),
		sql.Named("payload_hash", payloadHash),
		sql.Named("received_at", receivedAt),
	); err != nil {
		return "", false, fmt.Errorf("insert webhook receipt: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return "", false, fmt.Errorf("commit webhook receipt: %w", err)
	}
	return receiptID, true, nil
}

func (r *Repository) MarkWebhookProcessed(ctx context.Context, receiptID string, outcome string) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE dbo.webhook_inbox
SET processed_at = SYSUTCDATETIME(),
    processing_outcome = @outcome
WHERE webhook_receipt_id = @receipt_id`,
		sql.Named("outcome", outcome),
		sql.Named("receipt_id", receiptID),
	)
	if err != nil {
		return fmt.Errorf("mark webhook processed: %w", err)
	}
	return nil
}

var ErrInvalidTransition = errors.New("invalid transfer transition")

const transferSelect = `
SELECT
    CONVERT(NVARCHAR(36), internal_transfer_id),
    tenant_id,
    request_id,
    from_account_id,
    to_account_id,
    from_display,
    to_display,
    amount_minor,
    currency,
    status,
    COALESCE(partner_transfer_id, ''),
    COALESCE(last_error_category, ''),
    created_at,
    updated_at
FROM dbo.transfers
`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTransfer(row rowScanner) (domain.Transfer, error) {
	var transfer domain.Transfer
	if err := row.Scan(
		&transfer.ID,
		&transfer.TenantID,
		&transfer.RequestID,
		&transfer.FromAccountID,
		&transfer.ToAccountID,
		&transfer.FromDisplay,
		&transfer.ToDisplay,
		&transfer.Amount,
		&transfer.Currency,
		&transfer.Status,
		&transfer.PartnerTransferID,
		&transfer.LastErrorCategory,
		&transfer.CreatedAt,
		&transfer.UpdatedAt,
	); err != nil {
		return domain.Transfer{}, err
	}
	return transfer, nil
}

func getTransferTx(ctx context.Context, tx *sql.Tx, tenantID string, internalID string, requestID string) (domain.Transfer, error) {
	query := transferSelect + " WITH (UPDLOCK, HOLDLOCK) WHERE tenant_id = @tenant_id"
	args := []any{sql.Named("tenant_id", tenantID)}
	if internalID != "" {
		query += " AND internal_transfer_id = @id"
		args = append(args, sql.Named("id", internalID))
	} else {
		query += " AND request_id = @request_id"
		args = append(args, sql.Named("request_id", requestID))
	}
	return scanTransfer(tx.QueryRowContext(ctx, query, args...))
}

func insertTransferHistory(ctx context.Context, tx *sql.Tx, transferID string, status string, source string, detail string) error {
	_, err := tx.ExecContext(ctx, `
INSERT INTO dbo.transfer_status_history (
    internal_transfer_id,
    status,
    source,
    safe_detail
)
VALUES (@transfer_id, @status, @source, @detail)`,
		sql.Named("transfer_id", transferID),
		sql.Named("status", status),
		sql.Named("source", source),
		sql.Named("detail", nullableString(detail)),
	)
	if err != nil {
		return fmt.Errorf("insert transfer history: %w", err)
	}
	return nil
}

func listTransactionsTx(ctx context.Context, tx *sql.Tx, internalID int64) ([]domain.Transaction, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT
    northwind_transaction_id,
    amount_minor,
    currency,
    description,
    COALESCE(CONVERT(VARCHAR(4), merchant_category_code), ''),
    posted_at
FROM dbo.account_transactions WITH (UPDLOCK, HOLDLOCK)
WHERE linked_account_id = @internal_id`,
		sql.Named("internal_id", internalID),
	)
	if err != nil {
		return nil, fmt.Errorf("load current transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]domain.Transaction, 0)
	for rows.Next() {
		var transaction domain.Transaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.Amount,
			&transaction.Currency,
			&transaction.Description,
			&transaction.MerchantCategoryCode,
			&transaction.PostedAt,
		); err != nil {
			return nil, fmt.Errorf("scan current transaction: %w", err)
		}
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate current transactions: %w", err)
	}
	return transactions, nil
}

func transactionsEqual(left []domain.Transaction, right []domain.Transaction) bool {
	if len(left) != len(right) {
		return false
	}
	leftCopy := append([]domain.Transaction(nil), left...)
	rightCopy := append([]domain.Transaction(nil), right...)
	sort.Slice(leftCopy, func(i, j int) bool { return leftCopy[i].ID < leftCopy[j].ID })
	sort.Slice(rightCopy, func(i, j int) bool { return rightCopy[i].ID < rightCopy[j].ID })
	for index := range leftCopy {
		l := leftCopy[index]
		r := rightCopy[index]
		if l.ID != r.ID ||
			l.Amount != r.Amount ||
			l.Currency != r.Currency ||
			l.Description != r.Description ||
			l.MerchantCategoryCode != r.MerchantCategoryCode ||
			// The SQL Server driver returns DATETIMEOFFSET(7) at 100 ns
			// precision. Compare at that stored boundary so a partner timestamp
			// with finer Go precision does not create a false invalidation.
			!l.PostedAt.Truncate(100*time.Nanosecond).Equal(r.PostedAt.Truncate(100*time.Nanosecond)) {
			return false
		}
	}
	return true
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
