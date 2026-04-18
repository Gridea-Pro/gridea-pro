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

			baseName := filepath.Base(event.Name)

			// 1. 忽略 OS 垃圾文件
			if baseName == ".DS_Store" {
				continue
			}

			// 2. 忽略原子写临时文件（模式：*.tmp.*，覆盖所有 WriteFileAtomic 生成的临时文件）
			if strings.Contains(baseName, ".tmp.") {
				continue
			}

			// Determine event type based on path
			isThemeChange := strings.Contains(event.Name, filepath.Join(w.appDir, "themes"))

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
				evtName := event.Name
				themeDebounceTimer = time.AfterFunc(debounceDuration, func() {
					log.Printf("ResourceWatcher: theme change → emit app-site-reload (trigger: %s)", evtName)
					runtime.EventsEmit(w.ctx, "theme-list-updated")
					runtime.EventsEmit(w.ctx, "app-site-reload")
				})
			} else {
				// 3. Data 目录只接受外部工具放进来的 .md 文章变更。
				//    所有 config/*.json 都由 app 自己通过原子写管理（包括 config.json
				//    里的站点/主题配置），前端保存时会显式 emit app-site-reload，
				//    watcher 不应该再重复触发，否则会形成"保存 → watcher → 再次渲染"
				//    的叠加。用户若手动编辑 config/*.json，需要重启 app 才能生效
				//    —— 这是罕见场景，可接受的折中。
				if !strings.HasSuffix(baseName, ".md") {
					continue
				}

				if dataDebounceTimer != nil {
					dataDebounceTimer.Stop()
				}
				evtName := event.Name
				dataDebounceTimer = time.AfterFunc(debounceDuration, func() {
					log.Printf("ResourceWatcher: data change → emit app-site-reload (trigger: %s)", evtName)
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
