package deploy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func mkGetReq(ctx context.Context, url string) func() (*http.Request, error) {
	return func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	}
}

func TestDoHTTPWithRetry_RetriesOn5xxAndSucceeds(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := DoHTTPWithRetry(ctx, &http.Client{Timeout: 2 * time.Second},
		mkGetReq(ctx, srv.URL), HTTPRetryPolicy{MaxAttempts: 3, BaseDelay: 10 * time.Millisecond}, nil)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if got := hits.Load(); got != 3 {
		t.Errorf("expected 3 hits, got %d", got)
	}
}

func TestDoHTTPWithRetry_4xxNotRetried(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := DoHTTPWithRetry(ctx, &http.Client{Timeout: 2 * time.Second},
		mkGetReq(ctx, srv.URL), HTTPRetryPolicy{MaxAttempts: 3, BaseDelay: 5 * time.Millisecond}, nil)
	if err != nil {
		t.Fatalf("4xx 不应返回 error，应 return resp，实际 %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
	if got := hits.Load(); got != 1 {
		t.Errorf("4xx 不应触发重试，实际 %d hits", got)
	}
}

func TestDoHTTPWithRetry_ExhaustedReturnsError(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := DoHTTPWithRetry(ctx, &http.Client{Timeout: 2 * time.Second},
		mkGetReq(ctx, srv.URL), HTTPRetryPolicy{MaxAttempts: 3, BaseDelay: 5 * time.Millisecond}, nil)
	if err == nil {
		t.Fatal("expected exhaustion error")
	}
	if !strings.Contains(err.Error(), "重试 3 次") {
		t.Errorf("expected '重试 3 次' in err, got %v", err)
	}
	if got := hits.Load(); got != 3 {
		t.Errorf("expected 3 attempts, got %d", got)
	}
}

func TestDoHTTPWithRetry_RespectsRetryAfter(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	_, err := DoHTTPWithRetry(ctx, &http.Client{Timeout: 2 * time.Second},
		mkGetReq(ctx, srv.URL),
		// BaseDelay 刻意很小；如果 Retry-After 生效，总耗时应至少 1s
		HTTPRetryPolicy{MaxAttempts: 3, BaseDelay: 10 * time.Millisecond}, nil)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed < 900*time.Millisecond {
		t.Errorf("expected to wait at least ~1s for Retry-After, got %v", elapsed)
	}
}

func TestDoHTTPWithRetry_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_, err := DoHTTPWithRetry(ctx, &http.Client{Timeout: 2 * time.Second},
		mkGetReq(ctx, srv.URL),
		HTTPRetryPolicy{MaxAttempts: 5, BaseDelay: 100 * time.Millisecond}, nil)
	if err == nil {
		t.Fatal("expected ctx cancellation error")
	}
}

func TestParseRetryAfter(t *testing.T) {
	if d := parseRetryAfter("3"); d != 3*time.Second {
		t.Errorf("seconds parse: got %v", d)
	}
	if d := parseRetryAfter(""); d != 0 {
		t.Errorf("empty: got %v", d)
	}
	if d := parseRetryAfter("invalid"); d != 0 {
		t.Errorf("invalid: got %v", d)
	}
}
