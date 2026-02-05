package comment

import (
	"errors"
	"gridea-pro/backend/internal/domain"
)

// NewProvider 创建评论提供者
func NewProvider(settings domain.CommentSettings) (domain.CommentProvider, error) {
	if !settings.Enable {
		return nil, errors.New("comment system is disabled")
	}

	// 获取当前平台的配置
	config := settings.PlatformConfigs[settings.Platform]
	if config == nil {
		config = make(map[string]any)
	}

	switch settings.Platform {
	case domain.CommentPlatformValine:
		appID, _ := config["appId"].(string)
		appKey, _ := config["appKey"].(string)
		masterKey, _ := config["masterKey"].(string)
		serverURLs, _ := config["serverURLs"].(string)
		if appID == "" || appKey == "" {
			return nil, errors.New("Valine config missing AppID or AppKey")
		}
		return NewValineProvider(appID, appKey, masterKey, serverURLs), nil

	case domain.CommentPlatformTwikoo:
		envID, _ := config["envId"].(string)
		if envID == "" {
			return nil, errors.New("Twikoo config missing EnvID")
		}
		return NewTwikooProvider(envID), nil

	case domain.CommentPlatformWaline:
		appID, _ := config["appId"].(string)
		appKey, _ := config["appKey"].(string)
		masterKey, _ := config["masterKey"].(string)
		serverURLs, _ := config["serverURLs"].(string)
		if serverURLs == "" {
			return nil, errors.New("Waline config missing ServerURLs")
		}
		return NewWalineProvider(appID, appKey, masterKey, serverURLs), nil

	case domain.CommentPlatformGitalk:
		owner, _ := config["owner"].(string)
		repo, _ := config["repo"].(string)
		clientID, _ := config["clientId"].(string)
		clientSecret, _ := config["clientSecret"].(string)

		if owner == "" || repo == "" {
			return nil, errors.New("Gitalk config missing Owner or Repo")
		}
		return NewGitHubProvider(owner, repo, clientID, clientSecret), nil

	case domain.CommentPlatformGiscus:
		return nil, errors.New("Giscus provider not implemented yet")

	case domain.CommentPlatformDisqus:
		return NewDisqusProvider(config), nil

	default:
		return nil, errors.New("unsupported comment platform")
	}
}
