package httpapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type capturedHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (handler *capturedHandler) Enabled(context.Context, slog.Level) bool { return true }

func (handler *capturedHandler) Handle(_ context.Context, record slog.Record) error {
	handler.mu.Lock()
	defer handler.mu.Unlock()
	handler.records = append(handler.records, record.Clone())
	return nil
}

func (handler *capturedHandler) WithAttrs([]slog.Attr) slog.Handler { return handler }
func (handler *capturedHandler) WithGroup(string) slog.Handler      { return handler }

func (handler *capturedHandler) snapshot() []slog.Record {
	handler.mu.Lock()
	defer handler.mu.Unlock()
	return append([]slog.Record(nil), handler.records...)
}

func TestRequestLoggingExcludesQueryAndClassifiesFailures(t *testing.T) {
	capture := &capturedHandler{}
	server := &Server{logger: slog.New(capture)}
	handler := server.correlationMiddleware(server.loggingMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusBadGateway)
	})))
	request := httptest.NewRequest(http.MethodGet, "/api/accounts?api_key=must-not-be-logged-9876", nil)
	request.Header.Set("X-Correlation-ID", "correlation_1234")

	handler.ServeHTTP(httptest.NewRecorder(), request)

	records := capture.snapshot()
	if len(records) != 1 {
		t.Fatalf("record count = %d, want 1", len(records))
	}
	if records[0].Level != slog.LevelError {
		t.Fatalf("level = %s, want ERROR", records[0].Level)
	}
	attributes := recordAttributes(records[0])
	if got := attributes["path"]; got != "/api/accounts" {
		t.Fatalf("path = %#v, want path without query", got)
	}
	if got := attributes["correlation_id"]; got != "correlation_1234" {
		t.Fatalf("correlation_id = %#v, want correlation_1234", got)
	}
	if strings.Contains(fmt.Sprint(attributes), "must-not-be-logged-9876") {
		t.Fatalf("request log contains query-string secret: %#v", attributes)
	}
}

func TestRequestLoggingSuppressesSuccessfulHealthChecks(t *testing.T) {
	capture := &capturedHandler{}
	server := &Server{logger: slog.New(capture)}
	handler := server.correlationMiddleware(server.loggingMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if records := capture.snapshot(); len(records) != 0 {
		t.Fatalf("successful health check record count = %d, want 0", len(records))
	}
}

func TestRecoveredPanicKeepsCorrelationInErrorRecords(t *testing.T) {
	capture := &capturedHandler{}
	server := &Server{logger: slog.New(capture)}
	handler := server.correlationMiddleware(server.loggingMiddleware(server.recoveryMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("synthetic panic")
	}))))
	request := httptest.NewRequest(http.MethodGet, "/api/accounts", nil)
	request.Header.Set("X-Correlation-ID", "panic-correlation-1234")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("response status = %d, want 500", response.Code)
	}
	records := capture.snapshot()
	if len(records) != 2 {
		t.Fatalf("record count = %d, want panic and request records", len(records))
	}
	for _, record := range records {
		if got := recordAttributes(record)["correlation_id"]; got != "panic-correlation-1234" {
			t.Errorf("%q correlation_id = %#v", record.Message, got)
		}
	}
}

func recordAttributes(record slog.Record) map[string]any {
	attributes := make(map[string]any)
	record.Attrs(func(attr slog.Attr) bool {
		attributes[attr.Key] = attr.Value.Any()
		return true
	})
	return attributes
}
