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

// netlifyAPIBase 指向 Netlify API 根；测试可替换为本地 httptest 地址。
var netlifyAPIBase = "https://api.netlify.com/api/v1"

// NetlifyProvider 实现了 Netlify API 部署策略
type NetlifyProvider struct {
	client *http.Client
}

// NewNetlifyProvider 创建 NetlifyProvider，proxyURL 为空则不使用代理
func NewNetlifyProvider(proxyURL string) *NetlifyProvider {
	return &NetlifyProvider{client: newNetlifyHTTPClient(proxyURL)}
}

func newNetlifyHTTPClient(proxyURL string) *http.Client {
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
		Timeout:   120 * time.Second,
		Transport: transport,
	}
}

// Deploy 实现 Provider 接口
// 流程：扫描文件 → 创建部署(获取 required 列表) → 上传缺失文件 → 自动生效
func (p *NetlifyProvider) Deploy(ctx context.Context, outputDir string, setting *domain.Setting, logger LogFunc) error {
	logger("🚀 开始准备 Netlify 部署...")

	siteId := setting.NetlifySiteId()
	if siteId == "" {
		return fmt.Errorf(domain.ErrNetlifySiteIdMissing)
	}

	token := setting.NetlifyAccessToken()
	if token == "" {
		return fmt.Errorf(domain.ErrNetlifyTokenMissing)
	}

	logger(fmt.Sprintf("Netlify Site ID: %s", siteId))

	// 1. 扫描文件并计算 SHA1
	logger("正在扫描文件并计算哈希值...")
	fileMap, err := p.scanAndHashFiles(outputDir)
	if err != nil {
		return fmt.Errorf("扫描文件失败: %w", err)
	}

	if len(fileMap) == 0 {
		logger("没有发现可供部署的文件。")
		return nil
	}

	logger(fmt.Sprintf("文件扫描完成，共 %d 个文件。", len(fileMap)))

	// 2. 创建部署，Netlify 返回需要上传的文件 SHA 列表
	logger("正在创建部署...")
	deployId, required, err := p.createDeploy(ctx, siteId, fileMap, token)
	if err != nil {
		return fmt.Errorf("创建部署失败: %w", err)
	}

	// 3. 上传缺失文件
	if len(required) > 0 {
		logger(fmt.Sprintf("需要上传 %d / %d 个文件...", len(required), len(fileMap)))

		// 构建 sha → path 的反向映射
		shaToPath := make(map[string]string, len(fileMap))
		for path, sha := range fileMap {
			shaToPath[sha] = path
		}

		if err := p.uploadFiles(ctx, outputDir, deployId, required, shaToPath, token, logger); err != nil {
			return fmt.Errorf("上传文件失败: %w", err)
		}
	} else {
		logger("所有文件已在 Netlify 缓存中，无需上传。")
	}

	// 4. 轮询部署状态直到 ready / error / 超时（issue #50）。
	// "上传完成" != "部署成功"：Netlify 仍要做校验 / CDN 分发 / build plugin 等，
	// 可能最终变为 error。没有这一步前端会误报成功，站点实际仍是旧版本。
	logger("文件上传完成，等待 Netlify 处理...")
	status, err := p.pollDeployStatus(ctx, deployId, token, logger)
	if err != nil {
		return err
	}
	if status.DeployURL != "" {
		logger(fmt.Sprintf("✅ Netlify 部署成功：%s", status.DeployURL))
	} else {
		logger("✅ Netlify 部署成功！")
	}
	return nil
}

// netlifyDeployStatus 是 /api/v1/deploys/{id} 的状态响应，字段按需取。
type netlifyDeployStatus struct {
	ID           string `json:"id"`
	State        string `json:"state"`         // uploading / processing / ready / error
	DeployURL    string `json:"deploy_url"`
	ErrorMessage string `json:"error_message"`
	Title        string `json:"title"`
}

