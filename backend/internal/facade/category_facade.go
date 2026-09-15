package facade

import (
	"context"
	"sync/atomic"

	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

// CategoryFacade wraps CategoryService
type CategoryFacade struct {
	// internal / postRepo 用原子读写保存：切站（UpdateAppDir）会热替换它们，
	// 而 Wails 可能并发调用本 facade 的方法，原子读写避免"读到替换一半的状态"的 data race。
	internal atomic.Pointer[service.CategoryService]
	postRepo atomic.Value // 存 domain.PostRepository
}

func NewCategoryFacade(s *service.CategoryService, postRepo domain.PostRepository) *CategoryFacade {
	f := &CategoryFacade{}
	f.internal.Store(s)
	f.postRepo.Store(postRepo)
	return f
}

func (f *CategoryFacade) svc() *service.CategoryService { return f.internal.Load() }
func (f *CategoryFacade) posts() domain.PostRepository {
	return f.postRepo.Load().(domain.PostRepository)
}

func (f *CategoryFacade) LoadCategories() ([]domain.Category, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().LoadCategories(ctx)
}

func (f *CategoryFacade) SaveCategories(categories []domain.Category) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().SaveCategories(ctx, categories)
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

	// 一次调用内只取一次 svc/posts 快照：避免中途发生切站（UpdateAppDir）导致
	// "保存到旧站分类、却返回新站文章列表"这类被劈成两半的数据串号。
	svc := f.svc()
	posts := f.posts()

	newCategory := domain.Category{
		ID:          form.ID,
		Name:        form.Name,
		Slug:        form.Slug,
		Description: form.Description,
	}

	if err := svc.SaveCategory(ctx, newCategory, form.ID); err != nil {
		return nil, err
	}

	categories, err := svc.LoadCategories(ctx)
	if err != nil {
		return nil, err
	}

	allPosts, err := posts.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return &CategoryCascadeResult{Categories: categories, Posts: allPosts}, nil
}

// DeleteCategoryFromFrontend 按 ID 删除分类，返回更新后的列表
func (f *CategoryFacade) DeleteCategoryFromFrontend(id string) (*CategoryCascadeResult, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}

	svc := f.svc()
	posts := f.posts()

	if err := svc.DeleteCategory(ctx, id); err != nil {
		return nil, err
	}

	categories, err := svc.LoadCategories(ctx)
	if err != nil {
		return nil, err
	}

	allPosts, err := posts.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return &CategoryCascadeResult{Categories: categories, Posts: allPosts}, nil
}

// RegisterEvents 注册分类相关事件监听器
func (f *CategoryFacade) RegisterEvents(ctx context.Context) {
	// Events match logic removed.
	// Frontend should call SaveCategories for sorting.
}
