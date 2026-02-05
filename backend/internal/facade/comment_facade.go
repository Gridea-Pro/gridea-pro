package facade

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

// CommentFacade 评论外观 - 暴露给前端的接口
type CommentFacade struct {
	internal *service.CommentService
}

// NewCommentFacade 创建评论外观
func NewCommentFacade(s *service.CommentService) *CommentFacade {
	return &CommentFacade{internal: s}
}

// GetSettings 获取评论设置
func (f *CommentFacade) GetSettings() (domain.CommentSettings, error) {
	return f.internal.GetSettings(context.TODO())
}

// SaveSettings 保存评论设置
func (f *CommentFacade) SaveSettings(settings domain.CommentSettings) error {
	return f.internal.SaveSettings(context.TODO(), settings)
}

// FetchComments 获取评论列表
func (f *CommentFacade) FetchComments(page, pageSize int) (*domain.PaginatedComments, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	return f.internal.FetchComments(context.TODO(), page, pageSize)
}

// ReplyComment 回复评论
func (f *CommentFacade) ReplyComment(parentID string, content string, articleID string) error {
	return f.internal.ReplyComment(context.TODO(), parentID, content, articleID)
}

// DeleteComment 删除评论
func (f *CommentFacade) DeleteComment(commentID string) error {
	return f.internal.DeleteComment(context.TODO(), commentID)
}
