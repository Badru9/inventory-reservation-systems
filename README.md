# INDICO Flash-Sale Take-Home

High-concurrency inventory reservation system built with **Go + Gin + PostgreSQL** (backend, 80%) and **React + Vite** (frontend, 20%).

## Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.27, Gin, pgx/v5, embedded SQL migrations |
| Database | PostgreSQL 16 |
| Frontend | React 19, Vite 8, TypeScript, TanStack Query, axios |
| Orchestration | Docker Compose — one command spins up the full stack |

## Endpoints

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/api/v1/inventory/reserve` | Atomically reserve stock (5-minute TTL) |
| `POST` | `/api/v1/inventory/confirm` | Commit an active reservation |
| `GET`  | `/api/v1/inventory/stock?item_id=…` | Real-time stock breakdown |
| `GET`  | `/healthz` | Liveness |

## TL;DR — three commands

```bash
# 1. Start the full stack (postgres + backend + frontend)
docker compose up -d --build

# 2. Open the dashboard
#    Frontend → http://localhost:5173
#    Backend  → http://localhost:8080/healthz

# 3. Run the test suite (unit + integration + 300-goroutine stress)
docker compose up -d postgres
docker run --rm --network indico_default -v "%cd%/backend:/src" -w /src golang:1.27-alpine ^
  sh -c "apk add --no-cache gcc musl-dev >/dev/null && DATABASE_URL='postgres://indico:indico@indico_postgres:5432/indico?sslmode=disable' go test -v -count=1 ./..."
```

## Prerequisites

| Tool | Why | Install hint |
|---|---|---|
| **Docker Desktop** (with WSL2 on Windows) | Runs Postgres + the backend & frontend | <https://docker.com/products/docker-desktop> |
| **Go 1.27+** | Backend dev loop and tests (only if you want to run things natively) | `scoop install go` |
| **Node 20+** | Frontend dev loop with Vite | <https://nodejs.org> |
| **MinGW** (Windows only, optional) | `go test -race` needs CGO | `scoop install mingw` |

## Running the backend

You have two options — the Docker path is the easiest, the native path is the fastest for a tight edit-run loop.

### Option A — Docker (recommended for first run)

```bash
docker compose up -d --build
```

This brings up three containers:
- `indico_postgres` — PostgreSQL 16 on `localhost:5432`
- `indico_backend`  — Gin server on `localhost:8080`
- `indico_frontend` — nginx serving the Vite build on `localhost:5173`

Tail the logs:

```bash
docker compose logs -f backend
```

Stop the stack:

```bash
docker compose down               # keep Postgres data
docker compose down -v            # also wipe the Postgres volume
```

### Option B — backend native, Postgres in Docker

This is the fastest dev loop — edits to Go files are picked up by `go run` on the next restart, and you skip rebuilding a container.

**1. Start only Postgres:**

```bash
docker compose up -d postgres
```

Wait until it's healthy (`docker ps` shows `(healthy)`).

**2. Start the Go backend:**

```bash
cd backend
go run ./cmd/server
```

You should see:
```
{"time":"...","level":"INFO","msg":"migrations applied"}
{"time":"...","level":"INFO","msg":"expiry sweeper started","interval":"15s"}
{"time":"...","level":"INFO","msg":"server listening","addr":":8080"}
```

The server reads `DATABASE_URL` and `PORT` from the environment. Defaults:

```
DATABASE_URL=postgres://indico:indico@127.0.0.1:5432/indico?sslmode=disable
PORT=8080
```

## Running the frontend

### Option A — Docker

Already running as part of `docker compose up -d --build`. Open <http://localhost:5173>.

### Option B — Vite dev server (faster HMR)

```bash
cd frontend
npm install
npm run dev
```

Open <http://localhost:5173>. Vite reads `VITE_API_BASE_URL` from `frontend/.env.development` (already set to `http://localhost:8080`).

## Trying the dashboard

1. Open <http://localhost:5173>.
2. The **Live Inventory** card polls every 3 s — flip between `item_4021` and `item_7777` to see different items.
3. Fill the **Reserve Stock** form (e.g. `usr_9981`, `item_4021`, qty `2`) → click **Reserve**.
4. The right panel shows your reservation, a live 5-minute countdown, and a **Confirm Purchase** button.
5. Click **Confirm Purchase** — the reservation is committed and the available stock returns to baseline.

To watch the expiry sweep at work, reserve a quantity and then leave the page open past the 5-minute mark. The countdown turns red, the reservation is auto-cleared, and `/stock` shows the units back in `available_stock`.

## Testing the backend

The test suite has two layers: **unit tests** (no DB needed) and **integration + stress tests** (real Postgres).

### Unit tests (native)

```bash
cd backend
go test -race -v -count=1 ./internal/service/...
```

On Windows, `go test -race` needs a C compiler. If you have MinGW:

```bash
set PATH=C:\Users\RIVAN\scoop\apps\mingw\current\bin;%PATH%
go test -race -v -count=1 ./internal/service/...
```

### Integration + stress tests (native)

```bash
docker compose up -d postgres
cd backend
set DATABASE_URL=postgres://indico:indico@127.0.0.1:5432/indico?sslmode=disable
go test -v -count=1 ./internal/repository/...
```

> **Windows + Docker Desktop caveat:** host connections to `127.0.0.1:5432` are routed through the WSL2 NAT layer, which Postgres sees as a non-loopback IP and rejects with `password authentication failed`. The unit tests still work because they don't need a real DB, but the integration tests should be run via the Docker path below when you're on Windows.

