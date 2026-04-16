package service

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"sync"
)

type CategoryService struct {
	repo     domain.CategoryRepository
	postRepo domain.PostRepository
	mu       sync.RWMutex
}

func NewCategoryService(repo domain.CategoryRepository, postRepo domain.PostRepository) *CategoryService {
	return &CategoryService{repo: repo, postRepo: postRepo}
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

	if originalID == "" {
		err := s.repo.Create(ctx, &category)
		s.mu.Unlock()
		return err
	}

	existing, err := s.repo.GetByID(ctx, originalID)
	if err != nil {
		s.mu.Unlock()
		return err
	}

	isRename := existing.Name != category.Name
	category.ID = originalID
	err = s.repo.Update(ctx, originalID, &category)
	s.mu.Unlock()
	if err != nil {
		return err
	}

	if isRename {
		return s.cascadeCategoryRename(ctx, existing.Name, category.Name)
	}
	return nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id string) error {
	s.mu.Lock()

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	catName := existing.Name

	err = s.repo.Delete(ctx, id)
	s.mu.Unlock()
	if err != nil {
		return err
	}

	return s.cascadeCategoryDelete(ctx, id, catName)
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

func (s *CategoryService) cascadeCategoryRename(ctx context.Context, oldName, newName string) error {
	posts, _, err := s.postRepo.List(ctx, 1, 10000)
	if err != nil {
		return err
	}
	for i := range posts {
		changed := false
		for j, c := range posts[i].Categories {
			if c == oldName {
				posts[i].Categories[j] = newName
				changed = true
			}
		}
		if changed {
			if err := s.postRepo.Update(ctx, &posts[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *CategoryService) cascadeCategoryDelete(ctx context.Context, categoryID, categoryName string) error {
	posts, _, err := s.postRepo.List(ctx, 1, 10000)
	if err != nil {
		return err
	}
	for i := range posts {
		changed := false
		var newCats []string
		for _, c := range posts[i].Categories {
			if c != categoryName {
				newCats = append(newCats, c)
			} else {
				changed = true
			}
		}
		var newCatIDs []string
		for _, id := range posts[i].CategoryIDs {
			if id != categoryID {
				newCatIDs = append(newCatIDs, id)
			} else {
				changed = true
			}
		}
		if changed {
			posts[i].Categories = newCats
			if posts[i].Categories == nil {
				posts[i].Categories = []string{}
			}
			posts[i].CategoryIDs = newCatIDs
			if posts[i].CategoryIDs == nil {
				posts[i].CategoryIDs = []string{}
			}
			if err := s.postRepo.Update(ctx, &posts[i]); err != nil {
				return err
			}
		}
	}
	return nil
}
