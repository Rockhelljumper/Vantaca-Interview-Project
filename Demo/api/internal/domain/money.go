package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Money stores integer minor units. The demo currently accepts USD only, but
// currency validation remains outside this representation.
type Money int64

func ParseMoney(raw string) (Money, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.ContainsAny(raw, "eE") {
		return 0, fmt.Errorf("amount must be a plain decimal with at most two fractional digits")
	}

	negative := strings.HasPrefix(raw, "-")
	if negative {
		raw = strings.TrimPrefix(raw, "-")
	}

	parts := strings.Split(raw, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, errors.New("amount is invalid")
	}

	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
		if fraction == "" || len(fraction) > 2 {
			return 0, errors.New("amount must have at most two fractional digits")
		}
	}
	for len(fraction) < 2 {
		fraction += "0"
	}

	whole, err := strconv.ParseUint(parts[0], 10, 63)
	if err != nil {
		return 0, errors.New("amount has an invalid whole-number component")
	}

	var fractional uint64
	if fraction != "" {
		fractional, err = strconv.ParseUint(fraction, 10, 8)
		if err != nil {
			return 0, errors.New("amount has an invalid fractional component")
		}
	}
	if whole > (math.MaxInt64-fractional)/100 {
		return 0, errors.New("amount is too large")
	}

	minor := int64(whole*100 + fractional)
	if negative {
		minor = -minor
	}
	return Money(minor), nil
}

func MoneyFromJSONNumber(value json.Number) (Money, error) {
	return ParseMoney(value.String())
}

func (m Money) String() string {
	negative := m < 0
	var magnitude uint64
	if negative {
		magnitude = uint64(-(m + 1))
		magnitude++
	} else {
		magnitude = uint64(m)
	}

	value := fmt.Sprintf("%d.%02d", magnitude/100, magnitude%100)
	if negative {
		return "-" + value
	}
	return value
}

func (m Money) MarshalJSON() ([]byte, error) {
	return []byte(m.String()), nil
}

func (m *Money) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("money cannot be null")
	}
	value, err := ParseMoney(string(data))
	if err != nil {
		return err
	}
	*m = value
	return nil
}
