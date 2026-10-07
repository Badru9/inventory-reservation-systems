# INDICO Flash-Sale — PowerShell wrapper around the Makefile for Windows users
# who don't have make installed.
#
# Usage:
#   .\scripts\dev.ps1 help
#   .\scripts\dev.ps1 up
#   .\scripts\dev.ps1 db
#   .\scripts\dev.ps1 backend
#   .\scripts\dev.ps1 frontend
#   .\scripts\dev.ps1 test
#   .\scripts\dev.ps1 down

[CmdletBinding()]
param(
    [Parameter(Position=0)]
    [string]$Target = "help"
)

$RepoRoot = Split-Path -Parent $PSScriptRoot
$Backend  = Join-Path $RepoRoot "backend"
$Frontend = Join-Path $RepoRoot "frontend"
$Env:PORT = "8080"
$Env:DATABASE_URL = "postgres://indico:indico@127.0.0.1:5432/indico?sslmode=disable"
$Env:PATH = "C:\Users\RIVAN\scoop\apps\mingw\current\bin;$Env:PATH"

function Run($cmd, $cwd = $RepoRoot) {
    Write-Host ">> $cmd" -ForegroundColor Cyan
    Push-Location $cwd
    try { & cmd.exe /c $cmd } finally { Pop-Location }
}

switch ($Target) {
    "help" {
        @"
INDICO Flash-Sale - dev.ps1 targets

  up             start postgres + backend + frontend (Docker)
  down           stop the stack
  clean          stop and wipe Postgres data
  db             start only Postgres
  backend        run the Go backend natively (requires Postgres)
  frontend       run the Vite dev server
  test           run all tests (race detector)
  test-unit      unit tests only (no DB)
  test-integ     integration + stress tests against real Postgres
  stress         run only the no-overselling stress test (host)
  stress-docker  run the stress test inside a one-shot container
  test-docker    run the full test suite inside a one-shot container
  logs           tail container logs
  psql           open psql in the postgres container

Examples:
  .\scripts\dev.ps1 up
  .\scripts\dev.ps1 test
"@
    }
    "up"        { Run "docker compose up -d --build" }
    "down"      { Run "docker compose down" }
    "clean"     { Run "docker compose down -v" }
    "db"        { Run "docker compose up -d postgres" }
    "backend"   {
        Run "docker compose up -d postgres"
        Run "go run ./cmd/server" $Backend
    }
    "frontend"  {
        if (-not (Test-Path (Join-Path $Frontend "node_modules"))) {
            Run "npm install" $Frontend
        }
        Run "npm run dev" $Frontend
    }
    "test"      {
        Run "go test -race -v -count=1 ./internal/service/..." $Backend
        Run "docker compose up -d postgres" | Out-Null
        Run "go test -race -v -count=1 ./internal/repository/..." $Backend
    }
    "test-unit" { Run "go test -race -v -count=1 ./internal/service/..." $Backend }
    "test-integ" {
        Run "docker compose up -d postgres" | Out-Null
        Run "go test -race -v -count=1 ./internal/repository/..." $Backend
    }
    "stress"    {
        Run "docker compose up -d postgres" | Out-Null
        Run "go test -race -v -count=1 -run TestStress_NoOverselling ./internal/repository/..." $Backend
    }
    "stress-docker" {
        Run "docker compose up -d postgres" | Out-Null
        Run "docker run --rm --network indico_default -v `"$RepoRoot\backend:/src`" -w /src golang:1.27-alpine sh -c `"apk add --no-cache gcc musl-dev >/dev/null && DATABASE_URL='postgres://indico:indico@indico_postgres:5432/indico?sslmode=disable' go test -v -count=1 -run TestStress_NoOverselling ./internal/repository/...`""
    }
    "test-docker" {
        Run "docker compose up -d postgres" | Out-Null
        Run "docker run --rm --network indico_default -v `"$RepoRoot\backend:/src`" -w /src golang:1.27-alpine sh -c `"apk add --no-cache gcc musl-dev >/dev/null && DATABASE_URL='postgres://indico:indico@indico_postgres:5432/indico?sslmode=disable' go test -v -count=1 ./...`""
    }
    "logs"      { Run "docker compose logs -f" }
    "psql"      { Run "docker exec -it indico_postgres psql -U indico -d indico" }
    default {
        Write-Host "Unknown target '$Target'. Run '.\scripts\dev.ps1 help'." -ForegroundColor Yellow
        exit 1
    }
}
