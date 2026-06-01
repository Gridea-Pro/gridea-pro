package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// UserInfo OAuth 授权后获取的用户信息
type UserInfo struct {
	Username  string
	AvatarURL string
	Email     string
}

// TokenResponse Token 交换响应
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// Provider OAuth 提供商配置
type Provider struct {
	ID           string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	ClientID     string
	ClientSecret string
	Scopes       []string
	// FixedPort 某些平台要求回调地址与注册时完全匹配，
	// 不允许随机端口，此时使用固定端口
	FixedPort int
	// EmailURL 某些平台需要单独调用接口获取邮箱
	EmailURL       string
	EmailParser    func(body []byte) string
	UserInfoParser func(body []byte) UserInfo
	// Bootstrap 授权成功后的初始化钩子
	// 可用于自动创建默认部署目标（仓库、站点、项目等）
	// 返回平台特定的配置字段 map，将通过 oauth:success 事件传递给前端自动填充
	// 失败不阻断授权流程，仅记录日志
	Bootstrap func(client *http.Client, token, username string) (map[string]string, error)
	// CustomBuildAuthURL 某些平台（如 Vercel Integration）的授权 URL 格式与标准 OAuth 不同，
	// 如设置则使用此函数构建授权 URL
	CustomBuildAuthURL func(p *Provider, redirectURI, state string) string
}

// BuildAuthURL 构建授权 URL
func (p *Provider) BuildAuthURL(redirectURI, state string) string {
	if p.CustomBuildAuthURL != nil {
		return p.CustomBuildAuthURL(p, redirectURI, state)
	}
	params := url.Values{
		"client_id":     {p.ClientID},
		"redirect_uri":  {redirectURI},
		"state":         {state},
		"response_type": {"code"},
	}
	if len(p.Scopes) > 0 {
		params.Set("scope", strings.Join(p.Scopes, " "))
	}
	return p.AuthURL + "?" + params.Encode()
}

