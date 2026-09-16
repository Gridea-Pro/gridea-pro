package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

type ScaffoldService struct {
	// fs.FS 而不是 embed.FS：运行时传入的就是 embed.FS，放宽成接口只是为了
	// 让测试能注入内存文件系统，验证「补齐缺失文件但不覆盖已有文件」的行为。
	assets fs.FS
	mu     sync.Mutex
}

func NewScaffoldService(assets fs.FS) *ScaffoldService {
	return &ScaffoldService{
		assets: assets,
	}
}

// scaffoldManifest records what was copied during initialization.
// Used to distinguish "user deleted this" from "new content in a new version".
type scaffoldManifest struct {
	Version string   `json:"version"`
	Posts   []string `json:"posts"`
	Themes  []string `json:"themes"`
	Memos   bool     `json:"memos"`
}

func loadManifest(manifestPath string) (*scaffoldManifest, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	var m scaffoldManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func saveManifest(manifestPath string, m *scaffoldManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(manifestPath, data, 0644)
}

// InitSite ensures the site directory is properly initialized.
//
// Uses a .scaffold.json manifest to track what was copied:
//   - Posts & memos: only copied on first init, never again (user deletions are respected)
//   - Themes: copied on first init; new themes from new app versions are auto-added,
//     but themes the user previously deleted won't come back. Themes still present on
//     disk also get missing files topped up, so template files added by a new app
//     version reach existing sites — existing files are never overwritten.
//   - Config & static files: always checked, re-created if missing (skip if exists)
//   - output/ directory & config.json sourceFolder: patched on every startup
func (s *ScaffoldService) InitSite(appDir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Locate default-files source in embed.FS
	srcParams := []string{"frontend/dist/default-files", "frontend/public/default-files"}
	var srcRoot string
	for _, p := range srcParams {
		if _, err := fs.Stat(s.assets, p); err == nil {
			srcRoot = p
			break
		}
	}
	if srcRoot == "" {
		return fmt.Errorf("default-files not found in assets")
	}

	// 2. Load or create manifest
	manifestPath := filepath.Join(appDir, ".scaffold.json")
	manifest, err := loadManifest(manifestPath)
	isFirstInit := err != nil // file doesn't exist or is invalid

	if isFirstInit {
		manifest = &scaffoldManifest{}
	}

	// 3. Scaffold posts (first init only)
	if len(manifest.Posts) == 0 {
		postsSrc := path.Join(srcRoot, "posts")
		postsDst := filepath.Join(appDir, "posts")
		if entries, err := fs.ReadDir(s.assets, postsSrc); err == nil {
			_ = os.MkdirAll(postsDst, 0755)
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				src := path.Join(postsSrc, entry.Name())
				dst := filepath.Join(postsDst, entry.Name())
				_ = s.copyFileFromEmbed(src, dst)
				manifest.Posts = append(manifest.Posts, entry.Name())
			}
		}
		// Fill date placeholders
		s.fillPostDates(appDir)
	}

	// 4. Scaffold memos (first init only)
	if !manifest.Memos {
		memosSrc := path.Join(srcRoot, "config", "memos.json")
		memosDst := filepath.Join(appDir, "config", "memos.json")
		_ = os.MkdirAll(filepath.Join(appDir, "config"), 0755)
		_ = s.copyFileFromEmbed(memosSrc, memosDst)
		s.fillMemoDates(appDir)
		manifest.Memos = true
	}

	// 5. Scaffold themes (track individually — new themes get copied, deleted themes stay deleted)
	knownThemes := make(map[string]bool)
	for _, t := range manifest.Themes {
		knownThemes[t] = true
	}
	themesSrc := path.Join(srcRoot, "themes")
	if entries, err := fs.ReadDir(s.assets, themesSrc); err == nil {
		themesDst := filepath.Join(appDir, "themes")
		_ = os.MkdirAll(themesDst, 0755)
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			themeName := entry.Name()
			src := path.Join(themesSrc, themeName)
			dst := filepath.Join(themesDst, themeName)

			if knownThemes[themeName] {
				// Theme directory gone entirely — user deleted it on purpose, don't resurrect it.
				if _, err := os.Stat(dst); err != nil {
					continue
				}
				// Theme still in use: top up files this app version added (e.g. new
				// category templates). copyFileFromEmbed skips paths that already
				// exist, so the user's own edits are never overwritten.
				_ = s.copyDirFromEmbed(src, dst)
				continue
			}

			// New theme (not in manifest) — copy it
			_ = s.copyDirFromEmbed(src, dst)
			manifest.Themes = append(manifest.Themes, themeName)
		}
	}

	// 6. Scaffold config files, static files, images (always, skip if exists)
	// These are essential for app functionality — re-created if user accidentally deleted them
	alwaysCopyDirs := []string{"config", "static", "images", "post-images"}
	for _, dir := range alwaysCopyDirs {
		src := path.Join(srcRoot, dir)
		dst := filepath.Join(appDir, dir)
		if _, err := fs.ReadDir(s.assets, src); err == nil {
			_ = s.copyDirFromEmbed(src, dst)
		}
	}
	// Also copy root-level files (favicon.ico etc.)
	if entries, err := fs.ReadDir(s.assets, srcRoot); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			src := path.Join(srcRoot, entry.Name())
			dst := filepath.Join(appDir, entry.Name())
			_ = s.copyFileFromEmbed(src, dst)
		}
	}

	// 7. Save manifest
	_ = saveManifest(manifestPath, manifest)

	// 8. Always ensure output directory exists
	_ = os.MkdirAll(filepath.Join(appDir, "output"), 0755)

	// 9. Always patch config.json with current sourceFolder
	configPath := filepath.Join(appDir, "config", "config.json")
	if content, err := os.ReadFile(configPath); err == nil {
		var config map[string]interface{}
		if err := json.Unmarshal(content, &config); err == nil {
			config["sourceFolder"] = appDir
			if data, err := json.MarshalIndent(config, "", "  "); err == nil {
				_ = os.WriteFile(configPath, data, 0644)
			}
		}
	}

	// 10. Clean up legacy marker file
	_ = os.Remove(filepath.Join(appDir, ".initialized"))

	return nil
}

