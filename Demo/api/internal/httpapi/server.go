package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"vantaca-interview-project/Demo/api/internal/application"
	"vantaca-interview-project/Demo/api/internal/domain"
	"vantaca-interview-project/Demo/api/internal/northwind"
)

const maxRequestBody = 64 << 10

//go:embed openapi.yaml
var openAPISpec []byte

type Pinger interface {
	Ping(context.Context) error
}

type Options struct {
	TenantID              string
	CustomerID            string
	FreshnessThreshold    time.Duration
	DemoControlsEnabled   bool
	AllowUnsignedWebhooks bool
	DemoAdminKey          string
	SwaggerUIOrigin       string
}

type Server struct {
	options     Options
	pinger      Pinger
	repository  application.Repository
	syncService *application.SyncService
	coordinator *application.RefreshCoordinator
	transfers   *application.TransferService
	logger      *slog.Logger
	handler     http.Handler
}

func NewServer(
	options Options,
	pinger Pinger,
	repository application.Repository,
	syncService *application.SyncService,
	coordinator *application.RefreshCoordinator,
	transfers *application.TransferService,
	logger *slog.Logger,
) *Server {
	server := &Server{
		options:     options,
		pinger:      pinger,
		repository:  repository,
		syncService: syncService,
		coordinator: coordinator,
		transfers:   transfers,
		logger:      logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", server.handleHealth)
	mux.HandleFunc("GET /openapi.yaml", server.handleOpenAPI)
	mux.HandleFunc("GET /api/demo/info", server.handleDemoInfo)
	mux.HandleFunc("GET /api/accounts", server.handleAccounts)
	mux.HandleFunc("GET /api/accounts/{id}/transactions", server.handleTransactions)
	mux.HandleFunc("POST /api/transfers", server.handleSubmitTransfer)
	mux.HandleFunc("GET /api/transfers", server.handleTransfers)
	mux.HandleFunc("POST /api/webhooks/northwind", server.handleNorthwindWebhook)
	mux.HandleFunc("POST /api/demo/sync", server.handleDemoSync)
	mux.HandleFunc("POST /api/demo/accounts/{id}/external-activity", server.handleExternalActivity)
	mux.HandleFunc("POST /api/demo/transfers/{id}/advance", server.handleAdvanceTransfer)
	mux.HandleFunc("POST /api/internal/reconcile", server.handleInternalReconcile)

	server.handler = server.recoveryMiddleware(
		server.correlationMiddleware(
			server.loggingMiddleware(
				server.securityHeadersMiddleware(server.swaggerCORSMiddleware(mux)),
			),
		),
	)
	return server
}

func (s *Server) handleOpenAPI(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	writer.Header().Set("Content-Disposition", `inline; filename="vantaca-demo-openapi.yaml"`)
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(openAPISpec)
}

func (s *Server) Handler() http.Handler {
	return s.handler
}

func (s *Server) handleHealth(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(request.Context(), 2*time.Second)
	defer cancel()
	if err := s.pinger.Ping(ctx); err != nil {
		writeError(writer, request, http.StatusServiceUnavailable, "database_unavailable", "The demo database is unavailable.")
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{
		"status":   "ok",
		"database": "connected",
	})
}

func (s *Server) handleDemoInfo(writer http.ResponseWriter, request *http.Request) {
	if _, ok := s.requireTenant(writer, request); !ok {
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{
		"mode":                        "demo",
		"customer_id":                 s.options.CustomerID,
		"northwind_authoritative":     true,
		"read_model":                  "SQL Server 2022",
		"demo_controls_enabled":       s.options.DemoControlsEnabled,
		"unsigned_webhook_mode":       s.options.AllowUnsignedWebhooks,
		"transfer_submission_enabled": true,
		"production_blockers": []string{
			"Vantaca identity and tenant contract",
			"Northwind transfer idempotency and exact ambiguous-outcome lookup",
			"Webhook authenticity and event identity",
			"Security-approved protected-data and platform controls",
		},
		"demo_assumptions": []string{
			"A single synthetic tenant/customer is supplied by the UI header.",
			"Full account and routing values are fetched in memory and never persisted.",
			"Unsigned webhooks are reconciliation signals, not authoritative state.",
			"Frontend invalidation uses bounded SQL-version polling.",
		},
	})
}

func (s *Server) handleAccounts(writer http.ResponseWriter, request *http.Request) {
	tenantID, ok := s.requireTenant(writer, request)
	if !ok {
		return
	}
	accounts, err := s.repository.ListAccounts(request.Context(), tenantID)
	if err != nil {
		s.logger.Error("list accounts", "correlation_id", correlationID(request), "error", err)
		writeError(writer, request, http.StatusInternalServerError, "account_read_failed", "Accounts could not be loaded.")
		return
	}

	response := make([]accountResponse, 0, len(accounts))
	for _, account := range accounts {
		response = append(response, s.accountResponse(account))
	}
	writeJSON(writer, http.StatusOK, map[string]any{
		"accounts": response,
		"source":   "Vantaca SQL read model",
	})
}

func (s *Server) handleTransactions(writer http.ResponseWriter, request *http.Request) {
	tenantID, ok := s.requireTenant(writer, request)
	if !ok {
		return
	}
	accountID := request.PathValue("id")
	account, err := s.repository.GetAccount(request.Context(), tenantID, accountID)
	if err != nil {
		if application.IsNotFound(err) {
			writeError(writer, request, http.StatusNotFound, "account_not_found", "Account was not found.")
			return
		}
		s.logger.Error("get transaction account", "correlation_id", correlationID(request), "error", err)
		writeError(writer, request, http.StatusInternalServerError, "transaction_read_failed", "Transactions could not be loaded.")
		return
	}
	transactions, err := s.repository.ListTransactions(request.Context(), tenantID, accountID)
	if err != nil {
		s.logger.Error("list transactions", "correlation_id", correlationID(request), "error", err)
		writeError(writer, request, http.StatusInternalServerError, "transaction_read_failed", "Transactions could not be loaded.")
		return
	}

	refreshStarted := false
	refreshRequested := request.URL.Query().Get("refresh") != "false"
	scenario := request.URL.Query().Get("scenario")
	if refreshRequested {
		if !validReadScenario(scenario) {
			writeError(writer, request, http.StatusBadRequest, "invalid_scenario", "Unsupported demo read scenario.")
			return
		}
		refreshStarted = s.coordinator.Start(tenantID, accountID, scenario)
	}

	items := make([]transactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		items = append(items, transactionResponse{
			ID:                   transaction.ID,
			Amount:               transaction.Amount.String(),
			Currency:             transaction.Currency,
			Description:          transaction.Description,
			MerchantCategoryCode: nullableResponse(transaction.MerchantCategoryCode),
			PostedAt:             transaction.PostedAt,
		})
	}
	writeJSON(writer, http.StatusOK, map[string]any{
		"account":         s.accountResponse(account),
		"transactions":    items,
		"version":         account.Version,
		"refresh_started": refreshStarted,
		"refreshing":      s.coordinator.IsRunning(tenantID, accountID),
		"invalidation":    "bounded version polling",
	})
}

func (s *Server) handleDemoSync(writer http.ResponseWriter, request *http.Request) {
	tenantID, ok := s.requireDemoControl(writer, request)
	if !ok {
		return
	}
	var body struct {
		Scenario string `json:"scenario"`
	}
	if err := decodeJSON(writer, request, &body); err != nil {
		writeError(writer, request, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if !validReadScenario(body.Scenario) {
		writeError(writer, request, http.StatusBadRequest, "invalid_scenario", "Unsupported demo read scenario.")
		return
	}

	result, err := s.syncService.SyncAll(request.Context(), tenantID, body.Scenario)
	if err != nil {
		s.logger.Warn("demo synchronization failed", "correlation_id", correlationID(request), "category", northwind.Category(err))
		writeJSON(writer, http.StatusBadGateway, map[string]any{
			"error":          "northwind_sync_failed",
			"message":        "Northwind refresh failed; the last SQL snapshot was preserved.",
			"correlation_id": correlationID(request),
			"result":         result,
		})
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"result": result})
}

func (s *Server) handleExternalActivity(writer http.ResponseWriter, request *http.Request) {
	tenantID, ok := s.requireDemoControl(writer, request)
	if !ok {
		return
	}
	accountID := request.PathValue("id")
	account, err := s.repository.GetAccount(request.Context(), tenantID, accountID)
	if err != nil {
		if application.IsNotFound(err) {
			writeError(writer, request, http.StatusNotFound, "account_not_found", "Account was not found.")
			return
		}
		s.logger.Error("get external activity account", "correlation_id", correlationID(request), "error", err)
		writeError(writer, request, http.StatusInternalServerError, "account_read_failed", "The account could not be validated.")
		return
	}
	if account.Status != "open" {
		writeError(writer, request, http.StatusUnprocessableEntity, "external_activity_unavailable", "External activity can only be simulated for an open account.")
		return
	}

	activity, err := s.syncService.SimulateExternalActivity(request.Context(), accountID)
	if err != nil {
		s.logger.Warn("simulate external activity failed", "correlation_id", correlationID(request), "category", northwind.Category(err))
		writeError(writer, request, http.StatusBadGateway, "mock_control_failed", "The synthetic Northwind activity could not be created.")
		return
	}
	writeJSON(writer, http.StatusCreated, map[string]any{
		"message": "Northwind changed outside Vantaca. The SQL snapshot has not been refreshed yet.",
		"transaction": transactionResponse{
			ID:                   activity.Transaction.ID,
			Amount:               activity.Transaction.Amount.String(),
			Currency:             activity.Transaction.Currency,
			Description:          activity.Transaction.Description,
			MerchantCategoryCode: nullableResponse(activity.Transaction.MerchantCategoryCode),
			PostedAt:             activity.Transaction.PostedAt,
		},
	})
}

func (s *Server) handleSubmitTransfer(writer http.ResponseWriter, request *http.Request) {
	tenantID, ok := s.requireTenant(writer, request)
	if !ok {
		return
	}
	var body struct {
		RequestID     string `json:"request_id"`
		FromAccountID string `json:"from_account_id"`
		ToAccountID   string `json:"to_account_id"`
		Amount        string `json:"amount"`
		Currency      string `json:"currency"`
		Scenario      string `json:"scenario,omitempty"`
	}
	if err := decodeJSON(writer, request, &body); err != nil {
		writeError(writer, request, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if body.Scenario != "" && !s.options.DemoControlsEnabled {
		writeError(writer, request, http.StatusForbidden, "demo_controls_disabled", "Failure scenarios are disabled.")
		return
	}

	transfer, err := s.transfers.Submit(request.Context(), tenantID, application.SubmitTransferInput{
		RequestID:     body.RequestID,
		FromAccountID: body.FromAccountID,
		ToAccountID:   body.ToAccountID,
		Amount:        body.Amount,
		Currency:      body.Currency,
		Scenario:      body.Scenario,
	})
	if err != nil {
		if errors.Is(err, application.ErrValidation) {
			message := strings.TrimSuffix(err.Error(), ": "+application.ErrValidation.Error())
			writeError(writer, request, http.StatusUnprocessableEntity, "transfer_validation_failed", message)
			return
		}
		s.logger.Error("submit transfer", "correlation_id", correlationID(request), "error", err)
		writeError(writer, request, http.StatusInternalServerError, "transfer_submission_failed", "The transfer request could not be recorded.")
		return
	}
	writeJSON(writer, http.StatusAccepted, transferResponseFromDomain(transfer))
}

func (s *Server) handleTransfers(writer http.ResponseWriter, request *http.Request) {
	tenantID, ok := s.requireTenant(writer, request)
	if !ok {
		return
	}
	transfers, err := s.repository.ListTransfers(request.Context(), tenantID)
	if err != nil {
		s.logger.Error("list transfers", "correlation_id", correlationID(request), "error", err)
		writeError(writer, request, http.StatusInternalServerError, "transfer_read_failed", "Transfers could not be loaded.")
		return
	}
	response := make([]transferResponse, 0, len(transfers))
	for _, transfer := range transfers {
		response = append(response, transferResponseFromDomain(transfer))
	}
	writeJSON(writer, http.StatusOK, map[string]any{"transfers": response})
}

func (s *Server) handleAdvanceTransfer(writer http.ResponseWriter, request *http.Request) {
	tenantID, ok := s.requireDemoControl(writer, request)
	if !ok {
		return
	}
	var body struct {
		Status     string `json:"status"`
		Deliveries int    `json:"deliveries"`
	}
	if err := decodeJSON(writer, request, &body); err != nil {
		writeError(writer, request, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	transfer, err := s.transfers.AdvanceDemo(request.Context(), tenantID, request.PathValue("id"), body.Status, body.Deliveries)
	if err != nil {
		if errors.Is(err, application.ErrValidation) {
			message := strings.TrimSuffix(err.Error(), ": "+application.ErrValidation.Error())
			writeError(writer, request, http.StatusUnprocessableEntity, "invalid_transition", message)
			return
		}
		s.logger.Warn("advance demo transfer failed", "correlation_id", correlationID(request), "error", err)
		writeError(writer, request, http.StatusBadGateway, "mock_transition_failed", "The mock transfer transition failed.")
		return
	}
	writeJSON(writer, http.StatusOK, transferResponseFromDomain(transfer))
}

func (s *Server) handleNorthwindWebhook(writer http.ResponseWriter, request *http.Request) {
	if !s.options.AllowUnsignedWebhooks {
		writeError(writer, request, http.StatusUnauthorized, "webhook_authentication_required", "Webhook authenticity could not be verified.")
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(writer, request.Body, maxRequestBody))
	if err != nil {
		writeError(writer, request, http.StatusBadRequest, "invalid_webhook", "Webhook body could not be read.")
		return
	}
	var event struct {
		Event      string `json:"event"`
		TransferID string `json:"transfer_id"`
		Status     string `json:"status"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&event); err != nil {
		writeError(writer, request, http.StatusBadRequest, "invalid_webhook", "Webhook schema is invalid.")
		return
	}
	if event.Event != "transfer.updated" || event.TransferID == "" || !validPartnerStatus(event.Status) {
		writeError(writer, request, http.StatusBadRequest, "invalid_webhook", "Webhook fields are invalid.")
		return
	}

	hash := sha256.Sum256(body)
	dedupeKey := event.TransferID + ":" + event.Status
	receiptID, created, err := s.repository.RecordWebhook(
		request.Context(),
		dedupeKey,
		event.TransferID,
		event.Status,
		hex.EncodeToString(hash[:]),
		time.Now().UTC(),
	)
	if err != nil {
		s.logger.Error("record webhook", "correlation_id", correlationID(request), "error", err)
		writeError(writer, request, http.StatusInternalServerError, "webhook_persistence_failed", "Webhook could not be recorded.")
		return
	}
	if !created {
		writeJSON(writer, http.StatusOK, map[string]any{
			"status":  "duplicate_acknowledged",
			"trusted": false,
		})
		return
	}

	s.logger.Warn(
		"unsigned demo webhook received; using it only as a reconciliation signal",
		"correlation_id", correlationID(request),
		"partner_transfer_id", event.TransferID,
	)
	_, reconcileErr := s.transfers.ReconcilePartnerTransfer(request.Context(), s.options.TenantID, event.TransferID)
	outcome := "reconciled_from_partner_read"
	if reconcileErr != nil {
		outcome = "recorded_reconciliation_required"
		s.logger.Warn("webhook-triggered reconciliation incomplete", "partner_transfer_id", event.TransferID, "error", reconcileErr)
	}
	if err := s.repository.MarkWebhookProcessed(request.Context(), receiptID, outcome); err != nil {
		s.logger.Error("mark webhook processed", "receipt_id", receiptID, "error", err)
		writeError(writer, request, http.StatusInternalServerError, "webhook_processing_failed", "Webhook processing could not be finalized.")
		return
	}
	writeJSON(writer, http.StatusAccepted, map[string]any{
		"status":  outcome,
		"trusted": false,
	})
}

func (s *Server) handleInternalReconcile(writer http.ResponseWriter, request *http.Request) {
	if subtle.ConstantTimeCompare(
		[]byte(request.Header.Get("X-Demo-Admin-Key")),
		[]byte(s.options.DemoAdminKey),
	) != 1 {
		writeError(writer, request, http.StatusUnauthorized, "unauthorized", "Demo admin key is missing or invalid.")
		return
	}
	syncResult, syncErr := s.syncService.SyncAll(request.Context(), s.options.TenantID, "")
	transferUpdates, transferErr := s.transfers.ReconcileAllKnown(request.Context(), s.options.TenantID)
	if syncErr != nil || transferErr != nil {
		s.logger.Warn("internal reconciliation completed with errors", "sync_error", syncErr, "transfer_error", transferErr)
		writeJSON(writer, http.StatusBadGateway, map[string]any{
			"status":           "partial",
			"sync":             syncResult,
			"transfer_updates": transferUpdates,
		})
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{
		"status":           "complete",
		"sync":             syncResult,
		"transfer_updates": transferUpdates,
	})
}

func (s *Server) requireTenant(writer http.ResponseWriter, request *http.Request) (string, bool) {
	tenantID := request.Header.Get("X-Demo-Tenant")
	if subtle.ConstantTimeCompare([]byte(tenantID), []byte(s.options.TenantID)) != 1 {
		writeError(writer, request, http.StatusForbidden, "tenant_forbidden", "The synthetic demo tenant context is missing or invalid.")
		return "", false
	}
	return tenantID, true
}

func (s *Server) requireDemoControl(writer http.ResponseWriter, request *http.Request) (string, bool) {
	tenantID, ok := s.requireTenant(writer, request)
	if !ok {
		return "", false
	}
	if !s.options.DemoControlsEnabled {
		writeError(writer, request, http.StatusForbidden, "demo_controls_disabled", "Demo controls are disabled.")
		return "", false
	}
	return tenantID, true
}

func (s *Server) accountResponse(account domain.Account) accountResponse {
	state := "current"
	if account.LastSyncError != "" {
		state = "degraded"
	} else if time.Since(account.FetchedAt) > s.options.FreshnessThreshold {
		state = "stale"
	}
	return accountResponse{
		ID:          account.ID,
		DisplayName: account.DisplayName(),
		Type:        account.Type,
		LastFour:    account.LastFour,
		Balance:     account.Balance.String(),
		Currency:    account.Currency,
		Status:      account.Status,
		Version:     account.Version,
		Freshness: freshnessResponse{
			State:       state,
			FetchedAt:   account.FetchedAt,
			CheckedAt:   account.CheckedAt,
			LastError:   nullableResponse(account.LastSyncError),
			PolicyLabel: "Demo threshold; not a Northwind source as-of guarantee",
		},
	}
}

type freshnessResponse struct {
	State       string    `json:"state"`
	FetchedAt   time.Time `json:"fetched_at"`
	CheckedAt   time.Time `json:"checked_at"`
	LastError   *string   `json:"last_error"`
	PolicyLabel string    `json:"policy_label"`
}

type accountResponse struct {
	ID          string            `json:"id"`
	DisplayName string            `json:"display_name"`
	Type        string            `json:"type"`
	LastFour    string            `json:"last_four"`
	Balance     string            `json:"balance"`
	Currency    string            `json:"currency"`
	Status      string            `json:"status"`
	Version     int64             `json:"version"`
	Freshness   freshnessResponse `json:"freshness"`
}

type transactionResponse struct {
	ID                   string    `json:"id"`
	Amount               string    `json:"amount"`
	Currency             string    `json:"currency"`
	Description          string    `json:"description"`
	MerchantCategoryCode *string   `json:"merchant_category_code"`
	PostedAt             time.Time `json:"posted_at"`
}

type transferResponse struct {
	ID                string    `json:"id"`
	RequestID         string    `json:"request_id"`
	FromAccountID     string    `json:"from_account_id"`
	ToAccountID       string    `json:"to_account_id"`
	FromDisplay       string    `json:"from_display"`
	ToDisplay         string    `json:"to_display"`
	Amount            string    `json:"amount"`
	Currency          string    `json:"currency"`
	Status            string    `json:"status"`
	PartnerTransferID *string   `json:"partner_transfer_id"`
	ErrorCategory     *string   `json:"error_category"`
	Message           string    `json:"message"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func transferResponseFromDomain(transfer domain.Transfer) transferResponse {
	return transferResponse{
		ID:                transfer.ID,
		RequestID:         transfer.RequestID,
		FromAccountID:     transfer.FromAccountID,
		ToAccountID:       transfer.ToAccountID,
		FromDisplay:       transfer.FromDisplay,
		ToDisplay:         transfer.ToDisplay,
		Amount:            transfer.Amount.String(),
		Currency:          transfer.Currency,
		Status:            transfer.Status,
		PartnerTransferID: nullableResponse(transfer.PartnerTransferID),
		ErrorCategory:     nullableResponse(transfer.LastErrorCategory),
		Message:           transferMessage(transfer.Status),
		CreatedAt:         transfer.CreatedAt,
		UpdatedAt:         transfer.UpdatedAt,
	}
}

func transferMessage(status string) string {
	switch status {
	case domain.TransferIntentRecorded:
		return "The request is durable but has not received a Northwind outcome."
	case domain.TransferPending:
		return "Northwind accepted the transfer; it is pending and not yet complete."
	case domain.TransferPosted:
		return "Northwind reports the transfer posted. A later return remains possible."
	case domain.TransferFailed:
		return "The transfer received a definitive failure before completion."
	case domain.TransferReturned:
		return "Northwind reports that the posted transfer was returned."
	case domain.TransferUnknown:
		return "The submission outcome is unknown. Do not submit it again; exact reconciliation is required."
	default:
		return "Transfer state requires review."
	}
}

func nullableResponse(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func validReadScenario(value string) bool {
	switch value {
	case "", "429", "500", "503", "latency":
		return true
	default:
		return false
	}
}

func validPartnerStatus(value string) bool {
	switch value {
	case domain.TransferPending, domain.TransferPosted, domain.TransferFailed, domain.TransferReturned:
		return true
	default:
		return false
	}
}

type contextKey string

const correlationKey contextKey = "correlation_id"

var correlationPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{8,64}$`)

func (s *Server) correlationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		id := request.Header.Get("X-Correlation-ID")
		if !correlationPattern.MatchString(id) {
			var bytes [12]byte
			if _, err := rand.Read(bytes[:]); err != nil {
				id = fmt.Sprintf("fallback-%d", time.Now().UnixNano())
			} else {
				id = hex.EncodeToString(bytes[:])
			}
		}
		writer.Header().Set("X-Correlation-ID", id)
		ctx := context.WithValue(request.Context(), correlationKey, id)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

func correlationID(request *http.Request) string {
	value, _ := request.Context().Value(correlationKey).(string)
	return value
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: writer, status: http.StatusOK}
		next.ServeHTTP(recorder, request)
		s.logger.Info(
			"request",
			"method", request.Method,
			"path", request.URL.Path,
			"status", recorder.status,
			"duration_ms", time.Since(started).Milliseconds(),
			"correlation_id", correlationID(request),
		)
	})
}

func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				s.logger.Error("panic recovered", "path", request.URL.Path, "correlation_id", correlationID(request))
				writeError(writer, request, http.StatusInternalServerError, "internal_error", "The request could not be completed.")
			}
		}()
		next.ServeHTTP(writer, request)
	})
}

func (s *Server) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Cache-Control", "no-store")
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		writer.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(writer, request)
	})
}

// swaggerCORSMiddleware permits the explicitly configured local Swagger UI to
// exercise the synthetic APIs. It does not allow arbitrary origins and must
// not be copied into a production financial-data boundary unchanged.
func (s *Server) swaggerCORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if s.options.SwaggerUIOrigin != "" && origin == s.options.SwaggerUIOrigin {
			writer.Header().Add("Vary", "Origin")
			writer.Header().Set("Access-Control-Allow-Origin", origin)
			writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Demo-Tenant, X-Correlation-ID, X-Demo-Admin-Key")
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

func decodeJSON(writer http.ResponseWriter, request *http.Request, destination any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(writer, request.Body, maxRequestBody))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return errors.New("request body must be one valid JSON object with known fields")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain exactly one JSON object")
	}
	return nil
}

func writeError(writer http.ResponseWriter, request *http.Request, status int, code string, message string) {
	writeJSON(writer, status, map[string]string{
		"error":          code,
		"message":        message,
		"correlation_id": correlationID(request),
	})
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
