package deploy

import (
	"bytes"
	"context"
	"crypto/sha1"
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
	"time"

	"gridea-pro/backend/internal/domain"

	"golang.org/x/net/proxy"
	"golang.org/x/sync/errgroup"
)

// VercelProvider 实现了 Vercel API 直传部署策略
type VercelProvider struct {
	client *http.Client
}

// NewVercelProvider 创建 VercelProvider，proxyURL 为空则不使用代理
func NewVercelProvider(proxyURL string) *VercelProvider {
	return &VercelProvider{client: newVercelHTTPClient(proxyURL)}
}

// newVercelHTTPClient 创建支持代理的 HTTP client，支持 HTTP/HTTPS/SOCKS 协议
func newVercelHTTPClient(proxyURL string) *http.Client {
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

// VercelFileResult 表示用于创建部署的文件哈希映射
type VercelFileResult struct {
	File string `json:"file"`
	Sha  string `json:"sha"`
	Size int64  `json:"size"`
}

// Deploy 实现了 Provider 接口。
//
// 流程：扫描文件 → 创建部署（服务器返回 missing 列表） → 上传 missing →
// 再次创建部署 … 循环直到 missing 为空或达到 maxRounds 上限（避免大文件 /
// 网络抖动场景下被服务端吞文件却静默宣布成功，见 issue #48）。
func (p *VercelProvider) Deploy(ctx context.Context, outputDir string, setting *domain.Setting, logger LogFunc) error {
	const maxRounds = 3

	logger("🚀 开始准备 Vercel 部署...")

	projectName := setting.Repository()
	if projectName == "" {
		projectName = setting.Username()
	}
	if projectName == "" {
		return fmt.Errorf(domain.ErrVercelProjectMissing)
	}

	token := setting.Token()
	if token == "" {
		return fmt.Errorf(domain.ErrVercelTokenMissing)
	}

	logger(fmt.Sprintf("Vercel 项目名称: %s", projectName))

	// 1. 扫描文件并计算 SHA1
	logger("正在扫描文件并计算哈希值...")
	fileResults, err := p.scanAndHashFiles(outputDir)
	if err != nil {
		return fmt.Errorf("扫描文件失败: %w", err)
	}

	if len(fileResults) == 0 {
		logger("没有发现可供部署的文件。")
		return nil
	}

	logger(fmt.Sprintf("文件扫描完成，共 %d 个文件。", len(fileResults)))

	for round := 1; round <= maxRounds; round++ {
		logger(fmt.Sprintf("正在创建部署（第 %d 轮）...", round))
		resp, err := p.createDeployment(ctx, projectName, fileResults, token)
		if err != nil {
			return fmt.Errorf("创建部署失败: %w", err)
		}

		if len(resp.Missing) == 0 {
			if round == 1 {
				logger("所有文件已在 Vercel 缓存中，无需上传。")
			}
			logger("✅ Vercel 部署成功！")
			return nil
		}

		if round == maxRounds {
			return fmt.Errorf(
				"Vercel 在 %d 轮后仍有 %d 个文件被报缺失；请检查网络后重试（通常为上传中断 / digest 校验失败）",
				maxRounds, len(resp.Missing),
			)
		}

		logger(fmt.Sprintf("第 %d 轮需要上传 %d / %d 个文件...", round, len(resp.Missing), len(fileResults)))

		missingSet := make(map[string]bool, len(resp.Missing))
		for _, sha := range resp.Missing {
			missingSet[sha] = true
		}
		var filesToUpload []VercelFileResult
		for _, f := range fileResults {
			if missingSet[f.Sha] {
				filesToUpload = append(filesToUpload, f)
			}
		}

		if err := p.uploadFiles(ctx, outputDir, filesToUpload, token, logger); err != nil {
			return fmt.Errorf("上传文件失败: %w", err)
		}
	}
	// 循环保证此处不可达
	return nil
}

// scanAndHashFiles 遍历目录，计算每个文件的 SHA1 值及文件大小
func (p *VercelProvider) scanAndHashFiles(outputDir string) ([]VercelFileResult, error) {
	var results []VercelFileResult

	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == ".github" {
				return filepath.SkipDir
			}
			return nil
		}

		name := info.Name()
		if name == ".DS_Store" || name == ".gitignore" {
			return nil
		}

		relPath, err := filepath.Rel(outputDir, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		hash := sha1.New()
		if _, err := io.Copy(hash, file); err != nil {
			return err
		}
		shaStr := hex.EncodeToString(hash.Sum(nil))

		results = append(results, VercelFileResult{
			File: relPath,
			Sha:  shaStr,
			Size: info.Size(),
		})

		return nil
	})

	return results, err
}

