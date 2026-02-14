package facade

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

// TagFacade wraps TagService
type TagFacade struct {
	internal *service.TagService
}

func NewTagFacade(s *service.TagService) *TagFacade {
	return &TagFacade{internal: s}
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

func (f *TagFacade) GetTagColors() []string {
	return domain.TagColors
}

// TagForm for frontend usage
type TagForm struct {
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Color        string `json:"color"`
	OriginalName string `json:"originalName"`
}

// SaveTagFromFrontend accepts a TagForm directly from frontend
func (f *TagFacade) SaveTagFromFrontend(form TagForm) ([]domain.Tag, error) {
	newTag := domain.Tag{
		Name:  form.Name,
		Slug:  form.Slug,
		Color: form.Color,
	}

	// Use SaveTag to handle creation/update
	if err := f.SaveTag(newTag, form.OriginalName); err != nil {
		return nil, err
	}

	// Return updated list
	return f.LoadTags()
}

// DeleteTagFromFrontend accepts a tag name (or slug if logic changed) and returns updated list
func (f *TagFacade) DeleteTagFromFrontend(name string) ([]domain.Tag, error) {
	// First, if name is a slug, we might need to find the name, but current logic uses name for deletion
	// Use DeleteTag method
	if err := f.DeleteTag(name); err != nil {
		return nil, err
	}

	// Return updated list
	return f.LoadTags()
}

// RegisterEvents 注册标签相关事件监听器
func (f *TagFacade) RegisterEvents(ctx context.Context) {
	// Events are no longer used for Save/Delete
	// Keeping this empty or removing it entirely if no other events are needed.
	// We might still want Sort event if it was implemented, but it wasn't really.
}
