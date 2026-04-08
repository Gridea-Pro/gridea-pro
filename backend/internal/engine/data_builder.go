package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/template"
	"gridea-pro/backend/internal/utils"
	htmlTemplate "html/template"
	"log/slog"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// TemplateDataBuilder 负责构建模板渲染所需的数据
type TemplateDataBuilder struct {
	postRepo     domain.PostRepository
	themeRepo    domain.ThemeRepository
	settingRepo  domain.SettingRepository
	menuRepo     domain.MenuRepository
	commentRepo  domain.CommentRepository
	linkRepo     domain.LinkRepository
	tagRepo      domain.TagRepository
	memoRepo     domain.MemoRepository
	categoryRepo domain.CategoryRepository

	themeConfigService *ThemeConfigService
	logger             *slog.Logger

	// Build() 阶段缓存的查找映射，供 ConvertPost() 复用
	cachedTagByName      map[string]domain.Tag
	cachedCategoryByID   map[string]domain.Category
	cachedCategoryByName map[string]domain.Category
}

// NewTemplateDataBuilder 创建 TemplateDataBuilder
func NewTemplateDataBuilder(
	postRepo domain.PostRepository,
	themeRepo domain.ThemeRepository,
	settingRepo domain.SettingRepository,
	themeConfigService *ThemeConfigService,
) *TemplateDataBuilder {
	return &TemplateDataBuilder{
		postRepo:           postRepo,
		themeRepo:          themeRepo,
		settingRepo:        settingRepo,
		themeConfigService: themeConfigService,
		logger:             slog.Default(),
	}
}

func (b *TemplateDataBuilder) SetMenuRepo(repo domain.MenuRepository)       { b.menuRepo = repo }
func (b *TemplateDataBuilder) SetCommentRepo(repo domain.CommentRepository) { b.commentRepo = repo }
func (b *TemplateDataBuilder) SetLinkRepo(repo domain.LinkRepository)       { b.linkRepo = repo }
func (b *TemplateDataBuilder) SetTagRepo(repo domain.TagRepository)         { b.tagRepo = repo }
func (b *TemplateDataBuilder) SetMemoRepo(repo domain.MemoRepository)       { b.memoRepo = repo }
func (b *TemplateDataBuilder) SetCategoryRepo(repo domain.CategoryRepository) {
	b.categoryRepo = repo
}

