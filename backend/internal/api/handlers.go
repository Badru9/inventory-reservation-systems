package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/indico/flashsale/internal/domain"
	"github.com/indico/flashsale/internal/service"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Stock(c *gin.Context) {
	itemID := c.Query("item_id")
	if itemID == "" {
		writeError(c, http.StatusBadRequest, domain.NewValidationError("item_id query parameter is required", nil))
		return
	}
	stock, err := h.svc.GetStock(c.Request.Context(), itemID)
	if err != nil {
		writeError(c, statusFor(err), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": stock})
}

func (h *Handler) Reserve(c *gin.Context) {
	var req domain.ReserveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, domain.NewValidationError("invalid request body", map[string]any{"reason": err.Error()}))
		return
	}
	res, err := h.svc.Reserve(c.Request.Context(), req)
	if err != nil {
		writeError(c, statusFor(err), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"reservation_id": res.ReservationID,
		"item_id":        res.ItemID,
		"quantity":       res.Quantity,
		"expires_at":     res.ExpiresAt,
	})
}

func (h *Handler) Confirm(c *gin.Context) {
	var req domain.ConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, domain.NewValidationError("invalid request body", map[string]any{"reason": err.Error()}))
		return
	}
	res, err := h.svc.Confirm(c.Request.Context(), req.ReservationID)
	if err != nil {
		writeError(c, statusFor(err), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"reservation_id": res.ReservationID,
		"confirmed_at":   res.ConfirmedAt,
	})
}

func writeError(c *gin.Context, status int, err error) {
	apiErr, ok := domain.AsAPIError(err)
	if !ok {
		slog.Error("internal error", "err", err, "path", c.Request.URL.Path)
		apiErr = &domain.APIError{Code: domain.ErrInternal, Message: "internal server error"}
		status = http.StatusInternalServerError
	}
	c.JSON(status, domain.ErrorEnvelope{Status: "error", Error: *apiErr})
}

func statusFor(err error) int {
	apiErr, ok := domain.AsAPIError(err)
	if !ok {
		return http.StatusInternalServerError
	}
	switch apiErr.Code {
	case domain.ErrInvalidInput:
		return http.StatusBadRequest
	case domain.ErrNotFound:
		return http.StatusNotFound
	case domain.ErrInsufficientStock:
		return http.StatusConflict
	case domain.ErrReservationExpired, domain.ErrReservationAlreadyClosed:
		return http.StatusGone
	default:
		return http.StatusInternalServerError
	}
}

// Sentinel to make errors.Is usable in tests if needed.
var _ = errors.New
