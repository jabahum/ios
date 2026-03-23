# Regenerate OpenAPI docs into cmd/web/docs (what the app embeds via _ "case/cmd/web/docs").
# Requires: go install github.com/swaggo/swag/cmd/swag@latest
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
Set-Location (Split-Path -Parent $PSScriptRoot)

swag init -g docs.go -o cmd/web/docs -d cmd/web,internal/handlers

# CLI swag may emit LeftDelim/RightDelim; project embeds github.com/swaggo/swag v1.8.1 (Spec without those fields).
$dg = Join-Path (Get-Location) "cmd/web/docs/docs.go"
(Get-Content -LiteralPath $dg) | Where-Object { $_ -notmatch '^\s*LeftDelim:' -and $_ -notmatch '^\s*RightDelim:' } | Set-Content -LiteralPath $dg

Write-Host "OK: $dg, cmd/web/docs/swagger.json, swagger.yaml"
