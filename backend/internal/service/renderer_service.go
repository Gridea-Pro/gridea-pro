package service

import (
	"context"
	"errors"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/render"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type RendererService struct {
	postRepo     domain.PostRepository
	themeRepo    domain.ThemeRepository
	settingRepo  domain.SettingRepository
	menuRepo     domain.MenuRepository
	commentRepo  domain.CommentRepository
	linkRepo     domain.LinkRepository
	tagRepo      domain.TagRepository
	memoRepo     domain.MemoRepository
	categoryRepo domain.CategoryRepository // 用于渲染时查询分类信息
	appDir       string

	// 主题配置服务
	themeConfigService *ThemeConfigService

	// 资源管理器
	assetManager *AssetManager

	// 主题渲染器(新架构)
	renderer     render.ThemeRenderer
	currentTheme string
}

func NewRendererService(
	appDir string,
	postRepo domain.PostRepository,
	themeRepo domain.ThemeRepository,
	settingRepo domain.SettingRepository,
) *RendererService {
	themeConfigService := NewThemeConfigService(appDir)
	return &RendererService{
		postRepo:           postRepo,
		themeRepo:          themeRepo,
		settingRepo:        settingRepo,
		appDir:             appDir,
		themeConfigService: themeConfigService,
		assetManager:       NewAssetManager(appDir, themeConfigService),
	}
}

// SetMenuRepo 设置菜单仓库（用于获取菜单数据）
func (s *RendererService) SetMenuRepo(menuRepo domain.MenuRepository) {
	s.menuRepo = menuRepo
}

// SetCommentRepo 设置评论仓库（用于获取评论设置）
func (s *RendererService) SetCommentRepo(commentRepo domain.CommentRepository) {
	s.commentRepo = commentRepo
}

// SetLinkRepo 设置友链仓库（用于渲染友链页）
func (s *RendererService) SetLinkRepo(linkRepo domain.LinkRepository) {
	s.linkRepo = linkRepo
}

// SetTagRepo 设置标签仓库
func (s *RendererService) SetTagRepo(tagRepo domain.TagRepository) {
	s.tagRepo = tagRepo
}

// SetMemoRepo 设置闪念仓库（用于渲染闪念页）
func (s *RendererService) SetMemoRepo(memoRepo domain.MemoRepository) {
	s.memoRepo = memoRepo
}

// SetCategoryRepo 设置分类仓库（用于渲染时解析分类信息）
func (s *RendererService) SetCategoryRepo(categoryRepo domain.CategoryRepository) {
	s.categoryRepo = categoryRepo
}

// SetTheme 设置主题并初始化渲染器
func (s *RendererService) SetTheme(themeName string) error {
	// 缓存检查：如果渲染器已初始化且主题未变更，直接返回
	if s.renderer != nil && s.currentTheme == themeName {
		return nil
	}

	factory := render.NewRendererFactory(s.appDir, themeName)
	renderer, err := factory.CreateRenderer()
	if err != nil {
		return fmt.Errorf("创建渲染器失败: %w", err)
	}
	s.renderer = renderer
	s.currentTheme = themeName // 更新当前主题
	_, _ = fmt.Fprintf(os.Stderr, "✅ 使用 %s 引擎渲染主题: %s\n", renderer.GetEngineType(), themeName)
	return nil
}

