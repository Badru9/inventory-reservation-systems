# Fullstack take home assignment

High-throughput inventory reservation system and mini dashboard.

| Item | Detail |
|---|---|
| Target level | Mid to senior fullstack |
| Strict tech stack | Go + Gin + PostgreSQL, React + Vite |
| Time commitment | 4 to 8 hours max |

## 1. Assignment objective

During flash-sale events, thousands of concurrent users try to buy high-demand products at the same time. This assessment evaluates how well you can design a resilient, high-concurrency backend service using Go, the Gin framework and a real PostgreSQL database (80% weighting). You then connect it to a clean, functional frontend built strictly with React and Vite (20% weighting).

## 2. Backend specifications and mandatory stack (80% weighting)

The backend microservice must use:

- Go (Golang) with the Gin HTTP framework (`github.com/gin-gonic/gin`).
- A real PostgreSQL instance running via Docker Compose. In-memory databases, SQLite and mock data stores are not allowed.

The service must expose the REST endpoints below. For each endpoint, implement both success and error scenarios, and design the JSON response structure that best follows API best practices.

### 2.1 Reserve stock

- Endpoint: `POST /api/v1/inventory/reserve`
- Objective: atomically reserve inventory in PostgreSQL for a user if enough available stock exists. Reserved stock must expire automatically after 5 minutes unless confirmed.

Request payload:

```json
{
  "user_id": "usr_9981",
  "item_id": "item_4021",
  "quantity": 2
}
```

Success response:

```json
{
  "status": "success",
  "reservation_id": "res_883291",
  "item_id": "item_4021",
  "quantity": 2,
  "expires_at": "2026-07-20T16:35:00Z"
}
```

Negative cases: handle and implement error scenarios such as insufficient inventory and invalid input parameters. Design the JSON error structure yourself so it works well for client integration.

### 2.2 Confirm reservation

- Endpoint: `POST /api/v1/inventory/confirm`
- Objective: permanently commit an active reservation in PostgreSQL, complete the order and decrement physical inventory.

Request payload:

```json
{
  "reservation_id": "res_883291"
}
```

Success response:

```json
{
  "status": "success",
  "reservation_id": "res_883291",
  "confirmed_at": "2026-07-20T16:32:00Z"
}
```

Negative cases: handle and implement error scenarios such as expired reservations and non-existent reservation IDs. Design the JSON error structure yourself.

### 2.3 Check inventory status

- Endpoint: `GET /api/v1/inventory/stock?item_id=item_4021`
- Objective: query PostgreSQL for the real-time stock breakdown of an item. Handle edge cases where the item is not found or the parameters are invalid, and design appropriate error JSON structures.

Success response:

```json
{
  "item_id": "item_4021",
  "total_stock": 100,
  "reserved_stock": 15,
  "available_stock": 85
}
```

## 3. Frontend specifications (20% weighting)

The mini dashboard must be built strictly with React initialized with Vite (JavaScript or TypeScript). Vue, Svelte, Angular, Next.js, Nuxt and plain HTML/JS solutions are unacceptable and will cause the submission to be rejected.

Implement these core workflows:

1. Live inventory tracker. Show total, reserved and available stock for an item, with a manual refresh button or automated polling.
2. Reservation form. A clean form where the user enters a User ID and a Quantity to request a reservation from the Gin backend.
3. Expiration and confirmation workflow.
   - Show the active reservation details next to a live 5-minute countdown timer.
   - Provide a "Confirm Purchase" button that confirms the order.
   - Handle negative backend responses (failed reservations, expired sessions) and render them in the React UI.

## 4. Mandatory technical requirements and quality expectations

### 4.1 Strict stack enforcement

- Backend framework: Go with Gin.
- Database: a real PostgreSQL instance. Include SQL migration files or schema initialization scripts.
- Frontend framework: React + Vite only.
- Docker setup: provide a working `docker-compose.yaml` that starts the PostgreSQL container together with the backend (and optionally the frontend), so the whole system runs with one command such as `docker compose up`.

### 4.2 Data integrity and concurrency

Use strict PostgreSQL transaction handling and concurrency control. Overselling or negative stock allocations under heavy concurrent load are not acceptable.

### 4.3 Automatic expiry cleanup

Expired reservations must return to available stock automatically when they expire.

### 4.4 Graceful shutdown

The Gin server must intercept system signals and shut down cleanly, without corrupting state or dropping active database connections.

### 4.5 Backend testing

Include unit tests and stress tests. The Go codebase will be validated with the standard race detector:

```bash
go test -race -v ./...
```

## 5. Required written deliverable: `ARCHITECTURE.md`

Answer these three prompts in an `ARCHITECTURE.md` file (1 to 2 pages max):

1. Architectural design and database locking strategy. Defend your PostgreSQL schema design, Gin routing structure, and your choices for synchronization and transaction management under high concurrency.
2. Distributed scaling and failure modes. What breaks when the service is scaled horizontally to 10 instances pointing at PostgreSQL? How would you redesign it for a stateless, distributed deployment?
3. Engineering trade-offs and AI transparency. Describe the trade-offs you made within the 4 to 8 hour window across both the React and Go stacks. If you used AI tools (ChatGPT, Copilot, Claude), describe one case where an AI suggestion was flawed and how you fixed it.

## 6. Submission guidelines and Git requirements

Proof of work (incremental Git commit history):

- Submit a link to a private or public Git repository (GitHub or GitLab). Zip files and single-commit code dumps are not accepted.
- Commit incrementally across both the React frontend and Go backend phases. A submission with one massive commit will be rejected.

Setup instructions: include a `README.md` that explains how to:

1. Run PostgreSQL via `docker compose up`.
2. Start the Gin backend.
3. Run the React + Vite frontend.

Next stage: passing submissions advance to a 30-minute live code walkthrough and defense call. You will present your React and Go implementation and discuss your architectural choices.

## Deliverables checklist

- [ ] Go + Gin backend with the three endpoints (reserve, confirm, stock)
- [ ] Error scenarios and a consistent JSON error structure for every endpoint
- [ ] PostgreSQL schema or migration files
- [ ] Transaction and concurrency control with no overselling
- [ ] Automatic expiry cleanup (5-minute reservations)
- [ ] Graceful shutdown on system signals
- [ ] Unit tests and stress tests that pass `go test -race -v ./...`
- [ ] React + Vite dashboard (inventory tracker, reservation form, countdown and confirm)
- [ ] `docker-compose.yaml` that starts the full system with one command
- [ ] `ARCHITECTURE.md` (1 to 2 pages, three prompts answered)
- [ ] `README.md` with setup instructions
- [ ] Incremental Git history across backend and frontend