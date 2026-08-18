package database

import (
	"testing"
	"time"

	"vantaca-interview-project/Demo/api/internal/domain"
)

func TestTransactionsEqualUsesSQLServerTimestampPrecision(t *testing.T) {
	base := time.Date(2026, time.August, 16, 12, 0, 0, 123456700, time.UTC)
	stored := []domain.Transaction{{ID: "txn_1", Amount: domain.Money(12550), Currency: "USD", Description: "deposit", PostedAt: base}}
	incoming := []domain.Transaction{{ID: "txn_1", Amount: domain.Money(12550), Currency: "USD", Description: "deposit", PostedAt: base.Add(99 * time.Nanosecond)}}

	if !transactionsEqual(stored, incoming) {
		t.Fatal("timestamps that truncate to the same DATETIMEOFFSET(7) value should match")
	}

	incoming[0].PostedAt = base.Add(100 * time.Nanosecond)
	if transactionsEqual(stored, incoming) {
		t.Fatal("timestamps in different DATETIMEOFFSET(7) increments should differ")
	}
}
