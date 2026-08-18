package northwind

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"vantaca-interview-project/Demo/api/internal/domain"
)

func TestListAccountsInjectsQueryKeyAndMapsMoney(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("api_key") != "synthetic-key" {
			t.Errorf("api key = %q", request.URL.Query().Get("api_key"))
		}
		if request.URL.Query().Get("page") != "1" {
			t.Errorf("page = %q", request.URL.Query().Get("page"))
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`[{"id":"acc_1","account_number":"000000001234","routing_number":"021000021","type":"checking","balance":10.05,"currency":"USD","status":"open"}]`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "synthetic-key", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	accounts, err := client.ListAccounts(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 || accounts[0].Balance != domain.Money(1005) {
		t.Fatalf("accounts = %#v", accounts)
	}
}

func TestSafeReadRetriesThrottle(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 3 {
			writer.Header().Set("Retry-After", "0")
			writer.WriteHeader(http.StatusServiceUnavailable)
			_, _ = writer.Write([]byte(`{"error":"service_unavailable"}`))
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte("[]"))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "synthetic-key", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListAccounts(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("calls = %d, want 3", got)
	}
}

func TestTransferTransportFailureIsAmbiguousAndNotRetried(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls.Add(1)
		hijacker, ok := writer.(http.Hijacker)
		if !ok {
			t.Fatal("test server does not support hijacking")
		}
		connection, _, err := hijacker.Hijack()
		if err != nil {
			t.Fatal(err)
		}
		_ = connection.Close()
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "synthetic-key", 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.CreateTransfer(context.Background(), CreateTransferRequest{
		FromAccountNumber: "000000001234",
		ToAccountNumber:   "000000005678",
		RoutingNumber:     "021000021",
		Amount:            25000,
		Currency:          "USD",
	}, "")
	if !IsAmbiguous(err) {
		t.Fatalf("error = %v, want ambiguous", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("calls = %d, want one transfer attempt", got)
	}
}
