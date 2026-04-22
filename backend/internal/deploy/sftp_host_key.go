package deploy

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// TrustOnFirstUseHostKeyCallback 构造一个基于 known_hosts 的 SSH HostKeyCallback：
//
//   - 主机在 known_hosts 中且指纹匹配 → 通过
//   - 主机在 known_hosts 中但指纹已变（可能中间人攻击）→ 硬拒绝，给出清晰
//     的错误信息要求用户手工确认后再从 known_hosts 删除旧记录
//   - 主机不在 known_hosts（首次连接）→ 写入 known_hosts 并通过（TOFU），
//     同时通过 logger 告知用户当前信任的指纹，留下审计痕迹
//
// 相比 ssh.InsecureIgnoreHostKey 的无条件信任，此回调至少能阻断"持续性中间人"：
// 攻击者需要在用户首次连接的那一刻就已经介入，而且任何后续替换都会立刻被发现。
// 更严格的"首次连接前端弹窗确认指纹"的流程可在此基础上追加，不影响接口兼容。
//
// knownHostsPath 为空时回退到"内存级 TOFU"（只在当次进程生命周期内记录），
// 主要用于测试与降级场景（例如磁盘权限异常），生产路径应始终传入持久化路径。
func TrustOnFirstUseHostKeyCallback(knownHostsPath string, logger LogFunc) ssh.HostKeyCallback {
	var inMemory sync.Map // host+fingerprint → struct{}
	return func(host string, remote net.Addr, key ssh.PublicKey) error {
		fp := ssh.FingerprintSHA256(key)

		if knownHostsPath == "" {
			// 内存级兜底：当次进程复连同一 host 时检查指纹一致
			return checkOrStoreInMemory(&inMemory, host, fp, logger)
		}

		if err := ensureKnownHostsFile(knownHostsPath); err != nil {
			if logger != nil {
				logger(fmt.Sprintf("known_hosts 文件不可用（%v），降级为内存级 TOFU", err))
			}
			return checkOrStoreInMemory(&inMemory, host, fp, logger)
		}

		cb, err := knownhosts.New(knownHostsPath)
		if err != nil {
			return fmt.Errorf("加载 known_hosts 失败: %w", err)
		}

		kerr := cb(host, remote, key)
		if kerr == nil {
			return nil
		}

		// 区分两种失败：指纹变更（Want 非空，有旧记录）vs 未知主机（Want 为空）
		var keyErr *knownhosts.KeyError
		if errors.As(kerr, &keyErr) && len(keyErr.Want) > 0 {
			return fmt.Errorf(
				"SSH 主机密钥已变更（可能是中间人攻击或服务器重装）：\n"+
					"  host=%s\n"+
					"  新指纹=%s\n"+
					"若确认服务器合法变更，请手工编辑 %s 删除旧记录后重试",
				host, fp, knownHostsPath,
			)
		}

		// 未知主机：TOFU 写入
		if err := appendKnownHost(knownHostsPath, host, key); err != nil {
			return fmt.Errorf("写入 known_hosts 失败: %w", err)
		}
		if logger != nil {
			logger(fmt.Sprintf("🔐 首次连接 %s，已将其指纹 %s 写入 known_hosts", host, fp))
		}
		return nil
	}
}

// checkOrStoreInMemory 进程内记忆式 TOFU：仅用于磁盘不可用的降级路径。
func checkOrStoreInMemory(store *sync.Map, host, fp string, logger LogFunc) error {
	if existing, ok := store.Load(host); ok {
		if existing.(string) == fp {
			return nil
		}
		return fmt.Errorf("本次会话中 %s 的指纹已变更（可能中间人）：旧 %s → 新 %s", host, existing, fp)
	}
	store.Store(host, fp)
	if logger != nil {
		logger(fmt.Sprintf("⚠️ known_hosts 不可用，内存级 TOFU 记录 %s 指纹 %s（重启后失效）", host, fp))
	}
	return nil
}

// ensureKnownHostsFile 确保 known_hosts 文件及其父目录存在，不存在则创建空文件。
func ensureKnownHostsFile(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	return f.Close()
}

// appendKnownHost 以 OpenSSH 兼容的格式追加一条 host 记录。
func appendKnownHost(path, host string, key ssh.PublicKey) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	line := knownhosts.Line([]string{knownhosts.Normalize(host)}, key)
	if _, err := fmt.Fprintln(f, line); err != nil {
		return err
	}
	return nil
}
