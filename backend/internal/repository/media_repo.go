package repository

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/webpconvert"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type mediaRepository struct {
	mu           sync.RWMutex
	appDir       string
	imageOptRepo domain.ImageOptimizeSettingRepository // 可为 nil（如 MCP 场景），nil 时 WebP 转换禁用
}

func NewMediaRepository(appDir string, imageOptRepo domain.ImageOptimizeSettingRepository) domain.MediaRepository {
	return &mediaRepository{appDir: appDir, imageOptRepo: imageOptRepo}
}

func (r *mediaRepository) SaveImages(ctx context.Context, files []domain.UploadedFile) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	postImageDir := filepath.Join(r.appDir, "post-images")
	_ = os.MkdirAll(postImageDir, 0755)

	// 检查 WebP 转换设置
	webpEnabled, webpQuality := GetWebpSetting(r.imageOptRepo, ctx)

	var results []string
	for i, file := range files {
		srcPath := file.Path
		ext := filepath.Ext(file.Name)

		// 如果启用了 WebP 转换且格式支持，先转换
		if webpEnabled && webpconvert.NeedsConversion(srcPath) {
			if tmpPath, err := webpconvert.ConvertToWebP(srcPath, webpQuality); err == nil && tmpPath != "" {
				srcPath = tmpPath
				ext = ".webp"
				defer os.Remove(tmpPath) // 函数返回时清理临时文件（非循环迭代时）
			}
		}

		// Use UnixNano and index to ensure uniqueness even in same batch
		newName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), i, ext)
		newPath := filepath.Join(postImageDir, newName)

		if err := CopyFile(srcPath, newPath); err != nil {
			continue
		}
		// Return relative path for frontend usage
		results = append(results, "/post-images/"+newName)
	}

	return results, nil
}
