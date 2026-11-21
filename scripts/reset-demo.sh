#!/usr/bin/env bash
# Fast data-only reset for the demo stack. Keeps users, the pinned rule
# catalog and the coverage matrix; clears imports and all derived records.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if ! docker compose ps postgres >/dev/null 2>&1; then
  echo "postgres container is not available; run 'make up' first" >&2
  exit 1
fi

docker compose exec -T postgres sh -c 'psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
  < scripts/reset-demo.sql

echo "demo data cleared; users and rule catalog preserved"
