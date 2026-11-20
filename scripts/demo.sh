#!/usr/bin/env bash
# Five-minute demo: reset data, import the synthetic batch, triage one alert,
# freeze evidence, decide, then reconcile audit, coverage and metering.
# All data is synthetic; no real customer PII is used.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  . ./.env
  set +a
fi

IDENTITY="http://localhost:8081"
INGESTION="http://localhost:8082"
RULES="http://localhost:8083"
ALERT="http://localhost:8084"
CASE="http://localhost:8085"
AUDIT="http://localhost:8086"
METERING="http://localhost:8087"
REPORTING="http://localhost:8088"

for bin in curl jq; do
  command -v "$bin" >/dev/null || { echo "missing prerequisite: $bin" >&2; exit 1; }
done

step() { printf '\n== %s ==\n' "$*"; }
get()  { curl -fsS -H "Authorization: Bearer $1" "$2"; }
post() { curl -fsS -X POST -H "Authorization: Bearer $1" -H 'Content-Type: application/json' -d "$3" "$2"; }

post_until() {
  local url=$1 token=$2 body=$3 expect=$4 out=$5 tries=${6:-30} code=""
  for _ in $(seq 1 "$tries"); do
    code=$(curl -sS -o "$out" -w '%{http_code}' -X POST -H "Authorization: Bearer $token" \
      -H 'Content-Type: application/json' -d "$body" "$url")
    [[ $code == "$expect" ]] && return 0
    sleep 1
  done
  echo "POST $url returned $code, expected $expect" >&2
  cat "$out" >&2
  return 1
}

wait_for() {
  local url=$1 name=$2 tries=${3:-60}
  for _ in $(seq 1 "$tries"); do
    if curl -fsS "$url" >/dev/null 2>&1; then return 0; fi
    sleep 1
  done
  echo "timed out waiting for $name ($url)" >&2
  return 1
}

