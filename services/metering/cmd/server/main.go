package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
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
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	formulaUsage   = "USAGE-v1"
	formulaValue   = "AVS-v1"
	formulaBilling = "BILLING-v1"
)

var idNamespace = uuid.MustParse("6f2f1c1e-7a0e-4c4e-9a3e-0b6a7c8d9e10")

type api struct {
	app *app.App
}

func main() {
	ctx, cancel := server.SignalContext()
	defer cancel()
	a, err := app.New(ctx, "metering")
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
	relay := &outbox.Relay{Pool: a.Pool, Schema: "metering", Publisher: producer, Logger: a.Log, PollInterval: time.Second}
	go relay.Run(ctx)

	s := &api{app: a}
	go func() {
		err := kafka.RunConsumer(ctx, a.Cfg.KafkaBrokers, "metering-usage", []string{
			pipeline.Topic(pipeline.TopicClaimEvaluated),
			pipeline.Topic(pipeline.TopicAlertCreated),
			pipeline.Topic(pipeline.TopicCaseDecided),
		}, a.Log, s.handleEvent)
		if err != nil && ctx.Err() == nil {
			a.Log.Error("consumer_stopped", "error", err.Error())
		}
	}()

	verifier := a.Verifier()
	a.Router.Route("/v1/metering", func(r chi.Router) {
		r.Use(auth.Middleware(verifier))
		r.Get("/usage", s.usage)
		r.Get("/close/{period}", s.getClose)
		r.With(auth.RequireRoles(auth.RoleAdmin)).Post("/close", s.closePeriod)
		r.Get("/value-report/{period}", s.getValueReport)
		r.With(auth.RequireRoles(auth.RoleAdmin, auth.RoleSupervisor)).Post("/value-report", s.generateValueReport)
		r.With(auth.RequireRoles(auth.RoleAdmin)).Get("/billing-export/{period}", s.getBillingExport)
		r.With(auth.RequireRoles(auth.RoleAdmin)).Post("/billing-export", s.createBillingExport)
	})
	if err := a.Run(ctx); err != nil {
		a.Log.Error("server_failed", "error", err.Error())
	}
}

func (s *api) handleEvent(ctx context.Context, env events.Envelope) error {
	var metric string
	switch env.Type {
	case pipeline.TopicClaimEvaluated:
		metric = "claims_evaluated"
	case pipeline.TopicAlertCreated:
		metric = "alerts_created"
	case pipeline.TopicCaseDecided:
		metric = "decisions_recorded"
	default:
		return nil
	}
	period := periodOf(env.OccurredAt)
	return db.WithTenant(ctx, s.app.Pool, env.TenantID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `INSERT INTO metering.processed_event (consumer, event_id)
			VALUES ('metering-usage',$1) ON CONFLICT DO NOTHING`, env.ID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return nil
		}
		_, err = tx.Exec(ctx, `INSERT INTO metering.usage_event
			(usage_event_id, tenant_id, source_event_id, metric, quantity, period, occurred_at, payload)
			VALUES ($1,$2,$3,$4,1,$5,$6,$7)
			ON CONFLICT (tenant_id, source_event_id, metric) DO NOTHING`,
			id.New(), env.TenantID, env.ID, metric, period, env.OccurredAt.UTC(), env.Payload)
		return err
	})
}

func periodOf(t time.Time) string {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.UTC
	}
	return t.In(loc).Format("2006-01")
}

func periodBounds(period string) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.UTC
	}
	start, err := time.ParseInLocation("2006-01", period, loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("period must use YYYY-MM")
	}
	return start.UTC(), start.AddDate(0, 1, 0).UTC(), nil
}

func (s *api) metrics(ctx context.Context, tx pgx.Tx, tenant, period string) (map[string]int64, int64, error) {
	rows, err := tx.Query(ctx, `SELECT metric, count(*), coalesce(sum(quantity),0)
		FROM metering.usage_event WHERE tenant_id=$1 AND period=$2 GROUP BY metric ORDER BY metric`,
		tenant, period)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := map[string]int64{}
	var diff int64
	for rows.Next() {
		var metric string
		var count, sum int64
		if err := rows.Scan(&metric, &count, &sum); err != nil {
			return nil, 0, err
		}
		out[metric] = sum
		if sum-count < 0 {
			diff += count - sum
		} else {
			diff += sum - count
		}
	}
	return out, diff, rows.Err()
}

