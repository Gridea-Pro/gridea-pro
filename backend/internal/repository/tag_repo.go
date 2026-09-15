package repository

import (
	"context"
	"sync"

	"gridea-pro/backend/internal/domain"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type tagRepository struct {
	*BaseJSONRepository[domain.Tag]
	// writeMu 串行化 Create/Update 的"读列表→查唯一性→写入"整个流程，
	// 消除 List 与 Add/Update 分两次独立加锁之间的 TOCTOU 窗口（并发新建同名标签会产生重复）。
	writeMu sync.Mutex
}

func NewTagRepository(appDir string) domain.TagRepository {
	base := NewBaseJSONRepository[domain.Tag](appDir, "tags.json", "tags")
	return &tagRepository{BaseJSONRepository: base}
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
	r.writeMu.Lock()
	defer r.writeMu.Unlock()
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
	r.writeMu.Lock()
	defer r.writeMu.Unlock()
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
