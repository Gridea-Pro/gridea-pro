package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"gridea-pro/backend/internal/config"
	"gridea-pro/backend/internal/service/credential"
	"gridea-pro/backend/internal/service/oauth"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// PlatformStatus 平台连接状态（返回前端）
type PlatformStatus struct {
	Connected    bool   `json:"connected"`
	ConnectedVia string `json:"connectedVia"` // "oauth" | "manual" | ""
	Username     string `json:"username,omitempty"`
	AvatarURL    string `json:"avatarUrl,omitempty"`
	Email        string `json:"email,omitempty"`
}

// OAuthService 处理平台授权与凭证管理
type OAuthService struct {
	credService  *credential.Service
	configMgr    *config.ConfigManager
	mu           sync.Mutex
	activeServer *http.Server
}

func NewOAuthService(credService *credential.Service, configMgr *config.ConfigManager) *OAuthService {
	return &OAuthService{
		credService: credService,
		configMgr:   configMgr,
	}
}

// StartOAuthFlow 启动 OAuth 流程：打开本地回调服务器 + 唤起浏览器
func (s *OAuthService) StartOAuthFlow(ctx context.Context, providerID, lang string) error {
	if !oauth.IsOAuthConfigured(providerID) {
		return fmt.Errorf("平台 %s 的 OAuth 应用尚未配置，请使用「手动配置」填写 Token", providerID)
	}
	p := oauth.Providers[providerID]

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("无法启动本地回调服务器: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/oauth/callback", port)

	state, err := generateOAuthState()
	if err != nil {
		listener.Close()
		return err
	}

	go s.runCallbackServer(ctx, listener, providerID, lang, p, redirectURI, state)

	authURL := p.BuildAuthURL(redirectURI, state)
	runtime.BrowserOpenURL(ctx, authURL)
	return nil
}

// RevokeToken 撤销平台授权（删除 Keychain 中的凭证 + 清除 meta 信息）
func (s *OAuthService) RevokeToken(ctx context.Context, providerID string) error {
	fields := getAllCredentialFields(providerID)
	for _, field := range fields {
		key := providerID + ":" + field
		_ = s.credService.Delete(key)
	}
	return s.configMgr.SavePlatformMeta(providerID, config.PlatformMeta{})
}

// GetAllStatuses 获取所有平台的连接状态
func (s *OAuthService) GetAllStatuses(ctx context.Context) map[string]PlatformStatus {
	platforms := []string{"github", "netlify", "vercel", "gitee", "coding", "sftp"}
	result := make(map[string]PlatformStatus, len(platforms))
	for _, p := range platforms {
		result[p] = s.getStatus(p)
	}
	return result
}

// HasCredential 检查指定平台的某个凭证字段是否已存储
func (s *OAuthService) HasCredential(providerID, field string) bool {
	return s.credService.Has(providerID + ":" + field)
}

// GetCredential 读取凭证（部署 / 测试时内部使用）
func (s *OAuthService) GetCredential(providerID, field string) string {
	return s.credService.Get(providerID + ":" + field)
}

// SaveManualCredentials 手动保存多个凭证字段（从 SaveSettingFromFrontend 路由过来）
func (s *OAuthService) SaveManualCredentials(ctx context.Context, providerID string, credentials map[string]string) error {
	for field, value := range credentials {
		key := providerID + ":" + field
		if value == "" {
			_ = s.credService.Delete(key)
			continue
		}
		if err := s.credService.Set(key, value); err != nil {
			return fmt.Errorf("保存 %s 凭证失败: %w", field, err)
		}
	}
	// 更新 meta：标记为手动配置
	if len(credentials) > 0 {
		meta := s.configMgr.GetPlatformMeta(providerID)
		meta.ConnectedVia = "manual"
		_ = s.configMgr.SavePlatformMeta(providerID, meta)
	}
	return nil
}

// GetAllCredentials 批量读取所有平台所有凭证（给部署服务使用）
func (s *OAuthService) GetAllCredentials() map[string]string {
	result := make(map[string]string)
	for platform, fields := range sensitiveFieldsByPlatform {
		for _, field := range fields {
			key := platform + ":" + field
			val := s.credService.Get(key)
			if val != "" {
				result[key] = val
			}
		}
	}
	return result
}

// OAuthSupportedProviders 返回所有支持 OAuth 的平台（包括是否已配置 client_id）
func (s *OAuthService) OAuthSupportedProviders() []string {
	return oauth.SupportedProviders()
}

