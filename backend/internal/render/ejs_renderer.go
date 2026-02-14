package render

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gridea-pro/backend/internal/template"

	"github.com/dop251/goja"
)

//go:embed ejs.min.js
var ejsJS string
var ejsProgram *goja.Program
var ejsProgramOnce sync.Once

// EjsRenderer EJS 渲染器
// 使用 Goja (Go 的 JavaScript 运行时) + ejs.js 直接执行 EJS
type EjsRenderer struct {
	config RenderConfig

	// 模板缓存
	cache     map[string]string // 缓存模板内容
	cacheLock sync.RWMutex

	// VM Pool
	pool chan *goja.Runtime
}

// createVM 创建新的 VM 实例
func (r *EjsRenderer) createVM() (*goja.Runtime, error) {
	// 创建新的 VM
	vm := goja.New()

	// 1. 注入 Node.js 环境模拟 (fs, path, require, process)
	// 计算当前主题的根目录作为 CWD
	themeDir := filepath.Join(r.config.AppDir, "themes", r.config.ThemeName)
	SetupNodePolyfills(vm, themeDir)

	// 2. 模拟 CommonJS 环境 (module, exports)
	vm.Set("exports", vm.NewObject())
	moduleObj := vm.NewObject()
	moduleObj.Set("exports", vm.Get("exports"))
	vm.Set("module", moduleObj)

	// 3. 加载 ejs.js 库
	// Use pre-compiled program if possible
	var err error
	ejsProgramOnce.Do(func() {
		ejsProgram, err = goja.Compile("ejs.min.js", ejsJS, true)
	})
	if err != nil {
		return nil, fmt.Errorf("编译 ejs.min.js 失败: %w", err)
	}

	if ejsProgram != nil {
		_, err = vm.RunProgram(ejsProgram)
	} else {
		_, err = vm.RunString(ejsJS)
	}

	if err != nil {
		return nil, fmt.Errorf("加载 ejs.js 失败: %w", err)
	}

	// 4. 验证 ejs 全局变量
	ejsVal := vm.Get("ejs")
	if ejsVal == nil || goja.IsUndefined(ejsVal) {
		return nil, fmt.Errorf("EJS 加载失败：全局 ejs 对象未定义")
	}

	return vm, nil
}

// NewEjsRenderer 创建 EJS 渲染器
func NewEjsRenderer(config RenderConfig) *EjsRenderer {
	return &EjsRenderer{
		config: config,
		cache:  make(map[string]string),
		pool:   make(chan *goja.Runtime, 32), // Allow up to 32 concurrent VMs
	}
}

// Render 实现 ThemeRenderer 接口
func (r *EjsRenderer) Render(templateName string, data *template.TemplateData) (string, error) {
	return r.renderViaGoja(templateName, data)
}

// GetEngineType 实现 ThemeRenderer 接口
func (r *EjsRenderer) GetEngineType() string {
	return "ejs"
}

// ClearCache 实现 ThemeRenderer 接口
func (r *EjsRenderer) ClearCache() {
	r.cacheLock.Lock()
	defer r.cacheLock.Unlock()
	r.cache = make(map[string]string)

	// Clear pool
loop:
	for {
		select {
		case <-r.pool:
		default:
			break loop
		}
	}
}

