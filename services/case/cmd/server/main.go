package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
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
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	outcomeConfirmed = "CONFIRMED"
	outcomeCleared   = "CLEARED"
	outcomeNeedsInfo = "NEEDS_INFO"
)

var reasonRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,31}$`)

type api struct {
	app *app.App
}

func main() {
	ctx, cancel := server.SignalContext()
	defer cancel()
	a, err := app.New(ctx, "case")
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
	relay := &outbox.Relay{Pool: a.Pool, Schema: "cases", Publisher: producer, Logger: a.Log, PollInterval: time.Second}
	go relay.Run(ctx)

	s := &api{app: a}
	go func() {
		err := kafka.RunConsumer(ctx, a.Cfg.KafkaBrokers, "case-alert-projection",
			[]string{pipeline.Topic(pipeline.TopicAlertCreated)}, a.Log, s.handleAlertCreated)
		if err != nil && ctx.Err() == nil {
			a.Log.Error("consumer_stopped", "error", err.Error())
		}
	}()

	verifier := a.Verifier()
	a.Router.Route("/v1", func(r chi.Router) {
		r.Use(auth.Middleware(verifier))
		r.Get("/cases", s.listCases)
		r.Post("/cases", s.createCase)
		r.Get("/cases/{caseID}", s.getCase)
		r.Post("/cases/{caseID}/evidence", s.addEvidence)
		r.Post("/cases/{caseID}/package", s.freezePackage)
		r.Get("/cases/{caseID}/package/verify", s.verifyPackage)
		r.Post("/cases/{caseID}/decisions", s.recordDecision)
		r.Get("/suppressions", s.listSuppressions)
		r.With(auth.RequireRoles(auth.RoleSupervisor, auth.RoleAdmin)).Post("/suppressions", s.createSuppression)
		r.With(auth.RequireRoles(auth.RoleSupervisor, auth.RoleAdmin)).Post("/suppressions/{suppressionID}/revoke", s.revokeSuppression)
	})
	if err := a.Run(ctx); err != nil {
		a.Log.Error("server_failed", "error", err.Error())
	}
}

func (s *api) handleAlertCreated(ctx context.Context, env events.Envelope) error {
	var ev pipeline.AlertCreated
	if err := json.Unmarshal(env.Payload, &ev); err != nil {
		return err
	}
	flagsRaw, err := events.Canonical(ev.Flags)
	if err != nil {
		return err
	}
	return db.WithTenant(ctx, s.app.Pool, env.TenantID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `INSERT INTO cases.processed_event (consumer, event_id)
			VALUES ('case-alert-projection',$1) ON CONFLICT DO NOTHING`, env.ID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return nil
		}
		_, err = tx.Exec(ctx, `INSERT INTO cases.alert_projection
			(tenant_id, alert_id, claim_id, fingerprint, loss_type, severity, claimed_amount, priority_rank, flags, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			ON CONFLICT (tenant_id, claim_id) DO UPDATE SET
				alert_id=EXCLUDED.alert_id, fingerprint=EXCLUDED.fingerprint, loss_type=EXCLUDED.loss_type,
				severity=EXCLUDED.severity, claimed_amount=EXCLUDED.claimed_amount,
				priority_rank=EXCLUDED.priority_rank, flags=EXCLUDED.flags`,
			env.TenantID, ev.AlertID, ev.ClaimID, ev.Fingerprint, ev.LossType, string(ev.Severity),
			ev.ClaimedAmount, ev.SeverityRank, flagsRaw, ev.CreatedAt)
		return err
	})
}

type createCaseRequest struct {
	Title    string   `json:"title"`
	Summary  string   `json:"summary"`
	Priority int      `json:"priority"`
	ClaimIDs []string `json:"claim_ids"`
}

func (s *api) createCase(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	var req createCaseRequest
	if err := decodeJSON(r, &req); err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" || len(req.ClaimIDs) == 0 {
		httpx.WriteProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION", "title and at least one claim_id are required")
		return
	}
	if req.Priority == 0 {
		req.Priority = 2
	}
	if req.Priority < 1 || req.Priority > 4 {
		httpx.WriteProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION", "priority must be between 1 and 4")
		return
	}
	caseID := id.New()
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		missing := make([]string, 0)
		for _, claimID := range req.ClaimIDs {
			var exists bool
			if err := tx.QueryRow(r.Context(), `SELECT EXISTS (SELECT 1 FROM cases.alert_projection
				WHERE tenant_id=$1 AND claim_id=$2)`, p.TenantID, claimID).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				missing = append(missing, claimID)
			}
		}
		if len(missing) > 0 {
			return errUnprocessable("UNKNOWN_CLAIM", "claims have no alert projection: "+strings.Join(missing, ","))
		}
		if _, err := tx.Exec(r.Context(), `INSERT INTO cases.case_group
			(case_id, tenant_id, title, summary, priority, owner_id) VALUES ($1,$2,$3,$4,$5,$6)`,
			caseID, p.TenantID, req.Title, req.Summary, req.Priority, p.UserID); err != nil {
			return err
		}
		for _, claimID := range req.ClaimIDs {
			if _, err := tx.Exec(r.Context(), `INSERT INTO cases.case_claim (tenant_id, case_id, claim_id)
				VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, p.TenantID, caseID, claimID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		writeCaseError(w, r, err, "could not create case")
		return
	}
	c, err := s.loadCase(r.Context(), p.TenantID, caseID)
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "case created but could not be loaded")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, c)
}

