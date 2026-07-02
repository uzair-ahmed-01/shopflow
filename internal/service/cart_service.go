package service

import (
	"context"
	"errors"
	"fmt"

	"shopflow/internal/models"
	"shopflow/internal/repository"
)

// CartService defines business operations for Cart.
type CartService interface {
	GetOrCreateCart(ctx context.Context, userID int) (*models.Cart, error)
	AddOrUpdateItem(ctx context.Context, userID int, productID, quantity int) error
	RemoveItem(ctx context.Context, userID int, productID int) error
}

type cartService struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

// NewCartService creates a new CartService instance.
func NewCartService(cartRepo repository.CartRepository, productRepo repository.ProductRepository) CartService {
	return &cartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

// GetOrCreateCart retrieves the user's cart, creating one if not found.
func (s *cartService) GetOrCreateCart(ctx context.Context, userID int) (*models.Cart, error) {
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, models.ErrCartNotFound) {
			cart, err = s.cartRepo.CreateCart(ctx, userID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	items, err := s.cartRepo.GetCartItems(ctx, cart.ID)
	if err != nil {
		return nil, err
	}
	cart.Items = items

	return cart, nil
}

// AddOrUpdateItem validates stock limits and adds or updates an item in the cart.
func (s *cartService) AddOrUpdateItem(ctx context.Context, userID int, productID, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("%w: quantity must be greater than zero", models.ErrInvalidInput)
	}

	// Verify product exists and check stock
	product, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		return err // forward ErrProductNotFound
	}

	if product.Stock < quantity {
		return fmt.Errorf("%w: stock is %d but requested %d", models.ErrInsufficientStock, product.Stock, quantity)
	}

	// Get or create user's cart
	cart, err := s.GetOrCreateCart(ctx, userID)
	if err != nil {
		return err
	}

	return s.cartRepo.AddOrUpdateCartItem(ctx, cart.ID, productID, quantity)
}

// RemoveItem deletes an item from the user's cart.
func (s *cartService) RemoveItem(ctx context.Context, userID int, productID int) error {
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return err // forward ErrCartNotFound
	}

	return s.cartRepo.RemoveCartItem(ctx, cart.ID, productID)
}
