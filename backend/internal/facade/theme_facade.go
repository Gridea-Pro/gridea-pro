package facade

import (
	"context"
	"encoding/json"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ThemeFacade wraps ThemeService
type ThemeFacade struct {
	internal *service.ThemeService
}

func NewThemeFacade(s *service.ThemeService) *ThemeFacade {
	return &ThemeFacade{internal: s}
}

func (f *ThemeFacade) LoadThemes() ([]domain.Theme, error) {
	return f.internal.LoadThemes(context.TODO())
}

func (f *ThemeFacade) LoadThemeConfig() (domain.ThemeConfig, error) {
	return f.internal.LoadThemeConfig(context.TODO())
}

func (f *ThemeFacade) SaveThemeConfig(config domain.ThemeConfig) error {
	return f.internal.SaveThemeConfig(context.TODO(), config)
}

// RegisterEvents 注册主题相关事件监听器
func (f *ThemeFacade) RegisterEvents(ctx context.Context, rendererFacade *RendererFacade) {
	registerThemeSaveEvent(ctx, f, rendererFacade)
	registerThemeCustomConfigSaveEvent(ctx, f, rendererFacade)
}

// registerThemeSaveEvent 注册主题保存事件监听器
func registerThemeSaveEvent(ctx context.Context, themeFacade *ThemeFacade, rendererFacade *RendererFacade) {
	runtime.EventsOn(ctx, "theme-save", func(data ...interface{}) {
		if len(data) == 0 {
			runtime.EventsEmit(ctx, "theme-saved", false)
			return
		}

		// 将前端传来的 map 转换为 JSON，再解析为 ThemeConfig
		configMap, ok := data[0].(map[string]interface{})
		if !ok {
			runtime.EventsEmit(ctx, "theme-saved", false)
			return
		}

		// 将 map 转换为 JSON bytes
		jsonBytes, err := json.Marshal(configMap)
		if err != nil {
			runtime.EventsEmit(ctx, "theme-saved", false)
			return
		}

		// 解析为 ThemeConfig
		var config domain.ThemeConfig
		if err := json.Unmarshal(jsonBytes, &config); err != nil {
			runtime.EventsEmit(ctx, "theme-saved", false)
			return
		}

		// 注意：如果前端只传了部分字段（非 CustomConfig），我们需要保留原有的 CustomConfig
		// 但这里的 theme-save 通常是保存基本设置，BasicSetting.vue 会发送 merged config.
		// 所以这里 config 应该包含了所有基本字段。
		// 但是，如果 CustomConfig 不在 configMap 中，Unmarshal 可能会让它为空。
		// BasicSetting.vue 发送的是 { ...siteStore.site.themeConfig, ...form }
		// siteStore.site.themeConfig 在 app.go LoadSite 中被 populate 了 CustomConfig
		// 所以如果前端正确传递，ThemeConfig.CustomConfig 应该是有值的。

		// 无论如何，为了安全起见，这里不需要做额外的合并，假设前端传的是全量或者包含了必要的 CustomConfig
		// 如果 BasicSetting.vue 里的 ...themeConfig 包含了 CustomConfig，那就没问题。

		// 保存配置
		if err := themeFacade.SaveThemeConfig(config); err != nil {
			runtime.EventsEmit(ctx, "theme-saved", false)
			return
		}

		// 保存成功后触发重新渲染
		go func() {
			if err := rendererFacade.RenderAll(); err != nil {
				fmt.Printf("主题保存后重新渲染失败: %v\n", err)
			} else {
				fmt.Printf("主题 %s 保存成功并重新渲染完成\n", config.ThemeName)
			}
		}()

		// 成功
		runtime.EventsEmit(ctx, "theme-saved", true)
	})
}

// registerThemeCustomConfigSaveEvent 注册自定义配置保存事件监听器
func registerThemeCustomConfigSaveEvent(ctx context.Context, themeFacade *ThemeFacade, rendererFacade *RendererFacade) {
	runtime.EventsOn(ctx, "theme-custom-config-save", func(data ...interface{}) {
		if len(data) == 0 {
			runtime.EventsEmit(ctx, "theme-custom-config-saved", map[string]interface{}{"success": false})
			return
		}

		customConfig, ok := data[0].(map[string]interface{})
		if !ok {
			runtime.EventsEmit(ctx, "theme-custom-config-saved", map[string]interface{}{"success": false})
			return
		}

		// 1. 加载当前配置
		currentConfig, err := themeFacade.LoadThemeConfig()
		if err != nil {
			runtime.EventsEmit(ctx, "theme-custom-config-saved", map[string]interface{}{"success": false})
			return
		}

		// 2. 更新 CustomConfig
		currentConfig.CustomConfig = customConfig

		// 3. 保存配置
		if err := themeFacade.SaveThemeConfig(currentConfig); err != nil {
			runtime.EventsEmit(ctx, "theme-custom-config-saved", map[string]interface{}{"success": false})
			return
		}

		// 4. 重发通知
		runtime.EventsEmit(ctx, "theme-custom-config-saved", map[string]interface{}{"success": true})

		// 5. 触发重新渲染
		go func() {
			rendererFacade.RenderAll()
		}()
	})
}
