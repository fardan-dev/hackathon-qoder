package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/fardanwardhana/qoder-hackathon/packages/contracts"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/app"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/auth"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/db"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/events"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/httpx"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/id"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/kafka"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/outbox"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/pii"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/server"
	"github.com/fardanwardhana/qoder-hackathon/packages/pipeline"
	"github.com/fardanwardhana/qoder-hackathon/packages/rulesengine"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

const maxUpload = 64 << 20

type api struct {
	app *app.App
}

type upload struct {
	dataset  string
	filename string
	data     []byte
	sha      string
}

var errIdempotent = errors.New("idempotent replay")

func main() {
	ctx, cancel := server.SignalContext()
	defer cancel()
	a, err := app.New(ctx, "ingestion")
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
	relay := &outbox.Relay{Pool: a.Pool, Schema: "ingestion", Publisher: producer, Logger: a.Log, PollInterval: time.Second}
	go relay.Run(ctx)

	s := &api{app: a}
	verifier := a.Verifier()
	a.Router.Route("/v1", func(r chi.Router) {
		r.Use(auth.Middleware(verifier))
		r.With(auth.RequireRoles(auth.RoleAdmin, auth.RoleReviewer, auth.RoleSupervisor)).
			Post("/imports", s.importBatch)
		r.Get("/imports", s.listImports)
		r.Get("/imports/{batchID}", s.getImport)
	})
	a.Router.With(httpx.InternalOnly(a.Cfg.InternalToken)).
		Get("/internal/batches/{batchID}/claims", s.batchClaims)
	if err := a.Run(ctx); err != nil {
		a.Log.Error("server_failed", "error", err.Error())
	}
}

func (s *api) importBatch(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "INVALID_UPLOAD", "expected multipart form with dataset files")
		return
	}
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" {
		key = strings.TrimSpace(r.FormValue("idempotency_key"))
	}
	if key == "" {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "REQUIRED_FIELD", "Idempotency-Key is required")
		return
	}
	uploads, err := readUploads(r)
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "INVALID_UPLOAD", err.Error())
		return
	}
	if len(uploads) == 0 {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "EMPTY_FILE", "no dataset files supplied")
		return
	}
	merged, parseErrs := parseUploads(uploads)
	if len(merged.Claims) == 0 && len(parseErrs) == 0 {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "EMPTY_FILE", "claims dataset is required")
		return
	}
	requestHash := hashUploads(uploads, true)
	fileHash := hashUploads(uploads, false)
	batchID := id.New()

	var result batchResult
	var rowErrs []contracts.RowError
	err = db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		existing, found, err := findExistingBatch(r.Context(), tx, p.TenantID, key, fileHash)
		if err != nil {
			return err
		}
		if found {
			result = existing
			return errIdempotent
		}
		if _, err := tx.Exec(r.Context(), `INSERT INTO ingestion.import_batch
			(id, tenant_id, idempotency_key, request_hash, file_name, file_sha256, status)
			VALUES ($1,$2,$3,$4,$5,$6,'PROCESSING')`,
			batchID, p.TenantID, key, requestHash, uploads[0].filename, fileHash); err != nil {
			return err
		}

		rowErrs = append(rowErrs, parseErrs...)
		if len(parseErrs) == 0 {
			errs, err := contracts.Validate(r.Context(), merged, dbLookup{ctx: r.Context(), tx: tx, tenant: p.TenantID})
			if err != nil {
				return err
			}
			rowErrs = append(rowErrs, errs...)
			existingErrs, err := existingKeyErrors(r.Context(), tx, p.TenantID, merged)
			if err != nil {
				return err
			}
			rowErrs = append(rowErrs, existingErrs...)
		}
		if len(rowErrs) > 0 {
			if err := persistErrors(r.Context(), tx, p.TenantID, batchID, rowErrs); err != nil {
				return err
			}
			if _, err := tx.Exec(r.Context(), `UPDATE ingestion.import_batch
				SET status='REJECTED', error_code='VALIDATION_FAILED', total_rows=$2, rejected_rows=$2, completed_at=now()
				WHERE tenant_id=$1 AND id=$3`, p.TenantID, len(rowErrs), batchID); err != nil {
				return err
			}
			result = batchResult{BatchID: batchID, Status: "REJECTED", Errors: rowErrs}
			return nil
		}
		if err := insertBatch(r.Context(), tx, p.TenantID, batchID, merged, p.Email); err != nil {
			return err
		}
		if _, err := tx.Exec(r.Context(), `UPDATE ingestion.import_batch
			SET status='PROCESSED', total_rows=$2, accepted_rows=$2, rejected_rows=0, completed_at=now()
			WHERE tenant_id=$1 AND id=$3`, p.TenantID, len(merged.Claims), batchID); err != nil {
			return err
		}
		for _, c := range merged.Claims {
			if err := outbox.Add(r.Context(), tx, "ingestion", outbox.Event{
				ID: id.New(), TenantID: p.TenantID, Type: pipeline.TopicClaimIngested, AggregateKey: batchID,
				Payload: pipeline.ClaimIngested{
					BatchID: batchID, ClaimID: c.ClaimID, ClaimantID: c.ClaimantID, PolicyID: c.PolicyID,
					ProviderID: c.ProviderID, LossType: c.LossType, ClaimStatus: c.Status,
					ClaimDate: c.ClaimDate.Time, IncidentDate: c.IncidentDate.Time,
					ClaimedAmount: c.ClaimedAmount, Actor: p.Email,
				},
			}); err != nil {
				return err
			}
		}
		if err := outbox.Add(r.Context(), tx, "ingestion", outbox.Event{
			ID: id.New(), TenantID: p.TenantID, Type: pipeline.TopicBatchIngested, AggregateKey: batchID,
			Payload: pipeline.BatchIngested{BatchID: batchID, ClaimCount: len(merged.Claims), Actor: p.Email},
		}); err != nil {
			return err
		}
		result = batchResult{BatchID: batchID, Status: "PROCESSED", Total: len(merged.Claims), Accepted: len(merged.Claims)}
		return nil
	})
	if errors.Is(err, errIdempotent) {
		httpx.WriteJSON(w, http.StatusOK, withIdempotent(result, true))
		return
	}
	if err != nil {
		s.app.Log.Error("import_failed", "request_id", httpx.RequestIDFrom(r.Context()), "error", err.Error())
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "IMPORT_FAILED", "import could not be processed")
		return
	}
	s.app.Log.Info("import_processed", "request_id", httpx.RequestIDFrom(r.Context()),
		"batch_id", result.BatchID, "status", result.Status, "claims", result.Accepted, "errors", len(result.Errors))
	if result.Status == "REJECTED" {
		httpx.WriteJSON(w, http.StatusUnprocessableEntity, withIdempotent(result, false))
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, withIdempotent(result, false))
}

