package engine

import (
	"encoding/json"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/template"
	"gridea-pro/backend/internal/version"
	"strings"
)

// HtmlPostProcessor 在模板渲染完成后对 HTML 进行后处理，注入 SEO 标签、PWA 标签和 CDN URL 重写
type HtmlPostProcessor struct {
	seoSetting *domain.SeoSetting
	cdnSetting *domain.CdnSetting
	pwaSetting *domain.PwaSetting
	domain     string
	siteName   string
	siteDesc   string
	language   string
	avatar     string // 站点头像 URL，用作 og:image / JSON-LD logo 的兜底
}

// NewHtmlPostProcessor 创建后处理器
func NewHtmlPostProcessor(seo *domain.SeoSetting, cdn *domain.CdnSetting, pwa *domain.PwaSetting, domain, siteName, siteDesc, language, avatar string) *HtmlPostProcessor {
	return &HtmlPostProcessor{
		seoSetting: seo,
		cdnSetting: cdn,
		pwaSetting: pwa,
		domain:     domain,
		siteName:   siteName,
		siteDesc:   siteDesc,
		language:   language,
		avatar:     avatar,
	}
}

// Process 对渲染后的 HTML 进行后处理
// pageType: "index", "post", "tags", "tag", "archives", "links", "memos", "404", "blog", "category"
// pageURL: 当前页面的相对 URL（如 "/" 、"/post/hello/"）
// post: 文章页时传入当前文章数据，其他页传 nil
func (p *HtmlPostProcessor) Process(html, pageType, pageURL string, post *template.PostView) string {
	if p == nil {
		return html
	}

	// Generator meta 注入（无开关，始终注入）
	html = p.injectGenerator(html)

	// SEO 注入
	html = p.injectSeo(html, pageType, pageURL, post)

	// PWA meta 标签注入
	html = p.injectPwa(html)

	// 自定义 body 代码注入
	html = p.injectBody(html)

	// CDN URL 重写
	html = p.rewriteCdnURLs(html)

	return html
}

// injectGenerator 插入 <meta name="generator">，标识 Gridea Pro 渲染。
// 优先插到 <title> 之前（紧邻 charset/viewport 等元信息的位置）；无 <title> 时
// 兜底插到 </head> 前。
func (p *HtmlPostProcessor) injectGenerator(html string) string {
	tag := fmt.Sprintf(`<meta name="generator" content="%s">`, escapeAttr(version.Generator()))
	lowerHTML := strings.ToLower(html)

	if idx := strings.Index(lowerHTML, "<title"); idx != -1 {
		// 复用 <title> 所在行的缩进，避免 <title> 被挤到行首
		lineStart := strings.LastIndex(html[:idx], "\n") + 1
		indent := html[lineStart:idx]
		if strings.TrimLeft(indent, " \t") != "" {
			indent = "" // 非纯空白（title 与其他内容同行）则不复用
		}
		return html[:idx] + tag + "\n" + indent + html[idx:]
	}

	if idx := strings.LastIndex(lowerHTML, "</head>"); idx != -1 {
		return html[:idx] + tag + "\n" + html[idx:]
	}

	return html
}