// Build 构建模板数据
func (b *TemplateDataBuilder) Build(ctx context.Context, posts []domain.Post, config domain.ThemeConfig) (*template.TemplateData, error) {
	// 转换文章 (Concurrent)
	// 1. Filter published posts first to know the exact count
	var publishedPosts []domain.Post
	for _, post := range posts {
		if post.Published {
			publishedPosts = append(publishedPosts, post)
		}
	}

	// 2. 预加载分类映射（单索引）：
	//    - categoryByID: id → Category（主键，新老数据均已被洗净）
	//    - categoryByName: name → Category（兜底，用于老文章无 CategoryIDs 时按名称反查 slug）
	categoryByID := make(map[string]domain.Category)
	categoryByName := make(map[string]domain.Category)
	if b.categoryRepo != nil {
		if cats, err := b.categoryRepo.List(ctx); err == nil {
			for _, c := range cats {
				if c.ID != "" {
					categoryByID[c.ID] = c
				}
				if c.Name != "" {
					categoryByName[c.Name] = c
				}
			}
		}
	}

	// 3. 预加载标签映射，供 convertPost 使用
	tagByName := make(map[string]domain.Tag)
	if b.tagRepo != nil {
		if repoTags, err := b.tagRepo.List(ctx); err == nil {
			for _, rt := range repoTags {
				if rt.Name != "" {
					tagByName[rt.Name] = rt
				}
			}
		}
	}

	// 缓存查找映射，供 ConvertPost() 在渲染单篇文章时复用
	b.cachedTagByName = tagByName
	b.cachedCategoryByID = categoryByID
	b.cachedCategoryByName = categoryByName

	postViews := make([]template.PostView, len(publishedPosts))
	var wg sync.WaitGroup
	// Limit concurrency to number of CPUs
	sem := make(chan struct{}, runtime.NumCPU())

	for i, post := range publishedPosts {
		wg.Add(1)
		go func(idx int, p domain.Post) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			postViews[idx] = b.convertPost(p, config, categoryByID, categoryByName, tagByName)
		}(i, post)
	}
	wg.Wait()

	// 获取菜单
	normalizeLink := func(link, openType string) string {
		if openType == "Internal" && link != "" && !strings.HasPrefix(link, "/") && !strings.HasPrefix(link, "http") {
			return "/" + link
		}
		return link
	}
	var menuViews []template.MenuView
	if b.menuRepo != nil {
		menus, _ := b.menuRepo.List(ctx)
		for _, menu := range menus {
			mv := template.MenuView{
				Name:     menu.Name,
				Link:     normalizeLink(menu.Link, menu.OpenType),
				OpenType: menu.OpenType,
			}
			for _, child := range menu.Children {
				mv.Children = append(mv.Children, template.MenuView{
					Name:     child.Name,
					Link:     normalizeLink(child.Link, child.OpenType),
					OpenType: child.OpenType,
				})
			}
			menuViews = append(menuViews, mv)
		}
	}

	// 获取主题自定义配置
	customConfig := b.loadThemeCustomConfig(config.ThemeName)

	// 统计每个标签在文章中的出现次数
	tagCountMap := make(map[string]int)
	for _, pv := range postViews {
		for _, t := range pv.Tags {
			tagCountMap[t.Name]++
		}
	}
	tagPath := config.TagPath
	if tagPath == "" {
		tagPath = DefaultTagPath
	}

	// 从 tagRepo 获取标签列表（保持用户添加顺序），再用文章计数补充 count
	var allTags []template.TagView
	if b.tagRepo != nil {
		repoTags, err := b.tagRepo.List(ctx)
		if err == nil {
			for _, rt := range repoTags {
				count := tagCountMap[rt.Name]
				if count > 0 {
					allTags = append(allTags, template.TagView{
						Name:  rt.Name,
						Slug:  rt.Slug,
						Link:  "/" + tagPath + "/" + rt.Slug + "/",
						Count: count,
					})
					delete(tagCountMap, rt.Name)
				}
			}
		}
	}
	// 兜底：tagRepo 中不存在但文章中使用的标签，追加到末尾
	for name, count := range tagCountMap {
		allTags = append(allTags, template.TagView{
			Name:  name,
			Slug:  name,
			Link:  "/" + tagPath + "/" + name + "/",
			Count: count,
		})
	}

	// 从 linkRepo 获取友链数据，注入到 customConfig.friends
	if b.linkRepo != nil {
		links, err := b.linkRepo.List(ctx)
		if err == nil && len(links) > 0 {
			var friendList []map[string]interface{}
			for _, link := range links {
				friendList = append(friendList, map[string]interface{}{
					"siteName":    link.Name,
					"siteLink":    link.Url,
					"description": link.Description,
					"avatar":      link.Avatar,
				})
			}
			if customConfig == nil {
				customConfig = make(map[string]interface{})
			}
			customConfig["links"] = friendList
		}
	}

	// 注入站点头像路径（如果 images/avatar.png 存在）
	if customConfig == nil {
		customConfig = make(map[string]interface{})
	}
	if _, ok := customConfig["avatar"]; !ok {
		avatarPath := filepath.Join(b.themeConfigService.appDir, "images", "avatar.png")
		if _, err := os.Stat(avatarPath); err == nil {
			customConfig["avatar"] = "/images/avatar.png"
		}
	}

	// 获取评论设置
	commentSettingView := b.buildCommentSettingView(ctx)

	// 获取全局 Setting (包含真实的 CNAME 或 Domain)
	globalDomain := config.Domain
	if globalDomain == "" && b.settingRepo != nil {
		setting, err := b.settingRepo.GetSetting(ctx)
		if err == nil {
			if setting.CNAME() != "" {
				globalDomain = setting.CNAME()
				if !strings.HasPrefix(globalDomain, "http") {
					globalDomain = "https://" + globalDomain
				}
			} else if setting.Domain() != "" {
				globalDomain = setting.Domain()
				if !strings.HasPrefix(globalDomain, "http") {
					globalDomain = "https://" + globalDomain // HTTPS 兜底
				}
			}
		}
	}

	data := &template.TemplateData{
		ThemeConfig: template.ThemeConfigView{
			ThemeName:        config.ThemeName,
			SiteName:         config.SiteName,
			SiteDescription:  config.SiteDescription,
			FooterInfo:       config.FooterInfo,
			Domain:           globalDomain,
			PostPageSize:     config.PostPageSize,
			ArchivesPageSize: config.ArchivesPageSize,
			PostUrlFormat:    config.PostUrlFormat,
			TagUrlFormat:     config.TagUrlFormat,
			DateFormat:       config.DateFormat,
			Language:         config.Language,
			FeedFullText:     config.FeedFullText,
			FeedCount:        config.FeedCount,
			PostPath:         config.PostPath,
			TagPath:          config.TagPath,
			TagsPath:         config.TagsPath,
			LinkPath:         config.LinkPath,
			MemosPath:        config.MemosPath,
		},
		Site: template.SiteView{
			CustomConfig: customConfig,
			Utils:        template.NewSiteUtils(),
		},
		Posts:          postViews,
		Tags:           allTags,
		Memos:          b.buildMemoViews(ctx, config),
		Menus:          menuViews,
		CommentSetting: commentSettingView,
		Pagination: template.PaginationView{
			CurrentPage: 1,
			TotalPages:  1,
		},
	}

	return data, nil
}

