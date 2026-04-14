package facade

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

// CategoryFacade wraps CategoryService
type CategoryFacade struct {
	internal *service.CategoryService
	postRepo domain.PostRepository
}

func NewCategoryFacade(s *service.CategoryService, postRepo domain.PostRepository) *CategoryFacade {
	return &CategoryFacade{internal: s, postRepo: postRepo}
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

// CategoryCascadeResult 分类操作返回结果（含更新后的文章列表）
type CategoryCascadeResult struct {
	Categories []domain.Category `json:"categories"`
	Posts      []domain.Post     `json:"posts"`
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
func (f *CategoryFacade) SaveCategoryFromFrontend(form CategoryForm) (*CategoryCascadeResult, error) {
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

	if err := f.internal.SaveCategory(ctx, newCategory, form.ID); err != nil {
		return nil, err
	}

	categories, err := f.internal.LoadCategories(ctx)
	if err != nil {
		return nil, err
	}

	posts, _, err := f.postRepo.List(ctx, 1, 10000)
	if err != nil {
		return nil, err
	}

	return &CategoryCascadeResult{Categories: categories, Posts: posts}, nil
}

// DeleteCategoryFromFrontend 按 ID 删除分类，返回更新后的列表
func (f *CategoryFacade) DeleteCategoryFromFrontend(id string) (*CategoryCascadeResult, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}

	if err := f.internal.DeleteCategory(ctx, id); err != nil {
		return nil, err
	}

	categories, err := f.internal.LoadCategories(ctx)
	if err != nil {
		return nil, err
	}

	posts, _, err := f.postRepo.List(ctx, 1, 10000)
	if err != nil {
		return nil, err
	}

	return &CategoryCascadeResult{Categories: categories, Posts: posts}, nil
}

// RegisterEvents 注册分类相关事件监听器
func (f *CategoryFacade) RegisterEvents(ctx context.Context) {
	// Events match logic removed.
	// Frontend should call SaveCategories for sorting.
}
