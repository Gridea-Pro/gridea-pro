package app

import (
	"context"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// PreferencesWindow 管理系统设置窗口
type PreferencesWindow struct {
	ctx context.Context
	mu  sync.Mutex
}

// NewPreferencesWindow 创建新的设置窗口管理器
func NewPreferencesWindow() *PreferencesWindow {
	return &PreferencesWindow{}
}

// SetContext 设置上下文
func (pw *PreferencesWindow) SetContext(ctx context.Context) {
	pw.ctx = ctx
}

// Show 显示设置窗口
func (pw *PreferencesWindow) Show() {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	if pw.ctx == nil {
		return
	}

	// 创建新窗口配置
	runtime.WindowCenter(pw.ctx)

	// 发送事件到主窗口，让其显示设置对话框
	runtime.EventsEmit(pw.ctx, "show-preferences-dialog")
}
