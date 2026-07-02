package models

import (
	"errors"
	"time"
)

var (
	ErrCartNotFound = errors.New("cart not found")
)

// Cart represents a user's active shopping cart.
type Cart struct {
	ID        int         `json:"id"`
	UserID    int         `json:"user_id"`
	Items     []*CartItem `json:"items"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// CartItem represents an item and quantity in a cart.
type CartItem struct {
	ProductID   int    `json:"product_id"`
	ProductName string `json:"product_name,omitempty"`
	Price       int    `json:"price,omitempty"`
	Quantity    int    `json:"quantity"`
}