// fillPostDates replaces __INIT_DATE_XX__ placeholders in post frontmatter with current time.
// XX controls sort order: 00 is newest, each increment adds a 1-minute offset into the past.
func (s *ScaffoldService) fillPostDates(appDir string) {
	now := time.Now()
	postsDir := filepath.Join(appDir, "posts")
	entries, err := os.ReadDir(postsDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		postPath := filepath.Join(postsDir, entry.Name())
		data, err := os.ReadFile(postPath)
		if err != nil || !bytes.Contains(data, []byte("__INIT_DATE_")) {
			continue
		}
		newData := data
		for i := 0; i < 100; i++ {
			placeholder := fmt.Sprintf("__INIT_DATE_%02d__", i)
			if bytes.Contains(newData, []byte(placeholder)) {
				postTime := now.Add(-time.Duration(i) * time.Minute)
				dateStr := postTime.Format("2006-01-02 15:04:05")
				newData = bytes.Replace(newData, []byte(placeholder), []byte(dateStr), 1)
				break
			}
		}
		if !bytes.Equal(data, newData) {
			_ = os.WriteFile(postPath, newData, 0644)
		}
	}
}

// fillMemoDates replaces empty createdAt/updatedAt strings in memos.json with current time.
func (s *ScaffoldService) fillMemoDates(appDir string) {
	memosPath := filepath.Join(appDir, "config", "memos.json")
	data, err := os.ReadFile(memosPath)
	if err != nil || !bytes.Contains(data, []byte(`"createdAt": ""`)) {
		return
	}
	memoTime := time.Now().Format(time.RFC3339)
	data = bytes.ReplaceAll(data, []byte(`"createdAt": ""`), []byte(`"createdAt": "`+memoTime+`"`))
	data = bytes.ReplaceAll(data, []byte(`"updatedAt": ""`), []byte(`"updatedAt": "`+memoTime+`"`))
	_ = os.WriteFile(memosPath, data, 0644)
}

func (s *ScaffoldService) copyDirFromEmbed(src string, dst string) error {
	entries, err := fs.ReadDir(s.assets, src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		// embed.FS always uses forward slashes, even on Windows
		srcPath := path.Join(src, entry.Name())
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
