package mockapi

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrAccountNotFound   = errors.New("account not found")
	ErrTransferNotFound  = errors.New("transfer not found")
	ErrInvalidTransition = errors.New("invalid transfer status transition")
)

const pageSize = 50

type Store struct {
	mu           sync.RWMutex
	clock        func() time.Time
	accounts     []Account
	transactions map[string][]Transaction
	transfers    []Transfer
	nextTxn      int64
	nextTransfer int64
	bootID       string
}

func NewStore(clock func() time.Time) *Store {
	if clock == nil {
		clock = time.Now
	}

	return &Store{
		clock: clock,
		accounts: []Account{
			{
				ID:            "acc_1029",
				AccountNumber: "000123454321",
				RoutingNumber: "021000021",
				Type:          "checking",
				Balance:       Money(482055),
				Currency:      "USD",
				Status:        "open",
			},
			{
				ID:            "acc_2042",
				AccountNumber: "000987656789",
				RoutingNumber: "021000021",
				Type:          "savings",
				Balance:       Money(1205000),
				Currency:      "USD",
				Status:        "open",
			},
			{
				ID:            "acc_3097",
				AccountNumber: "000555551111",
				RoutingNumber: "021000021",
				Type:          "checking",
				Balance:       Money(0),
				Currency:      "USD",
				Status:        "closed",
			},
		},
		transactions: map[string][]Transaction{
			"acc_1029": {
				{
					ID:          "txn_88213",
					Amount:      Money(-4217),
					Currency:    "USD",
					Description: "COFFEE HOUSE #42",
					PostedAt:    time.Date(2026, time.July, 21, 14, 3, 0, 0, time.UTC),
				},
				{
					ID:                   "txn_88174",
					Amount:               Money(250000),
					Currency:             "USD",
					Description:          "PAYROLL DEPOSIT",
					MerchantCategoryCode: "0000",
					PostedAt:             time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC),
				},
			},
			"acc_2042": {
				{
					ID:          "txn_77120",
					Amount:      Money(12500),
					Currency:    "USD",
					Description: "INTEREST CREDIT",
					PostedAt:    time.Date(2026, time.July, 1, 0, 5, 0, 0, time.UTC),
				},
			},
			"acc_3097": {},
		},
		nextTxn:      90001,
		nextTransfer: 55120,
		bootID:       newBootID(clock),
	}
}

func (s *Store) AddExternalTransaction(
	accountID string,
	amount Money,
	description string,
	merchantCategoryCode string,
) (Account, Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	accountIndex := -1
	for index := range s.accounts {
		if s.accounts[index].ID == accountID {
			accountIndex = index
			break
		}
	}
	if accountIndex < 0 || s.accounts[accountIndex].Status != "open" {
		return Account{}, Transaction{}, ErrAccountNotFound
	}

	transaction := Transaction{
		ID:                   fmt.Sprintf("txn_demo_%05d", s.nextTxn),
		Amount:               amount,
		Currency:             s.accounts[accountIndex].Currency,
		Description:          description,
		MerchantCategoryCode: merchantCategoryCode,
		PostedAt:             s.clock().UTC(),
	}
	s.nextTxn++
	s.accounts[accountIndex].Balance += amount
	s.transactions[accountID] = append([]Transaction{transaction}, s.transactions[accountID]...)

	return s.accounts[accountIndex], transaction, nil
}

func (s *Store) ListAccounts(page int) []Account {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return paginate(s.accounts, page)
}

func (s *Store) AccountByID(id string) (Account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, account := range s.accounts {
		if account.ID == id {
			return account, true
		}
	}

	return Account{}, false
}

func (s *Store) AccountByNumber(number string) (Account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, account := range s.accounts {
		if account.AccountNumber == number {
			return account, true
		}
	}

	return Account{}, false
}

func (s *Store) ListTransactions(accountID string, page int) ([]Transaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	transactions, ok := s.transactions[accountID]
	if !ok {
		return nil, ErrAccountNotFound
	}

	return paginate(transactions, page), nil
}

func (s *Store) CreateTransfer(request TransferRequest) Transfer {
	s.mu.Lock()
	defer s.mu.Unlock()

	transfer := Transfer{
		ID:        fmt.Sprintf("trf_%s_%d", s.bootID, s.nextTransfer),
		Status:    StatusPending,
		Amount:    request.Amount,
		CreatedAt: s.clock().UTC(),
	}
	s.nextTransfer++
	s.transfers = append(s.transfers, transfer)

	return transfer
}

func newBootID(clock func() time.Time) string {
	var random [6]byte
	if _, err := rand.Read(random[:]); err == nil {
		return hex.EncodeToString(random[:])
	}
	return fmt.Sprintf("%x", clock().UTC().UnixNano())
}

func (s *Store) ListTransfers(page int) []Transfer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return paginate(s.transfers, page)
}

func (s *Store) UpdateTransferStatus(id string, status string) (Transfer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index := range s.transfers {
		transfer := s.transfers[index]
		if transfer.ID != id {
			continue
		}
		if !validTransition(transfer.Status, status) {
			return Transfer{}, ErrInvalidTransition
		}

		s.transfers[index].Status = status
		return s.transfers[index], nil
	}

	return Transfer{}, ErrTransferNotFound
}

func validTransition(current string, next string) bool {
	if current == next {
		return true
	}

	switch current {
	case StatusPending:
		return next == StatusPosted || next == StatusFailed
	case StatusPosted:
		return next == StatusReturned
	default:
		return false
	}
}

func validStatus(status string) bool {
	switch status {
	case StatusPending, StatusPosted, StatusFailed, StatusReturned:
		return true
	default:
		return false
	}
}

func paginate[T any](records []T, page int) []T {
	start := (page - 1) * pageSize
	if start >= len(records) {
		return []T{}
	}

	end := start + pageSize
	if end > len(records) {
		end = len(records)
	}

	result := make([]T, end-start)
	copy(result, records[start:end])
	return result
}
