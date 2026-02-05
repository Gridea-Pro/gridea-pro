package render

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dop251/goja"
)

// SetupNodePolyfills 注入 Node.js 核心模块模拟
// baseDir: 模拟的 process.cwd()，通常是主题目录
func SetupNodePolyfills(vm *goja.Runtime, baseDir string) {

	// --- 1. Process Module ---
	processObj := vm.NewObject()
	processObj.Set("cwd", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(baseDir)
	})
	processObj.Set("platform", runtime.GOOS)
	processObj.Set("env", vm.NewObject())
	processObj.Set("argv", []string{})
	processObj.Set("version", "v14.0.0") // Mock version
	vm.Set("process", processObj)

	// Helper to resolve paths relative to baseDir if they are relative
	resolvePath := func(p string) string {
		if filepath.IsAbs(p) {
			return filepath.Clean(p)
		}
		return filepath.Join(baseDir, p)
	}

	// --- 2. Console Module ---
	consoleObj := vm.NewObject()
	consoleObj.Set("log", func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.Export()
		}
		fmt.Println(args...)
		return goja.Undefined()
	})
	consoleObj.Set("error", func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.Export()
		}
		fmt.Printf("JS Error: %v\n", args...)
		return goja.Undefined()
	})
	consoleObj.Set("warn", func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.Export()
		}
		fmt.Printf("JS Warn: %v\n", args...)
		return goja.Undefined()
	})
	vm.Set("console", consoleObj)

	// --- 3. Path Module ---
	pathObj := vm.NewObject()
	pathObj.Set("join", func(call goja.FunctionCall) goja.Value {
		parts := make([]string, len(call.Arguments))
		for i, arg := range call.Arguments {
			parts[i] = arg.String()
		}
		return vm.ToValue(filepath.Join(parts...))
	})
	pathObj.Set("resolve", func(call goja.FunctionCall) goja.Value {
		// Node.js path.resolve resolves right-to-left
		resolved := ""
		for i := len(call.Arguments) - 1; i >= 0; i-- {
			p := call.Argument(i).String()
			if p == "" {
				continue
			}
			if resolved == "" {
				resolved = p
			} else {
				resolved = filepath.Join(p, resolved)
			}
			if filepath.IsAbs(resolved) {
				return vm.ToValue(filepath.Clean(resolved))
			}
		}
		// If still relative, resolve against baseDir (CWD)
		return vm.ToValue(filepath.Join(baseDir, resolved))
	})
	pathObj.Set("dirname", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(filepath.Dir(call.Argument(0).String()))
	})
	pathObj.Set("basename", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		ext := ""
		if len(call.Arguments) > 1 {
			ext = call.Argument(1).String()
		}
		base := filepath.Base(path)
		if ext != "" && strings.HasSuffix(base, ext) {
			return vm.ToValue(base[:len(base)-len(ext)])
		}
		return vm.ToValue(base)
	})
	pathObj.Set("extname", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(filepath.Ext(call.Argument(0).String()))
	})
	pathObj.Set("isAbsolute", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(filepath.IsAbs(call.Argument(0).String()))
	})
	pathObj.Set("sep", string(filepath.Separator))
	pathObj.Set("delimiter", string(os.PathListSeparator))
	vm.Set("path", pathObj)

	// --- 4. FS Module ---
	fsObj := vm.NewObject()

	fsObj.Set("existsSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		_, err := os.Stat(resolvePath(path))
		return vm.ToValue(err == nil)
	})

	fsObj.Set("readFileSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		data, err := os.ReadFile(resolvePath(path))
		if err != nil {
			panic(vm.ToValue(fmt.Sprintf("ENOENT: no such file or directory, open '%s'", path)))
		}
		// Encoding handling: EJS usually requests 'utf8'. We just always return string for simplicity in templates.
		// If explicit null encoding was requested (unlikely for EJS include), we might need to return ArrayBuffer,
		// but Goja string is safe for text.
		return vm.ToValue(string(data))
	})

	fsObj.Set("statSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		info, err := os.Stat(resolvePath(path))
		if err != nil {
			panic(vm.ToValue(fmt.Sprintf("ENOENT: no such file or directory, stat '%s'", path)))
		}

		stat := vm.NewObject()
		stat.Set("isFile", func(call goja.FunctionCall) goja.Value { return vm.ToValue(!info.IsDir()) })
		stat.Set("isDirectory", func(call goja.FunctionCall) goja.Value { return vm.ToValue(info.IsDir()) })
		return stat
	})

	fsObj.Set("lstatSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		info, err := os.Lstat(resolvePath(path))
		if err != nil {
			panic(vm.ToValue(fmt.Sprintf("ENOENT: no such file or directory, lstat '%s'", path)))
		}
		stat := vm.NewObject()
		stat.Set("isFile", func(call goja.FunctionCall) goja.Value { return vm.ToValue(!info.IsDir()) })
		stat.Set("isDirectory", func(call goja.FunctionCall) goja.Value { return vm.ToValue(info.IsDir()) })
		return stat
	})

	// realpathSync is also often used
	fsObj.Set("realpathSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		resolved, err := filepath.EvalSymlinks(resolvePath(path))
		if err != nil {
			// Fallback if not found or error, just return resolved path
			return vm.ToValue(resolvePath(path))
		}
		return vm.ToValue(resolved)
	})

	vm.Set("fs", fsObj)

	// --- 5. Global Require ---
	vm.Set("require", func(call goja.FunctionCall) goja.Value {
		moduleName := call.Argument(0).String()
		switch moduleName {
		case "fs":
			return fsObj
		case "path":
			return pathObj
		case "process":
			return processObj
		case "console":
			return consoleObj
		default:
			return goja.Undefined()
		}
	})
}
