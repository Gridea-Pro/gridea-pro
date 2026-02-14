package service

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dop251/goja"
)

// AssetManager handles theme and site asset operations
type AssetManager struct {
	appDir             string
	themeConfigService *ThemeConfigService
}

// NewAssetManager creates a new AssetManager instance
func NewAssetManager(appDir string, themeConfigService *ThemeConfigService) *AssetManager {
	return &AssetManager{
		appDir:             appDir,
		themeConfigService: themeConfigService,
	}
}

// CopyThemeAssets 复制主题静态资源
func (m *AssetManager) CopyThemeAssets(buildDir, themeName string) error {
	themePath := filepath.Join(m.appDir, DirThemes, themeName)
	assetsPath := filepath.Join(themePath, DirAssets)
	if _, err := os.Stat(assetsPath); os.IsNotExist(err) {
		return nil
	}

	// 1. 检查并编译 LESS 文件
	// 1. 检查并编译 LESS 文件
	lessPath := filepath.Join(assetsPath, DirStyles, FileMainLess)
	if _, err := os.Stat(lessPath); err == nil {
		// Use compileLess which has optimization
		if err := m.compileLess(lessPath, buildDir); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "警告：LESS 编译失败: %v\n", err)
		}
	}

	// 2. 复制其他静态资源
	destPath := filepath.Join(buildDir)
	return copyDir(assetsPath, destPath)
}

// compileLess 编译 LESS 文件为 CSS
func (m *AssetManager) compileLess(lessPath, buildDir string) error {
	// 输出路径
	cssPath := filepath.Join(buildDir, DirStyles, FileMainCSS)

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(cssPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// Optimization: Check if recompilation is needed
	// If main.css exists and is newer than main.less, skip compilation
	lessInfo, err := os.Stat(lessPath)
	if err == nil {
		cssInfo, err := os.Stat(cssPath)
		if err == nil && cssInfo.ModTime().After(lessInfo.ModTime()) {
			return nil
		}
	}

	// 调用 lessc 命令编译
	cmd := exec.Command("lessc", lessPath, cssPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("lessc 编译失败: %w\n输出: %s", err, string(output))
	}

	// 检查并应用 style-override.js
	// 从 lessPath 推导主题路径
	themePath := filepath.Dir(filepath.Dir(lessPath))
	overridePath := filepath.Join(themePath, FileStyleOverride)
	if _, err := os.Stat(overridePath); err == nil {
		fmt.Fprintln(os.Stderr, "检测到 style-override.js，应用自定义样式...")
		customCSS, err := m.applyStyleOverride(overridePath)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "警告：应用 style-override.js 失败: %v\n", err)
		} else {
			// 读取编译后的 CSS
			cssContent, err := os.ReadFile(cssPath)
			if err != nil {
				return fmt.Errorf("读取 CSS 文件失败: %w", err)
			}

			// 追加自定义 CSS
			cssContent = append(cssContent, []byte("\n"+customCSS)...)
			if err := os.WriteFile(cssPath, cssContent, 0644); err != nil {
				return fmt.Errorf("写入 CSS 文件失败: %w", err)
			}
			fmt.Fprintln(os.Stderr, "✅ 自定义样式应用成功")
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

	return nil
}

// loadThemeCustomConfig 加载主题自定义配置 (Helper for AssetManager)
func (m *AssetManager) loadThemeCustomConfig(themeName string) map[string]interface{} {
	// 使用 ThemeConfigService 加载配置
	config, err := m.themeConfigService.GetFinalConfig(themeName)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "警告：加载主题配置失败，使用空配置: %v\n", err)
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
