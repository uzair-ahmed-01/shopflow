package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"shopflow/internal/models"
)

// UserRepository defines the database operations for User.
type UserRepository interface {
	CreateUser(ctx context.Context, u *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
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
