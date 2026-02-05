package render

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RendererFactory 渲染器工厂
// 根据主题配置自动创建合适的渲染器
type RendererFactory struct {
	config RenderConfig
}

// NewRendererFactory 创建渲染器工厂
func NewRendererFactory(appDir, themeName string) *RendererFactory {
	cacheDir := filepath.Join(appDir, ".gridea-cache", "templates")
	_ = os.MkdirAll(cacheDir, 0755)

	return &RendererFactory{
		config: RenderConfig{
			AppDir:    appDir,
			ThemeName: themeName,
			CacheDir:  cacheDir,
		},
	}
}

// CreateRenderer 创建渲染器
// 自动识别引擎类型并返回对应的渲染器实例
func (f *RendererFactory) CreateRenderer() (ThemeRenderer, error) {
	engineType, err := f.detectEngineType()
	if err != nil {
		return nil, fmt.Errorf("检测引擎类型失败: %w", err)
	}

	switch engineType {
	case "gotemplate":
		return NewGoTemplateRenderer(f.config), nil
	case "ejs":
		return NewEjsRenderer(f.config), nil
	default:
		return nil, fmt.Errorf("不支持的引擎类型: %s", engineType)
	}
}

// detectEngineType 检测引擎类型
// 优先级:
// 1. config.json 中的 engine 字段
// 2. 根据文件扩展名自动检测
func (f *RendererFactory) detectEngineType() (string, error) {
	themePath := filepath.Join(f.config.AppDir, "themes", f.config.ThemeName)

	// 1. 读取 config.json
	configPath := filepath.Join(themePath, "config.json")
	if engine := f.readEngineFromConfig(configPath); engine != "" {
		return engine, nil
	}

	// 2. 根据文件扩展名检测
	templatesDir := filepath.Join(themePath, "templates")
	return f.detectEngineByExtension(templatesDir)
}

// readEngineFromConfig 从 config.json 读取 engine 字段
func (f *RendererFactory) readEngineFromConfig(configPath string) string {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}

	var config struct {
		Engine string `json:"engine"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return ""
	}

	// 标准化引擎名称
	engine := strings.ToLower(strings.TrimSpace(config.Engine))
	if engine == "go" || engine == "gotemplate" || engine == "gotemplates" {
		return "gotemplate"
	}
	if engine == "ejs" {
		return "ejs"
	}

	return ""
}

// detectEngineByExtension 根据文件扩展名检测引擎
func (f *RendererFactory) detectEngineByExtension(templatesDir string) (string, error) {
	// 检查是否存在 .ejs 文件
	ejsFiles, _ := filepath.Glob(filepath.Join(templatesDir, "*.ejs"))
	if len(ejsFiles) > 0 {
		return "ejs", nil
	}

	// 检查是否存在 .html 或 .gohtml 文件
	htmlFiles, _ := filepath.Glob(filepath.Join(templatesDir, "*.html"))
	gohtmlFiles, _ := filepath.Glob(filepath.Join(templatesDir, "*.gohtml"))
	if len(htmlFiles) > 0 || len(gohtmlFiles) > 0 {
		return "gotemplate", nil
	}

	// 默认使用 EJS (向后兼容)
	return "ejs", nil
}

// GetEngineType 获取当前主题的引擎类型(不创建渲染器)
func (f *RendererFactory) GetEngineType() (string, error) {
	return f.detectEngineType()
}
