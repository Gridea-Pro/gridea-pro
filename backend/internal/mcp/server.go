package mcp

import (
	"gridea-pro/backend/internal/engine"
	"gridea-pro/backend/internal/repository"
	"gridea-pro/backend/internal/service"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	mcpServer *server.MCPServer
	services  *Services
}

type Services struct {
	Post     *service.PostService
	Memo     *service.MemoService
	Tag      *service.TagService
	Category *service.CategoryService
	Link     *service.LinkService
	Menu     *service.MenuService
	Theme    *service.ThemeService
	Setting  *service.SettingService
	Renderer *engine.Engine
	Comment  *service.CommentService
}

func NewServer() *Server {
	// Initialize MCP Server
	s := server.NewMCPServer(
		"Gridea Pro MCP",
		"1.0.0",
		server.WithLogging(),
	)

	// Initialize Gridea Services
	appDir := GetAppDir()
	services := initServices(appDir)

	srv := &Server{
		mcpServer: s,
		services:  services,
	}

	srv.registerTools()
	srv.registerResources()
	srv.registerPrompts()

	return srv
}

func (s *Server) Start() error {
	// Stdio transport by default
	return server.ServeStdio(s.mcpServer)
}

func GetAppDir() string {
	if envDir := os.Getenv("SOURCE_DIR"); envDir != "" {
		return envDir
	}
	// 兼容旧环境变量名
	if envDir := os.Getenv("GRIDEA_SOURCE_DIR"); envDir != "" {
		return envDir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Documents", "Gridea Pro")
}

func IsDeployEnabled() bool {
	return os.Getenv("DEPLOY_ENABLED") == "true"
}

func initServices(appDir string) *Services {
	// Repositories
	postRepo := repository.NewPostRepository(appDir)
	categoryRepo := repository.NewCategoryRepository(appDir)
	tagRepo := repository.NewTagRepository(appDir)
	menuRepo := repository.NewMenuRepository(appDir)
	linkRepo := repository.NewLinkRepository(appDir)
	themeRepo := repository.NewThemeRepository(appDir)
	settingRepo := repository.NewSettingRepository(appDir)
	mediaRepo := repository.NewMediaRepository(appDir)
	memoRepo := repository.NewMemoRepository(appDir)
	commentRepo := repository.NewCommentRepository(appDir)
	seoSettingRepo := repository.NewSeoSettingRepository(appDir)
	cdnSettingRepo := repository.NewCdnSettingRepository(appDir)

	// Services
	tagService := service.NewTagService(tagRepo, postRepo)
	categoryService := service.NewCategoryService(categoryRepo, postRepo)
	postService := service.NewPostService(postRepo, tagRepo, tagService, categoryService, mediaRepo)
	menuService := service.NewMenuService(menuRepo)
	linkService := service.NewLinkService(linkRepo)
	themeService := service.NewThemeService(themeRepo, appDir)
	settingService := service.NewSettingService(appDir, settingRepo)
	commentService := service.NewCommentService(appDir, commentRepo, postRepo, themeRepo, settingRepo)
	memoService := service.NewMemoService(memoRepo)

	// Renderer (Complex dependencies)
	rendererService := engine.New(appDir, postRepo, themeRepo, settingRepo)
	rendererService.SetMenuRepo(menuRepo)
	rendererService.SetLinkRepo(linkRepo)
	rendererService.SetTagRepo(tagRepo)
	rendererService.SetMemoRepo(memoRepo)
	rendererService.SetCommentRepo(commentRepo)
	rendererService.SetCategoryRepo(categoryRepo)
	rendererService.SetSeoSettingRepo(seoSettingRepo)
	rendererService.SetCdnSettingRepo(cdnSettingRepo)

	return &Services{
		Post:     postService,
		Memo:     memoService,
		Tag:      tagService,
		Category: categoryService,
		Link:     linkService,
		Menu:     menuService,
		Theme:    themeService,
		Setting:  settingService,
		Renderer: rendererService,
		Comment:  commentService,
	}
}

func (s *Server) registerTools() {
	// Core Tools
	s.mcpServer.AddTool(listPostsTool(), listPostsHandler(s.services.Post))
	s.mcpServer.AddTool(getPostTool(), getPostHandler(s.services.Post))
	s.mcpServer.AddTool(createPostTool(), createPostHandler(s.services.Post))
	s.mcpServer.AddTool(updatePostTool(), updatePostHandler(s.services.Post))
	s.mcpServer.AddTool(deletePostTool(), deletePostHandler(s.services.Post))

	s.mcpServer.AddTool(listMemosTool(), listMemosHandler(s.services.Memo))
	s.mcpServer.AddTool(createMemoTool(), createMemoHandler(s.services.Memo))
	s.mcpServer.AddTool(updateMemoTool(), updateMemoHandler(s.services.Memo))
	s.mcpServer.AddTool(deleteMemoTool(), deleteMemoHandler(s.services.Memo))
	s.mcpServer.AddTool(getMemoStatsTool(), getMemoStatsHandler(s.services.Memo))

	// Secondary Tools
	s.mcpServer.AddTool(listTagsTool(), listTagsHandler(s.services.Tag))
	s.mcpServer.AddTool(createTagTool(), createTagHandler(s.services.Tag))
	s.mcpServer.AddTool(updateTagTool(), updateTagHandler(s.services.Tag))
	s.mcpServer.AddTool(deleteTagTool(), deleteTagHandler(s.services.Tag))

	s.mcpServer.AddTool(listCategoriesTool(), listCategoriesHandler(s.services.Category))
	s.mcpServer.AddTool(createCategoryTool(), createCategoryHandler(s.services.Category))
	s.mcpServer.AddTool(updateCategoryTool(), updateCategoryHandler(s.services.Category))
	s.mcpServer.AddTool(deleteCategoryTool(), deleteCategoryHandler(s.services.Category))

	s.mcpServer.AddTool(listLinksTool(), listLinksHandler(s.services.Link))
	s.mcpServer.AddTool(createLinkTool(), createLinkHandler(s.services.Link))
	s.mcpServer.AddTool(updateLinkTool(), updateLinkHandler(s.services.Link))
	s.mcpServer.AddTool(deleteLinkTool(), deleteLinkHandler(s.services.Link))

	s.mcpServer.AddTool(listMenusTool(), listMenusHandler(s.services.Menu))
	s.mcpServer.AddTool(createMenuTool(), createMenuHandler(s.services.Menu))
	s.mcpServer.AddTool(updateMenuTool(), updateMenuHandler(s.services.Menu))
	s.mcpServer.AddTool(deleteMenuTool(), deleteMenuHandler(s.services.Menu))

	// Config Tools
	s.mcpServer.AddTool(listThemesTool(), listThemesHandler(s.services.Theme))
	s.mcpServer.AddTool(getThemeConfigTool(), getThemeConfigHandler(s.services.Theme))
	s.mcpServer.AddTool(updateThemeConfigTool(), updateThemeConfigHandler(s.services.Theme))

	s.mcpServer.AddTool(getSettingsTool(), getSettingsHandler(s.services.Setting))
	s.mcpServer.AddTool(updateSettingsTool(), updateSettingsHandler(s.services.Setting))

	// Comment Tools
	s.mcpServer.AddTool(listCommentsTool(), listCommentsHandler(s.services.Comment))
	s.mcpServer.AddTool(replyCommentTool(), replyCommentHandler(s.services.Comment))
	s.mcpServer.AddTool(deleteCommentTool(), deleteCommentHandler(s.services.Comment))

	// Advanced Tools
	s.mcpServer.AddTool(renderSiteTool(), renderSiteHandler(s.services.Renderer))

	// Deploy Tool (opt-in via DEPLOY_ENABLED=true)
	if IsDeployEnabled() {
		s.mcpServer.AddTool(deploySiteTool(), deploySiteHandler(s.services.Setting, s.services.Renderer, GetAppDir()))
	}
}