// uploadFiles 并发上传文件到 Vercel
func (p *VercelProvider) uploadFiles(ctx context.Context, outputDir string, files []VercelFileResult, token string, logger LogFunc) error {
	var eg errgroup.Group
	eg.SetLimit(10)

	for _, result := range files {
		res := result
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			filePath := filepath.Join(outputDir, filepath.FromSlash(res.File))
			if err := p.uploadSingleFile(ctx, filePath, res.Sha, res.Size, token); err != nil {
				return fmt.Errorf("文件 %s 上传失败: %w", res.File, err)
			}

			logger(fmt.Sprintf("已上传: %s", res.File))
			return nil
		})
	}

	return eg.Wait()
}

// uploadSingleFile 上传单个文件到 Vercel
func (p *VercelProvider) uploadSingleFile(ctx context.Context, filePath, sha string, size int64, token string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.vercel.com/v2/files", file)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("x-vercel-digest", sha)
	req.ContentLength = size

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// vercelDeployResp 聚合 createDeployment 的返回：可能是"部署成功"、"仍有文件
// 需上传"或"真正的错误"。无论响应码是什么都尝试解析 missing —— Vercel 的
// v13 在 2xx 和非 2xx 响应里都可能出现 missing 字段（issue #48）。
type vercelDeployResp struct {
	ID      string
	Missing []string
}

// parseVercelDeployBody 根据 HTTP 状态码与 body 做分类：
//   - 2xx：部署请求被受理，missing 可能为 nil 或非空（Vercel 也会在 2xx 里告知缺文件）
//   - 非 2xx + error.missing 非空：服务端需要先补齐文件，当作"半成功"继续循环
//   - 其它非 2xx：返回真正的错误（优先 error.message）
//
// 抽出来便于测试各类响应体形态。
func parseVercelDeployBody(statusCode int, bodyBytes []byte) (*vercelDeployResp, error) {
	var parsed struct {
		ID      string   `json:"id"`
		Missing []string `json:"missing"`
		Error   *struct {
			Code    string   `json:"code"`
			Message string   `json:"message"`
			Missing []string `json:"missing"`
		} `json:"error"`
	}
	_ = json.Unmarshal(bodyBytes, &parsed)

	if statusCode >= 200 && statusCode < 300 {
		return &vercelDeployResp{ID: parsed.ID, Missing: parsed.Missing}, nil
	}
	if parsed.Error != nil && len(parsed.Error.Missing) > 0 {
		return &vercelDeployResp{Missing: parsed.Error.Missing}, nil
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return nil, fmt.Errorf("HTTP %d (%s): %s", statusCode, parsed.Error.Code, parsed.Error.Message)
	}
	return nil, fmt.Errorf("HTTP %d: %s", statusCode, string(bodyBytes))
}

// createDeployment 调用 Vercel v13 部署接口。
func (p *VercelProvider) createDeployment(ctx context.Context, projectName string, files []VercelFileResult, token string) (*vercelDeployResp, error) {
	payload := map[string]interface{}{
		"name":   projectName,
		"files":  files,
		"target": "production",
		"projectSettings": map[string]interface{}{
			"framework": nil,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.vercel.com/v13/deployments", bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	return parseVercelDeployBody(resp.StatusCode, bodyBytes)
}

// AddCustomDomain 通过 Vercel API 为项目绑定自定义域名
func (p *VercelProvider) AddCustomDomain(ctx context.Context, projectName, domainName, token string) error {
	payload, _ := json.Marshal(map[string]string{"name": domainName})

	u := fmt.Sprintf("https://api.vercel.com/v10/projects/%s/domains", projectName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 200: 已存在, 201: 新增成功
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return nil
	}

	// 409: 域名已绑定到该项目（也算成功）
	if resp.StatusCode == http.StatusConflict {
		return nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
}

// RemoveCustomDomain 通过 Vercel API 解绑项目的自定义域名
func (p *VercelProvider) RemoveCustomDomain(ctx context.Context, projectName, domainName, token string) error {
	u := fmt.Sprintf("https://api.vercel.com/v9/projects/%s/domains/%s", projectName, domainName)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 200: 删除成功, 404: 域名不存在（也算成功）
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
}
