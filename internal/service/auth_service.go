package service

import (
	"context"
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
	Login(ctx context.Context, email, password string) (string, error)
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

// Login validates credentials and generates JWT session token.
func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.TrimSpace(email)

	if email == "" || password == "" {
		return "", models.ErrInvalidCredentials
	}

	// Fetch user
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return "", models.ErrInvalidCredentials
		}
		return "", fmt.Errorf("failed to login: %w", err)
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", models.ErrInvalidCredentials
	}

	// Generate JWT claims
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	// Sign JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
