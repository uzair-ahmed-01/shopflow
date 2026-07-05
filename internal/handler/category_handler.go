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
		SendError(w, http.StatusInternalServerError, "failed to create category", "INTERNAL_SERVER_ERROR")
		return
	}

	SendJSON(w, http.StatusCreated, c)
}

// ListCategories handles GET /api/v1/categories.
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.ListCategories(r.Context())
	if err != nil {
		SendError(w, http.StatusInternalServerError, "failed to list categories", "INTERNAL_SERVER_ERROR")
		return
	}

	if categories == nil {
		categories = []*models.Category{}
	}

	SendJSON(w, http.StatusOK, categories)
}