type caseRow struct {
	CaseID      string        `json:"case_id"`
	Title       string        `json:"title"`
	Summary     string        `json:"summary"`
	Status      string        `json:"status"`
	Priority    int           `json:"priority"`
	OwnerID     *string       `json:"owner_id,omitempty"`
	Version     int           `json:"version"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	ClosedAt    *time.Time    `json:"closed_at,omitempty"`
	Claims      []caseClaim   `json:"claims"`
	Evidence    []evidenceRow `json:"evidence"`
	Package     *packageRow   `json:"package,omitempty"`
	Decisions   []decisionRow `json:"decisions"`
	OpenPackage bool          `json:"has_open_evidence"`
	ClaimCount  int           `json:"claim_count"`
}

type caseClaim struct {
	ClaimID       string          `json:"claim_id"`
	AlertID       string          `json:"alert_id"`
	LossType      string          `json:"loss_type"`
	Severity      string          `json:"severity"`
	ClaimedAmount int64           `json:"claimed_amount"`
	PriorityRank  int             `json:"priority_rank"`
	Flags         json.RawMessage `json:"flags"`
}

type evidenceRow struct {
	EvidenceID     string          `json:"evidence_id"`
	ArtifactType   string          `json:"artifact_type"`
	Source         string          `json:"source"`
	ContentSHA256  string          `json:"content_sha256"`
	ChainOfCustody json.RawMessage `json:"chain_of_custody"`
	Notes          string          `json:"notes"`
	CreatedAt      time.Time       `json:"created_at"`
}

type packageRow struct {
	PackageID       string    `json:"package_id"`
	PackageSHA256   string    `json:"package_sha256"`
	ArtifactCount   int       `json:"artifact_count"`
	EngineVersion   string    `json:"engine_version"`
	RulePackVersion string    `json:"rule_pack_version"`
	FrozenAt        time.Time `json:"frozen_at"`
}

type decisionRow struct {
	DecisionID     string    `json:"decision_id"`
	Outcome        string    `json:"outcome"`
	ReasonCode     string    `json:"reason_code"`
	ReviewerID     string    `json:"reviewer_id"`
	Notes          string    `json:"notes"`
	ValueProtected *int64    `json:"value_protected,omitempty"`
	ValueRecovered *int64    `json:"value_recovered,omitempty"`
	DecidedAt      time.Time `json:"decided_at"`
}

func (s *api) listCases(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	status := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("status")))
	limit := 100
	if n, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && n > 0 && n <= 500 {
		limit = n
	}
	cases := []caseRow{}
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT c.case_id::text, c.title, c.summary, c.status, c.priority,
			c.owner_id::text, c.version, c.created_at, c.updated_at, c.closed_at,
			(SELECT count(*) FROM cases.case_claim cc WHERE cc.case_id=c.case_id)
			FROM cases.case_group c
			WHERE c.tenant_id=$1 AND ($2='' OR c.status=$2)
			ORDER BY c.priority, c.created_at DESC LIMIT $3`, p.TenantID, status, limit)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var c caseRow
			if err := rows.Scan(&c.CaseID, &c.Title, &c.Summary, &c.Status, &c.Priority,
				&c.OwnerID, &c.Version, &c.CreatedAt, &c.UpdatedAt, &c.ClosedAt, &c.ClaimCount); err != nil {
				return err
			}
			cases = append(cases, c)
		}
		return rows.Err()
	})
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not load cases")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"cases": cases, "count": len(cases)})
}

