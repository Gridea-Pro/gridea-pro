package render

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// jinja2SyntaxRe 匹配 Jinja2/pongo2 独有的语句块：{% if ... %}、{% for ... %}、
// {% block ... %}、{% extends ... %} 等。Go 模板只使用 {{ ... }}，不会出现 {% 。
// 这里仅要求能匹配到 "{%" + 空白 + 标识符，足以区分两种引擎。
var jinja2SyntaxRe = regexp.MustCompile(`{%[-\s]*[A-Za-z_]`)

// RendererFactory 渲染器工厂
// 根据主题配置自动创建合适的渲染器
type RendererFactory struct {
	config RenderConfig
}

// NewRendererFactory 创建渲染器工厂
func NewRendererFactory(appDir, themeName string) *RendererFactory {
	return &RendererFactory{
		config: RenderConfig{
			AppDir:    appDir,
			ThemeName: themeName,
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
	case "jinja2":
		return NewJinja2Renderer(f.config), nil
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
		Engine         string `json:"engine"`
		TemplateEngine string `json:"templateEngine"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return ""
	}

	// 标准化引擎名称（优先使用 engine 字段，其次使用 templateEngine 字段）
	raw := config.Engine
	if raw == "" {
		raw = config.TemplateEngine
	}
	engine := strings.ToLower(strings.TrimSpace(raw))
	if engine == "go" || engine == "gotemplate" || engine == "gotemplates" {
		return "gotemplate"
	}
	if engine == "ejs" {
		return "ejs"
	}
	if engine == "jinja2" || engine == "jinja" || engine == "j2" {
		return "jinja2"
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

	// 检查是否存在 .jinja2 或 .j2 文件（明确后缀，无歧义）
	jinja2Files, _ := filepath.Glob(filepath.Join(templatesDir, "*.jinja2"))
	j2Files, _ := filepath.Glob(filepath.Join(templatesDir, "*.j2"))
	if len(jinja2Files) > 0 || len(j2Files) > 0 {
		return "jinja2", nil
	}

	// 检查是否存在 .gohtml 文件（明确后缀，仅 Go 模板使用）
	gohtmlFiles, _ := filepath.Glob(filepath.Join(templatesDir, "*.gohtml"))
	if len(gohtmlFiles) > 0 {
		return "gotemplate", nil
	}

	// 检查是否存在 .html 文件。.html 是 jinja2/pongo2 和 Go 模板通用的后缀，
	// 仅凭后缀无法判断，因此对内容做嗅探：若文件含 {% ... %} 语句块特征就判为
	// jinja2，否则回退到 gotemplate（原有默认行为）。
	htmlFiles, _ := filepath.Glob(filepath.Join(templatesDir, "*.html"))
	if len(htmlFiles) > 0 {
		if sniffJinja2FromHTMLFiles(htmlFiles) {
			return "jinja2", nil
		}
		return "gotemplate", nil
	}

	// 默认使用 EJS (向后兼容)
	return "ejs", nil
}

// sniffJinja2FromHTMLFiles 读取若干 .html 文件首部内容，检测是否包含 Jinja2/pongo2
// 独有的 {% ... %} 语句块。命中任一文件即返回 true。
//
// 采用"读首部"策略而非全文，避免超大模板拖慢启动；若首部就有 extends/block/if/for
// 任一语句，足以区分引擎；若首部全是静态 HTML，则继续读完该文件，保证覆盖率。
func sniffJinja2FromHTMLFiles(paths []string) bool {
	const headSize = 8 * 1024 // 8KB，覆盖绝大多数模板头部的 extends/block 块
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		target := data
		if len(target) > headSize {
			target = target[:headSize]
		}
		if jinja2SyntaxRe.Match(target) {
			return true
		}
	}
	return false
}

// GetEngineType 获取当前主题的引擎类型(不创建渲染器)
func (f *RendererFactory) GetEngineType() (string, error) {
	return f.detectEngineType()
}
