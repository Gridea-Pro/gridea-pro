package repository

import (
	"context"
	"gridea-pro/backend/internal/domain"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type menuRepository struct {
	*BaseJSONRepository[domain.Menu]
}

func NewMenuRepository(appDir string) domain.MenuRepository {
	base := NewBaseJSONRepository[domain.Menu](appDir, "menus.json", "menus")
	return &menuRepository{base}
}

func ensureMenuID(menu *domain.Menu) {
	if menu.ID == "" {
		const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		id, _ := gonanoid.Generate(alphabet, 6)
		menu.ID = id
	}
}

func (r *menuRepository) Create(ctx context.Context, menu *domain.Menu) error {
	ensureMenuID(menu)
	return r.Add(ctx, *menu)
}

func (r *menuRepository) Update(ctx context.Context, id string, menu *domain.Menu) error {
	return r.BaseJSONRepository.Update(ctx, id, *menu)
}

func (r *menuRepository) GetByID(ctx context.Context, id string) (*domain.Menu, error) {
	menu, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (r *menuRepository) SaveAll(ctx context.Context, menus []domain.Menu) error {
	for i := range menus {
		ensureMenuID(&menus[i])
	}
	return r.BaseJSONRepository.SaveAll(ctx, menus)
}
