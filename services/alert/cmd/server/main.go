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
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/id"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/kafka"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/outbox"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/server"
	"github.com/fardanwardhana/qoder-hackathon/packages/pipeline"
	"github.com/fardanwardhana/qoder-hackathon/packages/rulesengine"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"
)

type api struct {
	app *app.App
}

func main() {
	ctx, cancel := server.SignalContext()
	defer cancel()
	a, err := app.New(ctx, "alert")
	if err != nil {
		slog.Error("startup_failed", "error", err.Error())
		return
	}
	defer a.Close()
	producer, err := kafka.NewProducer(a.Cfg.KafkaBrokers)
	if err != nil {
		a.Log.Error("kafka_producer_failed", "error", err.Error())
		return
	}
	defer producer.Close()
	relay := &outbox.Relay{Pool: a.Pool, Schema: "alert", Publisher: producer, Logger: a.Log, PollInterval: time.Second}
	go relay.Run(ctx)
	registerQueueDepth(a)

	s := &api{app: a}
	go func() {
		err := kafka.RunConsumer(ctx, a.Cfg.KafkaBrokers, "alert-bundler", []string{
			pipeline.Topic(pipeline.TopicClaimEvaluated),
			pipeline.Topic(pipeline.TopicSuppressionApproved),
			pipeline.Topic(pipeline.TopicSuppressionRevoked),
			pipeline.Topic(pipeline.TopicCaseDecided),
		}, a.Log, s.handleEvent)
		if err != nil && ctx.Err() == nil {
			a.Log.Error("consumer_stopped", "error", err.Error())
		}
	}()

	verifier := a.Verifier()
	a.Router.Route("/v1", func(r chi.Router) {
		r.Use(auth.Middleware(verifier))
		r.Get("/alerts", s.queue)
		r.Get("/alerts/{alertID}", s.getAlert)
		r.Get("/alerts/by-claim/{claimID}", s.getByClaim)
	})
	if err := a.Run(ctx); err != nil {
		a.Log.Error("server_failed", "error", err.Error())
	}
}

func (s *api) handleEvent(ctx context.Context, env events.Envelope) error {
	var err error
	switch env.Type {
	case pipeline.TopicClaimEvaluated:
		err = s.handleEvaluated(ctx, env)
	case pipeline.TopicSuppressionApproved:
		err = s.handleSuppressionApproved(ctx, env)
	case pipeline.TopicSuppressionRevoked:
		err = s.handleSuppressionRevoked(ctx, env)
	case pipeline.TopicCaseDecided:
		err = s.handleCaseDecided(ctx, env)
	}
	if err != nil {
		s.app.Log.Error("alert_event_failed", "event_id", env.ID, "type", env.Type, "error", err.Error())
	}
	return nil
}

