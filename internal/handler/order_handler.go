package handler

import (
	"errors"
	"net/http"
	"strconv"

	"shopflow/internal/middleware"
	"shopflow/internal/models"
	"shopflow/internal/service"
)

// OrderHandler coordinates Order HTTP requests.
type OrderHandler struct {
	service service.OrderService
}

// NewOrderHandler creates a new OrderHandler instance.
func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// PlaceOrder handles POST /api/v1/orders.
func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		SendError(w, http.StatusUnauthorized, "unauthorized", "UNAUTHORIZED")
		return
	}

	o, err := h.service.PlaceOrder(r.Context(), userID)
	if err != nil {
		if errors.Is(err, models.ErrInvalidInput) {
			SendError(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
			return
		}
		if errors.Is(err, models.ErrInsufficientStock) {
			SendError(w, http.StatusConflict, err.Error(), "INSUFFICIENT_STOCK")
			return
		}
		if errors.Is(err, models.ErrCartNotFound) {
			SendError(w, http.StatusBadRequest, "cart not found", "CART_NOT_FOUND")
			return
		}
		SendError(w, http.StatusInternalServerError, "failed to place order", "INTERNAL_SERVER_ERROR")
		return
	}

	SendJSON(w, http.StatusCreated, o)
}

// ListOrders handles GET /api/v1/orders.
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		SendError(w, http.StatusUnauthorized, "unauthorized", "UNAUTHORIZED")
		return
	}

	orders, err := h.service.ListOrders(r.Context(), userID)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "failed to list orders", "INTERNAL_SERVER_ERROR")
		return
	}

	if orders == nil {
		orders = []*models.Order{}
	}

	SendJSON(w, http.StatusOK, orders)
}

// GetOrder handles GET /api/v1/orders/{id}.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		SendError(w, http.StatusUnauthorized, "unauthorized", "UNAUTHORIZED")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		SendError(w, http.StatusBadRequest, "invalid order ID", "BAD_REQUEST")
		return
	}

	o, err := h.service.GetOrderDetails(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, models.ErrOrderNotFound) {
			SendError(w, http.StatusNotFound, err.Error(), "ORDER_NOT_FOUND")
			return
		}
		SendError(w, http.StatusInternalServerError, "failed to get order details", "INTERNAL_SERVER_ERROR")
		return
	}

	SendJSON(w, http.StatusOK, o)
}