### Integration + stress tests (Docker — works on every OS)

This runs the test binary inside a one-shot Go container attached to the `indico_default` network, so the connection comes from a real docker IP and the trust `pg_hba.conf` rule applies.

**Linux / macOS / Git Bash:**

```bash
docker compose up -d postgres
docker run --rm --network indico_default \
  -v "$PWD/backend:/src" -w /src golang:1.27-alpine \
  sh -c "apk add --no-cache gcc musl-dev >/dev/null && \
         DATABASE_URL='postgres://indico:indico@indico_postgres:5432/indico?sslmode=disable' \
         go test -v -count=1 ./..."
```

**Windows (cmd):**

```cmd
docker compose up -d postgres
docker run --rm --network indico_default -v "%cd%\backend:/src" -w /src golang:1.27-alpine ^
  sh -c "apk add --no-cache gcc musl-dev >/dev/null && DATABASE_URL='postgres://indico:indico@indico_postgres:5432/indico?sslmode=disable' go test -v -count=1 ./..."
```

**Windows (PowerShell):**

```powershell
docker compose up -d postgres
docker run --rm --network indico_default -v "${PWD}/backend:/src" -w /src golang:1.27-alpine `
  sh -c "apk add --no-cache gcc musl-dev >/dev/null && DATABASE_URL='postgres://indico:indico@indico_postgres:5432/indico?sslmode=disable' go test -v -count=1 ./..."
```

### Stress test in isolation

`TestStress_NoOverselling` fires 300 concurrent `Reserve` calls against 100 stock and asserts **exactly 100 succeed and 200 fail with `INSUFFICIENT_STOCK`** — the proof that `SELECT … FOR UPDATE` keeps the system correct under contention.

Add `-run TestStress_NoOverselling` to any of the commands above:

```bash
# Native
go test -v -count=1 -run TestStress_NoOverselling ./internal/repository/...

# Docker
docker run --rm --network indico_default \
  -v "$PWD/backend:/src" -w /src golang:1.27-alpine \
  sh -c "apk add --no-cache gcc musl-dev >/dev/null && \
         DATABASE_URL='postgres://indico:indico@indico_postgres:5432/indico?sslmode=disable' \
         go test -v -count=1 -run TestStress_NoOverselling ./internal/repository/..."
```

## Environment variables

| Name | Default | Purpose |
|---|---|---|
| `DATABASE_URL` | `postgres://indico:indico@127.0.0.1:5432/indico?sslmode=disable` | Postgres connection string |
| `PORT` | `8080` | HTTP listen port |
| `RESERVATION_TTL` | `5m` | Time after which a reservation auto-expires |
| `SWEEPER_INTERVAL` | `15s` | How often the background worker sweeps expired reservations |
| `SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown window for in-flight requests |
| `VITE_API_BASE_URL` | `http://localhost:8080` | Frontend → backend base URL (build-time) |

## Trying the API with curl

```bash
# Health
curl http://localhost:8080/healthz

# Stock for an item
curl "http://localhost:8080/api/v1/inventory/stock?item_id=item_4021"

# Reserve 2 units
curl -X POST http://localhost:8080/api/v1/inventory/reserve \
  -H "Content-Type: application/json" \
  -d '{"user_id":"usr_9981","item_id":"item_4021","quantity":2}'

# Confirm (paste the reservation_id from the response above)
curl -X POST http://localhost:8080/api/v1/inventory/confirm \
  -H "Content-Type: application/json" \
  -d '{"reservation_id":"res_XXXXXX"}'
```

Negative cases worth trying:

```bash
# Item that does not exist → 404 NOT_FOUND
curl -i "http://localhost:8080/api/v1/inventory/stock?item_id=does_not_exist"

# Reservation that does not exist → 404 NOT_FOUND
curl -i -X POST http://localhost:8080/api/v1/inventory/confirm \
  -H "Content-Type: application/json" \
  -d '{"reservation_id":"res_nope"}'

# Quantity larger than stock → 409 INSUFFICIENT_STOCK with available/requested in details
curl -i -X POST http://localhost:8080/api/v1/inventory/reserve \
  -H "Content-Type: application/json" \
  -d '{"user_id":"usr_9981","item_id":"item_4021","quantity":9999}'
```

## Layout

```
.
├── backend/                # Go + Gin
│   ├── cmd/server/         # main.go (entrypoint, graceful shutdown, sweeper)
│   ├── internal/
│   │   ├── api/            # gin handlers + router
│   │   ├── service/        # business logic (depends on Repo interface)
│   │   ├── repository/     # pgx queries, FOR UPDATE concurrency
│   │   ├── domain/         # models, DTOs, error codes
│   │   └── config/         # env loading
│   ├── scripts/            # pg_hba.conf + init script for dev Postgres
│   └── migrations/         # SQL migration scripts
├── frontend/               # React + Vite
│   ├── src/
│   │   ├── api/            # axios client + types
│   │   ├── hooks/          # TanStack Query hooks
│   │   ├── components/     # StockCard, ReservationForm, Countdown, Confirm
│   │   └── App.tsx
│   ├── Dockerfile
│   └── nginx.conf
├── docker-compose.yaml     # one command: full stack
├── ARCHITECTURE.md         # design defense
└── README.md
```

See **[ARCHITECTURE.md](./ARCHITECTURE.md)** for the design rationale, scaling discussion, and AI-transparency notes.
