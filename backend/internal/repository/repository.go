package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/indico/flashsale/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Pool() *pgxpool.Pool { return r.pool }

func (r *Repository) Close() { r.pool.Close() }

// RunMigrations applies every .up.sql file in migrationsDir in lexical order.
// We use a tiny built-in migrator (not golang-migrate) so the binary is self-contained.
func (r *Repository) RunMigrations(ctx context.Context, upSQL string) error {
	if _, err := r.pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations(version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	var version string
	row := r.pool.QueryRow(ctx, `SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`)
	if err := row.Scan(&version); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("read schema_migrations: %w", err)
	}
	if version == "001_init" {
		return nil // already applied
	}
	if _, err := r.pool.Exec(ctx, upSQL); err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}
	if _, err := r.pool.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES ('001_init') ON CONFLICT DO NOTHING`); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}
	return nil
}

// GetStock returns the current stock breakdown for an item.
// It uses a single transaction with FOR UPDATE on the item row so the
// returned breakdown is consistent with any concurrent reservation in flight.
func (r *Repository) GetStock(ctx context.Context, itemID string) (*domain.StockView, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var total int
	err = tx.QueryRow(ctx, `SELECT total_stock FROM items WHERE item_id = $1 FOR SHARE`, itemID).Scan(&total)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}

	var reserved int
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(quantity), 0)
		FROM reservations
		WHERE item_id = $1 AND status = 'active' AND expires_at > NOW()
	`, itemID).Scan(&reserved)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &domain.StockView{
		ItemID:         itemID,
		TotalStock:     total,
		ReservedStock:  reserved,
		AvailableStock: total - reserved,
	}, nil
}

// Reserve atomically reserves quantity for an item.
// Uses SELECT ... FOR UPDATE to serialise concurrent reservations on the same item.
func (r *Repository) Reserve(ctx context.Context, userID, itemID string, quantity int, ttl time.Duration) (*domain.Reservation, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var total int
	err = tx.QueryRow(ctx, `SELECT total_stock FROM items WHERE item_id = $1 FOR UPDATE`, itemID).Scan(&total)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}

	var reserved int
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(quantity), 0)
		FROM reservations
		WHERE item_id = $1 AND status = 'active' AND expires_at > NOW()
	`, itemID).Scan(&reserved)
	if err != nil {
		return nil, err
	}

	available := total - reserved
	if available < quantity {
		return nil, domain.NewInsufficientError(available, quantity)
	}

	id, err := newID("res_")
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().UTC().Add(ttl)

	_, err = tx.Exec(ctx, `
		INSERT INTO reservations(reservation_id, user_id, item_id, quantity, status, expires_at)
		VALUES ($1, $2, $3, $4, 'active', $5)
	`, id, userID, itemID, quantity, expiresAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &domain.Reservation{
		ReservationID: id,
		UserID:        userID,
		ItemID:        itemID,
		Quantity:      quantity,
		Status:        domain.StatusActive,
		ExpiresAt:     expiresAt,
		CreatedAt:     time.Now().UTC(),
	}, nil
}

// Confirm permanently commits an active reservation.
// Returns ErrExpired if past expires_at, ErrAlreadyConfirmed if not active, ErrResNotFound if missing.
func (r *Repository) Confirm(ctx context.Context, reservationID string) (*domain.Reservation, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var res domain.Reservation
	var confirmedAt *time.Time
	row := tx.QueryRow(ctx, `
		SELECT reservation_id, user_id, item_id, quantity, status, expires_at, created_at, confirmed_at
		FROM reservations
		WHERE reservation_id = $1
		FOR UPDATE
	`, reservationID)
	err = row.Scan(&res.ReservationID, &res.UserID, &res.ItemID, &res.Quantity, &res.Status, &res.ExpiresAt, &res.CreatedAt, &confirmedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrResNotFound
	}
	if err != nil {
		return nil, err
	}
	res.ConfirmedAt = confirmedAt

	switch res.Status {
	case domain.StatusConfirmed:
		// Idempotent: return the existing confirmation.
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return &res, nil
	case domain.StatusExpired:
		return nil, domain.ErrExpired
	case domain.StatusActive:
		// fall through
	default:
		return nil, domain.ErrAlreadyConfirmed
	}

	if time.Now().UTC().After(res.ExpiresAt) {
		// Mark as expired and surface the right error.
		_, _ = tx.Exec(ctx, `UPDATE reservations SET status = 'expired' WHERE reservation_id = $1`, reservationID)
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return nil, domain.ErrExpired
	}

	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `UPDATE reservations SET status = 'confirmed', confirmed_at = $1 WHERE reservation_id = $2`, now, reservationID)
	if err != nil {
		return nil, err
	}
	res.Status = domain.StatusConfirmed
	res.ConfirmedAt = &now

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &res, nil
}

// SweepExpired marks every active reservation past its expires_at as expired.
// Returns the number of rows updated. Called by the background ticker.
func (r *Repository) SweepExpired(ctx context.Context) (int64, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE reservations
		SET status = 'expired'
		WHERE status = 'active' AND expires_at <= NOW()
	`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func newID(prefix string) (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b), nil
}
