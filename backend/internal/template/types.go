package template

import (
	"html/template"
	"time"
)

// TemplateData 主模板数据结构
type TemplateData struct {
	// 主题配置
	ThemeConfig ThemeConfigView `json:"themeConfig"`

	// 站点数据
	Site SiteView `json:"site"`

	// 文章列表（首页/标签页/归档页使用）
	Posts []PostView `json:"posts"`

	// 当前文章（文章详情页使用）
	Post PostView `json:"post"`

	// 菜单列表
	Menus []MenuView `json:"menus"`

	// 分页信息
	Pagination PaginationView `json:"pagination"`

	// 评论设置
	CommentSetting CommentSettingView `json:"commentSetting"`

	// 页面标题（供 includes/head 使用）
	SiteTitle string `json:"siteTitle"`

	// 当前标签（标签页使用）
	Tag TagView `json:"tag"`

	// 当前分类（分类页使用）
	Category CategoryView `json:"category"`

	// 所有标签
	Tags []TagView `json:"tags"`

	// 闪念列表（闪念页使用）
	Memos []MemoView `json:"memos"`

	// 归档数据（按年份分组，归档页使用）
	Archives []ArchiveYearView `json:"archives"`
}

// ThemeConfigView 主题配置视图
type ThemeConfigView struct {
	ThemeName        string `json:"themeName"`
	SiteName         string `json:"siteName"`
	SiteDescription  string `json:"siteDescription"`
	FooterInfo       string `json:"footerInfo"`
	Domain           string `json:"domain"`
	PostPageSize     int    `json:"postPageSize"`
	ArchivesPageSize int    `json:"archivesPageSize"`
	PostUrlFormat    string `json:"postUrlFormat"`
	TagUrlFormat     string `json:"tagUrlFormat"`
	DateFormat       string `json:"dateFormat"`
	Language         string `json:"language"`
	FeedEnabled      bool   `json:"feedEnabled"`
	FeedFullText     bool   `json:"feedFullText"`
	FeedCount        int    `json:"feedCount"`
	PostPath         string `json:"postPath"`
	TagPath          string `json:"tagPath"`
	TagsPath         string `json:"tagsPath"`
	LinkPath         string `json:"linkPath"`
	MemosPath        string `json:"memosPath"`
	ShowFeatureImage bool   `json:"showFeatureImage"`
}

// ArchiveYearView 归档年份分组视图
type ArchiveYearView struct {
	Year  int
	Posts []PostView
}

// SiteView 站点视图
type SiteView struct {
	// 自定义配置（从主题 config.json 读取的自定义字段）
	CustomConfig map[string]interface{} `json:"customConfig"`

	// 工具函数
	Utils SiteUtils `json:"utils"`
}

// SiteUtils 站点工具
type SiteUtils struct {
	// 当前时间戳
	Now int64 `json:"now"`
}

// SimplePostView 文章简要视图（用于上下文引用）
type SimplePostView struct {
	Title    string `json:"title"`
	Link     string `json:"link"`
	FileName string `json:"fileName"`
	Feature  string `json:"feature"`
}

// PostView 文章视图
type PostView struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	FileName        string          `json:"fileName"`
	Content         template.HTML   `json:"content"`  // HTML 内容，不会被转义
	Abstract        template.HTML   `json:"abstract"` // HTML 摘要，不会被转义
	Description     string          `json:"description"`
	Link            string          `json:"link"`
	Feature         string          `json:"feature"`
	CreatedAt       time.Time       `json:"createdAt"`
	Date            time.Time       `json:"date"` // 永远为老主题保留 date 字段
	DateFormat      string          `json:"dateFormat"`
	UpdatedAt       time.Time       `json:"updatedAt"`       // 最后修改时间
	UpdatedAtFormat string          `json:"updatedAtFormat"` // 格式化后的修改时间
	Published       bool            `json:"published"`
	HideInList      bool            `json:"hideInList"`
	IsTop           bool            `json:"isTop"`
	Tags            []TagView       `json:"tags"`
	Categories      []CategoryView  `json:"categories"`
	TagsString      string          `json:"tagsString"` // 标签逗号分隔字符串
	Stats           PostStats       `json:"stats"`
	Toc             template.HTML   `json:"toc"` // 目录 HTML，不会被转义
	NextPost        *SimplePostView `json:"nextPost"`
	PrevPost        *SimplePostView `json:"prevPost"`
}

// PostStats 文章统计
type PostStats struct {
	Words   int    `json:"words"`
	Minutes int    `json:"minutes"`
	Text    string `json:"text"` // "5 min read"
}

// TagView 标签视图
type TagView struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Link     string `json:"link"`
	Count    int    `json:"count"`
	UsedName string `json:"usedName"` // 兼容旧版
}

// MemoView 闪念视图
type MemoView struct {
	ID         string        `json:"id"`
	Content    template.HTML `json:"content"` // HTML 内容
	Tags       []string      `json:"tags"`
	CreatedAt  string        `json:"createdAt"`
	DateFormat string        `json:"dateFormat"`
}

// CategoryView 分类视图
type CategoryView struct {
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Link  string `json:"link"`
	Count int    `json:"count"`
}

// MenuView 菜单视图
type MenuView struct {
	Name     string     `json:"name"`
	Link     string     `json:"link"`
	OpenType string     `json:"openType"` // "Internal" 或 "External"
	Children []MenuView `json:"children,omitempty"`
}

// PaginationView 分页视图
type PaginationView struct {
	CurrentPage int    `json:"currentPage"` // 当前页码（从 1 开始）
	TotalPages  int    `json:"totalPages"`  // 总页数
	TotalPosts  int    `json:"totalPosts"`  // 文章总数
	HasPrev     bool   `json:"hasPrev"`     // 是否有上一页
	HasNext     bool   `json:"hasNext"`     // 是否有下一页
	PrevURL     string `json:"prevURL"`     // 上一页 URL
	NextURL     string `json:"nextURL"`     // 下一页 URL
	Prev        string `json:"prev"`        // 上一页 URL（别名，兼容旧主题）
	Next        string `json:"next"`        // 下一页 URL（别名，兼容旧主题）
}

// CommentSettingView 评论设置视图（用于模板渲染）
type CommentSettingView struct {
	ShowComment bool   `json:"showComment"`
	Platform    string `json:"commentPlatform"`

	// Valine/Waline 配置
	AppID      string `json:"appId"`
	AppKey     string `json:"appKey"`
	ServerURLs string `json:"serverURLs"`

	// Twikoo 配置
	EnvID string `json:"envId"`

	// Gitalk 配置
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Repo         string `json:"repo"`
	Owner        string `json:"owner"`
	Admin        string `json:"admin"`

	// Giscus 配置
	RepoID     string `json:"repoId"`
	Category   string `json:"category"`
	CategoryID string `json:"categoryId"`

	// Disqus 配置
	Shortname string `json:"shortname"`
	API       string `json:"api"`
	APIKey    string `json:"apiKey"`

	// Cusdis 配置
	Host string `json:"host"`
}

// NewSiteUtils 创建站点工具实例
func NewSiteUtils() SiteUtils {
	return SiteUtils{
		Now: time.Now().Unix(),
	}
}
