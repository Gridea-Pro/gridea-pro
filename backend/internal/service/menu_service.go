package service

import (
	"context"
	"gridea-pro/backend/internal/domain"
)

type MenuService struct {
	repo domain.MenuRepository
}

func NewMenuService(repo domain.MenuRepository) *MenuService {
	return &MenuService{repo: repo}
}

func (s *MenuService) LoadMenus(ctx context.Context) ([]domain.Menu, error) {
	return s.repo.GetAll(ctx)
}

func (s *MenuService) SaveMenus(ctx context.Context, menus []domain.Menu) error {
	return s.repo.SaveAll(ctx, menus)
}
