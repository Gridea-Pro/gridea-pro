package service

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/template"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// bufferPool optimizes memory usage for large strings
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// ─── 文件级别 Struct 定义（RSS / Sitemap）────────────────────────────────────

// CDATA 安全的原始 HTML 输出结构
type CDATA struct {
	Text string `xml:",cdata"`
}

// rssEnclosure RSS 2.0 附件（图片/音频/视频）
type rssEnclosure struct {
	XMLName xml.Name `xml:"enclosure"`
	URL     string   `xml:"url,attr"`
	Length  string   `xml:"length,attr"`
	Type    string   `xml:"type,attr"`
}

// rssGuid RSS 2.0 唯一标识
type rssGuid struct {
	XMLName     xml.Name `xml:"guid"`
	IsPermaLink bool     `xml:"isPermaLink,attr"`
	Value       string   `xml:",chardata"`
}

// rssAtomLink Atom 自引用链接
type rssAtomLink struct {
	XMLName xml.Name `xml:"atom:link"`
	Href    string   `xml:"href,attr"`
	Rel     string   `xml:"rel,attr"`
	Type    string   `xml:"type,attr"`
}

// rssItem RSS 2.0 条目
type rssItem struct {
	XMLName     xml.Name      `xml:"item"`
	Title       string        `xml:"title"`
	Link        string        `xml:"link"`
	Guid        rssGuid       `xml:"guid"`
	PubDate     string        `xml:"pubDate"`
	Description CDATA         `xml:"description"`
	Categories  []string      `xml:"category,omitempty"`
	Enclosure   *rssEnclosure `xml:"enclosure,omitempty"`
}

// rssChannel RSS 2.0 频道
type rssChannel struct {
	XMLName       xml.Name    `xml:"channel"`
	Title         string      `xml:"title"`
	Link          string      `xml:"link"`
	Description   string      `xml:"description"`
	Language      string      `xml:"language"`
	Generator     string      `xml:"generator"`
	LastBuildDate string      `xml:"lastBuildDate"`
	AtomLink      rssAtomLink `xml:"atom:link"`
	Items         []rssItem   `xml:"item"`
}

// rssFeed RSS 2.0 根元素
type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	Channel rssChannel `xml:"channel"`
}

// sitemapImage Sitemap 图片扩展
type sitemapImage struct {
	XMLName xml.Name `xml:"image:image"`
	Loc     string   `xml:"image:loc"`
}

// sitemapURL Sitemap 链接条目
type sitemapURL struct {
	XMLName xml.Name      `xml:"url"`
	Loc     string        `xml:"loc"`
	LastMod string        `xml:"lastmod"`
	Image   *sitemapImage `xml:"image:image,omitempty"`
}

// sitemapURLSet Sitemap 根元素
type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	ImageNs string       `xml:"xmlns:image,attr"`
	Urls    []sitemapURL `xml:"url"`
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
	_, _ = fmt.Fprintf(os.Stderr, "开始渲染首页，使用 %s 引擎\n", s.renderer.GetEngineType())

	var listPosts []template.PostView
	for _, p := range data.Posts {
		if !p.HideInList {
			listPosts = append(listPosts, p)
		}
	}

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
		_, _ = fmt.Fprintf(os.Stderr, "❌ 首页渲染失败: %v，使用简单模板\n", err)
		return s.renderSimpleIndex(buildDir, data)
	}

	total := len(listPosts)
	totalPages := (total + pageSize(data.ThemeConfig.PostPageSize, 10) - 1) / pageSize(data.ThemeConfig.PostPageSize, 10)
	if totalPages < 1 {
		totalPages = 1
	}
	_, _ = fmt.Fprintf(os.Stderr, "✅ 首页渲染成功（共 %d 页）\n", totalPages)
	return nil
}

