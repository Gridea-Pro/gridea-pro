package facade

import (
	"context"
	"embed"
	"gridea-pro/backend/internal/repository"
	"gridea-pro/backend/internal/service"
)

// AppServices holds all facades
type AppServices struct {
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
	}
	assets embed.FS // Keep reference for UpdateAppDir
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

	// 2. Init Services
	tagService := service.NewTagService(tagRepo)
	postService := service.NewPostService(postRepo, tagRepo, tagService, mediaRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	menuService := service.NewMenuService(menuRepo)
	linkService := service.NewLinkService(linkRepo)
	themeService := service.NewThemeService(themeRepo)
	deployService := service.NewDeployService(settingRepo, appDir)
	// RendererService
	rendererService := service.NewRendererService(appDir, postRepo, themeRepo, settingRepo)
	rendererService.SetMenuRepo(menuRepo)
	settingService := service.NewSettingService(appDir, settingRepo)
	scaffoldService := service.NewScaffoldService(assets)
	// CommentService
	commentRepo := repository.NewCommentRepository(appDir)
	commentService := service.NewCommentService(appDir, commentRepo, postRepo, themeRepo)
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
		},
		assets: assets,
	}
}

func (s *AppServices) UpdateAppDir(appDir string) {
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
	// Scaffold service doesn't need update generally, but good to keep in sync
	s.Services.Scaffold = newServices.Services.Scaffold
	s.Services.Comment = newServices.Services.Comment
}

func (s *AppServices) RegisterEvents(ctx context.Context) {
	// Register theme-save event (needs renderer)
	s.Theme.RegisterEvents(ctx, s.Renderer)

	// Register app-site-reload event
	s.Renderer.RegisterEvents(ctx)

	// Register post events
	s.Post.RegisterEvents(ctx)

	// Register link events
	s.Link.RegisterEvents(ctx)

	// Register menu events
	s.Menu.RegisterEvents(ctx)

	// Register category events
	s.Category.RegisterEvents(ctx)

	// Register tag events
	s.Tag.RegisterEvents(ctx)
}
