package facade

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gridea-pro/backend/internal/utils"
	"gridea-pro/backend/internal/version"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// UpdateFacade 处理程序更新检查、下载与应用
type UpdateFacade struct {
	releasesURL string
	httpClient  *http.Client

	mu              sync.Mutex
	downloadCancel  context.CancelFunc
	downloadingFile *os.File // 用于取消时清理
	readyPath       string   // 下载完成后的本地路径
	readyAssetName  string   // asset 名（macOS 判定 .zip / .dmg 所需）
	mockReady       bool     // 是否是 StartMockDownload 产生的假数据
}

// NewUpdateFacade 创建 UpdateFacade
func NewUpdateFacade() *UpdateFacade {
	return &UpdateFacade{
		releasesURL: "https://api.github.com/repos/Gridea-Pro/gridea-pro/releases/latest",
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// UpdateInfo 更新检查结果
type UpdateInfo struct {
	HasUpdate      bool   `json:"hasUpdate"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	PublishedAt    string `json:"publishedAt"`
	HtmlURL        string `json:"htmlUrl"`
	Body           string `json:"body"`
	BodyHTML       string `json:"bodyHtml"`
	// HasAsset 表示当前平台有匹配的下载资源，前端据此决定「立即更新」按钮是否可用
	HasAsset  bool   `json:"hasAsset"`
	AssetName string `json:"assetName"`
	AssetSize int64  `json:"assetSize"`
}

type githubAsset struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	DownloadURL string `json:"browser_download_url"`
}

type githubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	HtmlURL     string        `json:"html_url"`
	PublishedAt string        `json:"published_at"`
	Body        string        `json:"body"`
	Draft       bool          `json:"draft"`
	Prerelease  bool          `json:"prerelease"`
	Assets      []githubAsset `json:"assets"`
}

// CheckUpdate 请求 GitHub Releases 接口，返回版本对比结果
func (f *UpdateFacade) CheckUpdate() (*UpdateInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.releasesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Gridea-Pro/"+version.Version)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub Releases 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub Releases 返回 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("解析 Releases 响应失败: %w", err)
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	info := &UpdateInfo{
		CurrentVersion: version.Version,
		LatestVersion:  latest,
		PublishedAt:    rel.PublishedAt,
		HtmlURL:        rel.HtmlURL,
		Body:           rel.Body,
		BodyHTML:       utils.ToHTMLUnsafe(rel.Body),
		HasUpdate:      !rel.Draft && !rel.Prerelease && compareSemver(latest, version.Version) > 0,
	}
	if asset := pickAsset(rel.Assets, runtime.GOOS, runtime.GOARCH); asset != nil {
		info.HasAsset = true
		info.AssetName = asset.Name
		info.AssetSize = asset.Size

		f.mu.Lock()
		f.readyAssetName = "" // 新一轮检查重置下载态
		f.readyPath = ""
		f.mu.Unlock()
	}
	return info, nil
}

// MockUpdate 返回一份模拟数据，便于调试 UI（不访问网络）
func (f *UpdateFacade) MockUpdate() *UpdateInfo {
	body := "### ✨ 新功能\n\n- 全新更新检查弹窗\n- 支持从 GitHub Releases 拉取最新版本\n\n### 🐞 修复\n\n- 若干细节优化"
	return &UpdateInfo{
		HasUpdate:      true,
		CurrentVersion: version.Version,
		LatestVersion:  "9.9.9",
		PublishedAt:    time.Now().Format(time.RFC3339),
		HtmlURL:        "https://github.com/Gridea-Pro/gridea-pro/releases",
		Body:           body,
		BodyHTML:       utils.ToHTMLUnsafe(body),
		HasAsset:       true,
		AssetName:      fmt.Sprintf("Gridea-Pro-9.9.9-%s-%s.mock", runtime.GOOS, runtime.GOARCH),
		AssetSize:      42 * 1024 * 1024,
	}
}

// StartDownload 启动真实下载，全程通过 update:progress 事件推送进度
// 下载完成后发送 update:ready；失败发送 update:error。
func (f *UpdateFacade) StartDownload() error {
	f.mu.Lock()
	if f.downloadCancel != nil {
		f.mu.Unlock()
		return errors.New("已经有下载任务在运行")
	}
	ctx, cancel := context.WithCancel(context.Background())
	f.downloadCancel = cancel
	f.mu.Unlock()

	f.mu.Lock()
	f.mockReady = false
	f.mu.Unlock()

	// 重新拉一次 Release 信息，避免依赖前端缓存（也方便重试）
	go func() {
		defer f.clearDownloadState()

		asset, err := f.fetchAssetForCurrentPlatform(ctx)
		if err != nil {
			f.emitError(err)
			return
		}
		f.doDownload(ctx, asset.DownloadURL, asset.Name, asset.Size)
	}()
	return nil
}

// StartMockDownload SIMULATE 模式下用的假进度流（不触发真实 HTTP）
func (f *UpdateFacade) StartMockDownload() error {
	f.mu.Lock()
	if f.downloadCancel != nil {
		f.mu.Unlock()
		return errors.New("已经有下载任务在运行")
	}
	ctx, cancel := context.WithCancel(context.Background())
	f.downloadCancel = cancel
	f.mu.Unlock()

	go func() {
		defer f.clearDownloadState()

		const total int64 = 42 * 1024 * 1024 // 42MB
		const steps = 50
		step := total / steps
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		var received int64
		for i := 1; i <= steps; i++ {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				received += step
				if received > total {
					received = total
				}
				f.emitProgress(received, total)
			}
		}
		// 模拟一个本地临时文件以便 ApplyUpdate 的流程能跑通（不会真的去替换）
		tmp, err := os.CreateTemp("", "gridea-pro-mock-*.bin")
		if err == nil {
			_ = tmp.Close()
			f.mu.Lock()
			f.readyPath = tmp.Name()
			f.readyAssetName = fmt.Sprintf("Gridea-Pro-mock-%s-%s.bin", runtime.GOOS, runtime.GOARCH)
			f.mockReady = true
			f.mu.Unlock()
			f.emitReady(tmp.Name())
		} else {
			f.emitError(err)
		}
	}()
	return nil
}

// CancelDownload 取消正在进行的下载
func (f *UpdateFacade) CancelDownload() {
	f.mu.Lock()
	cancel := f.downloadCancel
	file := f.downloadingFile
	f.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if file != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}
}

// ApplyUpdate 应用已下载的更新并重启应用
func (f *UpdateFacade) ApplyUpdate() error {
	f.mu.Lock()
	path := f.readyPath
	name := f.readyAssetName
	mock := f.mockReady
	f.mu.Unlock()

	if path == "" {
		return errors.New("尚未完成下载，无法安装")
	}
	// mock 下载：不触发真实替换，也不重启应用，直接返回成功
	// 前端在 SIMULATE 模式下会自行处理 UI（Toast + 关闭弹窗）
	if mock {
		_ = os.Remove(path)
		f.mu.Lock()
		f.readyPath = ""
		f.readyAssetName = ""
		f.mockReady = false
		f.mu.Unlock()
		return nil
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("下载文件丢失: %w", err)
	}

	// 由平台专属实现完成替换 + 重启
	if err := installAndRelaunch(path, name); err != nil {
		return err
	}
	// installAndRelaunch 通常在触发重启前返回；再通知 Wails 退出
	if WailsContext != nil {
		go func() {
			// 留一点时间让前端收到消息
			time.Sleep(300 * time.Millisecond)
			wailsRuntime.Quit(WailsContext)
		}()
	}
	return nil
}

// ─── 内部辅助 ─────────────────────────────────────────

func (f *UpdateFacade) clearDownloadState() {
	f.mu.Lock()
	f.downloadCancel = nil
	f.downloadingFile = nil
	f.mu.Unlock()
}

func (f *UpdateFacade) fetchAssetForCurrentPlatform(ctx context.Context) (*githubAsset, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.releasesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Gridea-Pro/"+version.Version)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 Releases 失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Releases 返回 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("解析 Releases 失败: %w", err)
	}
	asset := pickAsset(rel.Assets, runtime.GOOS, runtime.GOARCH)
	if asset == nil {
		return nil, fmt.Errorf("没有匹配当前平台 (%s/%s) 的下载资源", runtime.GOOS, runtime.GOARCH)
	}
	return asset, nil
}

func (f *UpdateFacade) doDownload(ctx context.Context, url, assetName string, expectedSize int64) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		f.emitError(err)
		return
	}
	req.Header.Set("User-Agent", "Gridea-Pro/"+version.Version)

	// 下载客户端用较长超时（GitHub LFS 重定向也走这儿）
	dlClient := &http.Client{Timeout: 30 * time.Minute}
	resp, err := dlClient.Do(req)
	if err != nil {
		f.emitError(fmt.Errorf("下载失败: %w", err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		f.emitError(fmt.Errorf("下载返回 %d", resp.StatusCode))
		return
	}

	total := resp.ContentLength
	if total <= 0 {
		total = expectedSize
	}

	tmp, err := os.CreateTemp("", "gridea-pro-update-*-"+sanitizeName(assetName))
	if err != nil {
		f.emitError(fmt.Errorf("创建临时文件失败: %w", err))
		return
	}
	f.mu.Lock()
	f.downloadingFile = tmp
	f.mu.Unlock()

	// 边读边写 + 每 200ms 发一次进度
	buf := make([]byte, 64*1024)
	var received int64
	nextEmit := time.Now()
	for {
		select {
		case <-ctx.Done():
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
			return
		default:
		}

		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := tmp.Write(buf[:n]); werr != nil {
				_ = tmp.Close()
				_ = os.Remove(tmp.Name())
				f.emitError(fmt.Errorf("写入失败: %w", werr))
				return
			}
			received += int64(n)
			if time.Now().After(nextEmit) {
				f.emitProgress(received, total)
				nextEmit = time.Now().Add(200 * time.Millisecond)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
			f.emitError(fmt.Errorf("读取失败: %w", rerr))
			return
		}
	}
	// 最后再推一次 100%
	f.emitProgress(received, received)

	if err := tmp.Close(); err != nil {
		f.emitError(fmt.Errorf("关闭文件失败: %w", err))
		return
	}

	f.mu.Lock()
	f.readyPath = tmp.Name()
	f.readyAssetName = assetName
	f.mu.Unlock()

	f.emitReady(tmp.Name())
}

func (f *UpdateFacade) emitProgress(received, total int64) {
	if WailsContext == nil {
		return
	}
	percent := float64(0)
	if total > 0 {
		percent = float64(received) * 100.0 / float64(total)
	}
	wailsRuntime.EventsEmit(WailsContext, "update:progress", map[string]any{
		"received": received,
		"total":    total,
		"percent":  percent,
	})
}

func (f *UpdateFacade) emitReady(path string) {
	if WailsContext == nil {
		return
	}
	wailsRuntime.EventsEmit(WailsContext, "update:ready", map[string]any{
		"filePath": path,
	})
}

func (f *UpdateFacade) emitError(err error) {
	if WailsContext == nil {
		return
	}
	wailsRuntime.EventsEmit(WailsContext, "update:error", map[string]any{
		"message": err.Error(),
	})
}

// pickAsset 按当前 GOOS/GOARCH 找到匹配的 asset。
// 命名约定宽松：只要同时包含平台和架构关键字即可命中。
func pickAsset(assets []githubAsset, goos, goarch string) *githubAsset {
	osKeys := map[string][]string{
		"darwin":  {"darwin", "mac", "macos", "osx"},
		"windows": {"windows", "win"},
		"linux":   {"linux"},
	}
	archKeys := map[string][]string{
		"amd64": {"amd64", "x86_64", "x64", "intel"},
		"arm64": {"arm64", "aarch64", "apple"},
	}
	// 扩展名优先级（同平台匹配多个时取优先级高的）
	// macOS 自更新优先 .zip（installer 可直接解压 .app），.dmg 留给首次下载
	// Windows 自更新优先便携 .exe（selfupdate 可直接替换），.msi 安装器吃不下
	extPriority := map[string]int{
		".zip": 4, ".dmg": 3, // macOS
		".exe": 4, ".msi": 3, // windows
		".AppImage": 4, ".tar.gz": 3, ".tar.xz": 2, // linux
	}

	var best *githubAsset
	bestPri := -1
	for i := range assets {
		a := &assets[i]
		name := strings.ToLower(a.Name)
		if !containsAny(name, osKeys[goos]) {
			continue
		}
		if keys, ok := archKeys[goarch]; ok && !containsAny(name, keys) {
			// 没带架构关键字的通用包也允许，但优先级降一档
		}
		pri := 0
		for ext, p := range extPriority {
			if strings.HasSuffix(name, strings.ToLower(ext)) {
				pri = p
				break
			}
		}
		// 安装器类资源自更新无法直接替换二进制，降权避免被 selfupdate 选中
		if strings.Contains(name, "setup") || strings.Contains(name, "installer") {
			pri -= 2
		}
		if pri > bestPri {
			bestPri = pri
			best = a
		}
	}
	return best
}

func containsAny(s string, keys []string) bool {
	for _, k := range keys {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

func sanitizeName(name string) string {
	name = filepath.Base(name)
	// 防止 Windows/Unix 特殊字符影响 CreateTemp
	repl := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "*", "-", "?", "-")
	return repl.Replace(name)
}

// compareSemver 比较两个语义化版本号，a > b 返回 1，a < b 返回 -1，相等返回 0。
func compareSemver(a, b string) int {
	as := splitVersion(a)
	bs := splitVersion(b)
	n := max(len(as), len(bs))
	for i := range n {
		av := 0
		bv := 0
		if i < len(as) {
			av = as[i]
		}
		if i < len(bs) {
			bv = bs[i]
		}
		if av > bv {
			return 1
		}
		if av < bv {
			return -1
		}
	}
	return 0
}

func splitVersion(v string) []int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			out = append(out, 0)
			continue
		}
		out = append(out, n)
	}
	return out
}
