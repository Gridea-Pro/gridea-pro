package repository

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type categoryRepository struct {
	*BaseJSONRepository[domain.Category]
}

func NewCategoryRepository(appDir string) domain.CategoryRepository {
	base := NewBaseJSONRepository[domain.Category](appDir, "categories.json", "categories")
	return &categoryRepository{base}
}

// ensureID 若分类没有 ID（老数据），自动生成 NanoID
func ensureCategoryID(category *domain.Category) {
	if category.ID == "" {
		const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		id, _ := gonanoid.Generate(alphabet, 6)
		category.ID = id
	}
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	ensureCategoryID(category)
	return r.Add(ctx, *category)
}

func (r *categoryRepository) Update(ctx context.Context, id string, category *domain.Category) error {
	// 保持 ID 不变
	category.ID = id
	return r.BaseJSONRepository.Update(ctx, id, *category)
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	return r.BaseJSONRepository.Delete(ctx, id)
}

func (r *categoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	cat, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *categoryRepository) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	cats, err := r.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range cats {
		if cats[i].Slug == slug {
			return &cats[i], nil
		}
	}
	return nil, fmt.Errorf("category not found with slug: %s", slug)
}

func (r *categoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	return r.BaseJSONRepository.List(ctx)
}

// SaveAll 在保存时自动为没有 ID 的分类补充 UUID（老数据迁移）
func (r *categoryRepository) SaveAll(ctx context.Context, categories []domain.Category) error {
	for i := range categories {
		ensureCategoryID(&categories[i])
	}
	return r.BaseJSONRepository.SaveAll(ctx, categories)
}
