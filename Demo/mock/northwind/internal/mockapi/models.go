package mockapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	StatusPending  = "PENDING"
	StatusPosted   = "POSTED"
	StatusFailed   = "FAILED"
	StatusReturned = "RETURNED"
)

// Money stores USD cents so the mock never relies on binary floating-point
// arithmetic for balances, transactions, or transfer validation.
type Money int64

func (m Money) MarshalJSON() ([]byte, error) {
	negative := m < 0
	var magnitude uint64
	if negative {
		magnitude = uint64(-(m + 1))
		magnitude++
	} else {
		magnitude = uint64(m)
	}

	value := fmt.Sprintf("%d.%02d", magnitude/100, magnitude%100)
	if negative {
		value = "-" + value
	}

	return []byte(value), nil
}

func (m *Money) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return errors.New("money cannot be null")
	}

	value, err := parseCents(string(data))
	if err != nil {
		return err
	}

	*m = Money(value)
	return nil
}

func parseCents(raw string) (int64, error) {
	if raw == "" || strings.ContainsAny(raw, "eE") {
		return 0, fmt.Errorf("amount %q must be a plain decimal with at most two fractional digits", raw)
	}

	negative := strings.HasPrefix(raw, "-")
	if negative {
		raw = strings.TrimPrefix(raw, "-")
	}
	if raw == "" {
		return 0, errors.New("amount is missing digits")
	}

	parts := strings.Split(raw, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, fmt.Errorf("amount %q is invalid", raw)
	}

	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
		if fraction == "" || len(fraction) > 2 {
			return 0, fmt.Errorf("amount %q must have one or two fractional digits", raw)
		}
	}

	whole, err := strconv.ParseUint(parts[0], 10, 63)
	if err != nil {
		return 0, fmt.Errorf("amount %q has an invalid whole-number component", raw)
	}

	for len(fraction) < 2 {
		fraction += "0"
	}
	var fractional uint64
	if fraction != "" {
		fractional, err = strconv.ParseUint(fraction, 10, 8)
		if err != nil {
			return 0, fmt.Errorf("amount %q has an invalid fractional component", raw)
		}
	}

	if whole > (math.MaxInt64-fractional)/100 {
		return 0, fmt.Errorf("amount %q is too large", raw)
	}

	cents := int64(whole*100 + fractional)
	if negative {
		cents = -cents
	}

	return cents, nil
}

type Account struct {
	ID            string `json:"id"`
	AccountNumber string `json:"account_number"`
	RoutingNumber string `json:"routing_number"`
	Type          string `json:"type"`
	Balance       Money  `json:"balance"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
}

type Transaction struct {
	ID                   string    `json:"id"`
	Amount               Money     `json:"amount"`
	Currency             string    `json:"currency"`
	Description          string    `json:"description"`
	MerchantCategoryCode string    `json:"merchant_category_code,omitempty"`
	PostedAt             time.Time `json:"posted_at"`
}

type TransferRequest struct {
	FromAccountNumber string `json:"from_account_number"`
	ToAccountNumber   string `json:"to_account_number"`
	RoutingNumber     string `json:"routing_number"`
	Amount            Money  `json:"amount"`
	Currency          string `json:"currency"`
}

type Transfer struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Amount    Money     `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type WebhookEvent struct {
	Event      string `json:"event"`
	TransferID string `json:"transfer_id"`
	Status     string `json:"status"`
}

type StatusUpdateRequest struct {
	Status     string `json:"status"`
	Deliveries int    `json:"deliveries,omitempty"`
}

type StatusUpdateResponse struct {
	Transfer          Transfer `json:"transfer"`
	WebhookDeliveries int      `json:"webhook_deliveries"`
}

type ExternalTransactionRequest struct {
	Amount               Money  `json:"amount"`
	Description          string `json:"description"`
	MerchantCategoryCode string `json:"merchant_category_code,omitempty"`
}

type ExternalTransactionResponse struct {
	Account     Account     `json:"account"`
	Transaction Transaction `json:"transaction"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func decodeJSON(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request must contain one JSON object")
		}
		return err
	}

	return nil
}
