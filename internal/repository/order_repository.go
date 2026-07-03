package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"shopflow/internal/models"
)

// OrderRepository defines database operations for Order.
type OrderRepository interface {
	CreateOrder(ctx context.Context, o *models.Order) error
	GetOrderByID(ctx context.Context, id, userID int) (*models.Order, error)
	ListOrdersByUserID(ctx context.Context, userID int) ([]*models.Order, error)
	ListPendingOrders(ctx context.Context) ([]*models.Order, error)
	UpdateOrderStatus(ctx context.Context, id int, fromStatus, toStatus string) error
}

type sqlOrderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates a new OrderRepository instance.
func NewOrderRepository(db *sql.DB) OrderRepository {
	return &sqlOrderRepository{db: db}
}

// CreateOrder places an order within an atomic transaction.
func (r *sqlOrderRepository) CreateOrder(ctx context.Context, o *models.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Insert order metadata
	queryOrder := `
		INSERT INTO orders (user_id, status, total_amount)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRowContext(ctx, queryOrder, o.UserID, o.Status, o.TotalAmount).
		Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// 2. Insert order items and deduct stock
	queryItem := `
		INSERT INTO order_items (order_id, product_id, quantity, price_at_purchase)
		VALUES ($1, $2, $3, $4)
	`
	queryUpdateStock := `
		UPDATE products
		SET stock = stock - $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND stock >= $1
	`
	for _, item := range o.Items {
		// Insert order item row
		_, err = tx.ExecContext(ctx, queryItem, o.ID, item.ProductID, item.Quantity, item.PriceAtPurchase)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}

		// Deduct product stock atomically ensuring it does not drop below 0
		res, err := tx.ExecContext(ctx, queryUpdateStock, item.Quantity, item.ProductID)
		if err != nil {
			return fmt.Errorf("failed to update stock: %w", err)
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to check rows affected: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("%w: product ID %d has insufficient stock", models.ErrInsufficientStock, item.ProductID)
		}
	}

	// 3. Clear user's cart items
	queryClearCart := `
		DELETE FROM cart_items
		WHERE cart_id = (SELECT id FROM carts WHERE user_id = $1)
	`
	_, err = tx.ExecContext(ctx, queryClearCart, o.UserID)
	if err != nil {
		return fmt.Errorf("failed to clear cart items: %w", err)
	}

	// 4. Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetOrderByID retrieves order metadata and its items.
func (r *sqlOrderRepository) GetOrderByID(ctx context.Context, id, userID int) (*models.Order, error) {
	queryOrder := `
		SELECT id, user_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE id = $1 AND user_id = $2
	`
	o := &models.Order{}
	err := r.db.QueryRowContext(ctx, queryOrder, id, userID).
		Scan(&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order by id: %w", err)
	}

	// Retrieve items
	queryItems := `
		SELECT oi.id, oi.product_id, p.name, oi.quantity, oi.price_at_purchase
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = $1
		ORDER BY oi.id ASC
	`
	rows, err := r.db.QueryContext(ctx, queryItems, o.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []*models.OrderItem
	for rows.Next() {
		item := &models.OrderItem{}
		err := rows.Scan(&item.ID, &item.ProductID, &item.ProductName, &item.Quantity, &item.PriceAtPurchase)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}
	o.Items = items

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during order items rows scan: %w", err)
	}

	return o, nil
}

// ListOrdersByUserID lists all orders placed by a user (metadata only).
func (r *sqlOrderRepository) ListOrdersByUserID(ctx context.Context, userID int) ([]*models.Order, error) {
	query := `
		SELECT id, user_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY id DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		o := &models.Order{Items: []*models.OrderItem{}}
		err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during orders rows scan: %w", err)
	}

	return orders, nil
}

// ListPendingOrders retrieves all orders with status 'pending' that were created more than 10 seconds ago.
// This ensures that orders are only processed after a cool-off period,
// preventing race conditions where multiple workers might try to process the same order simultaneously.
// Only orders that have been in 'pending' status for at least 10 seconds are eligible for processing.
func (r *sqlOrderRepository) ListPendingOrders(ctx context.Context) ([]*models.Order, error) {
	query := `
		SELECT id, user_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE status = $1
			AND created_at <= NOW() - INTERVAL '10 seconds'
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, models.StatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		o := &models.Order{Items: []*models.OrderItem{}}
		err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending order: %w", err)
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during pending orders scan: %w", err)
	}

	return orders, nil
}

// UpdateOrderStatus updates order status in the database.
// It atomically changes the order status only if it is currently in the expected state.
func (r *sqlOrderRepository) UpdateOrderStatus(ctx context.Context, id int, fromStatus string, toStatus string) error {
	query := `
		UPDATE orders
		SET status = $1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		  AND status = $3
	`
	res, err := r.db.ExecContext(ctx, query, toStatus, id, fromStatus)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected on status update: %w", err)
	}

	// No rows updated means either:
	// - Order was already processed.
	// - Current state doesn't match expected state.
	// This is normal in concurrent background processing.
	if rows == 0 {
		return nil
	}

	return nil
}
