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

type ThemeWatcher struct {
	watcher *fsnotify.Watcher
	ctx     context.Context
	appDir  string
	stop    chan struct{}
	mu      sync.Mutex
	timers  map[string]*time.Timer
}

func NewThemeWatcher(appDir string) (*ThemeWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &ThemeWatcher{
		watcher: watcher,
		appDir:  appDir,
		stop:    make(chan struct{}),
		timers:  make(map[string]*time.Timer),
	}, nil
}

func (w *ThemeWatcher) Start(ctx context.Context) {
	w.ctx = ctx
	themesDir := filepath.Join(w.appDir, "themes")

	// Add themes root directory
	if err := w.watcher.Add(themesDir); err != nil {
		// Just log error, don't panic, directory might not exist yet
		log.Printf("Failed to watch themes directory: %v", err)
	}

	// Recursively add subdirectories
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

func (w *ThemeWatcher) watchLoop() {
	var debounceTimer *time.Timer
	const debounceDuration = 300 * time.Millisecond

	for {
		select {
		case <-w.stop:
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Handle new directories being created
			if event.Op&fsnotify.Create == fsnotify.Create {
				if isDir(event.Name) {
					w.watcher.Add(event.Name)
				}
			}

			// We care about config.json changes or any directory changes (themes added/removed)
			shouldUpdate := false
			if strings.Contains(event.Name, "config.json") || isDir(event.Name) {
				shouldUpdate = true
			}

			// Additional check: maybe just reload on any change in themes dir for now,
			// but we want to avoid reloading on simple asset changes if possible?
			// So `themes/*` (create dir) or `themes/*/config.json` (write).

			// For simplicity, let's trigger on everything locally to be safe,
			// but debounce heavily.
			shouldUpdate = true

			if shouldUpdate {
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(debounceDuration, func() {
					runtime.EventsEmit(w.ctx, "theme-list-updated")
					// Also reload site to ensure backend state is fresh
					runtime.EventsEmit(w.ctx, "app-site-reload")
				})
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Println("watcher error:", err)
		}
	}
}

func (w *ThemeWatcher) Close() {
	close(w.stop)
	w.watcher.Close()
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