// injectSeo 在 </head> 前插入 SEO 相关标签
func (p *HtmlPostProcessor) injectSeo(html, pageType, pageURL string, post *template.PostView) string {
	if p.seoSetting == nil {
		return html
	}

	var sb strings.Builder

	fullURL := strings.TrimRight(p.domain, "/") + pageURL

	// Meta Description：post.Description → post.Abstract → 内容前 160 字 → 站点描述
	description := p.siteDesc
	if post != nil && pageType == "post" {
		switch {
		case post.Description != "":
			description = post.Description
		case post.Abstract != "":
			if plain := strings.TrimSpace(stripHTML(string(post.Abstract))); plain != "" {
				description = plain
			}
		default:
			plain := stripHTML(string(post.Content))
			if len([]rune(plain)) > 160 {
				description = string([]rune(plain)[:160])
			} else {
				description = plain
			}
		}
	}
	if description != "" {
		sb.WriteString(fmt.Sprintf(`<meta name="description" content="%s">`+"\n", escapeAttr(description)))
	}

	// Meta Keywords
	if p.seoSetting.MetaKeywords != "" {
		sb.WriteString(fmt.Sprintf(`<meta name="keywords" content="%s">`+"\n", escapeAttr(p.seoSetting.MetaKeywords)))
	}

	// Canonical URL
	if p.seoSetting.EnableCanonicalURL && fullURL != "" {
		sb.WriteString(fmt.Sprintf(`<link rel="canonical" href="%s">`+"\n", escapeAttr(fullURL)))
	}

	// JSON-LD
	if p.seoSetting.EnableJsonLD {
		sb.WriteString(p.buildJsonLD(pageType, fullURL, post, description))
	}

	// Open Graph + Twitter Card
	if p.seoSetting.EnableOpenGraph {
		sb.WriteString(p.buildOpenGraph(pageType, fullURL, post, description))
	}

	// Google Analytics
	if p.seoSetting.GoogleAnalyticsID != "" {
		gaID := escapeAttr(p.seoSetting.GoogleAnalyticsID)
		sb.WriteString(fmt.Sprintf(`<script async src="https://www.googletagmanager.com/gtag/js?id=%s"></script>`+"\n", gaID))
		sb.WriteString(fmt.Sprintf(`<script>window.dataLayer=window.dataLayer||[];function gtag(){dataLayer.push(arguments);}gtag('js',new Date());gtag('config','%s');</script>`+"\n", gaID))
	}

	// 站长平台验证 meta
	if p.seoSetting.GoogleSearchConsoleCode != "" {
		sb.WriteString(fmt.Sprintf(`<meta name="google-site-verification" content="%s">`+"\n", escapeAttr(p.seoSetting.GoogleSearchConsoleCode)))
	}
	if p.seoSetting.BingVerificationCode != "" {
		sb.WriteString(fmt.Sprintf(`<meta name="msvalidate.01" content="%s">`+"\n", escapeAttr(p.seoSetting.BingVerificationCode)))
	}
	if p.seoSetting.BaiduVerificationCode != "" {
		sb.WriteString(fmt.Sprintf(`<meta name="baidu-site-verification" content="%s">`+"\n", escapeAttr(p.seoSetting.BaiduVerificationCode)))
	}
	if p.seoSetting.The360VerificationCode != "" {
		sb.WriteString(fmt.Sprintf(`<meta name="360-site-verification" content="%s">`+"\n", escapeAttr(p.seoSetting.The360VerificationCode)))
	}
	if p.seoSetting.YandexVerificationCode != "" {
		sb.WriteString(fmt.Sprintf(`<meta name="yandex-verification" content="%s">`+"\n", escapeAttr(p.seoSetting.YandexVerificationCode)))
	}

	// 百度统计
	if p.seoSetting.BaiduAnalyticsID != "" {
		baiduID := escapeAttr(p.seoSetting.BaiduAnalyticsID)
		sb.WriteString(fmt.Sprintf(`<script>var _hmt=_hmt||[];(function(){var hm=document.createElement("script");hm.src="https://hm.baidu.com/hm.js?%s";var s=document.getElementsByTagName("script")[0];s.parentNode.insertBefore(hm,s);})();</script>`+"\n", baiduID))
	}

	// Plausible Analytics
	if p.seoSetting.PlausibleDomain != "" {
		sb.WriteString(fmt.Sprintf(`<script defer data-domain="%s" src="https://plausible.io/js/script.js"></script>`+"\n", escapeAttr(p.seoSetting.PlausibleDomain)))
	}

	// Umami Analytics
	if p.seoSetting.UmamiWebsiteID != "" && p.seoSetting.UmamiScriptURL != "" {
		sb.WriteString(fmt.Sprintf(`<script defer src="%s" data-website-id="%s"></script>`+"\n",
			escapeAttr(p.seoSetting.UmamiScriptURL), escapeAttr(p.seoSetting.UmamiWebsiteID)))
	}

	// Cloudflare Web Analytics
	if p.seoSetting.CloudflareWebAnalyticsToken != "" {
		token := escapeAttr(p.seoSetting.CloudflareWebAnalyticsToken)
		sb.WriteString(fmt.Sprintf(`<script defer src="https://static.cloudflareinsights.com/beacon.min.js" data-cf-beacon='{"token":"%s"}'></script>`+"\n", token))
	}

	// 自定义 head 代码
	if p.seoSetting.CustomHeadCode != "" {
		sb.WriteString(p.seoSetting.CustomHeadCode + "\n")
	}

	inject := sb.String()
	if inject == "" {
		return html
	}

	// 在 </head> 前插入
	idx := strings.LastIndex(strings.ToLower(html), "</head>")
	if idx == -1 {
		return html
	}
	return html[:idx] + inject + html[idx:]
}

