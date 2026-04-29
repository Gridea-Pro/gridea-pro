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

	// 2.5 CDN 上传媒体文件。
	// 单文件失败不终止整组，UploadMediaForDeploy 返回 UploadResult 汇总成功 / 失败清单。
	// 失败占比超过阈值时中止部署，避免"toast 成功、线上图片大面积 404"的隐性故障（#44）。
	if s.cdnUploadService != nil {
		s.log(ctx, "Uploading media files to CDN...")
		result, err := s.cdnUploadService.UploadMediaForDeploy(ctx, s.appDir, func(msg string) {
			s.log(ctx, msg)
		})
		if err != nil {
			s.log(ctx, fmt.Sprintf("CDN upload warning: %v", err))
		}
		if reason := cdnFailureAbortReason(result); reason != "" {
			s.log(ctx, fmt.Sprintf("❌ %s，已中止部署以避免上线图片 404", reason))
			return fmt.Errorf("CDN 上传失败率过高：%s", reason)
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

// cdn 上传失败阈值：超过任一条件都视为"过多"，部署中止。
// 比例偏保守（10%），绝对数给定下限（5）避免小图库被 1~2 个误差锁死。
const (
	cdnFailureRatioThreshold = 0.10
	cdnFailureAbsoluteCap    = 5
)

// cdnFailureAbortReason 判断是否因 CDN 上传失败过多而中止部署。
// 返回空串表示可以继续；返回非空表示应中止，字符串即用户友好原因。
func cdnFailureAbortReason(r CdnUploadResultShape) string {
	if r.GetTotal() == 0 || len(r.GetFailures()) == 0 {
		return ""
	}
	failed := len(r.GetFailures())
	ratio := float64(failed) / float64(r.GetTotal())
	if ratio >= cdnFailureRatioThreshold && failed >= cdnFailureAbsoluteCap {
		return fmt.Sprintf("%d 个文件失败（共 %d 个，占比 %.0f%%）", failed, r.GetTotal(), ratio*100)
	}
	return ""
}

// CdnUploadResultShape 是对 UploadResult 的抽象，用于在不跨包循环依赖的前提下
// 在 service 包里做阈值判断。具体类型为 *UploadResult。
type CdnUploadResultShape interface {
	GetTotal() int
	GetFailures() []UploadFailure
}

// 让 UploadResult 满足 CdnUploadResultShape —— 方法放这里是为了让阈值函数能在
// 同一个 service 包内引用，不需要暴露到 domain 层。
func (r UploadResult) GetTotal() int                 { return r.Total }
func (r UploadResult) GetFailures() []UploadFailure  { return r.Failures }
