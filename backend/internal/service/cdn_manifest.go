package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// cdnManifest 记录一份"远端路径 → 本地 git blob SHA"的映射。
// 持久化在 <appDir>/.cdn-manifest.json，用于跨次部署避免重复的 GET+PUT（见 issue #45）。
//
// 核心假设：一旦某个 remotePath 在远端上传成功，本地就会记录对应的
// blob SHA；下次部署时若本地算出同样的 SHA，就可以直接跳过 API 调用，
// 不需要再向 GitHub 查询远端状态。
type cdnManifest struct {
	mu      sync.Mutex
	Entries map[string]string `json:"entries"` // remotePath → localSHA
}

// loadCdnManifest 从 appDir 读取 manifest；文件不存在或解析失败时返回空白 manifest。
func loadCdnManifest(appDir string) *cdnManifest {
	m := &cdnManifest{Entries: make(map[string]string)}
	data, err := os.ReadFile(filepath.Join(appDir, ".cdn-manifest.json"))
	if err != nil {
		return m
	}
	// 忽略解析错误：异常文件当作"空缓存"，迫使所有文件都重新走一遍完整路径。
	_ = json.Unmarshal(data, m)
	if m.Entries == nil {
		m.Entries = make(map[string]string)
	}
	return m
}

// save 将 manifest 写回磁盘；并发安全。
func (m *cdnManifest) save(appDir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(appDir, ".cdn-manifest.json"), data, 0o644)
}

// hit 查询某个远端路径的已知 SHA；未命中返回 ""，false。
func (m *cdnManifest) hit(remotePath string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	sha, ok := m.Entries[remotePath]
	return sha, ok
}

// record 写入一次成功上传后的映射。
func (m *cdnManifest) record(remotePath, sha string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Entries[remotePath] = sha
}
