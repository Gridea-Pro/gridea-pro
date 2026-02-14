package service

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type ScaffoldService struct {
	assets embed.FS
	mu     sync.Mutex
}

func NewScaffoldService(assets embed.FS) *ScaffoldService {
	return &ScaffoldService{
		assets: assets,
	}
}

// InitSite checks if the site is initialized, if not, it copies default files
func (s *ScaffoldService) InitSite(appDir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Locate default-files source path in embed.FS
	// Source path in embed.FS: frontend/dist/default-files or frontend/public/default-files
	srcParams := []string{"frontend/dist/default-files", "frontend/public/default-files"}
	var srcPath string

	for _, p := range srcParams {
		if _, err := fs.Stat(s.assets, p); err == nil {
			srcPath = p
			break
		}
	}

	if srcPath == "" {
		return fmt.Errorf("default-files not found in assets")
	}

	// 2. Recursively copy all default files to appDir
	// This will create directories (themes, posts, images, etc.) and copy files.
	// Existing files will be skipped.
	if err := s.copyDirFromEmbed(srcPath, appDir); err != nil {
		return fmt.Errorf("failed to copy default files: %w", err)
	}

	// 3. Create directories that might not be in default-files (e.g. output)
	// Ensure essential directories exist just in case
	ensureDirs := []string{
		filepath.Join(appDir, "output"),
	}
	for _, dir := range ensureDirs {
		_ = os.MkdirAll(dir, 0755)
	}

	// 4. Patch config.json with dynamic sourceFolder
	configPath := filepath.Join(appDir, "config", "config.json")
	// Only patch if file exists (it should, copied from defaults)
	if content, err := os.ReadFile(configPath); err == nil {
		var config map[string]interface{}
		if err := json.Unmarshal(content, &config); err == nil {
			// Update sourceFolder to actual appDir
			config["sourceFolder"] = appDir

			// Write back
			if data, err := json.MarshalIndent(config, "", "  "); err == nil {
				_ = os.WriteFile(configPath, data, 0644)
			}
		}
	}

	return nil
}

func (s *ScaffoldService) copyDirFromEmbed(src string, dst string) error {
	entries, err := s.assets.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := s.copyDirFromEmbed(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := s.copyFileFromEmbed(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ScaffoldService) copyFileFromEmbed(src string, dst string) error {
	// Check if destination file exists
	if _, err := os.Stat(dst); err == nil {
		// File exists, skip
		return nil
	}

	sourceFile, err := s.assets.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
