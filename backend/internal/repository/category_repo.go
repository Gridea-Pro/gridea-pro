package repository

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"path/filepath"
	"sync"
)

type categoryRepository struct {
	mu     sync.RWMutex
	appDir string
}

func NewCategoryRepository(appDir string) domain.CategoryRepository {
	return &categoryRepository{appDir: appDir}
}

func (r *categoryRepository) GetAll(ctx context.Context) ([]domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dbPath := filepath.Join(r.appDir, "config", "categories.json")
	var db struct {
		Categories []domain.Category `json:"categories"`
	}

	if err := LoadJSONFile(dbPath, &db); err != nil {
		// Log error but returns empty on first run for compatibility
		if filepath.Base(dbPath) == "categories.json" {
			// Try to handle missing file gracefully
			return []domain.Category{}, nil
		}
		return nil, fmt.Errorf("failed to load categories: %w", err)
	}

	return db.Categories, nil
}

func (r *categoryRepository) SaveAll(ctx context.Context, categories []domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	dbPath := filepath.Join(r.appDir, "config", "categories.json")
	db := map[string]interface{}{"categories": categories}
	return SaveJSONFile(dbPath, db)
}