// renderSimpleIndex 渲染简单首页（备用）
func (s *RendererService) renderSimpleIndex(buildDir string, data *template.TemplateData) error {
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

	// Use buffer to construct the final HTML to avoid huge string allocation
	// Note: We are still formatting string key parts.
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

	return os.WriteFile(filepath.Join(buildDir, FileIndexHTML), buf.Bytes(), 0644)
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
		_, _ = fmt.Fprintf(os.Stderr, "文章模板渲染失败: %v，使用简单模板\n", err)
		return s.renderSimplePost(postDir, &postData)
	}

	buf.WriteString(html)
	indexPath := filepath.Join(postDir, FileIndexHTML)
	if err := os.WriteFile(indexPath, buf.Bytes(), 0644); err != nil {
		// Retry once: maybe dir is missing?
		if os.IsNotExist(err) {
			if mkdirErr := os.MkdirAll(postDir, 0755); mkdirErr != nil {
				return fmt.Errorf("failed to retry create dir: %w, original write error: %v", mkdirErr, err)
			}
			return os.WriteFile(indexPath, buf.Bytes(), 0644)
		}
		return err
	}

	return nil
}

// renderSimplePost 渲染简单文章页（备用）
func (s *RendererService) renderSimplePost(postDir string, data *template.TemplateData) error {
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

	// Write file with retry
	indexPath := filepath.Join(postDir, FileIndexHTML)
	if err := os.WriteFile(indexPath, buf.Bytes(), 0644); err != nil {
		// Retry once: maybe dir is missing?
		if os.IsNotExist(err) {
			if mkdirErr := os.MkdirAll(postDir, 0755); mkdirErr != nil {
				return fmt.Errorf("failed to retry create dir: %w, original write error: %v", mkdirErr, err)
			}
			return os.WriteFile(indexPath, buf.Bytes(), 0644)
		}
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
		_, _ = fmt.Fprintf(os.Stderr, "博客列表页模板不存在或渲染失败: %v，跳过\n", err)
		return nil
	}

	postPath := data.ThemeConfig.PostPath
	if postPath == "" {
		postPath = DefaultPostPath
	}

	var listPosts []template.PostView
	for _, p := range data.Posts {
		if !p.HideInList {
			listPosts = append(listPosts, p)
		}
	}

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
		_, _ = fmt.Fprintf(os.Stderr, "博客列表页渲染失败: %v，跳过\n", err)
		return nil
	}

	total := len(listPosts)
	size := pageSize(data.ThemeConfig.PostPageSize, 10)
	totalPages := (total + size - 1) / size
	if totalPages < 1 {
		totalPages = 1
	}
	_, _ = fmt.Fprintf(os.Stderr, "✅ 博客列表页渲染成功（共 %d 页）\n", totalPages)
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
		_, _ = fmt.Fprintf(os.Stderr, "标签列表页模板不存在或渲染失败: %v，跳过\n", err)
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

	_, _ = fmt.Fprintf(os.Stderr, "✅ 标签列表页渲染成功\n")
	return os.WriteFile(filepath.Join(tagsDir, FileIndexHTML), buf.Bytes(), 0644)
}

// renderTagPages 渲染每个标签的文章列表页（支持分页）
func (s *RendererService) renderTagPages(ctx context.Context, buildDir string, data *template.TemplateData, config domain.ThemeConfig) error {
	tagPath := config.TagPath
	if tagPath == "" {
		tagPath = DefaultTagPath
	}

	size := pageSize(data.ThemeConfig.PostPageSize, 10)

	for _, tag := range data.Tags {
		// 检查 Context（每个标签循环入口处检查，避免已取消时继续大量渲染）
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 筛选该标签下的文章
		var tagPosts []template.PostView
		for _, post := range data.Posts {
			for _, pt := range post.Tags {
				if pt.Name == tag.Name {
					tagPosts = append(tagPosts, post)
					break
				}
			}
		}

		// 构建标签页专属基础数据
		tagBaseData := *data
		tagBaseData.Tag = tag
		tagBaseData.SiteTitle = tag.Name + " | " + data.ThemeConfig.SiteName

		tagDir := filepath.Join(buildDir, tagPath, tag.Name)
		err := s.renderPaginated(ctx, paginatedRenderConfig{
			templateName: "tag",
			baseURL:      "/" + tagPath + "/" + tag.Name + "/",
			firstPageDir: tagDir,
			pageBaseDir:  tagDir,
			pageSize:     size,
			items:        tagPosts,
			baseData:     &tagBaseData,
		})
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "标签 %s 页渲染失败: %v，跳过\n", tag.Name, err)
			// 单个标签失败不影响其他标签
			continue
		}
	}

	if len(data.Tags) > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "✅ %d 个标签页渲染成功\n", len(data.Tags))
	}
	return nil
}