func (s *api) getCase(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	c, err := s.loadCase(r.Context(), p.TenantID, chi.URLParam(r, "caseID"))
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteProblem(w, r, http.StatusNotFound, "NOT_FOUND", "case not found")
		return
	}
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not load case")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, c)
}

func (s *api) loadCase(ctx context.Context, tenant, caseID string) (caseRow, error) {
	var c caseRow
	err := db.WithTenant(ctx, s.app.Pool, tenant, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `SELECT case_id::text, title, summary, status, priority, owner_id::text,
			version, created_at, updated_at, closed_at FROM cases.case_group
			WHERE tenant_id=$1 AND case_id=$2`, tenant, caseID).Scan(&c.CaseID, &c.Title, &c.Summary,
			&c.Status, &c.Priority, &c.OwnerID, &c.Version, &c.CreatedAt, &c.UpdatedAt, &c.ClosedAt); err != nil {
			return err
		}
		claimRows, err := tx.Query(ctx, `SELECT cc.claim_id, ap.alert_id, ap.loss_type, ap.severity,
			ap.claimed_amount, ap.priority_rank, ap.flags
			FROM cases.case_claim cc JOIN cases.alert_projection ap
			  ON ap.tenant_id=cc.tenant_id AND ap.claim_id=cc.claim_id
			WHERE cc.case_id=$1 ORDER BY ap.priority_rank, ap.claim_id`, c.CaseID)
		if err != nil {
			return err
		}
		for claimRows.Next() {
			var cc caseClaim
			if err := claimRows.Scan(&cc.ClaimID, &cc.AlertID, &cc.LossType, &cc.Severity,
				&cc.ClaimedAmount, &cc.PriorityRank, &cc.Flags); err != nil {
				claimRows.Close()
				return err
			}
			c.Claims = append(c.Claims, cc)
		}
		claimRows.Close()
		if err := claimRows.Err(); err != nil {
			return err
		}
		c.ClaimCount = len(c.Claims)

		evRows, err := tx.Query(ctx, `SELECT evidence_id::text, artifact_type, source, content_sha256,
			chain_of_custody, notes, created_at FROM cases.evidence WHERE tenant_id=$1 AND case_id=$2
			ORDER BY created_at`, tenant, caseID)
		if err != nil {
			return err
		}
		for evRows.Next() {
			var ev evidenceRow
			if err := evRows.Scan(&ev.EvidenceID, &ev.ArtifactType, &ev.Source, &ev.ContentSHA256,
				&ev.ChainOfCustody, &ev.Notes, &ev.CreatedAt); err != nil {
				evRows.Close()
				return err
			}
			c.Evidence = append(c.Evidence, ev)
		}
		evRows.Close()
		if err := evRows.Err(); err != nil {
			return err
		}

		var pkg packageRow
		err = tx.QueryRow(ctx, `SELECT package_id::text, package_sha256, artifact_count,
			engine_version, rule_pack_version, frozen_at
			FROM cases.evidence_package WHERE tenant_id=$1 AND case_id=$2`, tenant, caseID).Scan(
			&pkg.PackageID, &pkg.PackageSHA256, &pkg.ArtifactCount,
			&pkg.EngineVersion, &pkg.RulePackVersion, &pkg.FrozenAt)
		if err == nil {
			c.Package = &pkg
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		c.OpenPackage = c.Package == nil

		decRows, err := tx.Query(ctx, `SELECT decision_id::text, outcome, reason_code, reviewer_id::text,
			notes, value_protected, value_recovered, decided_at FROM cases.decision
			WHERE tenant_id=$1 AND case_id=$2 ORDER BY decided_at`, tenant, caseID)
		if err != nil {
			return err
		}
		for decRows.Next() {
			var d decisionRow
			if err := decRows.Scan(&d.DecisionID, &d.Outcome, &d.ReasonCode, &d.ReviewerID,
				&d.Notes, &d.ValueProtected, &d.ValueRecovered, &d.DecidedAt); err != nil {
				decRows.Close()
				return err
			}
			c.Decisions = append(c.Decisions, d)
		}
		decRows.Close()
		return decRows.Err()
	})
	return c, err
}

