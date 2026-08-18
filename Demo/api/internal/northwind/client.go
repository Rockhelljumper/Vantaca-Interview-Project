package northwind

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"vantaca-interview-project/Demo/api/internal/domain"
)

const (
	scenarioHeader = "X-Northwind-Mock-Scenario"
	maxPages       = 100
	pageSize       = 50
	maxBodyBytes   = 2 << 20
)

type ErrorKind string

const (
	ErrorValidation ErrorKind = "validation"
	ErrorAuth       ErrorKind = "authentication"
	ErrorNotFound   ErrorKind = "not_found"
	ErrorThrottled  ErrorKind = "throttled"
	ErrorTransient  ErrorKind = "transient"
	ErrorTimeout    ErrorKind = "timeout"
	ErrorContract   ErrorKind = "contract"
	ErrorAmbiguous  ErrorKind = "ambiguous_outcome"
)

type PartnerError struct {
	Kind       ErrorKind
	StatusCode int
	Code       string
	RetryAfter time.Duration
	Err        error
}

func (e *PartnerError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("northwind %s error (HTTP %d)", e.Kind, e.StatusCode)
	}
	return fmt.Sprintf("northwind %s error", e.Kind)
}

func (e *PartnerError) Unwrap() error {
	return e.Err
}

func Category(err error) string {
	var partnerErr *PartnerError
	if errors.As(err, &partnerErr) {
		return string(partnerErr.Kind)
	}
	return "internal"
}

func IsAmbiguous(err error) bool {
	var partnerErr *PartnerError
	return errors.As(err, &partnerErr) && partnerErr.Kind == ErrorAmbiguous
}

type Client struct {
	baseURL    string
	controlURL string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL string, apiKey string, timeout time.Duration) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" || apiKey == "" {
		return nil, errors.New("northwind base URL and API key are required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("northwind base URL must be an absolute HTTP URL")
	}
	if timeout <= 0 {
		return nil, errors.New("northwind timeout must be positive")
	}

	controlURL := baseURL
	if strings.HasSuffix(controlURL, "/v1") {
		controlURL = strings.TrimSuffix(controlURL, "/v1")
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 20
	transport.MaxIdleConnsPerHost = 10
	transport.IdleConnTimeout = 30 * time.Second

	return &Client{
		baseURL:    baseURL,
		controlURL: controlURL,
		apiKey:     apiKey,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}, nil
}

