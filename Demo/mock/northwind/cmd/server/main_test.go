package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoadConfigDefaults(t *testing.T) {
	clearMockEnvironment(t)

	config, address, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}

	if address != ":8081" {
		t.Fatalf("address = %q, want :8081", address)
	}
	if config.APIKey != defaultAPIKey {
		t.Fatalf("API key = %q, want synthetic default", config.APIKey)
	}
	if config.WebhookAttempts != 3 {
		t.Fatalf("webhook attempts = %d, want 3", config.WebhookAttempts)
	}
	if config.WebhookBackoff != 100*time.Millisecond {
		t.Fatalf("webhook backoff = %s, want 100ms", config.WebhookBackoff)
	}
	if config.ScenarioDelay != 5*time.Second {
		t.Fatalf("scenario delay = %s, want 5s", config.ScenarioDelay)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	clearMockEnvironment(t)
	t.Setenv("NORTHWIND_MOCK_PORT", "18081")
	t.Setenv("NORTHWIND_MOCK_API_KEY", "another_synthetic_key")
	t.Setenv("NORTHWIND_MOCK_WEBHOOK_URL", "http://127.0.0.1:9999/webhook")
	t.Setenv("NORTHWIND_MOCK_WEBHOOK_ATTEMPTS", "4")
	t.Setenv("NORTHWIND_MOCK_WEBHOOK_BACKOFF", "250ms")
	t.Setenv("NORTHWIND_MOCK_SCENARIO_DELAY", "2s")

	config, address, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}

	if address != ":18081" ||
		config.APIKey != "another_synthetic_key" ||
		config.WebhookURL != "http://127.0.0.1:9999/webhook" ||
		config.WebhookAttempts != 4 ||
		config.WebhookBackoff != 250*time.Millisecond ||
		config.ScenarioDelay != 2*time.Second {
		t.Fatalf("unexpected config/address: %+v %q", config, address)
	}
}

func TestLoadConfigRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "port text", key: "NORTHWIND_MOCK_PORT", value: "abc"},
		{name: "port range", key: "NORTHWIND_MOCK_PORT", value: "70000"},
		{name: "scenario delay", key: "NORTHWIND_MOCK_SCENARIO_DELAY", value: "0s"},
		{name: "webhook backoff", key: "NORTHWIND_MOCK_WEBHOOK_BACKOFF", value: "not-a-duration"},
		{name: "webhook attempts", key: "NORTHWIND_MOCK_WEBHOOK_ATTEMPTS", value: "0"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			clearMockEnvironment(t)
			t.Setenv(test.key, test.value)

			if _, _, err := loadConfig(); err == nil {
				t.Fatalf("loadConfig unexpectedly accepted %s=%q", test.key, test.value)
			}
		})
	}
}

func TestHealthcheckStatusValidation(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{name: "healthy", status: http.StatusOK},
		{name: "unhealthy", status: http.StatusServiceUnavailable, wantErr: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(test.status)
			}))
			defer server.Close()

			client := &http.Client{Timeout: time.Second}
			if gotErr := checkHealth(client, server.URL) != nil; gotErr != test.wantErr {
				t.Fatalf("checkHealth error = %t, want %t", gotErr, test.wantErr)
			}
		})
	}
}

func clearMockEnvironment(t *testing.T) {
	t.Helper()

	for _, name := range []string{
		"NORTHWIND_MOCK_PORT",
		"NORTHWIND_MOCK_API_KEY",
		"NORTHWIND_MOCK_WEBHOOK_URL",
		"NORTHWIND_MOCK_WEBHOOK_ATTEMPTS",
		"NORTHWIND_MOCK_WEBHOOK_BACKOFF",
		"NORTHWIND_MOCK_SCENARIO_DELAY",
	} {
		t.Setenv(name, "")
	}
}
