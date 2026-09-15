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
	"log/slog"
	"path/filepath"
	"sync"
)

// WailsContext holds the global application context
var WailsContext context.Context

type AppServices struct {
	mu                   sync.RWMutex
	Category             *CategoryFacade
	Post                 *PostFacade
	Menu                 *MenuFacade
	Link                 *LinkFacade
	Tag                  *TagFacade
	Deploy               *DeployFacade
	Renderer             *RendererFacade
	Theme                *ThemeFacade
	Setting              *SettingFacade
	Comment              *CommentFacade
	Memo                 *MemoFacade
	Preview              *PreviewFacade
	SeoSetting           *SeoSettingFacade
	CdnSetting           *CdnSettingFacade
	PwaSetting           *PwaSettingFacade
	CdnUpload            *CdnUploadFacade
	AI                   *AIFacade
	OAuth                *OAuthFacade
	Update               *UpdateFacade
	ImageHosting         *ImageHostingFacade
	ImageOptimizeSetting *ImageOptimizeSettingFacade
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
	imageOptimizeSettingRepo := repository.NewImageOptimizeSettingRepository(appDir)
	postRepo := repository.NewPostRepository(appDir, imageOptimizeSettingRepo)
	categoryRepo := repository.NewCategoryRepository(appDir)
	tagRepo := repository.NewTagRepository(appDir)
	menuRepo := repository.NewMenuRepository(appDir)
	linkRepo := repository.NewLinkRepository(appDir)
	themeRepo := repository.NewThemeRepository(appDir)
	settingRepo := repository.NewSettingRepository(appDir)
	mediaRepo := repository.NewMediaRepository(appDir, imageOptimizeSettingRepo)
	memoRepo := repository.NewMemoRepository(appDir)
	seoSettingRepo := repository.NewSeoSettingRepository(appDir)
	cdnSettingRepo := repository.NewCdnSettingRepository(appDir)
	pwaSettingRepo := repository.NewPwaSettingRepository(appDir)

	// 1.1 扫描历史数据中的 Name / Slug 重复（通常是用户手工编辑 JSON 引入的），
	//     仅记录告警，不阻止启动 —— 数据层 Create/Update 会拒绝新的冲突。
	auditCtx := context.Background()
	for _, c := range repository.AuditTagUniqueness(auditCtx, tagRepo) {
		slog.Warn("检测到标签数据重复，渲染层可能出现覆盖，请修复 config/tags.json", "conflict", c)
	}
	for _, c := range repository.AuditCategoryUniqueness(auditCtx, categoryRepo) {
		slog.Warn("检测到分类数据重复，渲染层可能出现覆盖，请修复 config/categories.json", "conflict", c)
	}
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

	// 2.5 Image Hosting Service (S.EE)
	imageHostingRepo := repository.NewImageHostingRepository(appDir)
	imageHostingService := service.NewImageHostingService(imageHostingRepo)

	// 3. Wrap with Facades
	return &AppServices{
		Category:             NewCategoryFacade(categoryService, postRepo),
		Post:                 NewPostFacade(postService),
		Menu:                 NewMenuFacade(menuService),
		Link:                 NewLinkFacade(linkService),
		Tag:                  NewTagFacade(tagService, postRepo),
		Deploy:               NewDeployFacade(deployService),
		Renderer:             NewRendererFacade(rendererService),
		Theme:                NewThemeFacade(themeService),
		Setting:              NewSettingFacade(settingService, oauthService),
		Comment:              NewCommentFacade(commentService),
		Memo:                 NewMemoFacade(memoService),
		Preview:              NewPreviewFacade(previewService),
		SeoSetting:           NewSeoSettingFacade(seoSettingRepo),
		CdnSetting:           NewCdnSettingFacade(cdnSettingRepo),
		PwaSetting:           NewPwaSettingFacade(pwaSettingRepo, appDir),
		CdnUpload:            NewCdnUploadFacade(cdnUploadService),
		AI:                   NewAIFacade(aiSettingRepo, aiService),
		OAuth:                NewOAuthFacade(oauthService),
		Update:               NewUpdateFacade(),
		ImageHosting:         NewImageHostingFacade(imageHostingService),
		ImageOptimizeSetting: NewImageOptimizeSettingFacade(imageOptimizeSettingRepo),
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
	// Repositories.* 是普通字段，切站（UpdateAppDir 持 s.mu 写锁）会整体替换它们。
	// 这里先在 RLock 下取一份一致快照再操作，避免与切站并发时读到"替换一半"的指针集
	// （本函数从"刷新/重载"路径调用，与用户切站可能并发）。
	s.mu.RLock()
	category := s.Repositories.Category
	tag := s.Repositories.Tag
	menu := s.Repositories.Menu
	link := s.Repositories.Link
	memo := s.Repositories.Memo
	setting := s.Repositories.Setting
	post := s.Repositories.Post
	s.mu.RUnlock()

	type invalidatable interface{ Invalidate() }
	repos := []interface{}{category, tag, menu, link, memo, setting}
	for _, r := range repos {
		if inv, ok := r.(invalidatable); ok {
			inv.Invalidate()
		}
	}
	if post != nil {
		// Post repo 走的是 Reload（同步扫盘）而非 Invalidate。
		// 历史代码这里直接丢错——issue #107 撞到时连"posts 加载失败"这条线索都没有。
		// 至少 log 一笔，配合 Fix 1 的 partial scan / "all failed" 错误能精确定位。
		if err := post.Reload(context.Background()); err != nil {
			slog.Error("InvalidateAllCaches: Post.Reload failed",
				"error", err)
		}
	}
}

func (s *AppServices) UpdateAppDir(appDir string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-initialize logic
	newServices := NewAppServices(appDir, s.assets)
	// 各 facade 的 internal/repo/service 字段已改为原子读写（消除切站热替换与 Wails 并发调用的
	// data race），这里统一用 .Store 写入。从 newServices 取值：service 类走 newServices.Services.*
	// 原始指针；repo/appDir 类走对应 facade 的私有 getter。
	s.Category.internal.Store(newServices.Services.Category)
	s.Post.internal.Store(newServices.Services.Post)
	s.Menu.internal.Store(newServices.Services.Menu)
	s.Link.internal.Store(newServices.Services.Link)
	s.Tag.internal.Store(newServices.Services.Tag)
	s.Deploy.internal.Store(newServices.Services.Deploy)
	s.Renderer.internal.Store(newServices.Services.Renderer)
	s.Theme.internal.Store(newServices.Services.Theme)
	s.Setting.internal.Store(newServices.Services.Setting)
	s.Comment.internal.Store(newServices.Services.Comment)
	s.Memo.internal.Store(newServices.Services.Memo)
	s.Preview.internal.Store(newServices.Services.Preview)
	s.SeoSetting.repo.Store(newServices.SeoSetting.repository())
	s.CdnSetting.repo.Store(newServices.CdnSetting.repository())
	s.PwaSetting.repo.Store(newServices.PwaSetting.repository())
	s.CdnUpload.internal.Store(newServices.Services.CdnUpload)
	s.AI.repo.Store(newServices.AI.repository())
	s.AI.service.Store(newServices.AI.svc())
	// OAuth 服务是应用级单例（跨站点共享 Keychain 状态），但代理配置要跟随
	// 当前站点——切站时更新其 settingRepo 指针
	s.OAuth.service.SetSettingRepo(newServices.Repositories.Setting)
	// ImageHosting：就地更新底层 service 指针，而非替换 facade 对象——Wails 绑定的是 boot 时的
	// 原 facade 对象，替换字段无效会导致切站后图床读写仍打旧站点目录。
	s.ImageHosting.internal.Store(newServices.ImageHosting.svc())
	s.ImageOptimizeSetting.repo.Store(newServices.ImageOptimizeSetting.repository())
	// Scaffold service doesn't need update generally, but good to keep in sync
	s.Services.Scaffold = newServices.Services.Scaffold
	s.Services.Comment = newServices.Services.Comment
	s.Services.Memo = newServices.Services.Memo
	s.Services.Preview = newServices.Services.Preview
	s.Services.CdnUpload = newServices.Services.CdnUpload

	// 补全切站时遗漏的字段（否则切站后出现跨站点数据串号）：
	// 1) Repositories.* —— InvalidateAllCaches 只操作这些指针；不更新会导致"刷新/重载"
	//    清的是旧站点的缓存，对当前站点无效。
	s.Repositories.Category = newServices.Repositories.Category
	s.Repositories.Tag = newServices.Repositories.Tag
	s.Repositories.Post = newServices.Repositories.Post
	s.Repositories.Menu = newServices.Repositories.Menu
	s.Repositories.Link = newServices.Repositories.Link
	s.Repositories.Memo = newServices.Repositories.Memo
	s.Repositories.Setting = newServices.Repositories.Setting
	// 2) Category/Tag Facade 各自持有的 postRepo —— 保存/删除分类标签时用它返回文章列表，
	//    不更新会把旧站点的文章列表混进当前站点的返回结果。
	s.Category.postRepo.Store(newServices.Repositories.Post)
	s.Tag.postRepo.Store(newServices.Repositories.Post)
	// 3) PwaSetting Facade 的 appDir —— PWA 图标检测/保存/删除用它拼路径，
	//    不更新会把新站点的图标写进旧站点 images/ 目录。
	s.PwaSetting.appDir.Store(appDir)
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