func (s *api) handleEvaluated(ctx context.Context, env events.Envelope) error {
	var ev pipeline.ClaimEvaluated
	if err := json.Unmarshal(env.Payload, &ev); err != nil {
		return err
	}
	if len(ev.Flags) == 0 {
		return s.markProcessed(ctx, env, "alert-bundler")
	}
	alertID := alertID(env.TenantID, ev.ClaimID)
	created := env.OccurredAt.UTC()
	sev := highestSeverity(ev.Flags)
	due := slaDue(created, sev)
	flagsRaw, err := events.Canonical(ev.Flags)
	if err != nil {
		return err
	}
	suppressed := false
	err = db.WithTenant(ctx, s.app.Pool, env.TenantID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `INSERT INTO alert.processed_event (consumer, event_id)
			VALUES ('alert-bundler',$1) ON CONFLICT DO NOTHING`, env.ID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return nil
		}
		suppressionID, isSuppressed, err := activeSuppression(ctx, tx, env.TenantID, ev.Fingerprint)
		if err != nil {
			return err
		}
		suppressed = isSuppressed
		if _, err := tx.Exec(ctx, `INSERT INTO alert.bundle
			(alert_id, tenant_id, claim_id, batch_id, loss_type, severity, severity_rank, flag_count,
			 claimed_amount, fingerprint, flags, status, suppressed, suppression_id, created_at, sla_due_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'OPEN',$12,$13,$14,$15)
			ON CONFLICT (tenant_id, claim_id) DO UPDATE SET
				batch_id=EXCLUDED.batch_id, loss_type=EXCLUDED.loss_type, severity=EXCLUDED.severity,
				severity_rank=EXCLUDED.severity_rank, flag_count=EXCLUDED.flag_count,
				claimed_amount=EXCLUDED.claimed_amount, fingerprint=EXCLUDED.fingerprint,
				flags=EXCLUDED.flags, suppressed=EXCLUDED.suppressed, suppression_id=EXCLUDED.suppression_id,
				sla_due_at=LEAST(alert.bundle.sla_due_at, EXCLUDED.sla_due_at)`,
			alertID, env.TenantID, ev.ClaimID, ev.BatchID, ev.LossType, string(sev),
			rulesengine.SeverityRank(sev), len(ev.Flags), ev.ClaimedAmount, ev.Fingerprint, flagsRaw,
			suppressed, suppressionID, created, due); err != nil {
			return err
		}
		out := pipeline.AlertCreated{
			AlertID: alertID, ClaimID: ev.ClaimID, BatchID: ev.BatchID, LossType: ev.LossType,
			Severity: sev, SeverityRank: rulesengine.SeverityRank(sev), FlagCount: len(ev.Flags),
			ClaimedAmount: ev.ClaimedAmount, Fingerprint: ev.Fingerprint, Flags: ev.Flags,
			Suppressed: suppressed, SLADueAt: due, CreatedAt: created, Actor: "system",
		}
		if suppressionID != nil {
			out.SuppressionID = *suppressionID
		}
		return outbox.Add(ctx, tx, "alert", outbox.Event{
			ID: id.New(), TenantID: env.TenantID, Type: pipeline.TopicAlertCreated,
			AggregateKey: alertID, Payload: out,
		})
	})
	if err == nil {
		s.app.Log.Info("alert_bundled", "alert_id", alertID, "claim_id", ev.ClaimID,
			"severity", sev, "flags", len(ev.Flags), "suppressed", suppressed)
	}
	return err
}

func (s *api) markProcessed(ctx context.Context, env events.Envelope, consumer string) error {
	return db.WithTenant(ctx, s.app.Pool, env.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO alert.processed_event (consumer, event_id)
			VALUES ($1,$2) ON CONFLICT DO NOTHING`, consumer, env.ID)
		return err
	})
}

func activeSuppression(ctx context.Context, tx pgx.Tx, tenant, fingerprint string) (*string, bool, error) {
	var sid string
	err := tx.QueryRow(ctx, `SELECT suppression_id::text FROM alert.suppression_projection
		WHERE tenant_id=$1 AND fingerprint=$2 AND revoked_at IS NULL AND expires_at > now()`, tenant, fingerprint).Scan(&sid)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &sid, true, nil
}

func (s *api) handleSuppressionApproved(ctx context.Context, env events.Envelope) error {
	var ev pipeline.SuppressionApproved
	if err := json.Unmarshal(env.Payload, &ev); err != nil {
		return err
	}
	return db.WithTenant(ctx, s.app.Pool, env.TenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO alert.suppression_projection
			(tenant_id, fingerprint, suppression_id, expires_at) VALUES ($1,$2,$3,$4)
			ON CONFLICT (tenant_id, fingerprint) DO UPDATE SET
				suppression_id=EXCLUDED.suppression_id, expires_at=EXCLUDED.expires_at, revoked_at=NULL`,
			env.TenantID, ev.Fingerprint, ev.SuppressionID, ev.ExpiresAt); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `UPDATE alert.bundle SET suppressed=true, suppression_id=$3
			WHERE tenant_id=$1 AND fingerprint=$2`, env.TenantID, ev.Fingerprint, ev.SuppressionID)
		return err
	})
}