// pollDeployStatus 轮询部署直到终态（ready / error）或超时。
// 轮询间隔从 2s 起指数退避到 10s 上限，总超时 5 分钟（不超过外部 ctx 的剩余时间）。
func (p *NetlifyProvider) pollDeployStatus(ctx context.Context, deployID, token string, logger LogFunc) (*netlifyDeployStatus, error) {
	const (
		initialInterval = 2 * time.Second
		maxInterval     = 10 * time.Second
		totalBudget     = 5 * time.Minute
	)

	deadline := time.Now().Add(totalBudget)
	interval := initialInterval
	apiURL := fmt.Sprintf("%s/deploys/%s", netlifyAPIBase, deployID)

	var lastState string
	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("等待 Netlify 处理超时（最后状态：%s）。部署已提交，可到 Netlify 后台查看最终结果", lastState)
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, httpErr := p.client.Do(req)
		if httpErr != nil {
			// 瞬时 / 网络错误：继续轮询，不中止
			logger(fmt.Sprintf("查询状态失败（将重试）：%v", httpErr))
		} else {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var st netlifyDeployStatus
				if err := json.Unmarshal(bodyBytes, &st); err == nil {
					if st.State != lastState {
						logger(fmt.Sprintf("Netlify 状态：%s", st.State))
						lastState = st.State
					}
					switch st.State {
					case "ready":
						return &st, nil
					case "error":
						msg := st.ErrorMessage
						if msg == "" {
							msg = st.Title
						}
						if msg == "" {
							msg = "未知错误"
						}
						return &st, fmt.Errorf("Netlify 处理失败：%s", msg)
					}
				}
			} else if resp.StatusCode == http.StatusUnauthorized {
				return nil, fmt.Errorf("Netlify Token 无效，轮询已中止")
			}
		}

		// 等待下一轮 / 响应 ctx 取消
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
		if interval < maxInterval {
			interval *= 2
			if interval > maxInterval {
				interval = maxInterval
			}
		}
	}
}

// scanAndHashFiles 遍历目录，返回 map["/relPath"] = "sha1hex"
func (p *NetlifyProvider) scanAndHashFiles(outputDir string) (map[string]string, error) {
	fileMap := make(map[string]string)

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
		relPath = "/" + filepath.ToSlash(relPath)

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		hash := sha1.New()
		if _, err := io.Copy(hash, file); err != nil {
			return err
		}

		fileMap[relPath] = hex.EncodeToString(hash.Sum(nil))
		return nil
	})

	return fileMap, err
}

// netlifyDeployResponse Netlify 创建部署的响应
type netlifyDeployResponse struct {
	ID       string   `json:"id"`
	Required []string `json:"required"`
}

// createDeploy 调用 Netlify 部署接口，返回部署 ID 和需要上传的文件 SHA 列表
func (p *NetlifyProvider) createDeploy(ctx context.Context, siteId string, files map[string]string, token string) (string, []string, error) {
	payload := map[string]any{
		"files": files,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", nil, err
	}

	apiURL := fmt.Sprintf("%s/sites/%s/deploys", netlifyAPIBase, siteId)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return "", nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result netlifyDeployResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return result.ID, result.Required, nil
}

// uploadFiles 并发上传缺失文件到 Netlify
func (p *NetlifyProvider) uploadFiles(ctx context.Context, outputDir, deployId string, required []string, shaToPath map[string]string, token string, logger LogFunc) error {
	var eg errgroup.Group
	eg.SetLimit(10)

	for _, sha := range required {
		reqSha := sha
		remotePath, ok := shaToPath[reqSha]
		if !ok {
			continue
		}

		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			localPath := filepath.Join(outputDir, filepath.FromSlash(strings.TrimPrefix(remotePath, "/")))
			if err := p.uploadSingleFile(ctx, deployId, localPath, remotePath, token); err != nil {
				return fmt.Errorf("文件 %s 上传失败: %w", remotePath, err)
			}

			logger(fmt.Sprintf("已上传: %s", remotePath))
			return nil
		})
	}

	return eg.Wait()
}

// uploadSingleFile 上传单个文件到 Netlify
func (p *NetlifyProvider) uploadSingleFile(ctx context.Context, deployId, localPath, remotePath, token string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	apiURL := fmt.Sprintf("%s/deploys/%s/files%s", netlifyAPIBase, deployId, remotePath)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, apiURL, file)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
