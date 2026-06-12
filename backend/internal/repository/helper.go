package repository

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Helper functions for JSON DB

// LoadJSONFile 读取并解析 JSON 文件
func LoadJSONFile(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// SaveJSONFile saves data to a JSON file atomically.
// It writes to a temp file first, flushes to disk, and then renames to target.
func SaveJSONFile(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomic(path, data, 0644)
}

// SaveJSONFileIdempotent saves data to a JSON file atomically, but only if content changes.
func SaveJSONFileIdempotent(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	// Read existing file to compare
	if existingData, err := os.ReadFile(path); err == nil {
		if string(existingData) == string(data) {
			return nil // Content matches, skip write
		}
	}

	return WriteFileAtomic(path, data, 0644)
}

// WriteFileAtomic writes data to a file atomically by writing to a temp file and renaming.
func WriteFileAtomic(filename string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create temp file in the same directory to ensure atomic rename works (same FS)
	tmpFile, err := os.CreateTemp(dir, filepath.Base(filename)+".tmp.*")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName) // Clean up temp file if rename fails

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return err
	}

	// Ensure data is written to disk
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	// 在 rename 前打上 self-write 标记。rename 完成后 fsnotify 几乎立刻就会投递事件，
	// ResourceWatcher 据此过滤"app 保存 → watcher → 再次渲染"的自激循环。
	// 外部编辑器（不走本函数）的改动不会触发 WriteGate，watcher 仍会照常响应。
	DefaultWriteGate.MarkSelfWrite(filename)

	// Atomic rename
	return os.Rename(tmpName, filename)
}

func CopyFile(src, dst string) error {
	if sameFilePath(src, dst) {
		return nil
	}

	sourceFile, err := os.Open(src)
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

// sameFilePath 判断 src 和 dst 是否指向同一个文件。
// 分两层防御：先做路径规范化比较（处理 \ vs /、大小写等差异），
// 若仍无法判定则通过 os.Stat + os.SameFile 对比 inode。
// 用于 CopyFile 入口守卫，避免同一文件被 os.Create 截断为零字节。
func sameFilePath(src, dst string) bool {
	if sameCleanPath(src, dst) {
		return true
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return false
	}
	dstInfo, err := os.Stat(dst)
	if err != nil {
		return false
	}
	return os.SameFile(srcInfo, dstInfo)
}

// sameCleanPath 将两个路径规范化为绝对路径后比较。
// Windows 下用 EqualFold 处理大小写不敏感。
func sameCleanPath(src, dst string) bool {
	src = cleanAbsPath(src)
	dst = cleanAbsPath(dst)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(src, dst)
	}
	return src == dst
}

// cleanAbsPath 先 Clean 再 Abs，用于消除路径中的 ../、多余分隔符等。
// Abs 失败时安全退化为 Clean 后的结果。
//
// 注意：filepath.Abs 对相对路径会基于 os.Getwd()（Wails 进程 CWD）解析，
// 而非用户站点目录。调用方必须传入绝对路径；本仓库调用方（FeatureImage.Path）
// 均来自文件选择器，天然为绝对路径。
func cleanAbsPath(path string) string {
	cleaned := filepath.Clean(path)
	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return cleaned
	}
	return abs
}

// pathInsideDir 判断 path 是否严格位于 dir 目录内部（排除 dir 自身及父目录）。
// 用于 CopyFile 后将已复制到 postImageDir 内的源文件安全删除：
// 仅当源文件已在目标目录内时才删除，防御路径穿越。
func pathInsideDir(path, dir string) bool {
	pathAbs := cleanAbsPath(path)
	dirAbs := cleanAbsPath(dir)
	rel, err := filepath.Rel(dirAbs, pathAbs)
	if err != nil {
		return false
	}
	// 排除 dir 自身（"."）、父目录（".."）、绝对路径（跨盘符/卷），
	// 以及以 "..\" 或 "../" 开头的相对路径（在 dir 外但 Rel 未判定为 ".."）。
	return rel != "." && rel != ".." && !filepath.IsAbs(rel) &&
		!strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

// FileMutex 管理对应文件的读写锁
// 简单实现：使用全局 map 或 sync.Map 存储每个文件的锁？
// 或者更简单：每个 Repository 实例持有一个 Global Lock for that resource type.
// Gridea Pro 是单用户桌面应用，通常只会有一个实例在运行。
// 为了简化，我们在每个具体 Repository struct 中使用 RWMutex 即可。
