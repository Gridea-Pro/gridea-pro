package facade

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
	"gridea-pro/backend/internal/service/ai"
	"sync/atomic"
)

// AIFacade 暴露给前端的 AI 功能接口
type AIFacade struct {
	// repo / service 用原子读写保存：切站（UpdateAppDir）会热替换它们，
	// 而 Wails 可能并发调用本 facade 的方法，原子读写避免 data race。
	repo    atomic.Value // 存 domain.AISettingRepository
	service atomic.Pointer[service.AIService]
}

func NewAIFacade(repo domain.AISettingRepository, svc *service.AIService) *AIFacade {
	f := &AIFacade{}
	f.repo.Store(repo)
	f.service.Store(svc)
	return f
}

func (f *AIFacade) repository() domain.AISettingRepository {
	return f.repo.Load().(domain.AISettingRepository)
}

func (f *AIFacade) svc() *service.AIService { return f.service.Load() }

func (f *AIFacade) ctx() context.Context {
	if WailsContext == nil {
		return context.TODO()
	}
	return WailsContext
}

// GetAISetting 获取 AI 配置
func (f *AIFacade) GetAISetting() (domain.AISetting, error) {
	return f.repository().GetAISetting(f.ctx())
}

// SaveAISettingFromFrontend 保存 AI 配置
func (f *AIFacade) SaveAISettingFromFrontend(setting domain.AISetting) error {
	return f.repository().SaveAISetting(f.ctx(), setting)
}

// GenerateSlug 根据文章标题 AI 生成 SEO 友好的英文 Slug
func (f *AIFacade) GenerateSlug(title string) (string, error) {
	return f.svc().GenerateSlug(f.ctx(), title)
}

// Complete 行内 AI 续写：给定光标前后文，返回应插入的补全文本
func (f *AIFacade) Complete(prefix, suffix string) (string, error) {
	return f.svc().Complete(f.ctx(), prefix, suffix)
}

// Polish 文本润色：返回润色后的文本
func (f *AIFacade) Polish(text string) (string, error) {
	return f.svc().Polish(f.ctx(), text)
}

// Summary 生成文章摘要
func (f *AIFacade) Summary(content string) (string, error) {
	return f.svc().Summary(f.ctx(), content)
}

// TestConnection 测试自定义厂商连接
func (f *AIFacade) TestConnection(provider, model, apiKey string) error {
	return f.svc().TestConnection(f.ctx(), provider, model, apiKey)
}

// TestConnectionWithBaseURL 测试自定义兼容厂商连接
func (f *AIFacade) TestConnectionWithBaseURL(provider, model, apiKey, baseURL string) error {
	return f.svc().TestConnectionWithBaseURL(f.ctx(), provider, model, apiKey, baseURL)
}

// ListProviderModels 拉取指定厂商的真实模型列表
func (f *AIFacade) ListProviderModels(provider, apiKey string) ([]string, error) {
	return f.svc().ListProviderModels(f.ctx(), provider, apiKey)
}

// ListProviderModelsWithBaseURL 拉取自定义兼容厂商的真实模型列表
func (f *AIFacade) ListProviderModelsWithBaseURL(provider, apiKey, baseURL string) ([]string, error) {
	return f.svc().ListProviderModelsWithBaseURL(f.ctx(), provider, apiKey, baseURL)
}

// GetProviderRegistry 返回所有自定义厂商配置（供前端展示）
func (f *AIFacade) GetProviderRegistry() []ai.ProviderInfo {
	return f.svc().GetProviderRegistry()
}

// GetBuiltInModels 返回内置免费模型清单
func (f *AIFacade) GetBuiltInModels() []string {
	return f.svc().GetBuiltInModels()
}
