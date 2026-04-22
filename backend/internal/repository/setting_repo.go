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
	cache  *domain.Setting
	loaded bool
}

func NewSettingRepository(appDir string) domain.SettingRepository {
	return &settingRepository{
		appDir: appDir,
		cache:  nil,
		loaded: false,
	}
}

func (r *settingRepository) loadIfNeeded() error {
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

	settingPath := filepath.Join(r.appDir, "config", "setting.json")
	var setting domain.Setting
	if err := LoadJSONFile(settingPath, &setting); err != nil {
		// If load fails, return empty setting but mark loaded to avoid repeated disk reads
		// Assuming error means file missing or invalid.
		// Detailed error handling might be better, but for now:
		r.cache = &domain.Setting{}
		r.loaded = true
		return nil
	}

	r.cache = &setting
	r.loaded = true
	return nil
}

func (r *settingRepository) GetSetting(ctx context.Context) (domain.Setting, error) {
	if err := r.loadIfNeeded(); err != nil {
		return domain.Setting{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.cache == nil {
		return domain.Setting{}, nil
	}
	// 深拷贝：PlatformConfigs / 内嵌 map 是引用类型，调用方若在其上写入
	// 敏感字段（见 DeployService 的 InjectCredentials），会反向污染 cache。
	// 修复 issue #39：凭证反向泄漏到前端。
	return r.cache.Clone(), nil
}

func (r *settingRepository) SaveSetting(ctx context.Context, setting domain.Setting) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	settingPath := filepath.Join(r.appDir, "config", "setting.json")
	if err := SaveJSONFile(settingPath, setting); err != nil {
		return err
	}

	// 同样深拷贝再存入 cache，避免调用方后续修改入参 map 影响缓存一致性。
	cached := setting.Clone()
	r.cache = &cached
	r.loaded = true
	return nil
}

func (r *settingRepository) Invalidate() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = nil
	r.loaded = false
}
