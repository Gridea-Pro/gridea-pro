package facade

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

// TagFacade wraps TagService
type TagFacade struct {
	internal *service.TagService
	postRepo domain.PostRepository
}

func NewTagFacade(s *service.TagService, postRepo domain.PostRepository) *TagFacade {
	return &TagFacade{internal: s, postRepo: postRepo}
}

func (f *TagFacade) LoadTags() ([]domain.Tag, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.LoadTags(ctx)
}

func (f *TagFacade) SaveTag(tag domain.Tag, originalName string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SaveTag(ctx, tag, originalName)
}

func (f *TagFacade) DeleteTag(name string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.DeleteTag(ctx, name)
}

func (f *TagFacade) SaveTags(tags []domain.Tag) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SaveTags(ctx, tags)
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
	newTag := domain.Tag{
		Name:  form.Name,
		Slug:  form.Slug,
		Color: form.Color,
	}

	if err := f.SaveTag(newTag, form.OriginalName); err != nil {
		return nil, err
	}

	tags, err := f.LoadTags()
	if err != nil {
		return nil, err
	}

	posts, _, err := f.postRepo.List(ctx(), 1, 10000)
	if err != nil {
		return nil, err
	}

	return &TagCascadeResult{Tags: tags, Posts: posts}, nil
}

// DeleteTagFromFrontend accepts a tag name and returns updated list
func (f *TagFacade) DeleteTagFromFrontend(name string) (*TagCascadeResult, error) {
	if err := f.DeleteTag(name); err != nil {
		return nil, err
	}

	tags, err := f.LoadTags()
	if err != nil {
		return nil, err
	}

	posts, _, err := f.postRepo.List(ctx(), 1, 10000)
	if err != nil {
		return nil, err
	}

	return &TagCascadeResult{Tags: tags, Posts: posts}, nil
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
