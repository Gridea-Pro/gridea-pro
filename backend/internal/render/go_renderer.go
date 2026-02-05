package render

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gridea-pro/backend/internal/template"
)

// GoTemplateRenderer Go Templates 渲染器
// 使用 Go 标准库 html/template 实现
type GoTemplateRenderer struct {
	config RenderConfig

	// 模板缓存
	cache     map[string]*htmltemplate.Template
	cacheLock sync.RWMutex
}

// NewGoTemplateRenderer 创建 Go Templates 渲染器
func NewGoTemplateRenderer(config RenderConfig) *GoTemplateRenderer {
	return &GoTemplateRenderer{
		config: config,
		cache:  make(map[string]*htmltemplate.Template),
	}
}

// Render 实现 ThemeRenderer 接口
func (r *GoTemplateRenderer) Render(templateName string, data *template.TemplateData) (string, error) {
	// 获取模板
	tmpl, err := r.getTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("加载模板失败: %w", err)
	}

	// 执行渲染
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("渲染模板失败: %w", err)
	}

	return buf.String(), nil
}

// GetEngineType 实现 ThemeRenderer 接口
func (r *GoTemplateRenderer) GetEngineType() string {
	return "gotemplate"
}

// ClearCache 实现 ThemeRenderer 接口
func (r *GoTemplateRenderer) ClearCache() {
	r.cacheLock.Lock()
	defer r.cacheLock.Unlock()
	r.cache = make(map[string]*htmltemplate.Template)
}

// getTemplate 获取模板(带缓存)
func (r *GoTemplateRenderer) getTemplate(name string) (*htmltemplate.Template, error) {
	// 检查缓存
	r.cacheLock.RLock()
	if tmpl, ok := r.cache[name]; ok {
		r.cacheLock.RUnlock()
		return tmpl, nil
	}
	r.cacheLock.RUnlock()

	// 加载模板
	tmpl, err := r.loadTemplate(name)
	if err != nil {
		return nil, err
	}

	// 缓存
	r.cacheLock.Lock()
	r.cache[name] = tmpl
	r.cacheLock.Unlock()

	return tmpl, nil
}

// loadTemplate 加载模板文件
func (r *GoTemplateRenderer) loadTemplate(name string) (*htmltemplate.Template, error) {
	themePath := filepath.Join(r.config.AppDir, "themes", r.config.ThemeName)
	templatesDir := filepath.Join(themePath, "templates")

	// 主模板路径(支持 .html 和 .gohtml)
	mainTemplatePath := filepath.Join(templatesDir, name+".html")
	if _, err := os.Stat(mainTemplatePath); os.IsNotExist(err) {
		mainTemplatePath = filepath.Join(templatesDir, name+".gohtml")
	}

	// 创建模板并注册函数
	tmpl := htmltemplate.New(name).Funcs(template.TemplateFuncs())

	// 读取主模板
	content, err := os.ReadFile(mainTemplatePath)
	if err != nil {
		return nil, fmt.Errorf("读取主模板失败: %w", err)
	}

	tmpl, err = tmpl.Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("解析主模板失败: %w", err)
	}

	// 加载 includes
	includesDir := filepath.Join(templatesDir, "includes")
	if _, err := os.Stat(includesDir); err == nil {
		if err := r.loadIncludes(tmpl, includesDir); err != nil {
			return nil, fmt.Errorf("加载 includes 失败: %w", err)
		}
	}

	return tmpl, nil
}

// loadIncludes 加载所有 include 模板
func (r *GoTemplateRenderer) loadIncludes(tmpl *htmltemplate.Template, includesDir string) error {
	patterns := []string{
		filepath.Join(includesDir, "*.html"),
		filepath.Join(includesDir, "*.gohtml"),
	}

	for _, pattern := range patterns {
		files, _ := filepath.Glob(pattern)
		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
			_, err = tmpl.New(baseName).Parse(string(content))
			if err != nil {
				return fmt.Errorf("解析 include %s 失败: %w", file, err)
			}
		}
	}

	return nil
}
