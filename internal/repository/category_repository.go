package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"shopflow/internal/models"
)

// CategoryRepository defines database operations for Category.
type CategoryRepository interface {
	CreateCategory(ctx context.Context, c *models.Category) error
	ListCategories(ctx context.Context) ([]*models.Category, error)
	GetCategoryByID(ctx context.Context, id int) (*models.Category, error)
}

type sqlCategoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository creates a new CategoryRepository instance.
func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &sqlCategoryRepository{db: db}
}

// CreateCategory inserts a category into the database.
func (r *sqlCategoryRepository) CreateCategory(ctx context.Context, c *models.Category) error {
	query := `
		INSERT INTO categories (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, c.Name, c.Description).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key value") {
			return models.ErrCategoryAlreadyExists
		}
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

// ListCategories retrieves all categories.
func (r *sqlCategoryRepository) ListCategories(ctx context.Context) ([]*models.Category, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM categories
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		c := &models.Category{}
		err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during category rows scan: %w", err)
	}

	return categories, nil
}

// GetCategoryByID retrieves a category by its ID.
func (r *sqlCategoryRepository) GetCategoryByID(ctx context.Context, id int) (*models.Category, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM categories
		WHERE id = $1
	`
	c := &models.Category{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category by id: %w", err)
	}

	return c, nil
}
