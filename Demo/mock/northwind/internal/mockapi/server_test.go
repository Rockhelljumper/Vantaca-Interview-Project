package mockapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const testAPIKey = "synthetic_test_key"

var fixedTime = time.Date(2026, time.July, 28, 16, 22, 0, 0, time.UTC)

func TestMoneyJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		money Money
		want  string
	}{
		{name: "positive", money: Money(482055), want: "4820.55"},
		{name: "negative", money: Money(-4217), want: "-42.17"},
		{name: "zero", money: Money(0), want: "0.00"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(test.money)
			if err != nil {
				t.Fatalf("marshal money: %v", err)
			}
			if got := string(data); got != test.want {
				t.Fatalf("marshal money = %s, want %s", got, test.want)
			}
		})
	}

	for _, raw := range []string{"250.001", "1e2", "null"} {
		var amount Money
		if err := json.Unmarshal([]byte(raw), &amount); err == nil {
			t.Fatalf("json.Unmarshal(%s) unexpectedly succeeded", raw)
		}
	}
}

func TestAccountsAuthenticationAndSpecData(t *testing.T) {
	t.Parallel()

	mock := newTestServer(t, Config{})
	defer mock.Close()

	response, body := doRequest(t, mock.Client(), http.MethodGet, mock.URL+"/v1/accounts", "", nil)
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want %d; body=%s", response.StatusCode, http.StatusUnauthorized, body)
	}

	response, body = doRequest(t, mock.Client(), http.MethodGet, authorizedURL(mock.URL, "/v1/accounts"), "", nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("accounts status = %d, want 200; body=%s", response.StatusCode, body)
	}

	var accounts []Account
	mustDecode(t, body, &accounts)
	if len(accounts) != 3 {
		t.Fatalf("accounts length = %d, want 3", len(accounts))
	}

	got := accounts[0]
	if got.ID != "acc_1029" ||
		got.AccountNumber != "000123454321" ||
		got.RoutingNumber != "021000021" ||
		got.Type != "checking" ||
		got.Balance != Money(482055) ||
		got.Currency != "USD" ||
		got.Status != "open" {
		t.Fatalf("first account does not match documented seed: %+v", got)
	}

	response, body = doRequest(t, mock.Client(), http.MethodGet, authorizedURL(mock.URL, "/v1/accounts")+"&page=2", "", nil)
	if response.StatusCode != http.StatusOK || strings.TrimSpace(body) != "[]" {
		t.Fatalf("empty page status/body = %d/%q, want 200/[]", response.StatusCode, body)
	}
}

func TestTransactionsAndNotFound(t *testing.T) {
	t.Parallel()

	mock := newTestServer(t, Config{})
	defer mock.Close()

	response, body := doRequest(t, mock.Client(), http.MethodGet, authorizedURL(mock.URL, "/v1/accounts/acc_1029/transactions"), "", nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("transactions status = %d, want 200; body=%s", response.StatusCode, body)
	}

	var transactions []Transaction
	mustDecode(t, body, &transactions)
	if len(transactions) < 1 {
		t.Fatal("expected seeded transactions")
	}
	got := transactions[0]
	if got.ID != "txn_88213" || got.Amount != Money(-4217) || got.Description != "COFFEE HOUSE #42" || !got.PostedAt.Equal(time.Date(2026, time.July, 21, 14, 3, 0, 0, time.UTC)) {
		t.Fatalf("first transaction does not match documented seed: %+v", got)
	}

	response, body = doRequest(t, mock.Client(), http.MethodGet, authorizedURL(mock.URL, "/v1/accounts/does_not_exist/transactions"), "", nil)
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("missing account status = %d, want 404; body=%s", response.StatusCode, body)
	}
	assertAPIError(t, body, "invalid_account")
}

