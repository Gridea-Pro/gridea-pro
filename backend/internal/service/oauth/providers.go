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
	UserInfoParser func(body []byte) UserInfo
}

// BuildAuthURL 构建授权 URL
func (p *Provider) BuildAuthURL(redirectURI, state string) string {
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
	return p.UserInfoParser(body)
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
	githubClientID      = "YOUR_GITHUB_CLIENT_ID"
	githubClientSecret  = "YOUR_GITHUB_CLIENT_SECRET"
	giteeClientID       = "YOUR_GITEE_CLIENT_ID"
	giteeClientSecret   = "YOUR_GITEE_CLIENT_SECRET"
	netlifyClientID     = "YOUR_NETLIFY_CLIENT_ID"
	netlifyClientSecret = "YOUR_NETLIFY_CLIENT_SECRET"
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
		Scopes:       []string{"repo", "user:email"},
		UserInfoParser: func(body []byte) UserInfo {
			var v struct {
				Login     string `json:"login"`
				AvatarURL string `json:"avatar_url"`
				Email     string `json:"email"`
			}
			json.Unmarshal(body, &v)
			return UserInfo{Username: v.Login, AvatarURL: v.AvatarURL, Email: v.Email}
		},
	},
	"gitee": {
		ID:           "gitee",
		AuthURL:      "https://gitee.com/oauth/authorize",
		TokenURL:     "https://gitee.com/oauth/token",
		UserInfoURL:  "https://gitee.com/api/v5/user",
		ClientID:     giteeClientID,
		ClientSecret: giteeClientSecret,
		Scopes:       []string{"projects", "user_info", "emails"},
		UserInfoParser: func(body []byte) UserInfo {
			var v struct {
				Login     string `json:"login"`
				AvatarURL string `json:"avatar_url"`
				Email     string `json:"email"`
			}
			json.Unmarshal(body, &v)
			return UserInfo{Username: v.Login, AvatarURL: v.AvatarURL, Email: v.Email}
		},
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
	return []string{"github", "gitee", "netlify"}
}
