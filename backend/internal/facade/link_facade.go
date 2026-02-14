package facade

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

// LinkFacade wraps LinkService
type LinkFacade struct {
	internal *service.LinkService
}

func NewLinkFacade(s *service.LinkService) *LinkFacade {
	return &LinkFacade{internal: s}
}

func (f *LinkFacade) LoadLinks() ([]domain.Link, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	// No side effects here anymore. Migration/Fixing should be explicit or at startup.
	return f.internal.LoadLinks(ctx)
}

// SaveLinks wraps service SaveLinks
func (f *LinkFacade) SaveLinks(links []domain.Link) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SaveLinks(ctx, links)
}

// LinkForm for frontend usage
type LinkForm struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
}

// SaveLinkFromFrontend accepts a LinkForm directly from frontend
func (f *LinkFacade) SaveLinkFromFrontend(form LinkForm) ([]domain.Link, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	newLink := domain.Link{
		ID:          form.ID,
		Name:        form.Name,
		Url:         form.URL,
		Avatar:      form.Avatar,
		Description: form.Description,
	}

	if err := f.internal.SaveLink(ctx, newLink); err != nil {
		return nil, err
	}

	return f.internal.LoadLinks(ctx)
}

// DeleteLinkFromFrontend accepts a link ID and returns updated list
func (f *LinkFacade) DeleteLinkFromFrontend(id string) ([]domain.Link, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	if err := f.internal.DeleteLink(ctx, id); err != nil {
		return nil, err
	}
	return f.internal.LoadLinks(ctx)
}

// RegisterEvents 注册友链相关事件监听器
func (f *LinkFacade) RegisterEvents(ctx context.Context) {
	// No events needed via Wails runtime.EventsOn for basic CRUD.
}
