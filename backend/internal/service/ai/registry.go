package ai

// Protocol 厂商使用的协议类型
type Protocol string

const (
	ProtocolOpenAI    Protocol = "openai"    // OpenAI 兼容（多数厂商）
	ProtocolAnthropic Protocol = "anthropic" // Anthropic /v1/messages
	ProtocolGemini    Protocol = "gemini"    // Google Gemini
)

// ProviderInfo 一个厂商的元信息
type ProviderInfo struct {
	ID            string   `json:"id"`            // 厂商唯一 ID
	Name          string   `json:"name"`          // 显示名（中文/英文均可）
	Protocol      Protocol `json:"protocol"`      // 协议类型
	BaseURL       string   `json:"baseURL"`       // API 端点
	DefaultModels []string `json:"defaultModels"` // 精选默认模型
	APIKeyURL     string   `json:"apiKeyURL"`     // 获取 API Key 的页面
}

// providerRegistry 13 家自定义厂商的注册表
// 顺序即前端下拉框展示顺序
//
// 排序原则：
//  1. 国际主流（OpenAI / Anthropic / Google / xAI）—— 用户认知度最高
//  2. 国内主流（DeepSeek / 智谱 / Kimi / 通义 / 豆包 / 小米）—— 中国用户无需代理
//  3. 特殊用途（OpenRouter 聚合 / Mistral 欧洲 / Groq 速度特化）—— 进阶选项
var providerRegistry = []ProviderInfo{
	// ─── 国际主流 ──────────────────────────────────────
	{
		ID:       "openai",
		Name:     "OpenAI",
		Protocol: ProtocolOpenAI,
		BaseURL:  "https://api.openai.com/v1",
		DefaultModels: []string{
			"gpt-4o",
			"gpt-4o-mini",
			"gpt-4-turbo",
			"o1-mini",
		},
		APIKeyURL: "https://platform.openai.com/api-keys",
	},
	{
		ID:       "anthropic",
		Name:     "Anthropic",
		Protocol: ProtocolAnthropic,
		BaseURL:  "https://api.anthropic.com",
		DefaultModels: []string{
			"claude-opus-4-6",
			"claude-sonnet-4-6",
			"claude-haiku-4-5",
		},
		APIKeyURL: "https://console.anthropic.com/settings/keys",
	},
	{
		ID:       "gemini",
		Name:     "Google Gemini",
		Protocol: ProtocolGemini,
		BaseURL:  "https://generativelanguage.googleapis.com/v1beta",
		DefaultModels: []string{
			"gemini-2.0-flash",
			"gemini-1.5-pro",
			"gemini-1.5-flash",
		},
		APIKeyURL: "https://aistudio.google.com/apikey",
	},
	{
		ID:       "xai",
		Name:     "xAI (Grok)",
		Protocol: ProtocolOpenAI,
		BaseURL:  "https://api.x.ai/v1",
		DefaultModels: []string{
			"grok-2",
			"grok-2-mini",
			"grok-beta",
		},
		APIKeyURL: "https://console.x.ai",
	},
	// ─── 国内主流 ──────────────────────────────────────
	{
		ID:       "deepseek",
		Name:     "DeepSeek",
		Protocol: ProtocolOpenAI,
		BaseURL:  "https://api.deepseek.com/v1",
		DefaultModels: []string{
			"deepseek-chat",
			"deepseek-reasoner",
		},
		APIKeyURL: "https://platform.deepseek.com/api_keys",
	},
	{
		ID:       "glm",
		Name:     "GLM (智谱)",
		Protocol: ProtocolOpenAI,
		BaseURL:  "https://open.bigmodel.cn/api/paas/v4",
		DefaultModels: []string{
			"glm-4-plus",
			"glm-4-air",
			"glm-4-airx",
			"glm-4-long",
		},
		APIKeyURL: "https://open.bigmodel.cn/usercenter/apikeys",
	},
	{
		ID:       "kimi",
		Name:     "Kimi (Moonshot)",
		Protocol: ProtocolOpenAI,
		BaseURL:  "https://api.moonshot.cn/v1",
		DefaultModels: []string{
			"moonshot-v1-8k",
			"moonshot-v1-32k",
			"moonshot-v1-128k",
		},
		APIKeyURL: "https://platform.moonshot.cn/console/api-keys",
	},
	{
		ID:       "qwen",
		Name:     "通义千问 (阿里)",
		Protocol: ProtocolOpenAI,
		BaseURL:  "https://dashscope.aliyuncs.com/compatible-mode/v1",
		DefaultModels: []string{
			"qwen-max",
			"qwen-plus",
			"qwen-turbo",
		},
		APIKeyURL: "https://dashscope.console.aliyun.com/apiKey",
	},
	{
		ID:       "doubao",
		Name:     "豆包 (字节)",
		Protocol: ProtocolOpenAI,
		BaseURL:  "https://ark.cn-beijing.volces.com/api/v3",
		DefaultModels: []string{
			"doubao-pro-32k",
			"doubao-lite-32k",
		},
		APIKeyURL: "https://console.volcengine.com/ark",
	},
	{
		ID:       "xiaomi",
		Name:     "小米 MiMo",
		Protocol: ProtocolOpenAI,
		BaseURL:  "https://api.xiaomimimo.com/v1",
		DefaultModels: []string{
			"mimo-v2-flash",
			"mimo-v2-pro",
		},
		APIKeyURL: "https://api.xiaomimimo.com",
	},
	// ─── 特殊用途 ──────────────────────────────────────
	{
		ID:       "openrouter",
		Name:     "OpenRouter",
		Protocol: ProtocolOpenAI,
		BaseURL:  "https://openrouter.ai/api/v1",
		DefaultModels: []string{
			"anthropic/claude-3.5-sonnet",
			"openai/gpt-4o",
			"google/gemini-2.0-flash-exp",
		},
		APIKeyURL: "https://openrouter.ai/keys",
	},
	{
		ID:       "mistral",
		Name:     "Mistral AI",
		Protocol: ProtocolOpenAI,
		BaseURL:  "https://api.mistral.ai/v1",
		DefaultModels: []string{
			"mistral-large-latest",
			"mistral-medium-latest",
			"mistral-small-latest",
		},
		APIKeyURL: "https://console.mistral.ai/api-keys",
	},
	{
		ID:       "groq",
		Name:     "Groq",
		Protocol: ProtocolOpenAI,
		BaseURL:  "https://api.groq.com/openai/v1",
		DefaultModels: []string{
			"llama-3.3-70b-versatile",
			"llama-3.1-8b-instant",
			"mixtral-8x7b-32768",
		},
		APIKeyURL: "https://console.groq.com/keys",
	},
}

// AllProviders 返回所有自定义厂商配置
func AllProviders() []ProviderInfo {
	return providerRegistry
}

// FindProvider 按 ID 查找厂商配置
func FindProvider(id string) (ProviderInfo, bool) {
	for _, p := range providerRegistry {
		if p.ID == id {
			return p, true
		}
	}
	return ProviderInfo{}, false
}
