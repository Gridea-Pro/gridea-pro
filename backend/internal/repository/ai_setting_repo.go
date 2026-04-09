package repository

import (
	"context"
	"sync"

	"gridea-pro/backend/internal/config"
	"gridea-pro/backend/internal/domain"
)

// aiSettingRepository AI 配置存储
//
// 存储位置：应用级 config.json 中的 aiSetting 字段
// 路径示例：
//   - macOS:   ~/Library/Application Support/Gridea Pro/config.json
//   - Linux:   ~/.config/Gridea Pro/config.json
//   - Windows: %AppData%/Gridea Pro/config.json
//
// 与站点目录解耦的原因：
//  1. 包含 API Key 等敏感信息，不能跟随站点目录被 iCloud / Dropbox 同步
//  2. 不能跟随站点目录被推到 Git 仓库
//  3. API Key 是用户级配置，跨站点共享，避免每个站点都要重新填写
type aiSettingRepository struct {
	mu     sync.RWMutex
	cm     *config.ConfigManager
	cache  *domain.AISetting
	loaded bool
}

func NewAISettingRepository(cm *config.ConfigManager) domain.AISettingRepository {
	return &aiSettingRepository{cm: cm}
}

func (r *aiSettingRepository) loadIfNeeded() {
	r.mu.RLock()
	if r.loaded {
		r.mu.RUnlock()
		return
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.loaded {
		return
	}

	setting, err := r.cm.GetAISetting()
	if err != nil {
		r.cache = &domain.AISetting{}
	} else {
		r.cache = &setting
	}
	r.loaded = true
}

func (r *aiSettingRepository) GetAISetting(ctx context.Context) (domain.AISetting, error) {
	r.loadIfNeeded()

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.cache == nil {
		return domain.AISetting{}, nil
	}
	return *r.cache, nil
}

func (r *aiSettingRepository) SaveAISetting(ctx context.Context, setting domain.AISetting) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.cm.SaveAISetting(setting); err != nil {
		return err
	}
	r.cache = &setting
	r.loaded = true
	return nil
}
