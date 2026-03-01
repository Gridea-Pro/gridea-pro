package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// DefaultDevStartPort 开发模式默认起始端口
const DefaultDevStartPort = 3367

// DefaultProdStartPort 生产模式默认起始端口
const DefaultProdStartPort = 2077

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

	// Determine preferred port
	basePort := DefaultProdStartPort
	if s.IsDevelopmentMode() {
		basePort = DefaultDevStartPort
	}

	// Helper to try listen
	tryListen := func(p int) (net.Listener, error) {
		return net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", p))
	}

	// Try ports incrementally
	var listener net.Listener
	var err error
	maxRetries := 20

	for i := 0; i < maxRetries; i++ {
		port := basePort + i
		listener, err = tryListen(port)
		if err == nil {
			break // Successfully bound
		}
		// Only log if it's the specific port we wanted, to avoid spamming logs if we are just scanning
		if i == 0 {
			fmt.Fprintf(os.Stderr, "Preview Server: Port %d is in use, attempting to find next available port...\n", port)
		}
	}

	// If scanning fails, fallback to random port
	if err != nil {
		fmt.Fprintf(os.Stderr, "Preview Server: Could not find available port in range %d-%d, falling back to random port...\n", basePort, basePort+maxRetries-1)
		listener, err = tryListen(0)
		if err != nil {
			s.sendToast(ctx, "预览服务启动失败："+err.Error(), "error")
			return "", fmt.Errorf("无法获取可用端口: %w", err)
		}
	}

	// 2. 获取实际分配的端口
	s.port = listener.Addr().(*net.TCPAddr).Port

	// 3. 配置服务器
	mux := http.NewServeMux()

	// Create a custom handler that falls back to 404.html
	fileServer := http.FileServer(http.Dir(s.buildDir))
	customHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(s.buildDir, filepath.Clean(r.URL.Path))

		// Determine if the file or directory exists
		info, err := os.Stat(path)
		if os.IsNotExist(err) || (info != nil && info.IsDir() && r.URL.Path != "/") {
			// If it's a directory other than root, let's see if index.html exists inside it
			// http.FileServer auto-redirects or serves index.html if present.
			// However, if we just want a simple fallback for completely missing routes:
			if info != nil && info.IsDir() {
				indexPath := filepath.Join(path, "index.html")
				if _, err := os.Stat(indexPath); err == nil {
					fileServer.ServeHTTP(w, r)
					return
				}
			}

			// If file doesn't exist, serve 404.html
			notFoundPath := filepath.Join(s.buildDir, "404.html")
			if content, err := os.ReadFile(notFoundPath); err == nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(content)
				return
			}

			// If even 404.html doesn't exist, just let the original fileServer handle (and fail)
		}

		// Serve standard files
		fileServer.ServeHTTP(w, r)
	})

	// 禁用浏览器缓存，确保主题切换后立即加载最新的 CSS/JS
	mux.Handle("/", noCacheMiddleware(customHandler))

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

// noCacheMiddleware 禁用浏览器缓存，确保主题切换/配置修改后立即加载最新资源
func noCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}
