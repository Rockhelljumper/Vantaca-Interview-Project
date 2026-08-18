package application

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"vantaca-interview-project/Demo/api/internal/domain"
	"vantaca-interview-project/Demo/api/internal/northwind"
)

func TestSubmitAmbiguousTransferDoesNotRetry(t *testing.T) {
	t.Parallel()

	repository := newMemoryRepository()
	repository.accounts = []domain.Account{
		{ID: "acc_from", Type: "checking", LastFour: "1234", Status: "open"},
		{ID: "acc_to", Type: "savings", LastFour: "5678", Status: "open"},
	}
	partner := &fakePartner{
		accounts: []domain.PartnerAccount{
			{ID: "acc_from", AccountNumber: "000000001234", RoutingNumber: "021000021", Status: "open"},
			{ID: "acc_to", AccountNumber: "000000005678", RoutingNumber: "021000021", Status: "open"},
		},
		createErr: &northwind.PartnerError{Kind: northwind.ErrorAmbiguous},
	}
	service := NewTransferService(repository, partner, discardLogger())

	transfer, err := service.Submit(context.Background(), "tenant_demo", SubmitTransferInput{
		RequestID:     "request-demo-001",
		FromAccountID: "acc_from",
		ToAccountID:   "acc_to",
		Amount:        "25.00",
		Currency:      "USD",
	})
	if err != nil {
		t.Fatal(err)
	}
	if transfer.Status != domain.TransferUnknown {
		t.Fatalf("status = %s, want UNKNOWN", transfer.Status)
	}
	if partner.createCalls != 1 {
		t.Fatalf("partner create calls = %d, want 1", partner.createCalls)
	}

	repeated, err := service.Submit(context.Background(), "tenant_demo", SubmitTransferInput{
		RequestID:     "request-demo-001",
		FromAccountID: "acc_from",
		ToAccountID:   "acc_to",
		Amount:        "25.00",
		Currency:      "USD",
	})
	if err != nil {
		t.Fatal(err)
	}
	if repeated.ID != transfer.ID {
		t.Fatalf("repeated transfer id = %s, want %s", repeated.ID, transfer.ID)
	}
	if partner.createCalls != 1 {
		t.Fatalf("partner create calls after repeat = %d, want 1", partner.createCalls)
	}
}

func TestSubmitPartnerAcceptedPersistenceFailureBecomesUnknown(t *testing.T) {
	t.Parallel()

	repository := newMemoryRepository()
	repository.accounts = []domain.Account{
		{ID: "acc_from", Type: "checking", LastFour: "1234", Status: "open"},
		{ID: "acc_to", Type: "savings", LastFour: "5678", Status: "open"},
	}
	repository.partnerUpdateErr = errors.New("synthetic partner-id persistence failure")
	partner := &fakePartner{
		accounts: []domain.PartnerAccount{
			{ID: "acc_from", AccountNumber: "000000001234", RoutingNumber: "021000021", Status: "open"},
			{ID: "acc_to", AccountNumber: "000000005678", RoutingNumber: "021000021", Status: "open"},
		},
		created: domain.PartnerTransfer{ID: "trf_partner_1", Status: domain.TransferPending},
	}
	service := NewTransferService(repository, partner, discardLogger())
	input := SubmitTransferInput{
		RequestID:     "request-demo-002",
		FromAccountID: "acc_from",
		ToAccountID:   "acc_to",
		Amount:        "25.00",
		Currency:      "USD",
	}

	transfer, err := service.Submit(context.Background(), "tenant_demo", input)
	if err != nil {
		t.Fatal(err)
	}
	if transfer.Status != domain.TransferUnknown || transfer.LastErrorCategory != "post_accept_persistence" {
		t.Fatalf("transfer = %+v, want UNKNOWN post_accept_persistence", transfer)
	}

	repeated, err := service.Submit(context.Background(), "tenant_demo", input)
	if err != nil {
		t.Fatal(err)
	}
	if repeated.ID != transfer.ID || partner.createCalls != 1 {
		t.Fatalf("repeat = %+v, partner calls = %d; want same transfer and one call", repeated, partner.createCalls)
	}
}

