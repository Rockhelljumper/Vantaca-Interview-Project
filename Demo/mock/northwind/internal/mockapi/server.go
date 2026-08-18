package mockapi

import (
	"crypto/subtle"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	ScenarioHeader            = "X-Northwind-Mock-Scenario"
	scenarioRateLimited       = "429"
	scenarioServerError       = "500"
	scenarioUnavailable       = "503"
	scenarioLatency           = "latency"
	scenarioPostCommitTimeout = "post-commit-timeout"
	maxRequestBodyBytes       = 1 << 20
)

//go:embed openapi.yaml
var openAPISpec []byte

type Config struct {
	APIKey          string
	WebhookURL      string
	WebhookAttempts int
	WebhookBackoff  time.Duration
	ScenarioDelay   time.Duration
	HTTPClient      *http.Client
	Clock           func() time.Time
	SwaggerUIOrigin string
}

type Server struct {
	apiKey        string
	store         *Store
	logger        *slog.Logger
	webhooks      webhookSender
	scenarioDelay time.Duration
	swaggerOrigin string
	handler       http.Handler
}

func NewServer(config Config, logger *slog.Logger) (*Server, error) {
	if config.APIKey == "" {
		return nil, errors.New("mock API key is required")
	}
	if config.ScenarioDelay <= 0 {
		config.ScenarioDelay = 5 * time.Second
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	server := &Server{
		apiKey:        config.APIKey,
		store:         NewStore(config.Clock),
		logger:        logger,
		webhooks:      newWebhookSender(config.WebhookURL, config.HTTPClient, config.WebhookAttempts, config.WebhookBackoff),
		scenarioDelay: config.ScenarioDelay,
		swaggerOrigin: config.SwaggerUIOrigin,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", server.handleHealth)
	mux.HandleFunc("GET /openapi.yaml", server.handleOpenAPI)
	mux.HandleFunc("GET /v1/accounts", server.handleAccounts)
	mux.HandleFunc("GET /v1/accounts/{id}/transactions", server.handleTransactions)
	mux.HandleFunc("POST /v1/transfers", server.handleCreateTransfer)
	mux.HandleFunc("GET /v1/transfers", server.handleTransfers)
	mux.HandleFunc("POST /__mock/accounts/{id}/transactions", server.handleExternalTransaction)
	mux.HandleFunc("POST /__mock/transfers/{id}/status", server.handleStatusUpdate)

	server.handler = server.loggingMiddleware(server.swaggerCORSMiddleware(server.authenticationMiddleware(server.scenarioMiddleware(mux))))
	return server, nil
}

func (s *Server) handleOpenAPI(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	writer.Header().Set("Content-Disposition", `inline; filename="northwind-mock-openapi.yaml"`)
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(openAPISpec)
}

func (s *Server) Handler() http.Handler {
	return s.handler
}

func (s *Server) handleHealth(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleAccounts(writer http.ResponseWriter, request *http.Request) {
	page, ok := parsePage(writer, request)
	if !ok {
		return
	}

	writeJSON(writer, http.StatusOK, s.store.ListAccounts(page))
}

func (s *Server) handleTransactions(writer http.ResponseWriter, request *http.Request) {
	page, ok := parsePage(writer, request)
	if !ok {
		return
	}

	accountID := request.PathValue("id")
	transactions, err := s.store.ListTransactions(accountID, page)
	if errors.Is(err, ErrAccountNotFound) {
		writeError(writer, http.StatusNotFound, "invalid_account", "Account not found")
		return
	}
	if err != nil {
		s.logger.Error("list transactions", "account_id", accountID, "error", err)
		writeError(writer, http.StatusInternalServerError, "server_error", "Something went wrong")
		return
	}

	writeJSON(writer, http.StatusOK, transactions)
}

func (s *Server) handleCreateTransfer(writer http.ResponseWriter, request *http.Request) {
	var transferRequest TransferRequest
	if err := decodeRequest(writer, request, &transferRequest); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if code, message := s.validateTransfer(transferRequest); code != "" {
		status := http.StatusBadRequest
		if code == "invalid_account" {
			status = http.StatusNotFound
		}
		writeError(writer, status, code, message)
		return
	}

	transfer := s.store.CreateTransfer(transferRequest)

	if request.Header.Get(ScenarioHeader) == scenarioPostCommitTimeout {
		timer := time.NewTimer(s.scenarioDelay)
		defer timer.Stop()

		select {
		case <-request.Context().Done():
			return
		case <-timer.C:
		}
	}

	writeJSON(writer, http.StatusOK, transfer)
}

func (s *Server) handleTransfers(writer http.ResponseWriter, request *http.Request) {
	page, ok := parsePage(writer, request)
	if !ok {
		return
	}

	writeJSON(writer, http.StatusOK, s.store.ListTransfers(page))
}

func (s *Server) handleExternalTransaction(writer http.ResponseWriter, request *http.Request) {
	var transactionRequest ExternalTransactionRequest
	if err := decodeRequest(writer, request, &transactionRequest); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if transactionRequest.Amount == 0 {
		writeError(writer, http.StatusBadRequest, "invalid_amount", "Amount must not be zero")
		return
	}
	transactionRequest.Description = strings.TrimSpace(transactionRequest.Description)
	if transactionRequest.Description == "" || len(transactionRequest.Description) > 120 {
		writeError(writer, http.StatusBadRequest, "invalid_description", "Description must contain 1 through 120 characters")
		return
	}
	if transactionRequest.MerchantCategoryCode != "" {
		if len(transactionRequest.MerchantCategoryCode) != 4 {
			writeError(writer, http.StatusBadRequest, "invalid_mcc", "Merchant category code must contain four digits")
			return
		}
		for _, character := range transactionRequest.MerchantCategoryCode {
			if character < '0' || character > '9' {
				writeError(writer, http.StatusBadRequest, "invalid_mcc", "Merchant category code must contain four digits")
				return
			}
		}
	}

	account, transaction, err := s.store.AddExternalTransaction(
		request.PathValue("id"),
		transactionRequest.Amount,
		transactionRequest.Description,
		transactionRequest.MerchantCategoryCode,
	)
	if errors.Is(err, ErrAccountNotFound) {
		writeError(writer, http.StatusNotFound, "invalid_account", "Account not found")
		return
	}
	if err != nil {
		s.logger.Error("create external transaction", "account_id", request.PathValue("id"), "error", err)
		writeError(writer, http.StatusInternalServerError, "server_error", "Something went wrong")
		return
	}

	writeJSON(writer, http.StatusCreated, ExternalTransactionResponse{
		Account:     account,
		Transaction: transaction,
	})
}

func (s *Server) handleStatusUpdate(writer http.ResponseWriter, request *http.Request) {
	var update StatusUpdateRequest
	if err := decodeRequest(writer, request, &update); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if !validStatus(update.Status) {
		writeError(writer, http.StatusBadRequest, "invalid_status", "Status must be PENDING, POSTED, FAILED, or RETURNED")
		return
	}
	if update.Deliveries == 0 {
		update.Deliveries = 1
	}
	if update.Deliveries < 1 || update.Deliveries > 5 {
		writeError(writer, http.StatusBadRequest, "invalid_deliveries", "Deliveries must be between 1 and 5")
		return
	}

	transfer, err := s.store.UpdateTransferStatus(request.PathValue("id"), update.Status)
	switch {
	case errors.Is(err, ErrTransferNotFound):
		writeError(writer, http.StatusNotFound, "invalid_transfer", "Transfer not found")
		return
	case errors.Is(err, ErrInvalidTransition):
		writeError(writer, http.StatusBadRequest, "invalid_transition", "Transfer status transition is not allowed")
		return
	case err != nil:
		s.logger.Error("update transfer status", "transfer_id", request.PathValue("id"), "error", err)
		writeError(writer, http.StatusInternalServerError, "server_error", "Something went wrong")
		return
	}

	delivered := 0
	for range update.Deliveries {
		attempts, deliveryErr := s.webhooks.Deliver(request.Context(), WebhookEvent{
			Event:      "transfer.updated",
			TransferID: transfer.ID,
			Status:     transfer.Status,
		})
		if deliveryErr != nil {
			s.logger.Warn("webhook delivery failed", "transfer_id", transfer.ID, "attempts", attempts, "error", deliveryErr)
			writeError(writer, http.StatusBadGateway, "webhook_delivery_failed", "Transfer updated but webhook delivery failed")
			return
		}
		if attempts > 0 {
			delivered++
		}
	}

	writeJSON(writer, http.StatusOK, StatusUpdateResponse{
		Transfer:          transfer,
		WebhookDeliveries: delivered,
	})
}

func (s *Server) validateTransfer(request TransferRequest) (string, string) {
	if request.FromAccountNumber == "" || request.ToAccountNumber == "" || request.RoutingNumber == "" || request.Currency == "" {
		return "missing_field", "from_account_number, to_account_number, routing_number, amount, and currency are required"
	}
	if request.Amount <= 0 {
		return "invalid_amount", "Amount must be greater than zero"
	}
	if request.Currency != "USD" {
		return "invalid_currency", "Only USD is supported by this mock"
	}
	if request.FromAccountNumber == request.ToAccountNumber {
		return "invalid_account", "Source and destination accounts must differ"
	}

	fromAccount, fromOK := s.store.AccountByNumber(request.FromAccountNumber)
	toAccount, toOK := s.store.AccountByNumber(request.ToAccountNumber)
	if !fromOK || !toOK || fromAccount.Status != "open" || toAccount.Status != "open" {
		return "invalid_account", "Account not found"
	}

	// The public guide does not say which account the single routing_number
	// identifies. The mock treats it as the destination routing number.
	if request.RoutingNumber != toAccount.RoutingNumber {
		return "invalid_routing_number", "Routing number does not match the destination account"
	}

	return "", ""
}

func (s *Server) authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/healthz" || request.URL.Path == "/openapi.yaml" {
			next.ServeHTTP(writer, request)
			return
		}

		provided := request.URL.Query().Get("api_key")
		if subtle.ConstantTimeCompare([]byte(provided), []byte(s.apiKey)) != 1 {
			writeError(writer, http.StatusUnauthorized, "unauthorized", "Missing or invalid api_key")
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func (s *Server) scenarioMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		scenario := request.Header.Get(ScenarioHeader)
		switch scenario {
		case "":
			next.ServeHTTP(writer, request)
		case scenarioRateLimited:
			writer.Header().Set("Retry-After", "1")
			writeError(writer, http.StatusTooManyRequests, "rate_limited", "Too many requests")
		case scenarioServerError:
			writeError(writer, http.StatusInternalServerError, "server_error", "Something went wrong")
		case scenarioUnavailable:
			writeError(writer, http.StatusServiceUnavailable, "service_unavailable", "Service is unavailable")
		case scenarioLatency:
			timer := time.NewTimer(s.scenarioDelay)
			defer timer.Stop()
			select {
			case <-request.Context().Done():
				return
			case <-timer.C:
				next.ServeHTTP(writer, request)
			}
		case scenarioPostCommitTimeout:
			if request.Method != http.MethodPost || request.URL.Path != "/v1/transfers" {
				writeError(writer, http.StatusBadRequest, "invalid_scenario", "post-commit-timeout is valid only for POST /v1/transfers")
				return
			}
			next.ServeHTTP(writer, request)
		default:
			writeError(writer, http.StatusBadRequest, "invalid_scenario", "Unknown mock scenario")
		}
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: writer, status: http.StatusOK}

		next.ServeHTTP(recorder, request)

		s.logger.Info(
			"request",
			"method", request.Method,
			"path", request.URL.Path,
			"status", recorder.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// swaggerCORSMiddleware is intentionally limited to the configured local
// documentation origin. Production partner APIs need their own reviewed CORS
// policy rather than inheriting this synthetic-demo allowance.
func (s *Server) swaggerCORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if s.swaggerOrigin != "" && origin == s.swaggerOrigin {
			writer.Header().Add("Vary", "Origin")
			writer.Header().Set("Access-Control-Allow-Origin", origin)
			writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Northwind-Mock-Scenario")
			writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			writer.Header().Set("Access-Control-Max-Age", "600")
			if request.Method == http.MethodOptions {
				writer.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(writer, request)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func parsePage(writer http.ResponseWriter, request *http.Request) (int, bool) {
	raw := request.URL.Query().Get("page")
	if raw == "" {
		return 1, true
	}

	page, err := strconv.Atoi(raw)
	if err != nil || page < 1 {
		writeError(writer, http.StatusBadRequest, "invalid_page", "page must be a positive integer")
		return 0, false
	}

	return page, true
}

func decodeRequest(writer http.ResponseWriter, request *http.Request, destination any) error {
	body, err := io.ReadAll(http.MaxBytesReader(writer, request.Body, maxRequestBodyBytes))
	if err != nil {
		return fmt.Errorf("read request body: %w", err)
	}

	if len(strings.TrimSpace(string(body))) == 0 {
		return errors.New("request body is required")
	}

	if err := decodeJSON(body, destination); err != nil {
		return fmt.Errorf("decode request body: %w", err)
	}

	return nil
}

func writeError(writer http.ResponseWriter, status int, code string, message string) {
	writeJSON(writer, status, ErrorResponse{Error: code, Message: message})
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)

	if err := json.NewEncoder(writer).Encode(value); err != nil {
		slog.Default().Error("encode response", "status", status, "error", err)
	}
}
