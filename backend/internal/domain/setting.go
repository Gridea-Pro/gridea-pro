package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
)

// Setting 系统设置
// platform 标识当前启用的平台，platformConfigs 按平台独立存储所有配置
// 注意：敏感字段（token/password/privateKey）存储于系统 Keychain，不序列化到 JSON
type Setting struct {
	Platform        string                    `json:"platform"`
	PlatformConfigs map[string]map[string]any `json:"platformConfigs,omitempty"`
	ProxyEnabled    bool                      `json:"proxyEnabled"`
	ProxyURL        string                    `json:"proxyURL"`
}

// SensitiveFields 各平台需要存入 Keychain 的敏感字段
// key: 平台 ID，value: 字段名列表
var SensitiveFields = map[string][]string{
	"github":  {"token"},
	"gitee":   {"token"},
	"coding":  {"token"},
	"netlify": {"netlifyAccessToken"},
	"vercel":  {"token"},
	"sftp":    {"password", "privateKey"},
}

// ExtractSensitiveFields 从 PlatformConfigs 中提取敏感字段
// 返回 map[credentialKey]value，并将原始配置中的这些字段清空
// credentialKey 格式为 "{platform}:{field}"，与 Keychain 的 account 字段对应
func (s *Setting) ExtractSensitiveFields() map[string]string {
	result := make(map[string]string)
	if s.PlatformConfigs == nil {
		return result
	}
	for platform, fields := range SensitiveFields {
		cfg := s.PlatformConfigs[platform]
		if cfg == nil {
			continue
		}
		for _, field := range fields {
			val, ok := cfg[field].(string)
			if !ok || val == "" {
				continue
			}
			result[platform+":"+field] = val
			delete(cfg, field)
		}
	}
	return result
}

// InjectCredentials 将凭证值注入 PlatformConfigs（用于部署/测试时补全凭证）
// 仅在对应字段为空时才注入（不覆盖前端新传入的值）
func (s *Setting) InjectCredentials(credentials map[string]string) {
	if s.PlatformConfigs == nil {
		s.PlatformConfigs = make(map[string]map[string]any)
	}
	for platform, fields := range SensitiveFields {
		for _, field := range fields {
			key := platform + ":" + field
			val, ok := credentials[key]
			if !ok || val == "" {
				continue
			}
			// 只在字段为空时注入
			cfg := s.PlatformConfigs[platform]
			if cfg == nil {
				cfg = make(map[string]any)
				s.PlatformConfigs[platform] = cfg
			}
			if existing, _ := cfg[field].(string); existing == "" {
				cfg[field] = val
			}
		}
	}
}

// platformFieldOrder 定义各平台配置项的输出顺序，与前端表单顺序一致
var platformFieldOrder = map[string][]string{
	"github": {"domain", "repository", "branch", "username", "email", "tokenUsername", "token", "cname"},
	"gitee":  {"domain", "repository", "branch", "username", "email", "tokenUsername", "token", "cname"},
	"coding": {"domain", "repository", "branch", "username", "email", "tokenUsername", "token", "cname"},
	"netlify": {"domain", "netlifySiteId", "netlifyAccessToken"},
	"vercel": {"domain", "repository", "token", "cname"},
	"sftp":   {"domain", "transferProtocol", "server", "port", "username", "password", "privateKey", "remotePath"},
}

