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

// uploadToGitHub 通过 GitHub Contents API 上传单个文件
func (s *CdnUploadService) uploadToGitHub(ctx context.Context, setting domain.CdnSetting, localFilePath, remotePath string) error {
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

	// 检查文件是否已存在
	existingSHA, err := s.getGithubFileSHA(ctx, setting, remotePath, branch)
	if err == nil && existingSHA == localSHA {
		// 文件内容相同，跳过
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

	// 上传
	if err := s.uploadToGitHub(ctx, setting, tmpFile.Name(), remotePath); err != nil {
		return "", err
	}

	// 构建 CDN 访问 URL
	cdnURL := s.buildCdnURL(setting, remotePath)
	return cdnURL, nil
}

// UploadFailure 描述一次 CDN 上传失败，用于汇总向上报告（见 UploadResult）。
type UploadFailure struct {
	Path  string // 相对仓库根的远程路径（例如 "post-images/cover.png"）
	Error string // 人类可读错误（已脱敏，不包含 token）
}

// UploadResult 汇总 CDN 批量上传的结果。
// 调用方据此决定是否中止部署 / 展示失败列表。
type UploadResult struct {
	Total    int
	Success  int
	Failures []UploadFailure
}

// UploadMediaForDeploy 部署时扫描并上传媒体文件到 CDN。
// 返回 UploadResult 供调用方做失败汇总 / 阈值判断；error 仅用于"整体流程失败"
// （如读配置失败）。单文件失败会计入 Failures 但不直接返回 error。
func (s *CdnUploadService) UploadMediaForDeploy(ctx context.Context, appDir string, logger func(string)) (UploadResult, error) {
	var res UploadResult

	setting, err := s.cdnSettingRepo.GetCdnSetting(ctx)
	if err != nil {
		return res, fmt.Errorf("读取 CDN 配置失败: %w", err)
	}

	if !setting.Enabled || setting.GithubToken == "" {
		return res, nil
	}

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

	res.Total = len(filesToUpload)
	if res.Total == 0 {
		logger("没有需要上传的媒体文件")
		return res, nil
	}

	logger(fmt.Sprintf("发现 %d 个媒体文件，开始上传到 CDN...", res.Total))

	// 使用 errgroup 控制并发（限制 5 并发）。
	// 单文件失败不终止 errgroup（保持"尽量完成"语义），但会被收集到 failures 列表。
	g, gCtx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, 5)
	var mu sync.Mutex
	var failures []UploadFailure
	var successCount int

	for _, file := range filesToUpload {
		f := file
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			if err := s.uploadToGitHub(gCtx, setting, f.localPath, f.remotePath); err != nil {
				logger(fmt.Sprintf("上传 %s 失败: %v", f.remotePath, err))
				mu.Lock()
				failures = append(failures, UploadFailure{Path: f.remotePath, Error: err.Error()})
				mu.Unlock()
				return nil // 单个文件失败不中断整个上传
			}

			mu.Lock()
			successCount++
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return res, fmt.Errorf("上传媒体文件失败: %w", err)
	}

	res.Success = successCount
	res.Failures = failures

	if len(failures) == 0 {
		logger(fmt.Sprintf("CDN 上传完成，共上传 %d 个文件", res.Success))
	} else {
		logger(fmt.Sprintf("CDN 上传完成：成功 %d / 总数 %d（失败 %d 个，详见下方列表）",
			res.Success, res.Total, len(failures)))
		// 摘要列出头几条失败，避免把日志撑爆；调用方会拿到完整 Failures 做决策
		const previewN = 5
		for i, f := range failures {
			if i >= previewN {
				logger(fmt.Sprintf("  ...（其余 %d 个失败未列出）", len(failures)-previewN))
				break
			}
			logger(fmt.Sprintf("  ✗ %s: %s", f.Path, f.Error))
		}
	}
	return res, nil
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
