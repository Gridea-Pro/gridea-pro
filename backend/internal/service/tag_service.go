package service

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"strings"
	"unicode"

	"github.com/gosimple/slug"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/mozillazg/go-pinyin"
)

type TagService struct {
	repo domain.TagRepository
}

func NewTagService(repo domain.TagRepository) *TagService {
	return &TagService{repo: repo}
}

func (s *TagService) LoadTags(ctx context.Context) ([]domain.Tag, error) {
	return s.repo.GetAll(ctx)
}

func (s *TagService) SaveTag(ctx context.Context, tag domain.Tag, originalName string) error {
	tags, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	found := false
	for i, t := range tags {
		if t.Name == originalName {
			// Update existing tag, preserve ID if not present in new tag (though it should be)
			if tag.ID == "" {
				tag.ID = t.ID
			}
			tags[i] = tag
			found = true
			break
		}
	}
	if !found {
		if tag.ID == "" {
			const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
			tag.ID, _ = gonanoid.Generate(alphabet, 6)
		}
		tags = append(tags, tag)
	}

	return s.repo.SaveAll(ctx, tags)
}

func (s *TagService) DeleteTag(ctx context.Context, name string) error {
	tags, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	var newTags []domain.Tag
	for _, t := range tags {
		if t.Name != name {
			newTags = append(newTags, t)
		}
	}
	return s.repo.SaveAll(ctx, newTags)
}

// GetOrCreateTag gets an existing tag by name or creates a new one with standardized slug and ID
func (s *TagService) GetOrCreateTag(ctx context.Context, name string) (domain.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Tag{}, fmt.Errorf("tag name cannot be empty")
	}

	tags, err := s.repo.GetAll(ctx)
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
	id, err := gonanoid.Generate(alphabet, 5)
	if err != nil {
		return domain.Tag{}, err
	}

	// Generate Slug
	slugStr := s.generateSlug(name, tags)

	newTag := domain.Tag{
		ID:   id,
		Name: name,
		Slug: slugStr,
		Used: true, // Assuming creation means usage
	}

	// 3. Save
	tags = append(tags, newTag)
	if err := s.repo.SaveAll(ctx, tags); err != nil {
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
		finalSlug, _ = gonanoid.Generate(alphabet, 5)
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
