package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"shopflow/internal/middleware"
	"shopflow/internal/models"
	"shopflow/internal/service"
)

// CartHandler coordinates Cart HTTP requests.
type CartHandler struct {
	service service.CartService
}

// NewCartHandler creates a new CartHandler instance.
func NewCartHandler(service service.CartService) *CartHandler {
	return &CartHandler{service: service}
}

type addCartItemRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

// AddOrUpdateItem handles POST /api/v1/cart/items.
func (h *CartHandler) AddOrUpdateItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		SendError(w, http.StatusUnauthorized, "unauthorized", "UNAUTHORIZED")
		return
	}

	var req addCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "invalid request body format", "BAD_REQUEST")
		return
	}

	err := h.service.AddOrUpdateItem(r.Context(), userID, req.ProductID, req.Quantity)
	if err != nil {
		if errors.Is(err, models.ErrInvalidInput) {
			SendError(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
			return
		}
		if errors.Is(err, models.ErrProductNotFound) {
			SendError(w, http.StatusNotFound, err.Error(), "PRODUCT_NOT_FOUND")
			return
		}
		if errors.Is(err, models.ErrInsufficientStock) {
			SendError(w, http.StatusConflict, err.Error(), "INSUFFICIENT_STOCK")
			return
		}
		SendError(w, http.StatusInternalServerError, "failed to update cart item", "INTERNAL_SERVER_ERROR")
		return
	}

	SendJSON(w, http.StatusOK, map[string]string{"message": "cart item updated successfully"})
}

// ViewCart handles GET /api/v1/cart.
func (h *CartHandler) ViewCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		SendError(w, http.StatusUnauthorized, "unauthorized", "UNAUTHORIZED")
		return
	}

	cart, err := h.service.GetOrCreateCart(r.Context(), userID)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "failed to get cart", "INTERNAL_SERVER_ERROR")
		return
	}

	SendJSON(w, http.StatusOK, cart)
}

// RemoveItem handles DELETE /api/v1/cart/items/{productId}.
func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		SendError(w, http.StatusUnauthorized, "unauthorized", "UNAUTHORIZED")
		return
	}

	productIDStr := r.PathValue("productId")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil || productID <= 0 {
		SendError(w, http.StatusBadRequest, "invalid product ID", "BAD_REQUEST")
		return
	}

	err = h.service.RemoveItem(r.Context(), userID, productID)
	if err != nil {
		if errors.Is(err, models.ErrCartNotFound) {
			SendError(w, http.StatusNotFound, err.Error(), "CART_NOT_FOUND")
			return
		}
		if errors.Is(err, models.ErrProductNotFound) {
			SendError(w, http.StatusNotFound, "product not found in cart", "PRODUCT_NOT_FOUND")
			return
		}
		SendError(w, http.StatusInternalServerError, "failed to remove item from cart", "INTERNAL_SERVER_ERROR")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
