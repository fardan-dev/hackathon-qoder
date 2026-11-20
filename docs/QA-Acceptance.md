# QA Acceptance Matrix — Thin Slice MVP

Scope: Phase 0/1 thin slice for the P&C claim-fraud detection platform. Rule pack
`2026.11.1`, engine `rulesengine-1.0.0` (43 rules: 33 active A–G, 10 gated-off H).
Document-forensics rules stay gated off; AC-9/11/12 and entitlement (AC-15) are
recorded as N/A with residual risk, not silently skipped.

Evidence below is reproducible from a clean checkout:

```bash
go build ./... && go vet ./... && go test -race ./...
make fp-gate
cd apps/web && npm run lint && npm run typecheck && npm test && npm run build
```

The false-positive gate writes `backtest/fp-gate-report.json` (also uploaded by CI).

## Execution record

Verified in the current working tree on 2026-11-20:

- `go build ./...`, `go vet ./...`: pass.
- `go test -race ./...`: pass (all packages).
- `make fp-gate`: pass; 47 claims, 33/33 positives flagged (recall 10000 bps),
  14 controls, 35 evaluated rules, 0 violations, 8 H rules excluded from the
  gate and H9/H10 checked against the experimental threshold.
  `dataset_sha256=sha256:b8ee7d8956582bc2dcea23443b00c069293edb6f1af4a203ccccbc2f232efafa`,
  `output_hash=sha256:cd5c1af6e9707ad15310551640bcd434656ca8a6761a772c6d4662ed32994daf`,
  pack `2026.11.1`, engine `rulesengine-1.0.0`.
- Web: lint, typecheck, 44/44 Vitest tests (including the UI copy guardrail)
  and `next build` pass.
- Clean-environment acceptance passes: `docker compose down -v`, then
  `make up`, `make seed`, `make demo` on fresh volumes. 47 claims accepted,
  CLM-1001 flags A3/A6/B4, audit chain valid over 128 events, dashboard
  decisions=1, value report `AVS-v1` with value protected 20000000,
  metering reconciliation diff 0, billing export reproducible.

## Test plan

1. Unit/integration packages: contracts, engine determinism and label isolation,
   alert/case/audit/metering helpers.
2. Backtest gate: synthetic positives plus clean controls, per-rule FP thresholds
   by BRD priority; violations fail CI.
3. API walkthrough: `scripts/demo.sh` exercises ingest → evaluate → alert →
   case → evidence/package → decision → audit → coverage/dashboard → metering.
4. Security review: authz tests, RLS/tenant guard inspection, PII redaction
   tests, secret-scan CI job.
5. UI copy guard: `src/lib/copy-guard.test.ts` fails if conclusion wording
   ("terbukti", "palsu", "fraud") appears in web source.

## Acceptance criteria

