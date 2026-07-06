package handler

import (
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
	Role     string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Register handles POST /api/v1/auth/register requests.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	req, ok := DecodeJSON[registerRequest](w, r)
	if !ok {
		return
	}

	user, err := h.authService.Register(r.Context(), req.Name, req.Email, req.Password, req.Role)
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
	req, ok := DecodeJSON[loginRequest](w, r)
	if !ok {
		return
	}

	accessToken, refreshToken, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			SendError(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
			return
		}
		SendError(w, http.StatusInternalServerError, "failed to authenticate", "INTERNAL_SERVER_ERROR")
		return
	}

	response := map[string]any{
		"access_token":       accessToken,
		"refresh_token":      refreshToken,
		"expires_in_seconds": 900, // 15 minutes
	}

	SendJSON(w, http.StatusOK, response)
}

// Refresh handles POST /api/v1/auth/refresh requests.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	req, ok := DecodeJSON[refreshTokenRequest](w, r)
	if !ok {
		return
	}

	if req.RefreshToken == "" {
		SendError(w, http.StatusBadRequest, "refresh token is required", "BAD_REQUEST")
		return
	}

	newAccessToken, newRefreshToken, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, models.ErrInvalidToken) || errors.Is(err, models.ErrUserNotFound) {
			SendError(w, http.StatusBadRequest, err.Error(), "INVALID_TOKEN")
			return
		}
		SendError(w, http.StatusInternalServerError, "failed to refresh token", "INTERNAL_SERVER_ERROR")
		return
	}

	response := map[string]any{
		"access_token":       newAccessToken,
		"refresh_token":      newRefreshToken,
		"expires_in_seconds": 900, // 15 minutes
	}

	SendJSON(w, http.StatusOK, response)
}

// Logout handles POST /api/v1/auth/logout requests.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	req, ok := DecodeJSON[refreshTokenRequest](w, r)
	if !ok {
		return
	}

	if req.RefreshToken == "" {
		SendError(w, http.StatusBadRequest, "refresh token is required", "BAD_REQUEST")
		return
	}

	err := h.authService.Logout(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, models.ErrInvalidToken) {
			SendError(w, http.StatusBadRequest, err.Error(), "INVALID_TOKEN")
			return
		}
		SendError(w, http.StatusInternalServerError, "failed to logout", "INTERNAL_SERVER_ERROR")
		return
	}

	SendJSON(w, http.StatusOK, map[string]string{"message": "successfully logged out"})
}
