package service

import (
	"context"
	"encoding/json"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/render"
	"gridea-pro/backend/internal/template"
	"gridea-pro/backend/internal/utils"
	htmlTemplate "html/template"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dop251/goja"
)

type RendererService struct {
	postRepo    domain.PostRepository
	themeRepo   domain.ThemeRepository
	settingRepo domain.SettingRepository
	menuRepo    domain.MenuRepository
	commentRepo domain.CommentRepository
	appDir      string

	// 主题配置服务
	themeConfigService *ThemeConfigService

	// 主题渲染器(新架构)
	renderer render.ThemeRenderer
}

func NewRendererService(
	appDir string,
	postRepo domain.PostRepository,
	themeRepo domain.ThemeRepository,
	settingRepo domain.SettingRepository,
) *RendererService {
	return &RendererService{
		postRepo:           postRepo,
		themeRepo:          themeRepo,
		settingRepo:        settingRepo,
		appDir:             appDir,
		themeConfigService: NewThemeConfigService(appDir),
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

// SetTheme 设置主题并初始化渲染器
func (s *RendererService) SetTheme(themeName string) error {
	factory := render.NewRendererFactory(s.appDir, themeName)
	renderer, err := factory.CreateRenderer()
	if err != nil {
		return fmt.Errorf("创建渲染器失败: %w", err)
	}
	s.renderer = renderer
	fmt.Printf("✅ 使用 %s 引擎渲染主题: %s\n", renderer.GetEngineType(), themeName)
	return nil
}

func (s *RendererService) RenderAll(ctx context.Context) error {
	// 获取数据
	posts, err := s.postRepo.GetAll(ctx)
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

	buildDir := filepath.Join(s.appDir, "output")
	_ = os.RemoveAll(buildDir)
	_ = os.MkdirAll(buildDir, 0755)

	// 1. 复制主题资源
	if err := s.copyThemeAssets(buildDir, themeConfig.ThemeName); err != nil {
		fmt.Printf("警告：复制主题资源失败: %v\n", err)
	}

	// 2. 复制站点静态资源（images、media 等）
	if err := s.copySiteAssets(buildDir); err != nil {
		fmt.Printf("警告：复制站点资源失败: %v\n", err)
	}

	// 3. 构建模板数据
	templateData, err := s.buildTemplateData(ctx, posts, themeConfig)
	if err != nil {
		return fmt.Errorf("构建模板数据失败: %w", err)
	}

	// 4. 渲染首页
	if err := s.renderIndex(buildDir, templateData); err != nil {
		return fmt.Errorf("渲染首页失败: %w", err)
	}

	// 5. 渲染文章详情页
	for _, post := range posts {
		if !post.Data.Published {
			continue
		}
		if err := s.renderPost(buildDir, post, templateData); err != nil {
			return fmt.Errorf("渲染文章 %s 失败: %w", post.Data.Title, err)
		}
	}

	// 6. 渲染标签页（TODO：后续实现）
	// 7. 渲染归档页（TODO：后续实现）

	fmt.Printf("渲染完成，共 %d 篇文章\n", len(posts))
	return nil
}

// copyThemeAssets 复制主题静态资源
func (s *RendererService) copyThemeAssets(buildDir, themeName string) error {
	themePath := filepath.Join(s.appDir, "themes", themeName)
	assetsPath := filepath.Join(themePath, "assets")
	if _, err := os.Stat(assetsPath); os.IsNotExist(err) {
		return nil
	}

	// 1. 检查并编译 LESS 文件
	lessPath := filepath.Join(assetsPath, "styles", "main.less")
	if _, err := os.Stat(lessPath); err == nil {
		fmt.Println("检测到 LESS 文件，开始编译...")
		if err := s.compileLess(lessPath, buildDir); err != nil {
			fmt.Printf("警告：LESS 编译失败: %v\n", err)
		} else {
			fmt.Println("✅ LESS 编译成功")
		}
	}

	// 2. 复制其他静态资源
	destPath := filepath.Join(buildDir)
	return copyDir(assetsPath, destPath)
}

// compileLess 编译 LESS 文件为 CSS
func (s *RendererService) compileLess(lessPath, buildDir string) error {
	// 输出路径
	cssPath := filepath.Join(buildDir, "styles", "main.css")

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(cssPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 调用 lessc 命令编译
	cmd := exec.Command("lessc", lessPath, cssPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("lessc 编译失败: %w\n输出: %s", err, string(output))
	}

	// 检查并应用 style-override.js
	// 从 lessPath 推导主题路径
	themePath := filepath.Dir(filepath.Dir(lessPath))
	overridePath := filepath.Join(themePath, "style-override.js")
	if _, err := os.Stat(overridePath); err == nil {
		fmt.Println("检测到 style-override.js，应用自定义样式...")
		customCSS, err := s.applyStyleOverride(overridePath)
		if err != nil {
			fmt.Printf("警告：应用 style-override.js 失败: %v\n", err)
		} else {
			// 读取编译后的 CSS
			cssContent, err := os.ReadFile(cssPath)
			if err != nil {
				return fmt.Errorf("读取 CSS 文件失败: %w", err)
			}

			// 追加自定义 CSS
			cssContent = append(cssContent, []byte("\n"+customCSS)...)
			if err := os.WriteFile(cssPath, cssContent, 0644); err != nil {
				return fmt.Errorf("写入 CSS 文件失败: %w", err)
			}
			fmt.Println("✅ 自定义样式应用成功")
		}
	}

	return nil
}

// applyStyleOverride 执行 style-override.js 并返回自定义 CSS
func (s *RendererService) applyStyleOverride(jsPath string) (string, error) {
	// 读取 JS 文件
	jsCode, err := os.ReadFile(jsPath)
	if err != nil {
		return "", fmt.Errorf("读取 style-override.js 失败: %w", err)
	}

	// 创建 JS 运行时
	vm := goja.New()

	// 执行 JS 代码
	_, err = vm.RunString(string(jsCode))
	if err != nil {
		return "", fmt.Errorf("执行 JS 代码失败: %w", err)
	}

	// 获取 module.exports (generateOverride 函数)
	moduleExports := vm.Get("module")
	if moduleExports == nil || goja.IsUndefined(moduleExports) {
		// 尝试直接获取 generateOverride
		generateOverride := vm.Get("generateOverride")
		if generateOverride == nil || goja.IsUndefined(generateOverride) {
			return "", fmt.Errorf("未找到 generateOverride 函数")
		}

		// 调用函数
		// 从 jsPath 推导 themeName
		themePath := filepath.Dir(jsPath)
		themeName := filepath.Base(themePath)
		customConfig := s.loadThemeCustomConfig(themeName)
		result, err := vm.RunString(fmt.Sprintf("generateOverride(%s)", toJSON(customConfig)))
		if err != nil {
			return "", fmt.Errorf("调用 generateOverride 失败: %w", err)
		}

		return result.String(), nil
	}

	// CommonJS 模块格式：module.exports = generateOverride
	exports := moduleExports.ToObject(vm).Get("exports")
	if exports == nil || goja.IsUndefined(exports) {
		return "", fmt.Errorf("module.exports 未定义")
	}

	// 调用导出的函数
	fn, ok := goja.AssertFunction(exports)
	if !ok {
		return "", fmt.Errorf("module.exports 不是函数")
	}

	// 准备参数
	// 从 jsPath 推导 themeName
	themePath := filepath.Dir(jsPath)
	themeName := filepath.Base(themePath)
	customConfig := s.loadThemeCustomConfig(themeName)
	configValue := vm.ToValue(customConfig)

	// 调用函数
	result, err := fn(goja.Undefined(), configValue)
	if err != nil {
		return "", fmt.Errorf("调用 generateOverride 失败: %w", err)
	}

	return result.String(), nil
}

// toJSON 将 map 转换为 JSON 字符串（用于 JS 调用）
func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// copySiteAssets 复制站点静态资源
func (s *RendererService) copySiteAssets(buildDir string) error {
	// 复制 images 目录
	imagesPath := filepath.Join(s.appDir, "images")
	if _, err := os.Stat(imagesPath); err == nil {
		if err := copyDir(imagesPath, filepath.Join(buildDir, "images")); err != nil {
			return err
		}
	}

	// 复制 media 目录
	mediaPath := filepath.Join(s.appDir, "media")
	if _, err := os.Stat(mediaPath); err == nil {
		if err := copyDir(mediaPath, filepath.Join(buildDir, "media")); err != nil {
			return err
		}
	}

	// 复制 post-images 目录
	postImagesPath := filepath.Join(s.appDir, "post-images")
	if _, err := os.Stat(postImagesPath); err == nil {
		if err := copyDir(postImagesPath, filepath.Join(buildDir, "post-images")); err != nil {
			return err
		}
	}

	return nil
}

// buildTemplateData 构建模板数据
func (s *RendererService) buildTemplateData(ctx context.Context, posts []domain.Post, config domain.ThemeConfig) (*template.TemplateData, error) {
	// 转换文章
	postViews := make([]template.PostView, 0, len(posts))
	for _, post := range posts {
		if !post.Data.Published {
			continue
		}
		postViews = append(postViews, s.convertPost(post, config))
	}

	// 获取菜单
	var menuViews []template.MenuView
	if s.menuRepo != nil {
		menus, _ := s.menuRepo.GetAll(ctx)
		for _, menu := range menus {
			menuViews = append(menuViews, template.MenuView{
				Name:     menu.Name,
				Link:     menu.Link,
				OpenType: menu.OpenType,
			})
		}
	}

	// 获取主题自定义配置
	customConfig := s.loadThemeCustomConfig(config.ThemeName)

	// 获取评论设置
	commentSettingView := s.buildCommentSettingView(ctx)

	data := &template.TemplateData{
		ThemeConfig: template.ThemeConfigView{
			ThemeName:        config.ThemeName,
			SiteName:         config.SiteName,
			SiteDescription:  config.SiteDescription,
			FooterInfo:       config.FooterInfo,
			ShowFeatureImage: config.ShowFeatureImage,
			Domain:           config.Domain,
			PostPageSize:     config.PostPageSize,
			ArchivesPageSize: config.ArchivesPageSize,
			PostUrlFormat:    config.PostUrlFormat,
			TagUrlFormat:     config.TagUrlFormat,
			DateFormat:       config.DateFormat,
			FeedFullText:     config.FeedFullText,
			FeedCount:        config.FeedCount,
			ArchivesPath:     config.ArchivesPath,
			PostPath:         config.PostPath,
			TagPath:          config.TagPath,
			LinkPath:         config.LinkPath,
		},
		Site: template.SiteView{
			CustomConfig: customConfig,
			Utils:        template.NewSiteUtils(),
		},
		Posts:          postViews,
		Menus:          menuViews,
		CommentSetting: commentSettingView,
		Pagination: template.PaginationView{
			Current: 1,
			Total:   1,
		},
	}

	return data, nil
}

// loadThemeCustomConfig 加载主题自定义配置
func (s *RendererService) loadThemeCustomConfig(themeName string) map[string]interface{} {
	// 使用 ThemeConfigService 加载配置
	config, err := s.themeConfigService.GetFinalConfig(themeName)
	if err != nil {
		fmt.Printf("警告：加载主题配置失败，使用空配置: %v\n", err)
		return make(map[string]interface{})
	}

	return config
}

// buildCommentSettingView 构建评论设置视图
func (s *RendererService) buildCommentSettingView(ctx context.Context) template.CommentSettingView {
	if s.commentRepo == nil {
		return template.CommentSettingView{}
	}

	settings, err := s.commentRepo.GetSettings(ctx)
	if err != nil {
		fmt.Printf("警告：加载评论设置失败: %v\n", err)
		return template.CommentSettingView{}
	}

	if !settings.Enable {
		return template.CommentSettingView{}
	}

	view := template.CommentSettingView{
		ShowComment: settings.Enable,
		Platform:    string(settings.Platform),
	}

	// Config   map[string]any  `json:"config"` // Deprecated - Removed logic reading from Config
	// Instead, read from specific fields in PlatformConfigs

	getConfig := func(p domain.CommentPlatform) map[string]any {
		if settings.PlatformConfigs == nil {
			return nil
		}
		return settings.PlatformConfigs[p]
	}

	// 根据平台类型提取配置
	switch settings.Platform {
	case domain.CommentPlatformValine:
		config := getConfig(domain.CommentPlatformValine)
		if config != nil {
			view.AppID, _ = config["appId"].(string)
			view.AppKey, _ = config["appKey"].(string)
			view.ServerURLs, _ = config["serverURLs"].(string)
		}
	case domain.CommentPlatformWaline:
		config := getConfig(domain.CommentPlatformWaline)
		if config != nil {
			view.AppID, _ = config["appId"].(string)
			view.AppKey, _ = config["appKey"].(string)
			view.ServerURLs, _ = config["serverURLs"].(string)
		}
	case domain.CommentPlatformTwikoo:
		config := getConfig(domain.CommentPlatformTwikoo)
		if config != nil {
			view.EnvID, _ = config["envId"].(string)
		}
	case domain.CommentPlatformGitalk:
		config := getConfig(domain.CommentPlatformGitalk)
		if config != nil {
			view.ClientID, _ = config["clientId"].(string)
			view.ClientSecret, _ = config["clientSecret"].(string)
			view.Repo, _ = config["repo"].(string)
			view.Owner, _ = config["owner"].(string)
			view.Admin, _ = config["admin"].(string)
		}
	case domain.CommentPlatformGiscus:
		config := getConfig(domain.CommentPlatformGiscus)
		if config != nil {
			view.Repo, _ = config["repo"].(string)
			view.RepoID, _ = config["repoId"].(string)
			view.Category, _ = config["category"].(string)
			view.CategoryID, _ = config["categoryId"].(string)
		}
	case domain.CommentPlatformDisqus:
		config := getConfig(domain.CommentPlatformDisqus)
		if config != nil {
			view.Shortname, _ = config["shortname"].(string)
			view.API, _ = config["api"].(string)
			view.APIKey, _ = config["apiKey"].(string)
		}
	case domain.CommentPlatformCusdis:
		config := getConfig(domain.CommentPlatformCusdis)
		if config != nil {
			view.AppID, _ = config["appId"].(string)
			view.Host, _ = config["host"].(string)
		}
	}

	return view
}

// convertPost 将 domain.Post 转换为 template.PostView
func (s *RendererService) convertPost(post domain.Post, config domain.ThemeConfig) template.PostView {
	postPath := config.PostPath
	if postPath == "" {
		postPath = "post"
	}

	// 生成链接
	link := "/" + postPath + "/" + post.FileName + "/"

	// 转换标签
	var tags []template.TagView
	var tagNames []string
	for _, tag := range post.Data.Tags {
		tagView := template.TagView{
			Name: tag,
			Slug: tag,
			Link: "/" + config.TagPath + "/" + tag + "/",
		}
		tags = append(tags, tagView)
		tagNames = append(tagNames, tag)
	}

	// 转换分类
	var categories []template.CategoryView
	for _, category := range post.Data.Categories {
		categoryView := template.CategoryView{
			Name: category,
			Slug: category,                                    // 简单起见，暂用 name 作为 slug
			Link: "/" + config.TagPath + "/" + category + "/", // TODO: 后续应有单独的 categoryPath
		}
		categories = append(categories, categoryView)
	}

	// 计算阅读统计
	wordCount := utf8.RuneCountInString(post.Content)
	readingTime := wordCount / 200
	if readingTime < 1 {
		readingTime = 1
	}

	// 解析日期
	var postDate time.Time
	if post.Data.Date != "" {
		postDate, _ = time.Parse("2006-01-02 15:04:05", post.Data.Date)
		if postDate.IsZero() {
			postDate, _ = time.Parse("2006-01-02", post.Data.Date)
		}
	}

	// 格式化日期
	dateFormat := config.DateFormat
	if dateFormat == "" {
		dateFormat = "YYYY-MM-DD"
	}
	formattedDate := formatDate(postDate, dateFormat)

	// 生成摘要
	abstract := post.Abstract
	if abstract == "" && len(post.Content) > 200 {
		abstract = post.Content[:200] + "..."
	}

	// 将 Markdown 内容转换为 HTML
	contentHTML := utils.ToHTMLUnsafe(post.Content)
	abstractHTML := utils.ToHTML(abstract)

	return template.PostView{
		Title:       post.Data.Title,
		FileName:    post.FileName,
		Content:     htmlTemplate.HTML(contentHTML),  // 转换为 template.HTML 类型
		Abstract:    htmlTemplate.HTML(abstractHTML), // 转换为 template.HTML 类型
		Description: "",                              // TODO: 从文章 frontmatter 读取
		Link:        link,
		Feature:     post.Data.Feature,
		Date:        postDate,
		DateFormat:  formattedDate,
		Published:   post.Data.Published,
		HideInList:  post.Data.HideInList,
		IsTop:       post.Data.IsTop,
		Tags:        tags,
		Categories:  categories,
		TagsString:  strings.Join(tagNames, ","),
		Stats: template.PostStats{
			Words:   wordCount,
			Minutes: readingTime,
			Text:    fmt.Sprintf("%d 分钟阅读", readingTime),
		},
		Toc: "", // TODO: 生成文章目录
	}
}

// renderIndex 渲染首页
func (s *RendererService) renderIndex(buildDir string, data *template.TemplateData) error {
	fmt.Printf("开始渲染首页，使用 %s 引擎\n", s.renderer.GetEngineType())

	// 使用新的渲染器接口
	html, err := s.renderer.Render("index", data)
	if err != nil {
		fmt.Printf("❌ 渲染失败: %v，使用简单模板\n", err)
		return s.renderSimpleIndex(buildDir, data)
	}

	fmt.Printf("✅ 首页渲染成功\n")
	return os.WriteFile(filepath.Join(buildDir, "index.html"), []byte(html), 0644)
}

// renderSimpleIndex 渲染简单首页（备用）
func (s *RendererService) renderSimpleIndex(buildDir string, data *template.TemplateData) error {
	var postListHTML strings.Builder
	for _, p := range data.Posts {
		postListHTML.WriteString(fmt.Sprintf(`
			<article class="post">
				<h2 class="post-title"><a href="%s">%s</a></h2>
				<div class="post-meta">%s</div>
			</article>
		`, p.Link, p.Title, p.DateFormat))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
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

	return os.WriteFile(filepath.Join(buildDir, "index.html"), []byte(html), 0644)
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
		postPath = "post"
	}
	postDir := filepath.Join(buildDir, postPath, post.FileName)
	if err := os.MkdirAll(postDir, 0755); err != nil {
		return err
	}

	// 使用新的渲染器接口
	html, err := s.renderer.Render("post", &postData)
	if err != nil {
		fmt.Printf("文章模板渲染失败: %v，使用简单模板\n", err)
		return s.renderSimplePost(postDir, &postData)
	}

	return os.WriteFile(filepath.Join(postDir, "index.html"), []byte(html), 0644)
}

// renderSimplePost 渲染简单文章页（备用）
func (s *RendererService) renderSimplePost(postDir string, data *template.TemplateData) error {
	html := fmt.Sprintf(`<!DOCTYPE html>
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

	return os.WriteFile(filepath.Join(postDir, "index.html"), []byte(html), 0644)
}

// formatDate 格式化日期
func formatDate(t time.Time, format string) string {
	if t.IsZero() {
		return ""
	}

	// 转换格式
	format = strings.ReplaceAll(format, "YYYY", "2006")
	format = strings.ReplaceAll(format, "MM", "01")
	format = strings.ReplaceAll(format, "DD", "02")
	format = strings.ReplaceAll(format, "HH", "15")
	format = strings.ReplaceAll(format, "mm", "04")
	format = strings.ReplaceAll(format, "ss", "05")

	return t.Format(format)
}

// copyDir 递归复制目录
func copyDir(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("源路径不是目录")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Close()
}
