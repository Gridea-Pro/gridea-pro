package facade

import (
	"context"

	"gridea-pro/backend/internal/service"
)

// OAuthFacade 平台授权 / 凭证管理 Facade
type OAuthFacade struct {
	service *service.OAuthService
}

func NewOAuthFacade(svc *service.OAuthService) *OAuthFacade {
	return &OAuthFacade{service: svc}
}

// StartOAuthFlow 启动 OAuth 授权流程
// 会在系统浏览器打开授权页面，授权完成后通过 Wails 事件通知前端
// 事件名：oauth:success / oauth:error
func (f *OAuthFacade) StartOAuthFlow(provider, lang string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.service.StartOAuthFlow(ctx, provider, lang)
}

// RevokeToken 撤销指定平台的授权（清除 Keychain 凭证 + 连接元信息）
func (f *OAuthFacade) RevokeToken(provider string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.service.RevokeToken(ctx, provider)
}

// GetAllStatuses 获取所有平台的连接状态（前端列表页使用）
func (f *OAuthFacade) GetAllStatuses() map[string]service.PlatformStatus {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.service.GetAllStatuses(ctx)
}

// HasCredential 检查指定平台的某凭证字段是否已存储（用于 UI 显示"已配置"）
func (f *OAuthFacade) HasCredential(provider, field string) bool {
	return f.service.HasCredential(provider, field)
}

// IsOAuthAvailable 检查平台是否支持 OAuth 且 client credentials 已配置
func (f *OAuthFacade) IsOAuthAvailable(provider string) bool {
	return f.service.IsOAuthAvailable(provider)
}

// OAuthSupportedProviders 返回支持 OAuth 的平台 ID 列表
func (f *OAuthFacade) OAuthSupportedProviders() []string {
	return f.service.OAuthSupportedProviders()
}
