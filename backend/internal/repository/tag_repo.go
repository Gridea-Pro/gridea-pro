package repository

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"path/filepath"
	"sync"
)

type tagRepository struct {
	mu     sync.RWMutex
	appDir string
}

func NewTagRepository(appDir string) domain.TagRepository {
	return &tagRepository{appDir: appDir}
}

func (r *tagRepository) GetAll(ctx context.Context) ([]domain.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dbPath := filepath.Join(r.appDir, "config", "tags.json")
	var db struct {
		Tags []domain.Tag `json:"tags"`
	}
	if err := LoadJSONFile(dbPath, &db); err != nil {
		return []domain.Tag{}, nil
	}
	return db.Tags, nil
}

func (r *tagRepository) SaveAll(ctx context.Context, tags []domain.Tag) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	dbPath := filepath.Join(r.appDir, "config", "tags.json")
	db := map[string]interface{}{"tags": tags}
	return SaveJSONFile(dbPath, db)
}
