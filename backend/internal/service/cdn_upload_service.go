package service

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gridea-pro/backend/internal/domain"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/net/proxy"
	"golang.org/x/sync/errgroup"
)

const cdnAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

type CdnUploadService struct {
	cdnSettingRepo domain.CdnSettingRepository
	settingRepo    domain.SettingRepository
	appDir         string

	clientMu       sync.Mutex
	cachedClient   *http.Client
	cachedProxyURL string
}

func NewCdnUploadService(cdnSettingRepo domain.CdnSettingRepository, settingRepo domain.SettingRepository, appDir string) *CdnUploadService {
	return &CdnUploadService{
		cdnSettingRepo: cdnSettingRepo,
		settingRepo:    settingRepo,
		appDir:         appDir,
	}
}

// newHTTPClient 创建支持代理的 HTTP client，支持 HTTP/HTTPS/SOCKS 协议
func newHTTPClient(proxyURL string) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	}
	if proxyURL != "" {
		if u, err := url.Parse(proxyURL); err == nil {
			switch strings.ToLower(u.Scheme) {
			case "socks4", "socks4a", "socks5", "socks":
				if dialer, err := proxy.FromURL(u, proxy.Direct); err == nil {
					transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
						return dialer.Dial(network, addr)
					}
				}
			default:
				transport.Proxy = http.ProxyURL(u)
			}
		}
	}
	return &http.Client{
		Timeout:   60 * time.Second,
		Transport: transport,
	}
}

// httpClient 根据当前代理设置返回合适的 HTTP client。
// 代理 client 会被缓存复用，只有代理地址变更时才重建，保证连接池有效。
func (s *CdnUploadService) httpClient(ctx context.Context) *http.Client {
	proxyURL := ""
	if s.settingRepo != nil {
		setting, err := s.settingRepo.GetSetting(ctx)
		if err == nil && setting.ProxyEnabled && setting.ProxyURL != "" {
			proxyURL = setting.ProxyURL
		}
	}

	if proxyURL == "" {
		return http.DefaultClient
	}

	s.clientMu.Lock()
	defer s.clientMu.Unlock()

	if s.cachedClient != nil && s.cachedProxyURL == proxyURL {
		return s.cachedClient
	}

	s.cachedClient = newHTTPClient(proxyURL)
	s.cachedProxyURL = proxyURL
	return s.cachedClient
}

// ResolveSavePath 解析路径模板变量
func ResolveSavePath(template, filename string) string {
	now := time.Now()
	ext := filepath.Ext(filename)
	nameOnly := strings.TrimSuffix(filename, ext)

	replacer := strings.NewReplacer(
		"{year}", now.Format("2006"),
		"{month}", now.Format("01"),
		"{day}", now.Format("02"),
		"{hour}", now.Format("15"),
		"{minute}", now.Format("04"),
		"{second}", now.Format("05"),
		"{since_second}", fmt.Sprintf("%d", now.Unix()),
		"{since_millisecond}", fmt.Sprintf("%d", now.UnixMilli()),
		"{random}", randomString(12),
		"{filename}", nameOnly,
		"{.suffix}", ext,
		"{suffix}", strings.TrimPrefix(ext, "."),
	)

	return replacer.Replace(template)
}

func randomString(n int) string {
	id, _ := gonanoid.Generate(cdnAlphabet, n)
	return id
}

// githubContentsResponse GitHub Contents API 响应
type githubContentsResponse struct {
	SHA string `json:"sha"`
}

