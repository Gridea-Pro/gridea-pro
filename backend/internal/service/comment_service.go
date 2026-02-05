package service

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/comment"
	"gridea-pro/backend/internal/domain"
)

// CommentService 评论服务
type CommentService struct {
	repo      domain.CommentRepository
	postRepo  domain.PostRepository
	themeRepo domain.ThemeRepository
	appDir    string
}

// NewCommentService 创建评论服务
func NewCommentService(appDir string, repo domain.CommentRepository, postRepo domain.PostRepository, themeRepo domain.ThemeRepository) *CommentService {
	return &CommentService{
		appDir:    appDir,
		repo:      repo,
		postRepo:  postRepo,
		themeRepo: themeRepo,
	}
}

// GetSettings 获取评论设置
func (s *CommentService) GetSettings(ctx context.Context) (domain.CommentSettings, error) {
	return s.repo.GetSettings(ctx)
}

// SaveSettings 保存评论设置
func (s *CommentService) SaveSettings(ctx context.Context, settings domain.CommentSettings) error {
	return s.repo.SaveSettings(ctx, settings)
}

// FetchComments 获取管理端评论列表
func (s *CommentService) FetchComments(ctx context.Context, page, pageSize int) (*domain.PaginatedComments, error) {
	settings, err := s.repo.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	emptyResult := &domain.PaginatedComments{
		Comments: []domain.Comment{},
		Total:    0,
		Page:     page,
		PageSize: pageSize,
	}

	if !settings.Enable {
		return emptyResult, nil
	}

	provider, err := comment.NewProvider(settings)
	if err != nil {
		return emptyResult, fmt.Errorf("provider init failed: %w", err)
	}

	result, err := provider.GetAdminComments(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 填充文章标题
	posts, _ := s.postRepo.GetAll(ctx)
	postMap := make(map[string]string)
	if len(posts) > 0 {
		for _, p := range posts {
			// 匹配常见路径格式
			key1 := fmt.Sprintf("/post/%s/", p.FileName)
			key2 := fmt.Sprintf("/post/%s", p.FileName)
			postMap[key1] = p.Data.Title
			postMap[key2] = p.Data.Title
			// 兼容可能得根路径配置
			key3 := fmt.Sprintf("/%s/", p.FileName)
			key4 := fmt.Sprintf("/%s", p.FileName)
			postMap[key3] = p.Data.Title
			postMap[key4] = p.Data.Title
		}
	}

	// 获取站点作者信息，用于判定 Admin
	info := s.getSiteOwnerInfo(ctx)
	adminEmail := info.Email

	for i := range result.Comments {
		// 1. ArticleID (URL Path) -> Title
		if len(postMap) > 0 {
			if title, ok := postMap[result.Comments[i].ArticleID]; ok {
				result.Comments[i].ArticleTitle = title
			}
		}

		// 2. Avatar override for Admin
		if adminEmail != "" && result.Comments[i].Email == adminEmail {
			result.Comments[i].Avatar = info.Avatar
		}
	}

	return result, nil
}

// ReplyComment 回复评论
func (s *CommentService) ReplyComment(ctx context.Context, parentID string, content string, articleID string) error {
	settings, err := s.repo.GetSettings(ctx)
	if err != nil {
		return err
	}

	provider, err := comment.NewProvider(settings)
	if err != nil {
		return err
	}

	// 获取站点信息
	siteInfo := s.getSiteOwnerInfo(ctx)

	// 构造评论对象
	newComment := domain.Comment{
		Content:   content,
		ParentID:  parentID,
		ArticleID: articleID,
		Nickname:  siteInfo.Nickname,
		Email:     siteInfo.Email,
		URL:       siteInfo.URL,
		Avatar:    siteInfo.Avatar,
	}

	return provider.PostComment(ctx, newComment)
}

func (s *CommentService) getSiteOwnerInfo(ctx context.Context) domain.Comment {
	// 默认值
	info := domain.Comment{
		Nickname: "Admin",
		URL:      "",
		Avatar:   "",
	}

	// 从主题配置中获取
	if config, err := s.themeRepo.GetConfig(ctx); err == nil {
		if config.SiteAuthor != "" {
			info.Nickname = config.SiteAuthor
		}
		if config.SiteEmail != "" {
			info.Email = config.SiteEmail
		}
		if config.Domain != "" {
			info.URL = config.Domain
		}
	}

	// 构造默认头像地址 (相对于域名的 /images/avatar.png)
	// 如果 info.URL 是完整的 url (http/https)，则拼接
	// 暂时简单处理：假定前端或 Provider 会处理相对路径，或者我们在这里拼完整
	if info.URL != "" {
		// 简单的 URL 拼接
		domain := info.URL
		// remove trailing slash
		if len(domain) > 0 && domain[len(domain)-1] == '/' {
			domain = domain[:len(domain)-1]
		}
		info.Avatar = fmt.Sprintf("%s/images/avatar.png", domain)
	}

	return info
}

// DeleteComment 删除评论
func (s *CommentService) DeleteComment(ctx context.Context, commentID string) error {
	settings, err := s.repo.GetSettings(ctx)
	if err != nil {
		return err
	}

	provider, err := comment.NewProvider(settings)
	if err != nil {
		return err
	}

	return provider.DeleteComment(ctx, commentID)
}
