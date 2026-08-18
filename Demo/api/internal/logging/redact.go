package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	redactedValue = "[REDACTED]"
	omittedValue  = "[OMITTED]"
)

var (
	apiKeyAssignmentPattern = regexp.MustCompile(`(?i)((?:api[_-]?key|x-api-key|demo[_-]?admin[_-]?key)["']?\s*(?:=|:)\s*["']?)([^&,\s;"']+)`)
	secretAssignmentPattern = regexp.MustCompile(`(?i)((?:password|passwd|pwd|passphrase|client[_-]?secret|access[_-]?token|refresh[_-]?token|authorization|cookie|session[_-]?(?:id|token)|private[_-]?key|connection[_-]?string|dsn)["']?\s*(?:=|:)\s*["']?)([^&,\s;"']+)`)
	bearerPattern           = regexp.MustCompile(`(?i)(bearer\s+)([A-Za-z0-9._~+/=-]+)`)
	URLUserInfoPattern      = regexp.MustCompile(`(?i)([a-z][a-z0-9+.-]*://[^:/@\s]+:)([^@\s/]+)(@)`)
	accountNumberPattern    = regexp.MustCompile(`(?i)((?:account[_-]?number|card[_-]?number)["']?\s*(?:=|:)\s*["']?)([0-9 -]{5,32})`)
	routingNumberPattern    = regexp.MustCompile(`(?i)((?:routing[_-]?number)["']?\s*(?:=|:)\s*["']?)([0-9 -]{5,16})`)
)

// RedactingHandler applies the same deterministic policy to console logs that
// DatabaseHandler applies before persistence. Keeping this at the handler
// boundary prevents a missed call-site scrub from becoming a disclosure.
type RedactingHandler struct {
	next slog.Handler
}

func NewRedactingHandler(next slog.Handler) *RedactingHandler {
	return &RedactingHandler{next: next}
}

func (h *RedactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *RedactingHandler) Handle(ctx context.Context, record slog.Record) error {
	return h.next.Handle(ctx, sanitizedRecord(record))
}

func (h *RedactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	safe := make([]slog.Attr, 0, len(attrs))
	for _, attr := range attrs {
		safe = append(safe, sanitizeAttr(attr))
	}
	return &RedactingHandler{next: h.next.WithAttrs(safe)}
}

func (h *RedactingHandler) WithGroup(name string) slog.Handler {
	return &RedactingHandler{next: h.next.WithGroup(name)}
}

func sanitizedRecord(record slog.Record) slog.Record {
	safe := slog.NewRecord(record.Time, record.Level, sanitizeText(record.Message), record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		safe.AddAttrs(sanitizeAttr(attr))
		return true
	})
	return safe
}

func sanitizeAttr(attr slog.Attr) slog.Attr {
	attr.Value = attr.Value.Resolve()
	key := canonicalKey(attr.Key)

	switch {
	case isUsernameKey(key):
		return slog.String(attr.Key, sanitizeText(valueText(attr.Value)))
	case isAPIKeyLastFourKey(key):
		return slog.String(attr.Key, sanitizeLastFour(valueText(attr.Value)))
	case isAPIKey(key):
		return slog.String(attr.Key, maskLastFour(valueText(attr.Value)))
	case isAccountNumberKey(key), isCardNumberKey(key):
		return slog.String(attr.Key, maskLastFour(valueText(attr.Value)))
	case isSecretKey(key), isRoutingNumberKey(key):
		return slog.String(attr.Key, redactedValue)
	case isRawBodyKey(key):
		return slog.String(attr.Key, omittedValue)
	}

	if attr.Value.Kind() == slog.KindGroup {
		children := attr.Value.Group()
		safe := make([]slog.Attr, 0, len(children))
		for _, child := range children {
			safe = append(safe, sanitizeAttr(child))
		}
		return slog.Group(attr.Key, attrsToAny(safe)...)
	}

	return slog.Any(attr.Key, sanitizeValue(attr.Value.Any()))
}

func attrsToAny(attrs []slog.Attr) []any {
	values := make([]any, 0, len(attrs))
	for _, attr := range attrs {
		values = append(values, attr)
	}
	return values
}

func sanitizeValue(value any) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case string:
		return sanitizeText(typed)
	case error:
		return sanitizeText(typed.Error())
	case fmt.Stringer:
		return sanitizeText(typed.String())
	case time.Time, time.Duration, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, json.Number:
		return typed
	}

	reflected := reflect.ValueOf(value)
	return sanitizeReflected(reflected)
}

func sanitizeReflected(value reflect.Value) any {
	if !value.IsValid() {
		return nil
	}
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Map:
		keys := value.MapKeys()
		sort.Slice(keys, func(i, j int) bool { return fmt.Sprint(keys[i].Interface()) < fmt.Sprint(keys[j].Interface()) })
		result := make(map[string]any, len(keys))
		for _, mapKey := range keys {
			key := fmt.Sprint(mapKey.Interface())
			mapValue := value.MapIndex(mapKey)
			result[key] = sanitizedValueForKey(key, reflectedInterface(mapValue))
		}
		return result
	case reflect.Slice, reflect.Array:
		result := make([]any, value.Len())
		for index := 0; index < value.Len(); index++ {
			result[index] = sanitizeValue(reflectedInterface(value.Index(index)))
		}
		return result
	case reflect.Struct:
		encoded, err := json.Marshal(value.Interface())
		if err == nil {
			var decoded any
			decoder := json.NewDecoder(strings.NewReader(string(encoded)))
			decoder.UseNumber()
			if decoder.Decode(&decoded) == nil {
				return sanitizeValue(decoded)
			}
		}
		return sanitizeText(fmt.Sprint(value.Interface()))
	default:
		return sanitizeText(fmt.Sprint(value.Interface()))
	}
}

