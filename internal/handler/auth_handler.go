package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"shopflow/internal/models"
	"shopflow/internal/service"
)

// AuthHandler coordinates authentication HTTP requests.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles POST /api/v1/auth/register requests.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "invalid request body format", "BAD_REQUEST")
		return
	}

	user, err := h.authService.Register(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidInput) {
			SendError(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
			return
		}
		if errors.Is(err, models.ErrEmailAlreadyExists) {
			SendError(w, http.StatusConflict, err.Error(), "EMAIL_ALREADY_EXISTS")
			return
		}
		SendError(w, http.StatusInternalServerError, "failed to register user", "INTERNAL_SERVER_ERROR")
		return
	}

	response := map[string]any{
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"created_at": user.CreatedAt,
	}

	SendJSON(w, http.StatusCreated, response)
}

// Login handles POST /api/v1/auth/login requests.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "invalid request body format", "BAD_REQUEST")
		return
	}

	token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			SendError(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
			return
		}
		SendError(w, http.StatusInternalServerError, "failed to authenticate", "INTERNAL_SERVER_ERROR")
		return
	}

	response := map[string]any{
		"token":              token,
		"expires_in_seconds": 86400, // 24 hours
	}

	SendJSON(w, http.StatusOK, response)
}
