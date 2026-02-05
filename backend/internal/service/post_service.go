package service

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"strings"
)

type PostService struct {
	repo       domain.PostRepository
	tagRepo    domain.TagRepository
	tagService *TagService
	mediaRepo  domain.MediaRepository
}

func NewPostService(repo domain.PostRepository, tagRepo domain.TagRepository, tagService *TagService, mediaRepo domain.MediaRepository) *PostService {
	return &PostService{
		repo:       repo,
		tagRepo:    tagRepo,
		tagService: tagService,
		mediaRepo:  mediaRepo,
	}
}

func (s *PostService) LoadPosts(ctx context.Context) ([]domain.Post, error) {
	posts, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Load all tags to build lookup maps
	allTags, _ := s.tagRepo.GetAll(ctx)
	tagMapByID := make(map[string]domain.Tag)
	tagMapByName := make(map[string]domain.Tag)

	for _, t := range allTags {
		if t.ID != "" {
			tagMapByID[t.ID] = t
		}
		// Use lower case for name matching
		tagMapByName[strings.ToLower(t.Name)] = t
	}

	for i := range posts {
		p := &posts[i]
		needsSave := false

		// Scenario A: Legacy Post (No IDs) or Partial IDs
		// If TagIDs count doesn't match Tags count, we might have missing IDs
		if len(p.Data.TagIDs) != len(p.Data.Tags) {
			var newIDs []string
			var newNames []string

			// Re-evaluate all tags
			for _, tagName := range p.Data.Tags {
				// Try to find by name
				if t, ok := tagMapByName[strings.ToLower(tagName)]; ok {
					newIDs = append(newIDs, t.ID)
					newNames = append(newNames, t.Name) // Use canonical name
				} else {
					// New Tag - Create it
					newTag, err := s.tagService.GetOrCreateTag(ctx, tagName)
					if err == nil {
						newIDs = append(newIDs, newTag.ID)
						newNames = append(newNames, newTag.Name)
						// Update local maps
						tagMapByID[newTag.ID] = newTag
						tagMapByName[strings.ToLower(newTag.Name)] = newTag
					} else {
						// Error creating? Keep original
						newNames = append(newNames, tagName)
					}
				}
			}
			p.Data.TagIDs = newIDs
			p.Data.Tags = newNames
			needsSave = true
		} else {
			// Scenario B: Modern Post (Has IDs) - Check for Renames
			// Trust IDs, update Names
			var syncedNames []string

			for _, id := range p.Data.TagIDs {
				if t, ok := tagMapByID[id]; ok {
					syncedNames = append(syncedNames, t.Name)
				} else {
					// ID not found (Deleted?) - Keep it?
					// Or try to recover from Name at corresponding index?
					// For safety, let's keep the ID but maybe we can't rely on name being correct if index mismatch
					// Simpler: Just resolve what we can.
					// But wait, if we have ID "t1" and name "Old", and t1 is "New", we want "New".
					// If t1 is missing, we might have a dead tag.
					syncedNames = append(syncedNames, "Unknown") // Placeholder? Or try to keep old name?
					// If we can't find the tag, we can't really do much.
					// Let's assume we keep the old name if we can match it by index?
					// But here we are iterating IDs.
					// Let's just append the ID as name to indicate broken link? No that's bad UI.
					// Fallback: If ID not found, check if we have a name at this index in p.Data.Tags?
					// But indices might not align if we modified array.
					// Let's skip unknown IDs or keep them?
					// User said: "If found, update name".
				}
			}

			// Check if names changed
			if len(syncedNames) == len(p.Data.Tags) {
				for j, name := range syncedNames {
					if name != "Unknown" && name != p.Data.Tags[j] {
						p.Data.Tags[j] = name
						needsSave = true
					}
				}
			}
		}

		if needsSave {
			// Construct Input to Save
			input := &domain.PostInput{
				Title:            p.Data.Title,
				Date:             p.Data.Date,
				Tags:             p.Data.Tags,
				TagIDs:           p.Data.TagIDs,
				Categories:       p.Data.Categories,
				Published:        p.Data.Published,
				HideInList:       p.Data.HideInList,
				IsTop:            p.Data.IsTop,
				Content:          p.Content,
				FileName:         p.FileName,
				FeatureImagePath: p.Data.Feature,
			}
			_ = s.repo.Save(ctx, input)
		}
	}

	return posts, nil
}

func (s *PostService) LoadTags(ctx context.Context) ([]domain.Tag, error) {
	posts, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// 1. Get all existing tags to preserve order & metadata
	existingTags, _ := s.tagRepo.GetAll(ctx)

	// Map for quick lookup and tracking usage
	// Key: Tag Name, Value: Index in existingTags slice
	tagIndexMap := make(map[string]int)
	for i, t := range existingTags {
		existingTags[i].Used = false // Reset used status first
		tagIndexMap[t.Name] = i
	}

	// 2. Identify used tags from posts and find new tags
	var newTags []domain.Tag
	// Use a map to deduplicate new tags specifically
	newTagsMap := make(map[string]bool)

	for _, p := range posts {
		for _, tagName := range p.Data.Tags {
			if idx, exists := tagIndexMap[tagName]; exists {
				// Mark existing tag as used
				existingTags[idx].Used = true
			} else {
				// This is a new tag found in posts but not in tagRepo
				if !newTagsMap[tagName] {
					// Use TagService to get or create the tag (generating proper ID and Slug)
					tag, err := s.tagService.GetOrCreateTag(ctx, tagName)
					if err == nil {
						newTags = append(newTags, tag)
						newTagsMap[tagName] = true
					}
				}
			}
		}
	}

	// 3. Combine: Existing Tags (Ordered) + New Tags (Appended)
	finalTags := append(existingTags, newTags...)

	// 4. Save to ensure tags.json is always consistent with posts
	// This fixes the issue where tags found in posts are not present in tags.json
	// and ensures 'Used' status is correctly persisted.
	_ = s.tagRepo.SaveAll(ctx, finalTags)

	return finalTags, nil

}

func (s *PostService) SavePost(ctx context.Context, input *domain.PostInput) error {
	// 1. Resolve TagIDs from Tags (Names)
	// This ensures that whenever we save, we have the latest IDs for the names provided
	var ids []string
	// Use a map to deduplicate IDs if necessary, though GetOrCreate handles uniqueness by name
	for _, tagName := range input.Tags {
		tag, err := s.tagService.GetOrCreateTag(ctx, tagName)
		if err == nil {
			ids = append(ids, tag.ID)
		}
	}
	input.TagIDs = ids

	return s.repo.Save(ctx, input)
}

func (s *PostService) DeletePost(ctx context.Context, fileName string) error {
	return s.repo.Delete(ctx, fileName)
}

func (s *PostService) UploadImages(ctx context.Context, files []domain.UploadedFile) ([]string, error) {
	return s.mediaRepo.SaveImages(ctx, files)
}
