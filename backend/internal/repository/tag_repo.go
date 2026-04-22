package repository

import (
	"context"
	"gridea-pro/backend/internal/domain"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type tagRepository struct {
	*BaseJSONRepository[domain.Tag]
}

func NewTagRepository(appDir string) domain.TagRepository {
	base := NewBaseJSONRepository[domain.Tag](appDir, "tags.json", "tags")
	return &tagRepository{base}
}

func ensureTagID(tag *domain.Tag) {
	if tag.ID == "" {
		const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		id, _ := gonanoid.Generate(alphabet, 6)
		tag.ID = id
	}
}

func (r *tagRepository) Create(ctx context.Context, tag *domain.Tag) error {
	ensureTagID(tag)
	existing, err := r.List(ctx)
	if err != nil {
		return err
	}
	if err := checkTagUniqueness(existing, *tag, tag.ID); err != nil {
		return err
	}
	return r.Add(ctx, *tag)
}

func (r *tagRepository) Update(ctx context.Context, tag *domain.Tag) error {
	existing, err := r.List(ctx)
	if err != nil {
		return err
	}
	if err := checkTagUniqueness(existing, *tag, tag.ID); err != nil {
		return err
	}
	return r.BaseJSONRepository.Update(ctx, tag.ID, *tag)
}

func (r *tagRepository) SaveAll(ctx context.Context, tags []domain.Tag) error {
	for i := range tags {
		ensureTagID(&tags[i])
	}
	return r.BaseJSONRepository.SaveAll(ctx, tags)
}
