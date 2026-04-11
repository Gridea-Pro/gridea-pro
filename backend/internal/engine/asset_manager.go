package engine

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dop251/goja"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/typomedia/lessgo/less"
)

type lessFileReader struct {
	baseDir string
}

func (r *lessFileReader) ReadFile(path string) ([]byte, error) {
	if filepath.IsAbs(path) {
		return os.ReadFile(path)
	}
	return os.ReadFile(filepath.Join(r.baseDir, path))
}

func inlineLess(content string, baseDir string, visited map[string]bool) (string, error) {
	importRe := regexp.MustCompile(`@import\s+["']([^"']+)["'];?`)

	var result bytes.Buffer
	lastEnd := 0

	matches := importRe.FindAllStringSubmatchIndex(content, -1)

	for _, match := range matches {
		result.WriteString(content[lastEnd:match[0]])

		importPath := content[match[2]:match[3]]

		if strings.HasPrefix(importPath, "http://") ||
			strings.HasPrefix(importPath, "https://") ||
			strings.HasPrefix(importPath, "//") ||
			strings.HasPrefix(importPath, "~") {
			result.WriteString(content[match[0]:match[1]])
			lastEnd = match[1]
			continue
		}

		if !strings.HasSuffix(importPath, ".less") {
			importPath = importPath + ".less"
		}

		fullPath := importPath
		if !filepath.IsAbs(importPath) {
			fullPath = filepath.Join(baseDir, importPath)
		}

		absPath, _ := filepath.Abs(fullPath)
		if visited[absPath] {
			lastEnd = match[1]
			continue
		}
		visited[absPath] = true

		importedContent, err := os.ReadFile(fullPath)
		if err != nil {
			return "", fmt.Errorf("failed to read import %s: %w", importPath, err)
		}

		importedBaseDir := filepath.Dir(fullPath)
		inlined, err := inlineLess(string(importedContent), importedBaseDir, visited)
		if err != nil {
			return "", err
		}

		result.WriteString(inlined)
		lastEnd = match[1]
	}

	result.WriteString(content[lastEnd:])
	return result.String(), nil
}

// AssetManager handles theme and site asset operations
type AssetManager struct {
	appDir             string
	themeConfigService *ThemeConfigService
	logger             *slog.Logger
}

// NewAssetManager creates a new AssetManager instance
func NewAssetManager(appDir string, themeConfigService *ThemeConfigService) *AssetManager {
	return &AssetManager{
		appDir:             appDir,
		themeConfigService: themeConfigService,
		logger:             slog.Default(),
	}
}

// CopyThemeAssets 复制主题静态资源
func (m *AssetManager) CopyThemeAssets(buildDir, themeName string) error {
	themePath := filepath.Join(m.appDir, DirThemes, themeName)
	assetsPath := filepath.Join(themePath, DirAssets)
	if _, err := os.Stat(assetsPath); os.IsNotExist(err) {
		return nil
	}

	// 检查主题是否发生了切换，如果切换则删除旧的 CSS 等静态资源缓存，避免 compileLess 误命中
	themeCacheFile := filepath.Join(buildDir, ".current_theme")
	if cachedTheme, err := os.ReadFile(themeCacheFile); err != nil || string(cachedTheme) != themeName {
		_ = os.RemoveAll(filepath.Join(buildDir, DirStyles))
	}
	_ = os.WriteFile(themeCacheFile, []byte(themeName), 0644)

	// 1. 检查并编译 LESS 文件
	lessPath := filepath.Join(assetsPath, DirStyles, FileMainLess)
	hasLess := false
	if _, err := os.Stat(lessPath); err == nil {
		hasLess = true
		if err := m.compileLess(lessPath, buildDir); err != nil {
			m.logger.Warn("LESS 编译失败", "error", err)
		}
	}

	// 2. 复制其他静态资源
	destPath := filepath.Join(buildDir)
	if err := copyDir(assetsPath, destPath); err != nil {
		return err
	}

	// 3. 对于纯 CSS 主题（无 LESS），也应用 style-override.js
	if !hasLess {
		overridePath := filepath.Join(themePath, FileStyleOverride)
		cssPath := filepath.Join(buildDir, DirStyles, FileMainCSS)
		if _, err := os.Stat(overridePath); err == nil {
			if _, err := os.Stat(cssPath); err == nil {
				m.logger.Info("检测到 style-override.js（纯 CSS 主题），应用自定义样式...")
				customCSS, err := m.applyStyleOverride(overridePath)
				if err != nil {
					m.logger.Warn("应用 style-override.js 失败", "error", err)
				} else if customCSS != "" {
					cssContent, err := os.ReadFile(cssPath)
					if err != nil {
						m.logger.Warn("读取 CSS 文件失败", "error", err)
					} else {
						cssContent = append(cssContent, []byte("\n/* style-override */\n"+customCSS)...)
						if err := os.WriteFile(cssPath, cssContent, 0644); err != nil {
							m.logger.Warn("写入 CSS 文件失败", "error", err)
						} else {
							m.logger.Info("纯 CSS 主题自定义样式应用成功")
						}
					}
				}
			}
		}
	}

	return nil
}

