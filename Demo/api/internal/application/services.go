package application

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"vantaca-interview-project/Demo/api/internal/domain"
	"vantaca-interview-project/Demo/api/internal/northwind"
)

type Repository interface {
	UpsertAccount(context.Context, string, domain.Account) (bool, error)
	ListAccounts(context.Context, string) ([]domain.Account, error)
	GetAccount(context.Context, string, string) (domain.Account, error)
	ListTransactions(context.Context, string, string) ([]domain.Transaction, error)
	ReplaceTransactions(context.Context, string, string, []domain.Transaction, time.Time) (bool, int64, error)
	MarkAccountSyncFailure(context.Context, string, string, string, time.Time) error
	MarkOutboxPublished(context.Context, string, int64, string) error
	CreateTransferIntent(context.Context, domain.Transfer) (domain.Transfer, bool, error)
	UpdateTransferStatus(context.Context, string, string, string, string, string, string) (domain.Transfer, error)
	ListTransfers(context.Context, string) ([]domain.Transfer, error)
	GetTransfer(context.Context, string, string) (domain.Transfer, error)
	GetTransferByPartnerID(context.Context, string, string) (domain.Transfer, error)
	RecordWebhook(context.Context, string, string, string, string, time.Time) (string, bool, error)
	MarkWebhookProcessed(context.Context, string, string) error
}

type Partner interface {
	ListAccounts(context.Context, string) ([]domain.PartnerAccount, error)
	ListTransactions(context.Context, string, string) ([]domain.Transaction, error)
	CreateTransfer(context.Context, northwind.CreateTransferRequest, string) (domain.PartnerTransfer, error)
	ListTransfers(context.Context, string) ([]domain.PartnerTransfer, error)
	SimulateExternalActivity(context.Context, string) (northwind.ExternalActivity, error)
	AdvanceTransfer(context.Context, string, string, int) error
}

type SyncResult struct {
	AccountsSeen        int  `json:"accounts_seen"`
	AccountsChanged     int  `json:"accounts_changed"`
	TransactionSetsSeen int  `json:"transaction_sets_seen"`
	TransactionChanges  int  `json:"transaction_changes"`
	Partial             bool `json:"partial"`
}

type SyncService struct {
	repository Repository
	partner    Partner
	logger     *slog.Logger
}

func NewSyncService(repository Repository, partner Partner, logger *slog.Logger) *SyncService {
	return &SyncService{repository: repository, partner: partner, logger: logger}
}