| AC | Result | Evidence |
| --- | --- | --- |
| AC-1 Explain each flag | Pass | Every `rulesengine.Flag` carries `rule_id`, `rule_version`, `trigger_value`, `evidence_refs` and `explanation` (`packages/rulesengine/types.go`); `TestEngineFlagsAndFourElements` asserts flag shape and rule expectations; `GET /v1/evaluations/{claimID}` and alert bundles expose the same fields. |
| AC-2 Human confirm/clear | Pass | Decisions are explicit records (`POST /v1/cases/{caseID}/decisions`); `CLEARED` requires an `FP_`/`NI_`/`CF_` reason code and `CONFIRMED` requires notes (`TestValidateDecision`). Alerts close only after a decision event (`services/alert/cmd/server/main.go`), never on evaluation alone. |
| AC-3 No automatic rejection | Pass | The engine returns flags only; no service or endpoint updates `ingestion.claim_raw.status`. Claim status in the source CSV is preserved through import and evaluation (verified by code inspection). |
| AC-4 Planted-case recall, control FP gate | Pass | `TestControlClaimFalsePositiveGate`: 47 synthetic claims (33 positives, 14 controls), case recall 10000 bps, zero Must-rule false positives on controls; violations fail the test. Artifact: `backtest/fp-gate-report.json`. |
| AC-5 Cross-batch detection | Pass | `TestCrossBatchHistoryFlags` loads prior history: old invoice triggers A1, prior CONFIRMED triggers G1, prior claim enters C4. |
| AC-6 Coverage matrix + residual register | Pass | Seed loads 14 typologies (10 covered, 4 residual) in `cmd/seed/main.go`; `GET /v1/coverage` returns `summary` and `residual_risk_register`; `GET /v1/coverage/export` emits CSV. The web `/coverage` page shows the same matrix read-only with a CSV download. |
| AC-7 Determinism, diff 0 | Pass | `TestEngineDeterministic` (repeat + reordered input), `TestEngineOutputIndependentOfGroundTruth` (same output hash with labels removed); FP gate asserts stable dataset/output hashes. |
| AC-8 Decisions and suppression audited | Pass | Decision, package, suppression approve/revoke events flow through the outbox to the hash-chained audit log; `GET /v1/audit/verify` and `/v1/audit/trace/{claimID}` expose the chain. `TestPackageHashDetectsTampering`, `TestSuppressionApprovalRoleGate`. |
| AC-9 Document manipulation detection | N/A | H-family forensics are gated off in the thin slice; residual risk recorded in the coverage matrix. No claim of coverage. |
| AC-10 Chain of custody | Pass (thin slice) | Evidence artifacts store SHA-256; packages store `package_sha256`, `artifact_count`, `engine_version` and `rule_pack_version` (`db/migrations/0004_evidence_package_versions.sql`, `services/case`); `GET .../package/verify` recomputes the hash. Forensic report content is N/A because H rules are off; uploaded-document hashing is out of scope (no document-upload endpoint). |
| AC-11 OCR failure not auto-cleared | N/A | No OCR pipeline in the thin slice; H7 is gated off. Rows with unparseable data fail import instead of being cleared. |
| AC-12 Document false-positive control | N/A | No document analyzer/EXIF path in the thin slice; H-family gated off. |
| AC-13 Metering accuracy | Pass | Usage events are deduplicated per source event (`metering.processed_event`, unique `(tenant, source_event_id, metric)`); `GET /v1/metering/usage` reports `reconciliation_diff`; close stores the metric checksum and refuses recomputation drift by reuse. Demo asserts `reconciliation_diff=0`. |
| AC-14 Value report | Pass | `POST /v1/metering/value-report` stores formula version, decision counts, confirmed/cleared, value protected/recovered; payload contains no claim IDs or PII (aggregates only). |
| AC-15 Entitlement / volume bands | N/A | Commercial entitlement is out of the thin-slice build; metering exposes usage only and does not silently block. Recorded as residual product risk. |
| AC-16 Billing export | Pass | `POST /v1/metering/billing-export` returns a checksummed, versioned JSON export; re-export returns the stored payload (same checksum). `TestBillingPayloadCanonical` covers canonicalization; demo compares two exports byte for byte. |

## Determinism and backtest evidence

- Dataset generation is deterministic (`TestDatasetIsDeterministicAndValid`); the
  FP gate records `dataset_sha256`, `ground_truth_sha256` and evaluation
  `output_hash` in `backtest/fp-gate-report.json`.
- Ground truth is never an engine input: `TestLabelIsolationBoundary` and the
  gate's column-forbidding check fail if label/typology columns leak in.
- Rule-pack version is pinned; `TestRulePackVersionPinned` rejects an empty or
  mismatched version.

## Security checklist

| Area | Status | Evidence |
| --- | --- | --- |
| Authentication | Pass | Session tokens are random 256-bit values stored only as SHA-256 hashes; HttpOnly/SameSite cookie; login rate-limit 5/minute per IP+email (`services/identity`). |
| Authorization | Pass | Reviewer/admin/supervisor gates on imports, audit, metering close/export and suppression; `TestUnauthenticatedIsRejected`, `TestSuppressionApprovalRoleGate`. |
| Tenant isolation | Pass | RLS `ENABLE`/`FORCE` on every tenant table plus `db.WithTenant` transaction guard; internal endpoints require `INTERNAL_TOKEN`. |
| SQL injection | Pass | All user values are pgx parameters; the only dynamic column selector (alert load) is an internal constant, not request input. |
| Secrets | Pass | `.env` ignored; `.env.example` placeholders only; seed passwords require `SEED_*` env; CI secret-scan job. |
| PII in logs/exports | Pass | Identity attributes are salted hashes; `TestLoggerNeverWritesRawPIIWhenUsingRedact`, `TestRedact`; value/billing exports contain aggregates only. |
| Transport | Partial | Local Compose is HTTP; TLS is expected at the ingress. Not verified in this environment. |

Critical findings: **0**. No open high-severity finding in the reviewed scope.

## Residual risks (ranked)

1. The web queue and case pages read seeded mock data, not the live alert/case
   APIs; the API path is verified by `scripts/demo.sh` instead.
2. H-family document forensics are off (AC-9/11/12); the coverage matrix records
   this residual risk explicitly.
3. No load/soak test was run; rate limiting is in-process per replica, so
   horizontal scaling would need a shared limiter.
4. No external penetration test; the security review above is code-level.
5. Entitlement enforcement (AC-15) is not implemented in the thin slice.