// renderArchives 渲染归档页（支持分页）
func (s *RendererService) renderArchives(ctx context.Context, buildDir string, data *template.TemplateData) error {
	archivesPath := DefaultArchivesPath

	// 归档页只展示已发布且不隐藏的文章
	var listPosts []template.PostView
	for _, p := range data.Posts {
		if !p.HideInList {
			listPosts = append(listPosts, p)
		}
	}

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
		_, _ = fmt.Fprintf(os.Stderr, "归档页模板不存在或渲染失败: %v，跳过\n", err)
		return nil
	}

	total := len(listPosts)
	size := pageSize(data.ThemeConfig.ArchivesPageSize, 10)
	totalPages := (total + size - 1) / size
	if totalPages < 1 {
		totalPages = 1
	}
	_, _ = fmt.Fprintf(os.Stderr, "✅ 归档页渲染成功（共 %d 页）\n", totalPages)
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
		_, _ = fmt.Fprintf(os.Stderr, "友链页模板不存在或渲染失败: %v，跳过\n", err)
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

	_, _ = fmt.Fprintf(os.Stderr, "✅ 友链页渲染成功\n")
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
		_, _ = fmt.Fprintf(os.Stderr, "闪念页模板不存在或渲染失败: %v，跳过\n", err)
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

	_, _ = fmt.Fprintf(os.Stderr, "✅ 闪念页渲染成功\n")
	return os.WriteFile(filepath.Join(memosDir, FileIndexHTML), buf.Bytes(), 0644)
}

// render404 渲染 404 页面
func (s *RendererService) render404(buildDir string, data *template.TemplateData) error {
	// 如果主题没有 404 页面，直接跳过并不会报错，保证旧主题兼容性
	html, err := s.renderer.Render("404", data)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "404 页模板不存在或渲染失败: %v，跳过\n", err)
		return nil
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	buf.WriteString(html)

	_, _ = fmt.Fprintf(os.Stderr, "✅ 404 页面渲染成功\n")
	// 注意 404 页面通常直接在根目录生成 404.html 文件
	return os.WriteFile(filepath.Join(buildDir, "404.html"), buf.Bytes(), 0644)
}

// ─── 搜索数据 JSON ────────────────────────────────────────────────────────────

// searchEntry 搜索索引条目
type searchEntry struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Date    string `json:"date"`
	Content string `json:"content"`
}

// renderSearchJSON 生成搜索数据 /api/search.json
// 包含所有已发布文章的标题、链接、日期和纯文本内容，供客户端搜索使用
func (s *RendererService) renderSearchJSON(buildDir string, data *template.TemplateData) error {
	var entries []searchEntry
	for _, post := range data.Posts {
		if post.HideInList {
			continue
		}
		// 将 HTML 内容转为纯文本用于搜索
		plainContent := stripHTMLForSearch(string(post.Content))
		// 限制内容长度（搜索不需要全文，5000 字足够）
		if len([]rune(plainContent)) > 5000 {
			plainContent = string([]rune(plainContent)[:5000])
		}
		entries = append(entries, searchEntry{
			Title:   post.Title,
			Link:    post.Link,
			Date:    post.DateFormat,
			Content: plainContent,
		})
	}

	jsonData, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("序列化搜索数据失败: %w", err)
	}

	apiDir := filepath.Join(buildDir, "api")
	if err := os.MkdirAll(apiDir, 0755); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(os.Stderr, "✅ 搜索数据生成成功 (%d 篇文章)\n", len(entries))
	return os.WriteFile(filepath.Join(apiDir, "search.json"), jsonData, 0644)
}

// stripHTMLForSearch 移除 HTML 标签，返回纯文本（用于搜索索引）。
// 使用 golang.org/x/net/html 解析器，正确处理：
//   - <script> / <style> 标签内容（完整跳过，不写入索引）
//   - HTML 实体（&amp; → &，&nbsp; → 空格 等）
func stripHTMLForSearch(s string) string {
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		// 降级到简单状态机，保证不因解析错误而中断
		return stripHTMLFallback(s)
	}
	var b strings.Builder
	b.Grow(len(s))
	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// 完整跳过 script / style 节点（含其所有子节点）
			switch n.Data {
			case "script", "style":
				return
			}
		}
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	extract(doc)
	return strings.TrimSpace(b.String())
}

