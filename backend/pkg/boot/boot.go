package boot

import (
	"context"
	"embed"
	"gridea-pro/backend/internal/app"
	"gridea-pro/backend/internal/config"
	"gridea-pro/backend/internal/facade"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

// NewLocalFileHandler creates middleware for serving local files
func NewLocalFileMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only handle /local-file requests
		if r.URL.Path != "/local-file" {
			// Pass to next handler for all other requests
			next.ServeHTTP(w, r)
			return
		}

		// Get the file path from query parameter
		filePath := r.URL.Query().Get("path")
		if filePath == "" {
			http.Error(w, "Missing path parameter", http.StatusBadRequest)
			return
		}

		// Security check: ensure the file exists and is accessible
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "File not found", http.StatusNotFound)
			} else {
				http.Error(w, "Error accessing file", http.StatusInternalServerError)
			}
			return
		}

		// Don't allow directory access
		if fileInfo.IsDir() {
			http.Error(w, "Cannot serve directories", http.StatusForbidden)
			return
		}

		// Serve the file
		http.ServeFile(w, r, filePath)
	})
}

func Run(assets embed.FS) {
	// 初始化 ConfigManager
	configManager := config.NewConfigManager()
	conf, _ := configManager.LoadConfig()

	// 初始化路径：优先使用配置中的路径，否则使用默认的 Documents/Gridea
	var appDir string
	home, _ := os.UserHomeDir()

	if conf != nil && conf.SourceFolder != "" {
		// 验证配置中的路径是否存在
		if _, err := os.Stat(conf.SourceFolder); err == nil {
			appDir = conf.SourceFolder
		}
	}

	// 如果没有配置或配置路径无效，使用默认路径
	if appDir == "" {
		docs := filepath.Join(home, "Documents")
		appDir = filepath.Join(docs, "Gridea Pro")
	}

	// 初始化 Services (Facade)
	services := facade.NewAppServices(appDir, assets)

	application := app.NewApp(appDir, services)
	prefsWindow := app.NewPreferencesWindow()

	// 创建应用菜单
	appMenu := menu.NewMenu()

	// 应用程序菜单 (macOS)
	if AppMenu := appMenu.AddSubmenu("Gridea Pro"); AppMenu != nil {
		AppMenu.AddText("首选项...", keys.CmdOrCtrl(","), func(_ *menu.CallbackData) {
			prefsWindow.Show()
		})
		AppMenu.AddSeparator()
		AppMenu.AddText("退出", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
			os.Exit(0)
		})
	}

	FileMenu := appMenu.AddSubmenu("File")
	FileMenu.AddText("Open Site Folder...", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		// Open folder logic
		if paths, err := application.OpenFolderDialog(); err == nil && len(paths) > 0 {
			log.Println("Selected folder:", paths[0])
		}
	})
	FileMenu.AddSeparator()
	FileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		os.Exit(0)
	})

	appMenu.Append(menu.EditMenu())   // Edit menu (Copy, Paste, etc.)
	appMenu.Append(menu.WindowMenu()) // Window menu

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Gridea Pro",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Middleware: NewLocalFileMiddleware,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		StartHidden:      true,
		OnStartup: func(ctx context.Context) {
			prefsWindow.SetContext(ctx)
			application.Startup(ctx)
		},
		OnShutdown: application.Shutdown,
		Menu:       appMenu,
		Bind: []interface{}{
			application,
			// Bind Facades
			services.Category,
			services.Post,
			services.Menu,
			services.Link,
			services.Tag, // Added TagFacade
			services.Deploy,
			services.Theme,
			services.Renderer,
			services.Setting,
			services.Comment, // Added CommentFacade
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