func TestCreateAndListTransfer(t *testing.T) {
	t.Parallel()

	mock := newTestServer(t, Config{})
	defer mock.Close()

	response, body := createTransfer(t, mock.Client(), mock.URL, nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("create transfer status = %d, want 200; body=%s", response.StatusCode, body)
	}
	if !strings.Contains(body, `"amount":250.00`) {
		t.Fatalf("transfer response should preserve two-decimal JSON amount: %s", body)
	}

	var transfer Transfer
	mustDecode(t, body, &transfer)
	if !strings.HasPrefix(transfer.ID, "trf_") || transfer.Status != StatusPending || transfer.Amount != Money(25000) || !transfer.CreatedAt.Equal(fixedTime) {
		t.Fatalf("created transfer does not match contract: %+v", transfer)
	}

	response, body = doRequest(t, mock.Client(), http.MethodGet, authorizedURL(mock.URL, "/v1/transfers"), "", nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("list transfers status = %d, want 200; body=%s", response.StatusCode, body)
	}

	var transfers []Transfer
	mustDecode(t, body, &transfers)
	if len(transfers) != 1 || transfers[0].ID != transfer.ID {
		t.Fatalf("listed transfers = %+v, want created transfer", transfers)
	}
}

func TestTransferValidation(t *testing.T) {
	t.Parallel()

	mock := newTestServer(t, Config{})
	defer mock.Close()

	tests := []struct {
		name     string
		body     string
		wantCode string
		wantHTTP int
	}{
		{
			name:     "unknown field",
			body:     `{"from_account_number":"000123454321","to_account_number":"000987656789","routing_number":"021000021","amount":250.00,"currency":"USD","unexpected":true}`,
			wantCode: "invalid_request",
			wantHTTP: http.StatusBadRequest,
		},
		{
			name:     "too many decimals",
			body:     `{"from_account_number":"000123454321","to_account_number":"000987656789","routing_number":"021000021","amount":250.001,"currency":"USD"}`,
			wantCode: "invalid_request",
			wantHTTP: http.StatusBadRequest,
		},
		{
			name:     "unknown account",
			body:     `{"from_account_number":"does_not_exist","to_account_number":"000987656789","routing_number":"021000021","amount":250.00,"currency":"USD"}`,
			wantCode: "invalid_account",
			wantHTTP: http.StatusNotFound,
		},
		{
			name:     "invalid routing",
			body:     `{"from_account_number":"000123454321","to_account_number":"000987656789","routing_number":"999999999","amount":250.00,"currency":"USD"}`,
			wantCode: "invalid_routing_number",
			wantHTTP: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			response, body := doRequest(t, mock.Client(), http.MethodPost, authorizedURL(mock.URL, "/v1/transfers"), test.body, nil)
			if response.StatusCode != test.wantHTTP {
				t.Fatalf("status = %d, want %d; body=%s", response.StatusCode, test.wantHTTP, body)
			}
			assertAPIError(t, body, test.wantCode)
		})
	}
}

func TestDocumentedFailureScenarios(t *testing.T) {
	t.Parallel()

	mock := newTestServer(t, Config{})
	defer mock.Close()

	tests := []struct {
		scenario string
		status   int
		code     string
	}{
		{scenario: "429", status: http.StatusTooManyRequests, code: "rate_limited"},
		{scenario: "500", status: http.StatusInternalServerError, code: "server_error"},
		{scenario: "503", status: http.StatusServiceUnavailable, code: "service_unavailable"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.scenario, func(t *testing.T) {
			headers := map[string]string{ScenarioHeader: test.scenario}
			response, body := doRequest(t, mock.Client(), http.MethodGet, authorizedURL(mock.URL, "/v1/accounts"), "", headers)
			if response.StatusCode != test.status {
				t.Fatalf("status = %d, want %d; body=%s", response.StatusCode, test.status, body)
			}
			if test.status == http.StatusTooManyRequests && response.Header.Get("Retry-After") != "1" {
				t.Fatalf("Retry-After = %q, want 1", response.Header.Get("Retry-After"))
			}
			assertAPIError(t, body, test.code)
		})
	}
}

