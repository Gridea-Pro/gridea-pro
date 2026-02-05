package model

// ThemeConfigItem 主题配置项
type ThemeConfigItem struct {
	Name    string              `json:"name"`              // 配置项名称
	Label   string              `json:"label"`             // 显示标签
	Group   string              `json:"group"`             // 分组
	Value   interface{}         `json:"value"`             // 默认值
	Type    string              `json:"type"`              // 类型: input/select/radio/switch/textarea
	Options []ThemeConfigOption `json:"options,omitempty"` // 选项（select/radio）
	Note    string              `json:"note,omitempty"`    // 说明
	Card    string              `json:"card,omitempty"`    // 卡片类型（如 color）
}

// ThemeConfigOption 配置选项
type ThemeConfigOption struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

// ThemeConfig 主题配置
type ThemeConfig struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Engine       string            `json:"engine"` // 模板引擎：ejs 或 go
	Author       string            `json:"author"`
	Repository   string            `json:"repository,omitempty"`
	CustomConfig []ThemeConfigItem `json:"customConfig"`
}
