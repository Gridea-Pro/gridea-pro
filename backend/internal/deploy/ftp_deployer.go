package deploy

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"gridea-pro/backend/internal/domain"

	"github.com/jlaffaye/ftp"
)

// FtpProvider 实现了 FTP 文件上传部署策略
type FtpProvider struct{}

// NewFtpProvider 创建 FtpProvider
func NewFtpProvider() *FtpProvider {
	return &FtpProvider{}
}

// FTP 模式常量：对应 setting.ftpMode 字段。
const (
	ftpModePlain         = "ftp"            // 明文（不安全），仅为兼容老配置保留
	ftpModeExplicitTLS   = "ftps-explicit"  // 推荐：先建立 TCP，再 AUTH TLS 升级
	ftpModeImplicitTLS   = "ftps-implicit"  // 首包就是 TLS（通常 990 端口）
	plainFTPInsecureWarn = "⚠️ 警告：当前使用明文 FTP。用户名、密码和所有上传内容将以明文形式在网络上传输，\n" +
		"   任何可以嗅探链路的中间节点（公共 WiFi / 运营商 / 被攻陷的路由器）都能直接获取。\n" +
		"   强烈建议切换到 SFTP 或 FTPS（在平台设置里选择 \"ftps-explicit\"）。"
)

// Deploy 实现 Provider 接口
// 流程：FTP 连接 → 登录 → 清理远程目录 → 上传文件
func (p *FtpProvider) Deploy(ctx context.Context, outputDir string, setting *domain.Setting, logger LogFunc) error {
	logger("🚀 开始准备 FTP 部署...")

	// 1. 验证配置
	server := setting.Server()
	if server == "" {
		return fmt.Errorf(domain.ErrSftpConfigMissing)
	}

	username := setting.Username()
	if username == "" {
		return fmt.Errorf(domain.ErrSftpConfigMissing)
	}

	password := setting.Password()
	if password == "" {
		return fmt.Errorf(domain.ErrSftpConfigMissing)
	}

	port := 21
	if p := setting.Port(); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			port = v
		}
	}

	remotePath := setting.RemotePath()
	if remotePath == "" {
		remotePath = "/"
	}

	// 2. FTP 连接
	addr := fmt.Sprintf("%s:%d", server, port)
	mode := setting.FtpMode()
	if mode == "" {
		mode = ftpModePlain
	}
	logger(fmt.Sprintf("正在连接 %s (模式: %s) ...", addr, mode))

	dialOpts := []ftp.DialOption{ftp.DialWithTimeout(15 * time.Second)}
	switch mode {
	case ftpModeExplicitTLS, ftpModeImplicitTLS:
		tlsCfg := &tls.Config{
			ServerName:         server,
			InsecureSkipVerify: setting.AllowInsecureTLS(),
		}
		if tlsCfg.InsecureSkipVerify {
			logger("⚠️ 已启用\"允许不安全 TLS\"，服务端证书不会被校验，仅建议 NAS 自签场景使用")
		}
		if mode == ftpModeExplicitTLS {
			dialOpts = append(dialOpts, ftp.DialWithExplicitTLS(tlsCfg))
		} else {
			dialOpts = append(dialOpts, ftp.DialWithTLS(tlsCfg))
		}
	default:
		// 明文 FTP：给用户醒目警告，凭证与内容将以明文传输
		logger(plainFTPInsecureWarn)
	}

	conn, err := ftp.Dial(addr, dialOpts...)
	if err != nil {
		return fmt.Errorf("FTP 连接失败: %w", err)
	}
	defer conn.Quit()

	// 3. 登录
	if err := conn.Login(username, password); err != nil {
		return fmt.Errorf("FTP 登录失败: %w", err)
	}

	logger("FTP 连接成功")

	// 4. 清理远程目录
	logger(fmt.Sprintf("正在清理远程目录: %s", remotePath))
	p.cleanRemoteDir(conn, remotePath)

	// 确保远程目录存在
	_ = conn.MakeDir(remotePath)

	// 5. 上传文件
	fileCount := 0
	err = filepath.Walk(outputDir, func(localPath string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
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
		remoteFile := path.Join(remotePath, filepath.ToSlash(relPath))

		// 创建远程目录
		remoteDir := path.Dir(remoteFile)
		p.mkdirAll(conn, remoteDir)

		// 上传文件
		if err := p.uploadFile(conn, localPath, remoteFile); err != nil {
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

	logger(fmt.Sprintf("✅ FTP 部署成功！共上传 %d 个文件到 %s", fileCount, remotePath))
	return nil
}

// cleanRemoteDir 清理远程目录下的所有文件和子目录
func (p *FtpProvider) cleanRemoteDir(conn *ftp.ServerConn, remotePath string) {
	entries, err := conn.List(remotePath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.Name == "." || entry.Name == ".." {
			continue
		}
		fullPath := path.Join(remotePath, entry.Name)
		if entry.Type == ftp.EntryTypeFolder {
			p.cleanRemoteDir(conn, fullPath)
			_ = conn.RemoveDir(fullPath)
		} else {
			_ = conn.Delete(fullPath)
		}
	}
}

// mkdirAll 递归创建远程目录
func (p *FtpProvider) mkdirAll(conn *ftp.ServerConn, dir string) {
	if dir == "/" || dir == "." {
		return
	}
	// 尝试切换到目录，如果成功说明已存在
	if err := conn.ChangeDir(dir); err == nil {
		// 切回根目录
		_ = conn.ChangeDir("/")
		return
	}
	// 递归创建父目录
	parent := path.Dir(dir)
	p.mkdirAll(conn, parent)
	_ = conn.MakeDir(dir)
}

// uploadFile 上传单个文件
func (p *FtpProvider) uploadFile(conn *ftp.ServerConn, localPath, remotePath string) error {
	local, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer local.Close()

	return conn.Stor(remotePath, io.Reader(local))
}
