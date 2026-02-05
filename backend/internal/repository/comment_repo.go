package repository

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"path/filepath"
	"sync"
)

type commentRepository struct {
	mu     sync.RWMutex
	appDir string
}

// NewCommentRepository 创建评论仓库
func NewCommentRepository(appDir string) domain.CommentRepository {
	return &commentRepository{appDir: appDir}
}

// GetSettings 获取评论设置
func (r *commentRepository) GetSettings(ctx context.Context) (domain.CommentSettings, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	settingsPath := filepath.Join(r.appDir, "config", "comment.json")
	var settings domain.CommentSettings
	if err := LoadJSONFile(settingsPath, &settings); err != nil {
		// 返回默认设置
		return domain.CommentSettings{
			Enable:   false,
			Platform: domain.CommentPlatformValine,
		}, nil
	}
	return settings, nil
}

// SaveSettings 保存评论设置
func (r *commentRepository) SaveSettings(ctx context.Context, settings domain.CommentSettings) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	settingsPath := filepath.Join(r.appDir, "config", "comment.json")
	return SaveJSONFile(settingsPath, settings)
}