func reflectedInterface(value reflect.Value) any {
	if !value.IsValid() {
		return nil
	}
	if value.Kind() == reflect.Interface && !value.IsNil() {
		return value.Interface()
	}
	if value.CanInterface() {
		return value.Interface()
	}
	return redactedValue
}

func sanitizedValueForKey(key string, value any) any {
	canonical := canonicalKey(key)
	switch {
	case isUsernameKey(canonical):
		return sanitizeText(fmt.Sprint(value))
	case isAPIKeyLastFourKey(canonical):
		return sanitizeLastFour(fmt.Sprint(value))
	case isAPIKey(canonical), isAccountNumberKey(canonical), isCardNumberKey(canonical):
		return maskLastFour(fmt.Sprint(value))
	case isSecretKey(canonical), isRoutingNumberKey(canonical):
		return redactedValue
	case isRawBodyKey(canonical):
		return omittedValue
	default:
		return sanitizeValue(value)
	}
}

func sanitizeText(value string) string {
	value = URLUserInfoPattern.ReplaceAllString(value, `${1}`+redactedValue+`${3}`)
	value = bearerPattern.ReplaceAllString(value, `${1}`+redactedValue)
	value = apiKeyAssignmentPattern.ReplaceAllStringFunc(value, func(match string) string {
		parts := apiKeyAssignmentPattern.FindStringSubmatch(match)
		return parts[1] + maskLastFour(parts[2])
	})
	value = secretAssignmentPattern.ReplaceAllString(value, `${1}`+redactedValue)
	value = accountNumberPattern.ReplaceAllStringFunc(value, func(match string) string {
		parts := accountNumberPattern.FindStringSubmatch(match)
		return parts[1] + maskLastFour(strings.ReplaceAll(parts[2], " ", ""))
	})
	value = routingNumberPattern.ReplaceAllString(value, `${1}`+redactedValue)
	return value
}

func maskLastFour(value string) string {
	value = strings.TrimSpace(value)
	if value == redactedValue || value == omittedValue {
		return value
	}
	if strings.HasPrefix(value, "****") {
		suffix := strings.TrimPrefix(value, "****")
		if len(suffix) == 4 {
			return value
		}
		value = suffix
	}
	if len(value) <= 4 {
		return redactedValue
	}
	return "****" + value[len(value)-4:]
}

func sanitizeLastFour(value string) string {
	value = strings.TrimSpace(value)
	if len(value) == 4 {
		return sanitizeText(value)
	}
	masked := maskLastFour(value)
	if strings.HasPrefix(masked, "****") && len(masked) == 8 {
		return masked[4:]
	}
	return redactedValue
}

func valueText(value slog.Value) string {
	if value.Kind() == slog.KindString {
		return value.String()
	}
	return fmt.Sprint(value.Any())
}

func canonicalKey(key string) string {
	var builder strings.Builder
	for _, character := range strings.ToLower(key) {
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' {
			builder.WriteRune(character)
		}
	}
	return builder.String()
}

func isUsernameKey(key string) bool {
	return key == "username" || key == "user" || key == "dbuser" || strings.HasSuffix(key, "username")
}

func isAPIKey(key string) bool {
	return strings.Contains(key, "apikey") || key == "xapikey" || strings.Contains(key, "adminkey")
}

func isAPIKeyLastFourKey(key string) bool {
	return strings.Contains(key, "apikeylastfour") || strings.Contains(key, "adminkeylastfour")
}

func isSecretKey(key string) bool {
	return strings.Contains(key, "password") || key == "passwd" || key == "pwd" || key == "passphrase" ||
		strings.Contains(key, "secret") || strings.Contains(key, "token") || strings.Contains(key, "authorization") ||
		strings.Contains(key, "cookie") || strings.Contains(key, "credential") || strings.Contains(key, "signature") ||
		strings.Contains(key, "privatekey") || strings.Contains(key, "connectionstring") || key == "dsn"
}

func isRawBodyKey(key string) bool {
	if strings.Contains(key, "hash") || strings.Contains(key, "digest") || strings.Contains(key, "sha256") {
		return false
	}
	return key == "body" || key == "payload" || strings.Contains(key, "requestbody") ||
		strings.Contains(key, "responsebody") || strings.Contains(key, "rawbody") ||
		strings.Contains(key, "requestpayload") || strings.Contains(key, "responsepayload")
}

func isAccountNumberKey(key string) bool {
	return strings.Contains(key, "accountnumber")
}

func isCardNumberKey(key string) bool {
	return strings.Contains(key, "cardnumber") || key == "pan"
}

func isRoutingNumberKey(key string) bool {
	return strings.Contains(key, "routingnumber") || key == "ssn" || strings.Contains(key, "taxid")
}
