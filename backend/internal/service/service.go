package service

import (
	"context"
	"time"

	"github.com/indico/flashsale/internal/domain"
)

// Repo captures only the methods the service depends on. The concrete
// *repository.Repository satisfies this interface. Tests pass a fake.
type Repo interface {
	GetStock(ctx context.Context, itemID string) (*domain.StockView, error)
	Reserve(ctx context.Context, userID, itemID string, quantity int, ttl time.Duration) (*domain.Reservation, error)
	Confirm(ctx context.Context, reservationID string) (*domain.Reservation, error)
}

type Service struct {
	repo Repo
	ttl  time.Duration
}

func New(repo Repo, ttl time.Duration) *Service { return &Service{repo: repo, ttl: ttl} }

func (s *Service) GetStock(ctx context.Context, itemID string) (*domain.StockView, error) {
	if itemID == "" {
		return nil, domain.NewValidationError("item_id is required", nil)
	}
	return s.repo.GetStock(ctx, itemID)
}

func (s *Service) Reserve(ctx context.Context, req domain.ReserveRequest) (*domain.Reservation, error) {
	if req.Quantity <= 0 {
		return nil, domain.NewValidationError("quantity must be > 0", nil)
	}
	return s.repo.Reserve(ctx, req.UserID, req.ItemID, req.Quantity, s.ttl)
}

func (s *Service) Confirm(ctx context.Context, reservationID string) (*domain.Reservation, error) {
	if reservationID == "" {
		return nil, domain.NewValidationError("reservation_id is required", nil)
	}
	return s.repo.Confirm(ctx, reservationID)
}
