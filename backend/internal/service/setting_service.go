package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gridea-pro/backend/internal/domain"

	gogit "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	gittransport "github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"golang.org/x/crypto/ssh"
)

type SettingService struct {
	repo   domain.SettingRepository
	appDir string
	mu     sync.RWMutex
}

func NewSettingService(appDir string, repo domain.SettingRepository) *SettingService {
	return &SettingService{
		appDir: appDir,
		repo:   repo,
	}
}

func (s *SettingService) SaveAvatar(ctx context.Context, sourcePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	destPath := filepath.Join(s.appDir, "images", "avatar.png")
	return s.copyFile(sourcePath, destPath)
}

func (s *SettingService) SaveFavicon(ctx context.Context, sourcePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	destPath := filepath.Join(s.appDir, "favicon.ico")
	return s.copyFile(sourcePath, destPath)
}

func (s *SettingService) copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}
	return nil
}

func (s *SettingService) GetSetting(ctx context.Context) (domain.Setting, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.repo.GetSetting(ctx)
}

func (s *SettingService) SaveSetting(ctx context.Context, setting domain.Setting) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repo.SaveSetting(ctx, setting)
}

// RemoteDetect 检测远程连接是否正常
func (s *SettingService) RemoteDetect(ctx context.Context, setting domain.Setting) (map[string]interface{}, error) {
	success := false
	message := ""

	switch setting.Platform {
	case "github", "gitee", "coding":
		// 使用 go-git ls-remote 验证认证
		repoUrl := strings.TrimSpace(setting.Repository())
		repoUrl = strings.TrimPrefix(repoUrl, "https://")
		repoUrl = strings.TrimPrefix(repoUrl, "http://")
		repoUrl = strings.TrimPrefix(repoUrl, "git@github.com:")
		repoUrl = strings.TrimPrefix(repoUrl, "git@gitee.com:")

		hostname := "github.com"
		switch setting.Platform {
		case "gitee":
			hostname = "gitee.com"
		case "coding":
			hostname = "e.coding.net"
		}

		if !strings.Contains(repoUrl, "/") {
			repoUrl = fmt.Sprintf("%s/%s/%s", hostname, setting.Username(), repoUrl)
		} else if !strings.Contains(repoUrl, hostname) {
			repoUrl = fmt.Sprintf("%s/%s", hostname, repoUrl)
		}

		if !strings.HasSuffix(repoUrl, ".git") {
			repoUrl += ".git"
		}
		safeUrl := "https://" + repoUrl

		tokenUser := setting.TokenUsername()
		if tokenUser == "" {
			tokenUser = setting.Username()
		}

		listOptions := &gogit.ListOptions{
			Auth: &githttp.BasicAuth{
				Username: tokenUser,
				Password: setting.Token(),
			},
		}
		if setting.ProxyEnabled && setting.ProxyURL != "" {
			listOptions.ProxyOptions = gittransport.ProxyOptions{URL: setting.ProxyURL}
		}

		_, err := gogit.NewRemote(nil, &gitconfig.RemoteConfig{
			Name: "origin",
			URLs: []string{safeUrl},
		}).ListContext(ctx, listOptions)

		if err != nil {
			message = fmt.Sprintf("连接失败: %v", err)
		} else {
			success = true
			message = "Git 仓库连接成功"
		}

	case "vercel":
		// 通过 Vercel API 验证 token
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.vercel.com/v2/user", nil)
		if err != nil {
			message = fmt.Sprintf("无法创建请求: %v", err)
			break
		}
		req.Header.Set("Authorization", "Bearer "+setting.Token())

		vercelClient := http.DefaultClient
		if setting.ProxyEnabled && setting.ProxyURL != "" {
			vercelClient = newHTTPClient(setting.ProxyURL)
		}
		resp, err := vercelClient.Do(req)
		if err != nil {
			message = fmt.Sprintf("连接失败: %v", err)
			break
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			success = true
			message = "Vercel Token 验证成功"
		} else {
			message = fmt.Sprintf("Vercel Token 无效 (HTTP %d)", resp.StatusCode)
		}

	case "netlify":
		siteId := setting.NetlifySiteId()
		token := setting.NetlifyAccessToken()
		if siteId == "" || token == "" {
			message = "Site ID 或 Access Token 为空"
			break
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("https://api.netlify.com/api/v1/sites/%s", siteId), nil)
		if err != nil {
			message = fmt.Sprintf("无法创建请求: %v", err)
			break
		}
		req.Header.Set("Authorization", "Bearer "+token)

		netlifyClient := http.DefaultClient
		if setting.ProxyEnabled && setting.ProxyURL != "" {
			netlifyClient = newHTTPClient(setting.ProxyURL)
		}
		resp, err := netlifyClient.Do(req)
		if err != nil {
			message = fmt.Sprintf("连接失败: %v", err)
			break
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			success = true
			message = "Netlify 连接成功"
		} else {
			message = fmt.Sprintf("Netlify 验证失败 (HTTP %d)", resp.StatusCode)
		}

	case "sftp":
		server := setting.Server()
		if server == "" {
			message = "服务器地址为空"
			break
		}

		sftpPort := 22
		if p := setting.Port(); p != "" {
			if v, err := strconv.Atoi(p); err == nil && v > 0 {
				sftpPort = v
			}
		}

		var authMethods []ssh.AuthMethod
		if pk := setting.PrivateKey(); pk != "" {
			var keyData []byte
			if strings.HasPrefix(pk, "-----BEGIN") {
				keyData = []byte(pk)
			} else {
				var readErr error
				keyData, readErr = os.ReadFile(pk)
				if readErr != nil {
					message = fmt.Sprintf("读取私钥失败: %v", readErr)
					break
				}
			}
			signer, parseErr := ssh.ParsePrivateKey(keyData)
			if parseErr != nil {
				message = fmt.Sprintf("解析私钥失败: %v", parseErr)
				break
			}
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
		if pw := setting.Password(); pw != "" {
			authMethods = append(authMethods, ssh.Password(pw))
		}

		if len(authMethods) == 0 {
			message = "密码和私钥均为空"
			break
		}

		sshConfig := &ssh.ClientConfig{
			User:            setting.Username(),
			Auth:            authMethods,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         10 * time.Second,
		}

		sshConn, dialErr := ssh.Dial("tcp", fmt.Sprintf("%s:%d", server, sftpPort), sshConfig)
		if dialErr != nil {
			message = fmt.Sprintf("SSH 连接失败: %v", dialErr)
			break
		}
		sshConn.Close()
		success = true
		message = "SFTP 连接成功"

	default:
		message = "不支持的平台类型"
	}

	return map[string]interface{}{
		"success": success,
		"message": message,
	}, nil
}
