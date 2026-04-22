package facade

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ─── SHA256SUMS 解析 ──────────────────────────────────────────────────────────

func TestParseSha256Sums(t *testing.T) {
	content := `d41d8cd98f00b204e9800998ecf8427e  empty.txt
abc123def456  Gridea.Pro_v1.0.0_macos_arm64.zip
aaaaaaaaaaaa *binary-mode-file.bin
# this is a comment

invalidline
0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef  Gridea.Pro_v1.0.0_linux_amd64.AppImage
`

	tests := []struct {
		target string
		want   string
	}{
		{"empty.txt", "d41d8cd98f00b204e9800998ecf8427e"},
		{"Gridea.Pro_v1.0.0_macos_arm64.zip", "abc123def456"},
		{"binary-mode-file.bin", "aaaaaaaaaaaa"}, // '*' 前缀被剥掉
		{"Gridea.Pro_v1.0.0_linux_amd64.AppImage", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		{"missing.zip", ""},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			got, err := parseSha256Sums(strings.NewReader(content), tt.target)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseSha256Sums(%q) = %q, want %q", tt.target, got, tt.want)
			}
		})
	}
}

// ─── sha256File 计算 ─────────────────────────────────────────────────────────

func TestSha256File(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "data.bin")
	content := []byte("hello, gridea pro")
	if err := os.WriteFile(tmp, content, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := sha256File(tmp)
	if err != nil {
		t.Fatalf("sha256File: %v", err)
	}

	hh := sha256.Sum256(content)
	want := hex.EncodeToString(hh[:])
	if got != want {
		t.Errorf("sha256File = %q, want %q", got, want)
	}
}

// ─── verifyDownloadChecksum 集成场景 ──────────────────────────────────────────

func TestVerifyDownloadChecksum_NoSumsAssetReturnsNil(t *testing.T) {
	f := &UpdateFacade{httpClient: &http.Client{Timeout: 2 * time.Second}}
	tmp := filepath.Join(t.TempDir(), "asset.zip")
	_ = os.WriteFile(tmp, []byte("anything"), 0o644)

	if err := f.verifyDownloadChecksum(context.Background(), tmp, "asset.zip", nil); err != nil {
		t.Errorf("no sums asset should return nil for backward compat, got %v", err)
	}
}

func TestVerifyDownloadChecksum_HashMatches(t *testing.T) {
	// 准备一个文件并算出它的 SHA256
	tmp := filepath.Join(t.TempDir(), "asset.zip")
	content := []byte("binary payload")
	_ = os.WriteFile(tmp, content, 0o644)
	hh := sha256.Sum256(content)
	expected := hex.EncodeToString(hh[:])

	// 起一个 httptest 服务器充当 SHA256SUMS asset 源
	sumsBody := expected + "  asset.zip\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(sumsBody))
	}))
	defer srv.Close()

	f := &UpdateFacade{httpClient: &http.Client{Timeout: 2 * time.Second}}
	sumsAsset := &githubAsset{Name: "SHA256SUMS", DownloadURL: srv.URL}

	if err := f.verifyDownloadChecksum(context.Background(), tmp, "asset.zip", sumsAsset); err != nil {
		t.Errorf("expected verification to pass, got %v", err)
	}
}

func TestVerifyDownloadChecksum_HashMismatchFails(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "asset.zip")
	_ = os.WriteFile(tmp, []byte("binary payload"), 0o644)

	sumsBody := "0000000000000000000000000000000000000000000000000000000000000000  asset.zip\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(sumsBody))
	}))
	defer srv.Close()

	f := &UpdateFacade{httpClient: &http.Client{Timeout: 2 * time.Second}}
	sumsAsset := &githubAsset{Name: "SHA256SUMS", DownloadURL: srv.URL}

	err := f.verifyDownloadChecksum(context.Background(), tmp, "asset.zip", sumsAsset)
	if err == nil {
		t.Fatal("expected verification to fail on hash mismatch")
	}
	if !strings.Contains(err.Error(), "不匹配") {
		t.Errorf("expected 哈希不匹配 error, got %v", err)
	}
}

func TestVerifyDownloadChecksum_AssetNotInSums(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "asset.zip")
	_ = os.WriteFile(tmp, []byte("binary payload"), 0o644)

	sumsBody := "abc123  some-other-file.zip\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(sumsBody))
	}))
	defer srv.Close()

	f := &UpdateFacade{httpClient: &http.Client{Timeout: 2 * time.Second}}
	sumsAsset := &githubAsset{Name: "SHA256SUMS", DownloadURL: srv.URL}

	err := f.verifyDownloadChecksum(context.Background(), tmp, "asset.zip", sumsAsset)
	if err == nil {
		t.Fatal("expected error when asset not in SHA256SUMS")
	}
	if !strings.Contains(err.Error(), "未找到") {
		t.Errorf("expected '未找到' error, got %v", err)
	}
}

// ─── findSumsAsset 查找逻辑 ───────────────────────────────────────────────────

func TestFindSumsAsset(t *testing.T) {
	assets := []githubAsset{
		{Name: "Gridea.Pro_v1.0.0_macos_arm64.zip"},
		{Name: "SHA256SUMS"},
		{Name: "Gridea.Pro_v1.0.0_linux_amd64.AppImage"},
	}
	got := findSumsAsset(assets)
	if got == nil || got.Name != "SHA256SUMS" {
		t.Errorf("expected SHA256SUMS, got %+v", got)
	}

	got = findSumsAsset([]githubAsset{{Name: "foo.zip"}, {Name: "bar.zip"}})
	if got != nil {
		t.Errorf("expected nil when no SHA256SUMS present, got %+v", got)
	}
}

// keep the fmt import alive for future extensions
var _ = fmt.Sprintf
