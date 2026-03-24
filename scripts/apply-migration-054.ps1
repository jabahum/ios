# Creates activity_logs (+ rrt_teams / rrt_deployments if missing)
$ErrorActionPreference = "Stop"
$Repo = Split-Path -Parent $PSScriptRoot
$Sql = Join-Path $Repo "migrations\054_ensure_activity_logs.sql"
if (-not (Test-Path -LiteralPath $Sql)) { Write-Error "Missing $Sql" }

$conn = $env:DATABASE_URL
if (-not $conn -and $args.Count -ge 1) { $conn = $args[0] }
if (-not $conn) { Write-Error "Set DATABASE_URL or pass connection URL" }

& psql -v ON_ERROR_STOP=1 $conn -f $Sql
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "OK: 054 applied."
