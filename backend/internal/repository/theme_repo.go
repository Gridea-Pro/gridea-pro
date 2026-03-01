package repository

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"os"
	"path/filepath"
	"sync"
)

type themeRepository struct {
	mu           sync.RWMutex
	appDir       string
	configCache  *domain.ThemeConfig
	configLoaded bool
}

func NewThemeRepository(appDir string) domain.ThemeRepository {
	return &themeRepository{
		appDir:       appDir,
		configCache:  nil,
		configLoaded: false,
	}
}

func (r *themeRepository) GetAll(ctx context.Context) ([]domain.Theme, error) {
	// Functionally identical to previous implementation (no caching for theme list yet)
	// Could implement caching later if directory scanning becomes a bottleneck.

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

func (r *themeRepository) loadConfigIfNeeded() error {
	r.mu.RLock()
	if r.configLoaded {
		r.mu.RUnlock()
		return nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.configLoaded {
		return nil
	}

	configPath := filepath.Join(r.appDir, "config", "config.json")
	var config domain.ThemeConfig

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		r.configCache = &domain.ThemeConfig{
			ThemeName:        "default",
			PostPageSize:     10,
			ArchivesPageSize: 50,
			SiteName:         "My Site",
			SiteAuthor:       "Gridea",
			SiteDescription:  "Welcome to my site",
			FooterInfo:       "Powered by Gridea Pro",
			Domain:           "http://localhost",
			PostUrlFormat:    "SLUG",
			TagUrlFormat:     "SLUG",
			DateFormat:       "YYYY-MM-DD",
			FeedFullText:     true,
			FeedCount:        20,
			PostPath:         "post",
			TagPath:          "tag",
			LinkPath:         "link",
		}
		r.configLoaded = true
		return nil
	}

	if err := LoadJSONFile(configPath, &config); err != nil {
		return err
	}

	r.configCache = &config
	r.configLoaded = true
	return nil
}

func (r *themeRepository) GetConfig(ctx context.Context) (domain.ThemeConfig, error) {
	if err := r.loadConfigIfNeeded(); err != nil {
		return domain.ThemeConfig{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.configCache == nil {
		return domain.ThemeConfig{}, nil
	}
	return *r.configCache, nil
}

func (r *themeRepository) SaveConfig(ctx context.Context, config domain.ThemeConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	configPath := filepath.Join(r.appDir, "config", "config.json")
	if err := SaveJSONFile(configPath, config); err != nil {
		return err
	}

	r.configCache = &config
	r.configLoaded = true
	return nil
}
