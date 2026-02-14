package facade

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

// CategoryFacade wraps CategoryService
type CategoryFacade struct {
	internal *service.CategoryService
}

func NewCategoryFacade(s *service.CategoryService) *CategoryFacade {
	return &CategoryFacade{internal: s}
}

func (f *CategoryFacade) LoadCategories() ([]domain.Category, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.LoadCategories(ctx)
}

func (f *CategoryFacade) SaveCategories(categories []domain.Category) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SaveCategories(ctx, categories)
}

func (f *CategoryFacade) SaveCategory(category domain.Category, originalSlug string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SaveCategory(ctx, category, originalSlug)
}

func (f *CategoryFacade) DeleteCategory(slug string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.DeleteCategory(ctx, slug)
}

// CategoryForm for frontend usage
type CategoryForm struct {
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Description  string `json:"description"`
	OriginalSlug string `json:"originalSlug"`
}

// SaveCategoryFromFrontend accepts a CategoryForm directly from frontend
func (f *CategoryFacade) SaveCategoryFromFrontend(form CategoryForm) ([]domain.Category, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}

	newCategory := domain.Category{
		Name:        form.Name,
		Slug:        form.Slug,
		Description: form.Description,
	}

	// 调用 Service 方法
	if err := f.internal.SaveCategory(ctx, newCategory, form.OriginalSlug); err != nil {
		return nil, err
	}

	// 重新加载分类列表
	return f.internal.LoadCategories(ctx)
}

// DeleteCategoryFromFrontend accepts a slug and returns updated list
func (f *CategoryFacade) DeleteCategoryFromFrontend(slug string) ([]domain.Category, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}

	if err := f.internal.DeleteCategory(ctx, slug); err != nil {
		return nil, err
	}

	return f.internal.LoadCategories(ctx)
}

// RegisterEvents 注册分类相关事件监听器
func (f *CategoryFacade) RegisterEvents(ctx context.Context) {
	// Events match logic removed.
	// Frontend should call SaveCategories for sorting.
}
