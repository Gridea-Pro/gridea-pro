package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// DefaultDevStartPort 开发模式默认起始端口
const DefaultDevStartPort = 3367

// DefaultProdStartPort 生产模式默认起始端口
const DefaultProdStartPort = 6060

// PreviewService 管理预览服务器的生命周期
type PreviewService struct {
	server    *http.Server
	port      int
	buildDir  string
	mu        sync.RWMutex
	isRunning bool
}

// NewPreviewService 创建新的预览服务实例
func NewPreviewService(buildDir string) *PreviewService {
	return &PreviewService{
		buildDir: buildDir,
		port:     0,
	}
}

func (s *PreviewService) SetBuildDir(buildDir string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buildDir = buildDir
}

func (s *PreviewService) IsDevelopmentMode() bool {
	if os.Getenv("devserver") != "" {
		return true
	}
	if os.Getenv("WAILS_DEV") != "" {
		return true
	}
	return false
}

// StartPreviewServer 启动预览服务器
func (s *PreviewService) StartPreviewServer(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning && s.server != nil {
		return fmt.Sprintf("http://127.0.0.1:%d", s.port), nil
	}

	// 1. 原子操作：监听端口 0，系统自动分配空闲端口
	// 注意：我们不关闭 listener，而是直接传给 Server 使用，避免端口被抢占
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		s.sendToast(ctx, "预览服务启动失败："+err.Error(), "error")
		return "", fmt.Errorf("无法获取空闲端口: %w", err)
	}

	// 2. 获取实际分配的端口
	s.port = listener.Addr().(*net.TCPAddr).Port

	// 3. 配置服务器
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(s.buildDir))
	mux.Handle("/", fs)

	s.server = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
	}

	// 4. 在 goroutine 中启动，使用 Serve(listener) 而不是 ListenAndServe
	go func() {
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "预览服务器错误: %v\n", err)
		}
	}()

	s.isRunning = true

	// 给一点启动缓冲时间（可选，Server.Serve 已经是即时的了）
	time.Sleep(50 * time.Millisecond)

	url := fmt.Sprintf("http://127.0.0.1:%d", s.port)
	fmt.Fprintf(os.Stderr, "预览服务器已启动: %s\n", url)

	return url, nil
}

// StopPreviewServer 平滑关闭预览服务器
func (s *PreviewService) StopPreviewServer() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server == nil || !s.isRunning {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.server.Close()
		fmt.Fprintf(os.Stderr, "预览服务器强制关闭: %v\n", err)
	} else {
		fmt.Fprintln(os.Stderr, "预览服务器已平滑关闭")
	}

	s.server = nil
	s.isRunning = false
	s.port = 0

	return nil
}

func (s *PreviewService) GetPreviewURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.port == 0 {
		return ""
	}
	return fmt.Sprintf("http://127.0.0.1:%d", s.port)
}

func (s *PreviewService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

func (s *PreviewService) GetPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.port
}

func (s *PreviewService) sendToast(ctx context.Context, message, toastType string) {
	if ctx == nil {
		return
	}
	runtime.EventsEmit(ctx, "app:toast", map[string]interface{}{
		"message":  message,
		"type":     toastType,
		"duration": 3000,
	})
}
