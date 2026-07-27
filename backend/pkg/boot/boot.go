package boot

import (
	"context"
	"embed"
	"fmt"
	"gridea-pro/backend/internal/app"
	"gridea-pro/backend/internal/config"
	"gridea-pro/backend/internal/facade"
	"gridea-pro/backend/internal/localfileauth"
	versionpkg "gridea-pro/backend/internal/version"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var appCtx context.Context

// 菜单重建所需的包级引用
var (
	menuApplication  *app.App
	menuAboutHandler func(*menu.CallbackData)
	menuQuitHandler  func(*menu.CallbackData)
	menuAppDir       string
)

// NewFileServerMiddleware creates a secure middleware for serving local files.
// It explicitly denies access to any file outside the rootPath to prevent directory traversal attacks.
func NewFileServerMiddleware(rootPath string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 处理 /post-images/ 请求：从 appDir 提供文章内容图片
			if strings.HasPrefix(r.URL.Path, "/post-images/") {
				filePath := filepath.Join(rootPath, r.URL.Path)
				absPath, err := filepath.Abs(filePath)
				cleanRoot := filepath.Clean(rootPath)
				// 前缀比较必须带路径分隔符边界，否则同级目录（如 "<root>-backup"）能被
				// ../ 穿透读取——例如另一个站点含 AI Key 的 config.json。
				if err != nil || (absPath != cleanRoot && !strings.HasPrefix(absPath, cleanRoot+string(os.PathSeparator))) {
					http.Error(w, "Invalid path", http.StatusBadRequest)
					return
				}
				http.ServeFile(w, r, absPath)
				return
			}

			// Only handle /local-file requests
			if r.URL.Path != "/local-file" {
				next.ServeHTTP(w, r)
				return
			}

			// Get the requested file path
			requestedPath := r.URL.Query().Get("path")
			if requestedPath == "" {
				http.Error(w, "Missing path parameter", http.StatusBadRequest)
				return
			}

			// Security Check: Prevent Path Traversal
			// 1. Resolve absolute path of the requested file
			absRequestedPath, err := filepath.Abs(requestedPath)
			if err != nil {
				http.Error(w, "Invalid file path", http.StatusBadRequest)
				return
			}

			// 安全 1：只允许常见图片扩展名，堵住"webview XSS → fetch('/local-file?path=/任意文件') 外传"。
			switch strings.ToLower(filepath.Ext(absRequestedPath)) {
			case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".svg", ".ico", ".avif", ".tiff", ".tif":
			default:
				http.Error(w, "Forbidden file type", http.StatusForbidden)
				return
			}

			// 安全 2：仅放行「站点目录内」或「用户经文件对话框主动选择并授权」的路径。
			// 单靠扩展名白名单仍允许读取站点外任意图片（隐私图片外传）；这里再收一道：
			//  - 站点内路径（配置里的头像/特色图/主题资源，均在 appDir 下）：直接放行；
			//  - 站点外绝对路径：必须由 OpenImageDialog 等主动选择入口授权过（localfileauth）。
			cleanRoot := filepath.Clean(rootPath)
			inSite := absRequestedPath == cleanRoot || strings.HasPrefix(absRequestedPath, cleanRoot+string(os.PathSeparator))
			if !inSite && !localfileauth.IsAuthorized(absRequestedPath) {
				http.Error(w, "Forbidden path", http.StatusForbidden)
				return
			}

			// 3. Ensure the file exists and is not a directory
			fileInfo, err := os.Stat(absRequestedPath)
			if err != nil {
				if os.IsNotExist(err) {
					http.Error(w, "File not found", http.StatusNotFound)
				} else {
					http.Error(w, "Error accessing file", http.StatusInternalServerError)
				}
				return
			}

			if fileInfo.IsDir() {
				http.Error(w, "Cannot serve directories", http.StatusForbidden)
				return
			}

			// 安全 3：SVG 可内嵌脚本，若被直接导航/加载为文档会在应用同源执行。加 CSP 禁掉脚本与插件，
			// 既不影响 <img> 内联预览，也堵死"同源 SVG 脚本"这条 XSS 路径。
			w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src 'self' data:; style-src 'unsafe-inline'; sandbox")
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Serve the validated file
			http.ServeFile(w, r, absRequestedPath)
		})
	}
}

