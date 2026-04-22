package facade

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"gridea-pro/backend/internal/utils"
	"gridea-pro/backend/internal/version"
	"io"
	"log/slog"
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
}

// NewUpdateFacade 创建 UpdateFacade
func NewUpdateFacade() *UpdateFacade {
	return &UpdateFacade{
		releasesURL: "https://api.github.com/repos/Gridea-Pro/gridea-pro/releases/latest",
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// trustedDownloadPrefix 自更新下载 URL 必须的前缀：本仓库 releases 资源。
// 即便 Release 的 browser_download_url 字段被篡改指向第三方域，也会被这里拦掉。
// 不要改为可配置 —— 这条硬编码是自更新安全链的最后一道关。
const trustedDownloadPrefix = "https://github.com/Gridea-Pro/gridea-pro/releases/download/"

// isTrustedDownloadURL 校验 URL 前缀是否在白名单内。
func isTrustedDownloadURL(url string) bool {
	return strings.HasPrefix(url, trustedDownloadPrefix)
}

// trustedRedirectSuffixes 自更新下载允许落地的域名根：
// GitHub releases 首跳在 github.com，资源本体 302 到 *.githubusercontent.com。
// 用后缀匹配 + 边界校验，避免 github.com.evil.com 这类后缀伪装；
// 同时兼容 GitHub 将来新增的 CDN 子域，不用穷举。
var trustedRedirectSuffixes = []string{
	"github.com",
	"githubusercontent.com",
}

// hasTrustedRedirectHost 判定 host 严格等于白名单后缀，或为 "." + 后缀 结尾。
func hasTrustedRedirectHost(host string) bool {
	host = strings.ToLower(host)
	if i := strings.IndexByte(host, ':'); i >= 0 {
		host = host[:i]
	}
	for _, suffix := range trustedRedirectSuffixes {
		if host == suffix || strings.HasSuffix(host, "."+suffix) {
			return true
		}
	}
	return false
}

// trustedRedirectChecker 作为 http.Client.CheckRedirect 使用，
// 限制 3xx 跳转只能落到 github 体系内的 HTTPS 地址，且跳转总数 < 10。
func trustedRedirectChecker(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return fmt.Errorf("重定向次数过多（%d 次）", len(via))
	}
	if req.URL.Scheme != "https" {
		return fmt.Errorf("拒绝非 HTTPS 重定向: %s", req.URL.String())
	}
	if !hasTrustedRedirectHost(req.URL.Host) {
		return fmt.Errorf("拒绝重定向到非预期域名: %s", req.URL.Host)
	}
	return nil
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

// StartDownload 启动真实下载，全程通过 update:progress 事件推送进度
// 下载完成后发送 update:ready；失败发送 update:error。
func (f *UpdateFacade) StartDownload() error {
	f.mu.Lock()
	if f.downloadCancel != nil {
		f.mu.Unlock()
		return errors.New("已经有下载任务在运行")
	}
	// 清理上一轮可能遗留的 ready 状态：避免"上一次下载成功但用户没点立即更新
	// 之后这一次下载失败"的组合下，ApplyUpdate 误用上一次的旧包。
	// 同时尝试删除磁盘上的旧临时文件（容忍失败，文件可能被系统清理）。
	if f.readyPath != "" {
		_ = os.Remove(f.readyPath)
	}
	f.readyPath = ""
	f.readyAssetName = ""
	ctx, cancel := context.WithCancel(context.Background())
	f.downloadCancel = cancel
	f.mu.Unlock()

	// 重新拉一次 Release 信息，避免依赖前端缓存（也方便重试）
	go func() {
		defer f.clearDownloadState()

		asset, sums, err := f.fetchAssetForCurrentPlatform(ctx)
		if err != nil {
			f.emitError(err)
			return
		}
		f.doDownload(ctx, asset.DownloadURL, asset.Name, asset.Size, sums)
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
	f.mu.Unlock()

	if path == "" {
		return errors.New("尚未完成下载，无法安装")
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

// fetchAssetForCurrentPlatform 返回当前平台的下载 asset，以及同 Release 中
// 名为 SHA256SUMS 的校验文件（若存在）。校验文件缺失时 sums 返回 nil，
// 调用方应 warn 但允许下载继续 —— 这是为兼容未产出 SHA256SUMS 的历史 Release。
func (f *UpdateFacade) fetchAssetForCurrentPlatform(ctx context.Context) (asset *githubAsset, sums *githubAsset, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.releasesURL, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Gridea-Pro/"+version.Version)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("请求 Releases 失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("Releases 返回 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, nil, fmt.Errorf("解析 Releases 失败: %w", err)
	}
	asset = pickAsset(rel.Assets, runtime.GOOS, runtime.GOARCH)
	if asset == nil {
		return nil, nil, fmt.Errorf("没有匹配当前平台 (%s/%s) 的下载资源", runtime.GOOS, runtime.GOARCH)
	}
	sums = findSumsAsset(rel.Assets)
	return asset, sums, nil
}

// findSumsAsset 在 assets 里查找 SHA256SUMS 文件（约定命名，全大写匹配）。
// 未来如果改成 SHA256SUMS.sig / .asc 等格式，也可以在这里扩展识别规则。
func findSumsAsset(assets []githubAsset) *githubAsset {
	for i := range assets {
		if assets[i].Name == "SHA256SUMS" {
			return &assets[i]
		}
	}
	return nil
}

func (f *UpdateFacade) doDownload(ctx context.Context, url, assetName string, expectedSize int64, sumsAsset *githubAsset) {
	// 安全检查：即使 GitHub API 返回的 browser_download_url 被篡改（凭证泄漏 / Release 被接管 /
	// 上游代理中间人等场景），也必须走本仓库 releases 资源前缀；其他一律拒绝下载。
	if !isTrustedDownloadURL(url) {
		f.emitError(fmt.Errorf("拒绝下载：非预期的更新包 URL: %s", url))
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		f.emitError(err)
		return
	}
	req.Header.Set("User-Agent", "Gridea-Pro/"+version.Version)

	// 下载客户端用较长超时（GitHub LFS 重定向也走这儿）。
	// CheckRedirect 限制跳转只能落到 github.com / *.githubusercontent.com 体系内，
	// 避免入口 URL 过了白名单后被 302 重定向到第三方域名绕过防御。
	dlClient := &http.Client{
		Timeout:       30 * time.Minute,
		CheckRedirect: trustedRedirectChecker,
	}
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

	// SHA256 完整性校验：防中间人、防上游 Release 被动替换、防半下载损坏。
	// sumsAsset 为 nil 时仅告警（兼容未产出 SHA256SUMS 的历史 Release）；
	// 为非 nil 时必须通过，否则删掉临时文件并 emitError，不进入 readyPath 状态。
	if err := f.verifyDownloadChecksum(ctx, tmp.Name(), assetName, sumsAsset); err != nil {
		_ = os.Remove(tmp.Name())
		f.emitError(fmt.Errorf("完整性校验失败: %w", err))
		return
	}

	f.mu.Lock()
	f.readyPath = tmp.Name()
	f.readyAssetName = assetName
	f.mu.Unlock()

	f.emitReady(tmp.Name())
}

// verifyDownloadChecksum 拉取 SHA256SUMS、在其中查找 assetName 对应的哈希、
// 计算本地文件哈希并对比。sumsAsset 为 nil 时仅告警并放行（向后兼容），
// 非 nil 时任何一步失败都返回 error —— 调用方会丢弃下载文件。
func (f *UpdateFacade) verifyDownloadChecksum(ctx context.Context, localPath, assetName string, sumsAsset *githubAsset) error {
	if sumsAsset == nil {
		slog.Warn("本次 Release 无 SHA256SUMS 资源，跳过完整性校验（建议重新发布时补上）",
			"asset", assetName)
		return nil
	}

	expected, err := f.fetchExpectedChecksum(ctx, sumsAsset.DownloadURL, assetName)
	if err != nil {
		return err
	}
	if expected == "" {
		return fmt.Errorf("SHA256SUMS 中未找到 %s 的哈希", assetName)
	}

	actual, err := sha256File(localPath)
	if err != nil {
		return fmt.Errorf("计算本地哈希失败: %w", err)
	}
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("下载包哈希不匹配：期望 %s，实际 %s", expected, actual)
	}
	return nil
}

// fetchExpectedChecksum 下载 SHA256SUMS 并返回 assetName 对应的 hex 哈希。
// 用与二进制下载相同大小的合理上限（1 MiB）防止被超大文件拖垮。
func (f *UpdateFacade) fetchExpectedChecksum(ctx context.Context, sumsURL, assetName string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sumsURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Gridea-Pro/"+version.Version)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载 SHA256SUMS 失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("下载 SHA256SUMS 返回 %d", resp.StatusCode)
	}

	const maxSize = 1 << 20 // 1 MiB
	return parseSha256Sums(io.LimitReader(resp.Body, maxSize), assetName)
}

// parseSha256Sums 按 GNU sha256sum 格式解析（"<hex>  <filename>"），
// 返回目标文件名的哈希；未命中返回空串。
func parseSha256Sums(r io.Reader, target string) (string, error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 格式："<hex>  <filename>"（两个空格分隔）；部分实现会是单空格或 tab
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := strings.TrimPrefix(fields[1], "*") // sha256sum -b 会在文件名前加 '*'
		if name == target {
			return strings.ToLower(fields[0]), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", nil
}

// sha256File 计算文件内容的 SHA256 hex 哈希。
func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
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

// extRule 描述合法发布包后缀及其自更新优先级。
type extRule struct {
	ext string
	pri int
}

// binaryAssetRules 定义可作为自更新候选的 asset 后缀白名单。
// 按长度降序排列，避免 .tar.gz 被更短的 .gz 抢先匹配（预留扩展）。
// 优先级：.zip/.exe/.AppImage 这类可直接替换主程序的打包 > .dmg/.msi 安装器 > .deb/.rpm 包管理格式。
var binaryAssetRules = []extRule{
	{".AppImage", 4},
	{".tar.gz", 3},
	{".tar.xz", 2},
	{".zip", 4},
	{".exe", 4},
	{".dmg", 3},
	{".msi", 3},
	{".deb", 1},
	{".rpm", 1},
}

// matchAssetExt 返回 name 匹配的二进制后缀及其优先级；未命中返回空串与 0。
func matchAssetExt(name string) (string, int, bool) {
	lower := strings.ToLower(name)
	for _, r := range binaryAssetRules {
		if strings.HasSuffix(lower, strings.ToLower(r.ext)) {
			return r.ext, r.pri, true
		}
	}
	return "", 0, false
}

// pickAsset 按当前 GOOS/GOARCH 找到匹配的 asset。
//
// 匹配条件：
//  1. 文件名（lowercase）需包含当前平台关键字
//  2. 文件名后缀必须在二进制发布包白名单内（避免 changelog.md / notes.txt 等附件被误选）
//  3. 当指定平台有架构关键字集时，未带架构关键字的包允许匹配但权重降一档
//  4. setup/installer 命名降权，避免自更新流程去拉交互式安装器
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

	var best *githubAsset
	bestPri := -1
	for i := range assets {
		a := &assets[i]
		name := strings.ToLower(a.Name)
		if !containsAny(name, osKeys[goos]) {
			continue
		}
		_, pri, ok := matchAssetExt(name)
		if !ok {
			continue
		}
		// 没带架构关键字的通用包允许匹配，但权重降一档，优先选明确匹配架构的包
		if keys, ok := archKeys[goarch]; ok && !containsAny(name, keys) {
			pri--
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
