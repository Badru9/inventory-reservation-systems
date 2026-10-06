package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/indico/flashsale/internal/domain"
	"github.com/indico/flashsale/internal/service"
)

// fakeRepo is an in-memory implementation of the repository contract
// used by the service layer. It is intentionally minimal — it only models
// the fields the service consumes.
type fakeRepo struct {
	stock      map[string]int            // available stock
	reserved   map[string]int            // active reserved quantity
	resByID    map[string]*domain.Reservation
	confirmErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		stock:    map[string]int{"item_4021": 100, "item_7777": 50},
		reserved: map[string]int{"item_4021": 0, "item_7777": 0},
		resByID:  make(map[string]*domain.Reservation),
	}
}

func (f *fakeRepo) GetStock(ctx context.Context, itemID string) (*domain.StockView, error) {
	total, ok := f.stock[itemID]
	if !ok {
		return nil, domain.ErrItemNotFound
	}
	reserved := f.reserved[itemID]
	return &domain.StockView{ItemID: itemID, TotalStock: total, ReservedStock: reserved, AvailableStock: total - reserved}, nil
}

func (f *fakeRepo) Reserve(ctx context.Context, userID, itemID string, quantity int, ttl time.Duration) (*domain.Reservation, error) {
	total, ok := f.stock[itemID]
	if !ok {
		return nil, domain.ErrItemNotFound
	}
	if total-f.reserved[itemID] < quantity {
		return nil, domain.NewInsufficientError(total-f.reserved[itemID], quantity)
	}
	id := "res_test"
	res := &domain.Reservation{
		ReservationID: id,
		UserID:        userID,
		ItemID:        itemID,
		Quantity:      quantity,
		Status:        domain.StatusActive,
		ExpiresAt:     time.Now().Add(ttl),
		CreatedAt:     time.Now(),
	}
	f.resByID[id] = res
	f.reserved[itemID] += quantity
	return res, nil
}

func (f *fakeRepo) Confirm(ctx context.Context, id string) (*domain.Reservation, error) {
	if f.confirmErr != nil {
		return nil, f.confirmErr
	}
	r, ok := f.resByID[id]
	if !ok {
		return nil, domain.ErrResNotFound
	}
	if r.Status != domain.StatusActive {
		return nil, domain.ErrAlreadyConfirmed
	}
	r.Status = domain.StatusConfirmed
	now := time.Now()
	r.ConfirmedAt = &now
	return r, nil
}

func TestReserve_HappyPath(t *testing.T) {
	svc := service.New(newFakeRepo(), 5*time.Minute)

	res, err := svc.Reserve(context.Background(), domain.ReserveRequest{UserID: "u1", ItemID: "item_4021", Quantity: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Quantity != 2 || res.Status != domain.StatusActive {
		t.Fatalf("bad reservation: %+v", res)
	}
}

func TestReserve_InsufficientStock(t *testing.T) {
	svc := service.New(newFakeRepo(), 5*time.Minute)
	_, err := svc.Reserve(context.Background(), domain.ReserveRequest{UserID: "u1", ItemID: "item_7777", Quantity: 9999})
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := domain.AsAPIError(err)
	if !ok || apiErr.Code != domain.ErrInsufficientStock {
		t.Fatalf("expected INSUFFICIENT_STOCK, got %v", err)
	}
}

func TestReserve_ItemNotFound(t *testing.T) {
	svc := service.New(newFakeRepo(), 5*time.Minute)
	_, err := svc.Reserve(context.Background(), domain.ReserveRequest{UserID: "u1", ItemID: "item_missing", Quantity: 1})
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := domain.AsAPIError(err)
	if !ok || apiErr.Code != domain.ErrNotFound {
		t.Fatalf("expected NOT_FOUND, got %v", err)
	}
}

func TestConfirm_AlreadyConfirmed(t *testing.T) {
	repo := newFakeRepo()
	svc := service.New(repo, 5*time.Minute)
	res, err := svc.Reserve(context.Background(), domain.ReserveRequest{UserID: "u1", ItemID: "item_4021", Quantity: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Confirm(context.Background(), res.ReservationID); err != nil {
		t.Fatal(err)
	}
	_, err = svc.Confirm(context.Background(), res.ReservationID)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := domain.AsAPIError(err)
	if !ok {
		t.Fatalf("expected APIError, got %v", err)
	}
	if apiErr.Code != domain.ErrReservationAlreadyClosed {
		t.Fatalf("expected RESERVATION_ALREADY_CONFIRMED, got %v", apiErr.Code)
	}
}

func TestConfirm_NotFound(t *testing.T) {
	svc := service.New(newFakeRepo(), 5*time.Minute)
	_, err := svc.Confirm(context.Background(), "res_does_not_exist")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := domain.AsAPIError(err)
	if !ok || apiErr.Code != domain.ErrNotFound {
		t.Fatalf("expected NOT_FOUND, got %v", err)
	}
}

func TestGetStock_NotFound(t *testing.T) {
	svc := service.New(newFakeRepo(), 5*time.Minute)
	_, err := svc.GetStock(context.Background(), "ghost")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestService_Errors(t *testing.T) {
	if _, ok := domain.AsAPIError(errors.New("plain")); ok {
		t.Fatal("plain error should not be APIError")
	}
}
