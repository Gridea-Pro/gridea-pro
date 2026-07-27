package domain

// Added Validate() method.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

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
	ThemeName        string `json:"themeName"`
	PostPageSize     int    `json:"postPageSize"`
	ArchivesPageSize int    `json:"archivesPageSize"`
	SiteName         string `json:"siteName"`
	SiteAuthor       string `json:"siteAuthor"`
	SiteEmail        string `json:"siteEmail"`
	SiteDescription  string `json:"siteDescription"`
	FooterInfo       string `json:"footerInfo"`
	Domain           string `json:"domain"`
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
	// KatexEnabled 控制是否在站点渲染产物里自动注入 KaTeX 公式样式。
	// 关闭后即使文章里有 `$...$` 公式，主题不会被注入 katex.min.css，
	// 公式样式由主题自己负责（或不显示）。默认开。
	//
	// 兼容性：老配置文件没有这个字段，UnmarshalJSON 会把它默认成 true，
	// 升级后不需要用户重新打开开关。
	KatexEnabled bool                   `json:"katexEnabled"`
	CustomConfig map[string]interface{} `json:"customConfig,omitempty"`
}

// UnmarshalJSON 自定义反序列化：对老配置（缺 katexEnabled 字段）默认开 KaTeX。
// 这样从 v1.2.0 之前升级过来的站点不需要手动到主题设置里打开开关。
func (c *ThemeConfig) UnmarshalJSON(data []byte) error {
	type alias ThemeConfig
	if err := json.Unmarshal(data, (*alias)(c)); err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err == nil {
		if _, ok := raw["katexEnabled"]; !ok {
			c.KatexEnabled = true
		}
	}
	return nil
}

// Validate 校验主题配置
func (c *ThemeConfig) Validate() error {
	if strings.TrimSpace(c.ThemeName) == "" {
		return errors.New("theme name is required")
	}
	return nil
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
