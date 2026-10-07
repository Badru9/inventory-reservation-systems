# Architecture Notes

## 1. Architectural design and database locking strategy

### Schema

Two tables. `items` is the source of truth for *physical* stock; `reservations` is the source of truth for *what is currently held* against that physical stock. The `reserved_stock` value returned by `GET /stock` is **derived** at query time — never stored — so the two columns can never disagree.

```sql
items(item_id PK, name, total_stock, created_at)
reservations(reservation_id PK, user_id, item_id FK, quantity, status,
             expires_at, created_at, confirmed_at)
  WHERE status = 'active' AND expires_at > NOW()  -- partial index
```

A partial index on `expires_at WHERE status='active'` keeps the expiry sweep O(active rows), not O(table). Status is an enum-like `TEXT` (`active|confirmed|expired`) — never `DELETE` so we keep an audit trail.

### Concurrency: `SELECT … FOR UPDATE`

Every reservation is wrapped in a transaction that opens with:

```sql
BEGIN;
SELECT total_stock FROM items WHERE item_id = $1 FOR UPDATE;
SELECT COALESCE(SUM(quantity),0)
  FROM reservations
  WHERE item_id = $1 AND status = 'active' AND expires_at > NOW();
-- if available < quantity → ROLLBACK + return 409 INSUFFICIENT_STOCK
INSERT INTO reservations(...) VALUES (..., 'active', NOW() + INTERVAL '5 min');
COMMIT;
```

The row lock on `items` serialises any other transaction trying to reserve the same item. Crucially, **different items never block each other** — high aggregate throughput. The two queries inside the transaction also run under read-committed isolation: the first locks the row, the second reads the latest committed reservation count (which is the count after any other transaction that has already updated the row), so the available-stock calculation is always correct.

`Confirm` is similarly transactional: `SELECT … FOR UPDATE` on the reservation row, then conditional `UPDATE … SET status='confirmed' WHERE reservation_id = $1`. If a second client tries to confirm the same id, it sees `status='confirmed'` and returns the existing confirmation idempotently rather than double-charging.

`ReadCommitted` is sufficient because all invariant checks happen inside a single transaction that holds the row lock for the relevant item. There is no read-then-decide pattern outside the transaction.

### Routing

`internal/api/router.go` mounts three handlers under `/api/v1/inventory` plus `/healthz`. A small `requestLogger` middleware writes a one-liner per request to stdout — enough for ops without burying the signal in structured noise. There is no global auth layer; the spec did not require one, and the take-home window prefers a clean handler over a thin auth stub.

### Error envelope

Every non-2xx response uses the same shape, so the React client always parses the same way:

```json
{ "status": "error", "error": { "code": "INSUFFICIENT_STOCK", "message": "...", "details": { "available": 3, "requested": 5 } } }
```

Codes: `INVALID_INPUT` (400), `NOT_FOUND` (404), `INSUFFICIENT_STOCK` (409), `RESERVATION_EXPIRED` (410), `RESERVATION_ALREADY_CONFIRMED` (410), `INTERNAL` (500).

## 2. Distributed scaling and failure modes

Running 10 instances of this service against a single PostgreSQL works — the `FOR UPDATE` lock still prevents overselling — but two failure modes emerge quickly under flash-sale load.

**Hot-item contention.** When 50 000 goroutines on 10 instances all race to reserve the *same* `item_id`, every transaction queues on the same `items.item_id` row lock. Aggregate throughput collapses to the rate at which one row can be locked-and-released. Postgres serialises on that single row; CPU on the DB is barely 30 %.

**Statelessness is fine, but stateful coordination is not.** Today the expiry sweeper lives in every replica, so 10 instances all hammer Postgres with the same `UPDATE … WHERE status='active' AND expires_at <= NOW()`. The query is idempotent and `WHERE` filters on the partial index, so correctness is preserved, but the DB does 10× the work it needs to.

**Redesign for horizontal scale:**

1. **Sharding by item.** Place each `item_id` on one of N Postgres shards (consistent hash or modulo). A lightweight router in front of the API picks the shard. The hot-item lock is then a problem of *one* shard, not the whole fleet. 10 instances × 10 shards = 100 independent lock surfaces.
2. **Queue-based reservation worker.** Replace the synchronous "POST → tx" path with "POST → enqueue → worker commits". Workers consume the queue in small batches (e.g. 50 at a time) inside a single per-item transaction, so one DB round-trip handles 50 reservations and the lock is held for the same wall-clock time as one transaction today. Throughput rises ~50× per shard.
3. **Single-leader sweeper.** Run the expiry sweeper on **one** instance only — picked via Postgres `pg_try_advisory_lock` so failover is automatic. Other instances act as warm standbys.
4. **Stateless API replicas.** Keep them stateless, but front them with a sticky-by-`item_id` load balancer (Envoy route hint or `hash_key(item_id)`) so a user retrying the same item hits the same replica. This is purely a latency optimisation, not a correctness one.
5. **Read replicas for `/stock`.** Polling traffic from the dashboard can be served from a hot standby with a small `~50ms` lag, which is invisible to a human and frees the primary.

These changes preserve the invariant that no row lock is ever held longer than the time it takes to insert one reservation row, so the system stays correct as it scales.

## 3. Engineering trade-offs and AI transparency

### Trade-offs inside the 4-to-8-hour window

- **No JWT / user auth.** `user_id` is taken from the request body. The take-home prompt is about concurrency, not identity. A real product would issue opaque session tokens and resolve `user_id` server-side.
- **Embedded migrations instead of `golang-migrate`.** I shipped a 30-line `RunMigrations` that checks a `schema_migrations` table and applies the embedded SQL. It is intentionally less powerful than `golang-migrate` (no down-step planning) but it removes a binary dependency from the container image and keeps the dev loop to `go run`.
- **String IDs (`res_xxxxxx`) instead of UUIDs.** The prompt showed `res_883291`. Eight bytes of `crypto/rand` is more than enough entropy for this volume and is dramatically easier to read in logs and on the dashboard than a UUID.
- **No `pgx` prepared-statement cache configuration.** Default caching is enough at this load and keeps the code small.
- **Frontend: no skeleton loaders, no design system.** The dashboard is a single dark card layout with a 2-column grid. Time went into the reservation lifecycle (countdown, expiry, confirm, error rendering) instead of polish.
- **No frontend tests.** React Testing Library would be the right next step, but the test budget went to the backend where the correctness story lives.

### AI-flaw anecdote

Early in the build, an AI assistant suggested an "optimistic" `UPDATE … RETURNING` for `/reserve` that looks like this:

```sql
UPDATE items
SET total_stock = total_stock - $1
WHERE item_id = $2 AND total_stock >= $1
RETURNING total_stock;
```

The claim was that this is "lock-free and faster than `FOR UPDATE`." That is true *only* if `reserved_stock` lives in `items` as a single counter. In our schema `reserved_stock` is **derived** from the `reservations` table, so the available count requires a `SELECT SUM(quantity)` which races against concurrent inserts. Without a transaction wrapper, two concurrent reserves can both observe `available=2`, both insert a `quantity=2` reservation, and oversell.

The fix was to wrap the read and the insert in a single transaction with `SELECT … FOR UPDATE` on the parent `items` row. The `FOR UPDATE` row lock is what makes the read-then-insert atomic; the `UPDATE` returning approach only works when the counter is colocal. I caught the issue in the `TestStress_NoOverselling` test (300 concurrent goroutines, 100 stock) — which fails immediately under the optimistic version — and reverted to the pessimistic design before merging the code. The stress test now passes deterministically.
