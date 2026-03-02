package service

import (
	"bytes"
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/template"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/sync/errgroup"
)

// bufferPool optimizes memory usage for large strings
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// ─── 分页辅助函数 ─────────────────────────────────────────────────────────────

// buildPagination 构建分页信息对象
// baseURL 是第 1 页的 URL（如 "/"、"/archives/"、"/tag/Go/"），以 / 结尾
func buildPagination(currentPage, totalPages, totalPosts int, baseURL string) template.PaginationView {
	pv := template.PaginationView{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		TotalPosts:  totalPosts,
		HasPrev:     currentPage > 1,
		HasNext:     currentPage < totalPages,
	}
	if pv.HasPrev {
		if currentPage == 2 {
			pv.PrevURL = baseURL // 第 2 页的上一页是首页(baseURL)
		} else {
			pv.PrevURL = fmt.Sprintf("%spage/%d/", baseURL, currentPage-1)
		}
	}
	if pv.HasNext {
		pv.NextURL = fmt.Sprintf("%spage/%d/", baseURL, currentPage+1)
	}
	return pv
}

// pageSize 返回有效的分页大小，0 或负数时使用 defaultSize
func pageSize(configured, defaultSize int) int {
	if configured <= 0 {
		return defaultSize
	}
	return configured
}

// ─── 通用分页渲染 ─────────────────────────────────────────────────────────────

// paginatedRenderConfig 定义分页渲染的参数
type paginatedRenderConfig struct {
	// templateName 模板名称（如 "index"、"blog"、"archives"）
	templateName string
	// baseURL 第1页的规范 URL（如 "/"、"/post/"），用于构建分页链接
	baseURL string
	// firstPageDir 第1页输出的目录（已包含 buildDir 前缀）
	firstPageDir string
	// pageBaseDir 次页输出的目录前缀（page/2/ 等相对于此路径）
	pageBaseDir string
	// pageSize 每页文章数
	pageSize int
	// items 要分页的文章列表
	items []template.PostView
	// baseData 渲染基础数据（会被 copy，不修改原始数据）
	baseData *template.TemplateData
}

// renderPaginated 提取 renderIndex/renderBlog/renderArchives/renderTagPages 中
// 完全相同的分页核心逻辑：切片 → 构建分页对象 → 渲染 → 写文件。
// 在循环前统一创建第1页目录，循环内只按需创建次页目录（修复问题4）。
func (s *RendererService) renderPaginated(ctx context.Context, cfg paginatedRenderConfig) error {
	total := len(cfg.items)
	totalPages := (total + cfg.pageSize - 1) / cfg.pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	// 第1页目录在循环外预先创建（而非每次循环都调用）
	if err := os.MkdirAll(cfg.firstPageDir, 0755); err != nil {
		return err
	}

	for page := 1; page <= totalPages; page++ {
		// 检查 Context 是否已被取消（支持超时/外部中断，修复问题6）
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		start := (page - 1) * cfg.pageSize
		end := start + cfg.pageSize
		if end > total {
			end = total
		}

		pageData := *cfg.baseData
		if total > 0 {
			pageData.Posts = cfg.items[start:end]
		} else {
			pageData.Posts = nil
		}
		pageData.Pagination = buildPagination(page, totalPages, total, cfg.baseURL)

		html, err := s.renderer.Render(cfg.templateName, &pageData)
		if err != nil {
			return fmt.Errorf("%s 第 %d 页渲染失败: %w", cfg.templateName, page, err)
		}

		// 确定输出路径（次页才创建子目录）
		var outDir string
		if page == 1 {
			outDir = cfg.firstPageDir
		} else {
			outDir = filepath.Join(cfg.pageBaseDir, "page", fmt.Sprintf("%d", page))
			if err := os.MkdirAll(outDir, 0755); err != nil {
				return err
			}
		}

		buf := bufferPool.Get().(*bytes.Buffer)
		buf.Reset()
		buf.WriteString(html)
		writeErr := os.WriteFile(filepath.Join(outDir, FileIndexHTML), buf.Bytes(), 0644)
		bufferPool.Put(buf)
		if writeErr != nil {
			return writeErr
		}
	}
	return nil
}

// ─── 页面渲染函数 ─────────────────────────────────────────────────────────────

// renderIndex 渲染首页（支持分页）
func (s *RendererService) renderIndex(ctx context.Context, buildDir string, data *template.TemplateData) error {
	s.logger.Info(fmt.Sprintf("开始渲染首页，使用 %s 引擎", s.renderer.GetEngineType()))

	listPosts := getVisiblePosts(data.Posts)

	err := s.renderPaginated(ctx, paginatedRenderConfig{
		templateName: "index",
		baseURL:      "/",
		firstPageDir: buildDir,
		pageBaseDir:  buildDir,
		pageSize:     pageSize(data.ThemeConfig.PostPageSize, 10),
		items:        listPosts,
		baseData:     data,
	})
	if err != nil {
		s.logger.Error(fmt.Sprintf("❌ 首页渲染失败: %v，使用简单模板", err))
		return s.renderSimpleIndex(buildDir, data)
	}

	total := len(listPosts)
	totalPages := (total + pageSize(data.ThemeConfig.PostPageSize, 10) - 1) / pageSize(data.ThemeConfig.PostPageSize, 10)
	if totalPages < 1 {
		totalPages = 1
	}
	s.logger.Info(fmt.Sprintf("✅ 首页渲染成功（共 %d 页）", totalPages))
	return nil
}

