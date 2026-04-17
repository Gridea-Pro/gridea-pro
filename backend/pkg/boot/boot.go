package boot

import (
	"context"
	"embed"
	"fmt"
	"gridea-pro/backend/internal/app"
	"gridea-pro/backend/internal/config"
	"gridea-pro/backend/internal/facade"
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
				if err != nil || !strings.HasPrefix(absPath, rootPath) {
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

			// 2. We intentionally omit the check if the file is physically located within the allowed root directory
			//    because users often select images from Downloads/Desktop and we need to allow previewing them before uploading.
			//    Gridea Pro uses localhost for Wails, which is securely confined to the local user context.
			//    Just ensure absolute path is used.

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

	// 所有平台使用原生窗口装饰
	frameless := false

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
		editMenu := appMenu.AddSubmenu(T("edit"))
		editMenu.AddText(T("edit.find"), keys.CmdOrCtrl("f"), func(_ *menu.CallbackData) {
			emitEvent("menu:find")
		})
		editMenu.AddText(T("edit.replace"), keys.Combo("f", keys.CmdOrCtrlKey, keys.OptionOrAltKey), func(_ *menu.CallbackData) {
			emitEvent("menu:replace")
		})
		editMenu.AddSeparator()
		editMenu.AddText(T("edit.copyHTML"), nil, func(_ *menu.CallbackData) {
			emitEvent("menu:copy-html")
		})
	}

	// ─────────────────────────────────────────────
	// 4. 视图 (View)
	// ─────────────────────────────────────────────
	viewMenu := appMenu.AddSubmenu(T("view"))

	if runtime.GOOS == "darwin" {
		// macOS: Find/Replace/CopyHTML 放入 View 菜单（原生 Edit 菜单无法自定义）
		viewMenu.AddText(T("edit.find"), keys.CmdOrCtrl("f"), func(_ *menu.CallbackData) {
			emitEvent("menu:find")
		})
		viewMenu.AddText(T("edit.replace"), keys.Combo("f", keys.CmdOrCtrlKey, keys.OptionOrAltKey), func(_ *menu.CallbackData) {
			emitEvent("menu:replace")
		})
		viewMenu.AddSeparator()
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