// renderViaGoja 通过 Goja 直接执行 EJS
func (r *EjsRenderer) renderViaGoja(templateName string, data *template.TemplateData) (string, error) {
	// 获取模板内容
	templateContent, err := r.getTemplateContent(templateName)
	if err != nil {
		return "", err
	}

	// 提前序列化数据 (Move out of critical section/init)
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("序列化数据失败: %w", err)
	}

	// 获取或创建 VM
	vm, err := r.getVM()
	if err != nil {
		return "", fmt.Errorf("创建 JS 运行时失败: %w", err)
	}
	defer r.returnVM(vm)

	// 主题路径
	themePath := filepath.Join(r.config.AppDir, "themes", r.config.ThemeName)

	// 构造模板的绝对路径，用于 EJS 的 filename 选项
	// 这样 EJS 才能正确解析相对路径 include (例如 ./includes/head)
	templateAbsPath := filepath.Join(themePath, "templates", templateName)
	// 如果没有后缀，加上 .ejs (这里做一个简单的假设，因为 getTemplateContent 已经处理了读取)
	// 为了更严谨，最好是从 getTemplateContent 返回路径，但这里先简单拼接
	if filepath.Ext(templateAbsPath) == "" {
		templateAbsPath += ".ejs"
	}

	// 执行 EJS 渲染
	script := fmt.Sprintf(`
		(function() {
			var data = %s;
			var template = %s;

			// 数据清洗：将 null 数组转换为 []，防止 EJS 报错
			// Go 的 nil slice 会被序列化为 null，而 EJS 模板通常直接调用 .forEach 或 .map
			if (data.menus === null) data.menus = [];
			if (data.posts === null) data.posts = [];
			if (data.tags === null) data.tags = [];
			if (data.post && data.post.tags === null) data.post.tags = [];
			
			// 清洗 posts 数组中每个 post 的 tags
			if (data.posts && Array.isArray(data.posts)) {
				data.posts.forEach(function(post) {
					if (post && post.tags === null) {
						post.tags = [];
					}
				});
			}
			
			// 兼容性处理：某些主题期望 site.posts、site.tags、site.menus 存在
			// 确保 site 对象存在并包含必要的属性
			if (!data.site) data.site = {};
			if (!data.site.posts) {
				data.site.posts = data.posts || [];
			}
			if (!data.site.tags) {
				data.site.tags = data.tags || [];
			}
			if (!data.site.menus) {
				data.site.menus = data.menus || [];
			}
			
			try {
				return ejs.render(template, data, {
					filename: %s, // 重要的是提供文件名，以便相对路径 include 工作
					root: %s // 设置根目录
				});
			} catch (e) {
				return "EJS Error: " + e.message + "\n" + e.stack;
			}
		})();
	`, string(dataJSON), r.escapeForJS(templateContent), r.escapeForJS(templateAbsPath), r.escapeForJS(themePath))

	result, err := vm.RunString(script)
	if err != nil {
		return "", fmt.Errorf("EJS 执行出错: %w", err)
	}

	resultStr := result.String()
	if len(resultStr) > 10 && resultStr[:10] == "EJS Error:" {
		// Log error for debugging
		// fmt.Println(resultStr)
		return "", fmt.Errorf("%s", resultStr)
	}

	return resultStr, nil
}

// getVM 获取或创建 Goja VM
func (r *EjsRenderer) getVM() (*goja.Runtime, error) {
	select {
	case vm := <-r.pool:
		return vm, nil
	default:
		return r.createVM()
	}
}

// returnVM 归还 VM 到池中
func (r *EjsRenderer) returnVM(vm *goja.Runtime) {
	select {
	case r.pool <- vm:
	default:
		// Pool is full, discard
	}
}

// setupNodePolyfills has been replaced by node_polyfills.go SetupNodePolyfills

// getTemplateContent 获取模板内容
func (r *EjsRenderer) getTemplateContent(name string) (string, error) {
	// 检查缓存
	r.cacheLock.RLock()
	if content, ok := r.cache[name]; ok {
		r.cacheLock.RUnlock()
		return content, nil
	}
	r.cacheLock.RUnlock()

	// 读取文件
	themePath := filepath.Join(r.config.AppDir, "themes", r.config.ThemeName)
	templatePath := filepath.Join(themePath, "templates", name+".ejs")

	content, err := os.ReadFile(templatePath)
	if err != nil {
		// 尝试不带 .ejs 后缀
		content, err = os.ReadFile(filepath.Join(themePath, "templates", name))
		if err != nil {
			return "", fmt.Errorf("读取 EJS 模板失败: %w", err)
		}
	}

	contentStr := string(content)

	// 缓存
	r.cacheLock.Lock()
	r.cache[name] = contentStr
	r.cacheLock.Unlock()

	return contentStr, nil
}

// escapeForJS 将字符串转义为 JS 字符串字面量
func (r *EjsRenderer) escapeForJS(s string) string {
	jsonBytes, _ := json.Marshal(s)
	return string(jsonBytes)
}
