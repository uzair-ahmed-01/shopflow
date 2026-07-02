package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"shopflow/internal/models"
)

// CartRepository defines database operations for Cart.
type CartRepository interface {
	GetCartByUserID(ctx context.Context, userID int) (*models.Cart, error)
	CreateCart(ctx context.Context, userID int) (*models.Cart, error)
	GetCartItems(ctx context.Context, cartID int) ([]*models.CartItem, error)
	AddOrUpdateCartItem(ctx context.Context, cartID, productID, quantity int) error
	RemoveCartItem(ctx context.Context, cartID, productID int) error
}

type sqlCartRepository struct {
	db *sql.DB
}

// NewCartRepository creates a new CartRepository instance.
func NewCartRepository(db *sql.DB) CartRepository {
	return &sqlCartRepository{db: db}
}

// GetCartByUserID retrieves the cart for a user.
func (r *sqlCartRepository) GetCartByUserID(ctx context.Context, userID int) (*models.Cart, error) {
	query := `
		SELECT id, user_id, created_at, updated_at
		FROM carts
		WHERE user_id = $1
	`
	c := &models.Cart{}
	err := r.db.QueryRowContext(ctx, query, userID).
		Scan(&c.ID, &c.UserID, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrCartNotFound
		}
		return nil, fmt.Errorf("failed to get cart by user id: %w", err)
	}

	return c, nil
}

// CreateCart creates a new empty cart for a user.
func (r *sqlCartRepository) CreateCart(ctx context.Context, userID int) (*models.Cart, error) {
	query := `
		INSERT INTO carts (user_id)
		VALUES ($1)
		RETURNING id, user_id, created_at, updated_at
	`
	c := &models.Cart{}
	err := r.db.QueryRowContext(ctx, query, userID).
		Scan(&c.ID, &c.UserID, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create cart: %w", err)
	}

	return c, nil
}

// GetCartItems retrieves all items in a cart with product names and prices.
func (r *sqlCartRepository) GetCartItems(ctx context.Context, cartID int) ([]*models.CartItem, error) {
	query := `
		SELECT ci.product_id, p.name, p.price, ci.quantity
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.cart_id = $1
		ORDER BY p.name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, cartID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}
	defer rows.Close()

	var items []*models.CartItem
	for rows.Next() {
		item := &models.CartItem{}
		err := rows.Scan(&item.ProductID, &item.ProductName, &item.Price, &item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during cart items rows scan: %w", err)
	}

	return items, nil
}

// AddOrUpdateCartItem adds an item or updates its quantity (upsert).
func (r *sqlCartRepository) AddOrUpdateCartItem(ctx context.Context, cartID, productID, quantity int) error {
	query := `
		INSERT INTO cart_items (cart_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (cart_id, product_id)
		DO UPDATE SET quantity = $3
	`
	_, err := r.db.ExecContext(ctx, query, cartID, productID, quantity)
	if err != nil {
		return fmt.Errorf("failed to upsert cart item: %w", err)
	}
	return nil
}

// RemoveCartItem deletes an item from the cart.
func (r *sqlCartRepository) RemoveCartItem(ctx context.Context, cartID, productID int) error {
	query := `
		DELETE FROM cart_items
		WHERE cart_id = $1 AND product_id = $2
	`
	res, err := r.db.ExecContext(ctx, query, cartID, productID)
	if err != nil {
		return fmt.Errorf("failed to delete cart item: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return models.ErrProductNotFound // Using ErrProductNotFound as the item is missing
	}
	return nil
}
