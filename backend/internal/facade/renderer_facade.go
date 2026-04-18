package facade

import (
	"context"
	"gridea-pro/backend/internal/engine"
	"log/slog"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// RendererFacade wraps RendererService
type RendererFacade struct {
	internal *engine.Engine
	logger   *slog.Logger
}

func NewRendererFacade(s *engine.Engine) *RendererFacade {
	return &RendererFacade{internal: s, logger: slog.Default()}
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

// registerSiteReloadEvent 注册站点重新加载事件监听器。
// 实际的并发保护（串行化 + 合并）在 engine.RenderAll 内部，这里只负责转发事件。
func registerSiteReloadEvent(ctx context.Context, rendererFacade *RendererFacade) {
	runtime.EventsOn(ctx, "app-site-reload", func(data ...interface{}) {
		rendererFacade.logger.Info("收到 app-site-reload 事件，触发渲染")
		go func() {
			if err := rendererFacade.RenderAll(); err != nil {
				rendererFacade.logger.Error("站点重新加载失败", "error", err)
			} else {
				rendererFacade.logger.Info("站点重新加载成功")
			}
		}()
	})
}
