# Applies migrations/052_rrt_deployment_proposals_and_activity_logs.sql in one shot.
#
# Usage:
#   $env:DATABASE_URL = "postgres://user:pass@host:5432/ios"
#   .\scripts\apply-migration-052.ps1
#
# Or:
#   .\scripts\apply-migration-052.ps1 "postgres://user:pass@localhost:5432/ios"

$ErrorActionPreference = "Stop"
$RepoRoot = Split-Path -Parent $PSScriptRoot
$Sql = Join-Path $RepoRoot "migrations\052_rrt_deployment_proposals_and_activity_logs.sql"

if (-not (Test-Path -LiteralPath $Sql)) {
    Write-Error "Migration file not found: $Sql"
}

$conn = $env:DATABASE_URL
if (-not $conn -and $args.Count -ge 1) { $conn = $args[0] }
if (-not $conn) {
    Write-Error "Set DATABASE_URL or pass connection URL as first argument."
}

Write-Host "Applying: $Sql"
& psql -v ON_ERROR_STOP=1 $conn -f $Sql
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "OK: migration 052 applied."
