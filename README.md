# INDICO Flash-Sale Take-Home

High-concurrency inventory reservation system built with **Go + Gin + PostgreSQL** (backend, 80%) and **React + Vite** (frontend, 20%).

## Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.23, Gin, pgx/v5, embedded SQL migrations |
| Database | PostgreSQL 16 |
| Frontend | React 19, Vite 8, TypeScript, TanStack Query, axios |
| Orchestration | Docker Compose (one command: `docker compose up`) |

## Endpoints

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/api/v1/inventory/reserve` | Atomically reserve stock (5-minute TTL) |
| `POST` | `/api/v1/inventory/confirm` | Commit an active reservation |
| `GET`  | `/api/v1/inventory/stock?item_id=…` | Real-time stock breakdown |
| `GET`  | `/healthz` | Liveness |

## Quick start (one command)

```bash
docker compose up --build
```

After the stack is up:
- Frontend: <http://localhost:5173>
- Backend:  <http://localhost:8080/healthz>

Click **Reserve**, watch the 5-minute countdown, then **Confirm Purchase**.

## Run components individually

### 1. Database

```bash
docker compose up postgres -d
```

### 2. Backend

```bash
cd backend
DATABASE_URL="postgres://indico:indico@localhost:5432/indico?sslmode=disable" go run ./cmd/server
```

### 3. Frontend (Vite dev server)

```bash
cd frontend
npm install
npm run dev
```

Open <http://localhost:5173>.

## Tests

Unit tests (no DB needed) + race detector:

```bash
cd backend
PATH="/c/Users/RIVAN/scoop/apps/mingw/current/bin:$PATH" go test -race -v ./...
```

Integration + stress test (requires Postgres at `$DATABASE_URL`):

```bash
cd backend
docker compose up postgres -d
DATABASE_URL="postgres://indico:indico@localhost:5432/indico?sslmode=disable" \
  PATH="/c/Users/RIVAN/scoop/apps/mingw/current/bin:$PATH" \
  go test -race -v ./internal/repository/...
```

The stress test (`TestStress_NoOverselling`) reserves 100 stock from 300 concurrent goroutines and asserts exactly 100 succeed.

## Environment variables

| Name | Default | Purpose |
|---|---|---|
| `DATABASE_URL` | `postgres://indico:indico@localhost:5432/indico?sslmode=disable` | Postgres connection string |
| `PORT` | `8080` | HTTP listen port |
| `RESERVATION_TTL` | `5m` | Time after which a reservation auto-expires |
| `SWEEPER_INTERVAL` | `15s` | How often the background worker sweeps expired reservations |
| `SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown window for in-flight requests |
| `VITE_API_BASE_URL` | `http://localhost:8080` | Frontend → backend base URL (build-time) |

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
