package handler

import (
	"errors"
	"net/http"
	"time"

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

type registerResponse struct {
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type loginResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

type logoutResponse struct {
	Message string `json:"message" example:"successfully logged out"`
}

// Register handles POST /api/v1/auth/register requests.
// @Summary Register a new user
// @Description Create a customer or admin account.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body handler.registerRequest true "User registration details"
// @Success 201 {object} handler.SuccessResponse[handler.registerResponse] "User registered successfully"
// @Failure 400 {object} handler.ErrorResponse "Invalid input data"
// @Failure 409 {object} handler.ErrorResponse "Email already exists"
// @Failure 500 {object} handler.ErrorResponse "Internal server error"
// @Router /api/v1/auth/register [post]
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
		SendError(w, http.StatusInternalServerError, "failed to register user", "INTERNAL_SERVER_ERROR", err)
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
// @Summary Authenticate user
// @Description Log in with email and password to receive JWT tokens.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body handler.loginRequest true "User login credentials"
// @Success 200 {object} handler.SuccessResponse[handler.loginResponse] "Successful login"
// @Failure 400 {object} handler.ErrorResponse "Invalid request payload"
// @Failure 401 {object} handler.ErrorResponse "Invalid credentials"
// @Failure 500 {object} handler.ErrorResponse "Internal server error"
// @Router /api/v1/auth/login [post]
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
		SendError(w, http.StatusInternalServerError, "failed to authenticate", "INTERNAL_SERVER_ERROR", err)
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
// @Summary Refresh JWT tokens
// @Description Rotate expired access token using a valid refresh token.
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body handler.refreshTokenRequest true "Refresh token"
// @Success 200 {object} handler.SuccessResponse[handler.loginResponse] "Tokens refreshed successfully"
// @Failure 400 {object} handler.ErrorResponse "Invalid refresh token or token missing"
// @Failure 500 {object} handler.ErrorResponse "Internal server error"
// @Router /api/v1/auth/refresh [post]
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
		SendError(w, http.StatusInternalServerError, "failed to refresh token", "INTERNAL_SERVER_ERROR", err)
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
// @Summary Log out user
// @Description Invalidate the session refresh token.
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body handler.refreshTokenRequest true "Refresh token to invalidate"
// @Success 200 {object} handler.SuccessResponse[handler.logoutResponse] "Successfully logged out"
// @Failure 400 {object} handler.ErrorResponse "Invalid refresh token or token missing"
// @Failure 500 {object} handler.ErrorResponse "Internal server error"
// @Router /api/v1/auth/logout [post]
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
		SendError(w, http.StatusInternalServerError, "failed to logout", "INTERNAL_SERVER_ERROR", err)
		return
	}

	SendJSON(w, http.StatusOK, map[string]string{"message": "successfully logged out"})
}
