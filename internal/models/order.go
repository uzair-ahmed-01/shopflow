package models

import (
	"errors"
	"time"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

const (
	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusCancelled = "cancelled"
)

// Order represents a customer purchase order.
type Order struct {
	ID          int          `json:"id"`
	UserID      int          `json:"user_id"`
	Status      string       `json:"status"`
	TotalAmount int          `json:"total_amount"` // stored in cents/paise
	Items       []*OrderItem `json:"items"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// OrderItem represents a single product purchased within an order.
type OrderItem struct {
	ID              int    `json:"id,omitempty"`
	OrderID         int    `json:"order_id,omitempty"`
	ProductID       int    `json:"product_id"`
	ProductName     string `json:"product_name,omitempty"` // populated on retrieve via join
	Quantity        int    `json:"quantity"`
	PriceAtPurchase int    `json:"price_at_purchase"` // stored in cents/paise
}
