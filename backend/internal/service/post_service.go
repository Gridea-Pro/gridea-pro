package service

import (
	"context"
	"fmt"
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

// LoadPosts Pure read operation.
func (s *PostService) LoadPosts(ctx context.Context) ([]domain.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 假设默认分页大小足够大，或者我们实现一个新的 GetAll for internal use if needed
	// 但 Repository interface 只有 List(page, size).
	// 为了兼容 LoadPosts 语义 (返回所有)，我们传递大数
	posts, _, err := s.repo.List(ctx, 1, 10000)
	return posts, err
}

func (s *PostService) LoadTags(ctx context.Context) ([]domain.Tag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	posts, _, err := s.repo.List(ctx, 1, 10000)
	if err != nil {
		return nil, err
	}

	tags, err := s.tagRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate 'Used' status in memory
	tagUsage := make(map[string]bool)
	for _, p := range posts {
		for _, tagID := range p.TagIDs { // Updated field access
			tagUsage[tagID] = true
		}
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

func (s *PostService) SavePost(ctx context.Context, post *domain.Post) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Resolve TagIDs from Tags (Names)
	var ids []string
	for _, tagName := range post.Tags {
		tag, err := s.tagService.GetOrCreateTag(ctx, tagName)
		if err == nil {
			ids = append(ids, tag.ID)
		}
	}
	post.TagIDs = ids

	// 2. Ensure Categories Exist & Resolve CategoryIDs (UUID)
	var catIDs []string
	for _, catName := range post.Categories {
		cat, err := s.categoryService.GetOrCreateCategory(ctx, catName)
		if err != nil {
			return err
		}
		catIDs = append(catIDs, cat.ID) // 存储不可变 UUID
	}
	// 若前端已直接传入 CategoryIDs（UUID），优先使用；否则用上面解析的结果
	if len(post.CategoryIDs) == 0 {
		post.CategoryIDs = catIDs
	}

	// Check if update or create by trying to get by filename
	// Or simply call Update if we know it exists.
	// But `post` arg is generic.
	// If it's a new post, user might not provide "DeleteFileName" (rename source).
	// Let's try GetByFileName to see if it exists.
	// 3. check for rename
	if post.DeleteFileName != "" && post.DeleteFileName != post.FileName {
		if _, err := s.repo.GetByFileName(ctx, post.DeleteFileName); err == nil {
			if err := s.repo.Delete(ctx, post.DeleteFileName); err != nil {
				return fmt.Errorf("failed to delete old file during rename: %w", err)
			}
		}
		return s.repo.Create(ctx, post)
	}

	// 4. Standard Create/Update check
	_, err := s.repo.GetByFileName(ctx, post.FileName)
	if err == nil {
		return s.repo.Update(ctx, post)
	}

	return s.repo.Create(ctx, post)
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
