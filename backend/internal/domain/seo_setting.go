package domain

import "context"

type SeoSetting struct {
	EnableJsonLD            bool   `json:"enableJsonLD"`
	EnableOpenGraph         bool   `json:"enableOpenGraph"`
	EnableCanonicalURL      bool   `json:"enableCanonicalURL"`
	MetaKeywords            string `json:"metaKeywords"`
	GoogleAnalyticsID       string `json:"googleAnalyticsId"`
	GoogleSearchConsoleCode string `json:"googleSearchConsoleCode"`
	BaiduAnalyticsID        string `json:"baiduAnalyticsId"`
	CustomHeadCode          string `json:"customHeadCode"`
}

type SeoSettingRepository interface {
	GetSeoSetting(ctx context.Context) (SeoSetting, error)
	SaveSeoSetting(ctx context.Context, setting SeoSetting) error
}

// DefaultSeoSetting 返回首次使用时的 SEO 默认值——开箱即用的"合理全开"。
// 业界惯例是博客型站点默认启用 OpenGraph / JSON-LD / Canonical URL。
func DefaultSeoSetting() SeoSetting {
	return SeoSetting{
		EnableJsonLD:       true,
		EnableOpenGraph:    true,
		EnableCanonicalURL: true,
	}
}