// openInOS 使用系统默认程序打开文件或文件夹（跨平台）
func openInOS(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "windows":
		cmd = exec.Command("explorer", path)
	default:
		log.Printf("Unsupported OS for open: %s", runtime.GOOS)
		return
	}
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to open path %s: %v", path, err)
	}
}

// openInTerminal 在系统终端中打开指定目录（跨平台）
func openInTerminal(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-a", "Terminal", path)
	case "linux":
		// 尝试常见的 Linux 终端模拟器
		cmd = exec.Command("x-terminal-emulator", "--working-directory", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "cmd", "/k", "cd", "/d", path)
	default:
		log.Printf("Unsupported OS for terminal: %s", runtime.GOOS)
		return
	}
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to open terminal at %s: %v", path, err)
	}
}

func Run(assets embed.FS, version string) {
	// 把 main 侧 ldflags 注入的版本号写入 version 包，作为全局唯一真源
	// （main 无法直接导入 internal/version，所以由 boot 做桥接）
	if version != "" {
		versionpkg.Version = version
	}

	// 初始化 ConfigManager
	configManager, err := config.NewConfigManager()
	if err != nil {
		log.Printf("Warning: Failed to initialize config manager: %v", err)
	}

	// 初始化路径：多站点模式，优先从 Sites 列表找活跃站点
	var appDir string
	home, _ := os.UserHomeDir()
	defaultPath := filepath.Join(home, "Documents", "Gridea Pro")

	if configManager != nil {
		// 迁移旧配置 / 确保 Sites 列表存在
		sites, migrateErr := configManager.MigrateToSites(defaultPath)
		if migrateErr != nil {
			log.Printf("Warning: Failed to migrate sites: %v", migrateErr)
		}
		// 从 Sites 列表找活跃站点
		for _, s := range sites {
			if s.Active {
				if _, err := os.Stat(s.Path); err == nil {
					appDir = s.Path
				}
				break
			}
		}
	}

	// 兜底：使用默认路径
	if appDir == "" {
		appDir = defaultPath
	}

	// 初始化 Services (Facade)
	services := facade.NewAppServices(appDir, assets)

	application := app.NewApp(appDir, services, version)

	// Capture context for safe shutdown
	quitHandler := func(_ *menu.CallbackData) {
		if appCtx != nil {
			wailsRuntime.Quit(appCtx)
		} else {
			log.Println("Context missing, performing force exit")
			os.Exit(0)
		}
	}

	aboutHandler := func(_ *menu.CallbackData) {
		if appCtx != nil {
			_, _ = wailsRuntime.MessageDialog(appCtx, wailsRuntime.MessageDialogOptions{
				Type:    wailsRuntime.InfoDialog,
				Title:   T("app.about"),
				Message: "Gridea Pro\nVersion " + version + "\nCopyright © 2026 Gridea Pro",
			})
		}
	}

	// 保存菜单构建参数，用于运行时重建
	menuApplication = application
	menuAboutHandler = aboutHandler
	menuQuitHandler = quitHandler
	menuAppDir = appDir

	// Windows/Linux 使用 Frameless 窗口，前端自绘窗口控制按钮
	// macOS 保持原生标题栏 + 交通灯按钮
	frameless := runtime.GOOS != "darwin"

	var appMenu *menu.Menu
	if !frameless {
		// 仅 macOS 构建原生菜单
		appMenu = buildMenu(application, aboutHandler, quitHandler, appDir)
	}

	// Create application with options
	err = wails.Run(&options.App{
		Title:     "Gridea Pro",
		Width:     1280,
		Height:    800,
		Frameless: frameless,
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Middleware: NewFileServerMiddleware(appDir), // Inject secure middleware
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		StartHidden:      true,
		OnStartup: func(ctx context.Context) {
			appCtx = ctx // Capture context
			application.Startup(ctx)
			services.Startup(ctx)

			// 监听前端语言切换事件，运行时重建菜单（仅 macOS 有原生菜单）
			if !frameless {
				wailsRuntime.EventsOn(ctx, "app:change-locale", func(optionalData ...interface{}) {
					if len(optionalData) > 0 {
						if locale, ok := optionalData[0].(string); ok {
							SetLocale(locale)
							newMenu := buildMenu(menuApplication, menuAboutHandler, menuQuitHandler, menuAppDir)
							wailsRuntime.MenuSetApplicationMenu(ctx, newMenu)
							wailsRuntime.MenuUpdateApplicationMenu(ctx)
						}
					}
				})
			}
		},
		// StartHidden 下窗口需显式显示。前端 safeWindowShow 受 Wails 运行时注入时序影响，
		// 挂载时机早于运行时就绪时会跳过显示，导致主窗口永久隐藏（#102）。
		// OnDomReady 在运行时注入完成后才触发，在此兜底显示一次，消除该竞态。
		OnDomReady: func(ctx context.Context) {
			wailsRuntime.WindowShow(ctx)
		},
		OnShutdown: application.Shutdown,
		Menu:       appMenu,
		Bind: []interface{}{
			application,
			services.Category,
			services.Post,
			services.Menu,
			services.Link,
			services.Tag,
			services.Deploy,
			services.Theme,
			services.Renderer,
			services.Setting,
			services.Memo,
			services.Comment,
			services.Preview,
			services.SeoSetting,
			services.CdnSetting,
			services.CdnUpload,
			services.PwaSetting,
			services.AI,
			services.OAuth,
			services.ImageHosting,
			services.ImageOptimizeSetting,
			services.Update,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableFramelessWindowDecorations: false,
			Theme:                             windows.SystemDefault,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
		},
	})

	if err != nil {
		fmt.Printf("Error during application run: %v\n", err)
	}
}

