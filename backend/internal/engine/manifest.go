package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ManifestFileName 记录渲染产物清单的文件名，放在 appDir 根目录下
// 放在 appDir 根目录（而非 output 内）是为了避免被部署到服务器
const ManifestFileName = ".render-manifest.json"

// RenderManifest 跟踪本次渲染生成的文件。渲染结束时计算"上次有、本次没"的
// 孤儿文件并删除，既保证已删除文章的 HTML 不会残留（解决 issue #26），
// 也不会误删用户放在 output 目录里的自定义文件（CNAME、ads.txt 等）。
type RenderManifest struct {
	mu       sync.Mutex
	files    map[string]struct{}
	buildDir string
}

// manifestPayload 是持久化到磁盘的结构
type manifestPayload struct {
	Version     int       `json:"version"`
	GeneratedAt time.Time `json:"generatedAt"`
	Files       []string  `json:"files"`
}

const manifestVersion = 1

// NewRenderManifest 创建一个新的 manifest，buildDir 用于路径相对化
func NewRenderManifest(buildDir string) *RenderManifest {
	return &RenderManifest{
		files:    make(map[string]struct{}),
		buildDir: filepath.Clean(buildDir),
	}
}

// Track 记录一个已写入的绝对路径。buildDir 外的路径会被静默忽略
// 并发安全，可由多个 goroutine 同时调用
func (m *RenderManifest) Track(absPath string) {
	if m == nil {
		return
	}
	rel, ok := m.relPath(absPath)
	if !ok {
		return
	}
	m.mu.Lock()
	m.files[rel] = struct{}{}
	m.mu.Unlock()
}

// relPath 返回规范化的相对路径（使用 / 分隔，跨平台可读）
// 路径不在 buildDir 内时返回 false
func (m *RenderManifest) relPath(absPath string) (string, bool) {
	abs, err := filepath.Abs(absPath)
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(m.buildDir, abs)
	if err != nil {
		return "", false
	}
	if rel == "." || strings.HasPrefix(rel, "..") {
		return "", false
	}
	return filepath.ToSlash(rel), true
}

// WriteFile 等价于 os.WriteFile，但写入成功后会记录路径
func (m *RenderManifest) WriteFile(path string, data []byte, perm os.FileMode) error {
	if err := os.WriteFile(path, data, perm); err != nil {
		return err
	}
	m.Track(path)
	return nil
}

// CopyFile 拷贝单个文件并记录目标路径
func (m *RenderManifest) CopyFile(src, dst string) error {
	if err := copyFile(src, dst); err != nil {
		return err
	}
	m.Track(dst)
	return nil
}

// CopyDir 递归拷贝目录，跟踪所有被拷贝的文件
func (m *RenderManifest) CopyDir(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("源路径不是目录: %s", src)
	}

	if err := os.MkdirAll(dst, si.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := m.CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := m.CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// Files 返回本次记录的相对路径集合（返回拷贝，调用方可自由修改）
func (m *RenderManifest) Files() map[string]struct{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]struct{}, len(m.files))
	for k := range m.files {
		out[k] = struct{}{}
	}
	return out
}

// Save 将 manifest 持久化到 appDir/.render-manifest.json
func (m *RenderManifest) Save(appDir string) error {
	m.mu.Lock()
	files := make([]string, 0, len(m.files))
	for f := range m.files {
		files = append(files, f)
	}
	m.mu.Unlock()
	sort.Strings(files)

	payload := manifestPayload{
		Version:     manifestVersion,
		GeneratedAt: time.Now().UTC(),
		Files:       files,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(appDir, ManifestFileName), data, 0644)
}

// LoadPreviousManifest 读取 appDir 下的上次清单
// 若不存在或解析失败返回 nil, nil —— 调用方据此判断是否为首次渲染
func LoadPreviousManifest(appDir string) (map[string]struct{}, error) {
	data, err := os.ReadFile(filepath.Join(appDir, ManifestFileName))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var payload manifestPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("解析 manifest 失败: %w", err)
	}
	out := make(map[string]struct{}, len(payload.Files))
	for _, f := range payload.Files {
		out[f] = struct{}{}
	}
	return out, nil
}

// CleanOrphans 删除"上次有、本次没"的文件。清理完成后
// 删除过程中变空的目录（从深到浅）
// 仅操作 buildDir 内的文件，永远不触碰 buildDir 外的路径
func CleanOrphans(buildDir string, previous map[string]struct{}, current *RenderManifest) int {
	buildDir = filepath.Clean(buildDir)
	currentFiles := current.Files()

	var orphans []string
	for f := range previous {
		if _, kept := currentFiles[f]; !kept {
			orphans = append(orphans, f)
		}
	}
	if len(orphans) == 0 {
		return 0
	}

	removed := 0
	dirsToCheck := make(map[string]struct{})
	for _, rel := range orphans {
		abs := filepath.Join(buildDir, filepath.FromSlash(rel))
		// 防御：绝对路径必须仍在 buildDir 内
		absClean, err := filepath.Abs(abs)
		if err != nil {
			continue
		}
		if !strings.HasPrefix(absClean, buildDir+string(filepath.Separator)) {
			continue
		}
		if err := os.Remove(absClean); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				continue
			}
		} else {
			removed++
		}
		// 记录需检查是否变空的父目录（不包括 buildDir 本身）
		dir := filepath.Dir(absClean)
		for strings.HasPrefix(dir, buildDir+string(filepath.Separator)) {
			dirsToCheck[dir] = struct{}{}
			dir = filepath.Dir(dir)
		}
	}

	// 按深度从深到浅删除空目录。非空目录 os.Remove 会失败，静默忽略
	dirs := make([]string, 0, len(dirsToCheck))
	for d := range dirsToCheck {
		dirs = append(dirs, d)
	}
	sort.Slice(dirs, func(i, j int) bool {
		return strings.Count(dirs[i], string(filepath.Separator)) >
			strings.Count(dirs[j], string(filepath.Separator))
	})
	for _, d := range dirs {
		_ = os.Remove(d)
	}

	return removed
}
