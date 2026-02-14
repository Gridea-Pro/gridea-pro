package repository

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
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

// SaveJSONFile 将数据保存为 JSON 文件
func SaveJSONFile(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// SaveJSONFileIdempotent 将数据保存为 JSON 文件，但只有内容变化时才写入
func SaveJSONFileIdempotent(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Read existing file to compare
	existingData, err := os.ReadFile(path)
	if err == nil && string(existingData) == string(data) {
		return nil // Content matches, skip write
	}

	return os.WriteFile(path, data, 0644)
}

func CopyFile(src, dst string) error {
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

// FileMutex 管理对应文件的读写锁
// 简单实现：使用全局 map 或 sync.Map 存储每个文件的锁？
// 或者更简单：每个 Repository 实例持有一个 Global Lock for that resource type.
// Gridea Pro 是单用户桌面应用，通常只会有一个实例在运行。
// 为了简化，我们在每个具体 Repository struct 中使用 RWMutex 即可。