// orderedJSON 按给定顺序生成 JSON 字符串（保持属性语义排序，而非字母序）
func orderedJSON(pairs ...any) string {
	var sb strings.Builder
	sb.WriteByte('{')
	for i := 0; i < len(pairs)-1; i += 2 {
		if i > 0 {
			sb.WriteByte(',')
		}
		key, _ := json.Marshal(pairs[i])
		val, _ := json.Marshal(pairs[i+1])
		sb.Write(key)
		sb.WriteByte(':')
		sb.Write(val)
	}
	sb.WriteByte('}')
	return sb.String()
}

// buildJsonLD 生成 JSON-LD 结构化数据
func (p *HtmlPostProcessor) buildJsonLD(pageType, fullURL string, post *template.PostView, description string) string {
	var sb strings.Builder

	switch pageType {
	case "post":
		if post == nil {
			break
		}
		// BlogPosting: @context → @type → headline → url → description → image → datePublished → dateModified → author → publisher
		pairs := []any{
			"@context", "https://schema.org",
			"@type", "BlogPosting",
			"headline", post.Title,
			"url", fullURL,
			"description", description,
		}
		if post.Feature != "" {
			pairs = append(pairs, "image", post.Feature)
		}
		publisher := map[string]any{"@type": "Organization", "name": p.siteName}
		if p.avatar != "" {
			publisher["logo"] = map[string]any{"@type": "ImageObject", "url": p.avatar}
		}
		pairs = append(pairs,
			"datePublished", post.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"dateModified", post.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"author", map[string]any{"@type": "Person", "name": p.siteName},
			"publisher", publisher,
		)
		fmt.Fprintf(&sb, "<script type=\"application/ld+json\">%s</script>\n", orderedJSON(pairs...))

		// BreadcrumbList
		breadcrumb := orderedJSON(
			"@context", "https://schema.org",
			"@type", "BreadcrumbList",
			"itemListElement", []map[string]any{
				{"@type": "ListItem", "position": 1, "name": p.siteName, "item": p.domain},
				{"@type": "ListItem", "position": 2, "name": post.Title, "item": fullURL},
			},
		)
		fmt.Fprintf(&sb, "<script type=\"application/ld+json\">%s</script>\n", breadcrumb)

	default:
		// WebSite: @context → @type → name → url → description → potentialAction
		website := orderedJSON(
			"@context", "https://schema.org",
			"@type", "WebSite",
			"name", p.siteName,
			"url", p.domain,
			"description", p.siteDesc,
			"potentialAction", map[string]any{
				"@type":       "SearchAction",
				"target":      strings.TrimRight(p.domain, "/") + "/search?q={search_term_string}",
				"query-input": "required name=search_term_string",
			},
		)
		fmt.Fprintf(&sb, "<script type=\"application/ld+json\">%s</script>\n", website)
	}

	return sb.String()
}

