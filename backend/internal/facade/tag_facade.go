package facade

import (
	"context"
	"sync/atomic"

	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

// TagFacade wraps TagService
type TagFacade struct {
	// internal / postRepo 用原子读写保存：切站（UpdateAppDir）会热替换它们，
	// 而 Wails 可能并发调用本 facade 的方法，原子读写避免"读到替换一半的状态"的 data race。
	internal atomic.Pointer[service.TagService]
	postRepo atomic.Value // 存 domain.PostRepository
}

func NewTagFacade(s *service.TagService, postRepo domain.PostRepository) *TagFacade {
	f := &TagFacade{}
	f.internal.Store(s)
	f.postRepo.Store(postRepo)
	return f
}

func (f *TagFacade) svc() *service.TagService     { return f.internal.Load() }
func (f *TagFacade) posts() domain.PostRepository { return f.postRepo.Load().(domain.PostRepository) }

func (f *TagFacade) LoadTags() ([]domain.Tag, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().LoadTags(ctx)
}

func (f *TagFacade) SaveTag(tag domain.Tag, originalName string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().SaveTag(ctx, tag, originalName)
}

func (f *TagFacade) DeleteTag(name string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().DeleteTag(ctx, name)
}

func (f *TagFacade) SaveTags(tags []domain.Tag) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().SaveTags(ctx, tags)
}

func (f *TagFacade) GetTagColors() []string {
	return service.TagColors
}

// TagForm for frontend usage
type TagForm struct {
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Color        string `json:"color"`
	OriginalName string `json:"originalName"`
}

// TagCascadeResult 标签操作返回结果（含更新后的文章列表）
type TagCascadeResult struct {
	Tags  []domain.Tag  `json:"tags"`
	Posts []domain.Post `json:"posts"`
}

// SaveTagFromFrontend accepts a TagForm directly from frontend
func (f *TagFacade) SaveTagFromFrontend(form TagForm) (*TagCascadeResult, error) {
	// 一次调用内只取一次 svc/posts 快照，避免中途切站把"保存到旧站、返回新站列表"劈成两半。
	svc := f.svc()
	posts := f.posts()

	newTag := domain.Tag{
		Name:  form.Name,
		Slug:  form.Slug,
		Color: form.Color,
	}

	if err := svc.SaveTag(ctx(), newTag, form.OriginalName); err != nil {
		return nil, err
	}

	tags, err := svc.LoadTags(ctx())
	if err != nil {
		return nil, err
	}

	allPosts, err := posts.GetAll(ctx())
	if err != nil {
		return nil, err
	}

	return &TagCascadeResult{Tags: tags, Posts: allPosts}, nil
}

// DeleteTagFromFrontend accepts a tag name and returns updated list
func (f *TagFacade) DeleteTagFromFrontend(name string) (*TagCascadeResult, error) {
	svc := f.svc()
	posts := f.posts()

	if err := svc.DeleteTag(ctx(), name); err != nil {
		return nil, err
	}

	tags, err := svc.LoadTags(ctx())
	if err != nil {
		return nil, err
	}

	allPosts, err := posts.GetAll(ctx())
	if err != nil {
		return nil, err
	}

	return &TagCascadeResult{Tags: tags, Posts: allPosts}, nil
}

func ctx() context.Context {
	if WailsContext != nil {
		return WailsContext
	}
	return context.TODO()
}

// RegisterEvents 注册标签相关事件监听器
func (f *TagFacade) RegisterEvents(ctx context.Context) {
	// Events are no longer used for Save/Delete
	// Keeping this empty or removing it entirely if no other events are needed.
	// We might still want Sort event if it was implemented, but it wasn't really.
}
