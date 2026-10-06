package repository_test

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/indico/flashsale/internal/domain"
	"github.com/indico/flashsale/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// integrationDB returns a real Postgres pool, or skips the test if DATABASE_URL is unset
// or unreachable. We reuse the same migrations script the production binary uses.
func integrationDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Skipf("postgres not reachable: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("ping failed: %v", err)
	}
	// Wipe state for a clean run.
	if _, err := pool.Exec(ctx, `TRUNCATE reservations, items RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO items(item_id, name, total_stock) VALUES ('item_4021', 'Test', 100)`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return pool, func() { pool.Close() }
}

const initSchema = `
CREATE TABLE IF NOT EXISTS items (
  item_id      TEXT PRIMARY KEY,
  name         TEXT NOT NULL,
  total_stock  INTEGER NOT NULL CHECK (total_stock >= 0),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS reservations (
  reservation_id  TEXT PRIMARY KEY,
  user_id         TEXT NOT NULL,
  item_id         TEXT NOT NULL REFERENCES items(item_id),
  quantity        INTEGER NOT NULL CHECK (quantity > 0),
  status          TEXT NOT NULL CHECK (status IN ('active','confirmed','expired')),
  expires_at      TIMESTAMPTZ NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  confirmed_at    TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_reservations_item_status ON reservations(item_id, status);
CREATE INDEX IF NOT EXISTS idx_reservations_expires     ON reservations(expires_at) WHERE status = 'active';
`

func TestRepository_RoundTrip(t *testing.T) {
	pool, cleanup := integrationDB(t)
	defer cleanup()
	ctx := context.Background()

	repo := repository.New(pool)
	if err := repo.RunMigrations(ctx, initSchema); err != nil {
		t.Fatalf("migrations: %v", err)
	}

	// 1) Reserve.
	res, err := repo.Reserve(ctx, "u1", "item_4021", 3, 5*time.Minute)
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if res.Status != domain.StatusActive {
		t.Fatalf("expected active, got %s", res.Status)
	}

	// 2) Stock reflects reservation.
	stock, err := repo.GetStock(ctx, "item_4021")
	if err != nil {
		t.Fatalf("stock: %v", err)
	}
	if stock.ReservedStock != 3 || stock.AvailableStock != 97 {
		t.Fatalf("unexpected stock: %+v", stock)
	}

	// 3) Confirm.
	confirmed, err := repo.Confirm(ctx, res.ReservationID)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if confirmed.Status != domain.StatusConfirmed {
		t.Fatalf("expected confirmed, got %s", confirmed.Status)
	}

	// 4) Available is now total (confirmed is no longer active).
	stock2, _ := repo.GetStock(ctx, "item_4021")
	if stock2.AvailableStock != 100 {
		t.Fatalf("after confirm available should be 100, got %d", stock2.AvailableStock)
	}
}

// TestStress_NoOverselling hammers the reserve endpoint with N concurrent goroutines
// against a stock of 100. Exactly 100 must succeed; the remaining 50 must hit
// INSUFFICIENT_STOCK. No race detector complaints are the proof of correctness.
func TestStress_NoOverselling(t *testing.T) {
	pool, cleanup := integrationDB(t)
	defer cleanup()
	ctx := context.Background()

	repo := repository.New(pool)
	if err := repo.RunMigrations(ctx, initSchema); err != nil {
		t.Fatalf("migrations: %v", err)
	}

	const (
		totalStock = 100
		workers    = 300
	)
	var (
		success atomic.Int64
		insuf   atomic.Int64
		other   atomic.Int64
	)

	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := repo.Reserve(ctx, "u_stress", "item_4021", 1, 5*time.Minute)
			switch {
			case err == nil:
				success.Add(1)
			case isInsufficient(err):
				insuf.Add(1)
			default:
				other.Add(1)
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()

	if s := success.Load(); s != totalStock {
		t.Fatalf("expected %d successful reserves, got %d", totalStock, s)
	}
	if i := insuf.Load(); i != workers-totalStock {
		t.Fatalf("expected %d insufficient, got %d", workers-totalStock, i)
	}
	if o := other.Load(); o != 0 {
		t.Fatalf("got %d unexpected errors", o)
	}
}

func isInsufficient(err error) bool {
	apiErr, ok := domain.AsAPIError(err)
	return ok && apiErr.Code == domain.ErrInsufficientStock
}