// buildOpenGraph 生成 Open Graph 和 Twitter Card 标签
func (p *HtmlPostProcessor) buildOpenGraph(pageType, fullURL string, post *template.PostView, description string) string {
	var sb strings.Builder

	title := p.siteName
	ogType := "website"
	// 图片优先级：文章特色图 → SEO 设置默认图 → 站点头像
	image := ""
	if p.seoSetting != nil && p.seoSetting.OgDefaultImage != "" {
		image = p.seoSetting.OgDefaultImage
	} else if p.avatar != "" {
		image = p.avatar
	}

	if pageType == "post" && post != nil {
		title = post.Title
		ogType = "article"
		if post.Feature != "" {
			image = post.Feature
		}
	}

	sb.WriteString(fmt.Sprintf(`<meta property="og:title" content="%s">`+"\n", escapeAttr(title)))
	sb.WriteString(fmt.Sprintf(`<meta property="og:description" content="%s">`+"\n", escapeAttr(description)))
	sb.WriteString(fmt.Sprintf(`<meta property="og:url" content="%s">`+"\n", escapeAttr(fullURL)))
	sb.WriteString(fmt.Sprintf(`<meta property="og:type" content="%s">`+"\n", ogType))
	sb.WriteString(fmt.Sprintf(`<meta property="og:site_name" content="%s">`+"\n", escapeAttr(p.siteName)))
	if image != "" {
		sb.WriteString(fmt.Sprintf(`<meta property="og:image" content="%s">`+"\n", escapeAttr(image)))
	}

	// Twitter Card
	sb.WriteString(`<meta name="twitter:card" content="summary_large_image">` + "\n")
	sb.WriteString(fmt.Sprintf(`<meta name="twitter:url" content="%s">`+"\n", escapeAttr(fullURL)))
	sb.WriteString(fmt.Sprintf(`<meta name="twitter:title" content="%s">`+"\n", escapeAttr(title)))
	sb.WriteString(fmt.Sprintf(`<meta name="twitter:description" content="%s">`+"\n", escapeAttr(description)))
	if image != "" {
		sb.WriteString(fmt.Sprintf(`<meta name="twitter:image" content="%s">`+"\n", escapeAttr(image)))
	}
	if p.seoSetting != nil && p.seoSetting.TwitterSite != "" {
		sb.WriteString(fmt.Sprintf(`<meta name="twitter:site" content="%s">`+"\n", escapeAttr(p.seoSetting.TwitterSite)))
	}
	if p.seoSetting != nil && p.seoSetting.TwitterCreator != "" {
		sb.WriteString(fmt.Sprintf(`<meta name="twitter:creator" content="%s">`+"\n", escapeAttr(p.seoSetting.TwitterCreator)))
	}

	return sb.String()
}

// injectBody 在 <body> 之后与 </body> 之前注入用户自定义代码（如 GTM、客服挂件等）
func (p *HtmlPostProcessor) injectBody(html string) string {
	if p.seoSetting == nil {
		return html
	}

	// </body> 之前注入 customBodyEndCode（先注入这个，让 startCode 后注入时位置不受 endCode 偏移影响）
	if p.seoSetting.CustomBodyEndCode != "" {
		idx := strings.LastIndex(strings.ToLower(html), "</body>")
		if idx != -1 {
			html = html[:idx] + p.seoSetting.CustomBodyEndCode + "\n" + html[idx:]
		}
	}

	// <body[...]> 之后注入 customBodyStartCode
	if p.seoSetting.CustomBodyStartCode != "" {
		lowerHTML := strings.ToLower(html)
		// 找开 <body 标签的位置；body 可能带属性，需要找 > 关闭
		if start := strings.Index(lowerHTML, "<body"); start != -1 {
			if end := strings.Index(lowerHTML[start:], ">"); end != -1 {
				insertAt := start + end + 1
				html = html[:insertAt] + "\n" + p.seoSetting.CustomBodyStartCode + html[insertAt:]
			}
		}
	}

	return html
}

// rewriteCdnURLs 将静态资源路径替换为 CDN 地址
func (p *HtmlPostProcessor) rewriteCdnURLs(html string) string {
	if p.cdnSetting == nil || !p.cdnSetting.Enabled {
		return html
	}

	cdnBase := p.buildCdnBase()
	if cdnBase == "" {
		return html
	}

	// 需要替换的路径前缀
	paths := []string{"/images/", "/post-images/", "/media/"}
	// 需要替换的 HTML 属性上下文
	prefixes := []string{`src="`, `src='`, `href="`, `href='`}

	for _, path := range paths {
		for _, prefix := range prefixes {
			old := prefix + path
			newVal := prefix + cdnBase + path
			html = strings.ReplaceAll(html, old, newVal)
		}
	}

	return html
}

