package facade

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/service"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// RendererFacade wraps RendererService
type RendererFacade struct {
	internal *service.RendererService
}

func NewRendererFacade(s *service.RendererService) *RendererFacade {
	return &RendererFacade{internal: s}
}

func (f *RendererFacade) RenderAll() error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.RenderAll(ctx)
}

// RegisterEvents 注册渲染相关事件监听器
func (f *RendererFacade) RegisterEvents(ctx context.Context) {
	registerSiteReloadEvent(ctx, f)
}

// registerSiteReloadEvent 注册站点重新加载事件监听器
func registerSiteReloadEvent(ctx context.Context, rendererFacade *RendererFacade) {
	runtime.EventsOn(ctx, "app-site-reload", func(data ...interface{}) {
		// 触发重新渲染
		go func() {
			if err := rendererFacade.RenderAll(); err != nil {
				fmt.Fprintf(os.Stderr, "站点重新加载失败: %v\n", err)
			} else {
				fmt.Fprintln(os.Stderr, "站点重新加载成功")
			}
		}()
	})
}
