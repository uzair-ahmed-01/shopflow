package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"shopflow/internal/models"
)

// ProductRepository defines database operations for Product.
type ProductRepository interface {
	CreateProduct(ctx context.Context, p *models.Product) error
	UpdateProduct(ctx context.Context, p *models.Product) error
	DeleteProduct(ctx context.Context, id int) error
	GetProductByID(ctx context.Context, id int) (*models.Product, error)
	ListProducts(ctx context.Context, limit, offset int) ([]*models.Product, int, error)
}

type sqlProductRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new ProductRepository instance.
func NewProductRepository(db *sql.DB) ProductRepository {
	return &sqlProductRepository{db: db}
}

// CreateProduct inserts a product into the database.
func (r *sqlProductRepository) CreateProduct(ctx context.Context, p *models.Product) error {
	query := `
		INSERT INTO products (category_id, name, description, price, stock)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, p.CategoryID, p.Name, p.Description, p.Price, p.Stock).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

// UpdateProduct updates an existing product.
func (r *sqlProductRepository) UpdateProduct(ctx context.Context, p *models.Product) error {
	query := `
		UPDATE products
		SET category_id = $1, name = $2, description = $3, price = $4, stock = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query, p.CategoryID, p.Name, p.Description, p.Price, p.Stock, p.ID).
		Scan(&p.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ErrProductNotFound
		}
		return fmt.Errorf("failed to update product: %w", err)
	}
	return nil
}

// DeleteProduct deletes a product from the database.
func (r *sqlProductRepository) DeleteProduct(ctx context.Context, id int) error {
	query := `
		DELETE FROM products
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return models.ErrProductNotFound
	}
	return nil
}

// GetProductByID retrieves a product by its ID.
func (r *sqlProductRepository) GetProductByID(ctx context.Context, id int) (*models.Product, error) {
	query := `
		SELECT id, category_id, name, description, price, stock, created_at, updated_at
		FROM products
		WHERE id = $1
	`
	p := &models.Product{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}

	return p, nil
}

// ListProducts retrieves a page of products along with the total count.
func (r *sqlProductRepository) ListProducts(ctx context.Context, limit, offset int) ([]*models.Product, int, error) {
	countQuery := `SELECT COUNT(*) FROM products`
	var totalItems int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&totalItems)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	query := `
		SELECT id, category_id, name, description, price, stock, created_at, updated_at
		FROM products
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products query: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		p := &models.Product{}
		err := rows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error during products rows scan: %w", err)
	}

	return products, totalItems, nil
}
