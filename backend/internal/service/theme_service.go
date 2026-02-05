package service

import (
	"context"
	"gridea-pro/backend/internal/domain"
)

type ThemeService struct {
	repo domain.ThemeRepository
}

func NewThemeService(repo domain.ThemeRepository) *ThemeService {
	return &ThemeService{repo: repo}
}

func (s *ThemeService) LoadThemes(ctx context.Context) ([]domain.Theme, error) {
	return s.repo.GetAll(ctx)
}

func (s *ThemeService) LoadThemeConfig(ctx context.Context) (domain.ThemeConfig, error) {
	return s.repo.GetConfig(ctx)
}

func (s *ThemeService) SaveThemeConfig(ctx context.Context, config domain.ThemeConfig) error {
	// Should update config.json
	return s.repo.SaveConfig(ctx, config)
}
