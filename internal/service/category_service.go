package service

import (
	"context"
	"fmt"
	"strings"

	"shopflow/internal/models"
	"shopflow/internal/repository"
)

// CategoryService defines business operations for Category.
type CategoryService interface {
	CreateCategory(ctx context.Context, name, description string) (*models.Category, error)
	ListCategories(ctx context.Context) ([]*models.Category, error)
}

type categoryService struct {
	repo repository.CategoryRepository
}

// NewCategoryService creates a new CategoryService instance.
func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

// CreateCategory validates input and inserts a Category.
func (s *categoryService) CreateCategory(ctx context.Context, name, description string) (*models.Category, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if name == "" {
		return nil, fmt.Errorf("%w: category name cannot be empty", models.ErrInvalidInput)
	}

	c := &models.Category{
		Name:        name,
		Description: description,
	}

	if err := s.repo.CreateCategory(ctx, c); err != nil {
		return nil, err
	}

	return c, nil
}

// ListCategories returns all categories.
func (s *categoryService) ListCategories(ctx context.Context) ([]*models.Category, error) {
	return s.repo.ListCategories(ctx)
}
