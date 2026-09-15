package facade

import (
	"context"
	"log/slog"
	"sync/atomic"

	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

// ThemeFacade wraps ThemeService
type ThemeFacade struct {
	// internal 会被切站(UpdateAppDir)热替换，用原子指针避免与 Wails 并发调用的 data race。
	internal atomic.Pointer[service.ThemeService]
	renderer *RendererFacade // 切站不替换（指向同一个 RendererFacade，其内部 engine 自身热替换）
	logger   *slog.Logger
}

func NewThemeFacade(s *service.ThemeService) *ThemeFacade {
	f := &ThemeFacade{logger: slog.Default()}
	f.internal.Store(s)
	return f
}

func (f *ThemeFacade) svc() *service.ThemeService { return f.internal.Load() }

func (f *ThemeFacade) SetRenderer(renderer *RendererFacade) {
	f.renderer = renderer
}

func (f *ThemeFacade) LoadThemes() ([]domain.Theme, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().LoadThemes(ctx)
}

func (f *ThemeFacade) LoadThemeConfig() (domain.ThemeConfig, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().LoadThemeConfig(ctx)
}

func (f *ThemeFacade) SaveThemeConfig(config domain.ThemeConfig) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().SaveThemeConfig(ctx, config)
}

func (f *ThemeFacade) UploadThemeCustomConfigImage(sourcePath string) (string, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().SaveThemeImage(ctx, sourcePath)
}

// SaveThemeConfigFromFrontend saves theme config.
// 不在这里直接触发渲染——前端会在保存成功后 emit app-site-reload，
// 由 RendererFacade 的事件监听器统一处理。避免重复渲染。
func (f *ThemeFacade) SaveThemeConfigFromFrontend(config domain.ThemeConfig) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().SaveThemeConfig(ctx, config)
}

// SaveThemeCustomConfigFromFrontend saves custom config.
// 同 SaveThemeConfigFromFrontend，不在这里直接触发渲染，由前端 emit 的
// app-site-reload 事件统一处理，避免重复渲染。
func (f *ThemeFacade) SaveThemeCustomConfigFromFrontend(customConfig map[string]interface{}) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	// 方法内缓存一次当前 service：本方法先读(LoadThemeConfig)后写(SaveThemeConfig)，
	// 缓存后整个方法用同一个 service，避免中途切站导致读写分属两个站点。
	svc := f.svc()

	currentConfig, err := svc.LoadThemeConfig(ctx)
	if err != nil {
		return err
	}

	currentConfig.CustomConfig = customConfig

	return svc.SaveThemeConfig(ctx, currentConfig)
}
