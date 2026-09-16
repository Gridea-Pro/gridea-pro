package service

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/utils"
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
	// 新建：必须校验格式，在持锁前拦掉 —— 避免 Wails 调用链拿不到锁时堆积
	if originalID == "" {
		if err := utils.ValidateSlug(category.Slug); err != nil {
			return fmt.Errorf("%w：分类 URL slug %q 不合法，只能包含字母、数字和连字符", err, category.Slug)
		}
		s.mu.Lock()
		err := s.repo.Create(ctx, &category)
		s.mu.Unlock()
		return err
	}

	s.mu.Lock()

	existing, err := s.repo.GetByID(ctx, originalID)
	if err != nil {
		s.mu.Unlock()
		return err
	}

	// 编辑：只有 slug 真被改动时才校验格式。历史数据里可能存在不符合当前规则的
	// slug（旧版本写入、或经 MCP 由外部写入），不能因此连带阻止用户修改描述、封面
	// 等其它字段——那是在惩罚用户没做过的操作。
	if category.Slug != existing.Slug {
		if err := utils.ValidateSlug(category.Slug); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("%w：分类 URL slug %q 不合法，只能包含字母、数字和连字符", err, category.Slug)
		}
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
	// 走 RewritePostTags：读-改-写全程持 postRepo 写锁，与用户 SavePost 互斥、只改分类字段，
	// 杜绝级联抹掉用户刚保存的正文。
	return s.postRepo.RewritePostTags(ctx, func(p *domain.Post) bool {
		changed := false
		newCats := make([]string, len(p.Categories))
		for j, c := range p.Categories {
			if c == oldName {
				newCats[j] = newName
				changed = true
			} else {
				newCats[j] = c
			}
		}
		if changed {
			p.Categories = newCats
		}
		return changed
	})
}

func (s *CategoryService) cascadeCategoryDelete(ctx context.Context, categoryID, categoryName string) error {
	return s.postRepo.RewritePostTags(ctx, func(p *domain.Post) bool {
		changed := false
		newCats := make([]string, 0, len(p.Categories))
		for _, c := range p.Categories {
			if c != categoryName {
				newCats = append(newCats, c)
			} else {
				changed = true
			}
		}
		newCatIDs := make([]string, 0, len(p.CategoryIDs))
		for _, id := range p.CategoryIDs {
			if id != categoryID {
				newCatIDs = append(newCatIDs, id)
			} else {
				changed = true
			}
		}
		if changed {
			p.Categories = newCats
			p.CategoryIDs = newCatIDs
		}
		return changed
	})
}