func (s *api) handleSuppressionRevoked(ctx context.Context, env events.Envelope) error {
	var ev pipeline.SuppressionRevoked
	if err := json.Unmarshal(env.Payload, &ev); err != nil {
		return err
	}
	return db.WithTenant(ctx, s.app.Pool, env.TenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `UPDATE alert.suppression_projection SET revoked_at=$3
			WHERE tenant_id=$1 AND suppression_id=$2`, env.TenantID, ev.SuppressionID, ev.RevokedAt); err != nil {
			return err
		}
		sid, active, err := activeSuppression(ctx, tx, env.TenantID, ev.Fingerprint)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE alert.bundle SET suppressed=$3, suppression_id=$4
			WHERE tenant_id=$1 AND fingerprint=$2`, env.TenantID, ev.Fingerprint, active, sid)
		return err
	})
}

func (s *api) handleCaseDecided(ctx context.Context, env events.Envelope) error {
	var ev pipeline.CaseDecided
	if err := json.Unmarshal(env.Payload, &ev); err != nil {
		return err
	}
	if ev.Outcome != "CONFIRMED" && ev.Outcome != "CLEARED" {
		return nil
	}
	return db.WithTenant(ctx, s.app.Pool, env.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE alert.bundle SET status='CLOSED'
			WHERE tenant_id=$1 AND claim_id = ANY($2)`, env.TenantID, ev.ClaimIDs)
		return err
	})
}

type alertRow struct {
	AlertID       string               `json:"alert_id"`
	ClaimID       string               `json:"claim_id"`
	BatchID       string               `json:"batch_id"`
	LossType      string               `json:"loss_type"`
	Severity      rulesengine.Severity `json:"severity"`
	SeverityRank  int                  `json:"severity_rank"`
	FlagCount     int                  `json:"flag_count"`
	ClaimedAmount int64                `json:"claimed_amount"`
	Fingerprint   string               `json:"fingerprint"`
	Flags         []rulesengine.Flag   `json:"flags"`
	Status        string               `json:"status"`
	Suppressed    bool                 `json:"suppressed"`
	SuppressionID *string              `json:"suppression_id,omitempty"`
	SLADueAt      time.Time            `json:"sla_due_at"`
	SLABreached   bool                 `json:"sla_breached"`
	CreatedAt     time.Time            `json:"created_at"`
}

