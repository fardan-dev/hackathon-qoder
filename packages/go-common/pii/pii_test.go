package pii

import (
	"strings"
	"testing"
)

func TestHashIsDeterministicAndTenantScoped(t *testing.T) {
	a := Hash("salt-a", "3171234567890001")
	b := Hash("salt-a", "3171234567890001")
	c := Hash("salt-b", "3171234567890001")
	if a != b {
		t.Fatalf("hash must be deterministic")
	}
	if a == c {
		t.Fatalf("hash must change with tenant salt")
	}
	if strings.Contains(a, "3171234567890001") {
		t.Fatalf("hash leaked the raw value")
	}
	if !IsHashed(a) {
		t.Fatalf("hash must be recognisable: %s", a)
	}
}

func TestFingerprintIsStable(t *testing.T) {
	a := Fingerprint("t1", "C-1", "2026-09-10", 18500000, "P-22")
	b := Fingerprint("t1", "C-1", "2026-09-10", 18500000, "P-22")
	c := Fingerprint("t1", "C-1", "2026-09-10", 18500001, "P-22")
	if a != b || a == c {
		t.Fatalf("fingerprint must be deterministic and amount sensitive")
	}
}

func TestRedact(t *testing.T) {
	in := "email budi@example.com nik 3171234567890001 phone 081234567890"
	out := Redact(in)
	for _, raw := range []string{"budi@example.com", "3171234567890001"} {
		if strings.Contains(out, raw) {
			t.Fatalf("redaction failed for %q: %s", raw, out)
		}
	}
	if ContainsRawPII(out) {
		t.Fatalf("redacted text still contains raw PII: %s", out)
	}
}

func TestTokenHash(t *testing.T) {
	h := TokenHash("token-value")
	if h == "token-value" || !IsHashed(h) {
		t.Fatalf("token hash must be opaque: %s", h)
	}
}