func TestPostCommitTimeoutIsAmbiguous(t *testing.T) {
	t.Parallel()

	mock := newTestServer(t, Config{ScenarioDelay: 200 * time.Millisecond})
	defer mock.Close()

	timeoutClient := &http.Client{Timeout: 20 * time.Millisecond}
	request, err := http.NewRequest(http.MethodPost, authorizedURL(mock.URL, "/v1/transfers"), strings.NewReader(documentedTransferBody))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(ScenarioHeader, scenarioPostCommitTimeout)

	if _, err := timeoutClient.Do(request); err == nil {
		t.Fatal("post-commit timeout request unexpectedly returned a response")
	}

	response, body := doRequest(t, mock.Client(), http.MethodGet, authorizedURL(mock.URL, "/v1/transfers"), "", nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("list transfers status = %d, want 200; body=%s", response.StatusCode, body)
	}

	var transfers []Transfer
	mustDecode(t, body, &transfers)
	if len(transfers) != 1 || !strings.HasPrefix(transfers[0].ID, "trf_") {
		t.Fatalf("timed-out request should still be committed, got %+v", transfers)
	}
}

func TestStatusUpdateRetriesAndDuplicateWebhookDelivery(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	var eventsMu sync.Mutex
	var events []WebhookEvent

	webhookReceiver := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()

		var event WebhookEvent
		if err := json.NewDecoder(request.Body).Decode(&event); err != nil {
			t.Errorf("decode webhook: %v", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		eventsMu.Lock()
		events = append(events, event)
		eventsMu.Unlock()

		if calls.Add(1) <= 2 {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer webhookReceiver.Close()

	mock := newTestServer(t, Config{
		WebhookURL:      webhookReceiver.URL,
		WebhookAttempts: 3,
		WebhookBackoff:  time.Millisecond,
	})
	defer mock.Close()

	response, body := createTransfer(t, mock.Client(), mock.URL, nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("create transfer status = %d, want 200; body=%s", response.StatusCode, body)
	}
	var created Transfer
	mustDecode(t, body, &created)

	updateBody := `{"status":"POSTED","deliveries":2}`
	response, body = doRequest(t, mock.Client(), http.MethodPost, authorizedURL(mock.URL, "/__mock/transfers/"+created.ID+"/status"), updateBody, nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status update = %d, want 200; body=%s", response.StatusCode, body)
	}

	var update StatusUpdateResponse
	mustDecode(t, body, &update)
	if update.Transfer.Status != StatusPosted || update.WebhookDeliveries != 2 {
		t.Fatalf("status update response = %+v", update)
	}
	if got := calls.Load(); got != 4 {
		t.Fatalf("webhook calls = %d, want 4 (two failures, one success, one duplicate success)", got)
	}

	eventsMu.Lock()
	defer eventsMu.Unlock()
	for _, event := range events {
		if event.Event != "transfer.updated" || event.TransferID != created.ID || event.Status != StatusPosted {
			t.Fatalf("unexpected webhook event: %+v", event)
		}
	}

	response, body = doRequest(
		t,
		mock.Client(),
		http.MethodPost,
		authorizedURL(mock.URL, "/__mock/transfers/"+created.ID+"/status"),
		`{"status":"FAILED"}`,
		nil,
	)
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid transition status = %d, want 400; body=%s", response.StatusCode, body)
	}
	assertAPIError(t, body, "invalid_transition")
}

func TestConcurrentTransferIDsAreUnique(t *testing.T) {
	t.Parallel()

	mock := newTestServer(t, Config{})
	defer mock.Close()

	const requests = 20
	var wg sync.WaitGroup
	errs := make(chan string, requests)

	for range requests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			response, body, err := performRequest(
				mock.Client(),
				http.MethodPost,
				authorizedURL(mock.URL, "/v1/transfers"),
				documentedTransferBody,
				nil,
			)
			if err != nil {
				errs <- err.Error()
				return
			}
			if response.StatusCode != http.StatusOK {
				errs <- body
			}
		}()
	}

	wg.Wait()
	close(errs)
	for body := range errs {
		t.Errorf("concurrent create failed: %s", body)
	}

	response, body := doRequest(t, mock.Client(), http.MethodGet, authorizedURL(mock.URL, "/v1/transfers"), "", nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("list transfers status = %d, want 200; body=%s", response.StatusCode, body)
	}

	var transfers []Transfer
	mustDecode(t, body, &transfers)
	if len(transfers) != requests {
		t.Fatalf("transfers length = %d, want %d", len(transfers), requests)
	}

	ids := make(map[string]struct{}, requests)
	for _, transfer := range transfers {
		if _, exists := ids[transfer.ID]; exists {
			t.Fatalf("duplicate transfer ID %s", transfer.ID)
		}
		ids[transfer.ID] = struct{}{}
	}
}

