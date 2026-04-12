package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	ID             string
	AuthURL        string
	TokenURL       string
	UserInfoURL    string
	ClientID       string
	ClientSecret   string
	Scopes         []string
	// FixedPort 某些平台（如 Gitee）要求回调地址与注册时完全匹配，
	// 不允许随机端口，此时使用固定端口
	FixedPort int
	// EmailURL 某些平台（如 Gitee）需要单独调用接口获取邮箱
	EmailURL       string
	EmailParser    func(body []byte) string
	UserInfoParser func(body []byte) UserInfo
	// EnsureDefaultRepo 确保用户的默认部署仓库存在（不存在则创建，已存在则跳过）
	// 返回默认仓库名，或错误（不影响授权流程）
	EnsureDefaultRepo func(client *http.Client, token, username string) (repoName string, err error)
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
//   Gitee:   https://gitee.com/oauth/applications
//
// 在编译时通过 ldflags 注入凭证：
//   wails build -ldflags "-X 'gridea-pro/backend/internal/service/oauth.githubClientID=xxx' -X 'gridea-pro/backend/internal/service/oauth.githubClientSecret=xxx'"

var (
	githubClientID      = "Ov23li2hRBoUIkY83knT"
	githubClientSecret  = "43274fee8a5a8c922719fa1bd7b911a2e0022115"
	giteeClientID       = "d10d5bebeb569e48ab8b128e5151c8f67a24fb110498898ccb58eb34a6995d56"
	giteeClientSecret   = "dde3869c9c89edabb7b9a20b61044038c31fadeeacc2ca74565a218ba19209c8"
	netlifyClientID     = "YOUR_NETLIFY_CLIENT_ID"
	netlifyClientSecret = "YOUR_NETLIFY_CLIENT_SECRET"
	vercelClientID      = "oac_Dveh2ueHQLSym2UcJB1vRzXD"
	vercelClientSecret  = "gECj7VpFOS7fmH7cHGoxC4L1"
	// vercelIntegrationSlug 是在 Vercel Integration Console 创建时指定的 slug
	vercelIntegrationSlug = "gridea-pro"
)

// Providers 所有支持 OAuth 的平台
var Providers = map[string]*Provider{
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
		EnsureDefaultRepo: ensureGitHubRepo,
	},
	"gitee": {
		ID:           "gitee",
		AuthURL:      "https://gitee.com/oauth/authorize",
		TokenURL:     "https://gitee.com/oauth/token",
		UserInfoURL:  "https://gitee.com/api/v5/user",
		EmailURL:     "https://gitee.com/api/v5/emails",
		ClientID:     giteeClientID,
		ClientSecret: giteeClientSecret,
		Scopes:       []string{"projects", "user_info", "emails"},
		FixedPort:    53682, // Gitee 要求回调地址完全匹配，使用固定端口
		UserInfoParser: func(body []byte) UserInfo {
			var v struct {
				Login     string `json:"login"`
				AvatarURL string `json:"avatar_url"`
				Email     string `json:"email"`
			}
			json.Unmarshal(body, &v)
			return UserInfo{Username: v.Login, AvatarURL: v.AvatarURL, Email: v.Email}
		},
		EmailParser: func(body []byte) string {
			// Gitee /api/v5/emails 返回数组，优先返回 primary email
			var emails []struct {
				Email string   `json:"email"`
				State string   `json:"state"`
				Scope []string `json:"scope"`
			}
			if err := json.Unmarshal(body, &emails); err != nil {
				return ""
			}
			// 优先 primary
			for _, e := range emails {
				for _, s := range e.Scope {
					if s == "primary" {
						return e.Email
					}
				}
			}
			// 退而求其次返回第一个
			if len(emails) > 0 {
				return emails[0].Email
			}
			return ""
		},
		EnsureDefaultRepo: ensureGiteeRepo,
	},
	"netlify": {
		ID:           "netlify",
		AuthURL:      "https://app.netlify.com/authorize",
		TokenURL:     "https://api.netlify.com/oauth/token",
		UserInfoURL:  "https://api.netlify.com/api/v1/user",
		ClientID:     netlifyClientID,
		ClientSecret: netlifyClientSecret,
		Scopes:       []string{},
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
			avatarURL := ""
			if wrapper.User.Avatar != "" {
				avatarURL = "https://api.vercel.com/www/avatar/" + wrapper.User.Avatar + "?s=160"
			}
			return UserInfo{Username: name, AvatarURL: avatarURL, Email: wrapper.User.Email}
		},
	},
}

// IsOAuthSupported 平台是否支持 OAuth
func IsOAuthSupported(providerID string) bool {
	_, ok := Providers[providerID]
	return ok
}

// IsOAuthConfigured 平台 OAuth 凭证是否已配置（非占位符）
func IsOAuthConfigured(providerID string) bool {
	p, ok := Providers[providerID]
	if !ok {
		return false
	}
	return p.ClientID != "" && !strings.HasPrefix(p.ClientID, "YOUR_")
}

// SupportedProviders 返回所有支持 OAuth 的平台 ID 列表
func SupportedProviders() []string {
	return []string{"github", "gitee", "netlify", "vercel"}
}
