# Runbook — Demo & Local Deployment

Thin-slice stack: Postgres, Kafka, Redis, MinIO, migrations, and eight Go
services. The web queue is optional and runs separately. All demo data is
synthetic.

## Ports

| Component | URL / port |
| --- | --- |
| Identity | http://localhost:8081 |
| Ingestion | http://localhost:8082 |
| Rules | http://localhost:8083 |
| Alert | http://localhost:8084 |
| Case | http://localhost:8085 |
| Audit | http://localhost:8086 |
| Metering | http://localhost:8087 |
| Reporting | http://localhost:8088 |
| Postgres | localhost:5432 |
| Kafka | localhost:9092 |
| MinIO | http://localhost:9000 |
| Web queue (optional) | http://localhost:3000 |

## Prerequisites

- Docker Engine/Desktop with Compose v2 (`docker compose version`)
- Go 1.24+, Node.js 20.9+ (only for the web app)
- `make`, `curl`, `jq`
- ~4 GB free disk for images and volumes

Do not use real customer data. Demo users and datasets are for the thin slice
only; change all passwords before exposing the stack.

## First run

```bash
cp .env.example .env
# edit .env: change SEED_* passwords, POSTGRES_PASSWORD, INTERNAL_TOKEN,
# TENANT_SALT and MinIO credentials
make up
make seed
make datasetgen
```

`make up` builds the services, runs migrations, and waits for dependencies.
`make seed` creates the admin/reviewer/supervisor demo users and the 14-row
coverage matrix. `make datasetgen` writes deterministic CSVs to
`datasets/synthetic/` (ignored by git).

Check readiness:

```bash
make ps
curl -fsS http://localhost:8081/health
for p in 8082 8083 8084 8085 8086 8087 8088; do curl -fsS "http://localhost:$p/health"; done
```

## Demo (≤ 5 minutes)

The scripted path is the acceptance path:

```bash
make demo
```

It performs, in order:

1. Data-only reset (keeps users, rule catalog, coverage matrix).
2. Login as reviewer and admin.
3. Generate and import `datasets/synthetic/*.csv` (multipart, new
   `Idempotency-Key`).
4. Wait for evaluation and alerts; print CLM-1001 flags and output hash.
5. Show the reviewer queue.
6. Create a time-boxed suppression for a known false positive (admin).
7. Create a case, add a synthetic timeline artifact, freeze the evidence
   package and verify its SHA-256, engine version and rule-pack version.
8. Record a CONFIRMED decision with protected value; claim status is unchanged.
9. Verify the audit hash chain and the CLM-1001 trace.
10. Print coverage and dashboard totals.
11. Close metering for the current Asia/Jakarta month, generate the value
    report and export billing twice to prove reproducibility.

To inspect a stage manually, log in and reuse the token:

```bash
TOKEN=$(curl -fsS -X POST http://localhost:8081/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"reviewer@claimshield.local","password":"<SEED_REVIEWER_PASSWORD>"}' | jq -r .token)

curl -fsS -H "Authorization: Bearer $TOKEN" http://localhost:8083/v1/evaluations/CLM-1001 | jq
curl -fsS -H "Authorization: Bearer $TOKEN" 'http://localhost:8084/v1/alerts?limit=10' | jq
curl -fsS -H "Authorization: Bearer $TOKEN" http://localhost:8088/v1/coverage | jq
```

Billing export is admin-only:

```bash
curl -fsS -D - -o /tmp/billing.json -X POST http://localhost:8087/v1/metering/billing-export \
  -H "Authorization: Bearer $ADMIN" -H 'Content-Type: application/json' \
  -d '{"period":"2026-11","version":"v1"}'
```

### Timing budget

| Stage | Typical |
| --- | --- |
| Reset + health | 5–15 s |
| Login + dataset generation | 2–5 s |
| Import | 1–3 s |
| Evaluation + alert propagation | 5–30 s |
| Suppression, case, evidence, decision | 5–10 s |
| Audit trace + coverage/dashboard | 2–5 s |
| Metering close + value + billing | 3–8 s |

The script waits for asynchronous stages instead of using fixed sleeps.

## Reset

Fast data-only reset (keeps users, rule catalog and coverage):

```bash
make reset-demo
```

Full reset (destructive for local volumes only): stops the stack and deletes
Postgres, Kafka, Redis and MinIO data:

```bash
make reset
make up
make seed
```

Re-importing identical files is idempotent by `Idempotency-Key` **and** file
SHA-256. After `make reset-demo`, the import history is gone and the same
dataset can be imported again. Without a reset, a second upload returns the
original batch instead of creating duplicates.

## Troubleshooting

| Symptom | Check / fix |
| --- | --- |
| `make up` fails on port bind | Change `POSTGRES_PORT`, `KAFKA_PORT`, `REDIS_PORT`, `MINIO_PORT` in `.env`; stop the conflicting process. |
| Migrate container exits | `docker compose logs migrate`; fix the reported migration and rerun `make up`. |
| Login returns 401 | Run `make seed`; confirm the password matches `SEED_*` in `.env` and that services were restarted after changing it. |
| Import returns 422 | Inspect `errors` in the response or `GET /v1/imports/{batchID}`; ingestion validates references, dates and hashed identities instead of silently accepting bad rows. |
| No alerts after import | Wait longer (consumer lag) and check `make logs`; confirm rules health on 8083 and Kafka health. |
| `make demo` times out | `docker compose ps`; check service health and logs. Rerun after the stack is healthy. |
| Disk full | Remove old local images/volumes only after confirming they are not needed (`docker system df`). Never delete shared volumes blindly. |

Logs: `make logs` follows the stack. For a single service:
`docker compose logs -f rules`.

## Clean-environment acceptance

From a fresh clone:

```bash
cp .env.example .env   # set strong demo passwords
make up && make seed && make demo
```

Record the final lines (batch id, coverage, metering reconciliation and billing
reproducibility) as the demo evidence. Do not commit `.env` or any generated
dataset.
