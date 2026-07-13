package handler

import (
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

type updateCartResponse struct {
	Message string `json:"message" example:"cart item updated successfully"`
}

// AddOrUpdateItem handles POST /api/v1/cart/items.
// @Summary Add or update item in cart
// @Description Add a product to the user's shopping cart or update its quantity. Authenticated.
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body addCartItemRequest true "Cart item details"
// @Success 200 {object} SuccessResponse[updateCartResponse] "Item added/updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid input data"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Product not found"
// @Failure 409 {object} ErrorResponse "Insufficient stock"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/cart/items [post]
func (h *CartHandler) AddOrUpdateItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		SendError(w, http.StatusUnauthorized, "unauthorized", "UNAUTHORIZED")
		return
	}

	req, ok := DecodeJSON[addCartItemRequest](w, r)
	if !ok {
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
		SendError(w, http.StatusInternalServerError, "failed to update cart item", "INTERNAL_SERVER_ERROR", err)
		return
	}

	SendJSON(w, http.StatusOK, map[string]string{"message": "cart item updated successfully"})
}

// ViewCart handles GET /api/v1/cart.
// @Summary View shopping cart
// @Description Retrieve the authenticated user's active shopping cart. Authenticated.
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse[models.Cart] "Cart details retrieved successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/cart [get]
func (h *CartHandler) ViewCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		SendError(w, http.StatusUnauthorized, "unauthorized", "UNAUTHORIZED")
		return
	}

	cart, err := h.service.GetOrCreateCart(r.Context(), userID)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "failed to get cart", "INTERNAL_SERVER_ERROR", err)
		return
	}

	SendJSON(w, http.StatusOK, cart)
}

// RemoveItem handles DELETE /api/v1/cart/items/{productId}.
// @Summary Remove item from cart
// @Description Remove a product from the user's shopping cart completely. Authenticated.
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Param productId path int true "Product ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse "Invalid product ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Cart or product not found in cart"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/cart/items/{productId} [delete]
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
		SendError(w, http.StatusInternalServerError, "failed to remove item from cart", "INTERNAL_SERVER_ERROR", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
