package domain

import "context"

// Setting 系统设置
type Setting struct {
	Platform           string `json:"platform"`
	Domain             string `json:"domain"`
	Repository         string `json:"repository"`
	Branch             string `json:"branch"`
	Username           string `json:"username"`
	Email              string `json:"email"`
	TokenUsername      string `json:"tokenUsername"`
	Token              string `json:"token"`
	Cname              string `json:"cname"`
	Port               string `json:"port"`
	Server             string `json:"server"`
	Password           string `json:"password"`
	PrivateKey         string `json:"privateKey"`
	RemotePath         string `json:"remotePath"`
	ProxyPath          string `json:"proxyPath"`
	ProxyPort          string `json:"proxyPort"`
	EnabledProxy       string `json:"enabledProxy"`
	NetlifySiteId      string `json:"netlifySiteId"`
	NetlifyAccessToken string `json:"netlifyAccessToken"`
}

// SettingRepository 定义配置存储接口
type SettingRepository interface {
	GetSetting(ctx context.Context) (Setting, error)
	SaveSetting(ctx context.Context, setting Setting) error
}

type DeployResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
