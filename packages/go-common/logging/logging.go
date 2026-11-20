package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// New returns a JSON logger. Log records must never contain raw PII; use
// RedactMap or opaque identifiers before logging user supplied values.
func New(service, level string) *slog.Logger {
	return NewWithWriter(os.Stdout, service, level)
}

func NewWithWriter(w io.Writer, service, level string) *slog.Logger {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: parseLevel(level),
	})
	return slog.New(h).With("service", service)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

var sensitiveKeys = map[string]struct{}{
	"nik":                {},
	"national_id":        {},
	"ktp":                {},
	"password":           {},
	"password_hash":      {},
	"token":              {},
	"session":            {},
	"session_token":      {},
	"authorization":      {},
	"bank_account":       {},
	"bank_account_hash":  {},
	"phone":              {},
	"phone_hash":         {},
	"email":              {},
	"email_hash":         {},
	"address":            {},
	"address_hash":       {},
	"raw_nik":            {},
}

// RedactMap replaces values of known sensitive keys with [REDACTED]. It is a
// last line of defence: callers should avoid putting PII in log fields at all.
func RedactMap(fields map[string]any) map[string]any {
	out := make(map[string]any, len(fields))
	for k, v := range fields {
		if _, ok := sensitiveKeys[strings.ToLower(k)]; ok {
			out[k] = "[REDACTED]"
			continue
		}
		out[k] = v
	}
	return out
}

func IsSensitiveKey(key string) bool {
	_, ok := sensitiveKeys[strings.ToLower(key)]
	return ok
}