type batchResult struct {
	BatchID  string               `json:"batch_id"`
	Status   string               `json:"status"`
	Total    int                  `json:"total_rows,omitempty"`
	Accepted int                  `json:"accepted_rows,omitempty"`
	Errors   []contracts.RowError `json:"errors,omitempty"`
}

func withIdempotent(r batchResult, v bool) map[string]any {
	return map[string]any{
		"batch_id": r.BatchID, "status": r.Status, "total_rows": r.Total,
		"accepted_rows": r.Accepted, "errors": r.Errors, "idempotent": v,
	}
}

func readUploads(r *http.Request) ([]upload, error) {
	if r.MultipartForm == nil {
		return nil, fmt.Errorf("no multipart form")
	}
	var out []upload
	seen := map[string]bool{}
	for field, headers := range r.MultipartForm.File {
		for _, h := range headers {
			dataset := strings.ToLower(strings.TrimSpace(field))
			switch dataset {
			case contracts.DatasetClaims, contracts.DatasetInvoices, contracts.DatasetPolicies,
				contracts.DatasetProviders, contracts.DatasetIdentities:
			default:
				ds, ok := contracts.DatasetForFile(h.Filename)
				if !ok {
					return nil, fmt.Errorf("cannot determine dataset for %s", h.Filename)
				}
				dataset = ds
			}
			if seen[dataset] {
				return nil, fmt.Errorf("dataset %s supplied more than once", dataset)
			}
			f, err := h.Open()
			if err != nil {
				return nil, err
			}
			data, err := io.ReadAll(io.LimitReader(f, maxUpload))
			_ = f.Close()
			if err != nil {
				return nil, err
			}
			sum := sha256.Sum256(data)
			seen[dataset] = true
			out = append(out, upload{dataset: dataset, filename: h.Filename, data: data, sha: hex.EncodeToString(sum[:])})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].dataset < out[j].dataset })
	return out, nil
}

