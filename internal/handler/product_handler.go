package handler

import (
	"errors"
	"net/http"
	"strconv"

	"shopflow/internal/models"
	"shopflow/internal/service"
)

// ProductHandler coordinates Product HTTP requests.
type ProductHandler struct {
	service service.ProductService
}

// NewProductHandler creates a new ProductHandler instance.
func NewProductHandler(service service.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

type createProductRequest struct {
	CategoryID  int    `json:"category_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	Stock       int    `json:"stock"`
}

type updateProductRequest struct {
	CategoryID  *int    `json:"category_id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Price       *int    `json:"price"`
	Stock       *int    `json:"stock"`
}

// CreateProduct handles POST /api/v1/products.
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	req, ok := DecodeJSON[createProductRequest](w, r)
	if !ok {
		return
	}

	p, err := h.service.CreateProduct(r.Context(), req.CategoryID, req.Name, req.Description, req.Price, req.Stock)
	if err != nil {
		if errors.Is(err, models.ErrInvalidInput) {
			SendError(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
			return
		}
		SendError(w, http.StatusInternalServerError, "failed to create product", "INTERNAL_SERVER_ERROR")
		return
	}

	SendJSON(w, http.StatusCreated, p)
}

// UpdateProduct handles PUT /api/v1/products/{id}.
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		SendError(w, http.StatusBadRequest, "invalid product ID", "BAD_REQUEST")
		return
	}

	req, ok := DecodeJSON[updateProductRequest](w, r)
	if !ok {
		return
	}

	p, err := h.service.UpdateProduct(r.Context(), id, req.CategoryID, req.Name, req.Description, req.Price, req.Stock)
	if err != nil {
		if errors.Is(err, models.ErrInvalidInput) {
			SendError(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
			return
		}
		if errors.Is(err, models.ErrProductNotFound) {
			SendError(w, http.StatusNotFound, err.Error(), "PRODUCT_NOT_FOUND")
			return
		}
		SendError(w, http.StatusInternalServerError, "failed to update product", "INTERNAL_SERVER_ERROR")
		return
	}

	SendJSON(w, http.StatusOK, p)
}

// DeleteProduct handles DELETE /api/v1/products/{id}.
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		SendError(w, http.StatusBadRequest, "invalid product ID", "BAD_REQUEST")
		return
	}

	err = h.service.DeleteProduct(r.Context(), id)
	if err != nil {
		if errors.Is(err, models.ErrProductNotFound) {
			SendError(w, http.StatusNotFound, err.Error(), "PRODUCT_NOT_FOUND")
			return
		}
		SendError(w, http.StatusInternalServerError, "failed to delete product", "INTERNAL_SERVER_ERROR")
		return
	}

	SendJSON(w, http.StatusOK, map[string]string{
		"message": "Product deleted successfully",
	})
}

// ListProducts handles GET /api/v1/products.
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	products, totalItems, err := h.service.ListProducts(r.Context(), page, limit)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "failed to list products", "INTERNAL_SERVER_ERROR")
		return
	}

	if products == nil {
		products = []*models.Product{}
	}

	response := map[string]any{
		"products": products,
		"pagination": map[string]any{
			"current_page": page,
			"limit":        limit,
			"total_items":  totalItems,
		},
	}

	SendJSON(w, http.StatusOK, response)
}
