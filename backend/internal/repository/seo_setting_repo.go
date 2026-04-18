package repository

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"path/filepath"
	"sync"
)

type seoSettingRepository struct {
	mu     sync.RWMutex
	appDir string
	cache  *domain.SeoSetting
	loaded bool
}

func NewSeoSettingRepository(appDir string) domain.SeoSettingRepository {
	return &seoSettingRepository{
		appDir: appDir,
		cache:  nil,
		loaded: false,
	}
}

func (r *seoSettingRepository) loadIfNeeded() error {
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

	settingPath := filepath.Join(r.appDir, "config", "seo_setting.json")
	var setting domain.SeoSetting
	if err := LoadJSONFile(settingPath, &setting); err != nil {
		def := domain.DefaultSeoSetting()
		r.cache = &def
		r.loaded = true
		return nil
	}

	r.cache = &setting
	r.loaded = true
	return nil
}

func (r *seoSettingRepository) GetSeoSetting(ctx context.Context) (domain.SeoSetting, error) {
	if err := r.loadIfNeeded(); err != nil {
		return domain.SeoSetting{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.cache == nil {
		return domain.SeoSetting{}, nil
	}
	return *r.cache, nil
}

func (r *seoSettingRepository) SaveSeoSetting(ctx context.Context, setting domain.SeoSetting) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	settingPath := filepath.Join(r.appDir, "config", "seo_setting.json")
	if err := SaveJSONFile(settingPath, setting); err != nil {
		return err
	}

	r.cache = &setting
	r.loaded = true
	return nil
}
