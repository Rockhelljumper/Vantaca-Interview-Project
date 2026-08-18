package domain

import "testing"

func TestParseMoney(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw  string
		want Money
	}{
		{raw: "0", want: 0},
		{raw: "0.01", want: 1},
		{raw: "250.00", want: 25000},
		{raw: "-42.17", want: -4217},
	}

	for _, test := range tests {
		test := test
		t.Run(test.raw, func(t *testing.T) {
			t.Parallel()
			got, err := ParseMoney(test.raw)
			if err != nil {
				t.Fatalf("ParseMoney(%q): %v", test.raw, err)
			}
			if got != test.want {
				t.Fatalf("ParseMoney(%q) = %d, want %d", test.raw, got, test.want)
			}
			if got.String() == "" {
				t.Fatal("Money.String returned empty text")
			}
		})
	}
}

func TestParseMoneyRejectsUnsafeRepresentations(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"", "1.001", "1e3", "NaN", ".", "92233720368547758.08"} {
		if _, err := ParseMoney(raw); err == nil {
			t.Fatalf("ParseMoney(%q) succeeded, want error", raw)
		}
	}
}

func TestTransferTransitions(t *testing.T) {
	t.Parallel()

	if !CanTransitionTransfer(TransferPending, TransferPosted) {
		t.Fatal("PENDING -> POSTED should be allowed")
	}
	if !CanTransitionTransfer(TransferPosted, TransferReturned) {
		t.Fatal("POSTED -> RETURNED should be allowed")
	}
	if CanTransitionTransfer(TransferPosted, TransferPending) {
		t.Fatal("POSTED -> PENDING should be rejected")
	}
	if CanTransitionTransfer(TransferUnknown, TransferIntentRecorded) {
		t.Fatal("UNKNOWN -> INTENT_RECORDED should be rejected")
	}
}
