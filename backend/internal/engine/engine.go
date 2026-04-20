package engine

import (
	"context"
	"errors"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/render"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// Engine 渲染协调器，组合各个独立的渲染子模块
type Engine struct {
	appDir string

	// 子模块
	dataBuilder   *TemplateDataBuilder
	pageRenderer  *PageRenderer
	seoGenerator  *SeoGenerator
	pwaGenerator  *PwaGenerator
	searchBuilder *SearchIndexBuilder
	assetManager  *AssetManager

	// 主题配置
	themeConfigService *ThemeConfigService

	// 仓库引用（仅用于 RenderAll 获取数据）
	postRepo       domain.PostRepository
	themeRepo      domain.ThemeRepository
	seoSettingRepo domain.SeoSettingRepository
	cdnSettingRepo domain.CdnSettingRepository
	pwaSettingRepo domain.PwaSettingRepository

	// 主题渲染器
	renderer     render.ThemeRenderer
	currentTheme string
	logger       *slog.Logger
}

func New(
	appDir string,
	postRepo domain.PostRepository,
	themeRepo domain.ThemeRepository,
	settingRepo domain.SettingRepository,
) *Engine {
	themeConfigService := NewThemeConfigService(appDir)
	dataBuilder := NewTemplateDataBuilder(postRepo, themeRepo, settingRepo, themeConfigService)
	pageRenderer := NewPageRenderer(appDir, dataBuilder)

	return &Engine{
		appDir:             appDir,
		postRepo:           postRepo,
		themeRepo:          themeRepo,
		dataBuilder:        dataBuilder,
		pageRenderer:       pageRenderer,
		seoGenerator:       NewSeoGenerator(),
		pwaGenerator:       NewPwaGenerator(appDir),
		searchBuilder:      NewSearchIndexBuilder(),
		assetManager:       NewAssetManager(appDir, themeConfigService),
		themeConfigService: themeConfigService,
		logger:             slog.Default(),
	}
}

// SetMenuRepo 设置菜单仓库
func (s *Engine) SetMenuRepo(menuRepo domain.MenuRepository) {
	s.dataBuilder.SetMenuRepo(menuRepo)
}

// SetCommentRepo 设置评论仓库
func (s *Engine) SetCommentRepo(commentRepo domain.CommentRepository) {
	s.dataBuilder.SetCommentRepo(commentRepo)
}

// SetLinkRepo 设置友链仓库
func (s *Engine) SetLinkRepo(linkRepo domain.LinkRepository) {
	s.dataBuilder.SetLinkRepo(linkRepo)
}

// SetTagRepo 设置标签仓库
func (s *Engine) SetTagRepo(tagRepo domain.TagRepository) {
	s.dataBuilder.SetTagRepo(tagRepo)
}

// SetMemoRepo 设置闪念仓库
func (s *Engine) SetMemoRepo(memoRepo domain.MemoRepository) {
	s.dataBuilder.SetMemoRepo(memoRepo)
}

// SetCategoryRepo 设置分类仓库
func (s *Engine) SetCategoryRepo(categoryRepo domain.CategoryRepository) {
	s.dataBuilder.SetCategoryRepo(categoryRepo)
}

// SetSeoSettingRepo 设置 SEO 设置仓库
func (s *Engine) SetSeoSettingRepo(repo domain.SeoSettingRepository) {
	s.seoSettingRepo = repo
}

// SetCdnSettingRepo 设置 CDN 设置仓库
func (s *Engine) SetCdnSettingRepo(repo domain.CdnSettingRepository) {
	s.cdnSettingRepo = repo
}

// SetPwaSettingRepo 设置 PWA 设置仓库
func (s *Engine) SetPwaSettingRepo(repo domain.PwaSettingRepository) {
	s.pwaSettingRepo = repo
}

// SetTheme 设置主题并初始化渲染器
func (s *Engine) SetTheme(themeName string) error {
	factory := render.NewRendererFactory(s.appDir, themeName)

	// 检测当前主题的引擎类型
	engineType, err := factory.GetEngineType()
	if err != nil {
		return fmt.Errorf("检测引擎类型失败: %w", err)
	}

	// 缓存检查：主题未变更 且 引擎类型一致时，清除缓存后返回
	if s.renderer != nil && s.currentTheme == themeName && s.renderer.GetEngineType() == engineType {
		s.renderer.ClearCache()
		return nil
	}

	renderer, err := factory.CreateRenderer()
	if err != nil {
		return fmt.Errorf("创建渲染器失败: %w", err)
	}
	s.renderer = renderer
	s.currentTheme = themeName
	s.pageRenderer.SetRenderer(renderer)
	s.logger.Info(fmt.Sprintf("✅ 使用 %s 引擎渲染主题: %s", renderer.GetEngineType(), themeName))
	return nil
}

func (s *Engine) RenderAll(ctx context.Context) error {
	startTime := time.Now()
	// 获取数据
	posts, _, err := s.postRepo.List(ctx, 1, 10000)
	if err != nil {
		return fmt.Errorf("获取文章失败: %w", err)
	}

	themeConfig, err := s.themeRepo.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("获取主题配置失败: %w", err)
	}

	// 初始化渲染器
	if err := s.SetTheme(themeConfig.ThemeName); err != nil {
		return fmt.Errorf("初始化渲染器失败: %w", err)
	}

	buildDir := filepath.Join(s.appDir, DirOutput)

	// 清理旧的输出文件，避免已删除的文章残留 HTML
	_ = os.RemoveAll(buildDir)
	_ = os.MkdirAll(buildDir, 0755)

	var errs error

	// 1. 复制主题资源
	if err := s.assetManager.CopyThemeAssets(buildDir, themeConfig.ThemeName); err != nil {
		errs = errors.Join(errs, fmt.Errorf("复制主题资源失败: %w", err))
		s.logger.Error(fmt.Sprintf("警告：复制主题资源失败: %v", err))
	}

	// 2. 复制站点静态资源（images、media 等）
	if err := s.assetManager.CopySiteAssets(buildDir); err != nil {
		errs = errors.Join(errs, fmt.Errorf("复制站点资源失败: %w", err))
		s.logger.Error(fmt.Sprintf("警告：复制站点资源失败: %v", err))
	}

	// 3. 构建模板数据
	templateData, err := s.dataBuilder.Build(ctx, posts, themeConfig)
	if err != nil {
		return fmt.Errorf("构建模板数据失败: %w", err)
	}

	// 4. 初始化 HTML 后处理器（SEO + CDN + PWA）
	var postProcessor *HtmlPostProcessor
	var pwaSetting domain.PwaSetting
	{
		var seoSetting domain.SeoSetting
		var cdnSetting domain.CdnSetting
		if s.seoSettingRepo != nil {
			seoSetting, _ = s.seoSettingRepo.GetSeoSetting(ctx)
		}
		if s.cdnSettingRepo != nil {
			cdnSetting, _ = s.cdnSettingRepo.GetCdnSetting(ctx)
		}
		if s.pwaSettingRepo != nil {
			pwaSetting, _ = s.pwaSettingRepo.GetPwaSetting(ctx)
		}
		postProcessor = NewHtmlPostProcessor(
			&seoSetting, &cdnSetting, &pwaSetting,
			templateData.ThemeConfig.Domain,
			templateData.ThemeConfig.SiteName,
			templateData.ThemeConfig.SiteDescription,
			templateData.ThemeConfig.Language,
		)
		s.pageRenderer.SetPostProcessor(postProcessor)
	}

	type renderTask struct {
		name string
		fn   func() error
	}

	// 核心业务：列表类页面渲染
	tasks := []renderTask{
		{"首页", func() error { return s.pageRenderer.RenderIndex(ctx, buildDir, templateData) }},
		{"博客列表页", func() error { return s.pageRenderer.RenderBlog(ctx, buildDir, templateData) }},
		{"标签页", func() error { return s.pageRenderer.RenderTags(ctx, buildDir, templateData, themeConfig) }},
		{"归档页", func() error { return s.pageRenderer.RenderArchives(ctx, buildDir, templateData) }},
		{"标签文章页", func() error { return s.pageRenderer.RenderTagPages(ctx, buildDir, templateData, themeConfig) }},
		{"分类文章页", func() error { return s.pageRenderer.RenderCategoryPages(ctx, buildDir, templateData) }},
	}

	for _, task := range tasks {
		if err := task.fn(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("%s失败: %w", task.name, err))
			s.logger.Error(fmt.Sprintf("警告：%s失败: %v", task.name, err))
		}
	}

	// 渲染文章详情页 (并发)
	g, _ := errgroup.WithContext(ctx)
	g.SetLimit(runtime.NumCPU())

	for _, post := range posts {
		if !post.Published {
			continue
		}
		p := post
		g.Go(func() error {
			if err := s.pageRenderer.RenderPost(buildDir, p, templateData); err != nil {
				s.logger.Error(fmt.Sprintf("rendering post %s: %v", p.Title, err))
				return fmt.Errorf("rendering post %s: %w", p.Title, err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		errs = errors.Join(errs, err)
	}

	// 完全独立、无依赖的页面与元数据生成，运用 errgroup 并发执行
	asyncTasks := []renderTask{
		{"友链页", func() error { return s.pageRenderer.RenderFriends(ctx, buildDir, templateData) }},
		{"闪念页", func() error { return s.pageRenderer.RenderMemos(ctx, buildDir, templateData) }},
		{"404页面", func() error { return s.pageRenderer.Render404(buildDir, templateData) }},
		{"搜索数据(search.json)", func() error { return s.searchBuilder.RenderSearchJSON(buildDir, templateData) }},
		{"RSS订阅(feed.xml)", func() error { return s.seoGenerator.RenderRSS(buildDir, templateData) }},
		{"站点地图(sitemap.xml)", func() error { return s.seoGenerator.RenderSitemap(buildDir, templateData) }},
		{"Robots(robots.txt)", func() error { return s.seoGenerator.RenderRobotsTxt(buildDir, templateData) }},
		{"PWA Manifest(manifest.json)", func() error {
			if pwaSetting.Enabled {
				return s.pwaGenerator.RenderManifest(buildDir, &pwaSetting, templateData.ThemeConfig.SiteName, templateData.ThemeConfig.Language)
			}
			return nil
		}},
		{"PWA ServiceWorker(sw.js)", func() error {
			if pwaSetting.Enabled {
				return s.pwaGenerator.RenderServiceWorker(buildDir)
			}
			return nil
		}},
	}

	asyncGroup, asyncCtx := errgroup.WithContext(ctx)
	asyncGroup.SetLimit(10)

	var asyncErrs error
	var errsMu sync.Mutex

	for _, task := range asyncTasks {
		t := task
		asyncGroup.Go(func() error {
			select {
			case <-asyncCtx.Done():
				return asyncCtx.Err()
			default:
			}
			if err := t.fn(); err != nil {
				s.logger.Error(fmt.Sprintf("警告：%s并发生成失败: %v", t.name, err))
				errsMu.Lock()
				asyncErrs = errors.Join(asyncErrs, fmt.Errorf("%s失败: %w", t.name, err))
				errsMu.Unlock()
			}
			return nil
		})
	}

	if err := asyncGroup.Wait(); err != nil {
		errs = errors.Join(errs, err)
	}
	if asyncErrs != nil {
		errs = errors.Join(errs, asyncErrs)
	}

	// CSS 合并压缩（后处理）
	themePath := filepath.Join(s.appDir, DirThemes, themeConfig.ThemeName)
	if err := s.assetManager.BundleCSS(buildDir, themePath); err != nil {
		s.logger.Warn("CSS bundle 失败，使用原始文件", "error", err)
	}

	totalDuration := time.Since(startTime)
	s.logger.Info(fmt.Sprintf("渲染完成，共 %d 篇文章，耗时: %v", len(posts), totalDuration))
	return errs
}
