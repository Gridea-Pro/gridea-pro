package service

import (
	"context"
	"gridea-pro/backend/internal/domain"
)

type LinkService struct {
	repo domain.LinkRepository
}

func NewLinkService(repo domain.LinkRepository) *LinkService {
	return &LinkService{repo: repo}
}

func (s *LinkService) LoadLinks(ctx context.Context) ([]domain.Link, error) {
	return s.repo.GetAll(ctx)
}

func (s *LinkService) SaveLinks(ctx context.Context, links []domain.Link) error {
	return s.repo.SaveAll(ctx, links)
}
