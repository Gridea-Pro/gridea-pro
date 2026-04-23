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

// Deploy 实现了 Provider 接口
// 流程：扫描文件 → 创建部署(获取 missing 列表) → 只上传缺失文件 → 完成
func (p *VercelProvider) Deploy(ctx context.Context, outputDir string, setting *domain.Setting, logger LogFunc) error {
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

	// 2. 创建部署，Vercel 会返回需要上传的文件列表（missing）
	logger("正在创建部署...")
	missing, err := p.createDeployment(ctx, projectName, fileResults, token)
	if err != nil {
		return fmt.Errorf("创建部署失败: %w", err)
	}

	// 3. 只上传 missing 的文件
	if len(missing) > 0 {
		logger(fmt.Sprintf("需要上传 %d / %d 个文件...", len(missing), len(fileResults)))

		// 构建 sha → file 的映射，快速查找需要上传的文件
		missingSet := make(map[string]bool, len(missing))
		for _, sha := range missing {
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

		// 4. 重新创建部署（文件已上传完毕）
		logger("文件上传完成，正在触发最终部署...")
		if _, err := p.createDeployment(ctx, projectName, fileResults, token); err != nil {
			return fmt.Errorf("触发最终部署失败: %w", err)
		}
	} else {
		logger("所有文件已在 Vercel 缓存中，无需上传。")
	}

	logger("✅ Vercel 部署成功！")
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

// uploadSingleFile 上传单个文件到 Vercel。
// 带有 x-vercel-digest 的 POST 是内容寻址的幂等写入（见 #46），5xx/429/网络
// 错误可安全重试，因此这里走 DoHTTPWithRetry 而非直接 client.Do。
func (p *VercelProvider) uploadSingleFile(ctx context.Context, filePath, sha string, size int64, token string) error {
	buildReq := func() (*http.Request, error) {
		// 每次重试重新 Open 文件；body 在失败后已被 http.Transport 消费，不能复用
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.vercel.com/v2/files", file)
		if err != nil {
			_ = file.Close()
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("x-vercel-digest", sha)
		req.ContentLength = size
		return req, nil
	}

	resp, err := DoHTTPWithRetry(ctx, p.client, buildReq, HTTPRetryPolicy{MaxAttempts: 3}, nil)
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

// createDeployment 调用 Vercel v13 部署接口
// 返回 missing 文件的 SHA 列表（如果有文件需要先上传）
func (p *VercelProvider) createDeployment(ctx context.Context, projectName string, files []VercelFileResult, token string) ([]string, error) {
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

	// Vercel 返回 missing 文件列表时状态码可能不同：
	// - 200/201: 部署创建成功（无 missing 或所有文件已存在）
	// - 其他: 可能包含 error 信息
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		// 尝试解析 missing 列表
		var errResp struct {
			Error struct {
				Code    string   `json:"code"`
				Missing []string `json:"missing"`
			} `json:"error"`
		}
		if json.Unmarshal(bodyBytes, &errResp) == nil && len(errResp.Error.Missing) > 0 {
			return errResp.Error.Missing, nil
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil, nil
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
