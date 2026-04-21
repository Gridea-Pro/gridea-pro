package engine

import (
	"bytes"
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/render"
	"gridea-pro/backend/internal/template"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

// PageRenderer 负责将模板数据渲染为 HTML 页面并写入文件
type PageRenderer struct {
	renderer      render.ThemeRenderer
	dataBuilder   *TemplateDataBuilder
	appDir        string
	logger        *slog.Logger
	postProcessor *HtmlPostProcessor
	manifest      *RenderManifest
}

// NewPageRenderer 创建 PageRenderer
func NewPageRenderer(appDir string, dataBuilder *TemplateDataBuilder) *PageRenderer {
	return &PageRenderer{
		appDir:      appDir,
		dataBuilder: dataBuilder,
		logger:      slog.Default(),
	}
}

// SetRenderer 设置模板渲染器
func (r *PageRenderer) SetRenderer(renderer render.ThemeRenderer) {
	r.renderer = renderer
}

// SetPostProcessor 设置 HTML 后处理器
func (r *PageRenderer) SetPostProcessor(pp *HtmlPostProcessor) {
	r.postProcessor = pp
}

// SetManifest 设置渲染产物跟踪器
func (r *PageRenderer) SetManifest(m *RenderManifest) {
	r.manifest = m
}

// postProcess 对渲染后的 HTML 进行后处理
func (r *PageRenderer) postProcess(html, pageType, pageURL string, post *template.PostView) string {
	if r.postProcessor == nil {
		return html
	}
	return r.postProcessor.Process(html, pageType, pageURL, post)
}

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
		pv.Prev = pv.PrevURL
	}
	if pv.HasNext {
		pv.NextURL = fmt.Sprintf("%spage/%d/", baseURL, currentPage+1)
		pv.Next = pv.NextURL
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

// renderPaginated 通用分页渲染逻辑
func (r *PageRenderer) renderPaginated(ctx context.Context, cfg paginatedRenderConfig) error {
	total := len(cfg.items)
	totalPages := (total + cfg.pageSize - 1) / cfg.pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	// 第1页目录在循环外预先创建
	if err := os.MkdirAll(cfg.firstPageDir, 0755); err != nil {
		return err
	}

	for page := 1; page <= totalPages; page++ {
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
		// 如果 baseData 预置了按年归档（仅归档页）则按当前页的 posts 重建，避免每页都吐出全量
		if len(cfg.baseData.Archives) > 0 {
			pageData.Archives = buildArchivesByYear(pageData.Posts)
		}
		pageData.Pagination = buildPagination(page, totalPages, total, cfg.baseURL)

		html, err := r.renderer.Render(cfg.templateName, &pageData)
		if err != nil {
			return fmt.Errorf("%s 第 %d 页渲染失败: %w", cfg.templateName, page, err)
		}

		// 后处理：SEO 注入 + CDN URL 重写
		pageURL := cfg.baseURL
		if page > 1 {
			pageURL = fmt.Sprintf("%spage/%d/", cfg.baseURL, page)
		}
		html = r.postProcess(html, cfg.templateName, pageURL, nil)

		var outDir string
		if page == 1 {
			outDir = cfg.firstPageDir
		} else {
			outDir = filepath.Join(cfg.pageBaseDir, "page", fmt.Sprintf("%d", page))
			if err := os.MkdirAll(outDir, 0755); err != nil {
				return err
			}
		}

		// 用 IIFE 限定 buf 的作用域，让 defer bufferPool.Put 在每一页写完后立即触发，
		// panic 安全且与 RenderPost/RenderTags/RenderFriends 等其他位置风格一致。
		if writeErr := func() error {
			buf := bufferPool.Get().(*bytes.Buffer)
			defer bufferPool.Put(buf)
			buf.Reset()
			buf.WriteString(html)
			return r.manifest.WriteFile(filepath.Join(outDir, FileIndexHTML), buf.Bytes(), 0644)
		}(); writeErr != nil {
			return writeErr
		}
	}
	return nil
}

// ─── 页面渲染函数 ─────────────────────────────────────────────────────────────

// RenderIndex 渲染首页（支持分页）
func (r *PageRenderer) RenderIndex(ctx context.Context, buildDir string, data *template.TemplateData) error {
	r.logger.Info(fmt.Sprintf("开始渲染首页，使用 %s 引擎", r.renderer.GetEngineType()))

	listPosts := getVisiblePosts(data.Posts)

	err := r.renderPaginated(ctx, paginatedRenderConfig{
		templateName: "index",
		baseURL:      "/",
		firstPageDir: buildDir,
		pageBaseDir:  buildDir,
		pageSize:     pageSize(data.ThemeConfig.PostPageSize, 10),
		items:        listPosts,
		baseData:     data,
	})
	if err != nil {
		r.logger.Error(fmt.Sprintf("❌ 首页渲染失败: %v，使用简单模板", err))
		return r.renderSimpleIndex(buildDir, data)
	}

	total := len(listPosts)
	totalPages := (total + pageSize(data.ThemeConfig.PostPageSize, 10) - 1) / pageSize(data.ThemeConfig.PostPageSize, 10)
	if totalPages < 1 {
		totalPages = 1
	}
	r.logger.Info(fmt.Sprintf("✅ 首页渲染成功（共 %d 页）", totalPages))
	return nil
}

// RenderPost 渲染文章详情页
func (r *PageRenderer) RenderPost(buildDir string, post domain.Post, baseData *template.TemplateData) error {
	// 创建文章专属数据
	postData := *baseData
	postData.Post = r.dataBuilder.ConvertPost(post, domain.ThemeConfig{
		PostPath:   baseData.ThemeConfig.PostPath,
		TagPath:    baseData.ThemeConfig.TagPath,
		DateFormat: baseData.ThemeConfig.DateFormat,
	}, nil) // categoryByID 传 nil，ConvertPost 自动使用 Build() 阶段缓存的映射
	postData.SiteTitle = postData.Post.Title + " | " + baseData.ThemeConfig.SiteName

	// 检查主题是否有与文章同名的专属模板（如 about.html、privacy.html）
	// 有则使用专属模板渲染，且不显示上一篇/下一篇
	// 排除通用模板名（post、index、blog 等），避免误匹配
	templateName := "post"
	isSpecialPage := false
	commonTemplates := map[string]bool{
		"post": true, "index": true, "blog": true, "tag": true, "tags": true,
		"category": true, "archives": true, "links": true, "memos": true,
		"404": true, "base": true,
	}
	if !commonTemplates[post.FileName] {
		themeName := baseData.ThemeConfig.ThemeName
		for _, ext := range []string{".html", ".ejs", ".gohtml"} {
			tmplPath := filepath.Join(r.appDir, DirThemes, themeName, DirTemplates, post.FileName+ext)
			if _, err := os.Stat(tmplPath); err == nil {
				templateName = post.FileName
				isSpecialPage = true
				break
			}
		}
	}

	// 仅普通文章计算上一篇/下一篇
	if !isSpecialPage {
		visiblePosts := getVisiblePosts(baseData.Posts)
		for i, pv := range visiblePosts {
			if pv.FileName == post.FileName {
				if i > 0 {
					prev := visiblePosts[i-1]
					postData.Post.PrevPost = &template.SimplePostView{
						Title:    prev.Title,
						Link:     prev.Link,
						FileName: prev.FileName,
						Feature:  prev.Feature,
					}
				}
				if i < len(visiblePosts)-1 {
					next := visiblePosts[i+1]
					postData.Post.NextPost = &template.SimplePostView{
						Title:    next.Title,
						Link:     next.Link,
						FileName: next.FileName,
						Feature:  next.Feature,
					}
				}
				break
			}
		}
	}

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

	html, err := r.renderer.Render(templateName, &postData)
	if err != nil {
		r.logger.Error(fmt.Sprintf("文章模板渲染失败: %v，使用简单模板", err))
		return r.renderSimplePost(postDir, &postData)
	}

	// 后处理：SEO 注入 + CDN URL 重写
	postURL := "/" + postPath + "/" + post.FileName + "/"
	html = r.postProcess(html, "post", postURL, &postData.Post)

	buf.WriteString(html)
	indexPath := filepath.Join(postDir, FileIndexHTML)
	if err := r.manifest.WriteFile(indexPath, buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

// RenderBlog 渲染博客列表页（支持分页）
func (r *PageRenderer) RenderBlog(ctx context.Context, buildDir string, data *template.TemplateData) error {
	// 先用空数据测试模板是否存在
	_, err := r.renderer.Render("blog", data)
	if err != nil {
		r.logger.Error(fmt.Sprintf("博客列表页模板不存在或渲染失败: %v，跳过", err))
		return nil
	}

	postPath := data.ThemeConfig.PostPath
	if postPath == "" {
		postPath = DefaultPostPath
	}

	listPosts := getVisiblePosts(data.Posts)

	blogDir := filepath.Join(buildDir, postPath)
	err = r.renderPaginated(ctx, paginatedRenderConfig{
		templateName: "blog",
		baseURL:      "/" + postPath + "/",
		firstPageDir: blogDir,
		pageBaseDir:  blogDir,
		pageSize:     pageSize(data.ThemeConfig.PostPageSize, 10),
		items:        listPosts,
		baseData:     data,
	})
	if err != nil {
		r.logger.Error(fmt.Sprintf("博客列表页渲染失败: %v，跳过", err))
		return nil
	}

	total := len(listPosts)
	size := pageSize(data.ThemeConfig.PostPageSize, 10)
	totalPages := (total + size - 1) / size
	if totalPages < 1 {
		totalPages = 1
	}
	r.logger.Info(fmt.Sprintf("✅ 博客列表页渲染成功（共 %d 页）", totalPages))
	return nil
}

// RenderTags 渲染标签列表页
func (r *PageRenderer) RenderTags(ctx context.Context, buildDir string, data *template.TemplateData, _ domain.ThemeConfig) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	html, err := r.renderer.Render("tags", data)
	if err != nil {
		r.logger.Error(fmt.Sprintf("标签列表页模板不存在或渲染失败: %v，跳过", err))
		return nil
	}

	tagsPath := data.ThemeConfig.TagsPath
	if tagsPath == "" {
		tagsPath = DefaultTagsPath
	}
	html = r.postProcess(html, "tags", "/"+tagsPath+"/", nil)

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	buf.WriteString(html)

	tagsDir := filepath.Join(buildDir, tagsPath)
	if err := os.MkdirAll(tagsDir, 0755); err != nil {
		return err
	}

	r.logger.Info("✅ 标签列表页渲染成功")
	return r.manifest.WriteFile(filepath.Join(tagsDir, FileIndexHTML), buf.Bytes(), 0644)
}

// RenderTagPages 渲染每个标签的文章列表页（支持分页）
func (r *PageRenderer) RenderTagPages(ctx context.Context, buildDir string, data *template.TemplateData, config domain.ThemeConfig) error {
	tagPath := config.TagPath
	if tagPath == "" {
		tagPath = DefaultTagPath
	}

	size := pageSize(data.ThemeConfig.PostPageSize, 10)

	g, tagCtx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	for _, tag := range data.Tags {
		tg := tag
		g.Go(func() error {
			select {
			case <-tagCtx.Done():
				return tagCtx.Err()
			default:
			}

			var tagPosts []template.PostView
			for _, post := range data.Posts {
				for _, pt := range post.Tags {
					if pt.Name == tg.Name {
						tagPosts = append(tagPosts, post)
						break
					}
				}
			}

			tagBaseData := *data
			tagBaseData.Tag = tg
			tagBaseData.SiteTitle = tg.Name + " | " + data.ThemeConfig.SiteName

			tagDir := filepath.Join(buildDir, tagPath, tg.Slug)
			err := r.renderPaginated(tagCtx, paginatedRenderConfig{
				templateName: "tag",
				baseURL:      "/" + tagPath + "/" + tg.Slug + "/",
				firstPageDir: tagDir,
				pageBaseDir:  tagDir,
				pageSize:     size,
				items:        tagPosts,
				baseData:     &tagBaseData,
			})
			if err != nil {
				r.logger.Error(fmt.Sprintf("标签 %s 页渲染失败: %v，跳过", tg.Name, err))
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	if len(data.Tags) > 0 {
		r.logger.Info(fmt.Sprintf("✅ 标签页渲染成功（共 %d 个）", len(data.Tags)))
	}
	return nil
}

// RenderCategoryPages 渲染每个分类的文章列表页（支持分页，复用 tag 模板）
func (r *PageRenderer) RenderCategoryPages(ctx context.Context, buildDir string, data *template.TemplateData) error {
	// 收集所有分类
	categoryPosts := make(map[string][]template.PostView)
	categoryNames := make(map[string]string) // slug -> name
	for _, post := range data.Posts {
		for _, cat := range post.Categories {
			categoryPosts[cat.Slug] = append(categoryPosts[cat.Slug], post)
			categoryNames[cat.Slug] = cat.Name
		}
	}

	if len(categoryPosts) == 0 {
		return nil
	}

	size := pageSize(data.ThemeConfig.PostPageSize, 10)

	g, catCtx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	for slug, posts := range categoryPosts {
		catSlug := slug
		catName := categoryNames[slug]
		catPosts := posts
		g.Go(func() error {
			select {
			case <-catCtx.Done():
				return catCtx.Err()
			default:
			}

			catBaseData := *data
			catBaseData.Category = template.CategoryView{
				Name:  catName,
				Slug:  catSlug,
				Link:  "/" + DefaultCategoryPath + "/" + catSlug + "/",
				Count: len(catPosts),
			}
			catBaseData.SiteTitle = catName + " | " + data.ThemeConfig.SiteName

			catDir := filepath.Join(buildDir, DefaultCategoryPath, catSlug)
			err := r.renderPaginated(catCtx, paginatedRenderConfig{
				templateName: "category",
				baseURL:      "/" + DefaultCategoryPath + "/" + catSlug + "/",
				firstPageDir: catDir,
				pageBaseDir:  catDir,
				pageSize:     size,
				items:        catPosts,
				baseData:     &catBaseData,
			})
			if err != nil {
				r.logger.Error(fmt.Sprintf("分类 %s 页渲染失败: %v，跳过", catName, err))
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	r.logger.Info(fmt.Sprintf("✅ 分类页渲染成功（共 %d 个）", len(categoryPosts)))
	return nil
}

// buildArchivesByYear 将文章列表按年份分组
func buildArchivesByYear(posts []template.PostView) []template.ArchiveYearView {
	yearIndex := make(map[int]int)
	var archives []template.ArchiveYearView
	for _, p := range posts {
		y := p.Date.Year()
		if idx, ok := yearIndex[y]; ok {
			archives[idx].Posts = append(archives[idx].Posts, p)
		} else {
			yearIndex[y] = len(archives)
			archives = append(archives, template.ArchiveYearView{Year: y, Posts: []template.PostView{p}})
		}
	}
	return archives
}

// RenderArchives 渲染归档页（支持分页）
func (r *PageRenderer) RenderArchives(ctx context.Context, buildDir string, data *template.TemplateData) error {
	archivesPath := DefaultArchivesPath

	listPosts := getVisiblePosts(data.Posts)

	// 构建按年份分组的归档数据（供 Go 模板主题使用）
	archivesData := *data
	archivesData.Archives = buildArchivesByYear(listPosts)

	archivesDir := filepath.Join(buildDir, archivesPath)
	err := r.renderPaginated(ctx, paginatedRenderConfig{
		templateName: "archives",
		baseURL:      "/" + archivesPath + "/",
		firstPageDir: archivesDir,
		pageBaseDir:  archivesDir,
		pageSize:     pageSize(data.ThemeConfig.ArchivesPageSize, 10),
		items:        listPosts,
		baseData:     &archivesData,
	})
	if err != nil {
		r.logger.Error(fmt.Sprintf("归档页模板不存在或渲染失败: %v，跳过", err))
		return nil
	}

	total := len(listPosts)
	size := pageSize(data.ThemeConfig.ArchivesPageSize, 10)
	totalPages := (total + size - 1) / size
	if totalPages < 1 {
		totalPages = 1
	}
	r.logger.Info(fmt.Sprintf("✅ 归档页渲染成功（共 %d 页）", totalPages))
	return nil
}

// RenderFriends 渲染友链页
func (r *PageRenderer) RenderFriends(ctx context.Context, buildDir string, data *template.TemplateData) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	html, err := r.renderer.Render("links", data)
	if err != nil {
		r.logger.Error(fmt.Sprintf("友链页模板不存在或渲染失败: %v，跳过", err))
		return nil
	}

	linkPath := data.ThemeConfig.LinkPath
	if linkPath == "" {
		linkPath = DefaultLinksPath
	}
	html = r.postProcess(html, "links", "/"+linkPath+"/", nil)

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	buf.WriteString(html)

	friendsDir := filepath.Join(buildDir, linkPath)
	if err := os.MkdirAll(friendsDir, 0755); err != nil {
		return err
	}

	r.logger.Info("✅ 友链页渲染成功")
	return r.manifest.WriteFile(filepath.Join(friendsDir, FileIndexHTML), buf.Bytes(), 0644)
}

// RenderMemos 渲染闪念页
func (r *PageRenderer) RenderMemos(ctx context.Context, buildDir string, data *template.TemplateData) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	html, err := r.renderer.Render("memos", data)
	if err != nil {
		r.logger.Error(fmt.Sprintf("闪念页模板不存在或渲染失败: %v，跳过", err))
		return nil
	}

	memosPath := data.ThemeConfig.MemosPath
	if memosPath == "" {
		memosPath = DefaultMemosPath
	}
	html = r.postProcess(html, "memos", "/"+memosPath+"/", nil)

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	buf.WriteString(html)

	memosDir := filepath.Join(buildDir, memosPath)
	if err := os.MkdirAll(memosDir, 0755); err != nil {
		return err
	}

	r.logger.Info("✅ 闪念页渲染成功")
	return r.manifest.WriteFile(filepath.Join(memosDir, FileIndexHTML), buf.Bytes(), 0644)
}

// Render404 渲染 404 页面
func (r *PageRenderer) Render404(buildDir string, data *template.TemplateData) error {
	html, err := r.renderer.Render("404", data)
	if err != nil {
		r.logger.Error(fmt.Sprintf("404 页模板不存在或渲染失败: %v，跳过", err))
		return nil
	}

	html = r.postProcess(html, "404", "/404.html", nil)

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	buf.WriteString(html)

	r.logger.Info("✅ 404 页面渲染成功")
	return r.manifest.WriteFile(filepath.Join(buildDir, "404.html"), buf.Bytes(), 0644)
}

// renderSimpleIndex 渲染简单首页（备用）
func (r *PageRenderer) renderSimpleIndex(buildDir string, data *template.TemplateData) error {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	var postListHTML strings.Builder
	for _, p := range data.Posts {
		postListHTML.WriteString(fmt.Sprintf(`
			<article class="post">
				<h2 class="post-title"><a href="%s">%s</a></h2>
				<div class="post-meta">%s</div>
			</article>
		`, p.Link, p.Title, p.DateFormat))
	}

	fmt.Fprintf(buf, `<!DOCTYPE html>
<html lang="zh-CN">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s</title>
	<link rel="stylesheet" href="/styles/main.css">
	<style>
		body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; line-height: 1.6; max-width: 800px; margin: 0 auto; padding: 20px; }
		.site-header { text-align: center; padding: 40px 0; border-bottom: 1px solid #eee; }
		.site-title { font-size: 2em; margin: 0; }
		.site-description { color: #666; margin-top: 10px; }
		.post { margin: 40px 0; padding-bottom: 20px; border-bottom: 1px solid #eee; }
		.post-title a { color: #333; text-decoration: none; }
		.post-title a:hover { color: #0066cc; }
		.post-meta { color: #999; font-size: 0.9em; margin-top: 5px; }
	</style>
</head>
<body>
	<header class="site-header">
		<h1 class="site-title">%s</h1>
		<p class="site-description">%s</p>
	</header>
	<main class="site-main">%s</main>
	<footer style="text-align: center; padding: 40px 0; color: #999;">%s</footer>
</body>
</html>`, data.ThemeConfig.SiteName, data.ThemeConfig.SiteName, data.ThemeConfig.SiteDescription,
		postListHTML.String(), data.ThemeConfig.FooterInfo)

	return r.manifest.WriteFile(filepath.Join(buildDir, FileIndexHTML), buf.Bytes(), 0644)
}

// renderSimplePost 渲染简单文章页（备用）
func (r *PageRenderer) renderSimplePost(postDir string, data *template.TemplateData) error {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	fmt.Fprintf(buf, `<!DOCTYPE html>
<html lang="zh-CN">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s</title>
	<link rel="stylesheet" href="/styles/main.css">
	<style>
		body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; line-height: 1.6; max-width: 800px; margin: 0 auto; padding: 20px; }
		.post-header { text-align: center; padding: 40px 0; }
		.post-title { font-size: 2.5em; margin: 0; }
		.post-meta { color: #999; margin-top: 10px; }
		.post-content { margin-top: 40px; }
		.post-content img { max-width: 100%%; height: auto; }
		.back-link { display: inline-block; margin-top: 40px; color: #0066cc; text-decoration: none; }
	</style>
</head>
<body>
	<article class="post">
		<header class="post-header">
			<h1 class="post-title">%s</h1>
			<div class="post-meta">%s</div>
		</header>
		<div class="post-content">%s</div>
	</article>
	<a href="/" class="back-link">← 返回首页</a>
	<footer style="text-align: center; padding: 40px 0; color: #999;">%s</footer>
</body>
</html>`, data.SiteTitle, data.Post.Title, data.Post.DateFormat, data.Post.Content, data.ThemeConfig.FooterInfo)

	indexPath := filepath.Join(postDir, FileIndexHTML)
	if err := r.manifest.WriteFile(indexPath, buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}
