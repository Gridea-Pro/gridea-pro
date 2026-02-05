package repository

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"os"
	"path/filepath"
	"sync"
)

type themeRepository struct {
	mu     sync.RWMutex
	appDir string
}

func NewThemeRepository(appDir string) domain.ThemeRepository {
	return &themeRepository{appDir: appDir}
}

func (r *themeRepository) GetAll(ctx context.Context) ([]domain.Theme, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	themesDir := filepath.Join(r.appDir, "themes")
	entries, err := os.ReadDir(themesDir)
	if err != nil {
		return []domain.Theme{}, nil
	}

	var themes []domain.Theme
	for _, entry := range entries {
		if entry.IsDir() {
			themePath := filepath.Join(themesDir, entry.Name(), "config.json")
			var theme domain.Theme
			if err := LoadJSONFile(themePath, &theme); err == nil {
				theme.Folder = entry.Name()

				// Detect preview image
				// Check for preview.png, preview.jpg, preview.jpeg, preview.webp
				// Return relative path from theme root: assets/media/preview.xxx
				assetsDir := filepath.Join(themesDir, entry.Name(), "assets", "media")
				exts := []string{".png", ".jpg", ".jpeg", ".webp"}
				for _, ext := range exts {
					if _, err := os.Stat(filepath.Join(assetsDir, "preview"+ext)); err == nil {
						theme.PreviewImage = filepath.Join("assets", "media", "preview"+ext)
						break
					}
				}

				themes = append(themes, theme)
			}
		}
	}
	return themes, nil
}

func (r *themeRepository) GetConfig(ctx context.Context) (domain.ThemeConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	configPath := filepath.Join(r.appDir, "config", "config.json")
	var config domain.ThemeConfig

	// 如果文件不存在，返回默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return domain.ThemeConfig{
			ThemeName:        "default",
			PostPageSize:     10,
			ArchivesPageSize: 50,
			SiteName:         "My Site",
			SiteAuthor:       "Gridea",
			SiteDescription:  "Welcome to my site",
			FooterInfo:       "Powered by Gridea Pro",
			ShowFeatureImage: true,
			Domain:           "http://localhost",
			PostUrlFormat:    "SLUG",
			TagUrlFormat:     "SLUG",
			DateFormat:       "YYYY-MM-DD",
			FeedFullText:     true,
			FeedCount:        10,
			ArchivesPath:     "archives",
			PostPath:         "post",
			TagPath:          "tag",
			LinkPath:         "link",
		}, nil
	}

	if err := LoadJSONFile(configPath, &config); err != nil {
		return domain.ThemeConfig{}, err
	}
	return config, nil
}

func (r *themeRepository) SaveConfig(ctx context.Context, config domain.ThemeConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	configPath := filepath.Join(r.appDir, "config", "config.json")
	return SaveJSONFile(configPath, config)
}
