package app

import (
	"context"
	"gridea-pro/backend/internal/config"
	"gridea-pro/backend/internal/facade"
	"gridea-pro/backend/internal/service"
	"os"
	"path/filepath"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	EventAppReady               = "app-ready"
	EventAppSiteLoaded          = "app-site-loaded"
	EventAppSiteReload          = "app-site-reload"
	EventPreviewSite            = "preview-site"
	EventOpenExternal           = "open-external"
	EventShowPreferences        = "show-preferences"
	EventShowPreferencesDialog  = "show-preferences-dialog"
	EventAppSourceFolderSetting = "app-source-folder-setting"
	EventAppSourceFolderSet     = "app-source-folder-set"
	EventAppToast               = "app:toast" // Keep consistent with frontend if needed, or change to "app-toast"
)

type App struct {
	ctx             context.Context
	mu              sync.RWMutex
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
	// InitSite uses appDir, which is set in NewApp and not changed until handleSourceFolderChange.
	// Since Startup is single-threaded at this point, read is safe.
	if err := a.services.Services.Scaffold.InitSite(a.appDir); err != nil {
		a.ShowToast("初始化站点失败: "+err.Error(), "error")
		runtime.LogError(ctx, "Failed to init site: "+err.Error())
	}

	// 初始化预览服务
	a.previewService = a.services.Preview

	// Initialize and start ResourceWatcher
	var err error
	a.resourceWatcher, err = service.NewResourceWatcher(a.appDir)
	if err == nil {
		a.resourceWatcher.Start(ctx)
	} else {
		runtime.LogError(ctx, "Failed to start resource watcher: "+err.Error())
	}

	// Initialize Services Context and Events
	a.services.RegisterEvents(ctx)

	// Register App Events
	a.registerEvents(ctx)

	// 预启动预览服务
	if _, err := a.previewService.StartPreviewServer(); err != nil {
		runtime.LogError(ctx, "Failed to pre-start preview server: "+err.Error())
	}
}

func (a *App) registerEvents(ctx context.Context) {
	// App-specific events
	runtime.EventsOn(ctx, EventAppReady, a.handleSiteReload)
	runtime.EventsOn(ctx, EventAppSiteReload, a.handleSiteReload)
	runtime.EventsOn(ctx, EventPreviewSite, a.handlePreviewSite)

	runtime.EventsOn(ctx, EventOpenExternal, func(args ...interface{}) {
		if len(args) > 0 {
			if u, ok := args[0].(string); ok {
				runtime.BrowserOpenURL(ctx, u)
			}
		}
	})

	runtime.EventsOn(ctx, EventShowPreferences, func(_ ...interface{}) {
		// 转发事件到前端显示设置对话框
		a.ShowPreferences()
	})

	// 监听源文件夹设置更改
	runtime.EventsOn(ctx, EventAppSourceFolderSetting, func(args ...interface{}) {
		if len(args) == 0 {
			return
		}
		newPath, ok := args[0].(string)
		if !ok || newPath == "" {
			a.ShowToast("无效的路径", "error")
			return
		}
		a.handleSourceFolderChange(newPath)
	})
}

func (a *App) handleSourceFolderChange(newPath string) {
	// 验证路径是否存在
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		a.ShowToast("路径不存在", "error")
		return
	}

	// 0. Ensure new site is initialized
	if err := a.services.Services.Scaffold.InitSite(newPath); err != nil {
		a.ShowToast("初始化站点失败: "+err.Error(), "error")
		runtime.EventsEmit(a.ctx, EventAppSourceFolderSet, false)
		return
	}

	// 保存配置
	cm, err := config.NewConfigManager()
	if err != nil {
		a.ShowToast("初始化配置管理器失败: "+err.Error(), "error")
		runtime.EventsEmit(a.ctx, EventAppSourceFolderSet, false)
		return
	}
	if err := cm.UpdateSourceFolder(newPath); err != nil {
		a.ShowToast("保存配置失败: "+err.Error(), "error")
		runtime.EventsEmit(a.ctx, EventAppSourceFolderSet, false)
		return
	}

	// --- 热更新逻辑 ---

	newBuildDir := filepath.Join(newPath, "output")

	a.mu.Lock()
	// 1. 更新 App 状态
	a.appDir = newPath
	a.buildDir = newBuildDir
	a.mu.Unlock()

	// 3. 更新 PreviewService 的路径
	// 先停止服务，因为 UpdateAppDir 会替换 PreviewService 的内部实例
	shouldRestart := false
	if a.previewService.IsRunning() {
		_ = a.previewService.StopPreviewServer()
		shouldRestart = true
	}

	// 2. 更新所有业务 Service 的路径 (这会更新 PreviewFacade.internal)
	a.services.UpdateAppDir(newPath)

	if shouldRestart {
		// 重新启动
		go func() {
			// 自动触发一次渲染，确保新目录有内容
			if err := a.services.Renderer.RenderAll(); err != nil {
				runtime.LogError(a.ctx, "Auto render failed after source change: "+err.Error())
			}
			if _, err := a.previewService.StartPreviewServer(); err != nil {
				runtime.LogError(a.ctx, "Failed to restart preview server: "+err.Error())
			}
		}()
	}

	// Update ResourceWatcher
	if a.resourceWatcher != nil {
		a.resourceWatcher.Close()
	}
	a.resourceWatcher, err = service.NewResourceWatcher(newPath)
	if err != nil {
		runtime.LogError(a.ctx, "Failed to create resource watcher for new path: "+err.Error())
	} else if a.resourceWatcher != nil {
		a.resourceWatcher.Start(a.ctx)
	}

	// 4.重新加载站点数据
	siteData := a.LoadSite()

	// 5. 通知前端更新
	runtime.EventsEmit(a.ctx, EventAppSiteLoaded, siteData)
	runtime.EventsEmit(a.ctx, EventAppSourceFolderSet, true)

	a.ShowToast("源文件夹已更新", "success")
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

	a.mu.RLock()
	appDir := a.appDir
	a.mu.RUnlock()

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
		"appDir":             appDir,
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

// ShowPreferences 显示设置窗口
// 供 Go 代码（如菜单）直接调用
func (a *App) ShowPreferences() {
	if a.ctx == nil {
		return
	}
	runtime.WindowCenter(a.ctx)
	runtime.EventsEmit(a.ctx, EventShowPreferencesDialog)
}

// ShowToast 向前端发送 Toast 通知
// message: 显示的消息内容
// toastType: 类型 success, error, info, warning
func (a *App) ShowToast(message string, toastType string) {
	runtime.EventsEmit(a.ctx, EventAppToast, map[string]interface{}{
		"message":  message,
		"type":     toastType,
		"duration": 3000, // 默认 3 秒
	})
}

func (a *App) handleSiteReload(_ ...interface{}) {
	// 重新加载站点数据
	data := a.LoadSite()
	// 发送给前端更新 Store
	runtime.EventsEmit(a.ctx, EventAppSiteLoaded, data)
}

func (a *App) handlePreviewSite(_ ...interface{}) {
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
	runtime.BrowserOpenURL(a.ctx, url)
}
