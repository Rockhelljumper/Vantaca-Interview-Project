package domain

import "time"

const (
	TransferIntentRecorded = "INTENT_RECORDED"
	TransferPending        = "PENDING"
	TransferPosted         = "POSTED"
	TransferFailed         = "FAILED"
	TransferReturned       = "RETURNED"
	TransferUnknown        = "UNKNOWN"
)

type Account struct {
	ID            string
	Type          string
	LastFour      string
	Balance       Money
	Currency      string
	Status        string
	Version       int64
	FetchedAt     time.Time
	CheckedAt     time.Time
	LastSyncError string
}

func (a Account) DisplayName() string {
	name := "Account"
	switch a.Type {
	case "checking":
		name = "Checking"
	case "savings":
		name = "Savings"
	}
	return name + " ••••" + a.LastFour
}

type Transaction struct {
	ID                   string
	Amount               Money
	Currency             string
	Description          string
	MerchantCategoryCode string
	PostedAt             time.Time
}

type Transfer struct {
	ID                string
	TenantID          string
	RequestID         string
	FromAccountID     string
	ToAccountID       string
	FromDisplay       string
	ToDisplay         string
	Amount            Money
	Currency          string
	Status            string
	PartnerTransferID string
	LastErrorCategory string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type PartnerAccount struct {
	ID            string
	AccountNumber string
	RoutingNumber string
	Type          string
	Balance       Money
	Currency      string
	Status        string
}

type PartnerTransfer struct {
	ID        string
	Status    string
	Amount    Money
	CreatedAt time.Time
}

func CanTransitionTransfer(current string, next string) bool {
	if current == next {
		return true
	}
	switch current {
	case TransferIntentRecorded:
		return next == TransferPending || next == TransferFailed || next == TransferUnknown
	case TransferUnknown:
		return next == TransferPending || next == TransferPosted || next == TransferFailed || next == TransferReturned
	case TransferPending:
		return next == TransferPosted || next == TransferFailed
	case TransferPosted:
		return next == TransferReturned
	default:
		return false
	}
}