// emitEvent 安全地向前端发送事件
func emitEvent(event string, data ...interface{}) {
	if appCtx != nil {
		wailsRuntime.EventsEmit(appCtx, event, data...)
	}
}

// buildMenu creates the full application menu.
func buildMenu(
	application *app.App,
	aboutHandler func(*menu.CallbackData),
	quitHandler func(*menu.CallbackData),
	appDir string,
) *menu.Menu {
	appMenu := menu.NewMenu()

	// ─────────────────────────────────────────────
	// 1. 应用主菜单 (macOS only: "Gridea Pro")
	// ─────────────────────────────────────────────
	if runtime.GOOS == "darwin" {
		appSubMenu := appMenu.AddSubmenu("Gridea Pro")

		appSubMenu.AddText(T("app.about"), nil, aboutHandler)
		appSubMenu.AddSeparator()

		appSubMenu.AddText(T("app.preferences"), keys.CmdOrCtrl(","), func(_ *menu.CallbackData) {
			application.ShowPreferences()
		})

		appSubMenu.AddText(T("app.checkUpdate"), nil, func(_ *menu.CallbackData) {
			emitEvent("menu:check-update")
		})
		appSubMenu.AddSeparator()

		appSubMenu.AddText(T("app.hide"), keys.CmdOrCtrl("h"), func(_ *menu.CallbackData) {
			if appCtx != nil {
				wailsRuntime.Hide(appCtx)
			}
		})
		appSubMenu.AddText(T("app.hideOthers"), nil, nil)
		appSubMenu.AddText(T("app.showAll"), nil, func(_ *menu.CallbackData) {
			if appCtx != nil {
				wailsRuntime.Show(appCtx)
			}
		})
		appSubMenu.AddSeparator()
		appSubMenu.AddText(T("app.quit"), keys.CmdOrCtrl("q"), quitHandler)
	}

	// ─────────────────────────────────────────────
	// 2. 文件 (File)
	// ─────────────────────────────────────────────
	fileMenu := appMenu.AddSubmenu(T("file"))

	fileMenu.AddText(T("file.newPost"), keys.CmdOrCtrl("n"), func(_ *menu.CallbackData) {
		emitEvent("menu:new-post")
	})
	fileMenu.AddText(T("file.newPage"), keys.Combo("n", keys.CmdOrCtrlKey, keys.ShiftKey), func(_ *menu.CallbackData) {
		emitEvent("menu:new-page")
	})
	fileMenu.AddSeparator()

	fileMenu.AddText(T("file.openSite"), keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		openInOS(appDir)
	})
	fileMenu.AddText(T("file.openTerminal"), nil, func(_ *menu.CallbackData) {
		openInTerminal(appDir)
	})
	fileMenu.AddSeparator()

	fileMenu.AddText(T("file.save"), keys.CmdOrCtrl("s"), func(_ *menu.CallbackData) {
		emitEvent("menu:save")
	})
	fileMenu.AddSeparator()

	fileMenu.AddText(T("file.import"), nil, func(_ *menu.CallbackData) {
		emitEvent("menu:import")
	})
	fileMenu.AddText(T("file.export"), nil, func(_ *menu.CallbackData) {
		emitEvent("menu:export")
	})

	// Windows/Linux: 在文件菜单中也放置首选项和退出
	if runtime.GOOS != "darwin" {
		fileMenu.AddSeparator()
		fileMenu.AddText(T("app.preferences"), keys.CmdOrCtrl(","), func(_ *menu.CallbackData) {
			application.ShowPreferences()
		})
		fileMenu.AddSeparator()
		fileMenu.AddText(T("app.quit"), keys.CmdOrCtrl("q"), quitHandler)
	}

	// ─────────────────────────────────────────────
	// 3. 编辑 (Edit)
	// ─────────────────────────────────────────────
	if runtime.GOOS == "darwin" {
		// macOS: 必须使用原生 Edit 菜单，否则 WKWebView 的 Cut/Copy/Paste 无法工作
		appMenu.Append(menu.EditMenu())
	} else {
		// Windows/Linux: WebView2/WebKit2GTK 原生处理 Ctrl+C/V
		// 注：Tiptap 编辑器无内置查找/替换，故不再提供 Find/Replace 菜单项——
		// 保留会让用户按 Cmd+F 却毫无反应，反成体验缺陷。仅保留已实现的「复制 HTML」。
		editMenu := appMenu.AddSubmenu(T("edit"))
		editMenu.AddText(T("edit.copyHTML"), nil, func(_ *menu.CallbackData) {
			emitEvent("menu:copy-html")
		})
	}

	// ─────────────────────────────────────────────
	// 4. 视图 (View)
	// ─────────────────────────────────────────────
	viewMenu := appMenu.AddSubmenu(T("view"))

	if runtime.GOOS == "darwin" {
		// macOS: CopyHTML 放入 View 菜单（原生 Edit 菜单无法自定义）。
		// Find/Replace 已移除——Tiptap 无内置查找替换，保留只会让 Cmd+F 无响应误导用户。
		viewMenu.AddText(T("edit.copyHTML"), nil, func(_ *menu.CallbackData) {
			emitEvent("menu:copy-html")
		})
		viewMenu.AddSeparator()
	}

	viewMenu.AddText(T("view.toggleSidebar"), keys.CmdOrCtrl("b"), func(_ *menu.CallbackData) {
		emitEvent("menu:toggle-sidebar")
	})
	viewMenu.AddText(T("view.togglePreview"), keys.CmdOrCtrl("p"), func(_ *menu.CallbackData) {
		emitEvent("menu:toggle-preview")
	})
	viewMenu.AddSeparator()

	viewMenu.AddText(T("view.actualSize"), keys.CmdOrCtrl("0"), func(_ *menu.CallbackData) {
		emitEvent("menu:zoom-reset")
	})
	viewMenu.AddText(T("view.zoomIn"), keys.CmdOrCtrl("="), func(_ *menu.CallbackData) {
		emitEvent("menu:zoom-in")
	})
	viewMenu.AddText(T("view.zoomOut"), keys.CmdOrCtrl("-"), func(_ *menu.CallbackData) {
		emitEvent("menu:zoom-out")
	})
	viewMenu.AddSeparator()

	if runtime.GOOS == "darwin" {
		viewMenu.AddText(T("view.fullscreen"), keys.Combo("f", keys.CmdOrCtrlKey, keys.ControlKey), func(_ *menu.CallbackData) {
			if appCtx != nil {
				wailsRuntime.WindowFullscreen(appCtx)
			}
		})
	} else {
		viewMenu.AddText(T("view.fullscreen"), keys.Key("F11"), func(_ *menu.CallbackData) {
			if appCtx != nil {
				wailsRuntime.WindowFullscreen(appCtx)
			}
		})
	}
	viewMenu.AddSeparator()

	viewMenu.AddText(T("view.devTools"), keys.Combo("i", keys.CmdOrCtrlKey, keys.OptionOrAltKey), func(_ *menu.CallbackData) {
		emitEvent("menu:dev-tools")
	})

	// ─────────────────────────────────────────────
	// 5. 主题 (Theme)
	// ─────────────────────────────────────────────
	themeMenu := appMenu.AddSubmenu(T("theme"))

	themeMenu.AddText(T("theme.settings"), nil, func(_ *menu.CallbackData) {
		emitEvent("menu:navigate", "/theme")
	})
	themeMenu.AddText(T("theme.refresh"), nil, func(_ *menu.CallbackData) {
		emitEvent("menu:refresh-themes")
	})
	themeMenu.AddSeparator()

	themeMenu.AddText(T("theme.openDir"), nil, func(_ *menu.CallbackData) {
		themesDir := filepath.Join(appDir, "themes")
		openInOS(themesDir)
	})

	// ─────────────────────────────────────────────
	// 6. 站点 (Site) — 核心特色菜单
	// ─────────────────────────────────────────────
	siteMenu := appMenu.AddSubmenu(T("site"))

	siteMenu.AddText(T("site.preview"), keys.CmdOrCtrl("r"), func(_ *menu.CallbackData) {
		emitEvent("preview-site") // 复用已有事件
	})
	siteMenu.AddText(T("site.deploy"), keys.CmdOrCtrl("d"), func(_ *menu.CallbackData) {
		emitEvent("publish-site") // 复用已有事件
	})
	siteMenu.AddSeparator()

	siteMenu.AddText(T("site.clearCache"), nil, func(_ *menu.CallbackData) {
		outputDir := filepath.Join(appDir, "output")
		if err := os.RemoveAll(outputDir); err != nil {
			log.Printf("Failed to clear cache: %v", err)
			emitEvent("app:toast", map[string]interface{}{
				"message":  T("toast.cacheClearFail") + err.Error(),
				"type":     "error",
				"duration": 3000,
			})
		} else {
			// 重新创建 output 目录
			_ = os.MkdirAll(outputDir, 0755)
			emitEvent("app:toast", map[string]interface{}{
				"message":  T("toast.cacheClearSuccess"),
				"type":     "success",
				"duration": 3000,
			})
		}
	})

	siteMenu.AddText(T("site.openOutput"), nil, func(_ *menu.CallbackData) {
		outputDir := filepath.Join(appDir, "output")
		_ = os.MkdirAll(outputDir, 0755) // 确保目录存在
		openInOS(outputDir)
	})

	// ─────────────────────────────────────────────
	// 7. 窗口 (Window) — 手动构建以支持多语言
	// ─────────────────────────────────────────────
	windowMenu := appMenu.AddSubmenu(T("window"))

	windowMenu.AddText(T("window.minimize"), keys.CmdOrCtrl("m"), func(_ *menu.CallbackData) {
		if appCtx != nil {
			wailsRuntime.WindowMinimise(appCtx)
		}
	})
	windowMenu.AddText(T("window.zoom"), nil, func(_ *menu.CallbackData) {
		if appCtx != nil {
			wailsRuntime.WindowMaximise(appCtx)
		}
	})
	windowMenu.AddSeparator()
	windowMenu.AddText(T("window.close"), keys.CmdOrCtrl("w"), func(_ *menu.CallbackData) {
		if appCtx != nil {
			wailsRuntime.Hide(appCtx)
		}
	})
	windowMenu.AddSeparator()
	windowMenu.AddText(T("window.front"), nil, func(_ *menu.CallbackData) {
		if appCtx != nil {
			wailsRuntime.WindowSetAlwaysOnTop(appCtx, true)
			wailsRuntime.WindowSetAlwaysOnTop(appCtx, false)
		}
	})

	// ─────────────────────────────────────────────
	// 8. 帮助 (Help)
	// ─────────────────────────────────────────────
	helpMenu := appMenu.AddSubmenu(T("help"))

	helpMenu.AddText(T("help.docs"), nil, func(_ *menu.CallbackData) {
		if appCtx != nil {
			wailsRuntime.BrowserOpenURL(appCtx, "https://gridea.pro")
		}
	})
	helpMenu.AddText(T("help.feedback"), nil, func(_ *menu.CallbackData) {
		if appCtx != nil {
			wailsRuntime.BrowserOpenURL(appCtx, "https://github.com/Gridea-Pro/gridea-pro/issues")
		}
	})
	helpMenu.AddSeparator()

	helpMenu.AddText(T("help.viewLogs"), nil, func(_ *menu.CallbackData) {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return
		}
		logDir := filepath.Join(configDir, config.AppName)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			_ = os.MkdirAll(logDir, 0755)
		}
		// 再次检查，确保目录存在（或已被创建）
		if _, err := os.Stat(logDir); err == nil {
			openInOS(logDir)
		} else {
			// Fallback: 如果创建失败，打开 appDir
			openInOS(appDir)
		}
	})

	// Windows/Linux: 在帮助菜单中也放置"关于"
	if runtime.GOOS != "darwin" {
		helpMenu.AddSeparator()
		helpMenu.AddText(T("app.about"), nil, aboutHandler)
	}

	return appMenu
}
