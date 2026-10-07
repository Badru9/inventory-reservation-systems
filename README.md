# INDICO Flash-Sale Take-Home

High-concurrency inventory reservation system built with **Go + Gin + PostgreSQL** (backend, 80%) and **React + Vite** (frontend, 20%).

## Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.27, Gin, pgx/v5, embedded SQL migrations |
| Database | PostgreSQL 16 |
| Frontend | React 19, Vite 8, TypeScript, TanStack Query, axios |
| Orchestration | Docker Compose — one command spins up the full stack |
| Dev ergonomics | `Makefile` + `scripts/dev.ps1` for one-liner workflows |

## Endpoints

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/api/v1/inventory/reserve` | Atomically reserve stock (5-minute TTL) |
| `POST` | `/api/v1/inventory/confirm` | Commit an active reservation |
| `GET`  | `/api/v1/inventory/stock?item_id=…` | Real-time stock breakdown |
| `GET`  | `/healthz` | Liveness |

## TL;DR — three commands

```bash
# 1. Spin up Postgres + backend + frontend
make up            # or: .\scripts\dev.ps1 up

# 2. Open the dashboard
#    Frontend → http://localhost:5173
#    Backend  → http://localhost:8080/healthz

# 3. Run the test suite (unit + integration + 300-goroutine stress)
make test          # or: .\scripts\dev.ps1 test
```

Run `make help` (or `.\scripts\dev.ps1 help`) for the full target list.

## Prerequisites

| Tool | Why | Install hint |
|---|---|---|
| **Docker Desktop** (with WSL2 on Windows) | Runs Postgres + (optionally) the backend & frontend | <https://docker.com/products/docker-desktop> |
| **Go 1.27+** | Backend dev loop, tests | `scoop install go` |
| **Node 20+** | Frontend dev loop, Vite | <https://nodejs.org> |
| **MinGW** (Windows only) | `go test -race` needs CGO | `scoop install mingw` |
| **make** (optional) | Use the `Makefile` targets | `choco install make` or use `scripts\dev.ps1` instead |

## Daily workflow (one-liners)

### Using the Makefile (macOS / Linux / Windows-with-`make`)

| What you want | Command |
|---|---|
| Start the full stack | `make up` |
| Start only Postgres (so you can run the backend natively) | `make db` |
| Run the Go backend natively | `make backend` |
| Run the Vite dev server | `make frontend` |
| Run all tests (race detector) | `make test` |
| Run only the no-overselling stress test | `make stress` |
| Tail container logs | `make logs` |
| Open psql in the Postgres container | `make psql` |
| Stop the stack | `make down` |
| Stop and wipe the Postgres volume | `make clean` |

### Using `scripts\dev.ps1` (Windows, no `make` required)

| What you want | Command |
|---|---|
| Start the full stack | `.\scripts\dev.ps1 up` |
| Start only Postgres | `.\scripts\dev.ps1 db` |
| Run the Go backend natively | `.\scripts\dev.ps1 backend` |
| Run the Vite dev server | `.\scripts\dev.ps1 frontend` |
| Run all tests | `.\scripts\dev.ps1 test` |
| Run only the stress test | `.\scripts\dev.ps1 stress` |
| Stop the stack | `.\scripts\dev.ps1 down` |

## Manual setup (no `make` / `dev.ps1`)

If you'd rather run the commands yourself:

### 1. Start the database

```bash
docker compose up -d postgres
```

### 2. Start the backend

```bash
cd backend
DATABASE_URL="postgres://indico:indico@127.0.0.1:5432/indico?sslmode=disable" \
  go run ./cmd/server
```

The server listens on `:8080`, runs migrations on boot, and starts a background sweeper that releases expired reservations every 15 s.

### 3. Start the frontend

```bash
cd frontend
npm install
npm run dev
```

Open <http://localhost:5173>. Vite reads `VITE_API_BASE_URL` from `.env.development` (already set to `http://localhost:8080`).

## Trying the dashboard

1. Open <http://localhost:5173>.
2. The **Live Inventory** card polls every 3 s — flip between `item_4021` and `item_7777` to see different items.
3. Fill the **Reserve Stock** form (e.g. `usr_9981`, `item_4021`, qty `2`) → click **Reserve**.
4. The right panel shows your reservation, a live 5-minute countdown, and a **Confirm Purchase** button.
5. Click **Confirm Purchase** — the reservation is committed and the available stock returns to baseline.

To watch the expiry sweep at work, reserve a quantity and then leave the page open past the 5-minute mark. The countdown turns red, the reservation is auto-cleared, and `/stock` shows the units back in `available_stock`.

## Stress / load testing

The integration test `TestStress_NoOverselling` fires 300 concurrent `Reserve` calls against 100 stock and asserts **exactly 100 succeed and 200 fail with `INSUFFICIENT_STOCK`** — the proof that `SELECT … FOR UPDATE` keeps the system correct under contention.

Run it on its own:

```bash
make stress                              # or: .\scripts\dev.ps1 stress
# or
cd backend && DATABASE_URL=... go test -race -v -count=1 -run TestStress_NoOverselling ./internal/repository/...
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

The `Makefile` and `dev.ps1` already set `DATABASE_URL` and `PORT` for you.

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
├── scripts/
│   └── dev.ps1             # PowerShell wrapper for Make targets
├── docker-compose.yaml     # one command: full stack
├── Makefile                # one-liner dev / test / deploy targets
├── ARCHITECTURE.md         # design defense
└── README.md
```

See **[ARCHITECTURE.md](./ARCHITECTURE.md)** for the design rationale, scaling discussion, and AI-transparency notes.