// stripHTMLFallback 降级用简单状态机（当 html.Parse 失败时使用）
func stripHTMLFallback(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ─── 辅助生成函数 ─────────────────────────────────────────────────────────────

// renderRobotsTxt 自动生成 robots.txt
func (s *RendererService) renderRobotsTxt(buildDir string, data *template.TemplateData) error {
	domainUrl := strings.TrimRight(data.ThemeConfig.Domain, "/")

	var content strings.Builder
	content.WriteString("User-agent: *\n")
	content.WriteString("Allow: /\n")

	if domainUrl != "" {
		content.WriteString(fmt.Sprintf("\nSitemap: %s/sitemap.xml\n", domainUrl))
	}

	return os.WriteFile(filepath.Join(buildDir, "robots.txt"), []byte(content.String()), 0644)
}

// getMimeType 根据图片后缀返回 MIME
func getMimeType(imgUrl string) string {
	ext := strings.ToLower(filepath.Ext(imgUrl))
	switch ext {
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	default:
		return "image/jpeg"
	}
}

// safeUrl 将含有中文或空格的 URL 转成标准的百分号编码 URL
func safeUrl(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return parsed.String()
}

// ─── RSS ──────────────────────────────────────────────────────────────────────

// renderRSS 渲染 RSS 订阅 (feed.xml, RSS 2.0 规范)
func (s *RendererService) renderRSS(buildDir string, data *template.TemplateData) error {
	domainUrl := strings.TrimRight(data.ThemeConfig.Domain, "/")
	if domainUrl == "" {
		_, _ = fmt.Fprintf(os.Stderr, "警告：未配置域名，RSS (feed.xml) 中的链接可能无效\n")
	}

	lastBuild := time.Now().Format(time.RFC1123Z)
	if len(data.Posts) > 0 {
		// 使用最新文章的 UpdatedAt（最后修改时间）作为 lastBuildDate
		lastBuild = data.Posts[0].UpdatedAt.Format(time.RFC1123Z)
	}

	language := data.ThemeConfig.Language
	if language == "" {
		language = "zh-cn"
	}

	feed := rssFeed{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Channel: rssChannel{
			Title:         data.ThemeConfig.SiteName,
			Link:          safeUrl(domainUrl + "/"),
			Description:   data.ThemeConfig.SiteDescription,
			Language:      language,
			Generator:     "Gridea Pro",
			LastBuildDate: lastBuild,
			AtomLink: rssAtomLink{
				Href: safeUrl(domainUrl + "/feed.xml"),
				Rel:  "self",
				Type: "application/rss+xml",
			},
		},
	}

	feedCount := data.ThemeConfig.FeedCount
	if feedCount <= 0 {
		feedCount = 20 // 如果配置确实缺失，回退到前端默认的 20，但最好直接用读到的值
	}

	count := 0
	for _, post := range data.Posts {
		if post.HideInList || !post.Published {
			continue
		}
		if count >= feedCount {
			break
		}

		// 内容: 优先判断配置，如果关闭全文，强制使用摘要
		content := string(post.Content)
		if !data.ThemeConfig.FeedFullText {
			if string(post.Abstract) != "" {
				content = string(post.Abstract)
			} else if len(post.Content) > 200 {
				content = string(post.Content)[:200] + "..."
			}
		}

		link := domainUrl + post.Link
		if domainUrl == "" {
			link = post.Link
		}

		// 必须提供完整的绝对路径图片
		content = strings.ReplaceAll(content, "src=\"/", "src=\""+safeUrl(domainUrl)+"/")
		content = strings.ReplaceAll(content, "href=\"/", "href=\""+safeUrl(domainUrl)+"/")

		var enclosure *rssEnclosure
		if post.Feature != "" {
			featureImage := post.Feature
			if !strings.HasPrefix(featureImage, "http") {
				if strings.HasPrefix(featureImage, "/") {
					featureImage = domainUrl + featureImage
				} else {
					featureImage = domainUrl + "/" + featureImage
				}
			}
			enclosure = &rssEnclosure{
				URL:    safeUrl(featureImage),
				Length: "0",
				Type:   getMimeType(featureImage),
			}
		}

		var categories []string
		for _, t := range post.Tags {
			categories = append(categories, t.Name)
		}

		feed.Channel.Items = append(feed.Channel.Items, rssItem{
			Title:       post.Title,
			Link:        safeUrl(link),
			Guid:        rssGuid{IsPermaLink: true, Value: safeUrl(link)},
			PubDate:     post.Date.Format(time.RFC1123Z), // pubDate = 首次发布时间（RSS 2.0 规范语义）
			Description: CDATA{Text: content},
			Categories:  categories,
			Enclosure:   enclosure,
		})
		count++
	}

	rssData, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return fmt.Errorf("生成 feed.xml 失败: %w", err)
	}

	finalOutput := []byte(xml.Header + string(rssData))

	_, _ = fmt.Fprintf(os.Stderr, "✅ RSS (feed.xml) 生成成功 (%d 篇文章)\n", len(feed.Channel.Items))
	return os.WriteFile(filepath.Join(buildDir, "feed.xml"), finalOutput, 0644)
}

