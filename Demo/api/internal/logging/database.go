package logging

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

const maxPersistedAttributesBytes = 32 << 10

type Executor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type DatabaseHandlerOptions struct {
	Application  string
	MinimumLevel slog.Level
	WriteTimeout time.Duration
	OnError      func(error)
}

// DatabaseHandler synchronously persists a sanitized record. A bounded write
// makes persistence deterministic without allowing an unavailable logging
// table to hang the application request indefinitely.
type DatabaseHandler struct {
	executor Executor
	options  DatabaseHandlerOptions
	attrs    []slog.Attr
	groups   []string
}

func NewDatabaseHandler(executor Executor, options DatabaseHandlerOptions) *DatabaseHandler {
	if options.Application == "" {
		options.Application = "application"
	}
	if options.WriteTimeout <= 0 {
		options.WriteTimeout = 2 * time.Second
	}
	return &DatabaseHandler{executor: executor, options: options}
}

func (h *DatabaseHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.options.MinimumLevel
}

func (h *DatabaseHandler) Handle(ctx context.Context, record slog.Record) error {
	attributes := make(map[string]any)
	collectAttrs(attributes, h.attrs)

	recordAttrs := make([]slog.Attr, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		recordAttrs = append(recordAttrs, attr)
		return true
	})
	collectAttrs(attributes, wrapGroups(recordAttrs, h.groups))

	correlationID := limitedString(findAttribute(attributes, func(key string) bool {
		return key == "correlationid"
	}), 64)
	username := limitedString(findAttribute(attributes, isUsernameKey), 128)
	apiKeyLastFour := lastFourFromMasked(findAttribute(attributes, isAPIKey))

	encoded, err := json.Marshal(attributes)
	if err != nil {
		return h.report(fmt.Errorf("encode sanitized log attributes: %w", err))
	}
	if len(encoded) > maxPersistedAttributesBytes {
		encoded = []byte(`{"attributes_omitted":"exceeded_32_kib_limit"}`)
	}

	writeCtx := context.Background()
	if ctx != nil {
		writeCtx = context.WithoutCancel(ctx)
	}
	writeCtx, cancel := context.WithTimeout(writeCtx, h.options.WriteTimeout)
	defer cancel()

	_, err = h.executor.ExecContext(writeCtx, `
INSERT INTO dbo.application_logs (
    occurred_at,
    application_name,
    severity,
    event_name,
    correlation_id,
    username,
    api_key_last_four,
    attributes_json
)
VALUES (
    @occurred_at,
    @application_name,
    @severity,
    @event_name,
    @correlation_id,
    @username,
    @api_key_last_four,
    @attributes_json
)`,
		sql.Named("occurred_at", record.Time.UTC()),
		sql.Named("application_name", limitedString(h.options.Application, 64)),
		sql.Named("severity", severityName(record.Level)),
		sql.Named("event_name", limitedString(sanitizeText(record.Message), 256)),
		sql.Named("correlation_id", nullable(correlationID)),
		sql.Named("username", nullable(username)),
		sql.Named("api_key_last_four", nullable(apiKeyLastFour)),
		sql.Named("attributes_json", string(encoded)),
	)
	if err != nil {
		return h.report(fmt.Errorf("persist sanitized application log: %w", err))
	}
	return nil
}

func (h *DatabaseHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := h.clone()
	clone.attrs = append(clone.attrs, wrapGroups(attrs, clone.groups)...)
	return clone
}

func (h *DatabaseHandler) WithGroup(name string) slog.Handler {
	clone := h.clone()
	if name != "" {
		clone.groups = append(clone.groups, name)
	}
	return clone
}

func (h *DatabaseHandler) clone() *DatabaseHandler {
	clone := *h
	clone.attrs = append([]slog.Attr(nil), h.attrs...)
	clone.groups = append([]string(nil), h.groups...)
	return &clone
}

func (h *DatabaseHandler) report(err error) error {
	if h.options.OnError != nil {
		h.options.OnError(err)
	}
	return err
}

func wrapGroups(attrs []slog.Attr, groups []string) []slog.Attr {
	wrapped := append([]slog.Attr(nil), attrs...)
	for index := len(groups) - 1; index >= 0; index-- {
		wrapped = []slog.Attr{slog.Group(groups[index], attrsToAny(wrapped)...)}
	}
	return wrapped
}

func collectAttrs(target map[string]any, attrs []slog.Attr) {
	for _, raw := range attrs {
		attr := sanitizeAttr(raw)
		if attr.Equal(slog.Attr{}) {
			continue
		}
		if attr.Value.Kind() == slog.KindGroup {
			children := make(map[string]any)
			collectAttrs(children, attr.Value.Group())
			if attr.Key == "" {
				for key, value := range children {
					target[key] = value
				}
				continue
			}
			if existing, ok := target[attr.Key].(map[string]any); ok {
				for key, value := range children {
					existing[key] = value
				}
			} else {
				target[attr.Key] = children
			}
			continue
		}
		target[attr.Key] = attr.Value.Any()
	}
}

func findAttribute(attributes map[string]any, matches func(string) bool) string {
	keys := make([]string, 0, len(attributes))
	for key := range attributes {
		keys = append(keys, key)
	}
	// A sorted walk makes duplicate nested matches deterministic.
	sortStrings(keys)
	for _, key := range keys {
		value := attributes[key]
		canonical := canonicalKey(key)
		if matches(canonical) {
			return fmt.Sprint(value)
		}
		if nested, ok := value.(map[string]any); ok {
			if found := findAttribute(nested, matches); found != "" {
				return found
			}
		}
	}
	return ""
}

func sortStrings(values []string) {
	for index := 1; index < len(values); index++ {
		for cursor := index; cursor > 0 && values[cursor] < values[cursor-1]; cursor-- {
			values[cursor], values[cursor-1] = values[cursor-1], values[cursor]
		}
	}
}

func lastFourFromMasked(value string) string {
	if len(value) == 4 {
		return value
	}
	if strings.HasPrefix(value, "****") && len(value) == 8 {
		return value[4:]
	}
	return ""
}

func severityName(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return "ERROR"
	case level >= slog.LevelWarn:
		return "WARN"
	case level >= slog.LevelInfo:
		return "INFO"
	default:
		return "DEBUG"
	}
}

func limitedString(value string, maximum int) string {
	value = sanitizeText(value)
	runes := []rune(value)
	if len(runes) <= maximum {
		return value
	}
	return string(runes[:maximum])
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}

// FanoutHandler sends a record to every enabled sink. It is intentionally
// small so the console and database receive the exact same slog event.
type FanoutHandler struct {
	handlers []slog.Handler
}

func NewFanoutHandler(handlers ...slog.Handler) *FanoutHandler {
	return &FanoutHandler{handlers: append([]slog.Handler(nil), handlers...)}
}

func (h *FanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *FanoutHandler) Handle(ctx context.Context, record slog.Record) error {
	var combined error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			combined = errors.Join(combined, handler.Handle(ctx, record.Clone()))
		}
	}
	return combined
}

func (h *FanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithAttrs(attrs))
	}
	return NewFanoutHandler(handlers...)
}

func (h *FanoutHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithGroup(name))
	}
	return NewFanoutHandler(handlers...)
}