type evidenceRequest struct {
	ArtifactType   string          `json:"artifact_type"`
	Source         string          `json:"source"`
	Content        string          `json:"content"`
	ChainOfCustody json.RawMessage `json:"chain_of_custody"`
	Notes          string          `json:"notes"`
}

func (s *api) addEvidence(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	var req evidenceRequest
	if err := decodeJSON(r, &req); err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	req.ArtifactType = strings.TrimSpace(req.ArtifactType)
	req.Source = strings.TrimSpace(req.Source)
	if req.ArtifactType == "" || req.Source == "" || req.Content == "" {
		httpx.WriteProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION", "artifact_type, source and content are required")
		return
	}
	if len(req.ChainOfCustody) == 0 {
		req.ChainOfCustody = json.RawMessage(`{}`)
	} else if !json.Valid(req.ChainOfCustody) {
		httpx.WriteProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION", "chain_of_custody must be a JSON object")
		return
	}
	sum := sha256.Sum256([]byte(req.Content))
	contentHash := "sha256:" + hex.EncodeToString(sum[:])
	evidenceID := id.New()
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		var frozen bool
		if err := tx.QueryRow(r.Context(), `SELECT EXISTS (SELECT 1 FROM cases.evidence_package
			WHERE tenant_id=$1 AND case_id=$2)`, p.TenantID, chi.URLParam(r, "caseID")).Scan(&frozen); err != nil {
			return err
		}
		if frozen {
			return errUnprocessable("PACKAGE_FROZEN", "evidence package is frozen; evidence cannot be added")
		}
		_, err := tx.Exec(r.Context(), `INSERT INTO cases.evidence
			(evidence_id, tenant_id, case_id, artifact_type, source, content_sha256, chain_of_custody, notes)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, evidenceID, p.TenantID, chi.URLParam(r, "caseID"),
			req.ArtifactType, req.Source, contentHash, req.ChainOfCustody, req.Notes)
		return err
	})
	if err != nil {
		writeCaseError(w, r, err, "could not add evidence")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]any{
		"evidence_id": evidenceID, "content_sha256": contentHash, "hash_verified": true,
	})
}

func (s *api) freezePackage(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	caseID := chi.URLParam(r, "caseID")
	var pkg packageRow
	var reused bool
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		var exists bool
		if err := tx.QueryRow(r.Context(), `SELECT EXISTS (SELECT 1 FROM cases.case_group
			WHERE tenant_id=$1 AND case_id=$2)`, p.TenantID, caseID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return pgx.ErrNoRows
		}
		err := tx.QueryRow(r.Context(), `SELECT package_id::text, package_sha256, artifact_count,
			engine_version, rule_pack_version, frozen_at
			FROM cases.evidence_package WHERE tenant_id=$1 AND case_id=$2`, p.TenantID, caseID).Scan(
			&pkg.PackageID, &pkg.PackageSHA256, &pkg.ArtifactCount,
			&pkg.EngineVersion, &pkg.RulePackVersion, &pkg.FrozenAt)
		if err == nil {
			reused = true
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		hashes, err := evidenceHashes(r.Context(), tx, p.TenantID, caseID)
		if err != nil {
			return err
		}
		if len(hashes) == 0 {
			return errUnprocessable("NO_EVIDENCE", "a package needs at least one evidence artifact")
		}
		hash, err := packageHash(hashes)
		if err != nil {
			return err
		}
		pkg = packageRow{
			PackageID: id.New(), PackageSHA256: hash, ArtifactCount: len(hashes),
			EngineVersion: s.app.Cfg.EngineVersion, RulePackVersion: s.app.Cfg.RuleCatalogVersion,
			FrozenAt: time.Now().UTC(),
		}
		_, err = tx.Exec(r.Context(), `INSERT INTO cases.evidence_package
			(package_id, tenant_id, case_id, package_sha256, artifact_count, engine_version, rule_pack_version, frozen_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, pkg.PackageID, p.TenantID, caseID, pkg.PackageSHA256,
			pkg.ArtifactCount, pkg.EngineVersion, pkg.RulePackVersion, pkg.FrozenAt)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return tx.QueryRow(r.Context(), `SELECT package_id::text, package_sha256, artifact_count,
					engine_version, rule_pack_version, frozen_at
					FROM cases.evidence_package WHERE tenant_id=$1 AND case_id=$2`, p.TenantID, caseID).Scan(
					&pkg.PackageID, &pkg.PackageSHA256, &pkg.ArtifactCount,
					&pkg.EngineVersion, &pkg.RulePackVersion, &pkg.FrozenAt)
			}
			return err
		}
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteProblem(w, r, http.StatusNotFound, "NOT_FOUND", "case not found")
		return
	}
	if err != nil {
		writeCaseError(w, r, err, "could not freeze evidence package")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"package": pkg, "reused": reused})
}

