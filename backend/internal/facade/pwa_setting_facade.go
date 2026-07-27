package facade

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"os"
	"path/filepath"
	"sync/atomic"
)

type PwaSettingFacade struct {
	// repo / appDir 用原子读写保存：切站（UpdateAppDir）会热替换它们，
	// 而 Wails 可能并发调用本 facade 的方法，原子读写避免 data race。
	repo   atomic.Value // 存 domain.PwaSettingRepository
	appDir atomic.Value // 存 string
}

func NewPwaSettingFacade(repo domain.PwaSettingRepository, appDir string) *PwaSettingFacade {
	f := &PwaSettingFacade{}
	f.repo.Store(repo)
	f.appDir.Store(appDir)
	return f
}

func (f *PwaSettingFacade) repository() domain.PwaSettingRepository {
	return f.repo.Load().(domain.PwaSettingRepository)
}

func (f *PwaSettingFacade) dir() string { return f.appDir.Load().(string) }

func (f *PwaSettingFacade) GetPwaSetting() (domain.PwaSetting, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.repository().GetPwaSetting(ctx)
}

func (f *PwaSettingFacade) SavePwaSettingFromFrontend(setting domain.PwaSetting) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.repository().SavePwaSetting(ctx, setting)
}

// HasCustomPwaIcon 检查是否存在自定义 PWA 图标
func (f *PwaSettingFacade) HasCustomPwaIcon() bool {
	iconPath := filepath.Join(f.dir(), "images", "pwa-icon.png")
	_, err := os.Stat(iconPath)
	return err == nil
}

// SavePwaIcon 保存自定义 PWA 图标（512x512）
func (f *PwaSettingFacade) SavePwaIcon(sourcePath string) error {
	if sourcePath == "" {
		return fmt.Errorf("图片路径不能为空")
	}

	destDir := filepath.Join(f.dir(), "images")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	destPath := filepath.Join(destDir, "pwa-icon.png")

	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("读取源文件失败: %w", err)
	}

	if err := os.WriteFile(destPath, sourceData, 0644); err != nil {
		return fmt.Errorf("保存 PWA 图标失败: %w", err)
	}

	return nil
}

// RemovePwaIcon 删除自定义 PWA 图标，恢复使用头像
func (f *PwaSettingFacade) RemovePwaIcon() error {
	iconPath := filepath.Join(f.dir(), "images", "pwa-icon.png")
	if err := os.Remove(iconPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除 PWA 图标失败: %w", err)
	}
	return nil
}
