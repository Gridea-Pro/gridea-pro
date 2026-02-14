package service

import (
	"bytes"
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/template"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// bufferPool optimizes memory usage for large strings
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// renderIndex 渲染首页
func (s *RendererService) renderIndex(buildDir string, data *template.TemplateData) error {
	_, _ = fmt.Fprintf(os.Stderr, "开始渲染首页，使用 %s 引擎\n", s.renderer.GetEngineType())

	// Retrieving a buffer for potential usage (even if Render returns string, we can use it for file writing prep if needed)
	// In this specific flow, Render returns string, so we use it directly to write file.
	// But as per instruction, we should use the pool.
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	// 使用新的渲染器接口
	html, err := s.renderer.Render("index", data)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "❌ 渲染失败: %v，使用简单模板\n", err)
		return s.renderSimpleIndex(buildDir, data)
	}

	_, _ = fmt.Fprintf(os.Stderr, "✅ 首页渲染成功\n")

	// Use the buffer to hold the content before writing (optional optimization for io.Writer compatibility)
	buf.WriteString(html)
	return os.WriteFile(filepath.Join(buildDir, FileIndexHTML), buf.Bytes(), 0644)
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
	})
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

// renderBlog 渲染博客列表页
func (s *RendererService) renderBlog(buildDir string, data *template.TemplateData) error {
	// 尝试使用 blog.ejs 模板
	html, err := s.renderer.Render("blog", data)
	if err != nil {
		// 如果主题没有 blog.ejs 模板，跳过
		_, _ = fmt.Fprintf(os.Stderr, "博客列表页模板不存在或渲染失败: %v，跳过\n", err)
		return nil
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	buf.WriteString(html)

	// 输出路径: {buildDir}/{postPath}/index.html
	postPath := data.ThemeConfig.PostPath
	if postPath == "" {
		postPath = DefaultPostPath
	}
	blogDir := filepath.Join(buildDir, postPath)
	if err := os.MkdirAll(blogDir, 0755); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(os.Stderr, "✅ 博客列表页渲染成功\n")
	return os.WriteFile(filepath.Join(blogDir, FileIndexHTML), buf.Bytes(), 0644)
}

// renderTags 渲染标签列表页
func (s *RendererService) renderTags(buildDir string, _ context.Context, data *template.TemplateData, _ domain.ThemeConfig) error {
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

// renderTagPages 渲染每个标签的文章列表页
func (s *RendererService) renderTagPages(buildDir string, _ context.Context, data *template.TemplateData, config domain.ThemeConfig) error {
	tagPath := config.TagPath
	if tagPath == "" {
		tagPath = DefaultTagPath
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buf)

	for _, tag := range data.Tags {
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

		// 构建标签页专属数据
		tagData := *data
		tagData.Tag = tag
		tagData.Posts = tagPosts
		tagData.SiteTitle = tag.Name + " | " + data.ThemeConfig.SiteName

		html, err := s.renderer.Render("tag", &tagData)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "标签 %s 页模板不存在或渲染失败: %v，跳过\n", tag.Name, err)
			continue
		}

		// Reset buffer for each iteration
		buf.Reset()
		buf.WriteString(html)

		tagDir := filepath.Join(buildDir, tagPath, tag.Name)
		if err := os.MkdirAll(tagDir, 0755); err != nil {
			return err
		}

		if err := os.WriteFile(filepath.Join(tagDir, FileIndexHTML), buf.Bytes(), 0644); err != nil {
			return err
		}
	}

	if len(data.Tags) > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "✅ %d 个标签页渲染成功\n", len(data.Tags))
	}
	return nil
}

// renderArchives 渲染归档页
func (s *RendererService) renderArchives(buildDir string, data *template.TemplateData) error {
	html, err := s.renderer.Render("archives", data)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "归档页模板不存在或渲染失败: %v，跳过\n", err)
		return nil
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	buf.WriteString(html)

	archivesPath := data.ThemeConfig.ArchivesPath
	if archivesPath == "" {
		archivesPath = DefaultArchivesPath
	}
	archivesDir := filepath.Join(buildDir, archivesPath)
	if err := os.MkdirAll(archivesDir, 0755); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(os.Stderr, "✅ 归档页渲染成功\n")
	return os.WriteFile(filepath.Join(archivesDir, FileIndexHTML), buf.Bytes(), 0644)
}

// renderFriends 渲染友链页
func (s *RendererService) renderFriends(buildDir string, _ context.Context, data *template.TemplateData) error {
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
func (s *RendererService) renderMemos(buildDir string, _ context.Context, data *template.TemplateData) error {
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
