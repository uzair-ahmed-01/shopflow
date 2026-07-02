package service

import (
	"context"
	"fmt"

	"shopflow/internal/models"
	"shopflow/internal/repository"
)

// OrderService defines business operations for Order.
type OrderService interface {
	PlaceOrder(ctx context.Context, userID int) (*models.Order, error)
	GetOrderDetails(ctx context.Context, id, userID int) (*models.Order, error)
	ListOrders(ctx context.Context, userID int) ([]*models.Order, error)
}

type orderService struct {
	orderRepo   repository.OrderRepository
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

// NewOrderService creates a new OrderService instance.
func NewOrderService(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

// PlaceOrder executes stock verification, total calculation, writes order transaction.
func (s *orderService) PlaceOrder(ctx context.Context, userID int) (*models.Order, error) {
	// 1. Get user's cart
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cart: %w", err)
	}

	// 2. Fetch cart items
	cartItems, err := s.cartRepo.GetCartItems(ctx, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cart items: %w", err)
	}

	if len(cartItems) == 0 {
		return nil, fmt.Errorf("%w: cannot place order for an empty cart", models.ErrInvalidInput)
	}

	// 3. Verify stock and calculate total amount
	var orderItems []*models.OrderItem
	var totalAmount int

	for _, item := range cartItems {
		product, err := s.productRepo.GetProductByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve product info: %w", err)
		}

		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("%w: product %s has insufficient stock", models.ErrInsufficientStock, product.Name)
		}

		totalAmount += item.Quantity * product.Price

		orderItems = append(orderItems, &models.OrderItem{
			ProductID:       item.ProductID,
			Quantity:        item.Quantity,
			PriceAtPurchase: product.Price,
		})
	}

	// 4. Construct Order object
	o := &models.Order{
		UserID:      userID,
		Status:      models.StatusPending,
		TotalAmount: totalAmount,
		Items:       orderItems,
	}

	// 5. Save order inside transaction (deducts stock, clears cart)
	if err := s.orderRepo.CreateOrder(ctx, o); err != nil {
		return nil, err
	}

	return o, nil
}

// GetOrderDetails fetches complete order details (metadata + items).
func (s *orderService) GetOrderDetails(ctx context.Context, id, userID int) (*models.Order, error) {
	return s.orderRepo.GetOrderByID(ctx, id, userID)
}

// ListOrders fetches all orders (metadata only) placed by the user.
func (s *orderService) ListOrders(ctx context.Context, userID int) ([]*models.Order, error) {
	return s.orderRepo.ListOrdersByUserID(ctx, userID)
}
