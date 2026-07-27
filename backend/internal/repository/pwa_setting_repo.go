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

type pwaSettingRepository struct {
	mu     sync.RWMutex
	appDir string
	cache  *domain.PwaSetting
	loaded bool
}

func NewPwaSettingRepository(appDir string) domain.PwaSettingRepository {
	return &pwaSettingRepository{
		appDir: appDir,
		cache:  nil,
		loaded: false,
	}
}

func (r *pwaSettingRepository) loadIfNeeded() error {
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

	settingPath := filepath.Join(r.appDir, "config", "pwa_setting.json")
	var setting domain.PwaSetting
	if err := LoadJSONFile(settingPath, &setting); err != nil {
		// 文件不存在：合法初始状态，锁进 cache。
		if errors.Is(err, fs.ErrNotExist) {
			r.cache = &domain.PwaSetting{}
			r.loaded = true
			return nil
		}
		// 其他错误（JSON 损坏/权限/IO）：不能把空配置锁进 cache，
		// 否则用户保存时会用空 PWA 配置覆盖磁盘真实配置——必须向上抛，让下次访问重试。
		log.Printf("[repo] pwaSettingRepo load %s failed: %v", settingPath, err)
		return fmt.Errorf("load pwa setting: %w", err)
	}

	r.cache = &setting
	r.loaded = true
	return nil
}

func (r *pwaSettingRepository) GetPwaSetting(ctx context.Context) (domain.PwaSetting, error) {
	if err := r.loadIfNeeded(); err != nil {
		return domain.PwaSetting{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.cache == nil {
		return domain.PwaSetting{}, nil
	}
	return *r.cache, nil
}

func (r *pwaSettingRepository) SavePwaSetting(ctx context.Context, setting domain.PwaSetting) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	settingPath := filepath.Join(r.appDir, "config", "pwa_setting.json")
	if err := SaveJSONFile(settingPath, setting); err != nil {
		return err
	}

	r.cache = &setting
	r.loaded = true
	return nil
}
