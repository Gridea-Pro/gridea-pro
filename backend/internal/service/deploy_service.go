package service

import (
	"context"
	"errors"
	"fmt"
	"gridea-pro/backend/internal/deploy"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/engine"
	"gridea-pro/backend/internal/notify"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// defaultDeployTimeout 单次部署的总超时上限，超过就强制取消。
// 30 分钟对绝大多数站点（含大文件 CDN 上传）足够宽松，同时保证 deploy 卡死
// 不会无限占用互斥锁 —— 即便 provider 某个阻塞点对 ctx 不敏感，HTTP
// 请求 / 网络 I/O 也会在 ctx 超时后被 Transport 取消。
const defaultDeployTimeout = 30 * time.Minute

type DeployService struct {
	settingRepo      domain.SettingRepository
	renderer         *engine.Engine // Injected to trigger site build before deploy
	cdnUploadService *CdnUploadService
	oauthService     *OAuthService // 用于从 Keychain 补全凭证
	appDir           string
	knownHostsPath   string // SFTP HostKey TOFU 校验文件路径（跨站点共享，见 #37）
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

// SetKnownHostsPath 注入 known_hosts 路径，生产环境应在 bootstrap 时设置为
// AppConfigDir/known_hosts，以便 SFTP Provider 做 HostKey TOFU 校验。
func (s *DeployService) SetKnownHostsPath(path string) {
	s.knownHostsPath = path
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

func (s *DeployService) DeployToRemote(ctx context.Context) (err error) {
	// 两层 ctx 叠加：
	// - 外层 WithCancel：暴露 cancel 给 CancelDeploy 调用（#42 / PR #93 可取消）
	// - 内层 WithTimeout：避免 SFTP io.Copy / FTP Stor 等对 ctx 不敏感的阻塞点永久
	//   占锁（#49 / PR #87 总超时）。每个 provider 内部 HTTP 请求通过
	//   NewRequestWithContext 继承这个 ctx，超时后 Transport 层会主动取消并释放
	//   goroutine。
	cancelCtx, cancelFn := context.WithCancel(ctx)

	s.mu.Lock()
	if s.isDeploying {
		s.mu.Unlock()
		cancelFn()
		return fmt.Errorf(domain.ErrDeployInProgress)
	}
	s.isDeploying = true
	s.activeCancel = cancelFn
	s.mu.Unlock()

	deployCtx, timeoutCancel := context.WithTimeout(cancelCtx, defaultDeployTimeout)
	defer timeoutCancel()
	ctx = deployCtx

	// 通知用闭包变量：声明在 defer 之前，部署进展中再填实际值。
	startedAt := time.Now()
	var platform, siteDomain string
	var deployedPlatforms []string

	// Ensure we reset the flag when done
	defer func() {
		s.mu.Lock()
		s.isDeploying = false
		s.activeCancel = nil
		s.mu.Unlock()
		cancelFn()
		// 多平台时使用汇总信息
		notifyPlatform := platform
		notifyDomain := siteDomain
		if len(deployedPlatforms) > 1 {
		notifyPlatform = strings.Join(deployedPlatforms, " / ")
		notifyDomain = fmt.Sprintf("%d 个平台", len(deployedPlatforms))
		}
		// 部署结束（成功 / 失败 / 取消）后发系统通知中心；偏好关闭则跳过。
		s.notifyDeployResult(err, time.Since(startedAt), notifyPlatform, notifyDomain)
	}()

	s.log(ctx, "Starting deployment check...")

	// 1. Get Settings safely，并从 Keychain 补全凭证
	setting, err := s.settingRepo.GetSetting(ctx)
	if err != nil {
		s.log(ctx, fmt.Sprintf("Failed to load settings: %v", err))
		return err
	}
	// 双保险：即使上游 GetSetting 未返回深拷贝，这里也再 Clone 一次，
	// 避免 InjectCredentials 写入 PlatformConfigs 反向污染 repo cache（issue #39）
	setting = setting.Clone()
	if s.oauthService != nil {
		creds := s.oauthService.GetAllCredentials()
		setting.InjectCredentials(creds)
	}
	platformsToDeploy := setting.EnabledPlatforms
	if len(platformsToDeploy) == 0 {
		platformsToDeploy = []string{setting.Platform}
	}

	s.log(ctx, fmt.Sprintf("Platforms to deploy: %v", platformsToDeploy))

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
			// CDN 配置加载失败属于致命错误：再继续走下去 result 是零值，
			// cdnFailureAbortReason 会判定无失败而放行，最终用旧 CDN URL 上线、
			// 新加的图片全 404。必须 fail-fast，让用户立刻看到具体错误（如 token 失效）。
			s.log(ctx, fmt.Sprintf("❌ CDN 配置加载失败: %v", err))
			return fmt.Errorf("CDN 配置加载失败: %w", err)
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

	// 4. Wrap log function
	logger := func(msg string) {
		s.log(ctx, msg)
	}

	// 5. Deploy to each enabled platform
	var deployErrors []string
	for _, p := range platformsToDeploy {
		s.log(ctx, fmt.Sprintf("Deploying to %s...", p))

		// Clone setting and set active platform for per-platform config access
		platformSetting := setting.Clone()
		platformSetting.Platform = p

		platform = platformSetting.Platform
		siteDomain = platformSetting.Domain()

		var provider deploy.Provider
		switch platform {
		case "github", "coding":
			provider = deploy.NewGitProvider()
		case "vercel":
			proxyURL := ""
			if platformSetting.ProxyEnabled {
				proxyURL = platformSetting.ProxyURL
			}
			provider = deploy.NewVercelProvider(proxyURL)
		case "netlify":
			proxyURL := ""
			if platformSetting.ProxyEnabled {
				proxyURL = platformSetting.ProxyURL
			}
			provider = deploy.NewNetlifyProvider(proxyURL)
		case "sftp":
			if platformSetting.TransferProtocol() == "ftp" {
				provider = deploy.NewFtpProvider()
			} else {
				provider = deploy.NewSftpProviderWithKnownHosts(s.knownHostsPath)
			}
		default:
			provider = deploy.NewGitProvider()
		}

		if err := provider.Deploy(ctx, outputDir, &platformSetting, logger); err != nil {
			errMsg := fmt.Sprintf("%s: %v", platform, err)
			s.log(ctx, fmt.Sprintf("❌ %s 部署失败: %v", platform, err))
			deployErrors = append(deployErrors, errMsg)
		} else {
			s.log(ctx, fmt.Sprintf("✅ %s 部署成功", platform))
			deployedPlatforms = append(deployedPlatforms, platform)
		}
	}

	if len(deployErrors) > 0 {
		return fmt.Errorf("部分平台部署失败: %s", strings.Join(deployErrors, "; "))
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

// notifyDeployResult 部署结束后发系统通知中心通知。
// 用户在偏好里关闭后跳过；通知发送失败不影响主流程。
// 用独立的 context.Background() 读设置，避免部署 ctx 已超时/取消导致读不到。
func (s *DeployService) notifyDeployResult(deployErr error, dur time.Duration, platform, siteDomain string) {
	setting, err := s.settingRepo.GetSetting(context.Background())
	if err != nil || !setting.IsDeployNotifyEnabled() {
		return
	}
	site := siteDomain
	if site == "" {
		site = "—"
	}
	plat := platform
	if plat == "" {
		plat = "—"
	}
	var title, body string
	switch {
	case deployErr == nil:
		title = "部署成功"
		body = fmt.Sprintf("%s · %s · 用时 %s", site, plat, formatDeployDuration(dur))
	case errors.Is(deployErr, context.Canceled):
		title = "部署已取消"
		body = fmt.Sprintf("%s · %s · 已停止", site, plat)
	default:
		title = "部署失败"
		body = fmt.Sprintf("%s · %s · %s", site, plat, truncateRunes(deployErr.Error(), 100))
	}
	_ = notify.Send(title, body)
}

func formatDeployDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d 秒", int(d.Seconds()))
	}
	return fmt.Sprintf("%d 分 %d 秒", int(d.Minutes()), int(d.Seconds())%60)
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}

// cdn 上传失败阈值。两条独立路径，命中任一即中止部署：
//   - 灾难性失败率（>= 50%）：覆盖小站全裂场景（3/3 全失败、4/5 失败等），
//     不再受 cdnFailureAbsoluteCap 保护——失败比例已经说明是确定性问题（token 失效、
//     配置错），不是 transient 噪声。
//   - 常规双门槛：比例（>= 10%）AND 绝对数（>= 5），避免大站被 1~2 个噪声误差打断。
const (
	cdnFailureRatioThreshold = 0.10
	cdnFailureAbsoluteCap    = 5
	cdnCatastrophicRatio     = 0.50
)

// cdnFailureAbortReason 判断是否因 CDN 上传失败过多而中止部署。
// 返回空串表示可以继续；返回非空表示应中止，字符串即用户友好原因。
func cdnFailureAbortReason(r CdnUploadResultShape) string {
	if r.GetTotal() == 0 || len(r.GetFailures()) == 0 {
		return ""
	}
	failed := len(r.GetFailures())
	ratio := float64(failed) / float64(r.GetTotal())
	if ratio >= cdnCatastrophicRatio ||
		(ratio >= cdnFailureRatioThreshold && failed >= cdnFailureAbsoluteCap) {
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