func (s *SyncService) SyncAll(ctx context.Context, tenantID string, scenario string) (SyncResult, error) {
	if err := validateReadScenario(scenario); err != nil {
		return SyncResult{}, err
	}
	partnerAccounts, err := s.partner.ListAccounts(ctx, scenario)
	if err != nil {
		// Preserve the last known-good read model, but make the degraded
		// freshness visible for every locally known account when the
		// partner's account-list call fails before per-account sync begins.
		localAccounts, listErr := s.repository.ListAccounts(ctx, tenantID)
		if listErr == nil {
			failedAt := time.Now().UTC()
			category := northwind.Category(err)
			for _, account := range localAccounts {
				_ = s.repository.MarkAccountSyncFailure(ctx, tenantID, account.ID, category, failedAt)
			}
		}
		return SyncResult{}, fmt.Errorf("list Northwind accounts: %w", err)
	}

	now := time.Now().UTC()
	result := SyncResult{AccountsSeen: len(partnerAccounts)}
	var partialErrors []error
	for _, partnerAccount := range partnerAccounts {
		account := normalizeAccount(partnerAccount, now)
		changed, err := s.repository.UpsertAccount(ctx, tenantID, account)
		if err != nil {
			partialErrors = append(partialErrors, fmt.Errorf("upsert account %s: %w", partnerAccount.ID, err))
			continue
		}
		if changed {
			result.AccountsChanged++
			stored, readErr := s.repository.GetAccount(ctx, tenantID, partnerAccount.ID)
			if readErr == nil {
				if publishErr := s.repository.MarkOutboxPublished(ctx, partnerAccount.ID, stored.Version, "account.updated"); publishErr != nil {
					s.logger.Warn("account invalidation publish marker failed", "account_id", partnerAccount.ID, "error", publishErr)
				}
			}
		}

		transactions, err := s.partner.ListTransactions(ctx, partnerAccount.ID, scenario)
		if err != nil {
			_ = s.repository.MarkAccountSyncFailure(ctx, tenantID, partnerAccount.ID, northwind.Category(err), time.Now().UTC())
			partialErrors = append(partialErrors, fmt.Errorf("list transactions for %s: %w", partnerAccount.ID, err))
			continue
		}
		result.TransactionSetsSeen++
		transactionChanged, version, err := s.repository.ReplaceTransactions(
			ctx,
			tenantID,
			partnerAccount.ID,
			transactions,
			time.Now().UTC(),
		)
		if err != nil {
			partialErrors = append(partialErrors, fmt.Errorf("replace transactions for %s: %w", partnerAccount.ID, err))
			continue
		}
		if transactionChanged {
			result.TransactionChanges++
			if err := s.repository.MarkOutboxPublished(ctx, partnerAccount.ID, version, "recent_transactions.updated"); err != nil {
				s.logger.Warn("transaction invalidation publish marker failed", "account_id", partnerAccount.ID, "version", version, "error", err)
			}
		}
	}
	if len(partialErrors) > 0 {
		result.Partial = true
		return result, errors.Join(partialErrors...)
	}
	return result, nil
}

func (s *SyncService) RefreshTransactions(ctx context.Context, tenantID string, accountID string, scenario string) (bool, int64, error) {
	if err := validateReadScenario(scenario); err != nil {
		return false, 0, err
	}
	if _, err := s.repository.GetAccount(ctx, tenantID, accountID); err != nil {
		return false, 0, fmt.Errorf("find account: %w", err)
	}

	transactions, err := s.partner.ListTransactions(ctx, accountID, scenario)
	if err != nil {
		category := northwind.Category(err)
		_ = s.repository.MarkAccountSyncFailure(ctx, tenantID, accountID, category, time.Now().UTC())
		return false, 0, fmt.Errorf("refresh Northwind transactions: %w", err)
	}

	changed, version, err := s.repository.ReplaceTransactions(
		ctx,
		tenantID,
		accountID,
		transactions,
		time.Now().UTC(),
	)
	if err != nil {
		return false, 0, err
	}
	if changed {
		if err := s.repository.MarkOutboxPublished(ctx, accountID, version, "recent_transactions.updated"); err != nil {
			return true, version, fmt.Errorf("mark post-commit invalidation published: %w", err)
		}
	}
	return changed, version, nil
}

func (s *SyncService) SimulateExternalActivity(ctx context.Context, accountID string) (northwind.ExternalActivity, error) {
	return s.partner.SimulateExternalActivity(ctx, accountID)
}

func normalizeAccount(account domain.PartnerAccount, now time.Time) domain.Account {
	lastFour := account.AccountNumber
	if len(lastFour) > 4 {
		lastFour = lastFour[len(lastFour)-4:]
	}
	return domain.Account{
		ID:        account.ID,
		Type:      account.Type,
		LastFour:  lastFour,
		Balance:   account.Balance,
		Currency:  account.Currency,
		Status:    account.Status,
		FetchedAt: now,
		CheckedAt: now,
	}
}

type RefreshCoordinator struct {
	service *SyncService
	logger  *slog.Logger
	mu      sync.RWMutex
	running map[string]bool
	wg      sync.WaitGroup
}

func NewRefreshCoordinator(service *SyncService, logger *slog.Logger) *RefreshCoordinator {
	return &RefreshCoordinator{
		service: service,
		logger:  logger,
		running: make(map[string]bool),
	}
}

