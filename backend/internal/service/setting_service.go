package service

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	gogit "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
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
		repoUrl := strings.TrimSpace(setting.Repository)
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
			repoUrl = fmt.Sprintf("%s/%s/%s", hostname, setting.Username, repoUrl)
		} else if !strings.Contains(repoUrl, hostname) {
			repoUrl = fmt.Sprintf("%s/%s", hostname, repoUrl)
		}

		if !strings.HasSuffix(repoUrl, ".git") {
			repoUrl += ".git"
		}
		safeUrl := "https://" + repoUrl

		tokenUser := setting.TokenUsername
		if tokenUser == "" {
			tokenUser = setting.Username
		}

		_, err := gogit.NewRemote(nil, &gitconfig.RemoteConfig{
			Name: "origin",
			URLs: []string{safeUrl},
		}).ListContext(ctx, &gogit.ListOptions{
			Auth: &githttp.BasicAuth{
				Username: tokenUser,
				Password: setting.Token,
			},
		})

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
		req.Header.Set("Authorization", "Bearer "+setting.Token)

		resp, err := http.DefaultClient.Do(req)
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

	default:
		// 对于其他平台（Netlify/SFTP），简单验证配置非空
		if setting.Token != "" || setting.Password != "" || setting.PrivateKey != "" {
			success = true
			message = "配置已保存，凭据不为空"
		} else {
			message = "凭据为空，请检查配置"
		}
	}

	return map[string]interface{}{
		"success": success,
		"message": message,
	}, nil
}
