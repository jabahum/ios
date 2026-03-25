#!/usr/bin/env bash
# Applies migration 054: rrt_teams, rrt_deployments, full 029 RRT members/proposals
# stack, proposal trigger, activity_logs. Fixes missing tables / pq relation errors.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SQL="$(cd "$SCRIPT_DIR/.." && pwd)/migrations/054_ensure_activity_logs.sql"
CONN="${DATABASE_URL:-${1:-}}"
[[ -f "$SQL" ]] || { echo "Missing $SQL"; exit 1; }
[[ -n "$CONN" ]] || { echo "Set DATABASE_URL or pass URL as arg"; exit 1; }
psql -v ON_ERROR_STOP=1 "$CONN" -f "$SQL"
echo "OK: 054 applied."
