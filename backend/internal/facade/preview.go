package facade

import (
	"context"
	"sync/atomic"

	"gridea-pro/backend/internal/service"
)

// PreviewFacade 封装 PreviewService，提供预览功能的公开接口
type PreviewFacade struct {
	// internal 用原子读写保存：切站（UpdateAppDir）会热替换它，
	// 而 Wails 可能并发调用本 facade 的方法，原子读写避免"读到替换一半的状态"的 data race。
	internal atomic.Pointer[service.PreviewService]
}

// NewPreviewFacade 创建新的 PreviewFacade 实例
func NewPreviewFacade(s *service.PreviewService) *PreviewFacade {
	f := &PreviewFacade{}
	f.internal.Store(s)
	return f
}

func (f *PreviewFacade) svc() *service.PreviewService { return f.internal.Load() }

func (f *PreviewFacade) SetBuildDir(buildDir string) {
	f.svc().SetBuildDir(buildDir)
}

// SetContext 设置 Wails context
// Deprecated: PreviewFacade follows the stateless pattern now.
// The context is retrieved via the global WailsContext variable.
func (f *PreviewFacade) SetContext(ctx context.Context) {
	// f.internal.SetContext(ctx) // Service no longer has SetContext
}

// StartPreviewServer 启动预览服务器，返回预览 URL
func (f *PreviewFacade) StartPreviewServer() (string, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().StartPreviewServer(ctx)
}

// StopPreviewServer 停止预览服务器
func (f *PreviewFacade) StopPreviewServer() error {
	return f.svc().StopPreviewServer()
}

// GetPreviewURL 获取当前预览服务的 URL
func (f *PreviewFacade) GetPreviewURL() string {
	return f.svc().GetPreviewURL()
}

// IsRunning 检查预览服务是否正在运行
func (f *PreviewFacade) IsRunning() bool {
	return f.svc().IsRunning()
}

// GetPort 获取当前使用的端口
func (f *PreviewFacade) GetPort() int {
	return f.svc().GetPort()
}
