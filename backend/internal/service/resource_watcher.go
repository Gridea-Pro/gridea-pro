package service

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ResourceWatcher struct {
	watcher *fsnotify.Watcher
	ctx     context.Context
	appDir  string
	stop    chan struct{}
	mu      sync.Mutex
}

func NewResourceWatcher(appDir string) (*ResourceWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &ResourceWatcher{
		watcher: watcher,
		appDir:  appDir,
		stop:    make(chan struct{}),
	}, nil
}

func (w *ResourceWatcher) Start(ctx context.Context) {
	w.ctx = ctx

	// 1. Watch Data Directories (Flat)
	dataPaths := []string{
		filepath.Join(w.appDir, "posts"),
		filepath.Join(w.appDir, "config"),
	}
	for _, p := range dataPaths {
		if err := w.watcher.Add(p); err != nil {
			// Log but don't fail, directory might not exist yet
			log.Printf("ResourceWatcher: Failed to watch data path %s: %v", p, err)
		}
	}

	// 2. Watch Themes Directory (Recursive)
	themesDir := filepath.Join(w.appDir, "themes")
	if err := w.watcher.Add(themesDir); err != nil {
		log.Printf("ResourceWatcher: Failed to watch themes root %s: %v", themesDir, err)
	}

	filepath.WalkDir(themesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			w.watcher.Add(path)
		}
		return nil
	})

	go w.watchLoop()
}

func (w *ResourceWatcher) watchLoop() {
	var themeDebounceTimer *time.Timer
	var dataDebounceTimer *time.Timer
	const debounceDuration = 100 * time.Millisecond

	for {
		select {
		case <-w.stop:
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Ignore temporary files and generated cache files
			baseName := filepath.Base(event.Name)
			if baseName == ".DS_Store" || baseName == "posts.json" || baseName == "tags.json" {
				continue
			}

			// Determine event type based on path
			isThemeChange := strings.Contains(event.Name, filepath.Join(w.appDir, "themes"))

			// Calculate relative path for logging/debugging
			// relPath, _ := filepath.Rel(w.appDir, event.Name)

			if isThemeChange {
				// Theme Logic:
				// Handle new directories being created (recursive watch)
				if event.Op&fsnotify.Create == fsnotify.Create {
					if isResourceDir(event.Name) {
						w.watcher.Add(event.Name)
					}
				}

				// Trigger Theme Update
				if themeDebounceTimer != nil {
					themeDebounceTimer.Stop()
				}
				themeDebounceTimer = time.AfterFunc(debounceDuration, func() {
					log.Println("ResourceWatcher: Theme change detected, refreshing...", event.Name)
					runtime.EventsEmit(w.ctx, "theme-list-updated")
					// Also reload site to ensure backend state is fresh (e.g. config.json in theme)
					runtime.EventsEmit(w.ctx, "app-site-reload")
				})

			} else {
				// Data Logic (Posts/Config):
				// Trigger Data Update
				if dataDebounceTimer != nil {
					dataDebounceTimer.Stop()
				}
				dataDebounceTimer = time.AfterFunc(debounceDuration, func() {
					log.Println("ResourceWatcher: Data change detected, reloading site...", event.Name)
					runtime.EventsEmit(w.ctx, "app-site-reload")
				})
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Println("ResourceWatcher error:", err)
		}
	}
}

func (w *ResourceWatcher) Close() {
	close(w.stop)
	w.watcher.Close()
}

// isDir checks if a path is a directory
// Note: duplicated from theme_watcher.go/data_watcher.go but since those are being deleted, we keep it here.
func isResourceDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