// IsOAuthAvailable 检查平台是否支持 OAuth 且已配置凭证
func (s *OAuthService) IsOAuthAvailable(providerID string) bool {
	return oauth.IsOAuthConfigured(providerID)
}

// ─── Private ──────────────────────────────────────────────────────────────

func (s *OAuthService) getStatus(providerID string) PlatformStatus {
	fields := getAllCredentialFields(providerID)
	hasAny := false
	for _, f := range fields {
		if s.credService.Has(providerID + ":" + f) {
			hasAny = true
			break
		}
	}
	if !hasAny {
		return PlatformStatus{Connected: false}
	}
	meta := s.configMgr.GetPlatformMeta(providerID)
	return PlatformStatus{
		Connected:    true,
		ConnectedVia: meta.ConnectedVia,
		Username:     meta.Username,
		AvatarURL:    meta.AvatarURL,
		Email:        meta.Email,
	}
}

func (s *OAuthService) runCallbackServer(ctx context.Context, listener net.Listener, providerID, lang string, p *oauth.Provider, redirectURI, expectedState string) {
	mux := http.NewServeMux()
	server := &http.Server{Handler: mux}

	s.mu.Lock()
	if s.activeServer != nil {
		s.activeServer.Close()
	}
	s.activeServer = server
	s.mu.Unlock()

	// 5 分钟内无操作自动关闭
	go func() {
		select {
		case <-time.After(5 * time.Minute):
		case <-ctx.Done():
		}
		server.Close()
		listener.Close()
	}()

	mux.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		// 延迟关闭服务器，确保 HTML 响应完整发送到浏览器
		defer func() {
			go func() {
				time.Sleep(3 * time.Second)
				server.Close()
			}()
		}()

		errParam := r.URL.Query().Get("error")
		if errParam != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, oauthResultHTML(false, providerID, lang, errParam, "", ""))
			runtime.EventsEmit(ctx, "oauth:error", map[string]string{"provider": providerID, "error": errParam})
			return
		}

		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if state != expectedState {
			fmt.Fprint(w, oauthResultHTML(false, providerID, lang, "state 验证失败，请重新授权", "", ""))
			return
		}

		client := &http.Client{Timeout: 15 * time.Second}
		tokenResp, err := p.ExchangeCode(client, code, redirectURI)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, oauthResultHTML(false, providerID, lang, err.Error(), "", ""))
			runtime.EventsEmit(ctx, "oauth:error", map[string]string{"provider": providerID, "error": err.Error()})
			return
		}

		// 存入 Keychain
		credKey := providerID + ":" + primaryCredField(providerID)
		if err := s.credService.Set(credKey, tokenResp.AccessToken); err != nil {
			fmt.Fprint(w, oauthResultHTML(false, providerID, lang, "存储凭证失败: "+err.Error(), "", ""))
			return
		}

		// 获取用户信息
		userInfo := p.GetUserInfo(client, tokenResp.AccessToken)

		// 保存 meta
		_ = s.configMgr.SavePlatformMeta(providerID, config.PlatformMeta{
			ConnectedVia: "oauth",
			Username:     userInfo.Username,
			AvatarURL:    userInfo.AvatarURL,
			Email:        userInfo.Email,
		})

		// 通知前端
		runtime.EventsEmit(ctx, "oauth:success", map[string]interface{}{
			"provider":  providerID,
			"username":  userInfo.Username,
			"avatarUrl": userInfo.AvatarURL,
			"email":     userInfo.Email,
		})

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, oauthResultHTML(true, providerID, lang, "", userInfo.Username, userInfo.AvatarURL))
	})

	server.Serve(listener)
}

// ─── Helpers ──────────────────────────────────────────────────────────────

// sensitiveFieldsByPlatform 同 domain.SensitiveFields，此处冗余以避免包循环引用
var sensitiveFieldsByPlatform = map[string][]string{
	"github":  {"token"},
	"gitee":   {"token"},
	"coding":  {"token"},
	"netlify": {"netlifyAccessToken"},
	"vercel":  {"token"},
	"sftp":    {"password", "privateKey"},
}

func getAllCredentialFields(providerID string) []string {
	if fields, ok := sensitiveFieldsByPlatform[providerID]; ok {
		return fields
	}
	return nil
}

// primaryCredField 返回该平台的主 Token 字段名（OAuth 授权后存储的字段）
func primaryCredField(providerID string) string {
	switch providerID {
	case "netlify":
		return "netlifyAccessToken"
	default:
		return "token"
	}
}

func generateOAuthState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