// compileLess 编译 LESS 文件为 CSS
func (m *AssetManager) compileLess(lessPath, buildDir string) error {
	cssPath := filepath.Join(buildDir, DirStyles, FileMainCSS)

	if err := os.MkdirAll(filepath.Dir(cssPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// Optimization: Check if recompilation is needed
	lessInfo, errLess := os.Stat(lessPath)
	configInfo, errConf := os.Stat(filepath.Join(m.appDir, "config", "config.json"))
	if errLess == nil {
		if cssInfo, errCss := os.Stat(cssPath); errCss == nil {
			isNewerThanLess := cssInfo.ModTime().After(lessInfo.ModTime())
			isNewerThanConfig := true
			if errConf == nil {
				isNewerThanConfig = cssInfo.ModTime().After(configInfo.ModTime())
			}
			if isNewerThanLess && isNewerThanConfig {
				return nil
			}
		}
	}

	// 读取 LESS 文件内容
	lessContent, err := os.ReadFile(lessPath)
	if err != nil {
		return fmt.Errorf("读取 LESS 文件失败: %w", err)
	}

	// 内联所有 @import 语句
	stylesDir := filepath.Dir(lessPath)
	visited := make(map[string]bool)
	absPath, _ := filepath.Abs(lessPath)
	visited[absPath] = true

	inlinedContent, err := inlineLess(string(lessContent), stylesDir, visited)
	if err != nil {
		return fmt.Errorf("LESS import 内联失败: %w", err)
	}

	// 使用 lessgo 编译
	m.logger.Info("正在编译 LESS 文件", "less", lessPath, "css", cssPath)
	cssContent, err := less.Render(inlinedContent, map[string]interface{}{"compress": false})
	if err != nil {
		return fmt.Errorf("LESS 编译失败: %w", err)
	}

	// 写入 CSS 文件
	if err := os.WriteFile(cssPath, []byte(cssContent), 0644); err != nil {
		return fmt.Errorf("写入 CSS 文件失败: %w", err)
	}

	// 检查并应用 style-override.js
	themePath := filepath.Dir(filepath.Dir(filepath.Dir(lessPath)))
	overridePath := filepath.Join(themePath, FileStyleOverride)
	if _, err := os.Stat(overridePath); err == nil {
		m.logger.Info("检测到 style-override.js，应用自定义样式...")
		customCSS, err := m.applyStyleOverride(overridePath)
		if err != nil {
			m.logger.Warn("应用 style-override.js 失败", "error", err)
		} else if customCSS != "" {
			cssContent = cssContent + "\n/* style-override */\n" + customCSS
			if err := os.WriteFile(cssPath, []byte(cssContent), 0644); err != nil {
				return fmt.Errorf("写入 CSS 文件失败: %w", err)
			}
			m.logger.Info("自定义样式应用成功")
		}
	}

	return nil
}

// applyStyleOverride 执行 style-override.js 并返回自定义 CSS
func (m *AssetManager) applyStyleOverride(jsPath string) (string, error) {
	// 读取 JS 文件
	jsCode, err := os.ReadFile(jsPath)
	if err != nil {
		return "", fmt.Errorf("读取 style-override.js 失败: %w", err)
	}

	// 创建 JS 运行时
	vm := goja.New()

	// 注入 module 和 exports 环境，解决 module is not defined 报错
	moduleObj := vm.NewObject()
	exportsObj := vm.NewObject()
	_ = moduleObj.Set("exports", exportsObj)
	_ = vm.Set("module", moduleObj)
	_ = vm.Set("exports", exportsObj)

	// 执行 JS 代码
	_, err = vm.RunString(string(jsCode))
	if err != nil {
		return "", fmt.Errorf("执行 JS 代码失败: %w", err)
	}

	// 获取 module.exports (generateOverride 函数)
	moduleExports := vm.Get("module")
	if moduleExports == nil || goja.IsUndefined(moduleExports) {
		// 尝试直接获取 generateOverride
		generateOverride := vm.Get("generateOverride")
		if generateOverride == nil || goja.IsUndefined(generateOverride) {
			return "", fmt.Errorf("未找到 generateOverride 函数")
		}

		// 调用函数
		// 从 jsPath 推导 themeName
		themePath := filepath.Dir(jsPath)
		themeName := filepath.Base(themePath)
		customConfig := m.loadThemeCustomConfig(themeName) // Use helper method
		result, err := vm.RunString(fmt.Sprintf("generateOverride(%s)", toJSON(customConfig)))
		if err != nil {
			return "", fmt.Errorf("调用 generateOverride 失败: %w", err)
		}

		return result.String(), nil
	}

	// CommonJS 模块格式：module.exports = generateOverride
	exports := moduleExports.ToObject(vm).Get("exports")
	if exports == nil || goja.IsUndefined(exports) {
		return "", fmt.Errorf("module.exports 未定义")
	}

	// 调用导出的函数
	fn, ok := goja.AssertFunction(exports)
	if !ok {
		return "", fmt.Errorf("module.exports 不是函数")
	}

	// 准备参数
	// 从 jsPath 推导 themeName
	themePath := filepath.Dir(jsPath)
	themeName := filepath.Base(themePath)
	customConfig := m.loadThemeCustomConfig(themeName)
	configValue := vm.ToValue(customConfig)

	// 调用函数
	result, err := fn(goja.Undefined(), configValue)
	if err != nil {
		return "", fmt.Errorf("调用 generateOverride 失败: %w", err)
	}

	return result.String(), nil
}

// CopySiteAssets 复制站点静态资源
func (m *AssetManager) CopySiteAssets(buildDir string) error {
	// 复制 images 目录
	imagesPath := filepath.Join(m.appDir, DirImages)
	if _, err := os.Stat(imagesPath); err == nil {
		if err := copyDir(imagesPath, filepath.Join(buildDir, DirImages)); err != nil {
			return err
		}
	}

	// 复制 media 目录
	mediaPath := filepath.Join(m.appDir, DirMedia)
	if _, err := os.Stat(mediaPath); err == nil {
		if err := copyDir(mediaPath, filepath.Join(buildDir, DirMedia)); err != nil {
			return err
		}
	}

	// 复制 post-images 目录
	postImagesPath := filepath.Join(m.appDir, DirPostImages)
	if _, err := os.Stat(postImagesPath); err == nil {
		if err := copyDir(postImagesPath, filepath.Join(buildDir, DirPostImages)); err != nil {
			return err
		}
	}

	// 复制 favicon.ico
	faviconSrc := filepath.Join(m.appDir, "favicon.ico")
	if _, err := os.Stat(faviconSrc); err == nil {
		if err := copyFile(faviconSrc, filepath.Join(buildDir, "favicon.ico")); err != nil {
			return err
		}
	}

	return nil
}

// loadThemeCustomConfig 加载主题自定义配置 (Helper for AssetManager)
func (m *AssetManager) loadThemeCustomConfig(themeName string) map[string]interface{} {
	// 使用 ThemeConfigService 加载配置
	config, err := m.themeConfigService.GetFinalConfig(themeName)
	if err != nil {
		m.logger.Warn("加载主题配置失败，使用空配置", "error", err)
		return make(map[string]interface{})
	}
	return config
}

// copyDir 递归复制目录
func copyDir(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("源路径不是目录")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// BundleCSS 合并压缩主题 CSS 文件，减少 HTTP 请求
func (m *AssetManager) BundleCSS(buildDir, themePath string) error {
	// 支持多种模板引擎的 head 文件位置
	headPaths := []string{
		filepath.Join(themePath, DirTemplates, "partials", "head.html"),
		filepath.Join(themePath, DirTemplates, "partials", "head.ejs"),
		filepath.Join(themePath, DirTemplates, "includes", "head.html"),
		filepath.Join(themePath, DirTemplates, "includes", "head.ejs"),
		filepath.Join(themePath, DirTemplates, "_blocks", "head.html"),
		filepath.Join(themePath, DirTemplates, "_blocks", "head.ejs"),
	}

	var headContent []byte
	var headPath string
	var err error
	for _, p := range headPaths {
		headContent, err = os.ReadFile(p)
		if err == nil {
			headPath = p
			break
		}
	}
	if headContent == nil {
		m.logger.Info("未找到 head 模板文件，跳过 CSS 合并")
		return nil
	}
	m.logger.Info("使用 head 模板", "path", headPath)

	// 提取 /styles/xxx.css 引用
	cssRefRe := regexp.MustCompile(`href="/styles/([\w.\-]+\.css)"`)
	matches := cssRefRe.FindAllStringSubmatch(string(headContent), -1)
	if len(matches) < 2 {
		// 只有 0 或 1 个 CSS 文件，无需合并
		return nil
	}

	// 2. 按顺序读取并拼接 CSS 内容
	var combined bytes.Buffer
	stylesDir := filepath.Join(buildDir, DirStyles)
	var cssFiles []string
	for _, match := range matches {
		name := match[1]
		cssFiles = append(cssFiles, name)
		cssPath := filepath.Join(stylesDir, name)
		content, err := os.ReadFile(cssPath)
		if err != nil {
			m.logger.Warn("读取 CSS 文件失败，跳过", "file", name, "error", err)
			continue
		}
		combined.Write(content)
		combined.WriteByte('\n')
	}

	if combined.Len() == 0 {
		return fmt.Errorf("没有读取到任何 CSS 内容")
	}

	// 3. 用 tdewolff/minify 压缩
	minifier := minify.New()
	minifier.AddFunc("text/css", css.Minify)

	minified, err := minifier.Bytes("text/css", combined.Bytes())
	if err != nil {
		// 压缩失败，使用未压缩版本
		m.logger.Warn("CSS 压缩失败，使用未压缩版本", "error", err)
		minified = combined.Bytes()
	}

	// 4. 写入 main.bundle.css
	bundlePath := filepath.Join(stylesDir, FileMainBundleCSS)
	if err := os.WriteFile(bundlePath, minified, 0644); err != nil {
		return fmt.Errorf("写入 %s 失败: %w", FileMainBundleCSS, err)
	}
	m.logger.Info(fmt.Sprintf("CSS 合并压缩完成: %d 个文件 → %s (%d bytes)", len(cssFiles), FileMainBundleCSS, len(minified)))

	// 5. 遍历 buildDir 下所有 .html 文件，替换 CSS 引用
	// 构建匹配连续 <link> 标签块的正则
	var linkPatterns []string
	for _, name := range cssFiles {
		linkPatterns = append(linkPatterns, fmt.Sprintf(`\s*<link\s+rel="stylesheet"\s+href="/styles/%s"\s*/?>`, regexp.QuoteMeta(name)))
	}
	blockPattern := strings.Join(linkPatterns, `\s*`)
	blockRe, err := regexp.Compile(blockPattern)
	if err != nil {
		return fmt.Errorf("构建替换正则失败: %w", err)
	}

	replacement := "\n" + `<link rel="stylesheet" href="/styles/` + FileMainBundleCSS + `">`

	return filepath.Walk(buildDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".html") {
			return err
		}
		htmlContent, err := os.ReadFile(path)
		if err != nil {
			return nil // 跳过读取失败的文件
		}
		newContent := blockRe.ReplaceAll(htmlContent, []byte(replacement))
		if !bytes.Equal(htmlContent, newContent) {
			_ = os.WriteFile(path, newContent, 0644)
		}
		return nil
	})
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	// Check if destination exists and is up to date
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if dstInfo, err := os.Stat(dst); err == nil {
		if dstInfo.Size() == srcInfo.Size() && !dstInfo.ModTime().Before(srcInfo.ModTime()) {
			return nil // Skip copy
		}
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Close()
}
