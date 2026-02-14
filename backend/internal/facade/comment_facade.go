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
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.GetSettings(ctx)
}

// SaveSettings 保存评论设置
func (f *CommentFacade) SaveSettings(settings domain.CommentSettings) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SaveSettings(ctx, settings)
}

// FetchComments 获取评论列表
func (f *CommentFacade) FetchComments(page, pageSize int) (*domain.PaginatedComments, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	return f.internal.FetchComments(ctx, page, pageSize)
}

// ReplyComment 回复评论
func (f *CommentFacade) ReplyComment(parentID string, content string, articleID string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.ReplyComment(ctx, parentID, content, articleID)
}

// DeleteComment 删除评论
func (f *CommentFacade) DeleteComment(commentID string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.DeleteComment(ctx, commentID)
}
