package comment

import (
	"context"
	"gridea-pro/backend/internal/domain"
)

type DisqusProvider struct {
	config map[string]any
}

func NewDisqusProvider(config map[string]any) *DisqusProvider {
	return &DisqusProvider{config: config}
}

func (p *DisqusProvider) GetComments(ctx context.Context, articleID string) ([]domain.Comment, error) {
	return []domain.Comment{}, nil
}

func (p *DisqusProvider) GetRecentComments(ctx context.Context, limit int) ([]domain.Comment, error) {
	return []domain.Comment{}, nil
}

func (p *DisqusProvider) GetAdminComments(ctx context.Context, page, pageSize int) (*domain.PaginatedComments, error) {
	return &domain.PaginatedComments{
		Comments: []domain.Comment{},
	}, nil
}

func (p *DisqusProvider) PostComment(ctx context.Context, comment domain.Comment) error {
	return nil
}

func (p *DisqusProvider) DeleteComment(ctx context.Context, commentID string) error {
	return nil
}