func parseUploads(uploads []upload) (contracts.Batch, []contracts.RowError) {
	merged := contracts.Batch{Version: contracts.SchemaVersion, Sources: map[string]contracts.Source{}}
	var errs []contracts.RowError
	for _, u := range uploads {
		b, parseErrs := contracts.ParseFile(u.dataset, u.filename, u.data)
		errs = append(errs, parseErrs...)
		merged.Claims = append(merged.Claims, b.Claims...)
		merged.Invoices = append(merged.Invoices, b.Invoices...)
		merged.Policies = append(merged.Policies, b.Policies...)
		merged.Providers = append(merged.Providers, b.Providers...)
		merged.Identities = append(merged.Identities, b.Identities...)
		merged.Sources[u.dataset] = contracts.Source{File: u.filename, SHA256: u.sha, Rows: b.TotalRows()}
	}
	return merged, errs
}

func hashUploads(uploads []upload, names bool) string {
	h := sha256.New()
	for _, u := range uploads {
		if names {
			fmt.Fprintf(h, "%s|%s|%s\n", u.dataset, u.filename, u.sha)
		} else {
			fmt.Fprintf(h, "%s|%s\n", u.dataset, u.sha)
		}
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}

func findExistingBatch(ctx context.Context, tx pgx.Tx, tenant, key, fileHash string) (batchResult, bool, error) {
	var r batchResult
	var total, accepted int
	err := tx.QueryRow(ctx, `SELECT id::text, status, total_rows, accepted_rows FROM ingestion.import_batch
		WHERE tenant_id=$1 AND (idempotency_key=$2 OR file_sha256=$3) LIMIT 1`, tenant, key, fileHash).
		Scan(&r.BatchID, &r.Status, &total, &accepted)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, false, nil
	}
	if err != nil {
		return r, false, err
	}
	r.Total, r.Accepted = total, accepted
	if r.Status == "REJECTED" {
		r.Errors = loadErrors(ctx, tx, tenant, r.BatchID)
	}
	return r, true, nil
}

func persistErrors(ctx context.Context, tx pgx.Tx, tenant, batchID string, errs []contracts.RowError) error {
	if len(errs) > 1000 {
		errs = errs[:1000]
	}
	for _, e := range errs {
		if _, err := tx.Exec(ctx, `INSERT INTO ingestion.import_error
			(error_id, tenant_id, batch_id, dataset, row_number, field_name, code, message)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			id.New(), tenant, batchID, e.Dataset, e.Row, e.Field, e.Code, pii.Redact(e.Message)); err != nil {
			return err
		}
	}
	return nil
}

func loadErrors(ctx context.Context, tx pgx.Tx, tenant, batchID string) []contracts.RowError {
	rows, err := tx.Query(ctx, `SELECT dataset, row_number, coalesce(field_name,''), code, message
		FROM ingestion.import_error WHERE tenant_id=$1 AND batch_id=$2 ORDER BY dataset, row_number`, tenant, batchID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []contracts.RowError
	for rows.Next() {
		var e contracts.RowError
		if err := rows.Scan(&e.Dataset, &e.Row, &e.Field, &e.Code, &e.Message); err == nil {
			out = append(out, e)
		}
	}
	return out
}

type dbLookup struct {
	ctx    context.Context
	tx     pgx.Tx
	tenant string
}

func (l dbLookup) PolicyExists(ctx context.Context, policyID string) (bool, error) {
	return exists(ctx, l.tx, "ingestion.policy_raw", "policy_id", l.tenant, policyID)
}
func (l dbLookup) ProviderExists(ctx context.Context, providerID string) (bool, error) {
	return exists(ctx, l.tx, "ingestion.provider_raw", "provider_id", l.tenant, providerID)
}
func (l dbLookup) InvoiceExists(ctx context.Context, invoiceNumber string) (bool, error) {
	return exists(ctx, l.tx, "ingestion.invoice_raw", "invoice_number", l.tenant, invoiceNumber)
}
func (l dbLookup) ClaimExists(ctx context.Context, claimID string) (bool, error) {
	return exists(ctx, l.tx, "ingestion.claim_raw", "claim_id", l.tenant, claimID)
}

func exists(ctx context.Context, tx pgx.Tx, table, column, tenant, value string) (bool, error) {
	var ok bool
	err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE tenant_id=$1 AND %s=$2)`, table, column), tenant, value).Scan(&ok)
	return ok, err
}

