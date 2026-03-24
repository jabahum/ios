#!/usr/bin/env bash
# Applies migrations/052_rrt_deployment_proposals_and_activity_logs.sql in one shot.
# Avoids copy-paste mistakes (pasting only the trigger / activity_logs tail).
#
# Usage:
#   export DATABASE_URL="postgres://user:pass@host:5432/ios"
#   ./scripts/apply-migration-052.sh
#
# Or pass the connection URL as the first argument:
#   ./scripts/apply-migration-052.sh "postgres://user:pass@localhost:5432/ios"

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SQL="$REPO_ROOT/migrations/052_rrt_deployment_proposals_and_activity_logs.sql"

if [[ ! -f "$SQL" ]]; then
  echo "ERROR: migration file not found: $SQL" >&2
  exit 1
fi

CONN="${DATABASE_URL:-${1:-}}"
if [[ -z "$CONN" ]]; then
  echo "ERROR: set DATABASE_URL or pass connection URL as first argument." >&2
  echo "  export DATABASE_URL=postgres://..." >&2
  echo "  $0" >&2
  exit 1
fi

echo "Applying: $SQL"
psql -v ON_ERROR_STOP=1 "$CONN" -f "$SQL"
echo "OK: migration 052 applied."
