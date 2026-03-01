package repository

import (
	"context"
	"gridea-pro/backend/internal/domain"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type linkRepository struct {
	*BaseJSONRepository[domain.Link]
}

func NewLinkRepository(appDir string) domain.LinkRepository {
	base := NewBaseJSONRepository[domain.Link](appDir, "links.json", "links")
	return &linkRepository{base}
}

func ensureLinkID(link *domain.Link) {
	if link.ID == "" {
		const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		id, _ := gonanoid.Generate(alphabet, 6)
		link.ID = id
	}
}

func (r *linkRepository) Create(ctx context.Context, link *domain.Link) error {
	ensureLinkID(link)
	return r.Add(ctx, *link)
}

func (r *linkRepository) Update(ctx context.Context, id string, link *domain.Link) error {
	return r.BaseJSONRepository.Update(ctx, id, *link)
}

func (r *linkRepository) GetByID(ctx context.Context, id string) (*domain.Link, error) {
	link, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *linkRepository) SaveAll(ctx context.Context, links []domain.Link) error {
	for i := range links {
		ensureLinkID(&links[i])
	}
	return r.BaseJSONRepository.SaveAll(ctx, links)
}