func (c *Client) ListAccounts(ctx context.Context, scenario string) ([]domain.PartnerAccount, error) {
	records, err := listPages[accountDTO](ctx, c, "accounts", scenario)
	if err != nil {
		return nil, err
	}
	accounts := make([]domain.PartnerAccount, 0, len(records))
	for _, record := range records {
		account, err := record.domain()
		if err != nil {
			return nil, &PartnerError{Kind: ErrorContract, Err: err}
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (c *Client) ListTransactions(ctx context.Context, accountID string, scenario string) ([]domain.Transaction, error) {
	if strings.TrimSpace(accountID) == "" {
		return nil, &PartnerError{Kind: ErrorValidation, Err: errors.New("account id is required")}
	}
	path := "accounts/" + url.PathEscape(accountID) + "/transactions"
	records, err := listPages[transactionDTO](ctx, c, path, scenario)
	if err != nil {
		return nil, err
	}
	transactions := make([]domain.Transaction, 0, len(records))
	for _, record := range records {
		transaction, err := record.domain()
		if err != nil {
			return nil, &PartnerError{Kind: ErrorContract, Err: err}
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

type CreateTransferRequest struct {
	FromAccountNumber string
	ToAccountNumber   string
	RoutingNumber     string
	Amount            domain.Money
	Currency          string
}

func (c *Client) CreateTransfer(
	ctx context.Context,
	request CreateTransferRequest,
	scenario string,
) (domain.PartnerTransfer, error) {
	body, err := json.Marshal(transferRequestDTO{
		FromAccountNumber: request.FromAccountNumber,
		ToAccountNumber:   request.ToAccountNumber,
		RoutingNumber:     request.RoutingNumber,
		Amount:            request.Amount,
		Currency:          request.Currency,
	})
	if err != nil {
		return domain.PartnerTransfer{}, fmt.Errorf("encode transfer request: %w", err)
	}

	endpoint, err := c.endpoint(c.baseURL, "transfers", 0)
	if err != nil {
		return domain.PartnerTransfer{}, err
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return domain.PartnerTransfer{}, fmt.Errorf("create Northwind transfer request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	if scenario != "" {
		httpRequest.Header.Set(scenarioHeader, scenario)
	}

	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		kind := ErrorAmbiguous
		if isTimeout(err) {
			return domain.PartnerTransfer{}, &PartnerError{Kind: kind, Err: err}
		}
		return domain.PartnerTransfer{}, &PartnerError{Kind: kind, Err: err}
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		partnerErr := responseError(response)
		if partnerErr.Kind == ErrorTransient || partnerErr.Kind == ErrorThrottled {
			partnerErr.Kind = ErrorAmbiguous
		}
		return domain.PartnerTransfer{}, partnerErr
	}

	var record transferDTO
	if err := decodeResponse(response, &record); err != nil {
		return domain.PartnerTransfer{}, &PartnerError{Kind: ErrorAmbiguous, Err: err}
	}
	transfer, err := record.domain()
	if err != nil {
		return domain.PartnerTransfer{}, &PartnerError{Kind: ErrorContract, Err: err}
	}
	return transfer, nil
}

func (c *Client) ListTransfers(ctx context.Context, scenario string) ([]domain.PartnerTransfer, error) {
	records, err := listPages[transferDTO](ctx, c, "transfers", scenario)
	if err != nil {
		return nil, err
	}
	transfers := make([]domain.PartnerTransfer, 0, len(records))
	for _, record := range records {
		transfer, err := record.domain()
		if err != nil {
			return nil, &PartnerError{Kind: ErrorContract, Err: err}
		}
		transfers = append(transfers, transfer)
	}
	return transfers, nil
}

type ExternalActivity struct {
	Account     domain.PartnerAccount
	Transaction domain.Transaction
}

func (c *Client) SimulateExternalActivity(ctx context.Context, accountID string) (ExternalActivity, error) {
	body, err := json.Marshal(externalTransactionRequestDTO{
		Amount:               domain.Money(12550),
		Description:          "EXTERNAL DEMO DEPOSIT",
		MerchantCategoryCode: "0000",
	})
	if err != nil {
		return ExternalActivity{}, fmt.Errorf("encode demo activity: %w", err)
	}
	path := "__mock/accounts/" + url.PathEscape(accountID) + "/transactions"
	endpoint, err := c.endpoint(c.controlURL, path, 0)
	if err != nil {
		return ExternalActivity{}, err
	}
	var response externalTransactionResponseDTO
	if err := c.doControl(ctx, endpoint, body, &response); err != nil {
		return ExternalActivity{}, err
	}
	account, err := response.Account.domain()
	if err != nil {
		return ExternalActivity{}, err
	}
	transaction, err := response.Transaction.domain()
	if err != nil {
		return ExternalActivity{}, err
	}
	return ExternalActivity{Account: account, Transaction: transaction}, nil
}

func (c *Client) AdvanceTransfer(ctx context.Context, partnerTransferID string, status string, deliveries int) error {
	body, err := json.Marshal(map[string]any{
		"status":     status,
		"deliveries": deliveries,
	})
	if err != nil {
		return fmt.Errorf("encode demo transfer transition: %w", err)
	}
	path := "__mock/transfers/" + url.PathEscape(partnerTransferID) + "/status"
	endpoint, err := c.endpoint(c.controlURL, path, 0)
	if err != nil {
		return err
	}
	var response any
	return c.doControl(ctx, endpoint, body, &response)
}

func (c *Client) doControl(ctx context.Context, endpoint string, body []byte, destination any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create mock control request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("execute mock control request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return responseError(response)
	}
	if err := decodeResponse(response, destination); err != nil {
		return fmt.Errorf("decode mock control response: %w", err)
	}
	return nil
}

func listPages[T any](ctx context.Context, client *Client, path string, scenario string) ([]T, error) {
	records := make([]T, 0)
	for page := 1; page <= maxPages; page++ {
		var pageRecords []T
		if err := client.getJSON(ctx, path, page, scenario, &pageRecords); err != nil {
			return nil, err
		}
		records = append(records, pageRecords...)
		if len(pageRecords) < pageSize {
			return records, nil
		}
	}
	return nil, &PartnerError{Kind: ErrorContract, Err: errors.New("pagination exceeded safety limit")}
}

func (c *Client) getJSON(ctx context.Context, path string, page int, scenario string, destination any) error {
	endpoint, err := c.endpoint(c.baseURL, path, page)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return fmt.Errorf("create Northwind read request: %w", err)
		}
		if scenario != "" {
			request.Header.Set(scenarioHeader, scenario)
		}

		response, err := c.httpClient.Do(request)
		if err != nil {
			kind := ErrorTransient
			if isTimeout(err) {
				kind = ErrorTimeout
			}
			lastErr = &PartnerError{Kind: kind, Err: err}
		} else {
			if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
				decodeErr := decodeResponse(response, destination)
				closeErr := response.Body.Close()
				if decodeErr != nil {
					return &PartnerError{Kind: ErrorContract, Err: decodeErr}
				}
				if closeErr != nil {
					return fmt.Errorf("close Northwind response: %w", closeErr)
				}
				return nil
			}
			lastErr = responseError(response)
			_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
			_ = response.Body.Close()
		}

		var partnerErr *PartnerError
		if !errors.As(lastErr, &partnerErr) ||
			(partnerErr.Kind != ErrorThrottled && partnerErr.Kind != ErrorTransient && partnerErr.Kind != ErrorTimeout) ||
			attempt == 3 {
			return lastErr
		}

		delay := time.Duration(attempt) * 100 * time.Millisecond
		if partnerErr.RetryAfter > 0 {
			delay = partnerErr.RetryAfter
			if delay > 2*time.Second {
				delay = 2 * time.Second
			}
		}
		if err := wait(ctx, delay); err != nil {
			return errors.Join(lastErr, err)
		}
	}
	return lastErr
}

func (c *Client) endpoint(baseURL string, path string, page int) (string, error) {
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/"))
	if err != nil {
		return "", fmt.Errorf("build Northwind URL: %w", err)
	}
	query := parsed.Query()
	query.Set("api_key", c.apiKey)
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func decodeResponse(response *http.Response, destination any) error {
	reader := io.LimitReader(response.Body, maxBodyBytes)
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("decode HTTP %d response: %w", response.StatusCode, err)
	}
	return nil
}

func responseError(response *http.Response) *PartnerError {
	var body struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(io.LimitReader(response.Body, 4096)).Decode(&body)

	partnerErr := &PartnerError{
		StatusCode: response.StatusCode,
		Code:       body.Error,
		Kind:       ErrorContract,
	}
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		partnerErr.Kind = ErrorValidation
	case http.StatusUnauthorized, http.StatusForbidden:
		partnerErr.Kind = ErrorAuth
	case http.StatusNotFound:
		partnerErr.Kind = ErrorNotFound
	case http.StatusTooManyRequests:
		partnerErr.Kind = ErrorThrottled
		if seconds, err := strconv.Atoi(response.Header.Get("Retry-After")); err == nil && seconds > 0 {
			partnerErr.RetryAfter = time.Duration(seconds) * time.Second
		}
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		partnerErr.Kind = ErrorTransient
	}
	return partnerErr
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func wait(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type accountDTO struct {
	ID            string      `json:"id"`
	AccountNumber string      `json:"account_number"`
	RoutingNumber string      `json:"routing_number"`
	Type          string      `json:"type"`
	Balance       json.Number `json:"balance"`
	Currency      string      `json:"currency"`
	Status        string      `json:"status"`
}

func (a accountDTO) domain() (domain.PartnerAccount, error) {
	if a.ID == "" || len(a.AccountNumber) < 4 || a.RoutingNumber == "" {
		return domain.PartnerAccount{}, errors.New("account identity fields are missing")
	}
	balance, err := domain.MoneyFromJSONNumber(a.Balance)
	if err != nil {
		return domain.PartnerAccount{}, fmt.Errorf("parse account balance: %w", err)
	}
	if a.Currency != "USD" {
		return domain.PartnerAccount{}, fmt.Errorf("unsupported account currency %q", a.Currency)
	}
	if a.Type != "checking" && a.Type != "savings" {
		return domain.PartnerAccount{}, fmt.Errorf("unknown account type %q", a.Type)
	}
	if a.Status != "open" && a.Status != "closed" {
		return domain.PartnerAccount{}, fmt.Errorf("unknown account status %q", a.Status)
	}
	return domain.PartnerAccount{
		ID:            a.ID,
		AccountNumber: a.AccountNumber,
		RoutingNumber: a.RoutingNumber,
		Type:          a.Type,
		Balance:       balance,
		Currency:      a.Currency,
		Status:        a.Status,
	}, nil
}

type transactionDTO struct {
	ID                   string      `json:"id"`
	Amount               json.Number `json:"amount"`
	Currency             string      `json:"currency"`
	Description          string      `json:"description"`
	MerchantCategoryCode string      `json:"merchant_category_code,omitempty"`
	PostedAt             time.Time   `json:"posted_at"`
}

func (t transactionDTO) domain() (domain.Transaction, error) {
	if t.ID == "" || t.Description == "" || t.PostedAt.IsZero() {
		return domain.Transaction{}, errors.New("transaction required fields are missing")
	}
	amount, err := domain.MoneyFromJSONNumber(t.Amount)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("parse transaction amount: %w", err)
	}
	if t.Currency != "USD" {
		return domain.Transaction{}, fmt.Errorf("unsupported transaction currency %q", t.Currency)
	}
	return domain.Transaction{
		ID:                   t.ID,
		Amount:               amount,
		Currency:             t.Currency,
		Description:          t.Description,
		MerchantCategoryCode: t.MerchantCategoryCode,
		PostedAt:             t.PostedAt.UTC(),
	}, nil
}

type transferRequestDTO struct {
	FromAccountNumber string       `json:"from_account_number"`
	ToAccountNumber   string       `json:"to_account_number"`
	RoutingNumber     string       `json:"routing_number"`
	Amount            domain.Money `json:"amount"`
	Currency          string       `json:"currency"`
}

type transferDTO struct {
	ID        string      `json:"id"`
	Status    string      `json:"status"`
	Amount    json.Number `json:"amount"`
	CreatedAt time.Time   `json:"created_at"`
}

func (t transferDTO) domain() (domain.PartnerTransfer, error) {
	if t.ID == "" || t.CreatedAt.IsZero() {
		return domain.PartnerTransfer{}, errors.New("transfer required fields are missing")
	}
	switch t.Status {
	case domain.TransferPending, domain.TransferPosted, domain.TransferFailed, domain.TransferReturned:
	default:
		return domain.PartnerTransfer{}, fmt.Errorf("unknown transfer status %q", t.Status)
	}
	amount, err := domain.MoneyFromJSONNumber(t.Amount)
	if err != nil {
		return domain.PartnerTransfer{}, fmt.Errorf("parse transfer amount: %w", err)
	}
	return domain.PartnerTransfer{
		ID:        t.ID,
		Status:    t.Status,
		Amount:    amount,
		CreatedAt: t.CreatedAt.UTC(),
	}, nil
}

type externalTransactionRequestDTO struct {
	Amount               domain.Money `json:"amount"`
	Description          string       `json:"description"`
	MerchantCategoryCode string       `json:"merchant_category_code"`
}

type externalTransactionResponseDTO struct {
	Account     accountDTO     `json:"account"`
	Transaction transactionDTO `json:"transaction"`
}
