package facade

import (
	"context"
	"log/slog"

	"gridea-pro/backend/internal/deploy"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

// SettingFacade wraps SettingService
type SettingFacade struct {
	internal     *service.SettingService
	oauthService *service.OAuthService
	logger       *slog.Logger
}

func NewSettingFacade(s *service.SettingService, oauthSvc *service.OAuthService) *SettingFacade {
	return &SettingFacade{
		internal:     s,
		oauthService: oauthSvc,
		logger:       slog.Default(),
	}
}

func (f *SettingFacade) GetSetting() (domain.Setting, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.GetSetting(ctx)
}

func (f *SettingFacade) SaveAvatar(sourcePath string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SaveAvatar(ctx, sourcePath)
}

func (f *SettingFacade) SaveFavicon(sourcePath string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SaveFavicon(ctx, sourcePath)
}

// SaveSettingFromFrontend 保存设置：
// 1. 从 setting 中提取敏感字段（token/password 等）→ 存入 Keychain
// 2. 剩余非敏感字段 → 写入 setting.json
func (f *SettingFacade) SaveSettingFromFrontend(setting domain.Setting) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}

	// 1. 提取敏感字段并路由到 Keychain
	if f.oauthService != nil {
		credentials := setting.ExtractSensitiveFields()
		// 按平台分组保存
		byPlatform := make(map[string]map[string]string)
		for key, val := range credentials {
			// key 格式: "platform:field"
			for platform, fields := range domain.SensitiveFields {
				for _, field := range fields {
					if key == platform+":"+field {
						if byPlatform[platform] == nil {
							byPlatform[platform] = make(map[string]string)
						}
						byPlatform[platform][field] = val
					}
				}
			}
		}
		for platform, creds := range byPlatform {
			if err := f.oauthService.SaveManualCredentials(ctx, platform, creds); err != nil {
				f.logger.Error("保存凭证到 Keychain 失败", "platform", platform, "error", err)
			}
		}
	}

	// 2. 保存前获取旧配置，用于检测 Vercel 域名变更
	oldSetting, _ := f.internal.GetSetting(ctx)

	// 3. 保存非敏感配置到 setting.json
	if err := f.internal.SaveSetting(ctx, setting); err != nil {
		return err
	}

	// 4. Vercel 自定义域名自动绑定
	if setting.Platform == "vercel" {
		newCname := setting.CNAME()
		oldCname := oldSetting.GetFrom("vercel", "cname")
		projectName := setting.Repository()

		// 从 Keychain 取 token（已经被 ExtractSensitiveFields 从 setting 中移走）
		token := ""
		if f.oauthService != nil {
			token = f.oauthService.GetCredential("vercel", "token")
		}

		if projectName != "" && token != "" {
			proxyURL := ""
			if setting.ProxyEnabled {
				proxyURL = setting.ProxyURL
			}
			vercel := deploy.NewVercelProvider(proxyURL)

			if oldCname != "" && oldCname != newCname {
				if err := vercel.RemoveCustomDomain(ctx, projectName, oldCname, token); err != nil {
					f.logger.Error("Vercel 旧域名解绑失败", "domain", oldCname, "error", err)
				} else {
					f.logger.Info("Vercel 旧域名已解绑", "domain", oldCname)
				}
			}
			if newCname != "" {
				if err := vercel.AddCustomDomain(ctx, projectName, newCname, token); err != nil {
					f.logger.Error("Vercel 域名绑定失败", "domain", newCname, "error", err)
				} else {
					f.logger.Info("Vercel 域名绑定成功", "domain", newCname)
				}
			}
		}
	}

	return nil
}

// RemoteDetectFromFrontend 测试远程连接
// 前端传入的 setting 中若包含 token（手动输入），直接使用
// 若 token 为空，自动从 Keychain 补全
func (f *SettingFacade) RemoteDetectFromFrontend(setting domain.Setting) (map[string]interface{}, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}

	// 从 Keychain 补全空凭证（不覆盖前端已填入的值）
	if f.oauthService != nil {
		creds := f.oauthService.GetAllCredentials()
		setting.InjectCredentials(creds)
	}

	return f.internal.RemoteDetect(ctx, setting)
}
