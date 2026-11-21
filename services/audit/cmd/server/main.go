package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/app"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/auth"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/db"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/events"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/httpx"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/kafka"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/server"
	"github.com/fardanwardhana/qoder-hackathon/packages/pipeline"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

const genesisHash = "sha256:genesis"

var consumedTopics = []string{
	pipeline.Topic(pipeline.TopicBatchIngested),
	pipeline.Topic(pipeline.TopicClaimIngested),
	pipeline.Topic(pipeline.TopicClaimEvaluated),
	pipeline.Topic(pipeline.TopicAlertCreated),
	pipeline.Topic(pipeline.TopicCaseDecided),
	pipeline.Topic(pipeline.TopicSuppressionApproved),
	pipeline.Topic(pipeline.TopicSuppressionRevoked),
	pipeline.Topic(pipeline.TopicKillSwitchChanged),
	pipeline.Topic(pipeline.TopicBillingExported),
}

type api struct {
	app *app.App
}

func main() {
	ctx, cancel := server.SignalContext()
	defer cancel()
	a, err := app.New(ctx, "audit")
	if err != nil {
		slog.Error("startup_failed", "error", err.Error())
		return
	}
	defer a.Close()
	s := &api{app: a}
	go func() {
		err := kafka.RunConsumer(ctx, a.Cfg.KafkaBrokers, "audit-chain", consumedTopics, a.Log, s.appendEvent)
		if err != nil && ctx.Err() == nil {
			a.Log.Error("consumer_stopped", "error", err.Error())
		}
	}()

	verifier := a.Verifier()
	guard := []func(http.Handler) http.Handler{
		auth.Middleware(verifier),
		auth.RequireRoles(auth.RoleAdmin, auth.RoleSupervisor),
	}
	a.Router.With(guard...).Get("/v1/audit", s.list)
	a.Router.With(guard...).Get("/v1/audit/verify", s.verify)
	a.Router.With(guard...).Get("/v1/audit/trace/{claimID}", s.trace)
	if err := a.Run(ctx); err != nil {
		a.Log.Error("server_failed", "error", err.Error())
	}
}

