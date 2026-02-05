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

// MaxPortAttempts 最大端口探测尝试次数
const MaxPortAttempts = 100

// DefaultDevStartPort 开发模式默认起始端口
const DefaultDevStartPort = 3367

// DefaultProdStartPort 生产模式默认起始端口
const DefaultProdStartPort = 6060

// PreviewService 管理预览服务器的生命周期
type PreviewService struct {
	ctx       context.Context
	server    *http.Server
	port      int
	buildDir  string
	mu        sync.Mutex
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

// SetContext 设置 Wails context
func (s *PreviewService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// IsDevelopmentMode 检测当前是否为开发模式
// 通过检查 WAILS_DEV 环境变量或其他标志来判断
func (s *PreviewService) IsDevelopmentMode() bool {
	// Wails 在开发模式下会设置这个环境变量
	if os.Getenv("devserver") != "" {
		return true
	}
	// 备用检测：检查 wails 相关的开发模式环境变量
	if os.Getenv("WAILS_DEV") != "" {
		return true
	}
	return false
}

// GetStartPort 根据运行模式获取起始端口
func (s *PreviewService) GetStartPort() int {
	if s.IsDevelopmentMode() {
		return DefaultDevStartPort
	}
	return DefaultProdStartPort
}

// GetFreePort 从指定端口开始探测，找到一个空闲端口
// 最多尝试 MaxPortAttempts 次
func (s *PreviewService) GetFreePort(startPort int) (int, error) {
	for i := 0; i < MaxPortAttempts; i++ {
		port := startPort + i
		addr := fmt.Sprintf("127.0.0.1:%d", port)

		// 尝试在该地址上监听
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			// 端口被占用，尝试下一个
			continue
		}
		// 端口可用，关闭并返回
		listener.Close()
		return port, nil
	}

	return 0, fmt.Errorf("无法在端口 %d-%d 范围内找到空闲端口", startPort, startPort+MaxPortAttempts-1)
}

// StartPreviewServer 启动预览服务器
// 返回预览 URL 和可能的错误
func (s *PreviewService) StartPreviewServer() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果服务已经在运行，直接返回当前 URL
	if s.isRunning && s.server != nil {
		return s.GetPreviewURL(), nil
	}

	// 获取空闲端口
	startPort := s.GetStartPort()
	freePort, err := s.GetFreePort(startPort)
	if err != nil {
		s.sendToast("预览服务启动失败："+err.Error(), "error")
		return "", err
	}

	s.port = freePort

	// 创建文件服务器
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(s.buildDir))
	mux.Handle("/", fs)

	// 配置 HTTP 服务器
	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	s.server = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
	}

	// 在 goroutine 中启动服务器
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// 服务器异常关闭时记录错误
			fmt.Printf("预览服务器错误: %v\n", err)
		}
	}()

	s.isRunning = true

	// 给服务器一点时间启动
	time.Sleep(100 * time.Millisecond)

	url := s.GetPreviewURL()
	fmt.Printf("预览服务器已启动: %s\n", url)

	return url, nil
}

// StopPreviewServer 平滑关闭预览服务器
func (s *PreviewService) StopPreviewServer() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server == nil || !s.isRunning {
		return nil
	}

	// 使用 5 秒超时进行优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		// 如果优雅关闭失败，强制关闭
		s.server.Close()
		fmt.Printf("预览服务器强制关闭: %v\n", err)
	} else {
		fmt.Println("预览服务器已平滑关闭")
	}

	s.server = nil
	s.isRunning = false
	s.port = 0

	return nil
}

// GetPreviewURL 返回当前预览服务器的 URL
func (s *PreviewService) GetPreviewURL() string {
	if s.port == 0 {
		return ""
	}
	return fmt.Sprintf("http://127.0.0.1:%d", s.port)
}

// IsRunning 返回预览服务器是否正在运行
func (s *PreviewService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isRunning
}

// GetPort 返回当前使用的端口
func (s *PreviewService) GetPort() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.port
}

// sendToast 向前端发送 Toast 通知
func (s *PreviewService) sendToast(message, toastType string) {
	if s.ctx == nil {
		return
	}
	runtime.EventsEmit(s.ctx, "app:toast", map[string]interface{}{
		"message":  message,
		"type":     toastType,
		"duration": 3000,
	})
}
