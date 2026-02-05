package domain

import "context"

// CommentPlatform 评论平台类型
type CommentPlatform string

const (
	CommentPlatformValine CommentPlatform = "Valine"
	CommentPlatformWaline CommentPlatform = "Waline"
	CommentPlatformTwikoo CommentPlatform = "Twikoo"
	CommentPlatformGitalk CommentPlatform = "Gitalk"
	CommentPlatformGiscus CommentPlatform = "Giscus"
	CommentPlatformDisqus CommentPlatform = "Disqus"
	CommentPlatformCusdis CommentPlatform = "Cusdis"
)

// CommentSettings 评论设置
type CommentSettings struct {
	Enable   bool            `json:"enable"`
	Platform CommentPlatform `json:"platform"`

	// PlatformConfigs 存储各平台的配置 map[PlatformName]ConfigMap
	PlatformConfigs map[CommentPlatform]map[string]any `json:"platformConfigs"`
}

// Comment 统一评论模型
type Comment struct {
	ID           string `json:"id"`
	Avatar       string `json:"avatar"`
	Nickname     string `json:"nickname"`
	Email        string `json:"email,omitempty"`
	URL          string `json:"url,omitempty"`
	Content      string `json:"content"`
	CreatedAt    string `json:"createdAt"`
	ArticleID    string `json:"articleId"`
	ArticleTitle string `json:"articleTitle"`
	ArticleURL   string `json:"articleUrl,omitempty"`
	ParentID     string `json:"parentId,omitempty"`
	ParentNick   string `json:"parentNick,omitempty"`
	IsNew        bool   `json:"isNew,omitempty"`
}

// PaginatedComments 分页评论数据
type PaginatedComments struct {
	Comments   []Comment `json:"comments"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"pageSize"`
	TotalPages int       `json:"totalPages"`
}

// CommentRepository 评论存储接口
type CommentRepository interface {
	GetSettings(ctx context.Context) (CommentSettings, error)
	SaveSettings(ctx context.Context, settings CommentSettings) error
}

// CommentProvider 评论平台提供者接口
type CommentProvider interface {
	// GetComments 获取指定文章的评论
	GetComments(ctx context.Context, articleID string) ([]Comment, error)
	// GetRecentComments 获取最近评论 (Legacy or Widget use)
	GetRecentComments(ctx context.Context, limit int) ([]Comment, error)
	// GetAdminComments 获取管理端评论列表（支持分页）
	// page: 页码，从 1 开始
	// pageSize: 每页数量
	GetAdminComments(ctx context.Context, page, pageSize int) (*PaginatedComments, error)
	// PostComment 发送评论/回复
	PostComment(ctx context.Context, comment Comment) error
	// DeleteComment 删除评论
	DeleteComment(ctx context.Context, commentID string) error
}
