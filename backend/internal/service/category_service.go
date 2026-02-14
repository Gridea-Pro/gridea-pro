package service

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"sync"
)

type CategoryService struct {
	repo domain.CategoryRepository
	mu   sync.RWMutex
}

func NewCategoryService(repo domain.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) LoadCategories(ctx context.Context) ([]domain.Category, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.repo.GetAll(ctx)
}

func (s *CategoryService) SaveCategories(ctx context.Context, categories []domain.Category) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repo.SaveAll(ctx, categories)
}

func (s *CategoryService) SaveCategory(ctx context.Context, category domain.Category, originalSlug string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	categories, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	found := false
	for i, c := range categories {
		if c.Slug == originalSlug {
			categories[i] = category
			found = true
			break
		}
	}

	if !found {
		categories = append(categories, category)
	}

	return s.repo.SaveAll(ctx, categories)
}

func (s *CategoryService) DeleteCategory(ctx context.Context, slug string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	categories, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	var newCategories []domain.Category
	for _, c := range categories {
		if c.Slug != slug {
			newCategories = append(newCategories, c)
		}
	}

	return s.repo.SaveAll(ctx, newCategories)
}

// GetOrCreateCategory gets an existing category by name or creates a new one
func (s *CategoryService) GetOrCreateCategory(ctx context.Context, name string) (domain.Category, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	categories, err := s.repo.GetAll(ctx)
	if err != nil {
		return domain.Category{}, err
	}

	// 1. Check if exists
	for _, c := range categories {
		if c.Name == name {
			return c, nil
		}
	}

	// 2. Create New Category
	// Use name as slug for simplicity as per plan
	newCategory := domain.Category{
		Name: name,
		Slug: name,
	}

	// 3. Save
	categories = append(categories, newCategory)
	if err := s.repo.SaveAll(ctx, categories); err != nil {
		return domain.Category{}, err
	}

	return newCategory, nil
}
