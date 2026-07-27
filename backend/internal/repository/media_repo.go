package repository

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"net/http"
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
	for i, file := range files {
		ext := filepath.Ext(file.Name)
		// Use UnixNano and index to ensure uniqueness even in same batch
		newName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), i, ext)
		newPath := filepath.Join(postImageDir, newName)

		if err := CopyFile(file.Path, newPath); err != nil {
			continue
		}
		// Return relative path for frontend usage
		results = append(results, "/post-images/"+newName)
	}

	return results, nil
}

// SaveImageBytes 保存内存中的图片字节（粘贴 / 拖拽），返回相对路径 /post-images/xxx
func (r *mediaRepository) SaveImageBytes(ctx context.Context, name string, data []byte) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(data) == 0 {
		return "", fmt.Errorf("空图片数据")
	}

	// 按真实内容判定类型，忽略用户可控的扩展名，杜绝 .svg/.html 等非位图内容落盘
	ext, ok := map[string]string{
		"image/png":  ".png",
		"image/jpeg": ".jpg",
		"image/gif":  ".gif",
		"image/webp": ".webp",
	}[http.DetectContentType(data)]
	if !ok {
		return "", fmt.Errorf("不支持的图片类型: %s", http.DetectContentType(data))
	}

	postImageDir := filepath.Join(r.appDir, "post-images")
	if err := os.MkdirAll(postImageDir, 0755); err != nil {
		return "", err
	}

	newName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	newPath := filepath.Join(postImageDir, newName)
	if err := os.WriteFile(newPath, data, 0644); err != nil {
		return "", err
	}
	return "/post-images/" + newName, nil
}