type evidenceHash struct {
	EvidenceID string `json:"evidence_id"`
	SHA256     string `json:"sha256"`
}

func evidenceHashes(ctx context.Context, tx pgx.Tx, tenant, caseID string) ([]evidenceHash, error) {
	rows, err := tx.Query(ctx, `SELECT evidence_id::text, content_sha256 FROM cases.evidence
		WHERE tenant_id=$1 AND case_id=$2 ORDER BY evidence_id`, tenant, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []evidenceHash
	for rows.Next() {
		var h evidenceHash
		if err := rows.Scan(&h.EvidenceID, &h.SHA256); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func packageHash(hashes []evidenceHash) (string, error) {
	raw, err := events.Canonical(hashes)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func (s *api) verifyPackage(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	caseID := chi.URLParam(r, "caseID")
	var stored, engineVersion, rulePackVersion string
	var count int
	var valid bool
	var computed string
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(r.Context(), `SELECT package_sha256, artifact_count, engine_version, rule_pack_version
			FROM cases.evidence_package WHERE tenant_id=$1 AND case_id=$2`, p.TenantID, caseID).
			Scan(&stored, &count, &engineVersion, &rulePackVersion); err != nil {
			return err
		}
		hashes, err := evidenceHashes(r.Context(), tx, p.TenantID, caseID)
		if err != nil {
			return err
		}
		computed, err = packageHash(hashes)
		if err != nil {
			return err
		}
		valid = computed == stored && len(hashes) == count
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteProblem(w, r, http.StatusNotFound, "NOT_FOUND", "package not found")
		return
	}
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not verify package")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"valid": valid, "stored_sha256": stored, "computed_sha256": computed, "artifact_count": count,
		"engine_version": engineVersion, "rule_pack_version": rulePackVersion,
	})
}

type decisionRequest struct {
	Outcome        string `json:"outcome"`
	ReasonCode     string `json:"reason_code"`
	Notes          string `json:"notes"`
	ValueProtected *int64 `json:"value_protected"`
	ValueRecovered *int64 `json:"value_recovered"`
}

func (s *api) recordDecision(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	var req decisionRequest
	if err := decodeJSON(r, &req); err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	req.Outcome = strings.ToUpper(strings.TrimSpace(req.Outcome))
	req.ReasonCode = strings.ToUpper(strings.TrimSpace(req.ReasonCode))
	if err := validateDecision(req); err != nil {
		httpx.WriteProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION", err.Error())
		return
	}
	caseID := chi.URLParam(r, "caseID")
	decisionID := id.New()
	var decidedAt time.Time
	var claimIDs []string
	var out pipeline.CaseDecided
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		var status string
		if err := tx.QueryRow(r.Context(), `SELECT status FROM cases.case_group
			WHERE tenant_id=$1 AND case_id=$2 FOR UPDATE`, p.TenantID, caseID).Scan(&status); err != nil {
			return err
		}
		if status == "CLOSED" {
			return errUnprocessable("CASE_CLOSED", "case is closed and cannot be decided again")
		}
		rows, err := tx.Query(r.Context(), `SELECT claim_id FROM cases.case_claim WHERE case_id=$1`, caseID)
		if err != nil {
			return err
		}
		for rows.Next() {
			var claimID string
			if err := rows.Scan(&claimID); err != nil {
				rows.Close()
				return err
			}
			claimIDs = append(claimIDs, claimID)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		decidedAt = time.Now().UTC()
		var valueProtected, valueRecovered int64
		if req.ValueProtected != nil {
			valueProtected = *req.ValueProtected
		}
		if req.ValueRecovered != nil {
			valueRecovered = *req.ValueRecovered
		}
		if _, err := tx.Exec(r.Context(), `INSERT INTO cases.decision
			(decision_id, tenant_id, case_id, outcome, reason_code, reviewer_id, notes, value_protected, value_recovered, decided_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			decisionID, p.TenantID, caseID, req.Outcome, req.ReasonCode, p.UserID, req.Notes,
			req.ValueProtected, req.ValueRecovered, decidedAt); err != nil {
			return err
		}
		newStatus := req.Outcome
		if req.Outcome == outcomeConfirmed || req.Outcome == outcomeCleared {
			newStatus = "CLOSED"
		}
		if _, err := tx.Exec(r.Context(), `UPDATE cases.case_group SET status=$3, version=version+1,
			updated_at=now(), closed_at=CASE WHEN $3='CLOSED' THEN now() ELSE closed_at END
			WHERE tenant_id=$1 AND case_id=$2`, p.TenantID, caseID, newStatus); err != nil {
			return err
		}
		out = pipeline.CaseDecided{
			CaseID: caseID, ClaimIDs: claimIDs, Outcome: req.Outcome, ReasonCode: req.ReasonCode,
			ReviewerID: p.UserID, ReviewerName: p.DisplayName, Notes: req.Notes,
			ValueProtected: valueProtected, ValueRecovered: valueRecovered,
			DecidedAt: decidedAt, Actor: p.Email,
		}
		return outbox.Add(r.Context(), tx, "cases", outbox.Event{
			ID: id.New(), TenantID: p.TenantID, Type: pipeline.TopicCaseDecided,
			AggregateKey: caseID, Payload: out,
		})
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteProblem(w, r, http.StatusNotFound, "NOT_FOUND", "case not found")
		return
	}
	if err != nil {
		writeCaseError(w, r, err, "could not record decision")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]any{
		"decision_id": decisionID, "case_id": caseID, "outcome": req.Outcome,
		"reason_code": req.ReasonCode, "decided_at": decidedAt, "claim_status_writeback": false,
	})
}

func validateDecision(req decisionRequest) error {
	switch req.Outcome {
	case outcomeConfirmed, outcomeCleared, outcomeNeedsInfo:
	default:
		return fmt.Errorf("outcome must be CONFIRMED, CLEARED or NEEDS_INFO")
	}
	if !reasonRe.MatchString(req.ReasonCode) {
		return fmt.Errorf("reason_code is required and must be an uppercase code")
	}
	if req.Outcome == outcomeCleared {
		pri := strings.SplitN(req.ReasonCode, "_", 2)[0]
		if pri != "FP" && pri != "NI" && pri != "CF" {
			return fmt.Errorf("CLEARED reason_code must use the FP, NI or CF category")
		}
	}
	if req.Outcome == outcomeConfirmed && strings.TrimSpace(req.Notes) == "" {
		return fmt.Errorf("CONFIRMED requires notes explaining the reason")
	}
	if req.ValueProtected != nil && *req.ValueProtected < 0 {
		return fmt.Errorf("value_protected must not be negative")
	}
	if req.ValueRecovered != nil && *req.ValueRecovered < 0 {
		return fmt.Errorf("value_recovered must not be negative")
	}
	return nil
}

type suppressionRequest struct {
	ClaimID   string    `json:"claim_id"`
	Reason    string    `json:"reason"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (s *api) createSuppression(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	var req suppressionRequest
	if err := decodeJSON(r, &req); err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	req.ClaimID = strings.TrimSpace(req.ClaimID)
	req.Reason = strings.TrimSpace(req.Reason)
	now := time.Now().UTC()
	if req.ClaimID == "" || req.Reason == "" {
		httpx.WriteProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION", "claim_id and reason are required")
		return
	}
	if req.ExpiresAt.IsZero() || !req.ExpiresAt.After(now) {
		httpx.WriteProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION", "expires_at must be in the future")
		return
	}
	if req.ExpiresAt.After(now.Add(90 * 24 * time.Hour)) {
		httpx.WriteProblem(w, r, http.StatusUnprocessableEntity, "VALIDATION", "expires_at cannot be more than 90 days away")
		return
	}
	suppressionID := id.New()
	var fingerprint string
	var out pipeline.SuppressionApproved
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(r.Context(), `SELECT fingerprint FROM cases.alert_projection
			WHERE tenant_id=$1 AND claim_id=$2`, p.TenantID, req.ClaimID).Scan(&fingerprint); err != nil {
			return err
		}
		createdAt := time.Now().UTC()
		if _, err := tx.Exec(r.Context(), `INSERT INTO cases.suppression
			(suppression_id, tenant_id, claim_id, fingerprint, reason, approved_by, expires_at, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, suppressionID, p.TenantID, req.ClaimID, fingerprint,
			req.Reason, p.UserID, req.ExpiresAt.UTC(), createdAt); err != nil {
			return err
		}
		out = pipeline.SuppressionApproved{
			SuppressionID: suppressionID, ClaimID: req.ClaimID, Fingerprint: fingerprint,
			Reason: req.Reason, ApprovedBy: p.UserID, ExpiresAt: req.ExpiresAt.UTC(),
			CreatedAt: createdAt, Actor: p.Email,
		}
		return outbox.Add(r.Context(), tx, "cases", outbox.Event{
			ID: id.New(), TenantID: p.TenantID, Type: pipeline.TopicSuppressionApproved,
			AggregateKey: suppressionID, Payload: out,
		})
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteProblem(w, r, http.StatusNotFound, "NOT_FOUND", "claim has no alert to suppress")
		return
	}
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not create suppression")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]any{
		"suppression_id": suppressionID, "claim_id": req.ClaimID,
		"fingerprint": fingerprint, "expires_at": req.ExpiresAt.UTC(), "flags_preserved": true,
	})
}

func (s *api) revokeSuppression(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	suppressionID := chi.URLParam(r, "suppressionID")
	var out pipeline.SuppressionRevoked
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		var claimID, fingerprint string
		var revokedAt *time.Time
		if err := tx.QueryRow(r.Context(), `SELECT claim_id, fingerprint, revoked_at FROM cases.suppression
			WHERE tenant_id=$1 AND suppression_id=$2 FOR UPDATE`, p.TenantID, suppressionID).Scan(
			&claimID, &fingerprint, &revokedAt); err != nil {
			return err
		}
		if revokedAt != nil {
			return errUnprocessable("ALREADY_REVOKED", "suppression has already been revoked")
		}
		now := time.Now().UTC()
		if _, err := tx.Exec(r.Context(), `UPDATE cases.suppression SET revoked_at=$3, revoked_by=$4
			WHERE tenant_id=$1 AND suppression_id=$2`, p.TenantID, suppressionID, now, p.UserID); err != nil {
			return err
		}
		out = pipeline.SuppressionRevoked{
			SuppressionID: suppressionID, ClaimID: claimID, Fingerprint: fingerprint,
			RevokedBy: p.UserID, RevokedAt: now, Actor: p.Email,
		}
		return outbox.Add(r.Context(), tx, "cases", outbox.Event{
			ID: id.New(), TenantID: p.TenantID, Type: pipeline.TopicSuppressionRevoked,
			AggregateKey: suppressionID, Payload: out,
		})
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteProblem(w, r, http.StatusNotFound, "NOT_FOUND", "suppression not found")
		return
	}
	if err != nil {
		writeCaseError(w, r, err, "could not revoke suppression")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"suppression_id": suppressionID, "revoked_at": out.RevokedAt})
}

func (s *api) listSuppressions(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	filter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("status")))
	type row struct {
		SuppressionID string     `json:"suppression_id"`
		ClaimID       string     `json:"claim_id"`
		Fingerprint   string     `json:"fingerprint"`
		Reason        string     `json:"reason"`
		ApprovedBy    string     `json:"approved_by"`
		ExpiresAt     time.Time  `json:"expires_at"`
		CreatedAt     time.Time  `json:"created_at"`
		RevokedAt     *time.Time `json:"revoked_at,omitempty"`
		Active        bool       `json:"active"`
	}
	out := []row{}
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT suppression_id::text, claim_id, fingerprint, reason,
			approved_by::text, expires_at, created_at, revoked_at,
			(revoked_at IS NULL AND expires_at > now()) FROM cases.suppression
			WHERE tenant_id=$1 AND ($2='' OR ($2='active' AND revoked_at IS NULL AND expires_at > now())
				OR ($2='revoked' AND revoked_at IS NOT NULL) OR ($2='expired' AND revoked_at IS NULL AND expires_at <= now()))
			ORDER BY created_at DESC`, p.TenantID, filter)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var x row
			if err := rows.Scan(&x.SuppressionID, &x.ClaimID, &x.Fingerprint, &x.Reason, &x.ApprovedBy,
				&x.ExpiresAt, &x.CreatedAt, &x.RevokedAt, &x.Active); err != nil {
				return err
			}
			out = append(out, x)
		}
		return rows.Err()
	})
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not load suppressions")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"suppressions": out, "count": len(out)})
}

type requestError struct {
	code    string
	message string
}

func (e requestError) Error() string { return e.message }

func errUnprocessable(code, message string) error { return requestError{code: code, message: message} }

func writeCaseError(w http.ResponseWriter, r *http.Request, err error, fallback string) {
	var re requestError
	if errors.As(err, &re) {
		httpx.WriteProblem(w, r, http.StatusUnprocessableEntity, re.code, re.message)
		return
	}
	httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", fallback)
}

func decodeJSON(r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 2<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("request body must contain a single JSON object")
	}
	return nil
}
