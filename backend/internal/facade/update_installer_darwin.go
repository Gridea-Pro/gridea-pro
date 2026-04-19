//go:build darwin

package facade

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func detachAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}

// installAndRelaunch 在 macOS 上替换 .app 并重启
//
// 流程（用一个 detached 的 shell 辅助脚本完成，避免 Gridea Pro 自身在运行时替换
// 包含自己的二进制 / bundle）：
//  1. 写一个临时 shell 脚本，该脚本：
//     a. 等待当前进程 PID 退出
//     b. rm -rf 旧 .app
//     c. 把新 .app 移动到原路径
//     d. codesign --force --deep --sign - 重新 ad-hoc 签名
//     e. open 新的 .app
//  2. 以 nohup + & 启动该脚本
//  3. 返回，由调用方触发 wailsRuntime.Quit
//
// 约定：下载下来的 asset 是 .zip，根目录里直接一个 `Gridea Pro.app`。
func installAndRelaunch(downloadedPath, assetName string) error {
	if !strings.HasSuffix(strings.ToLower(assetName), ".zip") {
		return fmt.Errorf("macOS 仅支持 .zip 资源，收到 %q", assetName)
	}

	// 1. 定位当前 .app 路径
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("无法获取自身路径: %w", err)
	}
	// macOS 下 exe = .../Gridea Pro.app/Contents/MacOS/Gridea Pro
	appBundle := filepath.Dir(filepath.Dir(filepath.Dir(exe)))
	if !strings.HasSuffix(appBundle, ".app") {
		return fmt.Errorf("未运行在 .app bundle 中 (current: %s)", appBundle)
	}

	// 2. 解压 zip 到临时目录
	stageDir, err := os.MkdirTemp("", "gridea-pro-stage-*")
	if err != nil {
		return fmt.Errorf("创建解压目录失败: %w", err)
	}
	if err := unzip(downloadedPath, stageDir); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	// 找到解压出的 .app（通常根目录就是 xxx.app）
	newApp, err := findAppInDir(stageDir)
	if err != nil {
		return err
	}

	// 3. 写辅助脚本
	pid := os.Getpid()
	script := fmt.Sprintf(`#!/bin/bash
set -e

# 等待主进程退出
PID=%d
for i in {1..60}; do
    if ! kill -0 "$PID" 2>/dev/null; then break; fi
    sleep 0.2
done

OLD_APP=%q
NEW_APP=%q

rm -rf "$OLD_APP"
mv "$NEW_APP" "$OLD_APP"

# 重新打 ad-hoc 签名（避免原 .app 带签名被替换后失效）
codesign --force --deep --sign - "$OLD_APP" >/dev/null 2>&1 || true

# 清理 staging 目录
rm -rf %q

# 启动新版
open "$OLD_APP"
`, pid, appBundle, newApp, stageDir)

	scriptPath := filepath.Join(os.TempDir(), fmt.Sprintf("gridea-pro-update-%d.sh", pid))
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		return fmt.Errorf("写更新脚本失败: %w", err)
	}

	// 4. 用 nohup 脱离父进程启动
	cmd := exec.Command("/bin/bash", scriptPath)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	// 让子进程独立于 Gridea Pro 的进程组
	cmd.SysProcAttr = detachAttrs()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动更新脚本失败: %w", err)
	}
	// 不等待，主进程稍后被 Quit
	_ = cmd.Process.Release()
	return nil
}

func unzip(src, dst string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		path := filepath.Join(dst, f.Name)
		// zip slip 防护
		if !strings.HasPrefix(path, filepath.Clean(dst)+string(os.PathSeparator)) {
			return fmt.Errorf("非法路径 %q", path)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, f.Mode()); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		in, err := f.Open()
		if err != nil {
			_ = out.Close()
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			_ = out.Close()
			_ = in.Close()
			return err
		}
		_ = out.Close()
		_ = in.Close()
	}
	return nil
}

func findAppInDir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() && strings.HasSuffix(e.Name(), ".app") {
			return filepath.Join(dir, e.Name()), nil
		}
	}
	// 也允许嵌套一层（例如 zip 里是 Gridea-Pro/Gridea Pro.app）
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		sub := filepath.Join(dir, e.Name())
		subEntries, err := os.ReadDir(sub)
		if err != nil {
			continue
		}
		for _, se := range subEntries {
			if se.IsDir() && strings.HasSuffix(se.Name(), ".app") {
				return filepath.Join(sub, se.Name()), nil
			}
		}
	}
	return "", fmt.Errorf("解压内容中未找到 .app bundle")
}
