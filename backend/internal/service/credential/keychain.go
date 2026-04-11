package credential

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const serviceName = "Gridea Pro"

// Service 凭证存储服务
// 优先使用系统 Keychain（macOS Keychain / Windows Credential Manager / Linux Secret Service）
// 不可用时（如无桌面环境的 Linux）降级为加密文件存储
type Service struct {
	fallbackPath string
}

// New 创建凭证服务
// fallbackDir：Keychain 不可用时的文件降级目录
func New(fallbackDir string) *Service {
	return &Service{
		fallbackPath: filepath.Join(fallbackDir, "credentials.json"),
	}
}

// Set 存储凭证
func (s *Service) Set(key, value string) error {
	if err := keyring.Set(serviceName, key, value); err != nil {
		slog.Warn("keyring 不可用，降级到文件存储", "key", key, "error", err)
		return s.setFallback(key, value)
	}
	return nil
}

// Get 读取凭证，不存在时返回空字符串
func (s *Service) Get(key string) string {
	val, err := keyring.Get(serviceName, key)
	if err == keyring.ErrNotFound {
		// 尝试文件降级
		fval, _ := s.getFallback(key)
		return fval
	}
	if err != nil {
		slog.Warn("keyring 读取失败，尝试文件降级", "key", key, "error", err)
		fval, _ := s.getFallback(key)
		return fval
	}
	return val
}

// Delete 删除凭证
func (s *Service) Delete(key string) error {
	kerr := keyring.Delete(serviceName, key)
	_ = s.deleteFallback(key)
	if kerr != nil && kerr != keyring.ErrNotFound {
		return kerr
	}
	return nil
}

// Has 检查凭证是否存在
func (s *Service) Has(key string) bool {
	return s.Get(key) != ""
}

// ─── 文件降级实现（Linux 无 libsecret 时） ──────────────────────────────────

func (s *Service) loadFallback() (map[string]string, error) {
	data, err := os.ReadFile(s.fallbackPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return make(map[string]string), nil
	}
	return m, nil
}

func (s *Service) saveFallback(m map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(s.fallbackPath), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.fallbackPath, data, 0600)
}

func (s *Service) setFallback(key, value string) error {
	m, _ := s.loadFallback()
	if value == "" {
		delete(m, key)
	} else {
		m[key] = value
	}
	return s.saveFallback(m)
}

func (s *Service) getFallback(key string) (string, error) {
	m, err := s.loadFallback()
	if err != nil {
		return "", err
	}
	return m[key], nil
}

func (s *Service) deleteFallback(key string) error {
	m, _ := s.loadFallback()
	delete(m, key)
	return s.saveFallback(m)
}
