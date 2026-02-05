package facade

import (
	"context"

	"gridea-pro/backend/internal/service"
)

// PreviewFacade 封装 PreviewService，提供预览功能的公开接口
type PreviewFacade struct {
	internal *service.PreviewService
}

// NewPreviewFacade 创建新的 PreviewFacade 实例
func NewPreviewFacade(buildDir string) *PreviewFacade {
	return &PreviewFacade{
		internal: service.NewPreviewService(buildDir),
	}
}

func (f *PreviewFacade) SetBuildDir(buildDir string) {
	f.internal.SetBuildDir(buildDir)
}

// SetContext 设置 Wails context
func (f *PreviewFacade) SetContext(ctx context.Context) {
	f.internal.SetContext(ctx)
}

// StartPreviewServer 启动预览服务器，返回预览 URL
func (f *PreviewFacade) StartPreviewServer() (string, error) {
	return f.internal.StartPreviewServer()
}

// StopPreviewServer 停止预览服务器
func (f *PreviewFacade) StopPreviewServer() error {
	return f.internal.StopPreviewServer()
}

// GetPreviewURL 获取当前预览服务的 URL
func (f *PreviewFacade) GetPreviewURL() string {
	return f.internal.GetPreviewURL()
}

// IsRunning 检查预览服务是否正在运行
func (f *PreviewFacade) IsRunning() bool {
	return f.internal.IsRunning()
}

// GetPort 获取当前使用的端口
func (f *PreviewFacade) GetPort() int {
	return f.internal.GetPort()
}
