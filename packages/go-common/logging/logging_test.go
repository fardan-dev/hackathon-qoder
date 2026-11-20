package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestRedactMap(t *testing.T) {
	out := RedactMap(map[string]any{
		"nik":        "3171234567890001",
		"claim_id":   "CLM-1",
		"email":      "a@b.c",
		"note":       "ok",
		"password":   "secret",
		"bank_account": "123",
	})
	for _, key := range []string{"nik", "email", "password", "bank_account"} {
		if out[key] != "[REDACTED]" {
			t.Fatalf("%s was not redacted: %v", key, out[key])
		}
	}
	if out["claim_id"] != "CLM-1" || out["note"] != "ok" {
		t.Fatalf("non-sensitive values must be preserved: %v", out)
	}
}

func TestLoggerNeverWritesRawPIIWhenUsingRedact(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithWriter(&buf, "test", "info")
	logger.Info("login", "fields", RedactMap(map[string]any{"nik": "3171234567890001"}))
	if strings.Contains(buf.String(), "3171234567890001") {
		t.Fatalf("raw NIK leaked into log: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "[REDACTED]") {
		t.Fatalf("expected redaction marker, got %s", buf.String())
	}
	_ = slog.LevelInfo
}
