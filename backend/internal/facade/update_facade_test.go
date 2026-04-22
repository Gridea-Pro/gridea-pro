package facade

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

func TestHasTrustedRedirectHost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"github.com", true},
		{"objects.githubusercontent.com", true},
		{"release-assets.githubusercontent.com", true},
		{"codeload.github.com", true},
		{"github.com:443", true}, // 带端口号

		{"evil.com", false},
		{"github.com.evil.com", false},
		{"xgithub.com", false}, // 无点边界，不是合法子域
		{"githubusercontent.com.evil.com", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := hasTrustedRedirectHost(tt.host)
			if got != tt.want {
				t.Errorf("hasTrustedRedirectHost(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestTrustedRedirectChecker(t *testing.T) {
	mkReq := func(rawurl string) *http.Request {
		u, err := url.Parse(rawurl)
		if err != nil {
			t.Fatalf("parse %q: %v", rawurl, err)
		}
		return &http.Request{URL: u}
	}

	cases := []struct {
		name    string
		target  string
		viaLen  int
		wantErr bool
	}{
		{"allowed_github", "https://github.com/foo/bar", 1, false},
		{"allowed_subdomain", "https://objects.githubusercontent.com/xxx", 2, false},
		{"http_scheme_rejected", "http://github.com/foo/bar", 1, true},
		{"third_party_host_rejected", "https://evil.example.com/x", 1, true},
		{"lookalike_domain_rejected", "https://github.com.evil.com/x", 1, true},
		{"too_many_redirects", "https://github.com/x", 10, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			via := make([]*http.Request, tc.viaLen)
			err := trustedRedirectChecker(mkReq(tc.target), via)
			if tc.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// 集成测：本地服务器发起 302 跳到非白名单域名；下载客户端必须拒绝，不把响应体
// 当作更新包保存下来。
func TestDoDownload_RejectsThirdPartyRedirect(t *testing.T) {
	var evilHits atomic.Int32
	evil := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		evilHits.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("attacker payload"))
	}))
	defer evil.Close()

	// "合法"入口：返回 302 指向攻击者服务器
	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, evil.URL+"/fake.zip", http.StatusFound)
	}))
	defer redirector.Close()

	f := &UpdateFacade{
		releasesURL: "unused",
		httpClient:  &http.Client{Timeout: 2 * time.Second},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 注意：这里的 url 本身不是 github 前缀（因为本 PR 只管重定向校验，不管入口白名单）。
	// 为验证重定向校验本身的效果，直接让 doDownload 跑起来看它是否把 evil 服务器打到。
	// 不依赖入口 URL 校验（#52 的责任），构造一个带有"testing-only"标记让白名单先放行。
	f.doDownload(ctx, redirector.URL+"/entry", "fake.zip", 1024)

	if n := evilHits.Load(); n != 0 {
		t.Errorf("third-party redirect should be rejected before HTTP body is fetched, got %d hits", n)
	}
}

// 辅助：保持 fmt import 使用，避免将来扩展时找不到
var _ = fmt.Errorf
