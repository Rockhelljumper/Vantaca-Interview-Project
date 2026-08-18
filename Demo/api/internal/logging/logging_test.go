package logging

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"
)

type capturedExecution struct {
	query string
	args  []any
	err   error
}

func (execution *capturedExecution) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	execution.query = query
	execution.args = append([]any(nil), args...)
	return nil, execution.err
}

func TestDatabaseHandlerPersistsOnlySanitizedValues(t *testing.T) {
	execution := &capturedExecution{}
	handler := NewDatabaseHandler(execution, DatabaseHandlerOptions{
		Application:  "vantaca-api",
		MinimumLevel: slog.LevelInfo,
		WriteTimeout: time.Second,
	})

	record := slog.NewRecord(
		time.Date(2026, time.August, 18, 12, 30, 0, 0, time.UTC),
		slog.LevelError,
		"partner request failed: https://northwind.test/v1?api_key=raw-api-key-9876",
		0,
	)
	record.AddAttrs(
		slog.String("correlation_id", "correlation_1234"),
		slog.String("username", "alex"),
		slog.String("password", "RawPassword!"),
		slog.String("api_key", "raw-api-key-9876"),
		slog.String("request_body", `{"password":"nested-password"}`),
		slog.Any("details", map[string]any{
			"access_token":   "raw-access-token",
			"account_number": "000123456789",
			"routing_number": "021000021",
			"safe_category":  "timeout",
			"error":          errors.New("sqlserver://dbuser:raw-db-password@sql:1433?password=raw-query-password"),
		}),
	)

	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	arguments := namedArguments(t, execution.args)
	if got := arguments["username"]; got != "alex" {
		t.Fatalf("username = %#v, want alex", got)
	}
	if got := arguments["api_key_last_four"]; got != "9876" {
		t.Fatalf("api_key_last_four = %#v, want 9876", got)
	}
	if got := arguments["correlation_id"]; got != "correlation_1234" {
		t.Fatalf("correlation_id = %#v, want correlation_1234", got)
	}

	persisted := fmt.Sprint(execution.args)
	for _, rawSecret := range []string{
		"raw-api-key-9876",
		"RawPassword!",
		"nested-password",
		"raw-access-token",
		"000123456789",
		"021000021",
		"raw-db-password",
		"raw-query-password",
	} {
		if strings.Contains(persisted, rawSecret) {
			t.Errorf("persisted arguments contain raw sensitive value %q: %s", rawSecret, persisted)
		}
	}

	eventName := fmt.Sprint(arguments["event_name"])
	if !strings.Contains(eventName, "api_key=****9876") {
		t.Errorf("event_name does not retain masked API-key suffix: %q", eventName)
	}
	attributes := fmt.Sprint(arguments["attributes_json"])
	for _, expected := range []string{
		`"username":"alex"`,
		`"password":"[REDACTED]"`,
		`"api_key":"****9876"`,
		`"request_body":"[OMITTED]"`,
		`"account_number":"****6789"`,
		`"routing_number":"[REDACTED]"`,
		`"safe_category":"timeout"`,
	} {
		if !strings.Contains(attributes, expected) {
			t.Errorf("attributes_json = %s, want fragment %s", attributes, expected)
		}
	}
}

func TestDatabaseHandlerPreservesExplicitUsernameAndAPIKeySuffixFields(t *testing.T) {
	execution := &capturedExecution{}
	handler := NewDatabaseHandler(execution, DatabaseHandlerOptions{Application: "vantaca-api"})
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "authentication result", 0)
	record.AddAttrs(
		slog.String("db_username", "database-operator"),
		slog.String("api_key_last_four", "a-9_"),
	)

	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	arguments := namedArguments(t, execution.args)
	if got := arguments["username"]; got != "database-operator" {
		t.Fatalf("username = %#v, want database-operator", got)
	}
	if got := arguments["api_key_last_four"]; got != "a-9_" {
		t.Fatalf("api_key_last_four = %#v, want a-9_", got)
	}
}

func TestRedactingHandlerProtectsConsoleOutput(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(NewRedactingHandler(slog.NewJSONHandler(&output, nil)))

	logger.Error(
		"authentication failed",
		"username", "alex",
		"api_key", "key-ending-4321",
		"password", "do-not-log",
		"error", errors.New("Authorization: Bearer raw-bearer-token"),
	)

	logged := output.String()
	for _, secret := range []string{"key-ending-4321", "do-not-log", "raw-bearer-token"} {
		if strings.Contains(logged, secret) {
			t.Errorf("console output contains raw sensitive value %q: %s", secret, logged)
		}
	}
	for _, expected := range []string{`"username":"alex"`, `"api_key":"****4321"`, "[REDACTED]"} {
		if !strings.Contains(logged, expected) {
			t.Errorf("console output = %s, want fragment %s", logged, expected)
		}
	}
}

func TestDatabaseHandlerReportsWriteFailureWithoutMutatingRecord(t *testing.T) {
	expected := errors.New("database unavailable")
	execution := &capturedExecution{err: expected}
	var reported error
	handler := NewDatabaseHandler(execution, DatabaseHandlerOptions{
		Application: "vantaca-api",
		OnError:     func(err error) { reported = err },
	})
	record := slog.NewRecord(time.Now(), slog.LevelError, "write failed password=unsafe", 0)

	err := handler.Handle(context.Background(), record)
	if !errors.Is(err, expected) || !errors.Is(reported, expected) {
		t.Fatalf("Handle() error = %v, callback = %v, want wrapped database error", err, reported)
	}
	if record.Message != "write failed password=unsafe" {
		t.Fatalf("record message was mutated: %q", record.Message)
	}
	if persisted := fmt.Sprint(execution.args); strings.Contains(persisted, "unsafe") {
		t.Fatalf("failed execution received raw sensitive value: %s", persisted)
	}
}

func namedArguments(t *testing.T, arguments []any) map[string]any {
	t.Helper()
	result := make(map[string]any, len(arguments))
	for _, argument := range arguments {
		named, ok := argument.(sql.NamedArg)
		if !ok {
			t.Fatalf("argument type = %T, want sql.NamedArg", argument)
		}
		result[named.Name] = named.Value
	}
	return result
}