// uploadToGitHub 通过 GitHub Contents API 上传单个文件。
// manifest 非 nil 时走本地 manifest 缓存（#45）：若 remotePath 在上次成功上传后
// 仍是同一个 SHA，直接跳过整个 API 往返；否则仍需一次 GET 查远端 SHA 做增量。
// 成功上传后向 manifest 记录新的 (remotePath → localSHA)。
func (s *CdnUploadService) uploadToGitHub(ctx context.Context, setting domain.CdnSetting, localFilePath, remotePath string, manifest *cdnManifest) error {
	// 读取本地文件
	data, err := os.ReadFile(localFilePath)
	if err != nil {
		return fmt.Errorf("读取文件失败 %s: %w", localFilePath, err)
	}

	branch := setting.GithubBranch
	if branch == "" {
		branch = "main"
	}

	// 计算本地文件 SHA（git blob SHA1）
	localSHA := gitBlobSHA(data)

	// 快速路径：本地 manifest 命中且 SHA 未变，直接跳过 API。
	// 放开这条后大部分"重复部署"场景都不再调用 GitHub API，配额保留给真正的变更。
	if manifest != nil {
		if cached, ok := manifest.hit(remotePath); ok && cached == localSHA {
			return nil
		}
	}

	// 慢路径：manifest 未命中或 SHA 变更，还需一次 GET 拿远端 SHA 做增量更新
	existingSHA, err := s.getGithubFileSHA(ctx, setting, remotePath, branch)
	if err == nil && existingSHA == localSHA {
		// 文件内容相同，跳过上传，但补录到 manifest 避免下次再 GET
		if manifest != nil {
			manifest.record(remotePath, localSHA)
		}
		return nil
	}

	// 构建请求体
	content := base64.StdEncoding.EncodeToString(data)
	body := map[string]any{
		"message": fmt.Sprintf("Upload %s via Gridea Pro", remotePath),
		"content": content,
		"branch":  branch,
	}
	if existingSHA != "" {
		body["sha"] = existingSHA
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s",
		setting.GithubUser, setting.GithubRepo, remotePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+setting.GithubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient(ctx).Do(req)
	if err != nil {
		return fmt.Errorf("上传失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		respStr := string(respBody)

		switch resp.StatusCode {
		case http.StatusNotFound:
			if strings.Contains(respStr, "Branch") && strings.Contains(respStr, "not found") {
				return fmt.Errorf("分支 %s 不存在", branch)
			}
			return fmt.Errorf("仓库不存在或无权限")
		case http.StatusUnauthorized, http.StatusForbidden:
			return fmt.Errorf("Token 无效或权限不足")
		case http.StatusConflict:
			return fmt.Errorf("文件冲突，请重试")
		default:
			return fmt.Errorf("上传失败 (%d)", resp.StatusCode)
		}
	}

	// 成功：把新 SHA 写入 manifest，后续部署可以直接走快速路径
	if manifest != nil {
		manifest.record(remotePath, localSHA)
	}
	return nil
}

// getGithubFileSHA 获取 GitHub 上文件的 SHA
func (s *CdnUploadService) getGithubFileSHA(ctx context.Context, setting domain.CdnSetting, remotePath, branch string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		setting.GithubUser, setting.GithubRepo, remotePath, branch)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+setting.GithubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := s.httpClient(ctx).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("文件不存在")
	}

	var result githubContentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.SHA, nil
}

// gitBlobSHA 计算 git blob 的 SHA1（与 GitHub API 一致）
func gitBlobSHA(data []byte) string {
	header := fmt.Sprintf("blob %d\x00", len(data))
	h := sha1.New()
	h.Write([]byte(header))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// TestUpload 测试上传：上传一个小的测试文件到 CDN 仓库
func (s *CdnUploadService) TestUpload(ctx context.Context) (string, error) {
	setting, err := s.cdnSettingRepo.GetCdnSetting(ctx)
	if err != nil {
		return "", fmt.Errorf("读取 CDN 配置失败: %w", err)
	}

	if !setting.Enabled {
		return "", fmt.Errorf(domain.ErrCdnNotEnabled)
	}

	if setting.GithubToken == "" {
		return "", fmt.Errorf(domain.ErrCdnTokenMissing)
	}

	// 创建测试文件内容
	testContent := []byte("Gridea Pro CDN Upload Test - " + time.Now().Format("2006-01-02 15:04:05"))

	// 写入临时文件
	tmpFile, err := os.CreateTemp("", "gridea-cdn-test-*.txt")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(testContent); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("写入临时文件失败: %w", err)
	}
	tmpFile.Close()

	// 解析保存路径
	savePath := setting.SavePath
	if savePath == "" {
		savePath = "{year}/{month}/{filename}{.suffix}"
	}
	remotePath := ResolveSavePath(savePath, "gridea-test.txt")

	// 上传（测试场景不走 manifest 缓存：每次都实际触达 API 以便验证配置）
	if err := s.uploadToGitHub(ctx, setting, tmpFile.Name(), remotePath, nil); err != nil {
		return "", err
	}

	// 构建 CDN 访问 URL
	cdnURL := s.buildCdnURL(setting, remotePath)
	return cdnURL, nil
}

