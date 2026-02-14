package service

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"regexp"
	"sort"
	"sync"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

var (
	// compile once for performance
	tagRegexp = regexp.MustCompile(`#([\p{L}\p{N}_]+)`)
)

type MemoService struct {
	repo domain.MemoRepository
	mu   sync.RWMutex
}

func NewMemoService(repo domain.MemoRepository) *MemoService {
	return &MemoService{repo: repo}
}

func (s *MemoService) LoadMemos(ctx context.Context) ([]domain.Memo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.repo.GetAll(ctx)
}

func (s *MemoService) SaveMemos(ctx context.Context, memos []domain.Memo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repo.SaveAll(ctx, memos)
}

func (s *MemoService) CreateMemo(ctx context.Context, content string) (*domain.Memo, error) {
	if content == "" {
		return nil, fmt.Errorf("content is empty")
	}

	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	id, err := gonanoid.Generate(alphabet, 6)
	if err != nil {
		return nil, err
	}

	now := time.Now().Format(domain.TimeLayout)
	newMemo := domain.Memo{
		ID:        id,
		Content:   content,
		Tags:      extractTags(content),
		Images:    []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	memos, err := s.repo.GetAll(ctx)
	if err != nil {
		memos = []domain.Memo{}
	}

	// Prepend
	memos = append([]domain.Memo{newMemo}, memos...)

	if err := s.repo.SaveAll(ctx, memos); err != nil {
		return nil, err
	}

	return &newMemo, nil
}

func (s *MemoService) UpdateMemo(ctx context.Context, memo domain.Memo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	memos, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	// Parse content for tags
	memo.Tags = extractTags(memo.Content)
	memo.UpdatedAt = time.Now().Format(domain.TimeLayout)

	found := false
	for i := range memos {
		if memos[i].ID == memo.ID {
			memos[i].Content = memo.Content
			memos[i].Tags = memo.Tags
			memos[i].UpdatedAt = memo.UpdatedAt
			memos[i].Images = memo.Images
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("memo not found")
	}

	return s.repo.SaveAll(ctx, memos)
}

func (s *MemoService) DeleteMemo(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	memos, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	filtered := make([]domain.Memo, 0)
	found := false
	for _, memo := range memos {
		if memo.ID != id {
			filtered = append(filtered, memo)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("memo not found")
	}

	return s.repo.SaveAll(ctx, filtered)
}

func (s *MemoService) RenameTag(ctx context.Context, oldName, newName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	memos, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	count := 0
	updatedMemos := make([]domain.Memo, 0)

	// Pre-compile regex for replacement
	// Using QuoteMeta to safely escape the tag name
	re := regexp.MustCompile(`#` + regexp.QuoteMeta(oldName) + `([^\p{L}\p{N}_]|$)`)

	for i := range memos {
		hasTag := false
		for _, t := range memos[i].Tags {
			if t == oldName {
				hasTag = true
				break
			}
		}

		if hasTag {
			memos[i].Content = re.ReplaceAllString(memos[i].Content, "#"+newName+"$1")
			memos[i].Tags = extractTags(memos[i].Content)
			memos[i].UpdatedAt = time.Now().Format(domain.TimeLayout)
			count++
		}
		updatedMemos = append(updatedMemos, memos[i])
	}

	if count > 0 {
		return s.repo.SaveAll(ctx, updatedMemos)
	}
	return nil
}

func (s *MemoService) DeleteTag(ctx context.Context, tagName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	memos, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	count := 0
	updatedMemos := make([]domain.Memo, 0)

	// Pre-compile regex for deletion
	re := regexp.MustCompile(`#` + regexp.QuoteMeta(tagName) + `([^\p{L}\p{N}_]|$)`)

	for i := range memos {
		hasTag := false
		for _, t := range memos[i].Tags {
			if t == tagName {
				hasTag = true
				break
			}
		}

		if hasTag {
			memos[i].Content = re.ReplaceAllString(memos[i].Content, tagName+"$1")
			memos[i].Tags = extractTags(memos[i].Content)
			memos[i].UpdatedAt = time.Now().Format(domain.TimeLayout)
			count++
		}
		updatedMemos = append(updatedMemos, memos[i])
	}

	if count > 0 {
		return s.repo.SaveAll(ctx, updatedMemos)
	}
	return nil
}

func (s *MemoService) GetMemoStats(ctx context.Context) (*domain.MemoStats, error) {
	// Re-use LoadMemos which has Read Lock
	memos, err := s.LoadMemos(ctx)
	if err != nil {
		return nil, err
	}

	tagCount := make(map[string]int)
	for _, memo := range memos {
		for _, tag := range memo.Tags {
			tagCount[tag]++
		}
	}

	var tagStats []domain.TagStat
	for name, count := range tagCount {
		tagStats = append(tagStats, domain.TagStat{
			Name:  name,
			Count: count,
		})
	}

	sort.Slice(tagStats, func(i, j int) bool {
		return tagStats[i].Count > tagStats[j].Count
	})

	heatmap := make(map[string]int)
	now := time.Now()
	for i := 0; i < 365; i++ {
		date := now.AddDate(0, 0, -i).Format(domain.DateLayout)
		heatmap[date] = 0
	}

	for _, memo := range memos {
		t, err := time.Parse(domain.TimeLayout, memo.CreatedAt)
		if err == nil {
			date := t.Format(domain.DateLayout)
			if _, exists := heatmap[date]; exists {
				heatmap[date]++
			}
		}
	}

	return &domain.MemoStats{
		Total:   len(memos),
		Tags:    tagStats,
		Heatmap: heatmap,
	}, nil
}

// extractTags helper
func extractTags(content string) []string {
	matches := tagRegexp.FindAllStringSubmatch(content, -1)

	tagSet := make(map[string]bool)
	tags := make([]string, 0)
	for _, match := range matches {
		if len(match) > 1 {
			tag := match[1]
			if !tagSet[tag] {
				tagSet[tag] = true
				tags = append(tags, tag)
			}
		}
	}
	return tags
}
