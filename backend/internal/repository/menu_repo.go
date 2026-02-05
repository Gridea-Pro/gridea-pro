package repository

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"path/filepath"
	"sync"
)

type menuRepository struct {
	mu     sync.RWMutex
	appDir string
}

func NewMenuRepository(appDir string) domain.MenuRepository {
	return &menuRepository{appDir: appDir}
}

func (r *menuRepository) GetAll(ctx context.Context) ([]domain.Menu, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dbPath := filepath.Join(r.appDir, "config", "menus.json")
	var db struct {
		Menus []domain.Menu `json:"menus"`
	}
	if err := LoadJSONFile(dbPath, &db); err != nil {
		return []domain.Menu{}, nil
	}
	return db.Menus, nil
}

func (r *menuRepository) SaveAll(ctx context.Context, menus []domain.Menu) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	dbPath := filepath.Join(r.appDir, "config", "menus.json")
	db := map[string]interface{}{"menus": menus}
	return SaveJSONFile(dbPath, db)
}
