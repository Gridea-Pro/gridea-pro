package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
)

func TestNodePolyfills(t *testing.T) {
	// 1. Setup temporary theme directory
	tmpDir, err := os.MkdirTemp("", "gridea-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	appDir := tmpDir
	themeName := "test-theme"
	themeDir := filepath.Join(appDir, "themes", themeName)
	templatesDir := filepath.Join(themeDir, "templates")

	err = os.MkdirAll(templatesDir, 0755)
	assert.NoError(t, err)

	// Create dummy files
	err = os.WriteFile(filepath.Join(themeDir, "style.css"), []byte("body {}"), 0644)
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(templatesDir, "index.ejs"), []byte("<h1>Index</h1>"), 0644)
	assert.NoError(t, err)

	// 2. Initialize Renderer
	// config := RenderConfig{...} // Not needed
	// renderer := NewEjsRenderer(config) // Not needed for polyfill test
	vm := goja.New()

	// 3. Setup Polyfills
	SetupNodePolyfills(vm, themeDir)

	// 4. Test Cases
	t.Run("process.cwd", func(t *testing.T) {
		res, err := vm.RunString(`process.cwd()`)
		assert.NoError(t, err)
		assert.Equal(t, themeDir, res.String())
	})

	t.Run("fs.existsSync absolute", func(t *testing.T) {
		script := `
			var fs = require('fs');
			fs.existsSync('` + filepath.Join(themeDir, "style.css") + `');
		`
		res, err := vm.RunString(script)
		assert.NoError(t, err)
		assert.Equal(t, true, res.ToBoolean())
	})

	t.Run("fs.existsSync relative", func(t *testing.T) {
		script := `
			var fs = require('fs');
			fs.existsSync('style.css'); // Should resolve to themeDir/style.css
		`
		res, err := vm.RunString(script)
		assert.NoError(t, err)
		assert.Equal(t, true, res.ToBoolean())
	})

	t.Run("fs.readFileSync", func(t *testing.T) {
		script := `
			var fs = require('fs');
			fs.readFileSync('style.css', 'utf8');
		`
		res, err := vm.RunString(script)
		assert.NoError(t, err)
		assert.Equal(t, "body {}", res.String())
	})

	t.Run("path.resolve", func(t *testing.T) {
		script := `
			var path = require('path');
			path.resolve('templates', 'index.ejs');
		`
		res, err := vm.RunString(script)
		assert.NoError(t, err)
		expected := filepath.Join(themeDir, "templates", "index.ejs")
		assert.Equal(t, expected, res.String())
	})

	t.Run("path.join", func(t *testing.T) {
		script := `
			var path = require('path');
			path.join('a', 'b');
		`
		res, err := vm.RunString(script)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join("a", "b"), res.String())
	})
}
