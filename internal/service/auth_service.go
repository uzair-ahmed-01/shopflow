package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"shopflow/internal/config"
	"shopflow/internal/models"
	"shopflow/internal/repository"
)

// AuthService defines authentication business operations.
type AuthService interface {
	Register(ctx context.Context, name, email, password string) (*models.User, error)
	Login(ctx context.Context, email, password string) (string, string, error) // Returns access_token, refresh_token, error
	RefreshToken(ctx context.Context, token string) (string, string, error)     // Returns new_access_token, new_refresh_token, error
	Logout(ctx context.Context, token string) error
}

type authService struct {
	repo repository.UserRepository
	cfg  *config.Config
}

// NewAuthService creates a new AuthService implementation.
func NewAuthService(repo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{
		repo: repo,
		cfg:  cfg,
	}
}

// Register validates, hashes password, and persists a new user.
func (s *authService) Register(ctx context.Context, name, email, password string) (*models.User, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)

	// Validations
	if name == "" || len(name) < 2 || len(name) > 100 {
		return nil, fmt.Errorf("%w: name must be between 2 and 100 characters", models.ErrInvalidInput)
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("%w: invalid email address format", models.ErrInvalidInput)
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("%w: password must be at least 8 characters long", models.ErrInvalidInput)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err // Can return models.ErrEmailAlreadyExists
	}

	return user, nil
}

// Login validates credentials, generates JWT access token, and creates refresh token.
func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	email = strings.TrimSpace(email)

	if email == "" || password == "" {
		return "", "", models.ErrInvalidCredentials
	}

	// Fetch user
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return "", "", models.ErrInvalidCredentials
		}
		return "", "", fmt.Errorf("failed to login: %w", err)
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", "", models.ErrInvalidCredentials
	}

	// Generate access token (expiring in 15 minutes)
	accessToken, err := s.generateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", err
	}

	// Generate and save refresh token (expiring in 7 days)
	refreshToken, err := s.generateAndSaveRefreshToken(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// RefreshToken rotates tokens by invalidating old refresh token and generating new access + refresh pair.
func (s *authService) RefreshToken(ctx context.Context, token string) (string, string, error) {
	// Fetch refresh token record
	t, err := s.repo.GetRefreshToken(ctx, token)
	if err != nil {
		return "", "", err
	}

	// Verify expiration and revocation
	if t.ExpiresAt.Before(time.Now()) {
		return "", "", fmt.Errorf("%w: token expired", models.ErrInvalidToken)
	}
	if t.RevokedAt != nil {
		return "", "", fmt.Errorf("%w: token already revoked", models.ErrInvalidToken)
	}

	// Revoke the old refresh token (Compare-and-Swap prevents double reuse)
	if err := s.repo.RevokeRefreshToken(ctx, token); err != nil {
		return "", "", err
	}

	// Fetch user details to generate new access token
	user, err := s.repo.GetUserByID(ctx, t.UserID)
	if err != nil {
		return "", "", err
	}

	// Generate new access token
	newAccessToken, err := s.generateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", err
	}

	// Generate new refresh token
	newRefreshToken, err := s.generateAndSaveRefreshToken(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

// Logout revokes the given refresh token.
func (s *authService) Logout(ctx context.Context, token string) error {
	return s.repo.RevokeRefreshToken(ctx, token)
}

// generateAccessToken creates a standard short-lived signed JWT.
func (s *authService) generateAccessToken(userID int, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}
	return tokenString, nil
}

// generateAndSaveRefreshToken creates a random secure token and stores it in database.
func (s *authService) generateAndSaveRefreshToken(ctx context.Context, userID int) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes for refresh token: %w", err)
	}
	token := hex.EncodeToString(b)

	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days expiration

	if err := s.repo.SaveRefreshToken(ctx, userID, token, expiresAt); err != nil {
		return "", err
	}

	return token, nil
}
