package handler

import (
	"errors"
	"net/http"

	"shopflow/internal/models"
	"shopflow/internal/service"
)

// CategoryHandler coordinates Category HTTP requests.
type CategoryHandler struct {
	service service.CategoryService
}

// NewCategoryHandler creates a new CategoryHandler instance.
func NewCategoryHandler(service service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

type createCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateCategory handles POST /api/v1/categories.
// @Summary Create category
// @Description Create a new product category. Admin only.
// @Tags Category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body createCategoryRequest true "Category details"
// @Success 201 {object} SuccessResponse[models.Category] "Category created successfully"
// @Failure 400 {object} ErrorResponse "Invalid input data"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 409 {object} ErrorResponse "Category already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/categories [post]
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	req, ok := DecodeJSON[createCategoryRequest](w, r)
	if !ok {
		return
	}

	c, err := h.service.CreateCategory(r.Context(), req.Name, req.Description)
	if err != nil {
		if errors.Is(err, models.ErrInvalidInput) {
			SendError(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
			return
		}
		if errors.Is(err, models.ErrCategoryAlreadyExists) {
			SendError(w, http.StatusConflict, err.Error(), "CATEGORY_ALREADY_EXISTS")
			return
		}
		SendError(w, http.StatusInternalServerError, "failed to create category", "INTERNAL_SERVER_ERROR", err)
		return
	}

	SendJSON(w, http.StatusCreated, c)
}

// ListCategories handles GET /api/v1/categories.
// @Summary List categories
// @Description Retrieve a list of all product categories. Public.
// @Tags Category
// @Produce json
// @Success 200 {object} SuccessResponse[[]models.Category] "Categories retrieved successfully"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/categories [get]
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.ListCategories(r.Context())
	if err != nil {
		SendError(w, http.StatusInternalServerError, "failed to list categories", "INTERNAL_SERVER_ERROR", err)
		return
	}

	if categories == nil {
		categories = []*models.Category{}
	}

	SendJSON(w, http.StatusOK, categories)
}
