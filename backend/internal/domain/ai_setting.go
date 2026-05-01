package domain

import "context"

// AIMode AI 模型使用模式
const (
	AIModeBuiltIn = "builtin" // 使用内置免费模型
	AIModeCustom  = "custom"  // 使用用户自己的 API Key
)

// AISetting AI 功能配置
type AISetting struct {
	Mode           string                    `json:"mode"`           // "builtin" | "custom"
	ActiveProvider string                    `json:"activeProvider"` // 当前激活的自定义厂商 ID
	Customs        map[string]AICustomConfig `json:"customs"`        // 各厂商的独立配置（key 为厂商 ID）
}

// AICustomConfig 单个自定义厂商的配置
// 注意：厂商 ID 由 AISetting.Customs 的 map key 表示，不在此结构中冗余存储
type AICustomConfig struct {
	Model   string `json:"model"`   // 模型 ID
	APIKey  string `json:"apiKey"`  // API Key
	BaseURL string `json:"baseURL"` // 自定义 OpenAI 兼容接口地址（如中转站）
}

// AISettingRepository AI 配置存储接口
type AISettingRepository interface {
	GetAISetting(ctx context.Context) (AISetting, error)
	SaveAISetting(ctx context.Context, setting AISetting) error
}
