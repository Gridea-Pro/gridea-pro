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
	return s.repo.List(ctx)
}

func (s *CategoryService) SaveCategories(ctx context.Context, categories []domain.Category) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repo.SaveAll(ctx, categories)
}

// SaveCategory 创建或更新分类
// originalID: 若为空则创建新分类；若非空则按 ID 更新
func (s *CategoryService) SaveCategory(ctx context.Context, category domain.Category, originalID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if originalID == "" {
		// 新建：Create 会自动生成 UUID
		return s.repo.Create(ctx, &category)
	}

	// 更新：保持 ID 不变
	category.ID = originalID
	return s.repo.Update(ctx, originalID, &category)
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repo.Delete(ctx, id)
}

// GetOrCreateCategory 按名称查找分类，不存在则创建（自动生成 UUID）
func (s *CategoryService) GetOrCreateCategory(ctx context.Context, name string) (domain.Category, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	categories, err := s.repo.List(ctx)
	if err != nil {
		return domain.Category{}, err
	}

	// 1. 按名称查找（返回已有分类，包含其 ID）
	for _, c := range categories {
		if c.Name == name {
			return c, nil
		}
	}

	// 2. 创建新分类（Create 自动生成 UUID）
	newCategory := domain.Category{
		Name: name,
		Slug: name,
	}
	if err := s.repo.Create(ctx, &newCategory); err != nil {
		return domain.Category{}, err
	}

	return newCategory, nil
}

// GetByID 按 ID 获取分类
func (s *CategoryService) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.repo.GetByID(ctx, id)
}