func TestTransferIDsAreNamespacedPerMockBoot(t *testing.T) {
	t.Parallel()

	firstStore := NewStore(func() time.Time { return fixedTime })
	secondStore := NewStore(func() time.Time { return fixedTime })
	request := TransferRequest{Amount: Money(100)}

	first := firstStore.CreateTransfer(request)
	second := secondStore.CreateTransfer(request)
	if first.ID == second.ID {
		t.Fatalf("separate mock boots generated the same transfer ID %q", first.ID)
	}
}

func TestPaginationUsesPageSizeFifty(t *testing.T) {
	t.Parallel()

	records := make([]int, 51)
	for index := range records {
		records[index] = index
	}

	if got := len(paginate(records, 1)); got != 50 {
		t.Fatalf("page 1 length = %d, want 50", got)
	}
	if got := paginate(records, 2); len(got) != 1 || got[0] != 50 {
		t.Fatalf("page 2 = %+v, want [50]", got)
	}
	if got := paginate(records, 3); len(got) != 0 {
		t.Fatalf("page 3 = %+v, want empty", got)
	}
}

const documentedTransferBody = `{"from_account_number":"000123454321","to_account_number":"000987656789","routing_number":"021000021","amount":250.00,"currency":"USD"}`

func newTestServer(t *testing.T, overrides Config) *httptest.Server {
	t.Helper()

	config := overrides
	config.APIKey = testAPIKey
	config.Clock = func() time.Time { return fixedTime }

	server, err := NewServer(config, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	return httptest.NewServer(server.Handler())
}

func createTransfer(t *testing.T, client *http.Client, baseURL string, headers map[string]string) (*http.Response, string) {
	t.Helper()
	return doRequest(t, client, http.MethodPost, authorizedURL(baseURL, "/v1/transfers"), documentedTransferBody, headers)
}

func authorizedURL(baseURL string, path string) string {
	return baseURL + path + "?api_key=" + testAPIKey
}

func doRequest(
	t *testing.T,
	client *http.Client,
	method string,
	url string,
	body string,
	headers map[string]string,
) (*http.Response, string) {
	t.Helper()

	response, responseBody, err := performRequest(client, method, url, body, headers)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}

	return response, responseBody
}

func performRequest(
	client *http.Client,
	method string,
	url string,
	body string,
	headers map[string]string,
) (*http.Response, string, error) {
	request, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, "", err
	}
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	for name, value := range headers {
		request.Header.Set(name, value)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}

	return response, string(responseBody), nil
}

func mustDecode(t *testing.T, raw string, destination any) {
	t.Helper()

	if err := json.Unmarshal([]byte(raw), destination); err != nil {
		t.Fatalf("decode response %s: %v", strconv.Quote(raw), err)
	}
}

func assertAPIError(t *testing.T, raw string, wantCode string) {
	t.Helper()

	var response ErrorResponse
	mustDecode(t, raw, &response)
	if response.Error != wantCode {
		t.Fatalf("error code = %q, want %q; response=%s", response.Error, wantCode, raw)
	}
}
