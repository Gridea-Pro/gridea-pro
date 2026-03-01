package service

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/comment"
	"gridea-pro/backend/internal/domain"
	"strings"
	"sync"
)

// CommentService 评论服务
type CommentService struct {
	repo      domain.CommentRepository
	postRepo  domain.PostRepository
	themeRepo domain.ThemeRepository
	appDir    string
	mu        sync.RWMutex
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
func (s *CommentService) GetSettings(ctx context.Context) (*domain.CommentSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.repo.GetSettings(ctx)
}

// SaveSettings 保存评论设置
func (s *CommentService) SaveSettings(ctx context.Context, settings domain.CommentSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repo.SaveSettings(ctx, &settings)
}

// FetchComments 获取管理端评论列表
func (s *CommentService) FetchComments(ctx context.Context, page, pageSize int) (*domain.PaginatedComments, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

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

	// 未启用或未完整配置时，跳过网络请求
	if !comment.IsConfigured(*settings) {
		return emptyResult, nil
	}

	provider, err := comment.NewProvider(*settings)
	if err != nil {
		return emptyResult, fmt.Errorf("provider init failed: %w", err)
	}

	comments, count, err := provider.GetAdminComments(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	result := &domain.PaginatedComments{
		Comments: comments,
		Total:    count,
		Page:     page,
		PageSize: pageSize,
	}
	if pageSize > 0 {
		result.TotalPages = int((count + int64(pageSize) - 1) / int64(pageSize))
	}

	// 填充文章标题 - 优化 O(1) 查找
	posts, _, _ := s.postRepo.List(ctx, 1, 10000) // Revert to GetAll as PostRepository uses GetAll
	postMap := make(map[string]*domain.Post)      // Key: URL Path / ID, Value: Post Reference
	if len(posts) > 0 {
		for i := range posts {
			p := &posts[i]
			// 匹配新出的 ID 关联方式
			if p.ID != "" {
				postMap[p.ID] = p
			}
			// 匹配旧版常见路径格式 (向下兼容旧服务云端挂载的 path)
			key1 := fmt.Sprintf("/post/%s/", p.FileName)
			key2 := fmt.Sprintf("/post/%s", p.FileName)
			postMap[key1] = p
			postMap[key2] = p
			// 兼容可能得根路径配置
			key3 := fmt.Sprintf("/%s/", p.FileName)
			key4 := fmt.Sprintf("/%s", p.FileName)
			postMap[key3] = p
			postMap[key4] = p
		}
	}

	// 获取站点作者信息，用于判定 Admin
	info := s.getSiteOwnerInfo(ctx)
	adminEmail := info.Email

	for i := range result.Comments {
		// 1. ArticleID (URL Path) -> Title
		if len(postMap) > 0 {
			if postData, ok := postMap[result.Comments[i].ArticleID]; ok {
				result.Comments[i].ArticleTitle = postData.Title
				// result.Comments[i].ArticleID might be a URL or ID from the provider.
				// ensure ArticleURL is populated correctly for frontend navigation regardless.
				// For old provider comments, ArticleID IS the URL. We must preserve it as ArticleURL.
				if strings.HasPrefix(result.Comments[i].ArticleID, "/") {
					result.Comments[i].ArticleURL = result.Comments[i].ArticleID
				}
				result.Comments[i].ArticleID = postData.ID
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
	s.mu.Lock()
	defer s.mu.Unlock()

	settings, err := s.repo.GetSettings(ctx)
	if err != nil {
		return err
	}

	// 未启用或未完整配置时，拒绝操作
	if !comment.IsConfigured(*settings) {
		return fmt.Errorf("评论功能未启用或配置不完整，无法回复")
	}

	provider, err := comment.NewProvider(*settings)
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

	return provider.PostComment(ctx, &newComment)
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
	if info.URL != "" {
		// 简单的 URL 拼接
		domainUrl := info.URL
		// remove trailing slash
		if len(domainUrl) > 0 && domainUrl[len(domainUrl)-1] == '/' {
			domainUrl = domainUrl[:len(domainUrl)-1]
		}
		info.Avatar = fmt.Sprintf("%s/images/avatar.png", domainUrl)
	}

	return info
}

// DeleteComment 删除评论
func (s *CommentService) DeleteComment(ctx context.Context, commentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	settings, err := s.repo.GetSettings(ctx)
	if err != nil {
		return err
	}

	// 未启用或未完整配置时，拒绝操作
	if !comment.IsConfigured(*settings) {
		return fmt.Errorf("评论功能未启用或配置不完整，无法删除")
	}

	provider, err := comment.NewProvider(*settings)
	if err != nil {
		return err
	}

	return provider.DeleteComment(ctx, commentID)
}
