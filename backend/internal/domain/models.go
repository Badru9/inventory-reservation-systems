package domain

import (
	"errors"
	"time"
)

type Item struct {
	ItemID     string `json:"item_id"`
	Name       string `json:"name"`
	TotalStock int    `json:"total_stock"`
}

type ReservationStatus string

const (
	StatusActive    ReservationStatus = "active"
	StatusConfirmed ReservationStatus = "confirmed"
	StatusExpired   ReservationStatus = "expired"
)

type Reservation struct {
	ReservationID string            `json:"reservation_id"`
	UserID        string            `json:"user_id"`
	ItemID        string            `json:"item_id"`
	Quantity      int               `json:"quantity"`
	Status        ReservationStatus `json:"status"`
	ExpiresAt     time.Time         `json:"expires_at"`
	CreatedAt     time.Time         `json:"created_at"`
	ConfirmedAt   *time.Time        `json:"confirmed_at,omitempty"`
}

type StockView struct {
	ItemID         string `json:"item_id"`
	TotalStock     int    `json:"total_stock"`
	ReservedStock  int    `json:"reserved_stock"`
	AvailableStock int    `json:"available_stock"`
}

type ErrorCode string

const (
	ErrInvalidInput             ErrorCode = "INVALID_INPUT"
	ErrNotFound                 ErrorCode = "NOT_FOUND"
	ErrInsufficientStock        ErrorCode = "INSUFFICIENT_STOCK"
	ErrReservationExpired       ErrorCode = "RESERVATION_EXPIRED"
	ErrReservationAlreadyClosed ErrorCode = "RESERVATION_ALREADY_CONFIRMED"
	ErrInternal                 ErrorCode = "INTERNAL"
)

type APIError struct {
	Code    ErrorCode      `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func (e *APIError) Error() string { return string(e.Code) + ": " + e.Message }

var (
	ErrItemNotFound     = &APIError{Code: ErrNotFound, Message: "item not found"}
	ErrResNotFound      = &APIError{Code: ErrNotFound, Message: "reservation not found"}
	ErrInvalidPayload   = &APIError{Code: ErrInvalidInput, Message: "invalid request parameters"}
	ErrInsufficient     = &APIError{Code: ErrInsufficientStock, Message: "insufficient stock"}
	ErrAlreadyConfirmed = &APIError{Code: ErrReservationAlreadyClosed, Message: "reservation is no longer active"}
	ErrExpired          = &APIError{Code: ErrReservationExpired, Message: "reservation has expired"}
)

func NewInsufficientError(available, requested int) *APIError {
	return &APIError{
		Code:    ErrInsufficientStock,
		Message: "insufficient stock for requested quantity",
		Details: map[string]any{"available": available, "requested": requested},
	}
}

func NewValidationError(msg string, details map[string]any) *APIError {
	return &APIError{Code: ErrInvalidInput, Message: msg, Details: details}
}

func AsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}
