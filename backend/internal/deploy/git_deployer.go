package deploy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gridea-pro/backend/internal/domain"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type GitProvider struct{}

func NewGitProvider() *GitProvider {
	return &GitProvider{}
}

func (p *GitProvider) Deploy(ctx context.Context, outputDir string, setting *domain.Setting, logger LogFunc) error {
	logger("Preparing git repository...")

	// 3.1 Initialize or Open Git repo
	var r *git.Repository
	r, err := git.PlainOpen(outputDir)
	if err == git.ErrRepositoryNotExists {
		logger("Initializing new git repository in output directory...")
		r, err = git.PlainInit(outputDir, false)
		if err != nil {
			return fmt.Errorf("failed to init git: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to open git repo: %w", err)
	}

	// Read credentials
	token := setting.Token()
	tokenUser := setting.TokenUsername()
	if tokenUser == "" {
		tokenUser = setting.Username()
	}

	if tokenUser == "" || token == "" {
		return fmt.Errorf(domain.ErrGitTokenMissing)
	}

	// Prepare remote url
	repoUrl := setting.Repository()
	if repoUrl == "" {
		return fmt.Errorf(domain.ErrRepoNotConfigured)
	}

	repoUrl = strings.TrimPrefix(repoUrl, "https://")
	repoUrl = strings.TrimPrefix(repoUrl, "http://")
	repoUrl = strings.TrimPrefix(repoUrl, "git@github.com:")
	repoUrl = strings.TrimPrefix(repoUrl, "git@gitee.com:")
	repoUrl = strings.TrimPrefix(repoUrl, "git@e.coding.net:")

	if !strings.Contains(repoUrl, "/") {
		switch setting.Platform {
		case "github":
			repoUrl = fmt.Sprintf("github.com/%s/%s", setting.Username(), repoUrl)
		case "gitee":
			repoUrl = fmt.Sprintf("gitee.com/%s/%s", setting.Username(), repoUrl)
		case "coding":
			repoUrl = fmt.Sprintf("e.coding.net/%s/%s", setting.Username(), repoUrl)
		}
	} else {
		switch setting.Platform {
		case "github":
			if !strings.Contains(repoUrl, "github.com") {
				repoUrl = fmt.Sprintf("github.com/%s", repoUrl)
			}
		case "gitee":
			if !strings.Contains(repoUrl, "gitee.com") {
				repoUrl = fmt.Sprintf("gitee.com/%s", repoUrl)
			}
		case "coding":
			if !strings.Contains(repoUrl, "e.coding.net") {
				repoUrl = fmt.Sprintf("e.coding.net/%s", repoUrl)
			}
		}
	}

	if (setting.Platform == "github" || setting.Platform == "gitee" || setting.Platform == "coding") && !strings.HasSuffix(repoUrl, ".git") {
		repoUrl += ".git"
	}

	// 3.2 Set Remote Origin
	safeRemoteUrl := fmt.Sprintf("https://%s", repoUrl)
	logger("Configuring remote origin...")
	_ = r.DeleteRemote("origin")
	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{safeRemoteUrl},
	})
	if err != nil {
		return fmt.Errorf("failed to set remote origin: %w", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get git worktree: %w", err)
	}

	// 3.3 Ignore unnecessary files
	gitignorePath := filepath.Join(outputDir, ".gitignore")
	_ = os.WriteFile(gitignorePath, []byte(".DS_Store\nthumbnails/\n.gitignore\n"), 0644)

	// 3.4 Add CNAME if configured
	if setting.CNAME() != "" {
		cnamePath := filepath.Join(outputDir, "CNAME")
		_ = os.WriteFile(cnamePath, []byte(setting.CNAME()), 0644)
		logger(fmt.Sprintf("Generated CNAME file: %s", setting.CNAME()))
	}

	// 3.5 Add all files
	logger("Adding files to commit...")
	err = w.AddWithOptions(&git.AddOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	// 3.6 Commit
	logger("Committing changes...")
	commitMsg := fmt.Sprintf("Deployed by Gridea Pro: %s", time.Now().Format("2006-01-02 15:04:05"))
	email := setting.Email()
	if email == "" {
		email = "gridea-pro@deploy.local"
	}
	username := setting.Username()
	if username == "" {
		username = "Gridea Pro Deployer"
	}

	commitHash, commitErr := w.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  username,
			Email: email,
			When:  time.Now(),
		},
	})

	if commitErr == git.ErrEmptyCommit {
		logger("No changes added to commit, proceeding to push if remote needs update...")
	} else if commitErr != nil {
		return fmt.Errorf("failed to commit: %w", commitErr)
	} else {
		logger(fmt.Sprintf("Committed successfully: %s", commitHash.String()[:7]))
	}

	// 3.7 Define Target Branch
	branch := setting.Branch()
	if branch == "" {
		branch = "gh-pages"
	}

	// 获取当前的本地分支（通常是 master 或 main）
	headRef, err := r.Head()
	if err != nil {
		return fmt.Errorf("failed to get head ref: %w", err)
	}

	// 3.8 Push (修复核心 Bug 的地方)
	logger(fmt.Sprintf("Pushing to remote %s branch (this might take a while)...", branch))

	// 定义 RefSpec：将本地的当前 HEAD 强制推送到远程的指定 branch
	// 格式：+refs/heads/本地分支:refs/heads/远程分支 (+号代表强制推送 Force)
	refSpecStr := fmt.Sprintf("+%s:refs/heads/%s", headRef.Name().String(), branch)

	pushOptions := &git.PushOptions{
		RemoteName: "origin",
		Auth: &http.BasicAuth{
			Username: tokenUser,
			Password: token,
		},
		RefSpecs: []config.RefSpec{config.RefSpec(refSpecStr)},
		Force:    true,
	}
	if setting.ProxyEnabled && setting.ProxyURL != "" {
		pushOptions.ProxyOptions = transport.ProxyOptions{URL: setting.ProxyURL}
	}

	err = r.PushContext(ctx, pushOptions)

	if err == git.NoErrAlreadyUpToDate {
		logger("Remote is already up-to-date!")
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to push to remote: %w", err)
	}

	logger("Deployment successful!")
	return nil
}
