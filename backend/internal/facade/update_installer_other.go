//go:build !darwin

package facade

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/minio/selfupdate"
)

// installAndRelaunch Windows / Linux 下的原地替换
//
//	Windows: 下载的是 .exe（便携版），直接用 selfupdate.Apply 替换当前 exe
//	Linux:   下载的是 AppImage 或单二进制，同上
//
// selfupdate 负责：写 tmp → atomic rename → 失败时回滚。
//
// 替换完成后 spawn 新可执行并 detach，让父进程被 Wails 正常退出。
func installAndRelaunch(downloadedPath, assetName string) error {
	_ = assetName // 目前非 macOS 直接替换当前 exe，无需区分 asset 类型

	f, err := os.Open(downloadedPath)
	if err != nil {
		return fmt.Errorf("打开下载文件失败: %w", err)
	}
	defer f.Close()

	if err := selfupdate.Apply(f, selfupdate.Options{}); err != nil {
		// selfupdate 内部已尝试回滚
		return fmt.Errorf("替换可执行文件失败: %w", err)
	}

	// 找到新 exe 路径并拉起来
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("无法获取可执行路径: %w", err)
	}
	// Linux AppImage 需要可执行权限
	if runtime.GOOS == "linux" {
		_ = os.Chmod(exe, 0o755)
	}

	// 清理下载文件
	_ = os.Remove(downloadedPath)

	// 尝试重启
	return launchDetached(exe)
}

func launchDetached(path string) error {
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err == nil {
			path = abs
		}
	}
	cmd := exec.Command(path)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = detachAttrsOther()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("重启失败: %w", err)
	}
	// 对于 Windows 新进程已经独立
	if strings.EqualFold(runtime.GOOS, "windows") {
		_ = cmd.Process.Release()
	}
	return nil
}