type usageRecord struct {
	TenantID       string           `json:"tenant_id"`
	Period         string           `json:"period"`
	FormulaVersion string           `json:"formula_version"`
	Metrics        map[string]int64 `json:"metrics"`
	Reconciliation int64            `json:"reconciliation_diff"`
	Checksum       string           `json:"checksum"`
	ClosedAt       time.Time        `json:"closed_at"`
}

func (s *api) usage(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	period := strings.TrimSpace(r.URL.Query().Get("period"))
	if period == "" {
		period = periodOf(time.Now())
	}
	if _, _, err := periodBounds(period); err != nil {
		httpx.WriteProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION", err.Error())
		return
	}
	var metrics map[string]int64
	var diff int64
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		var err error
		metrics, diff, err = s.metrics(r.Context(), tx, p.TenantID, period)
		return err
	})
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not load usage")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"period": period, "metrics": metrics, "reconciliation_diff": diff,
	})
}

func (s *api) closePeriod(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	var req struct {
		Period string `json:"period"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "INVALID_JSON", "period is required")
		return
	}
	req.Period = strings.TrimSpace(req.Period)
	if _, _, err := periodBounds(req.Period); err != nil {
		httpx.WriteProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION", err.Error())
		return
	}
	rec, reused, err := s.closeUsage(r.Context(), p.TenantID, p.UserID, req.Period)
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not close period")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"usage_record": rec, "reused": reused})
}

func (s *api) closeUsage(ctx context.Context, tenant, userID, period string) (usageRecord, bool, error) {
	var rec usageRecord
	reused := false
	err := db.WithTenant(ctx, s.app.Pool, tenant, func(tx pgx.Tx) error {
		var metrics map[string]int64
		var diff int64
		metrics, diff, err := s.metrics(ctx, tx, tenant, period)
		if err != nil {
			return err
		}
		raw, err := events.Canonical(metrics)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(raw)
		checksum := "sha256:" + hex.EncodeToString(sum[:])
		rec = usageRecord{TenantID: tenant, Period: period, FormulaVersion: formulaUsage,
			Metrics: metrics, Reconciliation: diff, Checksum: checksum}
		var existing string
		var closedAt time.Time
		var existingMetrics []byte
		err = tx.QueryRow(ctx, `SELECT checksum, metrics, closed_at FROM metering.usage_record
			WHERE tenant_id=$1 AND period=$2`, tenant, period).Scan(&existing, &existingMetrics, &closedAt)
		if err == nil {
			reused = true
			rec.Checksum = existing
			rec.ClosedAt = closedAt
			_ = json.Unmarshal(existingMetrics, &rec.Metrics)
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		rec.ClosedAt = time.Now().UTC()
		_, err = tx.Exec(ctx, `INSERT INTO metering.usage_record
			(tenant_id, period, formula_version, metrics, checksum, closed_by, closed_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`, tenant, period, formulaUsage, raw, checksum, userID, rec.ClosedAt)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return tx.QueryRow(ctx, `SELECT checksum, metrics, closed_at FROM metering.usage_record
					WHERE tenant_id=$1 AND period=$2`, tenant, period).Scan(
					&rec.Checksum, &existingMetrics, &rec.ClosedAt)
			}
			return err
		}
		return nil
	})
	return rec, reused, err
}

func (s *api) getClose(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	period := chi.URLParam(r, "period")
	var rec usageRecord
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		var metrics []byte
		return tx.QueryRow(r.Context(), `SELECT period, formula_version, metrics, checksum, closed_at
			FROM metering.usage_record WHERE tenant_id=$1 AND period=$2`, p.TenantID, period).Scan(
			&rec.Period, &rec.FormulaVersion, &metrics, &rec.Checksum, &rec.ClosedAt)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteProblem(w, r, http.StatusNotFound, "NOT_FOUND", "period has not been closed")
		return
	}
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not load usage record")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, rec)
}

type valuePayload struct {
	Period             string `json:"period"`
	FormulaVersion     string `json:"formula_version"`
	ClaimsEvaluated    int64  `json:"claims_evaluated"`
	AlertsCreated      int64  `json:"alerts_created"`
	Decisions          int64  `json:"decisions"`
	Confirmed          int64  `json:"confirmed"`
	Cleared            int64  `json:"cleared"`
	NeedsInfo          int64  `json:"needs_info"`
	ValueProtected     int64  `json:"value_protected"`
	ValueRecovered     int64  `json:"value_recovered"`
	ValueAtRiskAvoided int64  `json:"value_at_risk_avoided"`
	FalsePositiveBPS   int64  `json:"false_positive_bps"`
}

func (s *api) generateValueReport(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	var req struct {
		Period string `json:"period"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "INVALID_JSON", "period is required")
		return
	}
	start, end, err := periodBounds(strings.TrimSpace(req.Period))
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION", err.Error())
		return
	}
	var payload valuePayload
	var checksum string
	var reportID string
	var reused bool
	err = db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		usage, _, err := s.metrics(r.Context(), tx, p.TenantID, req.Period)
		if err != nil {
			return err
		}
		payload = valuePayload{Period: req.Period, FormulaVersion: formulaValue,
			ClaimsEvaluated: usage["claims_evaluated"], AlertsCreated: usage["alerts_created"]}
		if err := tx.QueryRow(r.Context(), `SELECT count(*),
			count(*) FILTER (WHERE outcome='CONFIRMED'),
			count(*) FILTER (WHERE outcome='CLEARED'),
			count(*) FILTER (WHERE outcome='NEEDS_INFO'),
			coalesce(sum(value_protected),0), coalesce(sum(value_recovered),0)
			FROM reporting.decision_snapshot
			WHERE tenant_id=$1 AND decided_at >= $2 AND decided_at < $3`,
			p.TenantID, start, end).Scan(&payload.Decisions, &payload.Confirmed, &payload.Cleared,
			&payload.NeedsInfo, &payload.ValueProtected, &payload.ValueRecovered); err != nil {
			return err
		}
		payload.ValueAtRiskAvoided = payload.ValueProtected + payload.ValueRecovered
		if payload.Decisions > 0 {
			payload.FalsePositiveBPS = payload.Cleared * 10000 / payload.Decisions
		}
		raw, err := events.Canonical(payload)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(raw)
		checksum = "sha256:" + hex.EncodeToString(sum[:])
		reportID = uuid.NewSHA1(idNamespace, []byte(p.TenantID+"|"+req.Period+"|"+formulaValue)).String()
		tag, err := tx.Exec(r.Context(), `INSERT INTO metering.value_report
			(report_id, tenant_id, period, formula_version, payload, checksum)
			VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT (report_id) DO NOTHING`,
			reportID, p.TenantID, req.Period, formulaValue, raw, checksum)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			reused = true
			var storedRaw []byte
			if err := tx.QueryRow(r.Context(), `SELECT checksum, payload FROM metering.value_report WHERE report_id=$1`,
				reportID).Scan(&checksum, &storedRaw); err != nil {
				return err
			}
			return json.Unmarshal(storedRaw, &payload)
		}
		return nil
	})
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not generate value report")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"report_id": reportID, "formula_version": formulaValue, "checksum": checksum,
		"reused": reused, "payload": payload,
	})
}