// loadThemeCustomConfig 加载主题自定义配置
func (b *TemplateDataBuilder) loadThemeCustomConfig(themeName string) map[string]interface{} {
	// 使用 ThemeConfigService 加载配置
	config, err := b.themeConfigService.GetFinalConfig(themeName)
	if err != nil {
		b.logger.Warn("加载主题配置失败，使用空配置", "error", err)
		return make(map[string]interface{})
	}

	return config
}

// buildCommentSettingView 构建评论设置视图
func (b *TemplateDataBuilder) buildCommentSettingView(ctx context.Context) template.CommentSettingView {
	if b.commentRepo == nil {
		return template.CommentSettingView{}
	}

	settings, err := b.commentRepo.GetSettings(ctx)
	if err != nil {
		b.logger.Warn("加载评论设置失败", "error", err)
		return template.CommentSettingView{}
	}

	if !settings.Enable {
		return template.CommentSettingView{}
	}

	view := template.CommentSettingView{
		ShowComment: settings.Enable,
		Platform:    string(settings.Platform),
	}

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

// ConvertPost 将 domain.Post 转换为 template.PostView（公开方法，供 PageRenderer 使用）
// categoryByID: ID(NanoID) → domain.Category；若为 nil 则使用 Build() 阶段缓存的映射
func (b *TemplateDataBuilder) ConvertPost(post domain.Post, config domain.ThemeConfig, categoryByID map[string]domain.Category) template.PostView {
	if categoryByID == nil {
		categoryByID = b.cachedCategoryByID
	}
	return b.convertPost(post, config, categoryByID, b.cachedCategoryByName, b.cachedTagByName)
}

// convertPost 将 domain.Post 转换为 template.PostView
func (b *TemplateDataBuilder) convertPost(post domain.Post, config domain.ThemeConfig, categoryByID map[string]domain.Category, categoryByName map[string]domain.Category, tagByName map[string]domain.Tag) template.PostView {
	postPath := config.PostPath
	if postPath == "" {
		postPath = DefaultPostPath
	}

	// 生成链接
	link := "/" + postPath + "/" + post.FileName + "/"

	// 转换标签
	var tags []template.TagView
	var tagNames []string
	for _, tag := range post.Tags {
		tagSlug := tag
		if tagByName != nil {
			if t, ok := tagByName[tag]; ok {
				tagSlug = t.Slug
			}
		}
		tagView := template.TagView{
			Name: tag,
			Slug: tagSlug,
			Link: "/" + config.TagPath + "/" + tagSlug + "/",
		}
		tags = append(tags, tagView)
		tagNames = append(tagNames, tag)
	}

	// 转换分类：严格基于 CategoryIDs 查找
	var categories []template.CategoryView
	if len(post.CategoryIDs) > 0 && categoryByID != nil {
		for _, catID := range post.CategoryIDs {
			if cat, ok := categoryByID[catID]; ok {
				categories = append(categories, template.CategoryView{
					Name: cat.Name,
					Slug: cat.Slug,
					Link: "/" + DefaultCategoryPath + "/" + cat.Slug + "/",
				})
			} else {
				// ID 未命中（说明分类已删除被置空等）
				categories = append(categories, template.CategoryView{
					Name: catID,
					Slug: catID,
					Link: "/" + DefaultCategoryPath + "/" + catID + "/",
				})
			}
		}
	} else {
		// 向后兼容：老文章无 CategoryIDs，回退使用名称字符串
		for _, category := range post.Categories {
			slug := category
			if categoryByName != nil {
				if cat, ok := categoryByName[category]; ok {
					slug = cat.Slug
				}
			}
			categories = append(categories, template.CategoryView{
				Name: category,
				Slug: slug,
				Link: "/" + DefaultCategoryPath + "/" + slug + "/",
			})
		}
	}

	// 计算阅读统计
	wordCount := utf8.RuneCountInString(post.Content)
	readingTime := wordCount / 600
	if readingTime < 1 {
		readingTime = 1
	}

	// 解析日期 - 已经是 time.Time
	postDate := post.CreatedAt

	// 格式化日期
	dateFormat := config.DateFormat
	if dateFormat == "" {
		dateFormat = "YYYY-MM-DD"
	}
	formattedDate := formatDate(postDate, dateFormat)

	// 格式化更新时间（若与创建时间相同则复用）
	updatedAt := post.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = postDate
	}
	formattedUpdatedAt := formatDate(updatedAt, dateFormat)

	// 生成摘要
	abstract := post.Abstract
	if abstract == "" && len(post.Content) > 200 {
		abstract = post.Content[:200] + "..."
	}

	// 将 Markdown 内容转换为 HTML
	contentHTML := utils.ToHTMLUnsafe(post.Content)
	abstractHTML := utils.ToHTML(abstract)

	return template.PostView{
		ID:              post.ID,
		Title:           post.Title,
		FileName:        post.FileName,
		Content:         htmlTemplate.HTML(contentHTML),
		Abstract:        htmlTemplate.HTML(abstractHTML),
		Description:     "",
		Link:            link,
		Feature:         post.Feature,
		CreatedAt:       postDate,
		Date:            postDate, // 为老主题保留 date 字典
		DateFormat:      formattedDate,
		UpdatedAt:       updatedAt,
		UpdatedAtFormat: formattedUpdatedAt,
		Published:       post.Published,
		HideInList:      post.HideInList,
		IsTop:           post.IsTop,
		Tags:            tags,
		Categories:      categories,
		TagsString:      strings.Join(tagNames, ","),
		Stats: template.PostStats{
			Words:   wordCount,
			Minutes: readingTime,
			Text:    fmt.Sprintf("%d 分钟阅读", readingTime),
		},
		Toc: "", // TODO: 生成文章目录
	}
}

