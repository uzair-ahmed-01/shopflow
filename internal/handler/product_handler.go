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

type deleteProductResponse struct {
	Message string `json:"message" example:"Product deleted successfully"`
}

type paginationInfo struct {
	CurrentPage int `json:"current_page" example:"1"`
	Limit       int `json:"limit" example:"10"`
	TotalItems  int `json:"total_items" example:"25"`
}

type listProductsResponse struct {
	Products   []*models.Product `json:"products"`
	Pagination paginationInfo    `json:"pagination"`
}

// CreateProduct handles POST /api/v1/products.
// @Summary Create product
// @Description Create a new product. Admin only.
// @Tags Product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body createProductRequest true "Product details"
// @Success 201 {object} SuccessResponse[models.Product] "Product created successfully"
// @Failure 400 {object} ErrorResponse "Invalid input data"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/products [post]
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
		SendError(w, http.StatusInternalServerError, "failed to create product", "INTERNAL_SERVER_ERROR", err)
		return
	}

	SendJSON(w, http.StatusCreated, p)
}

// UpdateProduct handles PUT /api/v1/products/{id}.
// @Summary Update product
// @Description Update an existing product. Admin only.
// @Tags Product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param body body updateProductRequest true "Product update fields"
// @Success 200 {object} SuccessResponse[models.Product] "Product updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid input or ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Product not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/products/{id} [put]
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
		SendError(w, http.StatusInternalServerError, "failed to update product", "INTERNAL_SERVER_ERROR", err)
		return
	}

	SendJSON(w, http.StatusOK, p)
}

// DeleteProduct handles DELETE /api/v1/products/{id}.
// @Summary Delete product
// @Description Delete a product by its ID. Admin only.
// @Tags Product
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 200 {object} SuccessResponse[deleteProductResponse] "Product deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid product ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Product not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/products/{id} [delete]
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
		SendError(w, http.StatusInternalServerError, "failed to delete product", "INTERNAL_SERVER_ERROR", err)
		return
	}

	SendJSON(w, http.StatusOK, map[string]string{
		"message": "Product deleted successfully",
	})
}

// ListProducts handles GET /api/v1/products.
// @Summary List products
// @Description Retrieve a list of products with optional pagination. Public.
// @Tags Product
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10)"
// @Success 200 {object} SuccessResponse[listProductsResponse] "Products retrieved successfully"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/products [get]
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
		SendError(w, http.StatusInternalServerError, "failed to list products", "INTERNAL_SERVER_ERROR", err)
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
