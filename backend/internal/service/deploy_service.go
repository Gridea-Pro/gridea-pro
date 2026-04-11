package service

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/deploy"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/engine"
	"os"
	"path/filepath"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type DeployService struct {
	settingRepo      domain.SettingRepository
	renderer         *engine.Engine // Injected to trigger site build before deploy
	cdnUploadService *CdnUploadService
	oauthService     *OAuthService // 用于从 Keychain 补全凭证
	appDir           string
	mu               sync.Mutex
	isDeploying      bool
}

func NewDeployService(settingRepo domain.SettingRepository, appDir string) *DeployService {
	return &DeployService{
		settingRepo: settingRepo,
		appDir:      appDir,
	}
}

// SetOAuthService 注入 OAuthService（用于从 Keychain 读取凭证）
func (s *DeployService) SetOAuthService(oauthSvc *OAuthService) {
	s.oauthService = oauthSvc
}

// SetRenderer injects the RendererService into DeployService
func (s *DeployService) SetRenderer(renderer *engine.Engine) {
	s.renderer = renderer
}

// SetCdnUploadService injects the CdnUploadService into DeployService
func (s *DeployService) SetCdnUploadService(cdnUpload *CdnUploadService) {
	s.cdnUploadService = cdnUpload
}

func (s *DeployService) DeployToRemote(ctx context.Context) error {
	s.mu.Lock()
	if s.isDeploying {
		s.mu.Unlock()
		return fmt.Errorf(domain.ErrDeployInProgress)
	}
	s.isDeploying = true
	s.mu.Unlock()

	// Ensure we reset the flag when done
	defer func() {
		s.mu.Lock()
		s.isDeploying = false
		s.mu.Unlock()
	}()

	s.log(ctx, "Starting deployment check...")

	// 1. Get Settings safely，并从 Keychain 补全凭证
	setting, err := s.settingRepo.GetSetting(ctx)
	if err != nil {
		s.log(ctx, fmt.Sprintf("Failed to load settings: %v", err))
		return err
	}
	if s.oauthService != nil {
		creds := s.oauthService.GetAllCredentials()
		setting.InjectCredentials(creds)
	}

	s.log(ctx, fmt.Sprintf("Deploying to domain: %s", setting.Domain()))

	// 2. Render Site
	if s.renderer != nil {
		s.log(ctx, "Building static site...")
		if err := s.renderer.RenderAll(ctx); err != nil {
			s.log(ctx, fmt.Sprintf("Failed to build site: %v", err))
			return fmt.Errorf("render site failed: %w", err)
		}
	} else {
		s.log(ctx, "Warning: Renderer service not attached, skipping build.")
	}

	// 2.5 CDN 上传媒体文件
	if s.cdnUploadService != nil {
		s.log(ctx, "Uploading media files to CDN...")
		if err := s.cdnUploadService.UploadMediaForDeploy(ctx, s.appDir, func(msg string) {
			s.log(ctx, msg)
		}); err != nil {
			s.log(ctx, fmt.Sprintf("CDN upload warning: %v", err))
		}
	}

	// 3. Prepare Git Repository Path
	outputDir := filepath.Join(s.appDir, "output")
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		_ = os.MkdirAll(outputDir, 0755) // Ensure it exists before Git operations if not already
	}

	// 4. Instantiate strategy based on platform
	var provider deploy.Provider
	switch setting.Platform {
	case "github", "gitee", "coding":
		provider = deploy.NewGitProvider()
	case "vercel":
		proxyURL := ""
		if setting.ProxyEnabled {
			proxyURL = setting.ProxyURL
		}
		provider = deploy.NewVercelProvider(proxyURL)
	case "netlify":
		proxyURL := ""
		if setting.ProxyEnabled {
			proxyURL = setting.ProxyURL
		}
		provider = deploy.NewNetlifyProvider(proxyURL)
	case "sftp":
		if setting.TransferProtocol() == "ftp" {
			provider = deploy.NewFtpProvider()
		} else {
			provider = deploy.NewSftpProvider()
		}
	default:
		provider = deploy.NewGitProvider()
	}

	// 5. Wrap log function
	logger := func(msg string) {
		s.log(ctx, msg)
	}

	// 6. Execute deployment (without buildSite callback)
	if err := provider.Deploy(ctx, outputDir, &setting, logger); err != nil {
		return err
	}

	return nil
}

// log sends a message to the frontend safely
func (s *DeployService) log(ctx context.Context, msg string) {
	if ctx != nil {
		runtime.EventsEmit(ctx, "deploy-log", msg)
	}
}
