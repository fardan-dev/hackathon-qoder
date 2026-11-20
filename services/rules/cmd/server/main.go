package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
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
)

type api struct {
	app    *app.App
	client *http.Client
}

func main() {
	ctx, cancel := server.SignalContext()
	defer cancel()
	a, err := app.New(ctx, "rules")
	if err != nil {
		slog.Error("startup_failed", "error", err.Error())
		return
	}
	defer a.Close()
	if err := ensureCatalog(ctx, a); err != nil {
		a.Log.Error("catalog_seed_failed", "error", err.Error())
		return
	}
	producer, err := kafka.NewProducer(a.Cfg.KafkaBrokers)
	if err != nil {
		a.Log.Error("kafka_producer_failed", "error", err.Error())
		return
	}
	defer producer.Close()
	relay := &outbox.Relay{Pool: a.Pool, Schema: "rules", Publisher: producer, Logger: a.Log, PollInterval: time.Second}
	go relay.Run(ctx)

	s := &api{app: a, client: &http.Client{Timeout: 15 * time.Second}}
	go func() {
		err := kafka.RunConsumer(ctx, a.Cfg.KafkaBrokers, "rules-engine",
			[]string{pipeline.Topic(pipeline.TopicBatchIngested), pipeline.Topic(pipeline.TopicCaseDecided)},
			a.Log, s.handleEvent)
		if err != nil && ctx.Err() == nil {
			a.Log.Error("consumer_stopped", "error", err.Error())
		}
	}()

	verifier := a.Verifier()
	a.Router.Route("/v1", func(r chi.Router) {
		r.Use(auth.Middleware(verifier))
		r.Get("/rules", s.listRules)
		r.Get("/evaluations/{claimID}", s.getEvaluation)
		r.Get("/kill-switches", s.listKillSwitches)
		r.With(auth.RequireRoles(auth.RoleAdmin)).Post("/kill-switches", s.setKillSwitch)
		r.Get("/rules/lifecycle", s.listLifecycle)
		r.With(auth.RequireRoles(auth.RoleAdmin)).Post("/rules/{ruleID}/lifecycle", s.setLifecycle)
		r.With(auth.RequireRoles(auth.RoleAdmin, auth.RoleReviewer, auth.RoleSupervisor)).
			Post("/backtests/{batchID}", s.backtest)
	})
	if err := a.Run(ctx); err != nil {
		a.Log.Error("server_failed", "error", err.Error())
	}
}

func (s *api) handleEvent(ctx context.Context, env events.Envelope) error {
	switch env.Type {
	case pipeline.TopicBatchIngested:
		if err := s.handleBatch(ctx, env); err != nil {
			s.app.Log.Error("batch_evaluation_failed", "event_id", env.ID, "tenant_id", env.TenantID, "error", err.Error())
		}
	case pipeline.TopicCaseDecided:
		if err := s.applyDecision(ctx, env); err != nil {
			s.app.Log.Error("decision_projection_failed", "event_id", env.ID, "error", err.Error())
		}
	}
	return nil
}