func existingKeyErrors(ctx context.Context, tx pgx.Tx, tenant string, b contracts.Batch) ([]contracts.RowError, error) {
	var out []contracts.RowError
	claimIDs := make([]string, 0, len(b.Claims))
	for _, c := range b.Claims {
		claimIDs = append(claimIDs, c.ClaimID)
	}
	if len(claimIDs) > 0 {
		rows, err := tx.Query(ctx, `SELECT claim_id FROM ingestion.claim_raw WHERE tenant_id=$1 AND claim_id = ANY($2)`, tenant, claimIDs)
		if err != nil {
			return nil, err
		}
		existing := map[string]bool{}
		for rows.Next() {
			var v string
			if err := rows.Scan(&v); err != nil {
				rows.Close()
				return nil, err
			}
			existing[v] = true
		}
		rows.Close()
		for _, c := range b.Claims {
			if existing[c.ClaimID] {
				out = append(out, contracts.RowError{Dataset: contracts.DatasetClaims, Row: c.Row, Field: "claim_id", Code: contracts.CodeDuplicateKey, Message: "claim_id already exists in the tenant"})
			}
		}
	}
	return out, nil
}

func insertBatch(ctx context.Context, tx pgx.Tx, tenant, batchID string, b contracts.Batch, actor string) error {
	for _, p := range b.Providers {
		lineage := lineageJSON(batchID, contracts.DatasetProviders, p.Row, b.Sources)
		if _, err := tx.Exec(ctx, `INSERT INTO ingestion.provider_raw
			(tenant_id, provider_id, name, type, region, batch_id, lineage)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`, tenant, p.ProviderID, p.Name, p.Type, p.Region, batchID, lineage); err != nil {
			return err
		}
	}
	for _, p := range b.Policies {
		lineage := lineageJSON(batchID, contracts.DatasetPolicies, p.Row, b.Sources)
		if _, err := tx.Exec(ctx, `INSERT INTO ingestion.policy_raw
			(tenant_id, policy_id, product, loss_type_scope, inception_date, coverage_added_date,
			 expiry_date, status, sum_insured, premium, reinstatement_date, asset_id, agent_id, batch_id, lineage)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
			tenant, p.PolicyID, p.Product, p.LossTypeScope, p.InceptionDate.Time, dateOrNil(p.CoverageAddedDate),
			p.ExpiryDate.Time, p.Status, p.SumInsured, p.Premium, dateOrNil(p.ReinstatementDate),
			nullString(p.AssetID), nullString(p.AgentID), batchID, lineage); err != nil {
			return err
		}
	}
	for _, in := range b.Invoices {
		lineage := lineageJSON(batchID, contracts.DatasetInvoices, in.Row, b.Sources)
		if _, err := tx.Exec(ctx, `INSERT INTO ingestion.invoice_raw
			(tenant_id, invoice_number, provider_id, invoice_date, invoice_amount, currency, claim_id, doc_sha256, batch_id, lineage)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			tenant, in.InvoiceNumber, in.ProviderID, in.InvoiceDate.Time, in.InvoiceAmount, in.Currency,
			in.ClaimID, nullString(in.DocSHA256), batchID, lineage); err != nil {
			return err
		}
	}
	for _, c := range b.Claims {
		lineage := lineageJSON(batchID, contracts.DatasetClaims, c.Row, b.Sources)
		fp := pii.Fingerprint(tenant, c.ClaimantID, c.IncidentDate.String(), c.ClaimedAmount, c.ProviderID)
		if _, err := tx.Exec(ctx, `INSERT INTO ingestion.claim_raw
			(tenant_id, claim_id, batch_id, policy_id, claimant_id, claim_date, incident_date, loss_type,
			 claimed_amount, currency, provider_id, invoice_number, status, rep_id, fingerprint, lineage)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
			tenant, c.ClaimID, batchID, c.PolicyID, c.ClaimantID, c.ClaimDate.Time, c.IncidentDate.Time,
			c.LossType, c.ClaimedAmount, c.Currency, c.ProviderID, nullString(c.InvoiceNumber),
			c.Status, nullString(c.RepID), fp, lineage); err != nil {
			return err
		}
		for _, field := range []string{"claim_id", "policy_id", "claimant_id", "incident_date", "claimed_amount", "loss_type", "provider_id", "invoice_number", "currency"} {
			if _, err := tx.Exec(ctx, `INSERT INTO ingestion.claim_lineage
				(lineage_id, tenant_id, claim_id, field_name, source, source_row, transform)
				VALUES ($1,$2,$3,$4,$5,$6,$7)`,
				id.New(), tenant, c.ClaimID, field, b.Sources[contracts.DatasetClaims].File, c.Row, "direct"); err != nil {
				return err
			}
		}
	}
	for _, ia := range b.Identities {
		lineage := lineageJSON(batchID, contracts.DatasetIdentities, ia.Row, b.Sources)
		if _, err := tx.Exec(ctx, `INSERT INTO ingestion.identity_map
			(tenant_id, claimant_id, attribute_type, attribute_hash, batch_id, lineage)
			VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (tenant_id, claimant_id, attribute_type) DO UPDATE SET attribute_hash=EXCLUDED.attribute_hash`,
			tenant, ia.ClaimantID, ia.AttributeType, ia.AttributeHash, batchID, lineage); err != nil {
			return err
		}
	}
	_ = actor
	return nil
}

func lineageJSON(batchID, dataset string, row int, sources map[string]contracts.Source) []byte {
	src := sources[dataset]
	raw, _ := events.Canonical(map[string]any{
		"schema_version": contracts.SchemaVersion,
		"batch_id":       batchID,
		"source":         src.File,
		"source_sha256":  src.SHA256,
		"source_row":     row,
		"transform":      "direct",
	})
	return raw
}

func dateOrNil(d *contracts.Date) any {
	if d == nil || d.IsZero() {
		return nil
	}
	return d.Time
}

func nullString(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func (s *api) listImports(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	var out []map[string]any
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT id::text, file_name, status, total_rows, accepted_rows, rejected_rows, created_at, completed_at
			FROM ingestion.import_batch WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT 50`, p.TenantID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var idv, file, status string
			var total, accepted, rejected int
			var created time.Time
			var completed *time.Time
			if err := rows.Scan(&idv, &file, &status, &total, &accepted, &rejected, &created, &completed); err != nil {
				return err
			}
			out = append(out, map[string]any{"batch_id": idv, "file_name": file, "status": status,
				"total_rows": total, "accepted_rows": accepted, "rejected_rows": rejected,
				"created_at": created, "completed_at": completed})
		}
		return rows.Err()
	})
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not list imports")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"imports": out})
}