// ─── Sitemap ──────────────────────────────────────────────────────────────────

// renderSitemap 渲染站点地图 (sitemap.xml)
func (s *RendererService) renderSitemap(buildDir string, data *template.TemplateData) error {
	domainUrl := strings.TrimRight(data.ThemeConfig.Domain, "/")
	if domainUrl == "" {
		_, _ = fmt.Fprintf(os.Stderr, "警告：未配置域名，Sitemap (sitemap.xml) 中的链接可能无效\n")
	}

	nowDate := time.Now().Format("2006-01-02T15:04:05-07:00")

	urlset := sitemapURLSet{
		Xmlns:   "http://www.sitemaps.org/schemas/sitemap/0.9",
		ImageNs: "http://www.google.com/schemas/sitemap-image/1.1",
	}

	// 1. 首页
	urlset.Urls = append(urlset.Urls, sitemapURL{
		Loc:     safeUrl(domainUrl + "/"),
		LastMod: nowDate,
	})

	// 2. 文章页
	for _, post := range data.Posts {
		if !post.Published || post.HideInList {
			continue
		}
		link := domainUrl + post.Link
		if domainUrl == "" {
			link = post.Link
		}
		var imageNode *sitemapImage
		if post.Feature != "" {
			featureImage := post.Feature
			if !strings.HasPrefix(featureImage, "http") {
				if strings.HasPrefix(featureImage, "/") {
					featureImage = domainUrl + featureImage
				} else {
					featureImage = domainUrl + "/" + featureImage
				}
			}
			imageNode = &sitemapImage{Loc: safeUrl(featureImage)}
		}

		urlset.Urls = append(urlset.Urls, sitemapURL{
			Loc:     safeUrl(link),
			LastMod: post.UpdatedAt.Format("2006-01-02T15:04:05-07:00"), // 使用 UpdatedAt 而非创建时间
			Image:   imageNode,
		})
	}

	// 3. 标签页 (主标签列表)
	tagsPath := data.ThemeConfig.TagsPath
	if tagsPath == "" {
		tagsPath = DefaultTagsPath
	}
	urlset.Urls = append(urlset.Urls, sitemapURL{
		Loc:     safeUrl(domainUrl + "/" + tagsPath + "/"),
		LastMod: nowDate,
	})

	// 4. 每个标签的文章列表页
	for _, tag := range data.Tags {
		urlset.Urls = append(urlset.Urls, sitemapURL{
			Loc:     safeUrl(domainUrl + tag.Link),
			LastMod: nowDate, // 标签页的内容可能会经常变，使用生成时间
		})
	}

	// 5. 其他页面 (归档)
	archivesPath := "archives"
	if archivesPath == "" {
		archivesPath = DefaultArchivesPath
	}
	urlset.Urls = append(urlset.Urls, sitemapURL{
		Loc:     safeUrl(domainUrl + "/" + archivesPath + "/"),
		LastMod: nowDate,
	})

	sitemapData, err := xml.MarshalIndent(urlset, "", "  ")
	if err != nil {
		return fmt.Errorf("生成 sitemap.xml 失败: %w", err)
	}

	finalOutput := []byte(xml.Header + string(sitemapData))

	_, _ = fmt.Fprintf(os.Stderr, "✅ Sitemap (sitemap.xml) 生成成功 (%d 个链接)\n", len(urlset.Urls))
	return os.WriteFile(filepath.Join(buildDir, "sitemap.xml"), finalOutput, 0644)
}