// MarshalJSON 自定义 JSON 序列化，确保平台配置项按前端表单顺序输出
func (s Setting) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(`{"platform":`)
	p, _ := json.Marshal(s.Platform)
	buf.Write(p)

	if len(s.PlatformConfigs) > 0 {
		buf.WriteString(`,"platformConfigs":{`)

		// 平台名按字母序排列
		platforms := make([]string, 0, len(s.PlatformConfigs))
		for k := range s.PlatformConfigs {
			platforms = append(platforms, k)
		}
		sort.Strings(platforms)

		for i, platform := range platforms {
			if i > 0 {
				buf.WriteByte(',')
			}
			pk, _ := json.Marshal(platform)
			buf.Write(pk)
			buf.WriteByte(':')

			cfg := s.PlatformConfigs[platform]
			order := platformFieldOrder[platform]
			if order == nil {
				// 未知平台，使用默认序列化
				d, _ := json.Marshal(cfg)
				buf.Write(d)
			} else {
				buf.WriteByte('{')
				first := true
				// 按定义顺序输出已有字段
				for _, key := range order {
					v, ok := cfg[key]
					if !ok {
						continue
					}
					if !first {
						buf.WriteByte(',')
					}
					first = false
					kk, _ := json.Marshal(key)
					buf.Write(kk)
					buf.WriteByte(':')
					vv, _ := json.Marshal(v)
					buf.Write(vv)
				}
				// 输出不在 order ���的额外字段
				for key, v := range cfg {
					found := false
					for _, ok := range order {
						if ok == key {
							found = true
							break
						}
					}
					if !found {
						if !first {
							buf.WriteByte(',')
						}
						first = false
						kk, _ := json.Marshal(key)
						buf.Write(kk)
						buf.WriteByte(':')
						vv, _ := json.Marshal(v)
						buf.Write(vv)
					}
				}
				buf.WriteByte('}')
			}
		}
		buf.WriteByte('}')
	}

	// 序列化代理设置
	buf.WriteString(`,"proxyEnabled":`)
	buf.WriteString(strconv.FormatBool(s.ProxyEnabled))
	buf.WriteString(`,"proxyURL":`)
	proxyURL, _ := json.Marshal(s.ProxyURL)
	buf.Write(proxyURL)

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// getConfig 获取当前平台的配置 map
func (s *Setting) getConfig() map[string]any {
	if s.PlatformConfigs == nil {
		return nil
	}
	return s.PlatformConfigs[s.Platform]
}

// Get 获取当前平台的指定配置项
func (s *Setting) Get(key string) string {
	m := s.getConfig()
	if m == nil {
		return ""
	}
	v, _ := m[key].(string)
	return v
}

// GetFrom 获取指定平台的指定配置项
func (s *Setting) GetFrom(platform, key string) string {
	if s.PlatformConfigs == nil {
		return ""
	}
	m := s.PlatformConfigs[platform]
	if m == nil {
		return ""
	}
	v, _ := m[key].(string)
	return v
}

// Domain 当前平台的域名
func (s *Setting) Domain() string { return s.Get("domain") }

// Repository 当前平台的仓库名/项目名
func (s *Setting) Repository() string { return s.Get("repository") }

// Branch 当前平台的分支
func (s *Setting) Branch() string { return s.Get("branch") }

// Username 当前平台的用户名
func (s *Setting) Username() string { return s.Get("username") }

// Email 当前平台的邮箱
func (s *Setting) Email() string { return s.Get("email") }

// TokenUsername 当前平台的 Token 用户名
func (s *Setting) TokenUsername() string { return s.Get("tokenUsername") }

// Token 当前平台的 Token
func (s *Setting) Token() string { return s.Get("token") }

// CNAME 当前平台的 CNAME
func (s *Setting) CNAME() string { return s.Get("cname") }

// Password 当前平台的密码
func (s *Setting) Password() string { return s.Get("password") }

// PrivateKey 当前平台的私钥路径
func (s *Setting) PrivateKey() string { return s.Get("privateKey") }

// NetlifyAccessToken 当前平台的 Netlify Access Token
func (s *Setting) NetlifyAccessToken() string { return s.Get("netlifyAccessToken") }

// NetlifySiteId 当前平台的 Netlify Site ID
func (s *Setting) NetlifySiteId() string { return s.Get("netlifySiteId") }

// Server 当前平台的服务器地址
func (s *Setting) Server() string { return s.Get("server") }

// Port 当前平台的端口
func (s *Setting) Port() string { return s.Get("port") }

// RemotePath 当前平台的远程路径
func (s *Setting) RemotePath() string { return s.Get("remotePath") }

// TransferProtocol 当前平台的传输协议（sftp 或 ftp）
func (s *Setting) TransferProtocol() string { return s.Get("transferProtocol") }

// Validate 校验配置数据
func (s *Setting) Validate() error {
	if s.Platform == "" {
		return errors.New("platform is required")
	}
	return nil
}

// SetPlatformConfig 设置指定平台的某个配置项
func (s *Setting) SetPlatformConfig(platform, key string, value any) {
	if s.PlatformConfigs == nil {
		s.PlatformConfigs = make(map[string]map[string]any)
	}
	m := s.PlatformConfigs[platform]
	if m == nil {
		m = make(map[string]any)
	}
	m[key] = value
	s.PlatformConfigs[platform] = m
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