func (s *api) handleBatch(ctx context.Context, env events.Envelope) error {
	var ev pipeline.BatchIngested
	if err := json.Unmarshal(env.Payload, &ev); err != nil {
		return err
	}
	asOf := env.OccurredAt.UTC().Truncate(time.Second)
	ec, err := s.loadContext(ctx, env.TenantID, ev.BatchID, asOf)
	if err != nil {
		return err
	}
	br, err := rulesengine.Evaluate(ec, s.app.Cfg.RuleCatalogVersion)
	if err != nil {
		return err
	}
	err = db.WithTenant(ctx, s.app.Pool, env.TenantID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `INSERT INTO rules.processed_event (consumer, event_id) VALUES ('rules-engine',$1)
			ON CONFLICT DO NOTHING`, env.ID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return nil
		}
		for _, cr := range br.Results {
			claim := findClaim(ec.Current, cr.ClaimID)
			if claim == nil {
				continue
			}
			flagsRaw, err := events.Canonical(cr.Flags)
			if err != nil {
				return err
			}
			neRaw, err := events.Canonical(cr.NotEvaluated)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `INSERT INTO rules.evaluation
				(evaluation_id, tenant_id, batch_id, claim_id, rule_pack_version, engine_version,
				 config_version, as_of, flags, not_evaluated, output_hash)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
				ON CONFLICT (tenant_id, claim_id, rule_pack_version, engine_version, config_version) DO NOTHING`,
				id.New(), env.TenantID, claim.BatchID, claim.ClaimID, br.RulePackVersion, br.EngineVersion,
				br.ConfigVersion, br.AsOf, flagsRaw, neRaw, cr.OutputHash); err != nil {
				return err
			}
			if err := upsertHistory(ctx, tx, env.TenantID, *claim); err != nil {
				return err
			}
			severity := highestSeverity(cr.Flags)
			ev := pipeline.ClaimEvaluated{
				BatchID: claim.BatchID, ClaimID: claim.ClaimID, ClaimantID: claim.ClaimantID,
				PolicyID: claim.PolicyID, ProviderID: claim.ProviderID, LossType: claim.LossType,
				ClaimStatus: claim.Status, ClaimDate: claim.ClaimDate, IncidentDate: claim.IncidentDate,
				ClaimedAmount: claim.ClaimedAmount, Fingerprint: claim.Fingerprint(), Severity: severity,
				Flags: cr.Flags, NotEvaluated: cr.NotEvaluated, RulePackVersion: br.RulePackVersion,
				EngineVersion: br.EngineVersion, ConfigVersion: br.ConfigVersion, AsOf: br.AsOf,
				OutputHash: cr.OutputHash, Actor: "system",
			}
			if err := outbox.Add(ctx, tx, "rules", outbox.Event{
				ID: id.New(), TenantID: env.TenantID, Type: pipeline.TopicClaimEvaluated,
				AggregateKey: claim.ClaimID, Payload: ev,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil {
		s.app.Log.Info("batch_evaluated", "batch_id", ev.BatchID,
			"claims", len(br.Results), "output_hash", br.OutputHash, "rule_pack", br.RulePackVersion)
	}
	return err
}

func (s *api) applyDecision(ctx context.Context, env events.Envelope) error {
	var ev pipeline.CaseDecided
	if err := json.Unmarshal(env.Payload, &ev); err != nil {
		return err
	}
	return db.WithTenant(ctx, s.app.Pool, env.TenantID, func(tx pgx.Tx) error {
		for _, claimID := range ev.ClaimIDs {
			switch ev.Outcome {
			case "CONFIRMED":
				_, err := tx.Exec(ctx, `UPDATE rules.claim_history SET prior_confirmed=true
					WHERE tenant_id=$1 AND claim_id=$2`, env.TenantID, claimID)
				if err != nil {
					return err
				}
			case "CLEARED":
				_, err := tx.Exec(ctx, `UPDATE rules.claim_history SET prior_cleared_reason=$3
					WHERE tenant_id=$1 AND claim_id=$2`, env.TenantID, claimID, ev.ReasonCode)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s *api) loadContext(ctx context.Context, tenant, batchID string, asOf time.Time) (rulesengine.EvalContext, error) {
	if asOf.IsZero() {
		asOf = time.Now().UTC().Truncate(time.Second)
	}
	var current []rulesengine.ClaimSnapshot
	for attempt := 0; attempt < 5; attempt++ {
		claims, err := s.fetchBatch(ctx, tenant, batchID)
		if err == nil {
			current = claims
			break
		}
		if attempt == 4 {
			return rulesengine.EvalContext{}, err
		}
		time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
	}
	history, err := s.loadHistory(ctx, tenant, batchID, asOf)
	if err != nil {
		return rulesengine.EvalContext{}, err
	}
	switches, err := s.loadKillSwitches(ctx, tenant)
	if err != nil {
		return rulesengine.EvalContext{}, err
	}
	return rulesengine.EvalContext{
		TenantID: tenant, AsOf: asOf, Current: current, History: history,
		DocForensicsEnabled: s.app.Cfg.DocForensics, KillSwitches: switches,
	}, nil
}

func (s *api) fetchBatch(ctx context.Context, tenant, batchID string) ([]rulesengine.ClaimSnapshot, error) {
	endpoint := strings.TrimRight(s.app.Cfg.IngestionURL, "/") + "/internal/batches/" + url.PathEscape(batchID) +
		"/claims?tenant_id=" + url.QueryEscape(tenant)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Internal-Token", s.app.Cfg.InternalToken)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("ingestion returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out pipeline.BatchClaims
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Claims, nil
}

func (s *api) loadHistory(ctx context.Context, tenant, batchID string, asOf time.Time) ([]rulesengine.ClaimSnapshot, error) {
	var out []rulesengine.ClaimSnapshot
	from := asOf.AddDate(-2, 0, 0)
	err := db.WithTenant(ctx, s.app.Pool, tenant, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `SELECT claim_id, batch_id::text, claimant_id, policy_id, provider_id,
			coalesce(invoice_number,''), loss_type, claim_date, incident_date, claimed_amount, status,
			fingerprint, coalesce(doc_sha256,''), coalesce(asset_id,''), sum_insured, inception_date,
			coverage_added_date, expiry_date, reinstatement_date, policy_status, identity_attrs,
			needs_info_count, coalesce(prior_cleared_reason,''), prior_confirmed,
			invoice_date, coalesce(invoice_amount,0)
			FROM rules.claim_history
			WHERE tenant_id=$1 AND batch_id <> $2 AND claim_date >= $3
			ORDER BY claim_id`, tenant, batchID, from)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var c rulesengine.ClaimSnapshot
			var policy rulesengine.PolicySnapshot
			var attrs []byte
			var historyFingerprint string
			var coverageAdded, reinstatement, invoiceDate *time.Time
			var invoiceAmount int64
			if err := rows.Scan(&c.ClaimID, &c.BatchID, &c.ClaimantID, &c.PolicyID, &c.ProviderID,
				&c.InvoiceNumber, &c.LossType, &c.ClaimDate, &c.IncidentDate, &c.ClaimedAmount, &c.Status,
				&historyFingerprint, &c.DocSHA256, &policy.AssetID, &policy.SumInsured, &policy.Inception,
				&coverageAdded, &policy.Expiry, &reinstatement, &policy.Status, &attrs,
				&c.PriorNeedsInfo, &c.PriorClearedReason, &c.PriorConfirmed, &invoiceDate, &invoiceAmount); err != nil {
				return err
			}
			_ = historyFingerprint
			c.TenantID = tenant
			if len(attrs) > 0 {
				_ = json.Unmarshal(attrs, &c.Identities)
			}
			policy.PolicyID = c.PolicyID
			policy.CoverageAdded = coverageAdded
			policy.Reinstatement = reinstatement
			c.Policy = &policy
			if c.InvoiceNumber != "" && invoiceAmount > 0 && invoiceDate != nil {
				c.Invoice = &rulesengine.InvoiceSnapshot{
					Number: c.InvoiceNumber, Date: *invoiceDate, ProviderID: c.ProviderID,
					Amount: invoiceAmount, ClaimID: c.ClaimID, DocSHA256: c.DocSHA256,
				}
			}
			out = append(out, c)
		}
		return rows.Err()
	})
	return out, err
}

func (s *api) loadKillSwitches(ctx context.Context, tenant string) ([]rulesengine.KillSwitch, error) {
	var out []rulesengine.KillSwitch
	err := db.WithTenant(ctx, s.app.Pool, tenant, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `SELECT scope, target, enabled FROM rules.kill_switch
			WHERE tenant_id=$1 OR tenant_id IS NULL ORDER BY scope, target`, tenant)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var sw rulesengine.KillSwitch
			if err := rows.Scan(&sw.Scope, &sw.Target, &sw.Enabled); err != nil {
				return err
			}
			out = append(out, sw)
		}
		return rows.Err()
	})
	return out, err
}

