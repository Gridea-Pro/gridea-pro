package domain

import "context"

type SeoSetting struct {
	// —— 基础 SEO ——
	MetaKeywords       string `json:"metaKeywords"`
	EnableCanonicalURL bool   `json:"enableCanonicalURL"`
	OgDefaultImage     string `json:"ogDefaultImage"` // 默认社交分享图（无则用站点头像）

	// —— 社交分享 / Open Graph ——
	EnableOpenGraph bool   `json:"enableOpenGraph"`
	TwitterSite     string `json:"twitterSite"`    // @username
	TwitterCreator  string `json:"twitterCreator"` // @username

	// —— 结构化数据 ——
	EnableJsonLD bool `json:"enableJsonLD"`

	// —— 站长平台验证 ——
	GoogleSearchConsoleCode string `json:"googleSearchConsoleCode"`
	BingVerificationCode    string `json:"bingVerificationCode"`
	BaiduVerificationCode   string `json:"baiduVerificationCode"`
	The360VerificationCode  string `json:"360VerificationCode"`
	YandexVerificationCode  string `json:"yandexVerificationCode"`

	// —— 网站分析统计 ——
	GoogleAnalyticsID           string `json:"googleAnalyticsId"`
	BaiduAnalyticsID            string `json:"baiduAnalyticsId"`
	PlausibleDomain             string `json:"plausibleDomain"`
	UmamiWebsiteID              string `json:"umamiWebsiteId"`
	UmamiScriptURL              string `json:"umamiScriptUrl"`
	CloudflareWebAnalyticsToken string `json:"cloudflareWebAnalyticsToken"`

	// —— 站点索引 ——
	SitemapEnabled bool   `json:"sitemapEnabled"`
	RobotsEnabled  bool   `json:"robotsEnabled"`
	RobotsCustom   string `json:"robotsCustom"` // 非空时替换默认 robots.txt 内容

	// —— 自定义代码注入 ——
	CustomHeadCode      string `json:"customHeadCode"`
	CustomBodyStartCode string `json:"customBodyStartCode"` // 注入 <body> 开头
	CustomBodyEndCode   string `json:"customBodyEndCode"`   // 注入 </body> 之前
}

type SeoSettingRepository interface {
	GetSeoSetting(ctx context.Context) (SeoSetting, error)
	SaveSeoSetting(ctx context.Context, setting SeoSetting) error
}

// DefaultSeoSetting 返回首次使用时的 SEO 默认值——开箱即用的"合理全开"。
// 业界惯例是博客型站点默认启用 OpenGraph / JSON-LD / Canonical URL，
// 同时默认生成 sitemap.xml 与 robots.txt。
func DefaultSeoSetting() SeoSetting {
	return SeoSetting{
		EnableJsonLD:       true,
		EnableOpenGraph:    true,
		EnableCanonicalURL: true,
		SitemapEnabled:     true,
		RobotsEnabled:      true,
	}
}
