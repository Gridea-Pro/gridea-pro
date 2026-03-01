package facade

import (
	"context"
	"embed"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/repository"
	"gridea-pro/backend/internal/service"
	"path/filepath"
	"sync"
)

// WailsContext holds the global application context
var WailsContext context.Context

type AppServices struct {
	mu       sync.RWMutex
	Category *CategoryFacade
	Post     *PostFacade
	Menu     *MenuFacade
	Link     *LinkFacade
	Tag      *TagFacade
	Deploy   *DeployFacade
	Renderer *RendererFacade
	Theme    *ThemeFacade
	Setting  *SettingFacade
	Comment  *CommentFacade
	Memo     *MemoFacade
	Preview  *PreviewFacade
	// Internal services for event/update handling
	Services struct {
		Category *service.CategoryService
		Post     *service.PostService
		Menu     *service.MenuService
		Link     *service.LinkService
		Tag      *service.TagService
		Deploy   *service.DeployService
		Renderer *service.RendererService
		Theme    *service.ThemeService
		Setting  *service.SettingService
		Scaffold *service.ScaffoldService
		Comment  *service.CommentService
		Memo     *service.MemoService
		Preview  *service.PreviewService
	}
	Repositories struct {
		Category domain.CategoryRepository
		Tag      domain.TagRepository
		Post     domain.PostRepository
		Menu     domain.MenuRepository
		Link     domain.LinkRepository
		Memo     domain.MemoRepository
	}
	assets embed.FS // Keep reference for UpdateAppDir
}

// Startup captures the Wails context on application start
func (s *AppServices) Startup(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	WailsContext = ctx
}

func NewAppServices(appDir string, assets embed.FS) *AppServices {
	// 1. Init Repositories
	postRepo := repository.NewPostRepository(appDir)
	categoryRepo := repository.NewCategoryRepository(appDir)
	tagRepo := repository.NewTagRepository(appDir)
	menuRepo := repository.NewMenuRepository(appDir)
	linkRepo := repository.NewLinkRepository(appDir)
	themeRepo := repository.NewThemeRepository(appDir)
	settingRepo := repository.NewSettingRepository(appDir)
	mediaRepo := repository.NewMediaRepository(appDir)
	memoRepo := repository.NewMemoRepository(appDir)

	// 2. Init Services
	tagService := service.NewTagService(tagRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	postService := service.NewPostService(postRepo, tagRepo, tagService, categoryService, mediaRepo)
	menuService := service.NewMenuService(menuRepo)
	linkService := service.NewLinkService(linkRepo)
	themeService := service.NewThemeService(themeRepo, appDir)
	deployService := service.NewDeployService(settingRepo, appDir)
	// RendererService
	rendererService := service.NewRendererService(appDir, postRepo, themeRepo, settingRepo)
	rendererService.SetMenuRepo(menuRepo)
	rendererService.SetLinkRepo(linkRepo)
	rendererService.SetTagRepo(tagRepo)
	rendererService.SetMemoRepo(memoRepo)
	rendererService.SetCategoryRepo(categoryRepo)
	settingService := service.NewSettingService(appDir, settingRepo)
	scaffoldService := service.NewScaffoldService(assets)
	// CommentService
	commentRepo := repository.NewCommentRepository(appDir)
	commentService := service.NewCommentService(appDir, commentRepo, postRepo, themeRepo)
	memoService := service.NewMemoService(memoRepo)
	previewService := service.NewPreviewService(filepath.Join(appDir, "output"))
	// Set CommentRepo on RendererService for template injection
	rendererService.SetCommentRepo(commentRepo)

	// 3. Wrap with Facades
	return &AppServices{
		Category: NewCategoryFacade(categoryService),
		Post:     NewPostFacade(postService),
		Menu:     NewMenuFacade(menuService),
		Link:     NewLinkFacade(linkService),
		Tag:      NewTagFacade(tagService),
		Deploy:   NewDeployFacade(deployService),
		Renderer: NewRendererFacade(rendererService),
		Theme:    NewThemeFacade(themeService),
		Setting:  NewSettingFacade(settingService),
		Comment:  NewCommentFacade(commentService),
		Memo:     NewMemoFacade(memoService),
		Preview:  NewPreviewFacade(previewService),
		Services: struct {
			Category *service.CategoryService
			Post     *service.PostService
			Menu     *service.MenuService
			Link     *service.LinkService
			Tag      *service.TagService
			Deploy   *service.DeployService
			Renderer *service.RendererService
			Theme    *service.ThemeService
			Setting  *service.SettingService
			Scaffold *service.ScaffoldService
			Comment  *service.CommentService
			Memo     *service.MemoService
			Preview  *service.PreviewService
		}{
			Category: categoryService,
			Post:     postService,
			Menu:     menuService,
			Link:     linkService,
			Tag:      tagService,
			Deploy:   deployService,
			Renderer: rendererService,
			Theme:    themeService,
			Setting:  settingService,
			Scaffold: scaffoldService,
			Comment:  commentService,
			Memo:     memoService,
			Preview:  previewService,
		},
		Repositories: struct {
			Category domain.CategoryRepository
			Tag      domain.TagRepository
			Post     domain.PostRepository
			Menu     domain.MenuRepository
			Link     domain.LinkRepository
			Memo     domain.MemoRepository
		}{
			Category: categoryRepo,
			Tag:      tagRepo,
			Post:     postRepo,
			Menu:     menuRepo,
			Link:     linkRepo,
			Memo:     memoRepo,
		},
		assets: assets,
	}
}

func (s *AppServices) UpdateAppDir(appDir string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-initialize logic
	newServices := NewAppServices(appDir, s.assets)
	s.Category.internal = newServices.Services.Category
	s.Post.internal = newServices.Services.Post
	s.Menu.internal = newServices.Services.Menu
	s.Link.internal = newServices.Services.Link
	s.Tag.internal = newServices.Services.Tag
	s.Deploy.internal = newServices.Services.Deploy
	s.Renderer.internal = newServices.Services.Renderer
	s.Theme.internal = newServices.Services.Theme
	s.Setting.internal = newServices.Services.Setting
	s.Comment.internal = newServices.Services.Comment
	s.Memo.internal = newServices.Services.Memo
	s.Preview.internal = newServices.Services.Preview
	// Scaffold service doesn't need update generally, but good to keep in sync
	s.Services.Scaffold = newServices.Services.Scaffold
	s.Services.Comment = newServices.Services.Comment
	s.Services.Memo = newServices.Services.Memo
	s.Services.Preview = newServices.Services.Preview
}

func (s *AppServices) RegisterEvents(ctx context.Context) {
	// Inject dependencies
	s.Theme.SetRenderer(s.Renderer)

	// Register app-site-reload event
	s.Renderer.RegisterEvents(ctx)

	// Register link events
	s.Link.RegisterEvents(ctx)

	// Register menu events
	s.Menu.RegisterEvents(ctx)

	// Register category events
	s.Category.RegisterEvents(ctx)

	// Register tag events
	s.Tag.RegisterEvents(ctx)
}
