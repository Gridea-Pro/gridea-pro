package service

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/utils"
	"strings"
	"sync"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type TagService struct {
	repo     domain.TagRepository
	postRepo domain.PostRepository
	mu       sync.RWMutex
}

func NewTagService(repo domain.TagRepository, postRepo domain.PostRepository) *TagService {
	return &TagService{repo: repo, postRepo: postRepo}
}

func (s *TagService) LoadTags(ctx context.Context) ([]domain.Tag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.repo.List(ctx)
}

func (s *TagService) SaveTag(ctx context.Context, tag domain.Tag, originalName string) error {
	s.mu.Lock()

	tags, err := s.repo.List(ctx)
	if err != nil {
		s.mu.Unlock()
		return err
	}

	var existing *domain.Tag
	if originalName != "" {
		for _, t := range tags {
			if t.Name == originalName {
				existing = &t
				break
			}
		}
	} else if tag.ID != "" {
		for _, t := range tags {
			if t.ID == tag.ID {
				existing = &t
				break
			}
		}
	}

	if existing != nil {
		isRename := originalName != "" && existing.Name != tag.Name
		existing.Name = tag.Name
		existing.Slug = tag.Slug
		existing.Color = tag.Color
		if err := s.repo.Update(ctx, existing); err != nil {
			s.mu.Unlock()
			return err
		}
		s.mu.Unlock()
		if isRename {
			return s.cascadeTagRename(ctx, originalName, tag.Name)
		}
		return nil
	}

	// Create new
	if tag.ID == "" {
		const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		id, err := gonanoid.Generate(alphabet, 6)
		if err != nil {
			s.mu.Unlock()
			return err
		}
		tag.ID = id
	}
	tag.Used = true

	err = s.repo.Create(ctx, &tag)
	s.mu.Unlock()
	return err
}

func (s *TagService) DeleteTag(ctx context.Context, name string) error {
	s.mu.Lock()

	tags, err := s.repo.List(ctx)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	for _, t := range tags {
		if t.Name == name {
			if err := s.repo.Delete(ctx, t.ID); err != nil {
				s.mu.Unlock()
				return err
			}
			s.mu.Unlock()
			return s.cascadeTagDelete(ctx, t.ID, name)
		}
	}
	s.mu.Unlock()
	return nil
}

func (s *TagService) SaveTags(ctx context.Context, tags []domain.Tag) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repo.SaveAll(ctx, tags)
}

// GetOrCreateTag gets an existing tag by name or creates a new one with standardized slug and ID
func (s *TagService) GetOrCreateTag(ctx context.Context, name string) (domain.Tag, error) {
	// Critical Section: Ensure check and create are atomic
	s.mu.Lock()
	defer s.mu.Unlock()

	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Tag{}, fmt.Errorf("tag name cannot be empty")
	}

	tags, err := s.repo.List(ctx)
	if err != nil {
		return domain.Tag{}, err
	}

	// 1. Check if exists (Case insensitive for Name)
	for _, t := range tags {
		if strings.EqualFold(t.Name, name) {
			return t, nil
		}
	}

	// 2. Create New Tag
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	id, err := gonanoid.Generate(alphabet, 6)
	if err != nil {
		return domain.Tag{}, err
	}

	// Generate Slug
	slugStr := s.generateSlug(name, tags)

	// Random Color
	hash := 0
	for _, c := range name {
		hash += int(c)
	}
	color := TagColors[hash%len(TagColors)]

	newTag := domain.Tag{
		ID:    id,
		Name:  name,
		Slug:  slugStr,
		Used:  true, // Assuming creation means usage
		Color: color,
	}

	// 3. Save
	if err := s.repo.Create(ctx, &newTag); err != nil {
		return domain.Tag{}, err
	}

	return newTag, nil
}

func (s *TagService) generateSlug(name string, existingTags []domain.Tag) string {
	finalSlug := utils.SlugifyName(name)
	if finalSlug == "" {
		const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		finalSlug, _ = gonanoid.Generate(alphabet, 6)
	}

	uniqueSlug := finalSlug
	counter := 1
	for {
		exists := false
		for _, t := range existingTags {
			if t.Slug == uniqueSlug {
				exists = true
				break
			}
		}
		if !exists {
			break
		}
		uniqueSlug = fmt.Sprintf("%s-%d", finalSlug, counter)
		counter++
	}

	return uniqueSlug
}

func (s *TagService) cascadeTagRename(ctx context.Context, oldName, newName string) error {
	posts, err := s.postRepo.GetAll(ctx)
	if err != nil {
		return err
	}
	for i := range posts {
		changed := false
		for j, t := range posts[i].Tags {
			if t == oldName {
				posts[i].Tags[j] = newName
				changed = true
			}
		}
		if changed {
			if err := s.postRepo.Update(ctx, &posts[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *TagService) cascadeTagDelete(ctx context.Context, tagID, tagName string) error {
	posts, err := s.postRepo.GetAll(ctx)
	if err != nil {
		return err
	}
	for i := range posts {
		changed := false
		var newTags []string
		for _, t := range posts[i].Tags {
			if t != tagName {
				newTags = append(newTags, t)
			} else {
				changed = true
			}
		}
		var newTagIDs []string
		for _, id := range posts[i].TagIDs {
			if id != tagID {
				newTagIDs = append(newTagIDs, id)
			} else {
				changed = true
			}
		}
		if changed {
			posts[i].Tags = newTags
			if posts[i].Tags == nil {
				posts[i].Tags = []string{}
			}
			posts[i].TagIDs = newTagIDs
			if posts[i].TagIDs == nil {
				posts[i].TagIDs = []string{}
			}
			if err := s.postRepo.Update(ctx, &posts[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

var TagColors = []string{
	"#4B5CC4", "#705DC4", "#915DC4", "#AF5DC4", "#C45DB6", "#C45D99", "#C45D7C", "#C45D5D", "#C47C5D", "#C4995D",
	"#B6C45D", "#99C45D", "#7CC45D", "#5DC45D", "#5DC47C", "#5DC499", "#5DC4B6", "#5DAFC4", "#5D91C4", "#5D70C4",
}
