package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"gridea-pro/backend/internal/domain"
)

const (
	// AppName 应用名称，用于生成配置目录
	AppName = "Gridea Pro"
	// ConfigFileName 配置文件名
	ConfigFileName = "config.json"
)

// SiteEntry 站点条目
type SiteEntry struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Active bool   `json:"active"`
}

// AppConfig 定义应用级别的全局配置
type AppConfig struct {
	// Sites 多站点列表
	Sites []SiteEntry `json:"sites,omitempty"`
	// AISetting AI 模型配置（应用级，跨站点共享）
	// 包含 API Key 等敏感信息，必须放在应用级而非站点目录
	AISetting *domain.AISetting `json:"aiSetting,omitempty"`
}

// ConfigManager 负责管理 AppConfig 的加载和保存
type ConfigManager struct {
	configDir  string
	configPath string
}

// NewConfigManager 创建新的配置管理器实例
// 使用系统标准的配置目录 (os.UserConfigDir)
func NewConfigManager() (*ConfigManager, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	appConfigDir := filepath.Join(configDir, AppName)
	return &ConfigManager{
		configDir:  appConfigDir,
		configPath: filepath.Join(appConfigDir, ConfigFileName),
	}, nil
}

// LoadConfig 加载配置
// 如果文件不存在，返回默认的空配置和 nil 错误
func (m *ConfigManager) LoadConfig() (*AppConfig, error) {
	// 直接尝试读取文件，利用 os.IsNotExist 判断
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &AppConfig{}, nil
		}
		return nil, err
	}

	var config AppConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig 保存配置到文件
// 延迟创建目录 (Lazy Creation)，且权限为 0600 (仅当前用户读写)
func (m *ConfigManager) SaveConfig(config *AppConfig) error {
	// 确保配置目录存在
	if err := os.MkdirAll(m.configDir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 使用 0600 权限增强安全性
	return os.WriteFile(m.configPath, data, 0600)
}

// GetSites 获取站点列表
func (m *ConfigManager) GetSites() ([]SiteEntry, error) {
	config, err := m.LoadConfig()
	if err != nil {
		return nil, err
	}
	return config.Sites, nil
}

// SaveSites 保存站点列表
func (m *ConfigManager) SaveSites(sites []SiteEntry) error {
	config, err := m.LoadConfig()
	if err != nil {
		config = &AppConfig{}
	}
	config.Sites = sites
	return m.SaveConfig(config)
}

// GetActiveSite 获取当前活跃站点
func (m *ConfigManager) GetActiveSite() (*SiteEntry, error) {
	config, err := m.LoadConfig()
	if err != nil {
		return nil, err
	}
	for i := range config.Sites {
		if config.Sites[i].Active {
			return &config.Sites[i], nil
		}
	}
	return nil, nil
}

// GetAISetting 获取 AI 配置（不存在则返回零值）
func (m *ConfigManager) GetAISetting() (domain.AISetting, error) {
	cfg, err := m.LoadConfig()
	if err != nil {
		return domain.AISetting{}, err
	}
	if cfg.AISetting == nil {
		return domain.AISetting{}, nil
	}
	return *cfg.AISetting, nil
}

// SaveAISetting 保存 AI 配置
func (m *ConfigManager) SaveAISetting(setting domain.AISetting) error {
	cfg, err := m.LoadConfig()
	if err != nil {
		cfg = &AppConfig{}
	}
	cfg.AISetting = &setting
	return m.SaveConfig(cfg)
}

// AppConfigDir 返回应用级配置目录路径（如 ~/.config/Gridea Pro）
// 用于存放 AI 调用计数等设备级状态
func (m *ConfigManager) AppConfigDir() string {
	return m.configDir
}

// MigrateToSites 确保 Sites 列表存在
// 如果 Sites 已有数据则不操作，否则创建默认站点
func (m *ConfigManager) MigrateToSites(defaultPath string) ([]SiteEntry, error) {
	config, err := m.LoadConfig()
	if err != nil {
		config = &AppConfig{}
	}

	if len(config.Sites) > 0 {
		return config.Sites, nil
	}

	config.Sites = []SiteEntry{
		{Name: "My Site", Path: defaultPath, Active: true},
	}

	if err := m.SaveConfig(config); err != nil {
		return nil, err
	}
	return config.Sites, nil
}
