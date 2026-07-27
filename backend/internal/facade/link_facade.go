package facade

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
	"sync/atomic"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// LinkFacade wraps LinkService
type LinkFacade struct {
	// internal 用原子读写保存：切站（UpdateAppDir）会热替换它，
	// 而 Wails 可能并发调用本 facade 的方法，原子读写避免"读到替换一半的状态"的 data race。
	internal atomic.Pointer[service.LinkService]
}

func NewLinkFacade(s *service.LinkService) *LinkFacade {
	f := &LinkFacade{}
	f.internal.Store(s)
	return f
}

func (f *LinkFacade) svc() *service.LinkService { return f.internal.Load() }

func (f *LinkFacade) LoadLinks() ([]domain.Link, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	// No side effects here anymore. Migration/Fixing should be explicit or at startup.
	return f.svc().LoadLinks(ctx)
}

// SaveLinks wraps service SaveLinks
func (f *LinkFacade) SaveLinks(links []domain.Link) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().SaveLinks(ctx, links)
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
	// 一次调用内只取一次 svc 快照，避免"写旧站、读新站列表"被切站劈成两半。
	svc := f.svc()

	newLink := domain.Link{
		ID:          form.ID,
		Name:        form.Name,
		Url:         form.URL,
		Avatar:      form.Avatar,
		Description: form.Description,
	}

	if newLink.ID == "" {
		if err := svc.CreateLink(ctx, newLink); err != nil {
			runtime.LogError(ctx, fmt.Sprintf("Failed to create link: %v", err))
			return nil, err
		}
	} else {
		if err := svc.UpdateLink(ctx, newLink); err != nil {
			return nil, err
		}
	}

	return svc.LoadLinks(ctx)
}

// DeleteLinkFromFrontend accepts a link ID and returns updated list
func (f *LinkFacade) DeleteLinkFromFrontend(id string) ([]domain.Link, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	svc := f.svc()
	if err := svc.DeleteLink(ctx, id); err != nil {
		return nil, err
	}
	return svc.LoadLinks(ctx)
}

// RegisterEvents 注册友链相关事件监听器
func (f *LinkFacade) RegisterEvents(ctx context.Context) {
	// No events needed via Wails runtime.EventsOn for basic CRUD.
}
