package facade

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
	"gridea-pro/backend/internal/service/ai"
)

// AIFacade 暴露给前端的 AI 功能接口
type AIFacade struct {
	repo    domain.AISettingRepository
	service *service.AIService
}

func NewAIFacade(repo domain.AISettingRepository, svc *service.AIService) *AIFacade {
	return &AIFacade{repo: repo, service: svc}
}

func (f *AIFacade) ctx() context.Context {
	if WailsContext == nil {
		return context.TODO()
	}
	return WailsContext
}

// GetAISetting 获取 AI 配置
func (f *AIFacade) GetAISetting() (domain.AISetting, error) {
	return f.repo.GetAISetting(f.ctx())
}

// SaveAISettingFromFrontend 保存 AI 配置
func (f *AIFacade) SaveAISettingFromFrontend(setting domain.AISetting) error {
	return f.repo.SaveAISetting(f.ctx(), setting)
}

// GenerateSlug 根据文章标题 AI 生成 SEO 友好的英文 Slug
func (f *AIFacade) GenerateSlug(title string) (string, error) {
	return f.service.GenerateSlug(f.ctx(), title)
}

// TestConnection 测试自定义厂商连接
func (f *AIFacade) TestConnection(provider, model, apiKey string) error {
	return f.service.TestConnection(f.ctx(), provider, model, apiKey)
}

// TestConnectionWithBaseURL 测试自定义 OpenAI 兼容厂商连接
func (f *AIFacade) TestConnectionWithBaseURL(provider, model, apiKey, baseURL string) error {
	return f.service.TestConnectionWithBaseURL(f.ctx(), provider, model, apiKey, baseURL)
}

// ListProviderModels 拉取指定厂商的真实模型列表
func (f *AIFacade) ListProviderModels(provider, apiKey string) ([]string, error) {
	return f.service.ListProviderModels(f.ctx(), provider, apiKey)
}

// ListProviderModelsWithBaseURL 拉取自定义 OpenAI 兼容厂商的真实模型列表
func (f *AIFacade) ListProviderModelsWithBaseURL(provider, apiKey, baseURL string) ([]string, error) {
	return f.service.ListProviderModelsWithBaseURL(f.ctx(), provider, apiKey, baseURL)
}

// GetProviderRegistry 返回所有自定义厂商配置（供前端展示）
func (f *AIFacade) GetProviderRegistry() []ai.ProviderInfo {
	return f.service.GetProviderRegistry()
}

// GetBuiltInModels 返回内置免费模型清单
func (f *AIFacade) GetBuiltInModels() []string {
	return f.service.GetBuiltInModels()
}
