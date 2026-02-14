package service

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"sync"
)

type MenuService struct {
	repo domain.MenuRepository
	mu   sync.RWMutex
}

func NewMenuService(repo domain.MenuRepository) *MenuService {
	return &MenuService{repo: repo}
}

func (s *MenuService) LoadMenus(ctx context.Context) ([]domain.Menu, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.repo.GetAll(ctx)
}

func (s *MenuService) SaveMenus(ctx context.Context, menus []domain.Menu) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repo.SaveAll(ctx, menus)
}