func (s *api) getImport(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	batchID := chi.URLParam(r, "batchID")
	var file, status, errorCode string
	var total, accepted, rejected int
	var created time.Time
	var completed *time.Time
	var errs []contracts.RowError
	err := db.WithTenant(r.Context(), s.app.Pool, p.TenantID, func(tx pgx.Tx) error {
		err := tx.QueryRow(r.Context(), `SELECT file_name, status, coalesce(error_code,''), total_rows, accepted_rows, rejected_rows, created_at, completed_at
			FROM ingestion.import_batch WHERE tenant_id=$1 AND id=$2`, p.TenantID, batchID).
			Scan(&file, &status, &errorCode, &total, &accepted, &rejected, &created, &completed)
		if err != nil {
			return err
		}
		rows, err := tx.Query(r.Context(), `SELECT dataset, row_number, coalesce(field_name,''), code, message
			FROM ingestion.import_error WHERE tenant_id=$1 AND batch_id=$2 ORDER BY dataset, row_number`, p.TenantID, batchID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var e contracts.RowError
			if err := rows.Scan(&e.Dataset, &e.Row, &e.Field, &e.Code, &e.Message); err != nil {
				return err
			}
			errs = append(errs, e)
		}
		return rows.Err()
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteProblem(w, r, http.StatusNotFound, "NOT_FOUND", "import batch not found")
		return
	}
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "INVALID_BATCH_ID", "invalid batch id")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"batch_id": batchID, "file_name": file, "status": status, "error_code": errorCode,
		"total_rows": total, "accepted_rows": accepted, "rejected_rows": rejected,
		"created_at": created, "completed_at": completed, "errors": errs,
	})
}

