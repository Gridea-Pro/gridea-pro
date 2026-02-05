package repository

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"path/filepath"
	"sync"
)

type settingRepository struct {
	mu     sync.RWMutex
	appDir string
}

func NewSettingRepository(appDir string) domain.SettingRepository {
	return &settingRepository{appDir: appDir}
}

func (r *settingRepository) GetSetting(ctx context.Context) (domain.Setting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	settingPath := filepath.Join(r.appDir, "config", "setting.json")
	var setting domain.Setting
	if err := LoadJSONFile(settingPath, &setting); err != nil {
		return domain.Setting{}, nil
	}
	return setting, nil
}

func (r *settingRepository) SaveSetting(ctx context.Context, setting domain.Setting) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	settingPath := filepath.Join(r.appDir, "config", "setting.json")
	return SaveJSONFile(settingPath, setting)
}
