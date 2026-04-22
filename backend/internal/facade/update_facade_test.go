package facade

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ─── pickAsset 后缀白名单（#57 / PR #74） ────────────────────────────────────

func mkAssets(names ...string) []githubAsset {
	out := make([]githubAsset, len(names))
	for i, n := range names {
		out[i] = githubAsset{Name: n, DownloadURL: "https://example.com/" + n, Size: 1024}
	}
	return out
}

func TestPickAsset_BinaryWhitelist(t *testing.T) {
	tests := []struct {
		name    string
		assets  []githubAsset
		goos    string
		goarch  string
		wantHit string // 期望的 asset name；"" 表示 nil
	}{
		{
			name:    "macos_arm64_zip_wins",
			assets:  mkAssets("Gridea-Pro-1.0.0-darwin-arm64.zip", "Gridea-Pro-1.0.0-darwin-arm64.dmg"),
			goos:    "darwin",
			goarch:  "arm64",
			wantHit: "Gridea-Pro-1.0.0-darwin-arm64.zip",
		},
		{
			name:    "windows_amd64_exe_wins_over_msi",
			assets:  mkAssets("Gridea-Pro-1.0.0-windows-amd64.exe", "Gridea-Pro-1.0.0-windows-amd64.msi"),
			goos:    "windows",
			goarch:  "amd64",
			wantHit: "Gridea-Pro-1.0.0-windows-amd64.exe",
		},
		{
			name:    "linux_amd64_appimage_wins",
			assets:  mkAssets("Gridea-Pro-1.0.0-linux-amd64.AppImage", "Gridea-Pro-1.0.0-linux-amd64.tar.gz"),
			goos:    "linux",
			goarch:  "amd64",
			wantHit: "Gridea-Pro-1.0.0-linux-amd64.AppImage",
		},
		{
			// 核心修复：含平台关键字的非二进制附件（.md/.txt/.json）必须被忽略
			name: "markdown_with_macos_keyword_ignored",
			assets: mkAssets(
				"changelog-macos.md",
				"Gridea-Pro-1.0.0-darwin-arm64.zip",
			),
			goos:    "darwin",
			goarch:  "arm64",
			wantHit: "Gridea-Pro-1.0.0-darwin-arm64.zip",
		},
		{
			// 仅有非二进制附件时，pickAsset 应返回 nil 而非错选 .md
			name: "only_markdown_returns_nil",
			assets: mkAssets(
				"release-notes-macos.md",
				"install-guide-linux.txt",
				"build-manifest-windows.json",
			),
			goos:    "darwin",
			goarch:  "arm64",
			wantHit: "",
		},
		{
			// setup/installer 命名降权，便携 exe 胜出
			name: "portable_exe_beats_installer_exe",
			assets: mkAssets(
				"Gridea-Pro-1.0.0-windows-amd64-setup.exe",
				"Gridea-Pro-1.0.0-windows-amd64.exe",
			),
			goos:    "windows",
			goarch:  "amd64",
			wantHit: "Gridea-Pro-1.0.0-windows-amd64.exe",
		},
		{
			// 架构未指定：通用包允许命中但权重降一档，优先匹配明确架构的
			name: "arch_specific_beats_generic",
			assets: mkAssets(
				"Gridea-Pro-1.0.0-darwin.zip",       // 没带架构
				"Gridea-Pro-1.0.0-darwin-arm64.zip", // 明确 arm64
			),
			goos:    "darwin",
			goarch:  "arm64",
			wantHit: "Gridea-Pro-1.0.0-darwin-arm64.zip",
		},
		{
			// 没有当前平台的 asset 时返回 nil
			name:    "no_match_returns_nil",
			assets:  mkAssets("Gridea-Pro-1.0.0-linux-amd64.AppImage"),
			goos:    "darwin",
			goarch:  "arm64",
			wantHit: "",
		},
		{
			// deb/rpm 虽在白名单但优先级较低，zip 应胜出
			name: "zip_beats_deb",
			assets: mkAssets(
				"gridea-pro_1.0.0_linux_amd64.deb",
				"Gridea-Pro-1.0.0-linux-amd64.tar.gz",
			),
			goos:    "linux",
			goarch:  "amd64",
			wantHit: "Gridea-Pro-1.0.0-linux-amd64.tar.gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pickAsset(tt.assets, tt.goos, tt.goarch)
			if tt.wantHit == "" {
				if got != nil {
					t.Errorf("pickAsset returned %q, want nil", got.Name)
				}
				return
			}
			if got == nil {
				t.Fatalf("pickAsset returned nil, want %q", tt.wantHit)
			}
			if got.Name != tt.wantHit {
				t.Errorf("pickAsset returned %q, want %q", got.Name, tt.wantHit)
			}
		})
	}
}

