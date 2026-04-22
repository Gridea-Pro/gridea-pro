package deploy

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gridea-pro/backend/internal/domain"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SftpProvider 实现了 SFTP 文件上传部署策略
type SftpProvider struct {
	// knownHostsPath 用于 TOFU 形式的 HostKey 校验；为空时走内存级 TOFU（降级）。
	// 通过 NewSftpProviderWithKnownHosts 注入，生产路径应为 AppConfigDir/known_hosts。
	knownHostsPath string
}

// NewSftpProvider 创建默认 SftpProvider（无 known_hosts 持久化，仅进程内 TOFU）。
// 生产路径请用 NewSftpProviderWithKnownHosts。
func NewSftpProvider() *SftpProvider {
	return &SftpProvider{}
}

// NewSftpProviderWithKnownHosts 注入 known_hosts 文件路径，启用跨会话的 HostKey 校验。
func NewSftpProviderWithKnownHosts(knownHostsPath string) *SftpProvider {
	return &SftpProvider{knownHostsPath: knownHostsPath}
}

// Deploy 实现 Provider 接口
// 流程：SSH 连接 → SFTP 客户端 → 清理远程目录 → 上传文件
func (p *SftpProvider) Deploy(ctx context.Context, outputDir string, setting *domain.Setting, logger LogFunc) error {
	logger("🚀 开始准备 SFTP 部署...")

	// 1. 验证配置
	server := setting.Server()
	if server == "" {
		return fmt.Errorf(domain.ErrSftpConfigMissing)
	}

	username := setting.Username()
	if username == "" {
		return fmt.Errorf(domain.ErrSftpConfigMissing)
	}

	port := 22
	if p := setting.Port(); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			port = v
		}
	}

	remotePath := setting.RemotePath()
	if remotePath == "" {
		remotePath = "/var/www/html"
	}

	// 2. 构建 SSH 认证方式
	authMethods, err := p.buildAuthMethods(setting)
	if err != nil {
		return err
	}
	if len(authMethods) == 0 {
		return fmt.Errorf(domain.ErrSftpConfigMissing)
	}

	// 3. SSH 连接：使用 TOFU 形式的 HostKey 校验替代 InsecureIgnoreHostKey。
	//    首次连接会把指纹写入 known_hosts；后续任何指纹变化都会被硬拒绝（防 MITM）。
	sshConfig := &ssh.ClientConfig{
		User:            username,
		Auth:            authMethods,
		HostKeyCallback: TrustOnFirstUseHostKeyCallback(p.knownHostsPath, logger),
		Timeout:         15 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", server, port)
	logger(fmt.Sprintf("正在连接 %s ...", addr))

	conn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer conn.Close()

	logger("SSH 连接成功")

	// 4. 创建 SFTP 客户端
	client, err := sftp.NewClient(conn)
	if err != nil {
		return fmt.Errorf("SFTP 客户端创建失败: %w", err)
	}
	defer client.Close()

	// 5. 清理远程目录
	logger(fmt.Sprintf("正在清理远程目录: %s", remotePath))
	p.cleanRemoteDir(client, remotePath)

	// 确保远程目录存在
	if err := client.MkdirAll(remotePath); err != nil {
		return fmt.Errorf("创建远程目录失败: %w", err)
	}

	// 6. 上传文件
	fileCount := 0
	err = filepath.Walk(outputDir, func(localPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过无关目录和文件
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == ".github" {
				return filepath.SkipDir
			}
			return nil
		}
		name := info.Name()
		if name == ".DS_Store" || name == ".gitignore" {
			return nil
		}

		// 检查 context 是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		relPath, err := filepath.Rel(outputDir, localPath)
		if err != nil {
			return err
		}
		// 远程路径始终使用 Unix 风格
		remoteFile := path.Join(remotePath, filepath.ToSlash(relPath))

		// 创建远程目录
		remoteDir := path.Dir(remoteFile)
		if err := client.MkdirAll(remoteDir); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %w", remoteDir, err)
		}

		// 上传文件
		if err := p.uploadFile(client, localPath, remoteFile); err != nil {
			return fmt.Errorf("上传 %s 失败: %w", relPath, err)
		}

		fileCount++
		if fileCount%20 == 0 {
			logger(fmt.Sprintf("已上传 %d 个文件...", fileCount))
		}

		return nil
	})

	if err != nil {
		return err
	}

	logger(fmt.Sprintf("✅ SFTP 部署成功！共上传 %d 个文件到 %s", fileCount, remotePath))
	return nil
}

// buildAuthMethods 根据配置构建 SSH 认证方式
func (p *SftpProvider) buildAuthMethods(setting *domain.Setting) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	// 私钥认证
	if pk := setting.PrivateKey(); pk != "" {
		var keyData []byte
		if strings.HasPrefix(pk, "-----BEGIN") {
			// 内联 PEM 内容
			keyData = []byte(pk)
		} else {
			// 文件路径
			data, err := os.ReadFile(pk)
			if err != nil {
				return nil, fmt.Errorf("读取私钥文件失败: %w", err)
			}
			keyData = data
		}

		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %w", err)
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}

	// 密码认证
	if pw := setting.Password(); pw != "" {
		methods = append(methods, ssh.Password(pw))
	}

	return methods, nil
}

// cleanRemoteDir 清理远程目录下的所有文件和子目录
func (p *SftpProvider) cleanRemoteDir(client *sftp.Client, remotePath string) {
	entries, err := client.ReadDir(remotePath)
	if err != nil {
		return // 目录不存在或无法读取，忽略
	}

	for _, entry := range entries {
		fullPath := path.Join(remotePath, entry.Name())
		if entry.IsDir() {
			p.cleanRemoteDir(client, fullPath)
			_ = client.RemoveDirectory(fullPath)
		} else {
			_ = client.Remove(fullPath)
		}
	}
}

// uploadFile 上传单个文件
func (p *SftpProvider) uploadFile(client *sftp.Client, localPath, remotePath string) error {
	local, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer local.Close()

	remote, err := client.Create(remotePath)
	if err != nil {
		return err
	}
	defer remote.Close()

	_, err = io.Copy(remote, local)
	return err
}
