package deploy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// setNetlifyAPIBase 是 test helper，用于在测试期间把 netlifyAPIBase 指向 httptest。
// 通过 const 引用；为了可覆盖，把 API base 改为 package-level var。
// 不想改原文件常量定义 → 用 monkey patching 不成，只能在同包里直接赋值。

func mustStatusResponder(sequence []string, deployURL string) http.HandlerFunc {
	var idx atomic.Int32
	return func(w http.ResponseWriter, r *http.Request) {
		i := int(idx.Add(1)) - 1
		if i >= len(sequence) {
			i = len(sequence) - 1
		}
		state := sequence[i]
		var body string
		switch state {
		case "ready":
			body = fmt.Sprintf(`{"id":"d1","state":"ready","deploy_url":%q}`, deployURL)
		case "error":
			body = `{"id":"d1","state":"error","error_message":"build failed: missing _redirects"}`
		default:
			body = fmt.Sprintf(`{"id":"d1","state":%q}`, state)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}
}

func TestPollDeployStatus_UploadingProcessingReady(t *testing.T) {
	srv := httptest.NewServer(mustStatusResponder([]string{"uploading", "processing", "ready"}, "https://site.example.com"))
	defer srv.Close()

	oldBase := netlifyAPIBase
	netlifyAPIBase = srv.URL
	defer func() { netlifyAPIBase = oldBase }()

	p := &NetlifyProvider{client: &http.Client{Timeout: 2 * time.Second}}
	logs := []string{}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	status, err := p.pollDeployStatus(ctx, "d1", "faketoken", func(msg string) { logs = append(logs, msg) })
	if err != nil {
		t.Fatalf("expected ready, got err: %v", err)
	}
	if status.State != "ready" {
		t.Errorf("expected ready state, got %q", status.State)
	}
	if status.DeployURL != "https://site.example.com" {
		t.Errorf("expected deploy URL, got %q", status.DeployURL)
	}
}

func TestPollDeployStatus_ErrorState(t *testing.T) {
	srv := httptest.NewServer(mustStatusResponder([]string{"processing", "error"}, ""))
	defer srv.Close()

	oldBase := netlifyAPIBase
	netlifyAPIBase = srv.URL
	defer func() { netlifyAPIBase = oldBase }()

	p := &NetlifyProvider{client: &http.Client{Timeout: 2 * time.Second}}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := p.pollDeployStatus(ctx, "d1", "faketoken", func(string) {})
	if err == nil {
		t.Fatal("expected error for error state")
	}
	if !strings.Contains(err.Error(), "build failed") {
		t.Errorf("expected error_message passthrough, got %v", err)
	}
}

func TestPollDeployStatus_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	oldBase := netlifyAPIBase
	netlifyAPIBase = srv.URL
	defer func() { netlifyAPIBase = oldBase }()

	p := &NetlifyProvider{client: &http.Client{Timeout: 2 * time.Second}}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := p.pollDeployStatus(ctx, "d1", "bad", func(string) {})
	if err == nil {
		t.Fatal("expected auth error")
	}
	if !strings.Contains(err.Error(), "Token 无效") {
		t.Errorf("expected Token 无效 error, got %v", err)
	}
}

func TestPollDeployStatus_ContextCancelled(t *testing.T) {
	// Server 永远返回 processing
	srv := httptest.NewServer(mustStatusResponder([]string{"processing"}, ""))
	defer srv.Close()

	oldBase := netlifyAPIBase
	netlifyAPIBase = srv.URL
	defer func() { netlifyAPIBase = oldBase }()

	p := &NetlifyProvider{client: &http.Client{Timeout: 2 * time.Second}}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := p.pollDeployStatus(ctx, "d1", "tok", func(string) {})
	if err == nil {
		t.Fatal("expected ctx cancellation error")
	}
}