// UploadMediaForDeploy 部署时扫描并上传媒体文件到 CDN
func (s *CdnUploadService) UploadMediaForDeploy(ctx context.Context, appDir string, logger func(string)) error {
	setting, err := s.cdnSettingRepo.GetCdnSetting(ctx)
	if err != nil {
		return fmt.Errorf("读取 CDN 配置失败: %w", err)
	}

	if !setting.Enabled || setting.GithubToken == "" {
		return nil
	}

	// 加载本地 manifest：命中的文件跳过整轮 API 调用（#45）
	manifest := loadCdnManifest(appDir)
	defer func() {
		// 无论成功失败都持久化已有进度，单文件失败不影响其它文件的快路径
		if err := manifest.save(appDir); err != nil {
			logger(fmt.Sprintf("警告：写入 CDN manifest 失败（下次将重新全量检查）: %v", err))
		}
	}()

	// 需要扫描的目录
	mediaDirs := []string{"post-images", "images", "media"}
	var filesToUpload []struct {
		localPath  string
		remotePath string
	}

	for _, dir := range mediaDirs {
		dirPath := filepath.Join(appDir, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}

			// 获取相对路径（如 post-images/cover.png）
			relPath, err := filepath.Rel(appDir, path)
			if err != nil {
				return err
			}
			// 统一为正斜杠
			relPath = filepath.ToSlash(relPath)

			filesToUpload = append(filesToUpload, struct {
				localPath  string
				remotePath string
			}{
				localPath:  path,
				remotePath: relPath,
			})

			return nil
		})
		if err != nil {
			logger(fmt.Sprintf("扫描目录 %s 失败: %v", dir, err))
		}
	}

	if len(filesToUpload) == 0 {
		logger("没有需要上传的媒体文件")
		return nil
	}

	logger(fmt.Sprintf("发现 %d 个媒体文件，开始上传到 CDN...", len(filesToUpload)))

	// 使用 errgroup 控制并发（限制 5 并发）
	g, gCtx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, 5)
	var uploadCount int
	var mu sync.Mutex

	for _, file := range filesToUpload {
		f := file
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			if err := s.uploadToGitHub(gCtx, setting, f.localPath, f.remotePath, manifest); err != nil {
				logger(fmt.Sprintf("上传 %s 失败: %v", f.remotePath, err))
				return nil // 单个文件失败不中断整个上传
			}

			mu.Lock()
			uploadCount++
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("上传媒体文件失败: %w", err)
	}

	// 清理孤儿：manifest 中有、但本次扫描不存在的 remotePath。
	// 只基于 manifest 做对比，不做"列整库 diff"，天然安全：
	// 如果用户在同一个 CDN 仓库里还有其它非 Gridea 上传的文件，不会被误删。
	localSet := make(map[string]struct{}, len(filesToUpload))
	for _, f := range filesToUpload {
		localSet[f.remotePath] = struct{}{}
	}
	var orphans []string
	for remotePath := range manifest.Entries {
		if _, exists := localSet[remotePath]; !exists {
			orphans = append(orphans, remotePath)
		}
	}
	if len(orphans) > 0 {
		logger(fmt.Sprintf("检测到 %d 个 CDN 孤儿文件（本地已删除但 CDN 仍存在），开始清理...", len(orphans)))
		deleted := 0
		for _, remotePath := range orphans {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := s.deleteFromGitHub(ctx, setting, remotePath); err != nil {
				// 单文件删除失败不阻塞：下次部署还会再尝试
				logger(fmt.Sprintf("清理 %s 失败（将下次重试）: %v", remotePath, err))
				continue
			}
			manifest.mu.Lock()
			delete(manifest.Entries, remotePath)
			manifest.mu.Unlock()
			deleted++
		}
		logger(fmt.Sprintf("CDN 孤儿清理完成：成功删除 %d / %d 个文件", deleted, len(orphans)))
	}

	logger(fmt.Sprintf("CDN 上传完成，共上传 %d 个文件", uploadCount))
	return nil
}

// deleteFromGitHub 从 GitHub CDN 仓库里删除一个已上传的文件。
// 先 GET 当前 SHA，再 DELETE（Contents API 要求提供 sha 防覆盖冲突）。
func (s *CdnUploadService) deleteFromGitHub(ctx context.Context, setting domain.CdnSetting, remotePath string) error {
	branch := setting.GithubBranch
	if branch == "" {
		branch = "main"
	}

	// 1. 拿远端 SHA；404 认为已经不存在，视作删除成功
	sha, err := s.getGithubFileSHA(ctx, setting, remotePath, branch)
	if err != nil {
		// getGithubFileSHA 把"文件不存在"当成 error —— 这里视作孤儿已消失
		return nil
	}

	// 2. DELETE 请求
	body := map[string]any{
		"message": fmt.Sprintf("Delete orphan %s via Gridea Pro", remotePath),
		"sha":     sha,
		"branch":  branch,
	}
	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s",
		setting.GithubUser, setting.GithubRepo, remotePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+setting.GithubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient(ctx).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		if resp.StatusCode == http.StatusNotFound {
			return nil // 已消失
		}
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}

// buildCdnURL 构建 CDN 访问 URL
func (s *CdnUploadService) buildCdnURL(setting domain.CdnSetting, remotePath string) string {
	switch setting.Provider {
	case "jsdelivr":
		branch := setting.GithubBranch
		if branch == "" {
			branch = "main"
		}
		return fmt.Sprintf("https://cdn.jsdelivr.net/gh/%s/%s@%s/%s",
			setting.GithubUser, setting.GithubRepo, branch, remotePath)
	case "custom":
		return strings.TrimRight(setting.BaseURL, "/") + "/" + remotePath
	default:
		return remotePath
	}
}
