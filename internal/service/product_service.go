package service

import (
	"context"
	"fmt"
	"strings"

	"shopflow/internal/models"
	"shopflow/internal/repository"
)

// ProductService defines business operations for Product.
type ProductService interface {
	CreateProduct(ctx context.Context, categoryID int, name, description string, price, stock int) (*models.Product, error)
	UpdateProduct(ctx context.Context, id int, categoryID *int, name *string, description *string, price *int, stock *int) (*models.Product, error)
	DeleteProduct(ctx context.Context, id int) error
	GetProductByID(ctx context.Context, id int) (*models.Product, error)
	ListProducts(ctx context.Context, page, limit int) ([]*models.Product, int, error)
}

type productService struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
}

// NewProductService creates a new ProductService instance.
func NewProductService(productRepo repository.ProductRepository, categoryRepo repository.CategoryRepository) ProductService {
	return &productService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

// CreateProduct validates inputs, checks category existence, and inserts the product.
func (s *productService) CreateProduct(ctx context.Context, categoryID int, name, description string, price, stock int) (*models.Product, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if name == "" {
		return nil, fmt.Errorf("%w: product name cannot be empty", models.ErrInvalidInput)
	}
	if price <= 0 {
		return nil, fmt.Errorf("%w: price must be greater than zero", models.ErrInvalidInput)
	}
	if stock < 0 {
		return nil, fmt.Errorf("%w: stock cannot be negative", models.ErrInvalidInput)
	}

	// Verify category exists
	_, err := s.categoryRepo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("%w: category does not exist", models.ErrInvalidInput)
	}

	p := &models.Product{
		CategoryID:  categoryID,
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
	}

	if err := s.productRepo.CreateProduct(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

// UpdateProduct updates product properties after validation.
func (s *productService) UpdateProduct(ctx context.Context, id int, categoryID *int, name *string, description *string, price *int, stock *int) (*models.Product, error) {
	// Verify product exists
	p, err := s.productRepo.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates if present
	if categoryID != nil {
		// Verify category exists
		_, err = s.categoryRepo.GetCategoryByID(ctx, *categoryID)
		if err != nil {
			return nil, fmt.Errorf("%w: category does not exist", models.ErrInvalidInput)
		}
		p.CategoryID = *categoryID
	}

	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return nil, fmt.Errorf("%w: product name cannot be empty", models.ErrInvalidInput)
		}
		p.Name = trimmed
	}

	if description != nil {
		p.Description = strings.TrimSpace(*description)
	}

	if price != nil {
		if *price <= 0 {
			return nil, fmt.Errorf("%w: price must be greater than zero", models.ErrInvalidInput)
		}
		p.Price = *price
	}

	if stock != nil {
		if *stock < 0 {
			return nil, fmt.Errorf("%w: stock cannot be negative", models.ErrInvalidInput)
		}
		p.Stock = *stock
	}

	if err := s.productRepo.UpdateProduct(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

// DeleteProduct deletes a product by its ID.
func (s *productService) DeleteProduct(ctx context.Context, id int) error {
	return s.productRepo.DeleteProduct(ctx, id)
}

// GetProductByID retrieves a product by its ID.
func (s *productService) GetProductByID(ctx context.Context, id int) (*models.Product, error) {
	return s.productRepo.GetProductByID(ctx, id)
}

// ListProducts returns paginated list of products and the total items count.
func (s *productService) ListProducts(ctx context.Context, page, limit int) ([]*models.Product, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	return s.productRepo.ListProducts(ctx, limit, offset)
}
