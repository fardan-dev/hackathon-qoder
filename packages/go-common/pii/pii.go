package pii

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

const (
	HMACPrefix   = "hmac-sha256:"
	SHA256Prefix = "sha256:"
)

var (
	nikRe   = regexp.MustCompile(`\b\d{16}\b`)
	phoneRe = regexp.MustCompile(`\+?\b\d{10,15}\b`)
	emailRe = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
)

// Hash produces a tenant-salted HMAC-SHA256. The salt is per deployment (or
// per tenant in production) and must never be committed to the repository.
func Hash(tenantSalt, value string) string {
	mac := hmac.New(sha256.New, []byte(tenantSalt))
	mac.Write([]byte(strings.TrimSpace(value)))
	return HMACPrefix + hex.EncodeToString(mac.Sum(nil))
}

// TokenHash hashes a high-entropy session token before storage. No salt is
// needed because the token itself is random.
func TokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return SHA256Prefix + hex.EncodeToString(sum[:])
}

func IsHashed(v string) bool {
	return strings.HasPrefix(v, SHA256Prefix) || strings.HasPrefix(v, HMACPrefix)
}

// Fingerprint is the deterministic duplicate-claim fingerprint from rule A2:
// claimant + incident date + amount + provider. Tenant is included so values
// cannot be correlated across tenants.
func Fingerprint(tenantID, claimantID, incidentDate string, claimedAmount int64, providerID string) string {
	raw := fmt.Sprintf("v1|%s|%s|%s|%d|%s", tenantID, claimantID, incidentDate, claimedAmount, providerID)
	sum := sha256.Sum256([]byte(raw))
	return SHA256Prefix + hex.EncodeToString(sum[:])
}

// Redact removes common Indonesian PII patterns from free text. Structured
// fields should be hashed before they ever reach a log line; this is a
// defensive helper for error messages and adapter payloads.
func Redact(s string) string {
	s = emailRe.ReplaceAllString(s, "[REDACTED_EMAIL]")
	s = nikRe.ReplaceAllString(s, "[REDACTED_NIK]")
	s = phoneRe.ReplaceAllString(s, "[REDACTED_PHONE]")
	return s
}

// ContainsRawPII is used by log/export tests and guards.
func ContainsRawPII(s string) bool {
	return emailRe.MatchString(s) || nikRe.MatchString(s)
}
