package render

import "gridea-pro/backend/internal/template"

// ThemeRenderer 主题渲染器接口
// 定义统一的渲染行为,支持多种模板引擎实现
type ThemeRenderer interface {
	// Render 渲染指定模板
	// templateName: 模板名称(不含扩展名),如 "index", "post"
	// data: 模板数据,统一使用 TemplateData 结构
	// 返回: 渲染后的 HTML 字符串和可能的错误
	Render(templateName string, data *template.TemplateData) (string, error)

	// GetEngineType 获取引擎类型
	// 返回: "gotemplate" 或 "ejs"
	GetEngineType() string

	// ClearCache 清除模板缓存
	// 用于开发模式下热重载
	ClearCache()
}

// RenderConfig 渲染器配置
type RenderConfig struct {
	AppDir    string // 应用根目录
	ThemeName string // 主题名称
	CacheDir  string // 缓存目录
}
