package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fardanwardhana/qoder-hackathon/packages/contracts"
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

type api struct {
	app *app.App
}

func main() {
	ctx, cancel := server.SignalContext()
	defer cancel()
	a, err := app.New(ctx, "reporting")
	if err != nil {
		slog.Error("startup_failed", "error", err.Error())
		return
	}
	defer a.Close()
	s := &api{app: a}
	go func() {
		err := kafka.RunConsumer(ctx, a.Cfg.KafkaBrokers, "reporting-snapshots", []string{
			pipeline.Topic(pipeline.TopicAlertCreated),
			pipeline.Topic(pipeline.TopicCaseDecided),
			pipeline.Topic(pipeline.TopicSuppressionApproved),
			pipeline.Topic(pipeline.TopicSuppressionRevoked),
		}, a.Log, s.handleEvent)
		if err != nil && ctx.Err() == nil {
			a.Log.Error("consumer_stopped", "error", err.Error())
		}
	}()

	verifier := a.Verifier()
	a.Router.Route("/v1", func(r chi.Router) {
		r.Use(auth.Middleware(verifier))
		r.Get("/coverage", s.coverage)
		r.Get("/coverage/export", s.coverageExport)
		r.Get("/dashboard", s.dashboard)
	})
	if err := a.Run(ctx); err != nil {
		a.Log.Error("server_failed", "error", err.Error())
	}
}

func (s *api) handleEvent(ctx context.Context, env events.Envelope) error {
	switch env.Type {
	case pipeline.TopicAlertCreated:
		return s.snapshotAlert(ctx, env)
	case pipeline.TopicCaseDecided:
		return s.snapshotDecision(ctx, env)
	case pipeline.TopicSuppressionApproved:
		var ev pipeline.SuppressionApproved
		if err := json.Unmarshal(env.Payload, &ev); err != nil {
			return err
		}
		return db.WithTenant(ctx, s.app.Pool, env.TenantID, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `UPDATE reporting.alert_snapshot SET suppressed=true
				WHERE tenant_id=$1 AND claim_id=$2`, env.TenantID, ev.ClaimID)
			return err
		})
	case pipeline.TopicSuppressionRevoked:
		var ev pipeline.SuppressionRevoked
		if err := json.Unmarshal(env.Payload, &ev); err != nil {
			return err
		}
		return db.WithTenant(ctx, s.app.Pool, env.TenantID, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `UPDATE reporting.alert_snapshot SET suppressed=false
				WHERE tenant_id=$1 AND claim_id=$2`, env.TenantID, ev.ClaimID)
			return err
		})
	}
	return nil
}

func (s *api) snapshotAlert(ctx context.Context, env events.Envelope) error {
	var ev pipeline.AlertCreated
	if err := json.Unmarshal(env.Payload, &ev); err != nil {
		return err
	}
	flagsRaw, err := events.Canonical(ev.Flags)
	if err != nil {
		return err
	}
	return db.WithTenant(ctx, s.app.Pool, env.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO reporting.alert_snapshot
			(tenant_id, alert_id, claim_id, loss_type, severity, severity_rank, flag_count,
			 claimed_amount, flags, suppressed, status, sla_due_at, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'OPEN',$11,$12)
			ON CONFLICT (tenant_id, alert_id) DO UPDATE SET
				loss_type=EXCLUDED.loss_type, severity=EXCLUDED.severity,
				severity_rank=EXCLUDED.severity_rank, flag_count=EXCLUDED.flag_count,
				claimed_amount=EXCLUDED.claimed_amount, flags=EXCLUDED.flags,
				suppressed=EXCLUDED.suppressed, sla_due_at=EXCLUDED.sla_due_at`,
			env.TenantID, ev.AlertID, ev.ClaimID, ev.LossType, string(ev.Severity), ev.SeverityRank,
			ev.FlagCount, ev.ClaimedAmount, flagsRaw, ev.Suppressed, ev.SLADueAt, ev.CreatedAt)
		return err
	})
}

func (s *api) snapshotDecision(ctx context.Context, env events.Envelope) error {
	var ev pipeline.CaseDecided
	if err := json.Unmarshal(env.Payload, &ev); err != nil {
		return err
	}
	claimsRaw, err := events.Canonical(ev.ClaimIDs)
	if err != nil {
		return err
	}
	return db.WithTenant(ctx, s.app.Pool, env.TenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO reporting.decision_snapshot
			(tenant_id, case_id, outcome, reason_code, value_protected, value_recovered, claim_ids, decided_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			ON CONFLICT (tenant_id, case_id) DO UPDATE SET
				outcome=EXCLUDED.outcome, reason_code=EXCLUDED.reason_code,
				value_protected=EXCLUDED.value_protected, value_recovered=EXCLUDED.value_recovered,
				claim_ids=EXCLUDED.claim_ids, decided_at=EXCLUDED.decided_at`,
			env.TenantID, ev.CaseID, ev.Outcome, ev.ReasonCode, ev.ValueProtected,
			ev.ValueRecovered, claimsRaw, ev.DecidedAt); err != nil {
			return err
		}
		if ev.Outcome != "CONFIRMED" && ev.Outcome != "CLEARED" {
			return nil
		}
		_, err := tx.Exec(ctx, `UPDATE reporting.alert_snapshot SET status='CLOSED'
			WHERE tenant_id=$1 AND claim_id = ANY($2)`, env.TenantID, ev.ClaimIDs)
		return err
	})
}

