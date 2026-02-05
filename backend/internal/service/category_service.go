package service

import (
	"context"
	"gridea-pro/backend/internal/domain"
)

type CategoryService struct {
	repo domain.CategoryRepository
}

func NewCategoryService(repo domain.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) LoadCategories(ctx context.Context) ([]domain.Category, error) {
	return s.repo.GetAll(ctx)
}

func (s *CategoryService) SaveCategories(ctx context.Context, categories []domain.Category) error {
	return s.repo.SaveAll(ctx, categories)
}

func (s *CategoryService) SaveCategory(ctx context.Context, category domain.Category, originalSlug string) error {
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