func TestMatchAssetExt(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		priGT   int // priority must be > this
		wantHit bool
	}{
		{"app-1.0.0.AppImage", ".AppImage", 0, true},
		{"app-1.0.0.tar.gz", ".tar.gz", 0, true},
		{"app-1.0.0.tar.xz", ".tar.xz", 0, true},
		{"app-1.0.0-darwin-arm64.zip", ".zip", 0, true},
		{"app-1.0.0-darwin.dmg", ".dmg", 0, true},
		{"app-1.0.0-windows.exe", ".exe", 0, true},
		{"app-1.0.0-windows.msi", ".msi", 0, true},
		{"app.deb", ".deb", 0, true},
		{"app.rpm", ".rpm", 0, true},
		{"changelog.md", "", -1, false},
		{"notes.txt", "", -1, false},
		{"manifest.json", "", -1, false},
		{"release.yaml", "", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, pri, ok := matchAssetExt(tt.name)
			if ok != tt.wantHit {
				t.Errorf("matchAssetExt(%q) hit = %v, want %v", tt.name, ok, tt.wantHit)
			}
			if ok && ext != tt.want {
				t.Errorf("matchAssetExt(%q) ext = %q, want %q", tt.name, ext, tt.want)
			}
			if ok && pri <= tt.priGT {
				t.Errorf("matchAssetExt(%q) priority = %d, want > %d", tt.name, pri, tt.priGT)
			}
		})
	}
}

// ─── StartDownload readyPath 清理（#56 / PR #79） ────────────────────────────

// newTestFacadeWith404 返回一个 UpdateFacade，其 releasesURL 指向本地 404 服务，
// 用于模拟"新下载失败"的场景。
func newTestFacadeWith404(t *testing.T) (*UpdateFacade, func()) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "no release", http.StatusNotFound)
	}))
	f := &UpdateFacade{
		releasesURL: srv.URL,
		httpClient:  &http.Client{Timeout: 2 * time.Second},
	}
	return f, func() { srv.Close() }
}

// 关键修复：连续两次下载（第一次成功、第二次失败）后，readyPath 不应指向第一次的文件。
func TestStartDownload_ClearsPreviousReadyState(t *testing.T) {
	f, cleanup := newTestFacadeWith404(t)
	defer cleanup()

	// 模拟上一次下载成功后残留在 facade 上的状态
	stalePath := filepath.Join(t.TempDir(), "old-release.zip")
	if err := os.WriteFile(stalePath, []byte("old content"), 0o644); err != nil {
		t.Fatalf("seed stale file: %v", err)
	}
	f.mu.Lock()
	f.readyPath = stalePath
	f.readyAssetName = "old-release.zip"
	f.mu.Unlock()

	// 新一轮 StartDownload —— 这次因为 releasesURL 返回 404 一定会失败
	if err := f.StartDownload(); err != nil {
		t.Fatalf("StartDownload returned sync error: %v", err)
	}

	// 等待后台 goroutine 结束（clearDownloadState 会清空 downloadCancel）
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		f.mu.Lock()
		running := f.downloadCancel != nil
		f.mu.Unlock()
		if !running {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	f.mu.Lock()
	gotPath := f.readyPath
	gotName := f.readyAssetName
	f.mu.Unlock()

	if gotPath != "" {
		t.Errorf("readyPath should be cleared after failed new download, got %q", gotPath)
	}
	if gotName != "" {
		t.Errorf("readyAssetName should be cleared, got %q", gotName)
	}
	// 旧 zip 应该已经被 StartDownload 同步清理
	if _, err := os.Stat(stalePath); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("stale file should have been removed, stat err: %v", err)
	}
}

// ApplyUpdate 在新一轮下载失败后应明确报"尚未完成下载"，而不是静默安装旧版。
func TestApplyUpdate_AfterFailedRedownload_ReturnsNotReady(t *testing.T) {
	f, cleanup := newTestFacadeWith404(t)
	defer cleanup()

	stalePath := filepath.Join(t.TempDir(), "old-release.zip")
	_ = os.WriteFile(stalePath, []byte("old"), 0o644)

	f.mu.Lock()
	f.readyPath = stalePath
	f.readyAssetName = "old-release.zip"
	f.mu.Unlock()

	_ = f.StartDownload()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		f.mu.Lock()
		running := f.downloadCancel != nil
		f.mu.Unlock()
		if !running {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	err := f.ApplyUpdate()
	if err == nil {
		t.Fatal("expected ApplyUpdate to error after failed redownload")
	}
	if err.Error() != "尚未完成下载，无法安装" {
		t.Errorf("expected '尚未完成下载' error, got %q", err.Error())
	}
}
