package domain

import "context"

// Theme 主题结构
type Theme struct {
	Folder       string        `json:"folder"`
	Name         string        `json:"name"`
	Version      string        `json:"version"`
	Description  string        `json:"description,omitempty"`
	Author       string        `json:"author,omitempty"`
	Repository   string        `json:"repository,omitempty"`
	PreviewImage string        `json:"previewImage,omitempty"`
	CustomConfig []interface{} `json:"customConfig,omitempty"`
}

// ThemeConfig 主题配置
type ThemeConfig struct {
	ThemeName        string                 `json:"themeName"`
	PostPageSize     int                    `json:"postPageSize"`
	ArchivesPageSize int                    `json:"archivesPageSize"`
	SiteName         string                 `json:"siteName"`
	SiteAuthor       string                 `json:"siteAuthor"`
	SiteEmail        string                 `json:"siteEmail"`
	SiteDescription  string                 `json:"siteDescription"`
	FooterInfo       string                 `json:"footerInfo"`
	ShowFeatureImage bool                   `json:"showFeatureImage"`
	Domain           string                 `json:"domain"`
	PostUrlFormat    string                 `json:"postUrlFormat"`
	TagUrlFormat     string                 `json:"tagUrlFormat"`
	DateFormat       string                 `json:"dateFormat"`
	FeedFullText     bool                   `json:"feedFullText"`
	FeedCount        int                    `json:"feedCount"`
	ArchivesPath     string                 `json:"archivesPath"`
	PostPath         string                 `json:"postPath"`
	TagPath          string                 `json:"tagPath"`
	TagsPath         string                 `json:"tagsPath"`
	LinkPath         string                 `json:"linkPath"`
	MemosPath        string                 `json:"memosPath"`
	CustomConfig     map[string]interface{} `json:"customConfig,omitempty"`
}

// ThemeRepository 定义主题存储接口
type ThemeRepository interface {
	// GetAll 获取已安装主题列表
	GetAll(ctx context.Context) ([]Theme, error)
	// GetConfig 获取当前主题配置
	GetConfig(ctx context.Context) (ThemeConfig, error)
	// SaveConfig 保存主题配置
	SaveConfig(ctx context.Context, config ThemeConfig) error
}
