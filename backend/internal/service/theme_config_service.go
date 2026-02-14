package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gridea-pro/backend/internal/model"
)

// ThemeConfigService 主题配置服务
type ThemeConfigService struct {
	appDir string
	cache  map[string]*model.ThemeConfig
	mu     sync.RWMutex
}

// NewThemeConfigService 创建主题配置服务
func NewThemeConfigService(appDir string) *ThemeConfigService {
	return &ThemeConfigService{
		appDir: appDir,
		cache:  make(map[string]*model.ThemeConfig),
	}
}

// LoadThemeConfig 加载主题配置定义
func (s *ThemeConfigService) LoadThemeConfig(themeName string) (*model.ThemeConfig, error) {
	// 1. Check Cache (Read Lock)
	s.mu.RLock()
	if config, ok := s.cache[themeName]; ok {
		s.mu.RUnlock()
		return config, nil
	}
	s.mu.RUnlock()

	// 2. Load from Disk (Write Lock)
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double check cache after acquiring write lock
	if config, ok := s.cache[themeName]; ok {
		return config, nil
	}

	configPath := filepath.Join(s.appDir, "themes", themeName, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取主题配置失败: %w", err)
	}

	var config model.ThemeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析主题配置失败: %w", err)
	}

	// Update Cache
	s.cache[themeName] = &config

	return &config, nil
}

// GetDefaultConfig 获取主题默认配置（name -> value map）
func (s *ThemeConfigService) GetDefaultConfig(themeName string) (map[string]interface{}, error) {
	// Reuses LoadThemeConfig which is cached
	themeConfig, err := s.LoadThemeConfig(themeName)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, item := range themeConfig.CustomConfig {
		result[item.Name] = item.Value
	}

	return result, nil
}

// MergeConfig 合并默认配置和用户配置
func (s *ThemeConfigService) MergeConfig(
	defaultConfig map[string]interface{},
	userConfig map[string]interface{},
) map[string]interface{} {
	result := make(map[string]interface{})

	// 先复制默认配置
	for k, v := range defaultConfig {
		result[k] = v
	}

	// 用户配置覆盖默认值
	for k, v := range userConfig {
		result[k] = v
	}

	return result
}

// GetFinalConfig 获取最终配置（当前仅使用默认配置）
func (s *ThemeConfigService) GetFinalConfig(themeName string) (map[string]interface{}, error) {
	// Phase 1：仅返回默认配置
	return s.GetDefaultConfig(themeName)

	// TODO: Phase 2：读取用户配置并合并
	// userConfig := s.LoadUserConfig(themeName)
	// return s.MergeConfig(defaultConfig, userConfig), nil
}