func (s *RendererService) RenderAll(ctx context.Context) error {
	startTime := time.Now()
	// 获取数据
	posts, _, err := s.postRepo.List(ctx, 1, 10000) // Use List with large page size
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
	// Optimization: Do NOT remove the entire directory.
	// This causes significant performance issues (3s+ delay) on every preview.
	// Overwriting files is sufficient for preview purposes.
	_ = os.MkdirAll(buildDir, 0755)

	var errs error

	// 1. 复制主题资源
	if err := s.assetManager.CopyThemeAssets(buildDir, themeConfig.ThemeName); err != nil {
		errs = errors.Join(errs, fmt.Errorf("复制主题资源失败: %w", err))
		_, _ = fmt.Fprintf(os.Stderr, "警告：复制主题资源失败: %v\n", err)
	}

	// 2. 复制站点静态资源（images、media 等）
	if err := s.assetManager.CopySiteAssets(buildDir); err != nil {
		errs = errors.Join(errs, fmt.Errorf("复制站点资源失败: %w", err))
		_, _ = fmt.Fprintf(os.Stderr, "警告：复制站点资源失败: %v\n", err)
	}

	// 3. 构建模板数据
	templateData, err := s.buildTemplateData(ctx, posts, themeConfig)
	if err != nil {
		return fmt.Errorf("构建模板数据失败: %w", err)
	}

	// 4. 渲染首页
	if err := s.renderIndex(ctx, buildDir, templateData); err != nil {
		errs = errors.Join(errs, fmt.Errorf("渲染首页失败: %w", err))
	}

	// 5. 渲染文章详情页 (Parallel)
	var wg sync.WaitGroup
	var errMutex sync.Mutex
	// Limit concurrency to avoid resource exhaustion
	concurrency := runtime.NumCPU() // Use NumCPU for consistency
	sem := make(chan struct{}, concurrency)

	for _, post := range posts {
		if !post.Published {
			continue
		}

		wg.Add(1)
		go func(p domain.Post) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			if err := s.renderPost(buildDir, p, templateData); err != nil {
				// Thread-safe error collection
				errMutex.Lock()
				errs = errors.Join(errs, fmt.Errorf("rendering post %s: %w", p.Title, err))
				errMutex.Unlock()
				_, _ = fmt.Fprintf(os.Stderr, "Error rendering post %s: %v\n", p.Title, err)
			}
		}(post)
	}
	wg.Wait()

	// 6. 渲染博客列表页
	if err := s.renderBlog(ctx, buildDir, templateData); err != nil {
		errs = errors.Join(errs, fmt.Errorf("渲染博客列表页失败: %w", err))
		_, _ = fmt.Fprintf(os.Stderr, "警告：渲染博客列表页失败: %v\n", err)
	}

	// 7. 渲染标签列表页
	if err := s.renderTags(ctx, buildDir, templateData, themeConfig); err != nil {
		errs = errors.Join(errs, fmt.Errorf("渲染标签页失败: %w", err))
		_, _ = fmt.Fprintf(os.Stderr, "警告：渲染标签页失败: %v\n", err)
	}

	// 8. 渲染每个标签的文章列表页
	if err := s.renderTagPages(ctx, buildDir, templateData, themeConfig); err != nil {
		errs = errors.Join(errs, fmt.Errorf("渲染标签文章页失败: %w", err))
		_, _ = fmt.Fprintf(os.Stderr, "警告：渲染标签文章页失败: %v\n", err)
	}

	// 9. 渲染归档页
	if err := s.renderArchives(ctx, buildDir, templateData); err != nil {
		errs = errors.Join(errs, fmt.Errorf("渲染归档页失败: %w", err))
		_, _ = fmt.Fprintf(os.Stderr, "警告：渲染归档页失败: %v\n", err)
	}

	// 10. 渲染友链页
	if err := s.renderFriends(ctx, buildDir, templateData); err != nil {
		errs = errors.Join(errs, fmt.Errorf("渲染友链页失败: %w", err))
		_, _ = fmt.Fprintf(os.Stderr, "警告：渲染友链页失败: %v\n", err)
	}

	// 11. 渲染闪念页
	if err := s.renderMemos(ctx, buildDir, templateData); err != nil {
		errs = errors.Join(errs, fmt.Errorf("渲染闪念页失败: %w", err))
		_, _ = fmt.Fprintf(os.Stderr, "警告：渲染闪念页失败: %v\n", err)
	}

	// 12. 渲染 404 页
	if err := s.render404(buildDir, templateData); err != nil {
		errs = errors.Join(errs, fmt.Errorf("渲染 404 页面失败: %w", err))
		_, _ = fmt.Fprintf(os.Stderr, "警告：渲染 404 页面失败: %v\n", err)
	}

	// 13. 生成搜索数据 (api/search.json)
	if err := s.renderSearchJSON(buildDir, templateData); err != nil {
		errs = errors.Join(errs, fmt.Errorf("生成搜索数据失败: %w", err))
		_, _ = fmt.Fprintf(os.Stderr, "警告：生成搜索数据失败: %v\n", err)
	}

	// 14. 生成 RSS 订阅 (feed.xml)
	if err := s.renderRSS(buildDir, templateData); err != nil {
		errs = errors.Join(errs, fmt.Errorf("生成 RSS 失败: %w", err))
		_, _ = fmt.Fprintf(os.Stderr, "警告：生成 RSS 失败: %v\n", err)
	}

	// 15. 生成 Sitemap (sitemap.xml)
	if err := s.renderSitemap(buildDir, templateData); err != nil {
		errs = errors.Join(errs, fmt.Errorf("生成 Sitemap 失败: %w", err))
		_, _ = fmt.Fprintf(os.Stderr, "警告：生成 Sitemap 失败: %v\n", err)
	}

	// 16. 生成 Robots.txt
	if err := s.renderRobotsTxt(buildDir, templateData); err != nil {
		errs = errors.Join(errs, fmt.Errorf("生成 robots.txt 失败: %w", err))
		_, _ = fmt.Fprintf(os.Stderr, "警告：生成 robots.txt 失败: %v\n", err)
	}

	totalDuration := time.Since(startTime)
	_, _ = fmt.Fprintf(os.Stderr, "渲染完成，共 %d 篇文章，耗时: %v\n", len(posts), totalDuration)
	return errs
}
