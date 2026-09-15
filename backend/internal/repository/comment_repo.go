package repository

import (
	"context"
	"errors"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"io/fs"
	"log"
	"path/filepath"
	"sync"
)

type commentRepository struct {
	mu     sync.RWMutex
	appDir string
	cache  *domain.CommentSettings
	loaded bool
}

func NewCommentRepository(appDir string) domain.CommentRepository {
	return &commentRepository{
		appDir: appDir,
		cache:  nil,
		loaded: false,
	}
}

func (r *commentRepository) loadIfNeeded() error {
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

	dbPath := filepath.Join(r.appDir, "config", "comment.json")
	var settings domain.CommentSettings

	if err := LoadJSONFile(dbPath, &settings); err != nil {
		// 文件不存在：合法初始状态，锁进 cache。
		// （原判断 filepath.Base(dbPath)=="comment.json" 恒真，等于吞掉一切错误，是 bug。）
		if errors.Is(err, fs.ErrNotExist) {
			r.cache = &domain.CommentSettings{}
			r.loaded = true
			return nil
		}
		// 其他错误（JSON 损坏/权限/IO）：不能把空配置锁进 cache，
		// 否则用户保存时会用空评论平台配置覆盖磁盘真实配置——必须向上抛，让下次访问重试。
		log.Printf("[repo] commentRepo load %s failed: %v", dbPath, err)
		return fmt.Errorf("load comment settings: %w", err)
	}

	r.cache = &settings
	r.loaded = true
	return nil
}

func (r *commentRepository) GetSettings(ctx context.Context) (*domain.CommentSettings, error) {
	if err := r.loadIfNeeded(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.cache == nil {
		return &domain.CommentSettings{}, nil
	}
	// Return copy? Or pointer to cache?
	// If caller modifies it, cache is modified.
	// Safe to return copy.
	copy := *r.cache
	return &copy, nil
}

func (r *commentRepository) SaveSettings(ctx context.Context, settings *domain.CommentSettings) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	dbPath := filepath.Join(r.appDir, "config", "comment.json")
	if err := SaveJSONFile(dbPath, settings); err != nil {
		return err
	}

	r.cache = settings
	r.loaded = true
	return nil
}
