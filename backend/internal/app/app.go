package app

import (
	"context"
	"gridea-pro/backend/internal/config"
	"gridea-pro/backend/internal/facade"
	"gridea-pro/backend/internal/service"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx             context.Context
	appDir          string
	buildDir        string
	previewService  *facade.PreviewFacade
	services        *facade.AppServices
	resourceWatcher *service.ResourceWatcher
}

func NewApp(appDir string, services *facade.AppServices) *App {
	return &App{
		appDir:   appDir,
		services: services,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// 0. Ensure site is initialized (scaffold)
	if err := a.services.Services.Scaffold.InitSite(a.appDir); err != nil {
		a.ShowToast("初始化站点失败: "+err.Error(), "error")
		// Log error but try to continue?
	}

	// buildDir 使用站点目录下的 output 文件夹（与原版 Gridea 一致）
	a.buildDir = filepath.Join(a.appDir, "output")

	// 初始化预览服务
	a.previewService = facade.NewPreviewFacade(a.buildDir)
	a.previewService.SetContext(ctx)

	// Initialize and start ResourceWatcher
	var err error
	a.resourceWatcher, err = service.NewResourceWatcher(a.appDir)
	if err == nil {
		a.resourceWatcher.Start(ctx)
	} else {
		// Log error
	}

	// Initialize Services Context and Events
	a.services.RegisterEvents(ctx)

	// App-specific events
	runtime.EventsOn(ctx, "app-ready", func(optionalData ...interface{}) {
		data := a.LoadSite()
		runtime.EventsEmit(ctx, "app-site-loaded", data)
	})

	// 监听站点重新加载事件（由前端保存主题等操作触发）
	runtime.EventsOn(ctx, "app-site-reload", func(_ ...interface{}) {
		// 重新加载站点数据
		data := a.LoadSite()
		// 发送给前端更新 Store
		runtime.EventsEmit(ctx, "app-site-loaded", data)
	})

	runtime.EventsOn(ctx, "preview-site", func(_ ...interface{}) {
		// 预览前先执行本地渲染（生成最新的静态文件）
		if err := a.services.Renderer.RenderAll(); err != nil {
			a.ShowToast("渲染失败："+err.Error(), "error")
			return
		}

		url, err := a.previewService.StartPreviewServer()
		if err != nil {
			a.ShowToast("预览服务启动失败："+err.Error(), "error")
			return
		}
		runtime.BrowserOpenURL(ctx, url)
	})

	runtime.EventsOn(ctx, "open-external", func(args ...interface{}) {
		if len(args) > 0 {
			if u, ok := args[0].(string); ok {
				runtime.BrowserOpenURL(ctx, u)
			}
		}
	})

	runtime.EventsOn(ctx, "show-preferences", func(_ ...interface{}) {
		// 转发事件到前端显示设置对话框
		runtime.EventsEmit(ctx, "show-preferences-dialog")
	})

	// 预启动预览服务
	_, _ = a.previewService.StartPreviewServer()

	// 监听源文件夹设置更改
	runtime.EventsOn(ctx, "app-source-folder-setting", func(args ...interface{}) {
		if len(args) == 0 {
			return
		}
		newPath, ok := args[0].(string)
		if !ok || newPath == "" {
			a.ShowToast("无效的路径", "error")
			return
		}

		// 验证路径是否存在
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			a.ShowToast("路径不存在", "error")
			return
		}

		// 0. Ensure new site is initialized
		// Note: We do this BEFORE saving config, so if it fails we might want to stop?
		// But usually we want to try to initialize.
		if err := a.services.Services.Scaffold.InitSite(newPath); err != nil {
			a.ShowToast("初始化站点失败: "+err.Error(), "error")
			runtime.EventsEmit(ctx, "app-source-folder-set", false)
			return
		}

		// 保存配置
		cm := config.NewConfigManager()
		if err := cm.UpdateSourceFolder(newPath); err != nil {
			a.ShowToast("保存配置失败: "+err.Error(), "error")
			runtime.EventsEmit(ctx, "app-source-folder-set", false)
			return
		}

		// --- 热更新逻辑 ---

		// 1. 更新 App 状态
		a.appDir = newPath
		a.buildDir = filepath.Join(newPath, "output")

		// 2. 更新所有业务 Service 的路径
		a.services.UpdateAppDir(newPath)

		// 3. 更新 PreviewService 的路径
		// 更新内部状态，并尝试重启服务如果它正在运行。
		a.previewService.SetBuildDir(a.buildDir)

		if a.previewService.IsRunning() {
			_ = a.previewService.StopPreviewServer()
			// 重新启动
			go func() {
				// 自动触发一次渲染，确保新目录有内容
				_ = a.services.Renderer.RenderAll()
				_, _ = a.previewService.StartPreviewServer()
			}()
		}

		// Update ResourceWatcher
		if a.resourceWatcher != nil {
			a.resourceWatcher.Close()
		}
		a.resourceWatcher, _ = service.NewResourceWatcher(newPath)
		if a.resourceWatcher != nil {
			a.resourceWatcher.Start(ctx)
		}

		// 4.重新加载站点数据
		siteData := a.LoadSite()

		// 5. 通知前端更新
		runtime.EventsEmit(ctx, "app-site-loaded", siteData)
		runtime.EventsEmit(ctx, "app-source-folder-set", true)

		a.ShowToast("源文件夹已更新", "success")
	})
}