func formatDate(t time.Time, format string) string {
	if t.IsZero() {
		return ""
	}

	// 转换格式: 从 Moment.js 格式转换为 Go 的 time.Format 格式
	replacer := strings.NewReplacer(
		// Year
		"YYYY", "2006",
		"YY", "06",
		// Month
		"MMMM", "January",
		"MMM", "Jan",
		"MM", "01",
		"M", "1",
		// Day of month
		"DD", "02",
		"D", "2",
		// Day of week
		"dddd", "Monday",
		"ddd", "Mon",
		// Time (24h / 12h)
		"HH", "15",
		"hh", "03",
		"h", "3",
		"mm", "04",
		"m", "4",
		"ss", "05",
		"s", "5",
		"A", "PM",
		"a", "pm",
	)

	goFormat := replacer.Replace(format)

	return t.Format(goFormat)
}

// buildMemoViews 构建闪念视图数据
func (b *TemplateDataBuilder) buildMemoViews(ctx context.Context, config domain.ThemeConfig) []template.MemoView {
	if b.memoRepo == nil {
		return nil
	}

	memos, err := b.memoRepo.List(ctx)
	if err != nil {
		b.logger.Warn("获取闪念数据失败", "error", err)
		return nil
	}

	dateFormat := config.DateFormat
	if dateFormat == "" {
		dateFormat = "2006-01-02"
	}

	var views []template.MemoView
	for _, m := range memos {
		// 将 Markdown 内容转为 HTML
		htmlContent := utils.ToHTML(m.Content)

		// 格式化时间
		formatted := formatDate(m.CreatedAt, dateFormat)

		views = append(views, template.MemoView{
			ID:         m.ID,
			Content:    htmlTemplate.HTML(htmlContent),
			Tags:       m.Tags,
			CreatedAt:  formatted,
			DateFormat: formatted,
		})
	}
	return views
}

// toJSON 将 map 转换为 JSON 字符串（用于 JS 调用）
func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