// ExchangeCode 用 code 换取 access_token
func (p *Provider) ExchangeCode(client *http.Client, code, redirectURI string) (TokenResponse, error) {
	data := url.Values{
		"client_id":     {p.ClientID},
		"client_secret": {p.ClientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	}
	req, err := http.NewRequest(http.MethodPost, p.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return TokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("token 交换失败: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var tr TokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		// 部分提供商返回 form-encoded 格式（如 GitHub）
		vals, _ := url.ParseQuery(string(body))
		tr.AccessToken = vals.Get("access_token")
	}
	if tr.AccessToken == "" {
		return TokenResponse{}, fmt.Errorf("响应中无 access_token: %s", string(body))
	}
	return tr, nil
}

// GetUserInfo 获取授权用户的基本信息
func (p *Provider) GetUserInfo(client *http.Client, token string) UserInfo {
	if p.UserInfoURL == "" || p.UserInfoParser == nil {
		return UserInfo{}
	}
	req, err := http.NewRequest(http.MethodGet, p.UserInfoURL, nil)
	if err != nil {
		return UserInfo{}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	// GitHub 要求 User-Agent
	req.Header.Set("User-Agent", "Gridea-Pro")

	resp, err := client.Do(req)
	if err != nil {
		return UserInfo{}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	info := p.UserInfoParser(body)

	// 如果用户信息没有 email 且提供商有单独的 email 接口，再调一次
	if info.Email == "" && p.EmailURL != "" && p.EmailParser != nil {
		emailReq, err := http.NewRequest(http.MethodGet, p.EmailURL, nil)
		if err == nil {
			emailReq.Header.Set("Authorization", "Bearer "+token)
			emailReq.Header.Set("Accept", "application/json")
			emailReq.Header.Set("User-Agent", "Gridea-Pro")
			emailResp, err := client.Do(emailReq)
			if err == nil {
				defer emailResp.Body.Close()
				emailBody, _ := io.ReadAll(emailResp.Body)
				info.Email = p.EmailParser(emailBody)
			}
		}
	}
	return info
}

// ─── Provider Registry ─────────────────────────────────────────────────────
//
// 注册 OAuth App 地址：
//   GitHub:  https://github.com/settings/applications/new
//            Redirect URI: http://127.0.0.1/oauth/callback（Wails 使用 localhost 随机端口，无需固定端口）
//
// 凭证优先级：ldflags 编译时注入（CI/Release）> 环境变量（本地开发）> 空
//
// 编译时注入（CI/CD 使用）：
//   wails build -ldflags "-X 'gridea-pro/backend/internal/service/oauth.githubClientID=xxx' ..."
//
// 环境变量配置（推荐本地开发使用）：
//   cp .env.example .env && source .env
//
// ⚠ 下面的变量必须以字符串字面量形式初始化。Go 链接器的 -X 标志只对
//   "uninitialized or initialized to a constant string expression" 生效，
//   函数调用初始化（如 getEnvOrDefault(...)）会在 init 阶段覆盖注入值，
//   导致 CI 构建出的 Release 里凭证全部为空。

var (
	githubClientID      = ""
	githubClientSecret  = ""
	netlifyClientID     = ""
	netlifyClientSecret = ""
	vercelClientID      = ""
	vercelClientSecret  = ""
	// vercelIntegrationSlug 是在 Vercel Integration Console 创建时指定的 slug
	vercelIntegrationSlug = "gridea-pro"
)

// envFallback 未被 ldflags 注入的变量，尝试从环境变量读取（本地开发场景）
func envFallback(dst *string, key string) {
	if *dst == "" {
		if v := os.Getenv(key); v != "" {
			*dst = v
		}
	}
}

// Providers 所有支持 OAuth 的平台。在 init() 中构建，确保 envFallback 先于
// ClientID / ClientSecret 捕获执行——否则本地开发（只有 env、无 ldflags）时
// Providers 会捕获到空字符串。
var Providers map[string]*Provider

func init() {
	// 未被 ldflags 注入的变量，尝试从环境变量补齐（本地开发）
	envFallback(&githubClientID, "GH_OAUTH_CLIENT_ID")
	envFallback(&githubClientSecret, "GH_OAUTH_CLIENT_SECRET")
	envFallback(&netlifyClientID, "NETLIFY_CLIENT_ID")
	envFallback(&netlifyClientSecret, "NETLIFY_CLIENT_SECRET")
	envFallback(&vercelClientID, "VERCEL_CLIENT_ID")
	envFallback(&vercelClientSecret, "VERCEL_CLIENT_SECRET")

	Providers = map[string]*Provider{
		"github": {
		ID:           "github",
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
		ClientID:     githubClientID,
		ClientSecret: githubClientSecret,
		Scopes:       []string{"public_repo", "read:user", "user:email"},
		UserInfoParser: func(body []byte) UserInfo {
			var v struct {
				Login     string `json:"login"`
				AvatarURL string `json:"avatar_url"`
				Email     string `json:"email"`
			}
			json.Unmarshal(body, &v)
			return UserInfo{Username: v.Login, AvatarURL: v.AvatarURL, Email: v.Email}
		},
		Bootstrap: ensureGitHubRepo,
	},
	"netlify": {
		ID:           "netlify",
		AuthURL:      "https://app.netlify.com/authorize",
		TokenURL:     "https://api.netlify.com/oauth/token",
		UserInfoURL:  "https://api.netlify.com/api/v1/user",
		ClientID:     netlifyClientID,
		ClientSecret: netlifyClientSecret,
		Scopes:       []string{},
		// Netlify 要求回调地址完全匹配，使用固定端口
		FixedPort: 53684,
		Bootstrap: ensureNetlifySite,
		UserInfoParser: func(body []byte) UserInfo {
			var v struct {
				Email     string `json:"email"`
				FullName  string `json:"full_name"`
				AvatarURL string `json:"avatar_url"`
			}
			json.Unmarshal(body, &v)
			name := v.FullName
			if name == "" {
				name = v.Email
			}
			return UserInfo{Username: name, AvatarURL: v.AvatarURL, Email: v.Email}
		},
	},
	"vercel": {
		ID:           "vercel",
		AuthURL:      "https://vercel.com/integrations/" + vercelIntegrationSlug + "/new",
		TokenURL:     "https://api.vercel.com/v2/oauth/access_token",
		UserInfoURL:  "https://api.vercel.com/v2/user",
		ClientID:     vercelClientID,
		ClientSecret: vercelClientSecret,
		// Vercel Integration 的 redirect URL 必须与在 Vercel 后台注册的完全一致（含端口）
		FixedPort: 53683,
		// Vercel Integration 的授权 URL 不使用标准 OAuth 参数，仅需 state
		CustomBuildAuthURL: func(p *Provider, redirectURI, state string) string {
			return p.AuthURL + "?state=" + state
		},
		UserInfoParser: func(body []byte) UserInfo {
			var wrapper struct {
				User struct {
					Username string `json:"username"`
					Name     string `json:"name"`
					Email    string `json:"email"`
					Avatar   string `json:"avatar"`
				} `json:"user"`
			}
			json.Unmarshal(body, &wrapper)
			name := wrapper.User.Username
			if name == "" {
				name = wrapper.User.Name
			}
			// Vercel 官方头像 URL 格式：根据 username 获取
			avatarURL := ""
			if name != "" {
				avatarURL = "https://vercel.com/api/www/avatar/" + name + "?s=160"
			}
			return UserInfo{Username: name, AvatarURL: avatarURL, Email: wrapper.User.Email}
		},
	},
	}
}

// IsOAuthSupported 平台是否支持 OAuth
func IsOAuthSupported(providerID string) bool {
	_, ok := Providers[providerID]
	return ok
}

// IsOAuthConfigured 平台 OAuth 凭证是否已配置
func IsOAuthConfigured(providerID string) bool {
	p, ok := Providers[providerID]
	if !ok {
		return false
	}
	return p.ClientID != "" && p.ClientSecret != ""
}

// SupportedProviders 返回所有支持 OAuth 的平台 ID 列表
func SupportedProviders() []string {
	return []string{"github", "netlify", "vercel"}
}