// renderPost 渲染文章详情页
func (s *RendererService) renderPost(buildDir string, post domain.Post, baseData *template.TemplateData) error {
	// 创建文章专属数据
	postData := *baseData
	postData.Post = s.convertPost(post, domain.ThemeConfig{
		PostPath:   baseData.ThemeConfig.PostPath,
		TagPath:    baseData.ThemeConfig.TagPath,
		DateFormat: baseData.ThemeConfig.DateFormat,
	}, nil) // 单篇渲染不需要分类映射，降级为名称兜底
	postData.SiteTitle = postData.Post.Title + " | " + baseData.ThemeConfig.SiteName

	// 创建目录
	postPath := baseData.ThemeConfig.PostPath
	if postPath == "" {
		postPath = DefaultPostPath
	}
	postDir := filepath.Join(buildDir, postPath, post.FileName)
	if err := os.MkdirAll(postDir, 0755); err != nil {
		return err
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	// 使用新的渲染器接口
	html, err := s.renderer.Render("post", &postData)
	if err != nil {
		s.logger.Error(fmt.Sprintf("文章模板渲染失败: %v，使用简单模板", err))
		return s.renderSimplePost(postDir, &postData)
	}

	buf.WriteString(html)
	indexPath := filepath.Join(postDir, FileIndexHTML)
	if err := os.WriteFile(indexPath, buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

// templateExists 检查主题是否包含指定模板
func (s *RendererService) templateExists(templateName string) bool {
	themePath := filepath.Join(s.appDir, DirThemes)
	// 查找当前主题名称
	entries, err := os.ReadDir(themePath)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		tmplPath := filepath.Join(themePath, entry.Name(), DirTemplates, templateName+".ejs")
		if _, err := os.Stat(tmplPath); err == nil {
			return true
		}
	}
	return false
}

// renderBlog 渲染博客列表页（支持分页）
func (s *RendererService) renderBlog(ctx context.Context, buildDir string, data *template.TemplateData) error {
	// 先用空数据测试模板是否存在
	_, err := s.renderer.Render("blog", data)
	if err != nil {
		s.logger.Error(fmt.Sprintf("博客列表页模板不存在或渲染失败: %v，跳过", err))
		return nil
	}

	postPath := data.ThemeConfig.PostPath
	if postPath == "" {
		postPath = DefaultPostPath
	}

	listPosts := getVisiblePosts(data.Posts)

	blogDir := filepath.Join(buildDir, postPath)
	err = s.renderPaginated(ctx, paginatedRenderConfig{
		templateName: "blog",
		baseURL:      "/" + postPath + "/",
		firstPageDir: blogDir,
		pageBaseDir:  blogDir,
		pageSize:     pageSize(data.ThemeConfig.PostPageSize, 10),
		items:        listPosts,
		baseData:     data,
	})
	if err != nil {
		s.logger.Error(fmt.Sprintf("博客列表页渲染失败: %v，跳过", err))
		return nil
	}

	total := len(listPosts)
	size := pageSize(data.ThemeConfig.PostPageSize, 10)
	totalPages := (total + size - 1) / size
	if totalPages < 1 {
		totalPages = 1
	}
	s.logger.Info(fmt.Sprintf("✅ 博客列表页渲染成功（共 %d 页）", totalPages))
	return nil
}

// renderTags 渲染标签列表页
func (s *RendererService) renderTags(ctx context.Context, buildDir string, data *template.TemplateData, _ domain.ThemeConfig) error {
	// 检查 Context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	html, err := s.renderer.Render("tags", data)
	if err != nil {
		s.logger.Error(fmt.Sprintf("标签列表页模板不存在或渲染失败: %v，跳过", err))
		return nil
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	buf.WriteString(html)

	tagsPath := data.ThemeConfig.TagsPath
	if tagsPath == "" {
		tagsPath = DefaultTagsPath
	}
	tagsDir := filepath.Join(buildDir, tagsPath)
	if err := os.MkdirAll(tagsDir, 0755); err != nil {
		return err
	}

	s.logger.Info("✅ 标签列表页渲染成功")
	return os.WriteFile(filepath.Join(tagsDir, FileIndexHTML), buf.Bytes(), 0644)
}

// renderTagPages 渲染每个标签的文章列表页（支持分页）
func (s *RendererService) renderTagPages(ctx context.Context, buildDir string, data *template.TemplateData, config domain.ThemeConfig) error {
	tagPath := config.TagPath
	if tagPath == "" {
		tagPath = DefaultTagPath
	}

	size := pageSize(data.ThemeConfig.PostPageSize, 10)

	g, tagCtx := errgroup.WithContext(ctx)
	// 限制并发渲染标签页的数量
	g.SetLimit(10)

	for _, tag := range data.Tags {
		tg := tag
		g.Go(func() error {
			// 检查 Context（每个标签循环入口处检查，避免已取消时继续大量渲染）
			select {
			case <-tagCtx.Done():
				return tagCtx.Err()
			default:
			}

			// 筛选该标签下的文章
			var tagPosts []template.PostView
			for _, post := range data.Posts {
				for _, pt := range post.Tags {
					if pt.Name == tg.Name {
						tagPosts = append(tagPosts, post)
						break
					}
				}
			}

			// 构建标签页专属基础数据
			tagBaseData := *data
			tagBaseData.Tag = tg
			tagBaseData.SiteTitle = tg.Name + " | " + data.ThemeConfig.SiteName

			tagDir := filepath.Join(buildDir, tagPath, tg.Name)
			err := s.renderPaginated(tagCtx, paginatedRenderConfig{
				templateName: "tag",
				baseURL:      "/" + tagPath + "/" + tg.Name + "/",
				firstPageDir: tagDir,
				pageBaseDir:  tagDir,
				pageSize:     size,
				items:        tagPosts,
				baseData:     &tagBaseData,
			})
			if err != nil {
				s.logger.Error(fmt.Sprintf("标签 %s 页渲染失败: %v，跳过", tg.Name, err))
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	if len(data.Tags) > 0 {
		s.logger.Info(fmt.Sprintf("✅ 标签页渲染成功（共 %d 个）", len(data.Tags)))
	}
	return nil
}

// renderArchives 渲染归档页（支持分页）
func (s *RendererService) renderArchives(ctx context.Context, buildDir string, data *template.TemplateData) error {
	archivesPath := DefaultArchivesPath

	// 归档页只展示已发布且不隐藏的文章
	listPosts := getVisiblePosts(data.Posts)

	archivesDir := filepath.Join(buildDir, archivesPath)
	err := s.renderPaginated(ctx, paginatedRenderConfig{
		templateName: "archives",
		baseURL:      "/" + archivesPath + "/",
		firstPageDir: archivesDir,
		pageBaseDir:  archivesDir,
		pageSize:     pageSize(data.ThemeConfig.ArchivesPageSize, 10),
		items:        listPosts,
		baseData:     data,
	})
	if err != nil {
		s.logger.Error(fmt.Sprintf("归档页模板不存在或渲染失败: %v，跳过", err))
		return nil
	}

	total := len(listPosts)
	size := pageSize(data.ThemeConfig.ArchivesPageSize, 10)
	totalPages := (total + size - 1) / size
	if totalPages < 1 {
		totalPages = 1
	}
	s.logger.Info(fmt.Sprintf("✅ 归档页渲染成功（共 %d 页）", totalPages))
	return nil
}

// renderFriends 渲染友链页
func (s *RendererService) renderFriends(ctx context.Context, buildDir string, data *template.TemplateData) error {
	// 检查 Context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	html, err := s.renderer.Render("friends", data)
	if err != nil {
		s.logger.Error(fmt.Sprintf("友链页模板不存在或渲染失败: %v，跳过", err))
		return nil
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	buf.WriteString(html)

	linkPath := data.ThemeConfig.LinkPath
	if linkPath == "" {
		linkPath = DefaultLinksPath
	}
	friendsDir := filepath.Join(buildDir, linkPath)
	if err := os.MkdirAll(friendsDir, 0755); err != nil {
		return err
	}

	s.logger.Info("✅ 友链页渲染成功")
	return os.WriteFile(filepath.Join(friendsDir, FileIndexHTML), buf.Bytes(), 0644)
}

// renderMemos 渲染闪念页
func (s *RendererService) renderMemos(ctx context.Context, buildDir string, data *template.TemplateData) error {
	// 检查 Context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	html, err := s.renderer.Render("memos", data)
	if err != nil {
		s.logger.Error(fmt.Sprintf("闪念页模板不存在或渲染失败: %v，跳过", err))
		return nil
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	buf.WriteString(html)

	memosPath := data.ThemeConfig.MemosPath
	if memosPath == "" {
		memosPath = DefaultMemosPath
	}
	memosDir := filepath.Join(buildDir, memosPath)
	if err := os.MkdirAll(memosDir, 0755); err != nil {
		return err
	}

	s.logger.Info("✅ 闪念页渲染成功")
	return os.WriteFile(filepath.Join(memosDir, FileIndexHTML), buf.Bytes(), 0644)
}

// render404 渲染 404 页面
func (s *RendererService) render404(buildDir string, data *template.TemplateData) error {
	// 如果主题没有 404 页面，直接跳过并不会报错，保证旧主题兼容性
	html, err := s.renderer.Render("404", data)
	if err != nil {
		s.logger.Error(fmt.Sprintf("404 页模板不存在或渲染失败: %v，跳过", err))
		return nil
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	buf.WriteString(html)

	s.logger.Info("✅ 404 页面渲染成功")
	// 注意 404 页面通常直接在根目录生成 404.html 文件
	return os.WriteFile(filepath.Join(buildDir, "404.html"), buf.Bytes(), 0644)
}