type fakePartner struct {
	accounts    []domain.PartnerAccount
	created     domain.PartnerTransfer
	createErr   error
	createCalls int
}

func (p *fakePartner) ListAccounts(context.Context, string) ([]domain.PartnerAccount, error) {
	return p.accounts, nil
}

func (p *fakePartner) ListTransactions(context.Context, string, string) ([]domain.Transaction, error) {
	return nil, nil
}

func (p *fakePartner) CreateTransfer(context.Context, northwind.CreateTransferRequest, string) (domain.PartnerTransfer, error) {
	p.createCalls++
	return p.created, p.createErr
}

func (p *fakePartner) ListTransfers(context.Context, string) ([]domain.PartnerTransfer, error) {
	return nil, nil
}

func (p *fakePartner) SimulateExternalActivity(context.Context, string) (northwind.ExternalActivity, error) {
	return northwind.ExternalActivity{}, nil
}

func (p *fakePartner) AdvanceTransfer(context.Context, string, string, int) error {
	return nil
}

type memoryRepository struct {
	mu               sync.Mutex
	accounts         []domain.Account
	transfers        []domain.Transfer
	partnerUpdateErr error
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{}
}

func (r *memoryRepository) UpsertAccount(context.Context, string, domain.Account) (bool, error) {
	return false, nil
}

func (r *memoryRepository) ListAccounts(context.Context, string) ([]domain.Account, error) {
	return r.accounts, nil
}

func (r *memoryRepository) GetAccount(_ context.Context, _ string, id string) (domain.Account, error) {
	for _, account := range r.accounts {
		if account.ID == id {
			return account, nil
		}
	}
	return domain.Account{}, errors.New("not found")
}

func (r *memoryRepository) ListTransactions(context.Context, string, string) ([]domain.Transaction, error) {
	return nil, nil
}

func (r *memoryRepository) ReplaceTransactions(context.Context, string, string, []domain.Transaction, time.Time) (bool, int64, error) {
	return false, 1, nil
}

func (r *memoryRepository) MarkAccountSyncFailure(context.Context, string, string, string, time.Time) error {
	return nil
}

func (r *memoryRepository) MarkOutboxPublished(context.Context, string, int64, string) error {
	return nil
}

func (r *memoryRepository) CreateTransferIntent(_ context.Context, transfer domain.Transfer) (domain.Transfer, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.transfers {
		if existing.TenantID == transfer.TenantID && existing.RequestID == transfer.RequestID {
			return existing, false, nil
		}
	}
	r.transfers = append(r.transfers, transfer)
	return transfer, true, nil
}

func (r *memoryRepository) UpdateTransferStatus(
	_ context.Context,
	tenantID string,
	internalID string,
	status string,
	partnerID string,
	category string,
	_ string,
) (domain.Transfer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if partnerID != "" && r.partnerUpdateErr != nil {
		return domain.Transfer{}, r.partnerUpdateErr
	}
	for index := range r.transfers {
		if r.transfers[index].TenantID == tenantID && r.transfers[index].ID == internalID {
			r.transfers[index].Status = status
			r.transfers[index].PartnerTransferID = partnerID
			r.transfers[index].LastErrorCategory = category
			return r.transfers[index], nil
		}
	}
	return domain.Transfer{}, errors.New("not found")
}

func (r *memoryRepository) ListTransfers(context.Context, string) ([]domain.Transfer, error) {
	return r.transfers, nil
}

func (r *memoryRepository) GetTransfer(context.Context, string, string) (domain.Transfer, error) {
	return domain.Transfer{}, errors.New("not implemented")
}

func (r *memoryRepository) GetTransferByPartnerID(context.Context, string, string) (domain.Transfer, error) {
	return domain.Transfer{}, errors.New("not implemented")
}

func (r *memoryRepository) RecordWebhook(context.Context, string, string, string, string, time.Time) (string, bool, error) {
	return "", false, nil
}

func (r *memoryRepository) MarkWebhookProcessed(context.Context, string, string) error {
	return nil
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
