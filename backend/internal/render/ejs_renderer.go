package render

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"gridea-pro/backend/internal/template"

	"github.com/dop251/goja"
)

//go:embed ejs.min.js
var ejsJS string
var ejsProgram *goja.Program
var ejsProgramOnce sync.Once

// Constants for VM Pool
const (
	MaxPoolSize = 20
)

// EjsRenderer EJS 渲染器
// 使用 Goja (Go 的 JavaScript 运行时) + ejs.js 直接执行 EJS
type EjsRenderer struct {
	config RenderConfig

	// 模板缓存
	cache     map[string]string // 缓存模板内容
	cacheLock sync.RWMutex

	// VM Pool (Bounded)
	// pool 存储可用的 VM。如果不为空，直接取用。如果不为空但 pool 空，则阻塞等待。
	pool   chan *goja.Runtime
	logger *slog.Logger
}

// NewEjsRenderer 创建 EJS 渲染器
func NewEjsRenderer(config RenderConfig) (*EjsRenderer, error) {
	r := &EjsRenderer{
		config: config,
		cache:  make(map[string]string),
		pool:   make(chan *goja.Runtime, MaxPoolSize),
		logger: slog.Default(),
	}

	// 预热 VM 池 (Pre-fill)
	// 这样可以确保我们有一个固定大小的池，且不会动态无限制创建
	// 虽然启动时会有短暂开销，但保证了运行时的稳定性
	var lastErr error
	created := 0
	for i := 0; i < MaxPoolSize; i++ {
		vm, err := r.createVM()
		if err != nil {
			// 单个 VM 初始化失败容错：池不满只是吞吐量略低，不影响正确性。
			r.logger.Warn("Failed to initialize VM", "index", i, "error", err)
			lastErr = err
			continue
		}
		r.pool <- vm
		created++
	}

	// 全部失败则返回错误，绝不返回空池：空池会让每次取 VM 都空等到超时才失败，
	// 把整站渲染拖成分钟级假死，还占着 renderMu 拖垮其它请求——不如启动即失败可诊断。
	if created == 0 {
		return nil, fmt.Errorf("EJS 渲染器初始化失败：VM 池预热全部失败，最后错误：%w", lastErr)
	}

	return r, nil
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

	// 注意：我们不清除 VM Pool，因为 VM 的上下文环境（polyfills, ejs lib）是静态的
	// 每次使用时我们都会重新注入 data，所以重用 VM 是安全的
}

// renderViaGoja 通过 Goja 直接执行 EJS
func (r *EjsRenderer) renderViaGoja(templateName string, data *template.TemplateData) (string, error) {
	// 1. 获取模板内容
	templateContent, err := r.getTemplateContent(templateName)
	if err != nil {
		return "", err
	}

	// 2. 数据清洗 (Go Side)
	// 在序列化之前，确保数据结构符合前端预期，减少 JS 端处理负担
	r.sanitizeData(data)

	// 3. 序列化数据
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("序列化数据失败: %w", err)
	}

	// 4. 获取 VM（带超时兜底：若 VM 池初始化全部失败导致池为空，避免在此永久阻塞、
	//    进而拖垮持有 renderMu 的整条渲染流水线）
	var vm *goja.Runtime
	select {
	case vm = <-r.pool:
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("EJS 渲染无可用 VM（VM 池可能初始化失败）")
	}
	var vmInterrupted atomic.Bool
	defer func() {
		if vmInterrupted.Load() {
			// 被超时中断的 VM 内部状态不可信，绝不放回池子。尝试补一个新 VM 保持池容量；
			// 补建失败就让池缩容（谁也不还）——宁可少一个 VM，也不能把坏 VM 放回污染后续渲染。
			if nvm, cerr := r.createVM(); cerr == nil {
				r.pool <- nvm
			} else {
				r.logger.Warn("EJS: 中断的 VM 已丢弃，补建替换 VM 失败，池容量 -1", "error", cerr)
			}
			return
		}
		// 正常归还池子。
		r.pool <- vm
	}()

	// 5. 准备参数
	themePath := filepath.Join(r.config.AppDir, "themes", r.config.ThemeName)
	// 构造模板的绝对路径，用于 EJS 的 filename 选项，以便 include 相对路径工作
	templateAbsPath := filepath.Join(themePath, "templates", templateName)
	if filepath.Ext(templateAbsPath) == "" {
		templateAbsPath += ".ejs"
	}

	// 6. 执行脚本
	// 直接调用 ejs.render，不再在 JS 里写大段 logic
	// 我们把 dataJSON 解析为 JS 对象传递进去
	script := fmt.Sprintf(`
		(function() {
			try {
				var data = %s;
				var template = %s;
				
				// 兼容性处理：某些主题期望 site.posts、site.tags 存在
				// 虽然我们在 Go 里做了清洗，但对象引用结构最好在这里保证
				if (!data.site) data.site = {};
				// 建立引用，避免拷贝
				if (!data.site.posts) data.site.posts = data.posts;
				if (!data.site.tags) data.site.tags = data.tags;
				if (!data.site.menus) data.site.menus = data.menus;

				return ejs.render(template, data, {
					filename: %s, 
					root: %s
				});
			} catch (e) {
				return "EJS Error: " + e.message + "\n" + e.stack;
			}
		})();
	`, string(dataJSON), r.escapeForJS(templateContent), r.escapeForJS(templateAbsPath), r.escapeForJS(themePath))

	// 超时看门狗：第三方主题模板若存在死循环（哪怕只是作者手误），goja 执行无法被 ctx 抢占，
	// 用 Interrupt 强制中断，避免一段死循环耗尽 VM 池、挂死整条渲染流水线。
	timer := time.AfterFunc(15*time.Second, func() {
		vmInterrupted.Store(true)
		vm.Interrupt("EJS 渲染超时（模板可能存在死循环）")
	})
	result, err := vm.RunString(script)
	timer.Stop()
	if err != nil {
		return "", fmt.Errorf("EJS 执行系统错误: %w", err)
	}

	resultStr := result.String()
	if len(resultStr) > 10 && resultStr[:10] == "EJS Error:" {
		return "", fmt.Errorf("%s", resultStr)
	}

	return resultStr, nil
}

// sanitizeData 确保 TemplateData 中的 nil slice 被初始化为空 slice
// 避免 json.Marshal 生成 null，导致前端 EJS 遍历报错
func (r *EjsRenderer) sanitizeData(data *template.TemplateData) {
	if data.Menus == nil {
		data.Menus = []template.MenuView{}
	}
	if data.Posts == nil {
		data.Posts = []template.PostView{}
	} else {
		for i := range data.Posts {
			if data.Posts[i].Tags == nil {
				data.Posts[i].Tags = []template.TagView{}
			}
			if data.Posts[i].Categories == nil {
				data.Posts[i].Categories = []template.CategoryView{}
			}
		}
	}
	if data.Tags == nil {
		data.Tags = []template.TagView{}
	}
	if data.Memos == nil {
		data.Memos = []template.MemoView{}
	}

	// 处理单篇文章的 tags
	if data.Post.Tags == nil {
		data.Post.Tags = []template.TagView{}
	}
	if data.Post.Categories == nil {
		data.Post.Categories = []template.CategoryView{}
	}
}

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
