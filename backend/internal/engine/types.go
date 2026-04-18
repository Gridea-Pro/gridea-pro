package engine

import (
	"encoding/xml"
	"gridea-pro/backend/internal/template"
)

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
	XMLName    xml.Name      `xml:"url"`
	Loc        string        `xml:"loc"`
	LastMod    string        `xml:"lastmod"`
	ChangeFreq string        `xml:"changefreq,omitempty"`
	Priority   string        `xml:"priority,omitempty"`
	Image      *sitemapImage `xml:"image:image,omitempty"`
}

// sitemapURLSet Sitemap 根元素
type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	ImageNs string       `xml:"xmlns:image,attr"`
	Urls    []sitemapURL `xml:"url"`
}

// searchEntry 搜索索引条目
type searchEntry struct {
	Title   string   `json:"title"`
	Link    string   `json:"link"`
	Date    string   `json:"date"`
	Tags    []string `json:"tags,omitempty"`
	Content string   `json:"content"`
}

// getVisiblePosts 过滤出需要在列表（如首页、归档、RSS）中展示的文章
func getVisiblePosts(posts []template.PostView) []template.PostView {
	var list []template.PostView
	for _, p := range posts {
		if !p.HideInList && p.Published {
			list = append(list, p)
		}
	}
	return list
}