func (s *api) batchClaims(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "batchID")
	tenant := strings.TrimSpace(r.URL.Query().Get("tenant_id"))
	if tenant == "" {
		httpx.WriteProblem(w, r, http.StatusBadRequest, "TENANT_REQUIRED", "tenant_id query parameter is required")
		return
	}
	var claims []rulesengine.ClaimSnapshot
	err := db.WithTenant(r.Context(), s.app.Pool, tenant, func(tx pgx.Tx) error {
		var status string
		if err := tx.QueryRow(r.Context(), `SELECT status FROM ingestion.import_batch WHERE tenant_id=$1 AND id=$2`, tenant, batchID).Scan(&status); err != nil {
			return err
		}
		rows, err := tx.Query(r.Context(), `SELECT claim_id, policy_id, claimant_id, provider_id, coalesce(invoice_number,''),
			loss_type, status, coalesce(rep_id,''), claim_date, incident_date, claimed_amount, currency
			FROM ingestion.claim_raw WHERE tenant_id=$1 AND batch_id=$2 ORDER BY claim_id`, tenant, batchID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var c rulesengine.ClaimSnapshot
			if err := rows.Scan(&c.ClaimID, &c.PolicyID, &c.ClaimantID, &c.ProviderID, &c.InvoiceNumber,
				&c.LossType, &c.Status, &c.RepID, &c.ClaimDate, &c.IncidentDate, &c.ClaimedAmount, &c.Currency); err != nil {
				return err
			}
			c.TenantID = tenant
			c.BatchID = batchID
			if err := loadSnapshot(r.Context(), tx, tenant, &c); err != nil {
				return err
			}
			claims = append(claims, c)
		}
		return rows.Err()
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteProblem(w, r, http.StatusNotFound, "NOT_FOUND", "import batch not found")
		return
	}
	if err != nil {
		s.app.Log.Error("batch_claims_failed", "batch_id", batchID, "error", err.Error())
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not load batch claims")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, pipeline.BatchClaims{BatchID: batchID, Claims: claims})
}

func loadSnapshot(ctx context.Context, tx pgx.Tx, tenant string, c *rulesengine.ClaimSnapshot) error {
	var p rulesengine.PolicySnapshot
	var covAdded, reinstatement *time.Time
	err := tx.QueryRow(ctx, `SELECT policy_id, product, loss_type_scope, status, inception_date, coverage_added_date,
		expiry_date, reinstatement_date, sum_insured, premium, coalesce(asset_id,'')
		FROM ingestion.policy_raw WHERE tenant_id=$1 AND policy_id=$2`, tenant, c.PolicyID).
		Scan(&p.PolicyID, &p.Product, &p.LossTypeScope, &p.Status, &p.Inception,
			&covAdded, &p.Expiry, &reinstatement, &p.SumInsured, &p.Premium, &p.AssetID)
	if err == nil {
		p.CoverageAdded, p.Reinstatement = covAdded, reinstatement
		c.Policy = &p
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	var in rulesengine.InvoiceSnapshot
	err = tx.QueryRow(ctx, `SELECT invoice_number, invoice_date, provider_id, invoice_amount, claim_id, coalesce(doc_sha256,'')
		FROM ingestion.invoice_raw WHERE tenant_id=$1 AND invoice_number=$2 AND provider_id=$3`,
		tenant, c.InvoiceNumber, c.ProviderID).
		Scan(&in.Number, &in.Date, &in.ProviderID, &in.Amount, &in.ClaimID, &in.DocSHA256)
	if err == nil {
		c.Invoice = &in
		c.DocSHA256 = in.DocSHA256
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	var pr rulesengine.ProviderSnapshot
	err = tx.QueryRow(ctx, `SELECT provider_id, name, type, region FROM ingestion.provider_raw
		WHERE tenant_id=$1 AND provider_id=$2`, tenant, c.ProviderID).
		Scan(&pr.ProviderID, &pr.Name, &pr.Type, &pr.Region)
	if err == nil {
		c.Provider = &pr
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	rows, err := tx.Query(ctx, `SELECT attribute_type, attribute_hash FROM ingestion.identity_map
		WHERE tenant_id=$1 AND claimant_id=$2`, tenant, c.ClaimantID)
	if err != nil {
		return err
	}
	defer rows.Close()
	c.Identities = map[string]string{}
	for rows.Next() {
		var typ, hash string
		if err := rows.Scan(&typ, &hash); err != nil {
			return err
		}
		c.Identities[typ] = hash
	}
	return rows.Err()
}
