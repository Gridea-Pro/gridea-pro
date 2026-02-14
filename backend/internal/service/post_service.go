package service

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"sync"
)

type PostService struct {
	repo            domain.PostRepository
	tagRepo         domain.TagRepository
	tagService      *TagService
	categoryService *CategoryService
	mediaRepo       domain.MediaRepository
	mu              sync.RWMutex
}

func NewPostService(repo domain.PostRepository, tagRepo domain.TagRepository, tagService *TagService, categoryService *CategoryService, mediaRepo domain.MediaRepository) *PostService {
	return &PostService{
		repo:            repo,
		tagRepo:         tagRepo,
		tagService:      tagService,
		categoryService: categoryService,
		mediaRepo:       mediaRepo,
	}
}

// LoadPosts Pure read operation. No side effects.
func (s *PostService) LoadPosts(ctx context.Context) ([]domain.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.repo.GetAll(ctx)
}

func (s *PostService) LoadTags(ctx context.Context) ([]domain.Tag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// This was originally in PostService, but ideally should be in TagService or just use Repo.
	// Since the requirement didn't ask to move it, we keep it but ensure it's safe.
	// However, the original code had side effects (saving tags).
	// We will keep it as read-only or strictly simple aggregation if possible,
	// but the original logic synced 'Used' status.
	// If we must be pure read, we shouldn't save.
	// But resetting 'Used' status is a write operation conceptually.
	// Given the instructions for LoadPosts were CQS, lets assume LoadTags also shouldn't arbitrarily mutate state if it's a "Load".
	// But if the "Used" status is calculated on the fly, it's fine.
	// The original code SAVED the modifications.
	// We will RLock here, so we CANNOT save.
	// If 'Used' status is needed, we should calculate it and return it, but NOT save it to disk during a Load.

	posts, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	tags, err := s.tagRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate 'Used' status in memory
	tagUsage := make(map[string]bool)
	for _, p := range posts {
		for _, tagID := range p.Data.TagIDs {
			tagUsage[tagID] = true
		}
		// Also check names for backward compatibility if needed, but PostService Save guarantees IDs now.
	}

	for i := range tags {
		if tagUsage[tags[i].ID] {
			tags[i].Used = true
		} else {
			tags[i].Used = false
		}
	}

	return tags, nil
}

func (s *PostService) SavePost(ctx context.Context, input *domain.PostInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Resolve TagIDs from Tags (Names)
	// This ensures that whenever we save, we have the latest IDs for the names provided
	var ids []string

	// We need to call TagService.GetOrCreateTag.
	// Since TagService likely has its own locks, this is fine, but we must be careful of deadlocks if TagService calls back to PostService.
	// TagService only depends on TagRepository, so it should be fine.
	for _, tagName := range input.Tags {
		tag, err := s.tagService.GetOrCreateTag(ctx, tagName)
		if err == nil {
			ids = append(ids, tag.ID)
		}
	}
	input.TagIDs = ids

	// 2. Ensure Categories Exist
	for _, catName := range input.Categories {
		if _, err := s.categoryService.GetOrCreateCategory(ctx, catName); err != nil {
			return err
		}
	}

	return s.repo.Save(ctx, input)
}

func (s *PostService) DeletePost(ctx context.Context, fileName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repo.Delete(ctx, fileName)
}

func (s *PostService) UploadImages(ctx context.Context, files []domain.UploadedFile) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mediaRepo.SaveImages(ctx, files)
}

func (s *PostService) GetByFileName(ctx context.Context, fileName string) (*domain.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.repo.GetByFileName(ctx, fileName)
}
