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
func (s *OAuthService) StartOAuthFlow(ctx context.Context, providerID string) error {
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

	go s.runCallbackServer(ctx, listener, providerID, p, redirectURI, state)

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

func (s *OAuthService) runCallbackServer(ctx context.Context, listener net.Listener, providerID string, p *oauth.Provider, redirectURI, expectedState string) {
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
			fmt.Fprint(w, oauthResultHTML(false, providerID, errParam, "", ""))
			runtime.EventsEmit(ctx, "oauth:error", map[string]string{"provider": providerID, "error": errParam})
			return
		}

		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if state != expectedState {
			fmt.Fprint(w, oauthResultHTML(false, providerID, "state 验证失败，请重新授权", "", ""))
			return
		}

		client := &http.Client{Timeout: 15 * time.Second}
		tokenResp, err := p.ExchangeCode(client, code, redirectURI)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, oauthResultHTML(false, providerID, err.Error(), "", ""))
			runtime.EventsEmit(ctx, "oauth:error", map[string]string{"provider": providerID, "error": err.Error()})
			return
		}

		// 存入 Keychain
		credKey := providerID + ":" + primaryCredField(providerID)
		if err := s.credService.Set(credKey, tokenResp.AccessToken); err != nil {
			fmt.Fprint(w, oauthResultHTML(false, providerID, "存储凭证失败: "+err.Error(), "", ""))
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
		fmt.Fprint(w, oauthResultHTML(true, providerID, "", userInfo.Username, userInfo.AvatarURL))
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

var platformDisplayNames = map[string]string{
	"github":  "GitHub",
	"gitee":   "Gitee",
	"netlify": "Netlify",
	"vercel":  "Vercel",
	"coding":  "Coding",
	"sftp":    "SFTP",
}

func oauthResultHTML(success bool, providerID, errMsg, username, avatarURL string) string {
	platformName := platformDisplayNames[providerID]
	if platformName == "" {
		platformName = providerID
	}
	const pageStyle = `*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI','Helvetica Neue',sans-serif;display:flex;align-items:center;justify-content:center;min-height:100vh;background:#000}
.card{text-align:center;padding:48px 56px;background:rgba(255,255,255,.06);border:1px solid rgba(255,255,255,.08);border-radius:24px;backdrop-filter:blur(20px);max-width:420px;width:100%}
.brand{margin-bottom:32px}
.brand img{width:56px;height:56px;border-radius:14px;margin-bottom:10px}
.brand-name{font-size:20px;font-weight:600;color:rgba(255,255,255,.9);letter-spacing:-.2px}
.status-icon{width:52px;height:52px;border-radius:50%;display:flex;align-items:center;justify-content:center;margin:0 auto 12px}
.status-icon.ok{background:linear-gradient(135deg,#34c759,#30d158)}
.status-icon.fail{background:linear-gradient(135deg,#ff3b30,#ff453a)}
.status-icon svg{width:26px;height:26px;color:#fff}
.title{font-size:16px;font-weight:700;color:#fff;margin-bottom:16px}
.user-row{display:flex;align-items:center;justify-content:center;gap:10px;margin-bottom:8px}
.user-row img{width:24px;height:24px;border-radius:50%;border:1px solid rgba(255, 255, 255)}
.user-row span{font-size:14px;font-weight:600;color:rgba(255,255,255,.85)}
.hint{font-size:13px;color:rgba(255,255,255,.4);line-height:1.6}
.err{font-size:13px;color:rgba(255,255,255,.5);line-height:1.6;word-break:break-all}
.divider{height:1px;background:rgba(255,255,255,.06);margin:28px 0 20px}
.footer{font-size:11px;color:rgba(255,255,255,.2)}`

	// Gridea Pro logo (base64 encoded small PNG placeholder — replace with actual logo URL in production)
	const logoURL = "https://www.gridea.pro/appicon.png"

	if success {
		name := username
		if name == "" {
			name = "账号"
		}
		userHTML := ""
		if avatarURL != "" {
			userHTML = `<div class="user-row"><img src="` + avatarURL + `" alt="" /><span>` + name + ` 已连接</span></div>`
		} else {
			userHTML = `<div class="user-row"><span>` + name + ` 已连接</span></div>`
		}
		return `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>Gridea Pro - 授权成功</title><link rel="icon" href="https://www.gridea.pro/favicon.ico">
<style>` + pageStyle + `</style></head>
<body><div class="card">
<div class="brand"><img src="` + logoURL + `" alt="Gridea Pro" onerror="this.style.display='none'" /><div class="brand-name">Gridea Pro</div></div>
<div class="status-icon ok"><svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"><path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"/></svg></div>
<div class="title">` + platformName + ` 授权成功</div>
` + userHTML + `
<div class="hint">请返回 Gridea Pro 查看，可以关闭此标签页</div>
<div class="divider"></div>
<div class="footer">Gridea Pro · 下一代桌面静态博客写作客户端</div>
</div></body></html>`
	}
	return `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>Gridea Pro - 授权失败</title>
<style>` + pageStyle + `</style></head>
<body><div class="card">
<div class="brand"><img src="` + logoURL + `" alt="Gridea Pro" onerror="this.style.display='none'" /><div class="brand-name">Gridea Pro</div></div>
<div class="status-icon fail"><svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/></svg></div>
<div class="title">` + platformName + ` 授权失败</div>
<div class="err">` + errMsg + `</div>
<div class="divider"></div>
<div class="footer">Gridea Pro · 下一代桌面静态博客写作客户端</div>
</div></body></html>`
}