// buildCdnBase 构建 CDN 基础 URL
func (p *HtmlPostProcessor) buildCdnBase() string {
	switch p.cdnSetting.Provider {
	case "jsdelivr":
		if p.cdnSetting.GithubUser == "" || p.cdnSetting.GithubRepo == "" {
			return ""
		}
		branch := p.cdnSetting.GithubBranch
		if branch == "" {
			branch = "main"
		}
		return fmt.Sprintf("https://cdn.jsdelivr.net/gh/%s/%s@%s",
			p.cdnSetting.GithubUser, p.cdnSetting.GithubRepo, branch)
	case "custom":
		return strings.TrimRight(p.cdnSetting.BaseURL, "/")
	default:
		return ""
	}
}

// escapeAttr 转义 HTML 属性值中的特殊字符
func escapeAttr(s string) string {
	s = strings.ReplaceAll(s, `&`, `&amp;`)
	s = strings.ReplaceAll(s, `"`, `&quot;`)
	s = strings.ReplaceAll(s, `<`, `&lt;`)
	s = strings.ReplaceAll(s, `>`, `&gt;`)
	return s
}

// injectPwa 在 </head> 前插入 PWA 相关标签
func (p *HtmlPostProcessor) injectPwa(html string) string {
	if p.pwaSetting == nil || !p.pwaSetting.Enabled {
		return html
	}

	var sb strings.Builder

	sb.WriteString(`<link rel="manifest" href="/manifest.json">` + "\n")

	if p.pwaSetting.ThemeColor != "" {
		sb.WriteString(fmt.Sprintf(`<meta name="theme-color" content="%s">`+"\n", escapeAttr(p.pwaSetting.ThemeColor)))
	}

	sb.WriteString(`<meta name="apple-mobile-web-app-capable" content="yes">` + "\n")
	sb.WriteString(`<meta name="mobile-web-app-capable" content="yes">` + "\n")
	sb.WriteString(`<meta name="apple-mobile-web-app-status-bar-style" content="default">` + "\n")

	appName := p.pwaSetting.AppName
	if appName == "" {
		appName = p.siteName
	}
	sb.WriteString(fmt.Sprintf(`<meta name="apple-mobile-web-app-title" content="%s">`+"\n", escapeAttr(appName)))

	// apple-touch-icon 180x180（iOS 标准尺寸）
	icon180 := "/images/icons/icon-180x180.png"
	sb.WriteString(fmt.Sprintf(`<link rel="apple-touch-icon" sizes="180x180" href="%s">`+"\n", escapeAttr(icon180)))

	// 标准 favicon 192x192（构建时统一生成到固定路径）
	icon192 := "/images/icons/icon-192x192.png"
	sb.WriteString(fmt.Sprintf(`<link rel="icon" type="image/png" sizes="192x192" href="%s">`+"\n", escapeAttr(icon192)))

	// Windows 磁贴
	if p.pwaSetting.ThemeColor != "" {
		sb.WriteString(fmt.Sprintf(`<meta name="msapplication-TileColor" content="%s">`+"\n", escapeAttr(p.pwaSetting.ThemeColor)))
	}
	sb.WriteString(fmt.Sprintf(`<meta name="msapplication-TileImage" content="%s">`+"\n", escapeAttr(icon192)))

	sb.WriteString(`<script>if('serviceWorker' in navigator){navigator.serviceWorker.register('/sw.js')}</script>` + "\n")

	inject := sb.String()
	idx := strings.LastIndex(strings.ToLower(html), "</head>")
	if idx == -1 {
		return html
	}
	return html[:idx] + inject + html[idx:]
}

// stripHTML 简单去除 HTML 标签
func stripHTML(s string) string {
	var sb strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			sb.WriteRune(r)
		}
	}
	result := sb.String()
	// 压缩连续空白
	result = strings.Join(strings.Fields(result), " ")
	return strings.TrimSpace(result)
}
