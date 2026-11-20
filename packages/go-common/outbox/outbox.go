package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/events"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var schemaName = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

type Publisher interface {
	Publish(ctx context.Context, topic, key string, value []byte) error
}

type Event struct {
	ID           string
	TenantID     string
	Type         string
	AggregateKey string
	Payload      any
}

// Add writes an outbox row inside the caller's transaction, keeping the
// aggregate change and the event atomic.
func Add(ctx context.Context, tx pgx.Tx, schema string, ev Event) error {
	if !schemaName.MatchString(schema) {
		return fmt.Errorf("invalid outbox schema %q", schema)
	}
	raw, err := events.Canonical(ev.Payload)
	if err != nil {
		return err
	}
	sql := fmt.Sprintf(`INSERT INTO %s.outbox_event (id, tenant_id, event_type, aggregate_key, payload)
		VALUES ($1, $2, $3, $4, $5)`, schema)
	_, err = tx.Exec(ctx, sql, ev.ID, ev.TenantID, ev.Type, ev.AggregateKey, raw)
	return err
}

type Relay struct {
	Pool         *pgxpool.Pool
	Schema       string
	Publisher    Publisher
	Logger       *slog.Logger
	PollInterval time.Duration
}

func (r *Relay) Run(ctx context.Context) {
	interval := r.PollInterval
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if err := r.once(ctx); err != nil {
			r.Logger.Error("outbox_relay_failed", "error", err.Error())
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *Relay) once(ctx context.Context) error {
	if !schemaName.MatchString(r.Schema) {
		return fmt.Errorf("invalid outbox schema %q", r.Schema)
	}
	var published []string
	err := func() error {
		tx, err := r.Pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback(ctx) }()

		query := fmt.Sprintf(`SELECT id, tenant_id, event_type, aggregate_key, payload, attempts, created_at
			FROM %s.outbox_event
			WHERE status = 'PENDING' AND next_attempt_at <= now()
			ORDER BY created_at
			LIMIT 50
			FOR UPDATE SKIP LOCKED`, r.Schema)
		rows, err := tx.Query(ctx, query)
		if err != nil {
			return err
		}
		type pending struct {
			id, tenant, eventType, key string
			payload                    []byte
			attempts                   int
			createdAt                  time.Time
		}
		var items []pending
		for rows.Next() {
			var p pending
			if err := rows.Scan(&p.id, &p.tenant, &p.eventType, &p.key, &p.payload, &p.attempts, &p.createdAt); err != nil {
				rows.Close()
				return err
			}
			items = append(items, p)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}

		update := fmt.Sprintf(`UPDATE %s.outbox_event
			SET status=$2, published_at=$3, attempts=$4, next_attempt_at=$5
			WHERE id=$1`, r.Schema)
		for _, p := range items {
			env := events.Envelope{
				ID:           p.id,
				Type:         p.eventType,
				TenantID:     p.tenant,
				AggregateKey: p.key,
				OccurredAt:   p.createdAt.UTC(),
				Payload:      p.payload,
			}
			raw, err := events.Canonical(env)
			if err != nil {
				return err
			}
			if err := r.Publisher.Publish(ctx, topicFor(p.eventType), p.key, raw); err != nil {
				attempts := p.attempts + 1
				status := "PENDING"
				if attempts >= 10 {
					status = "FAILED"
				}
				backoff := time.Duration(1<<min(attempts, 6)) * time.Second
				_, _ = tx.Exec(ctx, update, p.id, status, nil, attempts, time.Now().Add(backoff))
				r.Logger.Warn("outbox_publish_failed", "event_id", p.id, "type", p.eventType, "attempts", attempts, "error", err.Error())
				continue
			}
			if _, err := tx.Exec(ctx, update, p.id, "PUBLISHED", time.Now().UTC(), p.attempts, time.Now().UTC()); err != nil {
				return err
			}
			published = append(published, p.id)
		}
		return tx.Commit(ctx)
	}()
	if err == nil && len(published) > 0 {
		r.Logger.Info("outbox_published", "count", len(published))
	}
	return err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func topicFor(eventType string) string {
	suffix := os.Getenv("KAFKA_TOPIC_SUFFIX")
	if suffix == "" {
		return eventType
	}
	if !strings.HasPrefix(suffix, "-") {
		suffix = "-" + suffix
	}
	return eventType + suffix
}
