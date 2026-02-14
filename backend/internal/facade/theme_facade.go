package facade

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
	"os"
)

// ThemeFacade wraps ThemeService
type ThemeFacade struct {
	internal *service.ThemeService
	renderer *RendererFacade
}

func NewThemeFacade(s *service.ThemeService) *ThemeFacade {
	return &ThemeFacade{internal: s}
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

// SaveThemeConfigFromFrontend saves theme config and triggers render
func (f *ThemeFacade) SaveThemeConfigFromFrontend(config domain.ThemeConfig) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	if err := f.internal.SaveThemeConfig(ctx, config); err != nil {
		return err
	}

	// Trigger render
	if f.renderer != nil {
		go func() {
			if err := f.renderer.RenderAll(); err != nil {
				fmt.Fprintf(os.Stderr, "Error rendering after theme save: %v\n", err)
			}
		}()
	}

	return nil
}

// SaveThemeCustomConfigFromFrontend saves custom config and triggers render
func (f *ThemeFacade) SaveThemeCustomConfigFromFrontend(customConfig map[string]interface{}) error {
	// 1. Load current config
	currentConfig, err := f.LoadThemeConfig()
	if err != nil {
		return err
	}

	// 2. Update CustomConfig
	currentConfig.CustomConfig = customConfig

	// 3. Save config
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	if err := f.internal.SaveThemeConfig(ctx, currentConfig); err != nil {
		return err
	}

	// 4. Trigger render
	if f.renderer != nil {
		go func() {
			if err := f.renderer.RenderAll(); err != nil {
				fmt.Fprintf(os.Stderr, "Error rendering after theme custom config save: %v\n", err)
			}
		}()
	}

	return nil
}