func (s *api) appendEvent(ctx context.Context, env events.Envelope) error {
	var payload map[string]any
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		s.app.Log.Warn("audit_payload_invalid", "event_id", env.ID, "error", err.Error())
		return nil
	}
	actor, _ := payload["actor"].(string)
	if actor == "" {
		actor = "system"
	}
	objectType, objectID := objectOf(env.Type, payload)
	hashInput, err := events.Canonical(struct {
		EventID    string          `json:"event_id"`
		TenantID   string          `json:"tenant_id"`
		Actor      string          `json:"actor"`
		Action     string          `json:"action"`
		ObjectType string          `json:"object_type"`
		ObjectID   string          `json:"object_id"`
		OccurredAt time.Time       `json:"occurred_at"`
		Payload    json.RawMessage `json:"payload"`
	}{env.ID, env.TenantID, actor, env.Type, objectType, objectID, env.OccurredAt.UTC(), env.Payload})
	if err != nil {
		return err
	}
	return db.WithTenant(ctx, s.app.Pool, env.TenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, "audit:"+env.TenantID); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `INSERT INTO audit.processed_event (consumer, event_id)
			VALUES ('audit-chain',$1) ON CONFLICT DO NOTHING`, env.ID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return nil
		}
		prev := genesisHash
		var last string
		err = tx.QueryRow(ctx, `SELECT hash FROM audit.audit_event WHERE tenant_id=$1
			ORDER BY seq DESC LIMIT 1`, env.TenantID).Scan(&last)
		if err == nil {
			prev = last
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		sum := sha256.New()
		sum.Write([]byte(prev))
		sum.Write(hashInput)
		hash := "sha256:" + hex.EncodeToString(sum.Sum(nil))
		_, err = tx.Exec(ctx, `INSERT INTO audit.audit_event
			(event_id, tenant_id, actor, action, object_type, object_id, occurred_at, payload, prev_hash, hash)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			env.ID, env.TenantID, actor, env.Type, objectType, objectID, env.OccurredAt.UTC(),
			env.Payload, prev, hash)
		return err
	})
}

func objectOf(eventType string, payload map[string]any) (string, string) {
	str := func(key string) string {
		v, _ := payload[key].(string)
		return v
	}
	switch eventType {
	case pipeline.TopicBatchIngested:
		return "batch", str("batch_id")
	case pipeline.TopicClaimIngested, pipeline.TopicClaimEvaluated:
		return "claim", str("claim_id")
	case pipeline.TopicAlertCreated:
		return "alert", str("alert_id")
	case pipeline.TopicCaseDecided:
		return "case", str("case_id")
	case pipeline.TopicSuppressionApproved, pipeline.TopicSuppressionRevoked:
		return "suppression", str("suppression_id")
	case pipeline.TopicKillSwitchChanged:
		return "kill_switch", str("switch_id")
	case pipeline.TopicBillingExported:
		return "billing_export", str("export_id")
	}
	return "event", ""
}

type auditRow struct {
	Seq        int64           `json:"seq"`
	EventID    string          `json:"event_id"`
	Actor      string          `json:"actor"`
	Action     string          `json:"action"`
	ObjectType string          `json:"object_type"`
	ObjectID   string          `json:"object_id"`
	OccurredAt time.Time       `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload"`
	PrevHash   string          `json:"prev_hash"`
	Hash       string          `json:"hash"`
}

func (s *api) list(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	q := r.URL.Query()
	limit := 200
	if n, err := strconv.Atoi(q.Get("limit")); err == nil && n > 0 && n <= 1000 {
		limit = n
	}
	out := []auditRow{}
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT seq, event_id::text, actor, action, object_type, object_id,
			occurred_at, payload, prev_hash, hash FROM audit.audit_event
			WHERE tenant_id=$1
			  AND ($2='' OR object_type=$2)
			  AND ($3='' OR object_id=$3)
			  AND ($4='' OR action=$4)
			ORDER BY seq DESC LIMIT $5`,
			p.TenantID, strings.TrimSpace(q.Get("object_type")), strings.TrimSpace(q.Get("object_id")),
			strings.TrimSpace(q.Get("action")), limit)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var a auditRow
			if err := rows.Scan(&a.Seq, &a.EventID, &a.Actor, &a.Action, &a.ObjectType, &a.ObjectID,
				&a.OccurredAt, &a.Payload, &a.PrevHash, &a.Hash); err != nil {
				return err
			}
			out = append(out, a)
		}
		return rows.Err()
	})
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not load audit log")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"events": out, "count": len(out), "mutable": false,
	})
}

func (s *api) trace(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	claimID := chi.URLParam(r, "claimID")
	out := []auditRow{}
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT seq, event_id::text, actor, action, object_type, object_id,
			occurred_at, payload, prev_hash, hash FROM audit.audit_event
			WHERE tenant_id=$1 AND (
				(object_type='claim' AND object_id=$2)
				OR payload->>'claim_id'=$2
				OR EXISTS (SELECT 1 FROM jsonb_array_elements_text(coalesce(payload->'claim_ids','[]'::jsonb)) c
				           WHERE c=$2)
			)
			ORDER BY seq`, p.TenantID, claimID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var a auditRow
			if err := rows.Scan(&a.Seq, &a.EventID, &a.Actor, &a.Action, &a.ObjectType, &a.ObjectID,
				&a.OccurredAt, &a.Payload, &a.PrevHash, &a.Hash); err != nil {
				return err
			}
			out = append(out, a)
		}
		return rows.Err()
	})
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not trace claim")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"claim_id": claimID, "events": out, "count": len(out),
		"stages": stages(out),
	})
}

func stages(events []auditRow) []string {
	seen := map[string]bool{}
	var out []string
	for _, e := range events {
		if !seen[e.Action] {
			seen[e.Action] = true
			out = append(out, e.Action)
		}
	}
	return out
}

func (s *api) verify(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	valid := true
	checked := 0
	var brokenAt *int64
	var expectedPrev = genesisHash
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT seq, event_id::text, actor, action, object_type, object_id,
			occurred_at, payload, prev_hash, hash FROM audit.audit_event
			WHERE tenant_id=$1 ORDER BY seq`, p.TenantID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var a auditRow
			if err := rows.Scan(&a.Seq, &a.EventID, &a.Actor, &a.Action, &a.ObjectType, &a.ObjectID,
				&a.OccurredAt, &a.Payload, &a.PrevHash, &a.Hash); err != nil {
				return err
			}
			checked++
			raw, err := events.Canonical(struct {
				EventID    string          `json:"event_id"`
				TenantID   string          `json:"tenant_id"`
				Actor      string          `json:"actor"`
				Action     string          `json:"action"`
				ObjectType string          `json:"object_type"`
				ObjectID   string          `json:"object_id"`
				OccurredAt time.Time       `json:"occurred_at"`
				Payload    json.RawMessage `json:"payload"`
			}{a.EventID, p.TenantID, a.Actor, a.Action, a.ObjectType, a.ObjectID, a.OccurredAt.UTC(), a.Payload})
			if err != nil {
				return err
			}
			sum := sha256.New()
			sum.Write([]byte(expectedPrev))
			sum.Write(raw)
			want := "sha256:" + hex.EncodeToString(sum.Sum(nil))
			if a.PrevHash != expectedPrev || a.Hash != want {
				valid = false
				seq := a.Seq
				brokenAt = &seq
				return nil
			}
			expectedPrev = a.Hash
		}
		return rows.Err()
	})
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not verify audit log")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"valid": valid, "checked": checked, "broken_seq": brokenAt, "genesis": genesisHash,
	})
}