func (s *api) getValueReport(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	period := chi.URLParam(r, "period")
	var reportID, checksum string
	var raw []byte
	var generatedAt time.Time
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `SELECT report_id::text, checksum, payload, generated_at
			FROM metering.value_report WHERE tenant_id=$1 AND period=$2
			ORDER BY generated_at DESC LIMIT 1`, p.TenantID, period).Scan(&reportID, &checksum, &raw, &generatedAt)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteProblem(w, r, http.StatusNotFound, "NOT_FOUND", "value report not found")
		return
	}
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not load value report")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Report-Id", reportID)
	w.Header().Set("X-Checksum", checksum)
	w.Header().Set("X-Generated-At", generatedAt.UTC().Format(time.RFC3339))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

type billingPayload struct {
	Period         string           `json:"period"`
	Version        string           `json:"version"`
	FormulaVersion string           `json:"formula_version"`
	Currency       string           `json:"currency"`
	Metrics        map[string]int64 `json:"metrics"`
	TotalEvents    int64            `json:"total_events"`
}

func (s *api) createBillingExport(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	var req struct {
		Period  string `json:"period"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "INVALID_JSON", "period is required")
		return
	}
	req.Period = strings.TrimSpace(req.Period)
	req.Version = strings.TrimSpace(req.Version)
	if req.Version == "" {
		req.Version = "v1"
	}
	if _, _, err := periodBounds(req.Period); err != nil {
		httpx.WriteProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION", err.Error())
		return
	}
	payload, checksum, exportID, createdAt, reused, err := s.buildBilling(r.Context(), p.TenantID, p.Email, req.Period, req.Version)
	if err != nil {
		writeBillingError(w, r, err)
		return
	}
	raw, _ := events.Canonical(payload)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Export-Id", exportID)
	w.Header().Set("X-Checksum", checksum)
	w.Header().Set("X-Reused", fmt.Sprintf("%t", reused))
	w.Header().Set("X-Created-At", createdAt.UTC().Format(time.RFC3339))
	status := http.StatusCreated
	if reused {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}

func (s *api) buildBilling(ctx context.Context, tenant, actor, period, version string) (billingPayload, string, string, time.Time, bool, error) {
	var payload billingPayload
	var checksum, exportID string
	var createdAt time.Time
	reused := false
	err := db.WithTenant(ctx, s.app.Pool, tenant, func(tx pgx.Tx) error {
		var closedChecksum string
		if err := tx.QueryRow(ctx, `SELECT checksum FROM metering.usage_record
			WHERE tenant_id=$1 AND period=$2`, tenant, period).Scan(&closedChecksum); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errUnprocessable("PERIOD_NOT_CLOSED", "close the usage period before exporting billing")
			}
			return err
		}
		metrics, _, err := s.metrics(ctx, tx, tenant, period)
		if err != nil {
			return err
		}
		var total int64
		for _, v := range metrics {
			total += v
		}
		payload = billingPayload{Period: period, Version: version, FormulaVersion: formulaBilling,
			Currency: "IDR", Metrics: metrics, TotalEvents: total}
		raw, err := events.Canonical(payload)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(raw)
		checksum = "sha256:" + hex.EncodeToString(sum[:])
		exportID = uuid.NewSHA1(idNamespace, []byte(tenant+"|"+period+"|"+version)).String()
		createdAt = time.Now().UTC()
		tag, err := tx.Exec(ctx, `INSERT INTO metering.billing_export
			(export_id, tenant_id, period, version, payload, checksum, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (tenant_id, period, version) DO NOTHING`,
			exportID, tenant, period, version, raw, checksum, createdAt)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			reused = true
			var stored []byte
			if err := tx.QueryRow(ctx, `SELECT export_id::text, checksum, payload, created_at
				FROM metering.billing_export WHERE tenant_id=$1 AND period=$2 AND version=$3`,
				tenant, period, version).Scan(&exportID, &checksum, &stored, &createdAt); err != nil {
				return err
			}
			return json.Unmarshal(stored, &payload)
		}
		return outbox.Add(ctx, tx, "metering", outbox.Event{
			ID: id.New(), TenantID: tenant, Type: pipeline.TopicBillingExported,
			AggregateKey: exportID,
			Payload: pipeline.BillingExported{ExportID: exportID, Period: period, Version: version,
				Checksum: checksum, CreatedAt: createdAt, Actor: actor},
		})
	})
	return payload, checksum, exportID, createdAt, reused, err
}

func (s *api) getBillingExport(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	period := chi.URLParam(r, "period")
	version := strings.TrimSpace(r.URL.Query().Get("version"))
	if version == "" {
		version = "v1"
	}
	var payload billingPayload
	var checksum, exportID string
	var createdAt time.Time
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		var raw []byte
		if err := tx.QueryRow(r.Context(), `SELECT export_id::text, checksum, payload, created_at
			FROM metering.billing_export WHERE tenant_id=$1 AND period=$2 AND version=$3`,
			p.TenantID, period, version).Scan(&exportID, &checksum, &raw, &createdAt); err != nil {
			return err
		}
		return json.Unmarshal(raw, &payload)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteProblem(w, r, http.StatusNotFound, "NOT_FOUND", "billing export not found")
		return
	}
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not load billing export")
		return
	}
	raw, _ := events.Canonical(payload)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Export-Id", exportID)
	w.Header().Set("X-Checksum", checksum)
	w.Header().Set("X-Created-At", createdAt.UTC().Format(time.RFC3339))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

type requestError struct {
	code    string
	message string
}

func (e requestError) Error() string { return e.message }

func errUnprocessable(code, message string) error { return requestError{code: code, message: message} }

func writeBillingError(w http.ResponseWriter, r *http.Request, err error) {
	var re requestError
	if errors.As(err, &re) {
		httpx.WriteProblem(w, r, http.StatusConflict, re.code, re.message)
		return
	}
	httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not build billing export")
}
