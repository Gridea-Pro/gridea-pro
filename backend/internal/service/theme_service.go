package service

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"sync"
)

type ThemeService struct {
	repo domain.ThemeRepository
	mu   sync.RWMutex
}

func NewThemeService(repo domain.ThemeRepository) *ThemeService {
	return &ThemeService{repo: repo}
}

func (s *ThemeService) LoadThemes(ctx context.Context) ([]domain.Theme, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.repo.GetAll(ctx)
}

func (s *ThemeService) LoadThemeConfig(ctx context.Context) (domain.ThemeConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.repo.GetConfig(ctx)
}

func (s *ThemeService) SaveThemeConfig(ctx context.Context, config domain.ThemeConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Should update config.json
	return s.repo.SaveConfig(ctx, config)
}
