package engine

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gridea-pro/backend/internal/template"
)

// SeoGenerator 负责生成 SEO 相关文件（RSS、Sitemap、Robots.txt）
type SeoGenerator struct {
	logger *slog.Logger
}

// NewSeoGenerator 创建 SeoGenerator
func NewSeoGenerator() *SeoGenerator {
	return &SeoGenerator{
		logger: slog.Default(),
	}
}

// RenderRobotsTxt 自动生成 robots.txt。
// custom 非空时直接以其作为完整内容（用户接管 robots.txt 全部内容）；为空时使用默认模板。
func (g *SeoGenerator) RenderRobotsTxt(buildDir string, data *template.TemplateData, custom string) error {
	var content string
	if strings.TrimSpace(custom) != "" {
		content = strings.TrimRight(custom, "\n") + "\n"
	} else {
		var sb strings.Builder
		sb.WriteString("User-agent: *\n")
		sb.WriteString("Allow: /\n")
		if domainUrl := strings.TrimRight(data.ThemeConfig.Domain, "/"); domainUrl != "" {
			sb.WriteString(fmt.Sprintf("\nSitemap: %s/sitemap.xml\n", domainUrl))
		}
		content = sb.String()
	}
	return os.WriteFile(filepath.Join(buildDir, "robots.txt"), []byte(content), 0644)
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
	needsParse := false
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if c >= 0x80 || c == ' ' || c == '"' || c == '<' || c == '>' || c == '\\' || c == '^' || c == '`' || c == '{' || c == '|' || c == '}' {
			needsParse = true
			break
		}
	}
	if !needsParse {
		return raw
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return parsed.String()
}

// RenderRSS 渲染 RSS 订阅 (feed.xml, RSS 2.0 规范)
func (g *SeoGenerator) RenderRSS(buildDir string, data *template.TemplateData) error {
	domainUrl := strings.TrimRight(data.ThemeConfig.Domain, "/")
	if domainUrl == "" {
		g.logger.Error("警告：未配置域名，RSS (feed.xml) 中的链接可能无效")
	}

	lastBuild := time.Now().Format(time.RFC1123Z)
	if len(data.Posts) > 0 {
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
		feedCount = 20
	}

	count := 0
	listPosts := getVisiblePosts(data.Posts)
	for _, post := range listPosts {
		if count >= feedCount {
			break
		}

		content := string(post.Content)
		if !data.ThemeConfig.FeedFullText {
			if string(post.Abstract) != "" {
				content = string(post.Abstract)
			} else {
				runes := []rune(content)
				if len(runes) > 200 {
					content = string(runes[:200]) + "..."
				}
			}
		}

		link := domainUrl + post.Link
		if domainUrl == "" {
			link = post.Link
		}

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
			PubDate:     post.Date.Format(time.RFC1123Z),
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

	g.logger.Info(fmt.Sprintf("✅ RSS (feed.xml) 生成成功 (%d 篇文章)", len(feed.Channel.Items)))
	return os.WriteFile(filepath.Join(buildDir, "feed.xml"), finalOutput, 0644)
}

// RenderSitemap 渲染站点地图 (sitemap.xml)
func (g *SeoGenerator) RenderSitemap(buildDir string, data *template.TemplateData) error {
	domainUrl := strings.TrimRight(data.ThemeConfig.Domain, "/")
	if domainUrl == "" {
		g.logger.Error("警告：未配置域名，Sitemap (sitemap.xml) 中的链接可能无效")
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
	listPosts := getVisiblePosts(data.Posts)
	for _, post := range listPosts {
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
			LastMod: post.UpdatedAt.Format("2006-01-02T15:04:05-07:00"),
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
			LastMod: nowDate,
		})
	}

	// 5. 归档页
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

	g.logger.Info(fmt.Sprintf("✅ Sitemap (sitemap.xml) 生成成功 (%d 个链接)", len(urlset.Urls)))
	return os.WriteFile(filepath.Join(buildDir, "sitemap.xml"), finalOutput, 0644)
}
