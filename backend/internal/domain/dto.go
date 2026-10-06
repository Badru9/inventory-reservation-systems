package domain

import "time"

type ReserveRequest struct {
	UserID   string `json:"user_id"   binding:"required,min=1,max=64"`
	ItemID   string `json:"item_id"   binding:"required,min=1,max=64"`
	Quantity int    `json:"quantity"  binding:"required,min=1,max=10000"`
}

type ReserveResponse struct {
	Status        string    `json:"status"`
	ReservationID string    `json:"reservation_id"`
	ItemID        string    `json:"item_id"`
	Quantity      int       `json:"quantity"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type ConfirmRequest struct {
	ReservationID string `json:"reservation_id" binding:"required,min=1,max=64"`
}

type ConfirmResponse struct {
	Status        string    `json:"status"`
	ReservationID string    `json:"reservation_id"`
	ConfirmedAt   time.Time `json:"confirmed_at"`
}

type ErrorEnvelope struct {
	Status string   `json:"status"`
	Error  APIError `json:"error"`
}

func Success(data any) map[string]any {
	return map[string]any{"status": "success", "data": data}
}
