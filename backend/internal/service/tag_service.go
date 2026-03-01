package service

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"strings"
	"sync"
	"unicode"

	"github.com/gosimple/slug"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/mozillazg/go-pinyin"
)

type TagService struct {
	repo domain.TagRepository
	mu   sync.RWMutex
}

func NewTagService(repo domain.TagRepository) *TagService {
	return &TagService{repo: repo}
}

func (s *TagService) LoadTags(ctx context.Context) ([]domain.Tag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.repo.List(ctx)
}

func (s *TagService) SaveTag(ctx context.Context, tag domain.Tag, originalName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tags, err := s.repo.List(ctx)
	if err != nil {
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
		existing.Name = tag.Name
		existing.Slug = tag.Slug
		existing.Color = tag.Color
		return s.repo.Update(ctx, existing)
	}

	// Create new
	if tag.ID == "" {
		const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		id, err := gonanoid.Generate(alphabet, 6)
		if err != nil {
			return err
		}
		tag.ID = id
	}
	tag.Used = true // Assuming creation means usage in this context

	return s.repo.Create(ctx, &tag)
}

func (s *TagService) DeleteTag(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// We need ID to delete efficiently, but legacy API passed name.
	// We must find ID first.
	tags, err := s.repo.List(ctx)
	if err != nil {
		return err
	}
	for _, t := range tags {
		if t.Name == name {
			return s.repo.Delete(ctx, t.ID)
		}
	}
	return nil
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
	// 1. Convert to Pinyin if it contains Chinese
	pinyinArgs := pinyin.NewArgs()
	pinyinArgs.Fallback = func(r rune, a pinyin.Args) []string {
		return []string{string(r)}
	}

	// Check if string contains chinese
	// Simple check: iterate runes
	hasChinese := false
	for _, r := range name {
		if unicode.Is(unicode.Han, r) {
			hasChinese = true
			break
		}
	}

	var preSlug string
	if hasChinese {
		// Pinyin conversion
		// "测试" -> [[ce], [shi]]
		pyRows := pinyin.Pinyin(name, pinyinArgs)
		var parts []string
		for _, row := range pyRows {
			if len(row) > 0 {
				parts = append(parts, row[0])
			}
		}
		preSlug = strings.Join(parts, "-")
	} else {
		preSlug = name
	}

	// 2. Slugify (handling special chars, lower case)
	finalSlug := slug.Make(preSlug)
	if finalSlug == "" {
		// Fallback for purely special chars or empty slug result
		const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		finalSlug, _ = gonanoid.Generate(alphabet, 6)
	}

	// 3. Handle Duplicates
	// Check against existing slugs
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

var TagColors = []string{
	"#4B5CC4", "#705DC4", "#915DC4", "#AF5DC4", "#C45DB6", "#C45D99", "#C45D7C", "#C45D5D", "#C47C5D", "#C4995D",
	"#B6C45D", "#99C45D", "#7CC45D", "#5DC45D", "#5DC47C", "#5DC499", "#5DC4B6", "#5DAFC4", "#5D91C4", "#5D70C4",
}
