package app

import (
	"context"
	"fmt"
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
	version         string
	previewService  *facade.PreviewFacade
	services        *facade.AppServices
	resourceWatcher *service.ResourceWatcher
}

func NewApp(appDir string, services *facade.AppServices, version string) *App {
	return &App{
		appDir:   appDir,
		services: services,
		version:  version,
	}
}

// GetVersion 返回应用版本号，供前端调用
func (a *App) GetVersion() string {
	return a.version
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

	// === 执行基础数据格式迁移与清洗 (ID 统一化) ===
	migrator := service.NewDataMigrator(
		a.appDir,
		a.services.Repositories.Category,
		a.services.Repositories.Tag,
		a.services.Repositories.Post,
		a.services.Repositories.Menu,
		a.services.Repositories.Link,
		a.services.Repositories.Memo,
	)
	if err := migrator.RunMigration(ctx); err != nil {
		runtime.LogError(ctx, "Data migration failed: "+err.Error())
	}

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
	if err := a.switchToPath(newPath); err != nil {
		a.ShowToast(err.Error(), "error")
		runtime.EventsEmit(a.ctx, EventAppSourceFolderSet, false)
		return
	}
	runtime.EventsEmit(a.ctx, EventAppSourceFolderSet, true)
	a.ShowToast("源文件夹已更新", "success")
}

// switchToPath 切换到指定路径的站点（核心热更新逻辑）
func (a *App) switchToPath(newPath string) error {
	// 验证路径是否存在
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		return fmt.Errorf("路径不存在: %s", newPath)
	}

	// 初始化站点目录
	if err := a.services.Services.Scaffold.InitSite(newPath); err != nil {
		return fmt.Errorf("初始化站点失败: %w", err)
	}

	// 热更新 App 状态
	a.mu.Lock()
	a.appDir = newPath
	a.buildDir = filepath.Join(newPath, "output")
	a.mu.Unlock()

	// 更新 PreviewService
	shouldRestart := false
	if a.previewService != nil && a.previewService.IsRunning() {
		_ = a.previewService.StopPreviewServer()
		shouldRestart = true
	}

	// 更新所有业务 Service
	a.services.UpdateAppDir(newPath)

	if shouldRestart {
		go func() {
			if err := a.services.Renderer.RenderAll(); err != nil {
				runtime.LogError(a.ctx, "Auto render failed after source change: "+err.Error())
			}
			if _, err := a.previewService.StartPreviewServer(); err != nil {
				runtime.LogError(a.ctx, "Failed to restart preview server: "+err.Error())
			}
		}()
	}

	// 更新 ResourceWatcher
	if a.resourceWatcher != nil {
		a.resourceWatcher.Close()
	}
	var watchErr error
	a.resourceWatcher, watchErr = service.NewResourceWatcher(newPath)
	if watchErr != nil {
		runtime.LogError(a.ctx, "Failed to create resource watcher: "+watchErr.Error())
	} else if a.resourceWatcher != nil {
		a.resourceWatcher.Start(a.ctx)
	}

	// 重新加载站点数据并通知前端
	siteData := a.LoadSite()
	runtime.EventsEmit(a.ctx, EventAppSiteLoaded, siteData)

	return nil
}

// GetSites 获取站点列表
func (a *App) GetSites() []config.SiteEntry {
	cm, err := config.NewConfigManager()
	if err != nil {
		return nil
	}
	sites, _ := cm.GetSites()
	return sites
}

// AddSite 添加新站点
func (a *App) AddSite(name, path string) ([]config.SiteEntry, error) {
	cm, err := config.NewConfigManager()
	if err != nil {
		return nil, err
	}

	// 初始化站点目录
	if err := a.services.Services.Scaffold.InitSite(path); err != nil {
		return nil, fmt.Errorf("初始化站点失败: %w", err)
	}

	sites, _ := cm.GetSites()

	// 检查路径是否已存在
	for _, s := range sites {
		if s.Path == path {
			return nil, fmt.Errorf("该站点路径已存在")
		}
	}

	sites = append(sites, config.SiteEntry{
		Name:   name,
		Path:   path,
		Active: false,
	})

	if err := cm.SaveSites(sites); err != nil {
		return nil, err
	}
	return sites, nil
}

// RemoveSite 删除站点（不删除源文件）
func (a *App) RemoveSite(path string) ([]config.SiteEntry, error) {
	cm, err := config.NewConfigManager()
	if err != nil {
		return nil, err
	}

	sites, _ := cm.GetSites()
	var newSites []config.SiteEntry
	for _, s := range sites {
		if s.Path != path {
			newSites = append(newSites, s)
		}
	}

	if err := cm.SaveSites(newSites); err != nil {
		return nil, err
	}
	return newSites, nil
}

// UpdateSites 更新站点列表（用于排序）
func (a *App) UpdateSites(sites []config.SiteEntry) error {
	cm, err := config.NewConfigManager()
	if err != nil {
		return err
	}
	return cm.SaveSites(sites)
}

// SwitchSite 切换活跃站点
func (a *App) SwitchSite(path string) error {
	cm, err := config.NewConfigManager()
	if err != nil {
		return err
	}

	sites, _ := cm.GetSites()

	// 更新 active 状态
	found := false
	for i := range sites {
		sites[i].Active = sites[i].Path == path
		if sites[i].Path == path {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("站点不存在")
	}

	// 保存配置
	if err := cm.SaveSites(sites); err != nil {
		return err
	}

	// 执行热切换
	return a.switchToPath(path)
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

// OpenKeyFileDialog 选择 SSH 私钥文件
func (a *App) OpenKeyFileDialog() (string, error) {
	opts := runtime.OpenDialogOptions{
		Title: "选择 SSH 私钥文件",
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
	// 清除所有仓库缓存，确保从磁盘加载最新数据
	a.services.InvalidateAllCaches()
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
