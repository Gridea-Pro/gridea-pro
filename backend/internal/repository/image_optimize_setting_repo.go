package repository

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type imageOptimizeSettingRepository struct {
	mu     sync.RWMutex
	appDir string
	cache  *domain.ImageOptimizeSetting
	loaded bool
}

func NewImageOptimizeSettingRepository(appDir string) domain.ImageOptimizeSettingRepository {
	return &imageOptimizeSettingRepository{
		appDir: appDir,
		cache:  nil,
		loaded: false,
	}
}

// loadIfNeeded 懒加载：RLock 快速路径检查 → Lock 双检锁写入。
// 文件不存在时不报错，返回零值设置。
func (r *imageOptimizeSettingRepository) loadIfNeeded() error {
	r.mu.RLock()
	if r.loaded {
		r.mu.RUnlock()
		return nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.loaded {
		return nil
	}

	settingPath := filepath.Join(r.appDir, "config", "image_optimize_setting.json") // 站点级配置
	var setting domain.ImageOptimizeSetting
	if err := LoadJSONFile(settingPath, &setting); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("[imageOptRepo] 读取配置失败，使用默认值: %v", err)
		}
		r.cache = &domain.ImageOptimizeSetting{} // 文件不存在 → 零值（功能关闭）
		r.loaded = true
		return nil
	}

	r.cache = &setting
	r.loaded = true
	return nil
}

func (r *imageOptimizeSettingRepository) GetImageOptimizeSetting(ctx context.Context) (domain.ImageOptimizeSetting, error) {
	if err := r.loadIfNeeded(); err != nil {
		return domain.ImageOptimizeSetting{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.cache == nil {
		return domain.ImageOptimizeSetting{}, nil
	}
	return *r.cache, nil
}

func (r *imageOptimizeSettingRepository) SaveImageOptimizeSetting(ctx context.Context, setting domain.ImageOptimizeSetting) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	settingPath := filepath.Join(r.appDir, "config", "image_optimize_setting.json")
	if err := SaveJSONFile(settingPath, setting); err != nil {
		return err
	}

	r.cache = &setting
	r.loaded = true
	return nil
}
