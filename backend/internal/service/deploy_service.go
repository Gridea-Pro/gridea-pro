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
	activeCancel     context.CancelFunc // 当前部署的取消函数；空闲时为 nil（issue #42）
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
	// 为本次部署创建可取消的 ctx，暴露 cancel 给 CancelDeploy 调用。
	deployCtx, cancel := context.WithCancel(ctx)

	s.mu.Lock()
	if s.isDeploying {
		s.mu.Unlock()
		cancel()
		return fmt.Errorf(domain.ErrDeployInProgress)
	}
	s.isDeploying = true
	s.activeCancel = cancel
	s.mu.Unlock()
	ctx = deployCtx

	// Ensure we reset the flag when done
	defer func() {
		s.mu.Lock()
		s.isDeploying = false
		s.activeCancel = nil
		s.mu.Unlock()
		cancel()
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

// CancelDeploy 中断当前正在进行的部署。
// 若当前空闲则 no-op，不返回错误。取消后 DeployToRemote 会收到 ctx.Canceled
// 并尽可能快地退出（各 provider 内部的 HTTP / walk 循环都要尊重 ctx）。
func (s *DeployService) CancelDeploy() {
	s.mu.Lock()
	cancel := s.activeCancel
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// IsDeploying 返回当前是否有部署在进行中，供前端按钮状态同步使用。
func (s *DeployService) IsDeploying() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isDeploying
}

// log sends a message to the frontend safely
func (s *DeployService) log(ctx context.Context, msg string) {
	if ctx != nil {
		runtime.EventsEmit(ctx, "deploy-log", msg)
	}
}
