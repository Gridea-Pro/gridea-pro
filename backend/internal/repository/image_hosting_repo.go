package repository

import (
	"context"
	"encoding/json"
	"gridea-pro/backend/internal/domain"
	"os"
	"path/filepath"
	"sync"
)

type imageHostingRepository struct {
	appDir  string
	mu      sync.RWMutex
	cached  *domain.ImageHostingSetting
}

func NewImageHostingRepository(appDir string) domain.ImageHostingRepository {
	return &imageHostingRepository{appDir: appDir}
}

func (r *imageHostingRepository) configPath() string {
	return filepath.Join(r.appDir, "config", "image_hosting.json")
}

func (r *imageHostingRepository) GetSetting() (*domain.ImageHostingSetting, error) {
	r.mu.RLock()
	if r.cached != nil {
		r.mu.RUnlock()
		return r.cached, nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cached != nil {
		return r.cached, nil
	}

	path := r.configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			r.cached = &domain.ImageHostingSetting{Enabled: false}
			return r.cached, nil
		}
		return nil, err
	}

	var setting domain.ImageHostingSetting
	if err := json.Unmarshal(data, &setting); err != nil {
		return nil, err
	}
	r.cached = &setting
	return r.cached, nil
}

func (r *imageHostingRepository) SaveSetting(setting *domain.ImageHostingSetting) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	path := r.configPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(setting, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	r.cached = setting
	return nil
}

func (r *imageHostingRepository) Invalidate() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cached = nil
}

var _ context.Context = nil