wait_http() {
  local url=$1 token=$2 tries=${3:-60}
  for _ in $(seq 1 "$tries"); do
    if [[ $(curl -sS -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $token" "$url") == 200 ]]; then
      return 0
    fi
    sleep 1
  done
  echo "timed out waiting for $url" >&2
  return 1
}

login() {
  local email=$1 pass=$2
  curl -fsS -X POST "$IDENTITY/v1/auth/login" -H 'Content-Type: application/json' \
    -d "$(jq -n --arg e "$email" --arg p "$pass" '{email:$e,password:$p}')" | jq -r .token
}

: "${SEED_ADMIN_PASSWORD:?SEED_ADMIN_PASSWORD is required (see .env.example)}"
: "${SEED_REVIEWER_PASSWORD:?SEED_REVIEWER_PASSWORD is required (see .env.example)}"

if [[ "${SKIP_RESET:-0}" != "1" ]]; then
  step "Reset demo data"
  scripts/reset-demo.sh
fi

step "Wait for services"
for port in 8081 8082 8083 8084 8085 8086 8087 8088; do
  wait_for "http://localhost:$port/health" "service on $port"
done

step "Login"
ADMIN=$(login admin@claimshield.local "$SEED_ADMIN_PASSWORD")
REVIEWER=$(login reviewer@claimshield.local "$SEED_REVIEWER_PASSWORD")
echo "admin and reviewer sessions created"

step "Generate synthetic dataset"
make datasetgen
DATA=datasets/synthetic

step "Import batch"
IMPORT=$(curl -fsS -X POST "$INGESTION/v1/imports" \
  -H "Authorization: Bearer $REVIEWER" \
  -H "Idempotency-Key: demo-$(date +%s)" \
  -F "claims=@$DATA/claims.csv" \
  -F "invoices=@$DATA/invoices.csv" \
  -F "policies=@$DATA/policies.csv" \
  -F "providers=@$DATA/providers.csv" \
  -F "identities=@$DATA/identities.csv")
BATCH_ID=$(jq -r .batch_id <<<"$IMPORT")
echo "batch $BATCH_ID accepted=$(jq -r .accepted_rows <<<"$IMPORT") rows"

step "Wait for rule evaluation and alerting"
wait_http "$RULES/v1/evaluations/CLM-1001" "$REVIEWER" 90
wait_http "$ALERT/v1/alerts/by-claim/CLM-1001" "$REVIEWER" 60
EVAL=$(get "$REVIEWER" "$RULES/v1/evaluations/CLM-1001")
echo "CLM-1001 flags: $(jq -c '[.flags[].rule_id]' <<<"$EVAL")"
echo "output hash: $(jq -r .output_hash <<<"$EVAL")"
QUEUE=$(get "$REVIEWER" "$ALERT/v1/alerts?limit=5")
echo "top queue alerts: $(jq -r '.count' <<<"$QUEUE") returned"
jq -r '.alerts[] | "  \(.claim_id) \(.severity) flags=\(.flag_count) amount=\(.claimed_amount)"' <<<"$QUEUE"

step "Suppress a known false positive"
EXPIRES=$(jq -nr 'now + 2592000 | todate')
post_until "$CASE/v1/suppressions" "$ADMIN" \
  "$(jq -n --arg c "CLM-1002" --arg r "Known legitimate resubmission (demo)" --arg e "$EXPIRES" \
    '{claim_id:$c,reason:$r,expires_at:$e}')" 201 /tmp/claimshield-suppression.json
SUPPRESSION=$(cat /tmp/claimshield-suppression.json)
echo "suppression $(jq -r .suppression_id <<<"$SUPPRESSION") active until $EXPIRES"

step "Create case and freeze evidence"
post_until "$CASE/v1/cases" "$REVIEWER" \
  "$(jq -n '{title:"Invoice reuse investigation",summary:"Synthetic demo case: shared invoice number across two providers.",priority:1,claim_ids:["CLM-1001"]}')" \
  201 /tmp/claimshield-case.json
CASE_JSON=$(cat /tmp/claimshield-case.json)
CASE_ID=$(jq -r .case_id <<<"$CASE_JSON")
post_until "$CASE/v1/cases/$CASE_ID/evidence" "$REVIEWER" \
  '{"artifact_type":"TIMELINE","source":"demo-script","content":"Synthetic timeline for CLM-1001. No customer data.","chain_of_custody":{"collected_by":"demo-script","source_system":"synthetic dataset"}}' \
  201 /tmp/claimshield-evidence.json
PACKAGE=$(post "$REVIEWER" "$CASE/v1/cases/$CASE_ID/package" '{}')
echo "package $(jq -r .package.package_sha256 <<<"$PACKAGE") engine=$(jq -r .package.engine_version <<<"$PACKAGE") rules=$(jq -r .package.rule_pack_version <<<"$PACKAGE")"
VERIFY=$(get "$REVIEWER" "$CASE/v1/cases/$CASE_ID/package/verify")
echo "package verify valid=$(jq -r .valid <<<"$VERIFY") artifacts=$(jq -r .artifact_count <<<"$VERIFY")"

step "Record reviewer decision"
DECISION=$(post "$REVIEWER" "$CASE/v1/cases/$CASE_ID/decisions" \
  '{"outcome":"CONFIRMED","reason_code":"INVOICE_REUSE_CONFIRMED","notes":"Invoice number reused across two providers in the synthetic batch.","value_protected":20000000,"value_recovered":0}')
echo "decision $(jq -r .outcome <<<"$DECISION") on case $CASE_ID (claim status unchanged)"

step "Verify audit chain and claim trace"
wait_http "$AUDIT/v1/audit/trace/CLM-1001" "$ADMIN" 60
for _ in $(seq 1 30); do
  TRACE=$(get "$ADMIN" "$AUDIT/v1/audit/trace/CLM-1001")
  [[ $(jq -r '.events | length' <<<"$TRACE") -ge 3 ]] && break
  sleep 1
done
AUDIT_VERIFY=$(get "$ADMIN" "$AUDIT/v1/audit/verify")
echo "audit chain valid=$(jq -r .valid <<<"$AUDIT_VERIFY") events=$(jq -r .checked <<<"$AUDIT_VERIFY")"
echo "CLM-1001 trace events: $(jq -r '.events | length' <<<"$TRACE") stages: $(jq -c .stages <<<"$TRACE")"

step "Coverage and dashboard"
COVERAGE=$(get "$REVIEWER" "$REPORTING/v1/coverage")
echo "coverage typologies=$(jq -r '.typologies | length' <<<"$COVERAGE") covered=$(jq -r '[.typologies[] | select(.phase01_covered)] | length' <<<"$COVERAGE")"
DASHBOARD=$(get "$REVIEWER" "$REPORTING/v1/dashboard")
echo "dashboard alerts=$(jq -r '.alerts.total' <<<"$DASHBOARD") decisions=$(jq -r '.decisions.total // 0' <<<"$DASHBOARD")"

step "Close metering and export billing"
PERIOD=$(TZ=Asia/Jakarta date +%Y-%m)
for _ in $(seq 1 30); do
  RECORDED=$(get "$ADMIN" "$METERING/v1/metering/usage?period=$PERIOD" | jq -r '.metrics.decisions_recorded // 0')
  [[ $RECORDED -ge 1 ]] && break
  sleep 1
done
CLOSE=$(post "$ADMIN" "$METERING/v1/metering/close" "$(jq -n --arg p "$PERIOD" '{period:$p}')")
echo "usage period $PERIOD reconciliation_diff=$(jq -r .usage_record.reconciliation_diff <<<"$CLOSE") checksum=$(jq -r .usage_record.checksum <<<"$CLOSE")"
VALUE=$(post "$ADMIN" "$METERING/v1/metering/value-report" "$(jq -n --arg p "$PERIOD" '{period:$p}')")
echo "value report decisions=$(jq -r .decisions <<<"$VALUE") value_protected=$(jq -r .value_protected <<<"$VALUE") formula=$(jq -r .formula_version <<<"$VALUE")"

H1=/tmp/claimshield-billing-1.json
H2=/tmp/claimshield-billing-2.json
curl -fsS -D /dev/null -o "$H1" -X POST "$METERING/v1/metering/billing-export" \
  -H "Authorization: Bearer $ADMIN" -H 'Content-Type: application/json' \
  -d "$(jq -n --arg p "$PERIOD" '{period:$p,version:"v1"}')"
curl -fsS -D /dev/null -o "$H2" -X POST "$METERING/v1/metering/billing-export" \
  -H "Authorization: Bearer $ADMIN" -H 'Content-Type: application/json' \
  -d "$(jq -n --arg p "$PERIOD" '{period:$p,version:"v1"}')"
if cmp -s "$H1" "$H2"; then
  echo "billing export reproducible: identical payload on re-export"
else
  echo "billing export differs between runs" >&2
  exit 1
fi

step "Done"
echo "demo completed for batch $BATCH_ID"
