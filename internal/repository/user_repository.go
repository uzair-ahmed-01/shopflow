package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"shopflow/internal/models"
)

// UserRepository defines the database operations for User.
type UserRepository interface {
	CreateUser(ctx context.Context, u *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	SaveRefreshToken(ctx context.Context, userID int, token string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}

// sqlUserRepository implements UserRepository using PostgreSQL database/sql.
type sqlUserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(db *sql.DB) UserRepository {
	return &sqlUserRepository{db: db}
}

// CreateUser inserts a user into the database.
func (r *sqlUserRepository) CreateUser(ctx context.Context, u *models.User) error {
	query := `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, u.Name, strings.ToLower(u.Email), u.PasswordHash).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key value") {
			return models.ErrEmailAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByEmail retrieves a user by their email address.
func (r *sqlUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	u := &models.User{}
	err := r.db.QueryRowContext(ctx, query, strings.ToLower(email)).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return u, nil
}

// GetUserByID retrieves a user by their user ID.
func (r *sqlUserRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	u := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return u, nil
}

// SaveRefreshToken persists a new refresh token record.
func (r *sqlUserRepository) SaveRefreshToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, query, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}
	return nil
}

// GetRefreshToken retrieves the refresh token record.
func (r *sqlUserRepository) GetRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, revoked_at
		FROM refresh_tokens
		WHERE token = $1
	`
	t := &models.RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, token).
		Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.CreatedAt, &t.RevokedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return t, nil
}

// RevokeRefreshToken sets revoked_at timestamp for token.
func (r *sqlUserRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = CURRENT_TIMESTAMP
		WHERE token = $1 AND revoked_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected on revoke: %w", err)
	}

	if rows == 0 {
		return models.ErrInvalidToken
	}

	return nil
}
