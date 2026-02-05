package repository

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"path/filepath"
	"sync"
)

type linkRepository struct {
	mu     sync.RWMutex
	appDir string
}

func NewLinkRepository(appDir string) domain.LinkRepository {
	return &linkRepository{appDir: appDir}
}

func (r *linkRepository) GetAll(ctx context.Context) ([]domain.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dbPath := filepath.Join(r.appDir, "config", "links.json")
	var db struct {
		Links []domain.Link `json:"links"`
	}
	if err := LoadJSONFile(dbPath, &db); err != nil {
		return []domain.Link{}, nil
	}
	return db.Links, nil
}

func (r *linkRepository) SaveAll(ctx context.Context, links []domain.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	dbPath := filepath.Join(r.appDir, "config", "links.json")
	db := map[string]interface{}{"links": links}
	return SaveJSONFile(dbPath, db)
}