func upsertHistory(ctx context.Context, tx pgx.Tx, tenant string, c rulesengine.ClaimSnapshot) error {
	var policy rulesengine.PolicySnapshot
	if c.Policy != nil {
		policy = *c.Policy
	}
	var invoiceDate *time.Time
	var invoiceAmount int64
	if c.Invoice != nil {
		invoiceDate = &c.Invoice.Date
		invoiceAmount = c.Invoice.Amount
	}
	attrs, err := events.Canonical(c.Identities)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO rules.claim_history
		(tenant_id, claim_id, batch_id, claimant_id, policy_id, provider_id, invoice_number, loss_type,
		 claim_date, incident_date, claimed_amount, status, fingerprint, doc_sha256, asset_id,
		 sum_insured, inception_date, coverage_added_date, expiry_date, reinstatement_date, policy_status,
		 identity_attrs, needs_info_count, prior_cleared_reason, prior_confirmed, invoice_date, invoice_amount)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27)
		ON CONFLICT (tenant_id, claim_id) DO UPDATE SET
			batch_id=EXCLUDED.batch_id, claimant_id=EXCLUDED.claimant_id, policy_id=EXCLUDED.policy_id,
			provider_id=EXCLUDED.provider_id, invoice_number=EXCLUDED.invoice_number, loss_type=EXCLUDED.loss_type,
			claim_date=EXCLUDED.claim_date, incident_date=EXCLUDED.incident_date, claimed_amount=EXCLUDED.claimed_amount,
			status=EXCLUDED.status, fingerprint=EXCLUDED.fingerprint, doc_sha256=EXCLUDED.doc_sha256,
			asset_id=EXCLUDED.asset_id, sum_insured=EXCLUDED.sum_insured, inception_date=EXCLUDED.inception_date,
			coverage_added_date=EXCLUDED.coverage_added_date, expiry_date=EXCLUDED.expiry_date,
			reinstatement_date=EXCLUDED.reinstatement_date, policy_status=EXCLUDED.policy_status,
			identity_attrs=EXCLUDED.identity_attrs, invoice_date=EXCLUDED.invoice_date, invoice_amount=EXCLUDED.invoice_amount,
			prior_confirmed=rules.claim_history.prior_confirmed OR EXCLUDED.prior_confirmed,
			prior_cleared_reason=COALESCE(NULLIF(EXCLUDED.prior_cleared_reason,''), rules.claim_history.prior_cleared_reason)`,
		tenant, c.ClaimID, c.BatchID, c.ClaimantID, c.PolicyID, c.ProviderID, null(c.InvoiceNumber),
		c.LossType, c.ClaimDate, c.IncidentDate, c.ClaimedAmount, c.Status, c.Fingerprint(), null(c.DocSHA256),
		null(policy.AssetID), policy.SumInsured, policy.Inception, policy.CoverageAdded, policy.Expiry,
		policy.Reinstatement, policy.Status, attrs, c.PriorNeedsInfo, null(c.PriorClearedReason),
		c.PriorConfirmed, invoiceDate, invoiceAmount)
	return err
}

func null(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func findClaim(claims []rulesengine.ClaimSnapshot, idv string) *rulesengine.ClaimSnapshot {
	for i := range claims {
		if claims[i].ClaimID == idv {
			return &claims[i]
		}
	}
	return nil
}

func highestSeverity(flags []rulesengine.Flag) rulesengine.Severity {
	best := rulesengine.Severity("")
	for _, f := range flags {
		if best == "" || rulesengine.SeverityRank(f.Severity) < rulesengine.SeverityRank(best) {
			best = f.Severity
		}
	}
	return best
}

func ensureCatalog(ctx context.Context, a *app.App) error {
	if _, err := a.Pool.Exec(ctx, `INSERT INTO rules.rule_pack (version, engine_version, status)
		VALUES ($1,$2,'ACTIVE') ON CONFLICT (version) DO NOTHING`, rulesengine.CatalogVersion, rulesengine.EngineVersion); err != nil {
		return err
	}
	for _, m := range rulesengine.Catalog() {
		params, err := events.Canonical(m.Params)
		if err != nil {
			return err
		}
		if _, err := a.Pool.Exec(ctx, `INSERT INTO rules.catalog
			(rule_id, version, family, name, severity, params, params_version, status, owner, mapping_typology, activated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,now())
			ON CONFLICT (rule_id, version) DO UPDATE SET
				family=EXCLUDED.family, name=EXCLUDED.name, severity=EXCLUDED.severity, params=EXCLUDED.params,
				params_version=EXCLUDED.params_version, status=EXCLUDED.status, owner=EXCLUDED.owner,
				mapping_typology=EXCLUDED.mapping_typology`,
			m.ID, m.Version, m.Family, m.Name, string(m.Severity), params, m.ParamsVersion,
			string(m.Status), m.Owner, m.Typologies); err != nil {
			return err
		}
	}
	return nil
}

func (s *api) listRules(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"rule_pack_version": rulesengine.CatalogVersion,
		"engine_version":    rulesengine.EngineVersion,
		"rules":             rulesengine.Catalog(),
	})
}

func (s *api) getEvaluation(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	claimID := chi.URLParam(r, "claimID")
	var batchID, pack, engine, config, hash string
	var asOf, at time.Time
	var flagsRaw, neRaw []byte
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `SELECT batch_id::text, rule_pack_version, engine_version, config_version,
			as_of, flags, not_evaluated, output_hash, evaluated_at
			FROM rules.evaluation WHERE tenant_id=$1 AND claim_id=$2
			ORDER BY evaluated_at DESC LIMIT 1`, p.TenantID, claimID).
			Scan(&batchID, &pack, &engine, &config, &asOf, &flagsRaw, &neRaw, &hash, &at)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteProblem(w, r, http.StatusNotFound, "NOT_FOUND", "no evaluation for this claim")
		return
	}
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not load evaluation")
		return
	}
	var flags []rulesengine.Flag
	var ne []rulesengine.NotEvaluated
	_ = json.Unmarshal(flagsRaw, &flags)
	_ = json.Unmarshal(neRaw, &ne)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"claim_id": claimID, "batch_id": batchID, "rule_pack_version": pack, "engine_version": engine,
		"config_version": config, "as_of": asOf, "flags": flags, "not_evaluated": ne,
		"output_hash": hash, "evaluated_at": at,
	})
}

func (s *api) listKillSwitches(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	type row struct {
		SwitchID  string    `json:"switch_id"`
		Scope     string    `json:"scope"`
		Target    string    `json:"target"`
		Enabled   bool      `json:"enabled"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	var out []row
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT switch_id::text, scope, target, enabled, updated_at
			FROM rules.kill_switch WHERE tenant_id=$1 OR tenant_id IS NULL ORDER BY scope, target`, p.TenantID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var v row
			if err := rows.Scan(&v.SwitchID, &v.Scope, &v.Target, &v.Enabled, &v.UpdatedAt); err != nil {
				return err
			}
			out = append(out, v)
		}
		return rows.Err()
	})
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not list kill switches")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"kill_switches": out})
}

func (s *api) setKillSwitch(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	var req struct {
		Scope   string `json:"scope"`
		Target  string `json:"target"`
		Enabled bool   `json:"enabled"`
	}
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "INVALID_JSON", "invalid kill switch payload")
		return
	}
	req.Scope = strings.TrimSpace(req.Scope)
	req.Target = strings.TrimSpace(req.Target)
	switch req.Scope {
	case "global":
		req.Target = "ALL"
	case "family", "rule":
		if req.Target == "" {
			httpx.WriteProblem(w, r, http.StatusBadRequest, "REQUIRED_FIELD", "target is required")
			return
		}
	default:
		httpx.WriteProblem(w, r, http.StatusBadRequest, "INVALID_ENUM", "scope must be global, family, or rule")
		return
	}
	var switchID string
	now := time.Now().UTC()
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		var err error
		if req.Scope == "global" {
			err = tx.QueryRow(r.Context(), `UPDATE rules.kill_switch SET enabled=$1, updated_by=$2, updated_at=now()
				WHERE tenant_id IS NULL AND scope='global' AND target=$3 RETURNING switch_id::text`, req.Enabled, p.UserID, req.Target).Scan(&switchID)
			if errors.Is(err, pgx.ErrNoRows) {
				switchID = id.New()
				_, err = tx.Exec(r.Context(), `INSERT INTO rules.kill_switch
					(switch_id, tenant_id, scope, target, enabled, updated_by) VALUES ($1,NULL,'global',$2,$3,$4)`,
					switchID, req.Target, req.Enabled, p.UserID)
			}
		} else {
			switchID = id.New()
			err = tx.QueryRow(r.Context(), `INSERT INTO rules.kill_switch
				(switch_id, tenant_id, scope, target, enabled, updated_by) VALUES ($1,$2,$3,$4,$5,$6)
				ON CONFLICT (tenant_id, scope, target) DO UPDATE SET
					enabled=EXCLUDED.enabled, updated_by=EXCLUDED.updated_by, updated_at=now()
				RETURNING switch_id::text`, switchID, p.TenantID, req.Scope, req.Target, req.Enabled, p.UserID).Scan(&switchID)
		}
		if err != nil {
			return err
		}
		return outbox.Add(r.Context(), tx, "rules", outbox.Event{
			ID: id.New(), TenantID: p.TenantID, Type: pipeline.TopicKillSwitchChanged, AggregateKey: switchID,
			Payload: pipeline.KillSwitchChanged{SwitchID: switchID, Scope: req.Scope, Target: req.Target,
				Enabled: req.Enabled, ChangedBy: p.Email, ChangedAt: now, Actor: p.Email},
		})
	})
	if err != nil {
		s.app.Log.Error("kill_switch_failed", "error", err.Error())
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not update kill switch")
		return
	}
	s.app.Log.Info("kill_switch_updated", "switch_id", switchID, "scope", req.Scope, "target", req.Target, "enabled", req.Enabled)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"switch_id": switchID, "scope": req.Scope,
		"target": req.Target, "enabled": req.Enabled})
}

func (s *api) listLifecycle(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	type statusRow struct {
		Status    string    `json:"status"`
		Reason    string    `json:"reason,omitempty"`
		ChangedAt time.Time `json:"changed_at,omitempty"`
	}
	states := map[string]statusRow{}
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT rule_id, status, reason, changed_at
			FROM rules.lifecycle WHERE tenant_id=$1`, p.TenantID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var ruleID string
			var st statusRow
			if err := rows.Scan(&ruleID, &st.Status, &st.Reason, &st.ChangedAt); err != nil {
				return err
			}
			states[ruleID] = st
		}
		return rows.Err()
	})
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not load rule lifecycle")
		return
	}
	type ruleState struct {
		RuleID          string    `json:"rule_id"`
		CatalogStatus   string    `json:"catalog_status"`
		LifecycleStatus string    `json:"lifecycle_status"`
		Reason          string    `json:"reason,omitempty"`
		ChangedAt       time.Time `json:"changed_at,omitempty"`
	}
	out := make([]ruleState, 0, 43)
	for _, m := range rulesengine.Catalog() {
		st := ruleState{RuleID: m.ID, CatalogStatus: string(m.Status), LifecycleStatus: "ACTIVE"}
		if s, ok := states[m.ID]; ok {
			st.LifecycleStatus = s.Status
			st.Reason = s.Reason
			st.ChangedAt = s.ChangedAt
		}
		out = append(out, st)
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"rule_pack_version": rulesengine.CatalogVersion, "rules": out,
	})
}

