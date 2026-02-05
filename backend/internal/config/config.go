package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AppConfig 定义应用级别的全局配置
type AppConfig struct {
	// SourceFolder 站点源文件目录
	SourceFolder string `json:"sourceFolder"`
}

// ConfigManager 负责管理 AppConfig 的加载和保存
type ConfigManager struct {
	configDir  string
	configPath string
}

// NewConfigManager 创建新的配置管理器实例
func NewConfigManager() *ConfigManager {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".gridea pro")

	// 确保配置目录存在
	_ = os.MkdirAll(configDir, 0755)

	return &ConfigManager{
		configDir:  configDir,
		configPath: filepath.Join(configDir, "config.json"),
	}
}

// LoadConfig 加载配置，如果文件不存在则返回特定错误或空配置
func (m *ConfigManager) LoadConfig() (*AppConfig, error) {
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return &AppConfig{}, nil
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
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
func (m *ConfigManager) SaveConfig(config *AppConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// UpdateSourceFolder 更新源文件路径并保存
func (m *ConfigManager) UpdateSourceFolder(path string) error {
	config, err := m.LoadConfig()
	if err != nil {
		// 如果加载失败，尝试创建一个新的覆盖
		config = &AppConfig{}
	}

	config.SourceFolder = path
	return m.SaveConfig(config)
}
