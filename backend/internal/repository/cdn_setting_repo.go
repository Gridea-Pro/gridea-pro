package repository

import (
	"context"
	"errors"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"io/fs"
	"log"
	"path/filepath"
	"sync"
)

type cdnSettingRepository struct {
	mu     sync.RWMutex
	appDir string
	cache  *domain.CdnSetting
	loaded bool
}

func NewCdnSettingRepository(appDir string) domain.CdnSettingRepository {
	return &cdnSettingRepository{
		appDir: appDir,
		cache:  nil,
		loaded: false,
	}
}

func (r *cdnSettingRepository) loadIfNeeded() error {
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

	settingPath := filepath.Join(r.appDir, "config", "cdn_setting.json")
	var setting domain.CdnSetting
	if err := LoadJSONFile(settingPath, &setting); err != nil {
		// 文件不存在：合法初始状态，锁进 cache。
		if errors.Is(err, fs.ErrNotExist) {
			r.cache = &domain.CdnSetting{}
			r.loaded = true
			return nil
		}
		// 其他错误（JSON 损坏/权限/IO）：绝不能把空配置锁进 cache，
		// 否则用户保存时会用空 Token/仓库配置覆盖磁盘真实配置——必须向上抛，让下次访问重试。
		log.Printf("[repo] cdnSettingRepo load %s failed: %v", settingPath, err)
		return fmt.Errorf("load cdn setting: %w", err)
	}

	r.cache = &setting
	r.loaded = true
	return nil
}

func (r *cdnSettingRepository) GetCdnSetting(ctx context.Context) (domain.CdnSetting, error) {
	if err := r.loadIfNeeded(); err != nil {
		return domain.CdnSetting{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.cache == nil {
		return domain.CdnSetting{}, nil
	}
	return *r.cache, nil
}

func (r *cdnSettingRepository) SaveCdnSetting(ctx context.Context, setting domain.CdnSetting) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	settingPath := filepath.Join(r.appDir, "config", "cdn_setting.json")
	if err := SaveJSONFile(settingPath, setting); err != nil {
		return err
	}

	r.cache = &setting
	r.loaded = true
	return nil
}