func (s *api) setLifecycle(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	ruleID := chi.URLParam(r, "ruleID")
	var meta *rulesengine.RuleMeta
	for _, m := range rulesengine.Catalog() {
		if m.ID == ruleID {
			m := m
			meta = &m
			break
		}
	}
	if meta == nil {
		httpx.WriteProblem(w, r, http.StatusNotFound, "NOT_FOUND", "rule is not in the pinned catalog")
		return
	}
	var req struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "INVALID_JSON", "invalid lifecycle payload")
		return
	}
	req.Status = strings.ToUpper(strings.TrimSpace(req.Status))
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Status != "ACTIVE" && req.Status != "PAUSED" && req.Status != "RETIRED" {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "INVALID_ENUM", "status must be ACTIVE, PAUSED or RETIRED")
		return
	}
	if req.Status != "ACTIVE" && req.Reason == "" {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "REQUIRED_FIELD", "reason is required when pausing or retiring a rule")
		return
	}
	enabled := req.Status != "ACTIVE"
	switchID := id.New()
	now := time.Now().UTC()
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(r.Context(), `INSERT INTO rules.lifecycle
			(tenant_id, rule_id, status, reason, changed_by, changed_at)
			VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (tenant_id, rule_id) DO UPDATE SET
				status=EXCLUDED.status, reason=EXCLUDED.reason,
				changed_by=EXCLUDED.changed_by, changed_at=EXCLUDED.changed_at`,
			p.TenantID, ruleID, req.Status, req.Reason, p.UserID, now); err != nil {
			return err
		}
		if err := tx.QueryRow(r.Context(), `INSERT INTO rules.kill_switch
			(switch_id, tenant_id, scope, target, enabled, updated_by)
			VALUES ($1,$2,'rule',$3,$4,$5)
			ON CONFLICT (tenant_id, scope, target) DO UPDATE SET
				enabled=EXCLUDED.enabled, updated_by=EXCLUDED.updated_by, updated_at=now()
			RETURNING switch_id::text`, switchID, p.TenantID, ruleID, enabled, p.UserID).Scan(&switchID); err != nil {
			return err
		}
		return outbox.Add(r.Context(), tx, "rules", outbox.Event{
			ID: id.New(), TenantID: p.TenantID, Type: pipeline.TopicKillSwitchChanged, AggregateKey: switchID,
			Payload: pipeline.KillSwitchChanged{SwitchID: switchID, Scope: "rule", Target: ruleID,
				Enabled: enabled, Reason: "lifecycle:" + req.Status + ":" + req.Reason,
				ChangedBy: p.Email, ChangedAt: now, Actor: p.Email},
		})
	})
	if err != nil {
		s.app.Log.Error("rule_lifecycle_failed", "rule_id", ruleID, "error", err.Error())
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not change rule lifecycle")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"rule_id": ruleID, "catalog_status": string(meta.Status), "lifecycle_status": req.Status,
		"kill_switch_enabled": enabled, "effective_next_run": !enabled, "switch_id": switchID,
	})
}

func (s *api) backtest(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	batchID := chi.URLParam(r, "batchID")
	asOf := time.Now().UTC().Truncate(time.Second)
	ec, err := s.loadContext(r.Context(), p.TenantID, batchID, asOf)
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusBadGateway, "INGESTION_UNAVAILABLE", "could not load batch for backtest")
		return
	}
	first, err := rulesengine.Evaluate(ec, s.app.Cfg.RuleCatalogVersion)
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "BACKTEST_FAILED", err.Error())
		return
	}
	second, err := rulesengine.Evaluate(ec, s.app.Cfg.RuleCatalogVersion)
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "BACKTEST_FAILED", err.Error())
		return
	}
	flags := 0
	for _, res := range first.Results {
		flags += len(res.Flags)
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"batch_id": batchID, "rule_pack_version": first.RulePackVersion, "engine_version": first.EngineVersion,
		"config_version": first.ConfigVersion, "output_hash": first.OutputHash,
		"replay_hash": second.OutputHash, "identical": first.OutputHash == second.OutputHash,
		"claims": len(first.Results), "flags": flags, "as_of": first.AsOf,
	})
}
