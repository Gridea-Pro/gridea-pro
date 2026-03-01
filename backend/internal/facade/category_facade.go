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

// CategoryForm 前端提交的分类表单
type CategoryForm struct {
	ID          string `json:"id"` // 分类 UUID（新建时为空，更新时必填）
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	// 已废弃：OriginalSlug 保留字段以防老版前端调用，逻辑忽略
	OriginalSlug string `json:"originalSlug"`
}

// SaveCategoryFromFrontend 创建或更新分类
// 若 form.ID 为空则创建新分类；否则按 ID 更新
func (f *CategoryFacade) SaveCategoryFromFrontend(form CategoryForm) ([]domain.Category, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}

	newCategory := domain.Category{
		ID:          form.ID,
		Name:        form.Name,
		Slug:        form.Slug,
		Description: form.Description,
	}

	// originalID 为空 → 新建；非空 → 更新（保持 ID 不变）
	if err := f.internal.SaveCategory(ctx, newCategory, form.ID); err != nil {
		return nil, err
	}

	return f.internal.LoadCategories(ctx)
}

// DeleteCategoryFromFrontend 按 ID 删除分类，返回更新后的列表
func (f *CategoryFacade) DeleteCategoryFromFrontend(id string) ([]domain.Category, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}

	if err := f.internal.DeleteCategory(ctx, id); err != nil {
		return nil, err
	}

	return f.internal.LoadCategories(ctx)
}

// RegisterEvents 注册分类相关事件监听器
func (f *CategoryFacade) RegisterEvents(ctx context.Context) {
	// Events match logic removed.
	// Frontend should call SaveCategories for sorting.
}