func (a *App) Shutdown(ctx context.Context) {
	if a.previewService != nil {
		_ = a.previewService.StopPreviewServer()
	}
	if a.resourceWatcher != nil {
		a.resourceWatcher.Close()
	}
}

// OpenFolderDialog 映射前端 invoke('open-folder-dialog')
func (a *App) OpenFolderDialog() ([]string, error) {
	opts := runtime.OpenDialogOptions{
		Title: "选择站点源文件夹",
	}
	res, err := runtime.OpenDirectoryDialog(a.ctx, opts)
	if err != nil {
		return []string{}, err
	}
	if res == "" {
		return []string{}, nil
	}
	return []string{res}, nil
}

// OpenImageDialog 映射前端 invoke('open-image-dialog')
func (a *App) OpenImageDialog() (string, error) {
	opts := runtime.OpenDialogOptions{
		Title: "选择图片",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "图片文件 (*.jpg;*.png;*.gif;*.webp;*.ico;*.svg)",
				Pattern:     "*.jpg;*.jpeg;*.png;*.gif;*.webp;*.ico;*.svg;*.JPG;*.JPEG;*.PNG;*.GIF;*.WEBP;*.ICO;*.SVG",
			},
		},
	}
	res, err := runtime.OpenFileDialog(a.ctx, opts)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (a *App) LoadSite() map[string]interface{} {
	// 确保预览服务已启动
	if a.previewService != nil && !a.previewService.IsRunning() {
		_, _ = a.previewService.StartPreviewServer()
	}

	// Load data using services
	posts, _ := a.services.Post.LoadPosts()
	categories, _ := a.services.Category.LoadCategories()
	tags, _ := a.services.Post.LoadTags()

	menus, _ := a.services.Menu.LoadMenus()
	links, _ := a.services.Link.LoadLinks()
	themes, _ := a.services.Theme.LoadThemes()
	themeConfig, _ := a.services.Theme.LoadThemeConfig()

	// Load settings via service
	setting, _ := a.services.Setting.GetSetting()

	// Find current theme's custom config schema
	var currentThemeConfig []interface{}
	for _, t := range themes {
		if t.Folder == themeConfig.ThemeName {
			currentThemeConfig = t.CustomConfig
			break
		}
	}

	// Ensure maps are not nil
	if themeConfig.CustomConfig == nil {
		themeConfig.CustomConfig = make(map[string]interface{})
	}
	if currentThemeConfig == nil {
		currentThemeConfig = make([]interface{}, 0)
	}

	// Construct SiteData map
	return map[string]interface{}{
		"appDir":             a.appDir,
		"posts":              posts,
		"tags":               tags,
		"categories":         categories,
		"menus":              menus,
		"links":              links,
		"themes":             themes,
		"themeConfig":        themeConfig,
		"setting":            setting,
		"themeCustomConfig":  themeConfig.CustomConfig,
		"currentThemeConfig": currentThemeConfig,
	}
}

// GetPreviewURL 返回当前预览服务的 URL
// 如果服务未启动，会先启动服务然后返回 URL
func (a *App) GetPreviewURL() (string, error) {
	return a.previewService.StartPreviewServer()
}

// ShowToast 向前端发送 Toast 通知
// message: 显示的消息内容
// toastType: 类型 success, error, info, warning
func (a *App) ShowToast(message string, toastType string) {
	runtime.EventsEmit(a.ctx, "app:toast", map[string]interface{}{
		"message":  message,
		"type":     toastType,
		"duration": 3000, // 默认 3 秒
	})
}
