package facade

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
	"sync/atomic"
)

// CommentFacade 评论外观 - 暴露给前端的接口
type CommentFacade struct {
	// internal 用原子读写保存：切站（UpdateAppDir）会热替换它，
	// 而 Wails 可能并发调用本 facade 的方法，原子读写避免"读到替换一半的状态"的 data race。
	internal atomic.Pointer[service.CommentService]
}

// NewCommentFacade 创建评论外观
func NewCommentFacade(s *service.CommentService) *CommentFacade {
	f := &CommentFacade{}
	f.internal.Store(s)
	return f
}

func (f *CommentFacade) svc() *service.CommentService { return f.internal.Load() }

// GetSettings 获取评论设置
func (f *CommentFacade) GetSettings() (domain.CommentSettings, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	settings, err := f.svc().GetSettings(ctx)
	if err != nil {
		return domain.CommentSettings{}, err
	}
	return *settings, nil
}

// SaveSettings 保存评论设置
func (f *CommentFacade) SaveSettings(settings domain.CommentSettings) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().SaveSettings(ctx, settings)
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
	return f.svc().FetchComments(ctx, page, pageSize)
}

// ReplyComment 回复评论
func (f *CommentFacade) ReplyComment(parentID string, content string, articleID string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().ReplyComment(ctx, parentID, content, articleID)
}

// DeleteComment 删除评论
func (f *CommentFacade) DeleteComment(commentID string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().DeleteComment(ctx, commentID)
}
