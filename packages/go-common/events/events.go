package events

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/id"
)

type Envelope struct {
	ID            string          `json:"event_id"`
	Type          string          `json:"event_type"`
	TenantID      string          `json:"tenant_id"`
	AggregateKey  string          `json:"aggregate_key"`
	OccurredAt    time.Time       `json:"occurred_at"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	Payload       json.RawMessage `json:"payload"`
}

func New(eventType, tenantID, aggregateKey string, payload any) (Envelope, error) {
	raw, err := Canonical(payload)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{
		ID:           id.New(),
		Type:         eventType,
		TenantID:     tenantID,
		AggregateKey: aggregateKey,
		OccurredAt:   time.Now().UTC(),
		Payload:      raw,
	}, nil
}

// Canonical marshals without HTML escaping so hashes over identical values are
// stable. encoding/json sorts map keys.
func Canonical(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

func Hash(v any) (string, error) {
	raw, err := Canonical(v)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func Parse(raw []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return Envelope{}, fmt.Errorf("decode event: %w", err)
	}
	if env.ID == "" || env.Type == "" || env.TenantID == "" {
		return Envelope{}, fmt.Errorf("event envelope is incomplete")
	}
	return env, nil
}
