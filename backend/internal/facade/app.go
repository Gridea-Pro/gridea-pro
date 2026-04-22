package facade

import (
	"context"
	"embed"
	"gridea-pro/backend/internal/config"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/engine"
	"gridea-pro/backend/internal/repository"
	"gridea-pro/backend/internal/service"
	"gridea-pro/backend/internal/service/credential"
	"path/filepath"
	"sync"
)

// WailsContext holds the global application context
var WailsContext context.Context

type AppServices struct {
	mu         sync.RWMutex
	Category   *CategoryFacade
	Post       *PostFacade
	Menu       *MenuFacade
	Link       *LinkFacade
	Tag        *TagFacade
	Deploy     *DeployFacade
	Renderer   *RendererFacade
	Theme      *ThemeFacade
	Setting    *SettingFacade
	Comment    *CommentFacade
	Memo       *MemoFacade
	Preview    *PreviewFacade
	SeoSetting *SeoSettingFacade
	CdnSetting *CdnSettingFacade
	PwaSetting *PwaSettingFacade
	CdnUpload  *CdnUploadFacade
	AI         *AIFacade
	OAuth      *OAuthFacade
	Update     *UpdateFacade
	// Internal services for event/update handling
	Services struct {
		Category  *service.CategoryService
		Post      *service.PostService
		Menu      *service.MenuService
		Link      *service.LinkService
		Tag       *service.TagService
		Deploy    *service.DeployService
		Renderer  *engine.Engine
		Theme     *service.ThemeService
		Setting   *service.SettingService
		Scaffold  *service.ScaffoldService
		Comment   *service.CommentService
		Memo      *service.MemoService
		Preview   *service.PreviewService
		CdnUpload *service.CdnUploadService
	}
	Repositories struct {
		Category domain.CategoryRepository
		Tag      domain.TagRepository
		Post     domain.PostRepository
		Menu     domain.MenuRepository
		Link     domain.LinkRepository
		Memo     domain.MemoRepository
		Setting  domain.SettingRepository
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
	seoSettingRepo := repository.NewSeoSettingRepository(appDir)
	cdnSettingRepo := repository.NewCdnSettingRepository(appDir)
	pwaSettingRepo := repository.NewPwaSettingRepository(appDir)
	// AI 配置和调用计数器使用应用级目录（跨站点共享，避免敏感信息泄露到站点目录）
	cm, _ := config.NewConfigManager()
	aiSettingRepo := repository.NewAISettingRepository(cm)
	aiUsageRepo := repository.NewAIUsageRepository(cm.AppConfigDir())
	// 凭证服务：系统 Keychain（Linux 无 libsecret 时降级为加密文件）
	credService := credential.New(cm.AppConfigDir())
	oauthService := service.NewOAuthService(credService, cm, settingRepo)

	// 2. Init Services
	tagService := service.NewTagService(tagRepo, postRepo)
	categoryService := service.NewCategoryService(categoryRepo, postRepo)
	postService := service.NewPostService(postRepo, tagRepo, tagService, categoryService, mediaRepo)
	menuService := service.NewMenuService(menuRepo)
	linkService := service.NewLinkService(linkRepo)
	themeService := service.NewThemeService(themeRepo, appDir)
	deployService := service.NewDeployService(settingRepo, appDir)
	// RendererService
	rendererService := engine.New(appDir, postRepo, themeRepo, settingRepo)
	rendererService.SetMenuRepo(menuRepo)
	rendererService.SetLinkRepo(linkRepo)
	rendererService.SetTagRepo(tagRepo)
	rendererService.SetMemoRepo(memoRepo)
	rendererService.SetCategoryRepo(categoryRepo)
	settingService := service.NewSettingService(appDir, settingRepo)
	scaffoldService := service.NewScaffoldService(assets)
	// CommentService
	commentRepo := repository.NewCommentRepository(appDir)
	commentService := service.NewCommentService(appDir, commentRepo, postRepo, themeRepo, settingRepo)
	memoService := service.NewMemoService(memoRepo)
	previewService := service.NewPreviewService(filepath.Join(appDir, "output"))
	// Set CommentRepo on RendererService for template injection
	rendererService.SetCommentRepo(commentRepo)
	rendererService.SetSeoSettingRepo(seoSettingRepo)
	rendererService.SetCdnSettingRepo(cdnSettingRepo)
	rendererService.SetPwaSettingRepo(pwaSettingRepo)
	cdnUploadService := service.NewCdnUploadService(cdnSettingRepo, settingRepo, appDir)
	deployService.SetCdnUploadService(cdnUploadService)
	deployService.SetRenderer(rendererService)
	deployService.SetOAuthService(oauthService)
	// SFTP HostKey TOFU 校验文件位置：app 级目录跨站点共享（#37）
	deployService.SetKnownHostsPath(filepath.Join(cm.AppConfigDir(), "known_hosts"))
	aiService := service.NewAIService(aiSettingRepo, settingRepo, aiUsageRepo)

	// 3. Wrap with Facades
	return &AppServices{
		Category:   NewCategoryFacade(categoryService, postRepo),
		Post:       NewPostFacade(postService),
		Menu:       NewMenuFacade(menuService),
		Link:       NewLinkFacade(linkService),
		Tag:        NewTagFacade(tagService, postRepo),
		Deploy:     NewDeployFacade(deployService),
		Renderer:   NewRendererFacade(rendererService),
		Theme:      NewThemeFacade(themeService),
		Setting:    NewSettingFacade(settingService, oauthService),
		Comment:    NewCommentFacade(commentService),
		Memo:       NewMemoFacade(memoService),
		Preview:    NewPreviewFacade(previewService),
		SeoSetting: NewSeoSettingFacade(seoSettingRepo),
		CdnSetting: NewCdnSettingFacade(cdnSettingRepo),
		PwaSetting: NewPwaSettingFacade(pwaSettingRepo, appDir),
		CdnUpload:  NewCdnUploadFacade(cdnUploadService),
		AI:         NewAIFacade(aiSettingRepo, aiService),
		OAuth:      NewOAuthFacade(oauthService),
		Update:     NewUpdateFacade(),
		Services: struct {
			Category  *service.CategoryService
			Post      *service.PostService
			Menu      *service.MenuService
			Link      *service.LinkService
			Tag       *service.TagService
			Deploy    *service.DeployService
			Renderer  *engine.Engine
			Theme     *service.ThemeService
			Setting   *service.SettingService
			Scaffold  *service.ScaffoldService
			Comment   *service.CommentService
			Memo      *service.MemoService
			Preview   *service.PreviewService
			CdnUpload *service.CdnUploadService
		}{
			Category:  categoryService,
			Post:      postService,
			Menu:      menuService,
			Link:      linkService,
			Tag:       tagService,
			Deploy:    deployService,
			Renderer:  rendererService,
			Theme:     themeService,
			Setting:   settingService,
			Scaffold:  scaffoldService,
			Comment:   commentService,
			Memo:      memoService,
			Preview:   previewService,
			CdnUpload: cdnUploadService,
		},
		Repositories: struct {
			Category domain.CategoryRepository
			Tag      domain.TagRepository
			Post     domain.PostRepository
			Menu     domain.MenuRepository
			Link     domain.LinkRepository
			Memo     domain.MemoRepository
			Setting  domain.SettingRepository
		}{
			Category: categoryRepo,
			Tag:      tagRepo,
			Post:     postRepo,
			Menu:     menuRepo,
			Link:     linkRepo,
			Memo:     memoRepo,
			Setting:  settingRepo,
		},
		assets: assets,
	}
}

// InvalidateAllCaches 清除所有仓库的内存缓存，使下次访问时从磁盘重新加载
func (s *AppServices) InvalidateAllCaches() {
	type invalidatable interface{ Invalidate() }
	repos := []interface{}{
		s.Repositories.Category,
		s.Repositories.Tag,
		s.Repositories.Menu,
		s.Repositories.Link,
		s.Repositories.Memo,
		s.Repositories.Setting,
	}
	for _, r := range repos {
		if inv, ok := r.(invalidatable); ok {
			inv.Invalidate()
		}
	}
	if s.Repositories.Post != nil {
		s.Repositories.Post.Reload(context.Background())
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
	s.SeoSetting.repo = newServices.SeoSetting.repo
	s.CdnSetting.repo = newServices.CdnSetting.repo
	s.PwaSetting.repo = newServices.PwaSetting.repo
	s.CdnUpload.internal = newServices.Services.CdnUpload
	s.AI.repo = newServices.AI.repo
	s.AI.service = newServices.AI.service
	// OAuth 服务是应用级单例（跨站点共享 Keychain 状态），但代理配置要跟随
	// 当前站点——切站时更新其 settingRepo 指针
	s.OAuth.service.SetSettingRepo(newServices.Repositories.Setting)
	// Scaffold service doesn't need update generally, but good to keep in sync
	s.Services.Scaffold = newServices.Services.Scaffold
	s.Services.Comment = newServices.Services.Comment
	s.Services.Memo = newServices.Services.Memo
	s.Services.Preview = newServices.Services.Preview
	s.Services.CdnUpload = newServices.Services.CdnUpload
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
