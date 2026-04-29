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
type SftpProvider struct{}

// NewSftpProvider 创建 SftpProvider
func NewSftpProvider() *SftpProvider {
	return &SftpProvider{}
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

	// 3. SSH 连接
	sshConfig := &ssh.ClientConfig{
		User:            username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
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

	// 5. 原子切换策略（issue #40）：
	//    a. 上传到 <remotePath>.new-<ts>（staging），旧站点完好无损
	//    b. 全部成功 → 把旧目录 rename 到 <remotePath>.old-<ts>，再把 staging rename 为 remotePath
	//    c. 失败 → 清理 staging，旧站点保持原样，访客零感知
	//
	// 好处：上传期间访客全程访问旧版本；中途断网不毁站；失败可直接 dropping staging 重新开始。
	ts := time.Now().Format("20060102-150405")
	stagingPath := remotePath + ".new-" + ts
	backupPath := remotePath + ".old-" + ts

	logger(fmt.Sprintf("正在上传到暂存目录: %s（旧站点保持可用）", stagingPath))
	if err := client.MkdirAll(stagingPath); err != nil {
		return fmt.Errorf("创建暂存目录失败: %w", err)
	}

	// 6. 上传文件到 staging
	fileCount := 0
	uploadErr := filepath.Walk(outputDir, func(localPath string, info os.FileInfo, err error) error {
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
		// 远程路径始终使用 Unix 风格；上传到 staging 目录
		remoteFile := path.Join(stagingPath, filepath.ToSlash(relPath))

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

	if uploadErr != nil {
		// 清理 staging（尽力而为），旧站点保持原样
		logger(fmt.Sprintf("上传失败，正在清理暂存目录 %s...", stagingPath))
		p.cleanRemoteDir(client, stagingPath)
		_ = client.RemoveDirectory(stagingPath)
		return uploadErr
	}

	// 7. 原子切换：rename 旧→backup、rename staging→remote
	logger("上传完成，正在切换到新版本（原子 rename）...")

	// 旧目录可能不存在（首次部署），尝试 rename 失败不视为错误
	oldExists := false
	if _, err := client.Stat(remotePath); err == nil {
		oldExists = true
	}
	if oldExists {
		if err := client.Rename(remotePath, backupPath); err != nil {
			// 清理 staging，保留旧目录
			p.cleanRemoteDir(client, stagingPath)
			_ = client.RemoveDirectory(stagingPath)
			return fmt.Errorf("重命名旧目录失败: %w", err)
		}
	}
	if err := client.Rename(stagingPath, remotePath); err != nil {
		// 尝试把 backup 改回来，尽量不让站点挂掉
		if oldExists {
			_ = client.Rename(backupPath, remotePath)
		}
		p.cleanRemoteDir(client, stagingPath)
		_ = client.RemoveDirectory(stagingPath)
		return fmt.Errorf("重命名新目录失败: %w", err)
	}

	// 8. 清理旧备份（best-effort，失败不影响部署成功）
	if oldExists {
		p.cleanRemoteDir(client, backupPath)
		_ = client.RemoveDirectory(backupPath)
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
