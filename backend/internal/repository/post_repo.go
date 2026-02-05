package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"gridea-pro/backend/internal/domain"

	"gopkg.in/yaml.v3"
)

type postRepository struct {
	mu     sync.RWMutex
	appDir string
}

func NewPostRepository(appDir string) domain.PostRepository {
	return &postRepository{
		appDir: appDir,
	}
}

func (r *postRepository) GetAll(ctx context.Context) ([]domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	postsDir := filepath.Join(r.appDir, "posts")
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create posts dir: %w", err)
	}

	files, err := os.ReadDir(postsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read posts dir: %w", err)
	}

	var posts []domain.Post
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(postsDir, file.Name()))
		if err != nil {
			continue
		}

		post, err := r.parsePost(string(content), file.Name())
		if err != nil {
			continue
		}
		posts = append(posts, post)
	}

	// Sort by date desc
	sort.Slice(posts, func(i, j int) bool {
		ti, _ := time.Parse("2006-01-02 15:04:05", posts[i].Data.Date)
		tj, _ := time.Parse("2006-01-02 15:04:05", posts[j].Data.Date)
		return ti.After(tj)
	})

	// Cache to JSON (Side effect suitable for Repo layer or Service?
	// In strict Clean Arch, Repo shouldn't have side effects like caching unless invisible.
	// But Gridea implementation seems to rely on this JSON cache.
	// We preserve it here.)
	dbPath := filepath.Join(r.appDir, "config", "posts.json")
	db := map[string]interface{}{"posts": posts}
	_ = SaveJSONFile(dbPath, db)

	return posts, nil
}

func (r *postRepository) GetByFileName(ctx context.Context, fileName string) (*domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	postPath := filepath.Join(r.appDir, "posts", fileName+".md")
	content, err := os.ReadFile(postPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read post file: %w", err)
	}

	post, err := r.parsePost(string(content), fileName+".md")
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) Save(ctx context.Context, input *domain.PostInput) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	postsDir := filepath.Join(r.appDir, "posts")
	postImageDir := filepath.Join(r.appDir, "post-images")
	_ = os.MkdirAll(postsDir, 0755)
	_ = os.MkdirAll(postImageDir, 0755)

	tagsStr := strings.Join(input.Tags, ",")
	tagsStr = escapeYAMLString(tagsStr) // Sanitize? Actually tags are usually safe chars but nice to be safe.
	// The current manual construction assumes simple tags.

	// Convert tags to formatted array string "tag1", "tag2"
	var formattedTags []string
	for _, t := range input.Tags {
		formattedTags = append(formattedTags, fmt.Sprintf("'%s'", escapeYAMLString(t)))
	}
	tagsStr = strings.Join(formattedTags, ", ")

	// Convert tagIDs to formatted array string
	var formattedTagIDs []string
	for _, id := range input.TagIDs {
		formattedTagIDs = append(formattedTagIDs, fmt.Sprintf("'%s'", escapeYAMLString(id)))
	}
	tagIDsStr := strings.Join(formattedTagIDs, ", ")

	categoriesStr := strings.Join(input.Categories, ",")
	feature := input.FeatureImagePath

	// Handle Image Copy
	if input.FeatureImage.Name != "" && input.FeatureImage.Path != "" {
		ext := filepath.Ext(input.FeatureImage.Name)
		newPath := filepath.Join(postImageDir, input.FileName+ext)
		// Check valid path to prevent path traversal?
		if err := CopyFile(input.FeatureImage.Path, newPath); err == nil {
			feature = "/post-images/" + input.FileName + ext
			// Cleanup temp file if necessary
			if input.FeatureImage.Path != newPath && strings.Contains(input.FeatureImage.Path, postImageDir) {
				_ = os.Remove(input.FeatureImage.Path)
			}
		}
	}

	mdContent := fmt.Sprintf(`---
title: '%s'
date: %s
tags: [%s]
tag_ids: [%s]
categories: [%s]
published: %t
hideInList: %t
feature: %s
isTop: %t
---
%s`,
		escapeYAMLString(input.Title),
		input.Date,
		tagsStr,
		tagIDsStr,
		categoriesStr,
		input.Published,
		input.HideInList,
		feature,
		input.IsTop,
		input.Content,
	)

	postPath := filepath.Join(postsDir, input.FileName+".md")
	if err := os.WriteFile(postPath, []byte(mdContent), 0644); err != nil {
		return fmt.Errorf("failed to write post file: %w", err)
	}

	if input.DeleteFileName != "" && input.DeleteFileName != input.FileName {
		oldPath := filepath.Join(postsDir, input.DeleteFileName+".md")
		_ = os.Remove(oldPath)
	}

	return nil
}

func (r *postRepository) Delete(ctx context.Context, fileName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	postsDir := filepath.Join(r.appDir, "posts")
	postPath := filepath.Join(postsDir, fileName+".md")

	// Read content to find assets to delete
	content, err := os.ReadFile(postPath)
	if err == nil {
		post, _ := r.parsePost(string(content), fileName+".md")

		// Delete feature image
		if post.Data.Feature != "" && !strings.HasPrefix(post.Data.Feature, "http") {
			featurePath := filepath.Join(r.appDir, strings.TrimPrefix(post.Data.Feature, "/"))
			_ = os.Remove(featurePath)
		}

		// Delete embedded images
		re := regexp.MustCompile(`!\[.*?\]\((.+?)\)`)
		matches := re.FindAllStringSubmatch(post.Content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				imgPath := match[1]
				if !strings.HasPrefix(imgPath, "http") {
					fullPath := filepath.Join(r.appDir, strings.TrimPrefix(imgPath, "/"))
					_ = os.Remove(fullPath)
				}
			}
		}
	}

	return os.Remove(postPath)
}

// Internal helpers

func (r *postRepository) parsePost(content string, filename string) (domain.Post, error) {
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return domain.Post{}, fmt.Errorf("invalid post format")
	}

	var data domain.PostData
	if err := yaml.Unmarshal([]byte(parts[1]), &data); err != nil {
		return domain.Post{}, err
	}

	postContent := strings.TrimSpace(parts[2])
	abstract := r.extractAbstract(postContent)

	post := domain.Post{
		Data:     data,
		Content:  postContent,
		Abstract: abstract,
		FileName: strings.TrimSuffix(filename, ".md"),
	}

	if post.Data.Date == "" {
		post.Data.Date = time.Now().Format("2006-01-02 15:04:05")
	}

	return post, nil
}

func (r *postRepository) extractAbstract(content string) string {
	re := regexp.MustCompile(`(?i)\n\s*<!--\s*more\s*-->\s*\n`)
	loc := re.FindStringIndex(content)
	if loc != nil {
		return strings.TrimSpace(content[:loc[0]])
	}
	return ""
}

func escapeYAMLString(s string) string {
	s = strings.ReplaceAll(s, "'", "''")
	return s
}