type coverageRow struct {
	TypologyKey    string   `json:"typology_key"`
	Label          string   `json:"label"`
	InScope        bool     `json:"in_scope"`
	Phase01Covered bool     `json:"phase01_covered"`
	RuleIDs        []string `json:"rule_ids"`
	ResidualRisk   string   `json:"residual_risk,omitempty"`
	SortOrder      int      `json:"sort_order"`
}

func (s *api) coverageRows(ctx context.Context) ([]coverageRow, error) {
	rows, err := s.app.Pool.Query(ctx, `SELECT typology_key, label, in_scope, phase01_covered,
		rule_ids, coalesce(residual_risk,''), sort_order FROM reporting.coverage_typology ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []coverageRow
	for rows.Next() {
		var c coverageRow
		if err := rows.Scan(&c.TypologyKey, &c.Label, &c.InScope, &c.Phase01Covered,
			&c.RuleIDs, &c.ResidualRisk, &c.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *api) coverage(w http.ResponseWriter, r *http.Request) {
	rows, err := s.coverageRows(r.Context())
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not load coverage")
		return
	}
	total := 0
	covered := 0
	var residual []coverageRow
	for _, c := range rows {
		if c.InScope {
			total++
		}
		if c.Phase01Covered {
			covered++
		}
		if c.InScope && !c.Phase01Covered {
			residual = append(residual, c)
		}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"typologies": rows,
		"summary": map[string]any{
			"denominator": total, "covered": covered, "residual": len(residual),
			"coverage_bps": bps(int64(covered), int64(total)),
		},
		"residual_risk_register": residual,
	})
}

func (s *api) coverageExport(w http.ResponseWriter, r *http.Request) {
	rows, err := s.coverageRows(r.Context())
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not load coverage")
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="coverage-matrix.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"typology_key", "label", "in_scope", "phase01_covered", "rule_ids", "residual_risk"})
	for _, c := range rows {
		_ = cw.Write([]string{
			contracts.EscapeCSVCell(c.TypologyKey),
			contracts.EscapeCSVCell(c.Label),
			strconv.FormatBool(c.InScope),
			strconv.FormatBool(c.Phase01Covered),
			contracts.EscapeCSVCell(strings.Join(c.RuleIDs, "|")),
			contracts.EscapeCSVCell(c.ResidualRisk),
		})
	}
	cw.Flush()
}

func (s *api) dashboard(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	from, to, err := dateRange(r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION", err.Error())
		return
	}
	lob := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("lob")))
	if lob != "" && lob != "MOTOR" && lob != "PROPERTY" {
		httpx.WriteProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION", "lob must be MOTOR or PROPERTY")
		return
	}
	out := map[string]any{"from": from, "to": to, "lob": lob}
	err = db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		var total, suppressed, flags, high, medium, low, breached int
		var avgFlags float64
		if err := tx.QueryRow(r.Context(), `SELECT count(*),
			count(*) FILTER (WHERE suppressed),
			coalesce(sum(flag_count),0),
			coalesce(avg(flag_count),0),
			count(*) FILTER (WHERE severity='High'),
			count(*) FILTER (WHERE severity='Medium'),
			count(*) FILTER (WHERE severity='Low'),
			count(*) FILTER (WHERE status='OPEN' AND sla_due_at < now())
			FROM reporting.alert_snapshot
			WHERE tenant_id=$1 AND created_at >= $2 AND created_at < $3
			  AND ($4='' OR loss_type=$4)`, p.TenantID, from, to, lob).Scan(
			&total, &suppressed, &flags, &avgFlags, &high, &medium, &low, &breached); err != nil {
			return err
		}
		out["alerts"] = map[string]any{
			"total": total, "suppressed": suppressed, "total_flags": flags,
			"avg_flags": avgFlags, "sla_breached_open": breached,
			"by_severity":        map[string]int{"High": high, "Medium": medium, "Low": low},
			"sla_compliance_bps": bps(int64(total-breached), int64(total)),
		}

		var decisions, confirmed, cleared, needsInfo int64
		var valueProtected, valueRecovered int64
		if err := tx.QueryRow(r.Context(), `SELECT count(*),
			count(*) FILTER (WHERE outcome='CONFIRMED'),
			count(*) FILTER (WHERE outcome='CLEARED'),
			count(*) FILTER (WHERE outcome='NEEDS_INFO'),
			coalesce(sum(value_protected),0), coalesce(sum(value_recovered),0)
			FROM reporting.decision_snapshot d
			WHERE d.tenant_id=$1 AND d.decided_at >= $2 AND d.decided_at < $3
			  AND ($4='' OR EXISTS (SELECT 1 FROM reporting.alert_snapshot a
			    WHERE a.tenant_id=d.tenant_id AND a.loss_type=$4
			      AND a.claim_id IN (SELECT jsonb_array_elements_text(d.claim_ids))))`,
			p.TenantID, from, to, lob).Scan(&decisions, &confirmed, &cleared, &needsInfo,
			&valueProtected, &valueRecovered); err != nil {
			return err
		}
		out["decisions"] = map[string]any{
			"total": decisions, "confirmed": confirmed, "cleared": cleared, "needs_info": needsInfo,
			"value_protected": valueProtected, "value_recovered": valueRecovered,
			"false_positive_bps": bps(cleared, decisions),
		}

		flagRows, err := tx.Query(r.Context(), `SELECT coalesce(f->>'rule_id', f->>'rule'), count(*)
			FROM reporting.alert_snapshot a, jsonb_array_elements(a.flags) f
			WHERE a.tenant_id=$1 AND a.created_at >= $2 AND a.created_at < $3
			  AND ($4='' OR a.loss_type=$4)
			GROUP BY 1 ORDER BY 2 DESC, 1`, p.TenantID, from, to, lob)
		if err != nil {
			return err
		}
		dist := map[string]int64{}
		for flagRows.Next() {
			var rule string
			var count int64
			if err := flagRows.Scan(&rule, &count); err != nil {
				flagRows.Close()
				return err
			}
			dist[rule] = count
		}
		flagRows.Close()
		out["flag_distribution"] = dist
		return nil
	})
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not build dashboard")
		return
	}
	rows, err := s.coverageRows(r.Context())
	if err == nil {
		total, covered := 0, 0
		for _, c := range rows {
			if c.InScope {
				total++
			}
			if c.Phase01Covered {
				covered++
			}
		}
		out["coverage"] = map[string]any{"denominator": total, "covered": covered}
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func dateRange(fromRaw, toRaw string) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.UTC
	}
	to := time.Now().In(loc)
	from := to.AddDate(0, 0, -30)
	if strings.TrimSpace(fromRaw) != "" {
		parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(fromRaw), loc)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("from must use YYYY-MM-DD")
		}
		from = parsed
	}
	if strings.TrimSpace(toRaw) != "" {
		parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(toRaw), loc)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("to must use YYYY-MM-DD")
		}
		to = parsed
	}
	start := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, loc).UTC()
	end := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1).UTC()
	if !start.Before(end) {
		return time.Time{}, time.Time{}, errors.New("from must be on or before to")
	}
	return start, end, nil
}

func bps(part, total int64) int64 {
	if total <= 0 {
		return 0
	}
	return part * 10000 / total
}
