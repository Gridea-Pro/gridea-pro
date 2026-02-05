package comment

import (
	"context"
	"encoding/json"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"net/http"
	"time"
)

// GitHubProvider 处理 Gitalk 和 Giscus (REST API fallback)
type GitHubProvider struct {
	Owner        string
	Repo         string
	ClientID     string
	ClientSecret string
	AccessToken  string // 可选，如果有 Personal Access Token 更好
}

func NewGitHubProvider(owner, repo, clientID, clientSecret string) *GitHubProvider {
	return &GitHubProvider{
		Owner:        owner,
		Repo:         repo,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

// GitHub Issue Comment
type githubComment struct {
	ID        int64      `json:"id"`
	Body      string     `json:"body"`
	User      githubUser `json:"user"`
	CreatedAt time.Time  `json:"created_at"`
	HtmlUrl   string     `json:"html_url"`
	IssueUrl  string     `json:"issue_url"`
}

type githubUser struct {
	Login     string `json:"login"`
	AvatarUrl string `json:"avatar_url"`
	HtmlUrl   string `json:"html_url"`
}

func (p *GitHubProvider) GetComments(ctx context.Context, articleID string) ([]domain.Comment, error) {
	// 获取指定文章的评论需要先找到对应的 Issue，逻辑较复杂，暂时只实现 GetRecentComments
	return nil, fmt.Errorf("GitHubProvider GetComments not implemented yet (requires Issue lookup strategy)")
}

func (p *GitHubProvider) GetRecentComments(ctx context.Context, limit int) ([]domain.Comment, error) {
	// https://docs.github.com/en/rest/issues/comments?apiVersion=2022-11-28#list-issue-comments-for-a-repository
	// GET /repos/{owner}/{repo}/issues/comments

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/comments?sort=created&direction=desc&per_page=%d", p.Owner, p.Repo, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	// 如果配置了 ClientID/Secret，目前 GitHub API 主要通过 Token 鉴权。
	// 对于公开仓库的公开评论，可以直接访问，但有速率限制。
	// Gitalk 配置中有 ClientID/Secret，主要用于前端 OAuth 换取 Token。
	// 后端如果没有 Token，只能匿名访问（60次/小时限制）。

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var ghComments []githubComment
	if err := json.NewDecoder(resp.Body).Decode(&ghComments); err != nil {
		return nil, err
	}

	var comments []domain.Comment
	for _, c := range ghComments {
		comments = append(comments, domain.Comment{
			ID:        fmt.Sprintf("%d", c.ID),
			Avatar:    c.User.AvatarUrl,
			Nickname:  c.User.Login,
			URL:       c.User.HtmlUrl,
			Content:   c.Body, // Markdown
			CreatedAt: c.CreatedAt.Format(time.RFC3339),
			ArticleID: c.IssueUrl, // 暂时用 Issue URL 代替
			// ParentID: "",
		})
	}

	return comments, nil
}

// GetAdminComments implementation
func (p *GitHubProvider) GetAdminComments(ctx context.Context, page, pageSize int) (*domain.PaginatedComments, error) {
	// TODO: Implement pagination for GitHub
	comments, err := p.GetRecentComments(ctx, pageSize)
	if err != nil {
		return nil, err
	}
	// Warning: Total count is fake here
	return &domain.PaginatedComments{
		Comments:   comments,
		Total:      len(comments), // Incorrect, just to pass interface
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 1,
	}, nil
}

func (p *GitHubProvider) PostComment(ctx context.Context, comment domain.Comment) error {
	// 需要 Token 才能发送
	return fmt.Errorf("PostComment requires authentication token")
}

func (p *GitHubProvider) DeleteComment(ctx context.Context, commentID string) error {
	return fmt.Errorf("DeleteComment requires authentication token")
}
