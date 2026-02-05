package repository

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type mediaRepository struct {
	mu     sync.RWMutex
	appDir string
}

func NewMediaRepository(appDir string) domain.MediaRepository {
	return &mediaRepository{appDir: appDir}
}

func (r *mediaRepository) SaveImages(ctx context.Context, files []domain.UploadedFile) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	postImageDir := filepath.Join(r.appDir, "post-images")
	_ = os.MkdirAll(postImageDir, 0755)

	var results []string
	for _, file := range files {
		ext := filepath.Ext(file.Name)
		newName := fmt.Sprintf("%d%s", time.Now().Unix(), ext)
		newPath := filepath.Join(postImageDir, newName)

		if err := CopyFile(file.Path, newPath); err != nil {
			continue
		}
		// Return relative path or absolute? Original code returned what?
		// Original code: results = append(results, newPath) -> Absolute path.
		results = append(results, newPath)
	}

	return results, nil
}
