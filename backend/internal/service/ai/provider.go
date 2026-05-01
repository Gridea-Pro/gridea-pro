package ai

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
)

// ChatRequest 通用 chat 请求
type ChatRequest struct {
	Model       string
	Prompt      string
	Temperature float64
	MaxTokens   int
}

// Provider 模型厂商适配器接口
type Provider interface {
	// Chat 发起一次 chat 请求并返回模型输出文本
	Chat(ctx context.Context, req ChatRequest, apiKey string, httpClient *http.Client) (string, error)
	// ListModels 拉取该厂商可用模型列表（用户填了 Key 后才能调用）
	ListModels(ctx context.Context, apiKey string, httpClient *http.Client) ([]string, error)
}

// NewProvider 根据厂商 ID 创建对应的 Provider 实例
func NewProvider(providerID string) (Provider, ProviderInfo, error) {
	info, ok := FindProvider(providerID)
	if !ok {
		return nil, ProviderInfo{}, errors.New("未知的模型厂商: " + providerID)
	}
	return NewProviderFromInfo(info)
}

// NewOpenAICompatibleProvider 创建使用指定 Base URL 的 OpenAI 兼容 Provider。
// 用于支持各类中转站 / 自定义 OpenAI Compatible 接口。
func NewOpenAICompatibleProvider(baseURL string) Provider {
	return &openAICompatProvider{baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/")}
}

// NewProviderFromInfo 根据厂商元信息创建对应的 Provider 实例
func NewProviderFromInfo(info ProviderInfo) (Provider, ProviderInfo, error) {
	info.BaseURL = strings.TrimRight(strings.TrimSpace(info.BaseURL), "/")
	switch info.Protocol {
	case ProtocolOpenAI:
		return &openAICompatProvider{baseURL: info.BaseURL}, info, nil
	case ProtocolAnthropic:
		return &anthropicProvider{baseURL: info.BaseURL}, info, nil
	case ProtocolGemini:
		return &geminiProvider{baseURL: info.BaseURL}, info, nil
	}
	return nil, ProviderInfo{}, errors.New("不支持的协议: " + string(info.Protocol))
}

// 模型黑名单关键词，用于过滤 ListModels 结果中非 chat 类模型
var modelBlocklist = []string{
	"embed", "whisper", "dall-e", "tts", "audio",
	"realtime", "image", "video", "cogview", "cogvideo",
	"moderation", "search", "rerank",
}

func filterChatModels(models []string) []string {
	out := make([]string, 0, len(models))
	for _, m := range models {
		lower := strings.ToLower(m)
		blocked := false
		for _, kw := range modelBlocklist {
			if strings.Contains(lower, kw) {
				blocked = true
				break
			}
		}
		if !blocked {
			out = append(out, m)
		}
	}
	return out
}

// defaultHTTPClient 默认 HTTP 客户端（30s 超时）
func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}
