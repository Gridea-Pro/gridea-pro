package repository

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/utils"
	"path/filepath"
	"sync"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type memoRepository struct {
	mu     sync.RWMutex
	appDir string
	cache  []domain.Memo
	loaded bool
}

func NewMemoRepository(appDir string) domain.MemoRepository {
	return &memoRepository{
		appDir: appDir,
		cache:  make([]domain.Memo, 0),
		loaded: false,
	}
}

// Internal DTO for JSON serialization compatibility
type MemoDTO struct {
	ID        string      `json:"id"`
	Content   string      `json:"content"`
	Tags      []string    `json:"tags"`
	Images    []string    `json:"images"`
	CreatedAt interface{} `json:"createdAt"` // Can be string or int64
	UpdatedAt interface{} `json:"updatedAt"`
}

// forceLoad loads memos from disk into cache. Caller must hold Lock.
func (r *memoRepository) forceLoad() error {
	dbPath := filepath.Join(r.appDir, "config", "memos.json")
	var db struct {
		Memos []MemoDTO `json:"memos"`
	}

	if err := LoadJSONFile(dbPath, &db); err != nil {
		r.cache = []domain.Memo{}
		r.loaded = true
		return nil // Return nil on error (assuming empty or missing file)
	}

	memos := make([]domain.Memo, len(db.Memos))
	for i, m := range db.Memos {
		memos[i] = domain.Memo{
			ID:        m.ID,
			Content:   m.Content,
			Tags:      m.Tags,
			Images:    m.Images,
			CreatedAt: parseTime(m.CreatedAt),
			UpdatedAt: parseTime(m.UpdatedAt),
		}
	}
	r.cache = memos
	r.loaded = true
	return nil
}

// loadIfNeeded ensures cache is populated.
func (r *memoRepository) loadIfNeeded() error {
	r.mu.RLock()
	if r.loaded {
		r.mu.RUnlock()
		return nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.loaded {
		return nil
	}
	return r.forceLoad()
}

// saveCache writes the current cache to disk. Caller must hold Lock.
func (r *memoRepository) saveCache() error {
	dbPath := filepath.Join(r.appDir, "config", "memos.json")

	dtos := make([]MemoDTO, len(r.cache))
	for i, m := range r.cache {
		dtos[i] = MemoDTO{
			ID:        m.ID,
			Content:   m.Content,
			Tags:      m.Tags,
			Images:    m.Images,
			CreatedAt: m.CreatedAt.Format(time.RFC3339),
			UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
		}
	}

	db := map[string]interface{}{
		"memos": dtos,
	}
	return SaveJSONFile(dbPath, db)
}

func parseTime(v interface{}) time.Time {
	switch t := v.(type) {
	case string:
		if t == "" {
			return time.Now()
		}
		if parsed, err := utils.ParseTime(t, time.Local); err == nil {
			return parsed
		}
		if t2, err := time.ParseInLocation(domain.TimeLayout, t, time.Local); err == nil {
			return t2
		}
		return time.Now()
	case float64:
		return time.UnixMilli(int64(t))
	case int64:
		return time.UnixMilli(t)
	default:
		return time.Now()
	}
}

func ensureMemoID(memo *domain.Memo) {
	if memo.ID == "" {
		const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		id, _ := gonanoid.Generate(alphabet, 6)
		memo.ID = id
	}
}

func (r *memoRepository) SaveAll(ctx context.Context, memos []domain.Memo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range memos {
		ensureMemoID(&memos[i])
	}

	originalCache := r.cache
	originalLoaded := r.loaded

	r.cache = memos
	r.loaded = true

	if err := r.saveCache(); err != nil {
		r.cache = originalCache // Revert
		r.loaded = originalLoaded
		return err
	}
	return nil
}

func (r *memoRepository) Create(ctx context.Context, memo *domain.Memo) error {
	ensureMemoID(memo)

	if err := r.loadIfNeeded(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 1. Prepare new data
	newCache := make([]domain.Memo, len(r.cache)+1)
	copy(newCache, r.cache)
	newCache[len(r.cache)] = *memo

	// 2. Save to disk
	originalCache := r.cache
	r.cache = newCache
	if err := r.saveCache(); err != nil {
		r.cache = originalCache // Revert
		return err
	}

	// 3. Cache updated
	return nil
}

func (r *memoRepository) Update(ctx context.Context, id string, memo *domain.Memo) error {
	if err := r.loadIfNeeded(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	idx := -1
	for i, m := range r.cache {
		if m.ID == id {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("memo not found")
	}

	// 1. Prepare new data
	newCache := make([]domain.Memo, len(r.cache))
	copy(newCache, r.cache)
	newCache[idx] = *memo

	// 2. Save to disk
	originalCache := r.cache
	r.cache = newCache
	if err := r.saveCache(); err != nil {
		r.cache = originalCache // Revert
		return err
	}

	// 3. Cache updated
	return nil
}

func (r *memoRepository) Delete(ctx context.Context, id string) error {
	if err := r.loadIfNeeded(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	idx := -1
	for i, m := range r.cache {
		if m.ID == id {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("memo not found")
	}

	// 1. Prepare new data
	newCache := make([]domain.Memo, 0, len(r.cache)-1)
	newCache = append(newCache, r.cache[:idx]...)
	newCache = append(newCache, r.cache[idx+1:]...)

	// 2. Save to disk
	originalCache := r.cache
	r.cache = newCache
	if err := r.saveCache(); err != nil {
		r.cache = originalCache // Revert
		return err
	}

	// 3. Cache updated
	return nil
}

func (r *memoRepository) GetByID(ctx context.Context, id string) (*domain.Memo, error) {
	if err := r.loadIfNeeded(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, m := range r.cache {
		if m.ID == id {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("memo not found")
}

func (r *memoRepository) List(ctx context.Context) ([]domain.Memo, error) {
	if err := r.loadIfNeeded(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]domain.Memo, len(r.cache))
	copy(result, r.cache)
	return result, nil
}
