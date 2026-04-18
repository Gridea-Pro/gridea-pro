package facade

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
	"log/slog"
)

// ThemeFacade wraps ThemeService
type ThemeFacade struct {
	internal *service.ThemeService
	renderer *RendererFacade
	logger   *slog.Logger
}

func NewThemeFacade(s *service.ThemeService) *ThemeFacade {
	return &ThemeFacade{internal: s, logger: slog.Default()}
}

func (f *ThemeFacade) SetRenderer(renderer *RendererFacade) {
	f.renderer = renderer
}

func (f *ThemeFacade) LoadThemes() ([]domain.Theme, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.LoadThemes(ctx)
}

func (f *ThemeFacade) LoadThemeConfig() (domain.ThemeConfig, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.LoadThemeConfig(ctx)
}

func (f *ThemeFacade) SaveThemeConfig(config domain.ThemeConfig) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SaveThemeConfig(ctx, config)
}

func (f *ThemeFacade) UploadThemeCustomConfigImage(sourcePath string) (string, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SaveThemeImage(ctx, sourcePath)
}

// SaveThemeConfigFromFrontend saves theme config.
// 不在这里直接触发渲染——前端会在保存成功后 emit app-site-reload，
// 由 RendererFacade 的事件监听器统一处理。避免重复渲染。
func (f *ThemeFacade) SaveThemeConfigFromFrontend(config domain.ThemeConfig) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SaveThemeConfig(ctx, config)
}

// SaveThemeCustomConfigFromFrontend saves custom config.
// 同 SaveThemeConfigFromFrontend，不在这里直接触发渲染，由前端 emit 的
// app-site-reload 事件统一处理，避免重复渲染。
func (f *ThemeFacade) SaveThemeCustomConfigFromFrontend(customConfig map[string]interface{}) error {
	currentConfig, err := f.LoadThemeConfig()
	if err != nil {
		return err
	}

	currentConfig.CustomConfig = customConfig

	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SaveThemeConfig(ctx, currentConfig)
}