func (c *RefreshCoordinator) Start(tenantID string, accountID string, scenario string) bool {
	key := tenantID + ":" + accountID
	c.mu.Lock()
	if c.running[key] {
		c.mu.Unlock()
		return false
	}
	c.running[key] = true
	c.wg.Add(1)
	c.mu.Unlock()

	go func() {
		defer c.wg.Done()
		defer func() {
			c.mu.Lock()
			delete(c.running, key)
			c.mu.Unlock()
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		changed, version, err := c.service.RefreshTransactions(ctx, tenantID, accountID, scenario)
		if err != nil {
			c.logger.Warn("asynchronous transaction refresh failed", "account_id", accountID, "category", northwind.Category(err))
			return
		}
		c.logger.Info("asynchronous transaction refresh complete", "account_id", accountID, "changed", changed, "version", version)
	}()
	return true
}

func (c *RefreshCoordinator) IsRunning(tenantID string, accountID string) bool {
	key := tenantID + ":" + accountID
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running[key]
}

func (c *RefreshCoordinator) Wait() {
	c.wg.Wait()
}

type SubmitTransferInput struct {
	RequestID     string
	FromAccountID string
	ToAccountID   string
	Amount        string
	Currency      string
	Scenario      string
}

var ErrValidation = errors.New("validation error")

func validationError(message string) error {
	return fmt.Errorf("%s: %w", message, ErrValidation)
}

type TransferService struct {
	repository Repository
	partner    Partner
	logger     *slog.Logger
}

func NewTransferService(repository Repository, partner Partner, logger *slog.Logger) *TransferService {
	return &TransferService{repository: repository, partner: partner, logger: logger}
}

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{7,99}$`)

func (s *TransferService) Submit(ctx context.Context, tenantID string, input SubmitTransferInput) (domain.Transfer, error) {
	if !requestIDPattern.MatchString(input.RequestID) {
		return domain.Transfer{}, validationError("request_id must contain 8 through 100 safe characters")
	}
	if input.FromAccountID == "" || input.ToAccountID == "" || input.FromAccountID == input.ToAccountID {
		return domain.Transfer{}, validationError("source and destination accounts must differ")
	}
	if input.Currency != "USD" {
		return domain.Transfer{}, validationError("the demo supports USD transfers only")
	}
	if err := validateTransferScenario(input.Scenario); err != nil {
		return domain.Transfer{}, fmt.Errorf("%v: %w", err, ErrValidation)
	}
	amount, err := domain.ParseMoney(input.Amount)
	if err != nil || amount <= 0 {
		return domain.Transfer{}, validationError("amount must be greater than zero with at most two decimals")
	}

	fromAccount, err := s.repository.GetAccount(ctx, tenantID, input.FromAccountID)
	if err != nil {
		return domain.Transfer{}, validationError("source account is unavailable")
	}
	toAccount, err := s.repository.GetAccount(ctx, tenantID, input.ToAccountID)
	if err != nil {
		return domain.Transfer{}, validationError("destination account is unavailable")
	}
	if fromAccount.Status != "open" || toAccount.Status != "open" {
		return domain.Transfer{}, validationError("both accounts must be open")
	}

	internalID, err := newID()
	if err != nil {
		return domain.Transfer{}, err
	}
	now := time.Now().UTC()
	intent := domain.Transfer{
		ID:            internalID,
		TenantID:      tenantID,
		RequestID:     input.RequestID,
		FromAccountID: input.FromAccountID,
		ToAccountID:   input.ToAccountID,
		FromDisplay:   fromAccount.DisplayName(),
		ToDisplay:     toAccount.DisplayName(),
		Amount:        amount,
		Currency:      input.Currency,
		Status:        domain.TransferIntentRecorded,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	intent, created, err := s.repository.CreateTransferIntent(ctx, intent)
	if err != nil {
		return domain.Transfer{}, fmt.Errorf("persist transfer intent: %w", err)
	}
	if !created {
		return intent, nil
	}

	partnerAccounts, err := s.partner.ListAccounts(ctx, "")
	if err != nil {
		return s.failBeforeSubmission(ctx, tenantID, intent, "partner_account_lookup")
	}
	fromPartner, fromOK := findPartnerAccount(partnerAccounts, input.FromAccountID)
	toPartner, toOK := findPartnerAccount(partnerAccounts, input.ToAccountID)
	if !fromOK || !toOK || fromPartner.Status != "open" || toPartner.Status != "open" {
		return s.failBeforeSubmission(ctx, tenantID, intent, "partner_account_unavailable")
	}

	partnerTransfer, err := s.partner.CreateTransfer(ctx, northwind.CreateTransferRequest{
		FromAccountNumber: fromPartner.AccountNumber,
		ToAccountNumber:   toPartner.AccountNumber,
		RoutingNumber:     toPartner.RoutingNumber,
		Amount:            amount,
		Currency:          input.Currency,
	}, input.Scenario)
	if err != nil {
		status := domain.TransferFailed
		category := northwind.Category(err)
		if northwind.IsAmbiguous(err) {
			status = domain.TransferUnknown
		}
		updated, updateErr := s.repository.UpdateTransferStatus(
			ctx,
			tenantID,
			intent.ID,
			status,
			"",
			category,
			"northwind_submission",
		)
		if updateErr != nil {
			return domain.Transfer{}, errors.Join(err, updateErr)
		}
		return updated, nil
	}

	updated, err := s.repository.UpdateTransferStatus(
		ctx,
		tenantID,
		intent.ID,
		partnerTransfer.Status,
		partnerTransfer.ID,
		"",
		"northwind_submission",
	)
	if err != nil {
		// Northwind accepted the monetary request, so a local persistence
		// failure cannot be reported as a definitive failure. Preserve the
		// durable intent as UNKNOWN when SQL is still reachable and never
		// invite a client or worker to submit it again.
		unknown, unknownErr := s.repository.UpdateTransferStatus(
			ctx,
			tenantID,
			intent.ID,
			domain.TransferUnknown,
			"",
			"post_accept_persistence",
			"northwind_submission",
		)
		if unknownErr != nil {
			return domain.Transfer{}, errors.Join(err, unknownErr)
		}
		return unknown, nil
	}
	return updated, nil
}

func (s *TransferService) failBeforeSubmission(
	ctx context.Context,
	tenantID string,
	intent domain.Transfer,
	category string,
) (domain.Transfer, error) {
	updated, err := s.repository.UpdateTransferStatus(
		ctx,
		tenantID,
		intent.ID,
		domain.TransferFailed,
		"",
		category,
		"pre_submission_validation",
	)
	if err != nil {
		return domain.Transfer{}, err
	}
	return updated, nil
}

func (s *TransferService) ReconcilePartnerTransfer(ctx context.Context, tenantID string, partnerID string) (domain.Transfer, error) {
	local, err := s.repository.GetTransferByPartnerID(ctx, tenantID, partnerID)
	if err != nil {
		return domain.Transfer{}, err
	}
	partnerTransfers, err := s.partner.ListTransfers(ctx, "")
	if err != nil {
		return domain.Transfer{}, fmt.Errorf("list Northwind transfers: %w", err)
	}
	partnerTransfer, ok := findPartnerTransfer(partnerTransfers, partnerID)
	if !ok {
		return domain.Transfer{}, errors.New("exact partner transfer was not found")
	}
	if partnerTransfer.Amount != local.Amount {
		return domain.Transfer{}, errors.New("exact partner id returned a conflicting amount")
	}
	return s.repository.UpdateTransferStatus(
		ctx,
		tenantID,
		local.ID,
		partnerTransfer.Status,
		partnerID,
		"",
		"partner_reconciliation",
	)
}

func (s *TransferService) ReconcileAllKnown(ctx context.Context, tenantID string) (int, error) {
	locals, err := s.repository.ListTransfers(ctx, tenantID)
	if err != nil {
		return 0, err
	}
	partnerTransfers, err := s.partner.ListTransfers(ctx, "")
	if err != nil {
		return 0, err
	}
	partnerByID := make(map[string]domain.PartnerTransfer, len(partnerTransfers))
	for _, transfer := range partnerTransfers {
		partnerByID[transfer.ID] = transfer
	}

	updated := 0
	var reconcileErrors []error
	for _, local := range locals {
		if local.PartnerTransferID == "" {
			continue
		}
		partnerTransfer, ok := partnerByID[local.PartnerTransferID]
		if !ok {
			reconcileErrors = append(reconcileErrors, fmt.Errorf("partner transfer %s not found", local.PartnerTransferID))
			continue
		}
		if partnerTransfer.Amount != local.Amount {
			reconcileErrors = append(reconcileErrors, fmt.Errorf("partner transfer %s amount mismatch", local.PartnerTransferID))
			continue
		}
		if partnerTransfer.Status == local.Status {
			continue
		}
		if _, err := s.repository.UpdateTransferStatus(
			ctx,
			tenantID,
			local.ID,
			partnerTransfer.Status,
			local.PartnerTransferID,
			"",
			"scheduled_reconciliation",
		); err != nil {
			reconcileErrors = append(reconcileErrors, err)
			continue
		}
		updated++
	}
	return updated, errors.Join(reconcileErrors...)
}

func (s *TransferService) AdvanceDemo(
	ctx context.Context,
	tenantID string,
	internalTransferID string,
	status string,
	deliveries int,
) (domain.Transfer, error) {
	if deliveries < 1 || deliveries > 5 {
		return domain.Transfer{}, validationError("deliveries must be between 1 and 5")
	}
	switch status {
	case domain.TransferPosted, domain.TransferFailed, domain.TransferReturned:
	default:
		return domain.Transfer{}, validationError("demo status must be POSTED, FAILED, or RETURNED")
	}
	local, err := s.repository.GetTransfer(ctx, tenantID, internalTransferID)
	if err != nil {
		return domain.Transfer{}, err
	}
	if local.PartnerTransferID == "" {
		return domain.Transfer{}, validationError("ambiguous transfer has no exact partner id and cannot be advanced safely")
	}
	if err := s.partner.AdvanceTransfer(ctx, local.PartnerTransferID, status, deliveries); err != nil {
		return domain.Transfer{}, err
	}
	return s.ReconcilePartnerTransfer(ctx, tenantID, local.PartnerTransferID)
}

func findPartnerAccount(accounts []domain.PartnerAccount, id string) (domain.PartnerAccount, bool) {
	for _, account := range accounts {
		if account.ID == id {
			return account, true
		}
	}
	return domain.PartnerAccount{}, false
}

func findPartnerTransfer(transfers []domain.PartnerTransfer, id string) (domain.PartnerTransfer, bool) {
	for _, transfer := range transfers {
		if transfer.ID == id {
			return transfer, true
		}
	}
	return domain.PartnerTransfer{}, false
}

func validateReadScenario(scenario string) error {
	switch scenario {
	case "", "429", "500", "503", "latency":
		return nil
	default:
		return errors.New("unsupported read failure scenario")
	}
}

func validateTransferScenario(scenario string) error {
	switch scenario {
	case "", "500", "503", "latency", "post-commit-timeout":
		return nil
	default:
		return errors.New("unsupported transfer failure scenario")
	}
}

func newID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate identifier: %w", err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		value[0:4],
		value[4:6],
		value[6:8],
		value[8:10],
		value[10:16],
	), nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows")
}