func (s *api) queue(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	q := r.URL.Query()
	status := strings.ToUpper(strings.TrimSpace(q.Get("status")))
	severity := strings.TrimSpace(q.Get("severity"))
	lob := strings.ToUpper(strings.TrimSpace(q.Get("lob")))
	sla := strings.ToLower(strings.TrimSpace(q.Get("sla")))
	includeSuppressed, _ := strconv.ParseBool(q.Get("include_suppressed"))
	limit := 100
	if n, err := strconv.Atoi(q.Get("limit")); err == nil && n > 0 && n <= 500 {
		limit = n
	}
	var out []alertRow
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT b.alert_id, b.claim_id, b.batch_id::text, b.loss_type, b.severity,
			b.severity_rank, b.flag_count, b.claimed_amount, b.fingerprint, b.flags, b.status,
			EXISTS (SELECT 1 FROM alert.suppression_projection sp WHERE sp.tenant_id=b.tenant_id
				AND sp.fingerprint=b.fingerprint AND sp.revoked_at IS NULL AND sp.expires_at > now()) AS effective_suppressed,
			b.sla_due_at, (b.status='OPEN' AND now() > b.sla_due_at) AS breached, b.created_at
			FROM alert.bundle b
			WHERE b.tenant_id=$1
			  AND ($2='' OR b.status=$2)
			  AND ($3='' OR b.severity=$3)
			  AND ($4='' OR b.loss_type=$4)
			  AND ($5='' OR ($5='breached' AND b.status='OPEN' AND b.sla_due_at < now())
			             OR ($5='ok' AND b.sla_due_at >= now()))
			  AND ($6 OR NOT EXISTS (SELECT 1 FROM alert.suppression_projection sp
			        WHERE sp.tenant_id=b.tenant_id AND sp.fingerprint=b.fingerprint
			          AND sp.revoked_at IS NULL AND sp.expires_at > now()))
			ORDER BY b.severity_rank, b.flag_count DESC, b.claimed_amount DESC, b.alert_id
			LIMIT $7`, p.TenantID, status, severity, lob, sla, includeSuppressed, limit)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var a alertRow
			var flagsRaw []byte
			if err := rows.Scan(&a.AlertID, &a.ClaimID, &a.BatchID, &a.LossType, &a.Severity,
				&a.SeverityRank, &a.FlagCount, &a.ClaimedAmount, &a.Fingerprint, &flagsRaw, &a.Status,
				&a.Suppressed, &a.SLADueAt, &a.SLABreached, &a.CreatedAt); err != nil {
				return err
			}
			_ = json.Unmarshal(flagsRaw, &a.Flags)
			out = append(out, a)
		}
		return rows.Err()
	})
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not load alert queue")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"alerts": out, "count": len(out)})
}

func (s *api) getAlert(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	a, err := s.loadAlert(r.Context(), p.TenantID, "alert_id", chi.URLParam(r, "alertID"))
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteProblem(w, r, http.StatusNotFound, "NOT_FOUND", "alert not found")
		return
	}
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not load alert")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, a)
}

func (s *api) getByClaim(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	a, err := s.loadAlert(r.Context(), p.TenantID, "claim_id", chi.URLParam(r, "claimID"))
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteProblem(w, r, http.StatusNotFound, "NOT_FOUND", "alert not found")
		return
	}
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not load alert")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, a)
}

func (s *api) loadAlert(ctx context.Context, tenant, column, value string) (alertRow, error) {
	var a alertRow
	err := db.WithTenant(ctx, s.app.Pool, tenant, func(tx pgx.Tx) error {
		var flagsRaw []byte
		query := `SELECT b.alert_id, b.claim_id, b.batch_id::text, b.loss_type, b.severity, b.severity_rank,
			b.flag_count, b.claimed_amount, b.fingerprint, b.flags, b.status,
			EXISTS (SELECT 1 FROM alert.suppression_projection sp WHERE sp.tenant_id=b.tenant_id
				AND sp.fingerprint=b.fingerprint AND sp.revoked_at IS NULL AND sp.expires_at > now()),
			b.sla_due_at, (b.status='OPEN' AND now() > b.sla_due_at), b.created_at
			FROM alert.bundle b WHERE b.tenant_id=$1 AND b.` + column + `=$2`
		if err := tx.QueryRow(ctx, query, tenant, value).Scan(&a.AlertID, &a.ClaimID, &a.BatchID, &a.LossType,
			&a.Severity, &a.SeverityRank, &a.FlagCount, &a.ClaimedAmount, &a.Fingerprint, &flagsRaw,
			&a.Status, &a.Suppressed, &a.SLADueAt, &a.SLABreached, &a.CreatedAt); err != nil {
			return err
		}
		return json.Unmarshal(flagsRaw, &a.Flags)
	})
	return a, err
}

func highestSeverity(flags []rulesengine.Flag) rulesengine.Severity {
	best := rulesengine.SeverityLow
	for _, f := range flags {
		if rulesengine.SeverityRank(f.Severity) < rulesengine.SeverityRank(best) {
			best = f.Severity
		}
	}
	return best
}

func alertID(tenant, claimID string) string {
	sum := sha256.Sum256([]byte("v1|" + tenant + "|" + claimID))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func registerQueueDepth(a *app.App) {
	gauge := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "claimshield_queue_depth",
		Help:        "Open review items by queue.",
		ConstLabels: prometheus.Labels{"queue": "alerts_open"},
	}, func() float64 {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var n float64
		err := db.WithTenant(ctx, a.Pool, a.Cfg.TenantID, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `SELECT count(*) FROM alert.bundle
				WHERE tenant_id=$1 AND status='OPEN' AND suppressed=false`, a.Cfg.TenantID).Scan(&n)
		})
		if err != nil {
			a.Log.Warn("queue_depth_failed", "error", err.Error())
			return 0
		}
		return n
	})
	if err := a.Metrics.Register(gauge); err != nil {
		a.Log.Warn("queue_depth_register_failed", "error", err.Error())
	}
}

func slaDue(created time.Time, sev rulesengine.Severity) time.Time {
	if sev == rulesengine.SeverityHigh {
		return created.Add(24 * time.Hour).UTC()
	}
	days := 3
	if sev == rulesengine.SeverityLow {
		days = 7
	}
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.UTC
	}
	local := created.In(loc)
	added := 0
	for added < days {
		local = local.AddDate(0, 0, 1)
		if local.Weekday() != time.Saturday && local.Weekday() != time.Sunday {
			added++
		}
	}
	return local.UTC()
}
