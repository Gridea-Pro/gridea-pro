package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

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

// PlatformMeta 平台连接元信息（非敏感，可安全存入应用级配置）
// 凭证（Token / 密码）存储在系统 Keychain，不在此结构中
type PlatformMeta struct {
	ConnectedVia string `json:"connectedVia"` // "oauth" | "manual"
	Username     string `json:"username,omitempty"`
	AvatarURL    string `json:"avatarUrl,omitempty"`
	Email        string `json:"email,omitempty"`
}

// AppConfig 定义应用级别的全局配置
type AppConfig struct {
	// Sites 多站点列表
	Sites []SiteEntry `json:"sites,omitempty"`
	// AISetting AI 模型配置（应用级，跨站点共享）
	// 包含 API Key 等敏感信息，必须放在应用级而非站点目录
	AISetting *domain.AISetting `json:"aiSetting,omitempty"`
	// PlatformMeta 各平台连接状态元信息（非敏感，凭证在 Keychain）
	PlatformMeta map[string]PlatformMeta `json:"platformMeta,omitempty"`
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
		// 极端环境（容器/沙箱/无图形会话缺 HOME/XDG_CONFIG_HOME/%AppData%）兜底到临时目录，
		// 保证返回非 nil 实例——调用方常写 `cm, _ := NewConfigManager()` 后立即 cm.AppConfigDir()，
		// 返回 nil 会直接空指针 panic 崩溃整个应用。仍返回 err 供关心降级的调用方感知。
		log.Printf("[config] os.UserConfigDir 失败，降级到临时目录: %v", err)
		appConfigDir := filepath.Join(os.TempDir(), AppName)
		return &ConfigManager{
			configDir:  appConfigDir,
			configPath: filepath.Join(appConfigDir, ConfigFileName),
		}, err
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
		// config.json 损坏（如写入中途被杀留下的半截 JSON）。
		// 关键：绝不能直接返回空配置继续用同一路径——否则上层写操作会用空配置覆盖磁盘上
		// 的多站点列表 / AI Key / OAuth 身份。这里把损坏文件改名保留现场（config.json.corrupted-<ts>），
		// 再让应用以空配置从头开始：既避免覆盖损坏但可人工抢救的原文件，又让应用可继续使用而非半瘫。
		backupPath := fmt.Sprintf("%s.corrupted-%d", m.configPath, time.Now().Unix())
		if renameErr := os.Rename(m.configPath, backupPath); renameErr != nil {
			// 连改名都失败（权限/占用）时，退回严格模式：返回错误阻止写入，绝不用空配置覆盖。
			log.Printf("[config] config.json 解析失败且无法改名保留现场，已阻止后续写入以防覆盖: parse=%v rename=%v", err, renameErr)
			return nil, fmt.Errorf("parse config.json: %w", err)
		}
		log.Printf("[config] config.json 解析失败（可能损坏），已改名保留为 %s，应用以空配置继续: %v", backupPath, err)
		return &AppConfig{}, nil
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

	// 原子写：先写临时文件再 rename，避免写入中途被杀/断电留下半截 JSON，
	// 那种损坏文件会在下次启动时触发配置读取失败。使用 0600 权限增强安全性。
	tmpPath := m.configPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, m.configPath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
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
		// 配置损坏时中止，不能用空配置覆盖掉 AISetting / PlatformMeta 等其他字段。
		return err
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
		// 配置损坏时中止，不能用空配置覆盖掉 Sites / PlatformMeta 等其他字段。
		return err
	}
	cfg.AISetting = &setting
	return m.SaveConfig(cfg)
}

// AppConfigDir 返回应用级配置目录路径（如 ~/.config/Gridea Pro）
// 用于存放 AI 调用计数等设备级状态
func (m *ConfigManager) AppConfigDir() string {
	return m.configDir
}

// GetPlatformMeta 获取指定平台的连接元信息
func (m *ConfigManager) GetPlatformMeta(providerID string) PlatformMeta {
	cfg, err := m.LoadConfig()
	if err != nil || cfg.PlatformMeta == nil {
		return PlatformMeta{}
	}
	return cfg.PlatformMeta[providerID]
}

// SavePlatformMeta 保存指定平台的连接元信息
// 传入零值 PlatformMeta 表示清除该平台的元信息
func (m *ConfigManager) SavePlatformMeta(providerID string, meta PlatformMeta) error {
	cfg, err := m.LoadConfig()
	if err != nil {
		// 配置损坏时中止，不能用空配置覆盖掉 Sites / AISetting 等其他字段。
		return err
	}
	if cfg.PlatformMeta == nil {
		cfg.PlatformMeta = make(map[string]PlatformMeta)
	}
	if meta.ConnectedVia == "" {
		delete(cfg.PlatformMeta, providerID)
	} else {
		cfg.PlatformMeta[providerID] = meta
	}
	return m.SaveConfig(cfg)
}

// GetAllPlatformMeta 获取所有平台的连接元信息
func (m *ConfigManager) GetAllPlatformMeta() map[string]PlatformMeta {
	cfg, err := m.LoadConfig()
	if err != nil || cfg.PlatformMeta == nil {
		return make(map[string]PlatformMeta)
	}
	return cfg.PlatformMeta
}

// MigrateToSites 确保 Sites 列表存在
// 如果 Sites 已有数据则不操作，否则创建默认站点
func (m *ConfigManager) MigrateToSites(defaultPath string) ([]SiteEntry, error) {
	config, err := m.LoadConfig()
	if err != nil {
		// 配置损坏时中止，不能用空配置覆盖并把默认站点写回去（会抹掉真实的多站点列表）。
		return nil, err
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
