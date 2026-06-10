package repository

import (
	"context"
	"errors"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"gopkg.in/yaml.v3"
)

type postRepository struct {
	mu     sync.RWMutex
	appDir string
	cache  []domain.Post
	loaded bool
}

func NewPostRepository(appDir string) domain.PostRepository {
	return &postRepository{
		appDir: appDir,
		cache:  make([]domain.Post, 0),
		loaded: false,
	}
}

// local struct to handle YAML frontmatter parsing and marshalling
type postYaml struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	CreatedAt   string   `yaml:"createdAt,omitempty"`
	Date        string   `yaml:"date,omitempty"`    // 兼容老版本
	UpdatedAt   string   `yaml:"updated,omitempty"` // 最后修改时间
	Tags        []string `yaml:"tags"`
	TagIDs      []string `yaml:"tag_ids"`
	Categories  []string `yaml:"categories"`
	CategoryIDs []string `yaml:"category_ids,omitempty"` // 分类 Slug
	Published   bool     `yaml:"published"`
	HideInList  bool     `yaml:"hideInList"`
	Feature     string   `yaml:"feature"`
	IsTop       bool     `yaml:"isTop"`
}

func (r *postRepository) Create(ctx context.Context, post *domain.Post) error {
	return r.save(ctx, post, false)
}

func (r *postRepository) Update(ctx context.Context, post *domain.Post) error {
	return r.save(ctx, post, true)
}

func (r *postRepository) scanPosts() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.loaded {
		return nil
	}

	postsDir := filepath.Join(r.appDir, "posts")
	if _, err := os.Stat(postsDir); os.IsNotExist(err) {
		r.cache = []domain.Post{}
		r.loaded = true
		return nil
	}

	files, err := os.ReadDir(postsDir)
	if err != nil {
		return fmt.Errorf("failed to read posts dir: %w", err)
	}

	var allPosts []domain.Post
	var fileErrs []error
	mdCount := 0
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		mdCount++
		content, err := os.ReadFile(filepath.Join(postsDir, file.Name()))
		if err != nil {
			log.Printf("[postRepo] read %s failed: %v", file.Name(), err)
			fileErrs = append(fileErrs, fmt.Errorf("read %s: %w", file.Name(), err))
			continue
		}
		post, err := r.parsePost(string(content), file.Name())
		if err != nil {
			log.Printf("[postRepo] parse %s failed: %v", file.Name(), err)
			fileErrs = append(fileErrs, fmt.Errorf("parse %s: %w", file.Name(), err))
			continue
		}
		allPosts = append(allPosts, post)
	}

	// 关键判断（issue #107）：
	// - 磁盘上有 N 个 .md 文件，但**一篇都没成功加载** → 不是"用户没文章"，
	//   而是底层 I/O 全军覆没（权限、文件系统瞬态错误、批量 YAML 损坏等）。
	//   不能把空结果锁进 cache + loaded=true，否则用户永远看不到文章直到下次 invalidate。
	// - 部分失败（mdCount > len(allPosts)）：保留成功的部分，log 错误供排查。
	if mdCount > 0 && len(allPosts) == 0 {
		return fmt.Errorf("posts dir has %d .md files but none could be loaded: %w",
			mdCount, errors.Join(fileErrs...))
	}
	if len(fileErrs) > 0 {
		log.Printf("[postRepo] partial scan: %d/%d .md files failed to load",
			len(fileErrs), mdCount)
	}

	// Sort: 置顶优先，再按 CreatedAt 降序
	sort.Slice(allPosts, func(i, j int) bool {
		if allPosts[i].IsTop != allPosts[j].IsTop {
			return allPosts[i].IsTop
		}
		return allPosts[i].CreatedAt.After(allPosts[j].CreatedAt)
	})

	r.cache = allPosts
	r.loaded = true
	return nil
}

func (r *postRepository) Reload(ctx context.Context) error {
	r.mu.Lock()
	r.loaded = false
	r.cache = nil
	r.mu.Unlock()
	return r.scanPosts()
}

func (r *postRepository) save(ctx context.Context, post *domain.Post, isUpdate bool) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 入口归一：保证写盘/缓存键用的 FileName 是裸名（消除 .md 双扩展 / 缓存失同步）
	post.FileName = normalizeFileName(post.FileName)
	post.DeleteFileName = normalizeFileName(post.DeleteFileName)

	// Ensure cache is loaded before modifying it (to avoid partial state if saving without listing first)
	// Although we are locking, consistency suggests we should have loaded state.
	// But simply appending to cache if not loaded might be risky if we later load/overwrite.
	// So let's ensure loaded.
	if err := r.scanPosts(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	postsDir := filepath.Join(r.appDir, "posts")
	postImageDir := filepath.Join(r.appDir, "post-images")
	_ = os.MkdirAll(postsDir, 0755)
	_ = os.MkdirAll(postImageDir, 0755)

	if err := post.Validate(); err != nil {
		return err
	}

	// 强制补充唯一 ID
	if post.ID == "" {
		const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		id, _ := gonanoid.Generate(alphabet, 6)
		post.ID = id
	}

	feature := post.FeatureImagePath

	// Handle Image Copy
	if post.FeatureImage.Name != "" && post.FeatureImage.Path != "" {
		ext := filepath.Ext(post.FeatureImage.Name)
		newPath := filepath.Join(postImageDir, post.FileName+ext)
		// 源路径与目标路径相同时跳过复制，避免 CopyFile 将文件截断为 0 字节
		if sameFilePath(post.FeatureImage.Path, newPath) {
			feature = "/post-images/" + post.FileName + ext
		} else if err := CopyFile(post.FeatureImage.Path, newPath); err == nil {
			feature = "/post-images/" + post.FileName + ext
			if pathInsideDir(post.FeatureImage.Path, postImageDir) {
				_ = os.Remove(post.FeatureImage.Path)
			}
		}
	}
	if feature == "" && post.Feature != "" {
		feature = post.Feature
	}

	post.Feature = feature

	// Prepare YAML
	// Update 时自动设置 UpdatedAt；保留原始 CreatedAt（Create 时两者相同）
	updatedAt := post.CreatedAt
	if isUpdate {
		updatedAt = time.Now()
	}

	meta := postYaml{
		ID:          post.ID,
		Title:       post.Title,
		CreatedAt:   post.CreatedAt.Format(domain.TimeLayout),
		UpdatedAt:   updatedAt.Format(domain.TimeLayout),
		Tags:        post.Tags,
		TagIDs:      post.TagIDs,
		Categories:  post.Categories,
		CategoryIDs: post.CategoryIDs,
		Published:   post.Published,
		HideInList:  post.HideInList,
		Feature:     post.Feature,
		IsTop:       post.IsTop,
	}

	yamlBytes, err := yaml.Marshal(&meta)
	if err != nil {
		return fmt.Errorf("failed to marshal post yaml: %w", err)
	}

	mdContent := fmt.Sprintf("---\n%s---\n\n%s", string(yamlBytes), post.Content)

	postPath := filepath.Join(postsDir, post.FileName+".md")

	if isUpdate {
		if post.DeleteFileName != "" && post.DeleteFileName != post.FileName {
			oldPath := filepath.Join(postsDir, post.DeleteFileName+".md")
			// 标记为 self-write：rename 场景下 fsnotify 会看到 oldPath REMOVE 事件，
			// 让 watcher 跳过这次自激，避免"改文件名保存 → 两次渲染"。
			DefaultWriteGate.MarkSelfWrite(oldPath)
			_ = os.Remove(oldPath)
			// Remove from cache logic below handles "old" file by filtering/looping
		}
	} else {
		if _, err := os.Stat(postPath); err == nil {
			return fmt.Errorf("post file already exists: %s", post.FileName)
		}
	}

	// Idempotent check optimization could be here, but with cache update we probably want to proceed.
	// Write file atomically to prevent data loss
	if err := WriteFileAtomic(postPath, []byte(mdContent), 0644); err != nil {
		return fmt.Errorf("failed to write post file: %w", err)
	}

	// Update Cache
	// If update, finding existing and replacing. If deleteFileName changed, we might have issues if we don't know the original index well or if ID isn't unique.
	// But `post.DeleteFileName` helps us find the old one if renamed.
	// If not renamed, `post.FileName` is the key.

	// Strategy: Filter out old (by filename or deleteFileName), then append new, then sort.
	newCache := make([]domain.Post, 0, len(r.cache)+1)
	targetFileName := post.FileName
	if isUpdate && post.DeleteFileName != "" {
		targetFileName = post.DeleteFileName
	}

	// Remove existing if any
	for _, p := range r.cache {
		// If isUpdate, strictly remove the one we are updating.
		// If Create, check collision? (already checked file existence)
		if isUpdate {
			if p.FileName == targetFileName {
				continue
			}
		}
		// If we are creating, we assume it's new, but safeguard:
		if !isUpdate && p.FileName == post.FileName {
			continue // Should not happen if file check passed, but just in case
		}
		newCache = append(newCache, p)
	}

	newCache = append(newCache, *post)

	// Sort: 置顶优先，再按 CreatedAt 降序
	sort.Slice(newCache, func(i, j int) bool {
		if newCache[i].IsTop != newCache[j].IsTop {
			return newCache[i].IsTop
		}
		return newCache[i].CreatedAt.After(newCache[j].CreatedAt)
	})

	r.cache = newCache
	r.saveCacheJSON() // Side effect

	return nil
}

func (r *postRepository) Delete(ctx context.Context, fileName string) error {
	fileName = normalizeFileName(fileName)
	if err := r.scanPosts(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	postsDir := filepath.Join(r.appDir, "posts")
	postPath := filepath.Join(postsDir, fileName+".md")

	// Read file logic to cleanup images (preserved)
	content, err := os.ReadFile(postPath)
	if err == nil {
		post, _ := r.parsePost(string(content), fileName+".md")
		if post.Feature != "" && !strings.HasPrefix(post.Feature, "http") {
			featurePath := filepath.Join(r.appDir, strings.TrimPrefix(post.Feature, "/"))
			_ = os.Remove(featurePath)
		}
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

	// 标记为 self-write 后再 Remove：前端 DeletePostFromFrontend 已经显式 emit
	// app-site-reload，watcher 不需要对这次 REMOVE 事件重复触发渲染。
	DefaultWriteGate.MarkSelfWrite(postPath)
	if err := os.Remove(postPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// 文件不在磁盘（已被删/路径不符）——归一后这通常是幂等删除，记一条以防隐性不一致。
		log.Printf("[postRepo] delete: file already absent on disk: %s", postPath)
	}

	// Update Cache
	newCache := make([]domain.Post, 0, len(r.cache))
	for _, p := range r.cache {
		if p.FileName != fileName {
			newCache = append(newCache, p)
		}
	}
	r.cache = newCache
	r.saveCacheJSON()

	return nil
}

func (r *postRepository) GetByFileName(ctx context.Context, fileName string) (*domain.Post, error) {
	fileName = normalizeFileName(fileName)
	if err := r.scanPosts(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.cache {
		if p.FileName == fileName {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("post not found: %s", fileName)
}

// List 按分页返回文章。用于真实的分页展示场景（如 MCP 列表接口）。
// 注意：若调用方需要「全部文章」，请使用 GetAll —— 不要用 List(ctx, 1, 大数字) 的惯用法，
// 这种写法有静默丢弃数据的风险。
func (r *postRepository) List(ctx context.Context, page, size int) ([]domain.Post, int64, error) {
	if err := r.scanPosts(); err != nil {
		return nil, 0, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	total := int64(len(r.cache))
	start := (page - 1) * size
	if start < 0 {
		start = 0
	}
	if start >= len(r.cache) {
		return []domain.Post{}, total, nil
	}
	end := start + size
	if end > len(r.cache) {
		end = len(r.cache)
	}

	// Return copy to prevent external mutation affecting cache
	result := make([]domain.Post, end-start)
	copy(result, r.cache[start:end])

	return result, total, nil
}

// GetAll 返回所有文章的副本（按 IsTop desc、CreatedAt desc 排序，与 List 一致）。
// 供渲染、级联修改、数据迁移等需要全量遍历的场景使用。
func (r *postRepository) GetAll(ctx context.Context) ([]domain.Post, error) {
	if err := r.scanPosts(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Clone(r.cache), nil
}

func (r *postRepository) saveCacheJSON() {
	dbPath := filepath.Join(r.appDir, "config", "posts.json")
	db := map[string]interface{}{"posts": r.cache}
	_ = SaveJSONFileIdempotent(dbPath, db)
}

// normalizeFileName 去掉文件名末尾的 .md（含历史遗留的多重 .md）。
// repo 内部不变量：FileName 始终为裸名——save 写盘时补 .md、parsePost 读盘时去 .md，
// 二者约定一致。所有外部入口（save / Delete / GetByFileName）先归一，避免调用方
// 传 "x.md" 造成 "x.md.md" 落盘或缓存键失同步。
func normalizeFileName(name string) string {
	for strings.HasSuffix(name, ".md") {
		name = strings.TrimSuffix(name, ".md")
	}
	return name
}

func (r *postRepository) parsePost(content string, filename string) (domain.Post, error) {

	// Use regex to extract frontmatter and content.
	// (?s) allows . to match newlines
	// ^\s* allows leading whitespace
	// ---\s* matches start separator
	// \n(.+?)\n matches YAML content non-greedily
	// \s*---\s* matches end separator
	// (.*)$ matches optional body
	re := regexp.MustCompile(`(?s)^\s*---\s*\n(.+?)\n\s*---\s*(?:$|\n(.*))`)
	matches := re.FindStringSubmatch(content)

	var yamlPart, bodyPart string

	if len(matches) >= 2 {
		yamlPart = matches[1]
		if len(matches) > 2 {
			bodyPart = matches[2]
		}
	} else {
		// Fallback for files that might not strictly match regex
		parts := strings.SplitN(content, "---", 3)
		if len(parts) < 3 {
			// Handle case where split fails (e.g. valid file but regex mismatch?)
			// Or completely invalid.
			return domain.Post{}, fmt.Errorf("invalid post format: %s", filename)
		}
		yamlPart = parts[1]
		bodyPart = parts[2]
	}

	var meta postYaml
	if err := yaml.Unmarshal([]byte(yamlPart), &meta); err != nil {
		return domain.Post{}, fmt.Errorf("failed to parse yaml in %s: %w", filename, err)
	}

	postContent := strings.TrimSpace(bodyPart)
	abstract := r.extractAbstract(postContent)

	// 解析 CreatedAt (优先取 CreatedAt，若无则取老的 Date)
	primaryDateStr := meta.CreatedAt
	if primaryDateStr == "" {
		primaryDateStr = meta.Date
	}
	// YAML 中存储的是不带时区的本地时间字符串，必须用 ParseInLocation 按本地时区解析
	parsedDate, err := time.ParseInLocation(domain.TimeLayout, primaryDateStr, time.Local)
	if err != nil {
		parsedDate = time.Now()
	}

	// 解析 UpdatedAt；若 “updated” 字段缺失则降级为 CreatedAt
	updatedAt, err := time.ParseInLocation(domain.TimeLayout, meta.UpdatedAt, time.Local)
	if err != nil || updatedAt.IsZero() {
		updatedAt = parsedDate
	}

	// Update Post struct
	post := domain.Post{
		ID:          meta.ID,
		Title:       meta.Title,
		CreatedAt:   parsedDate,
		UpdatedAt:   updatedAt,
		Tags:        meta.Tags,
		TagIDs:      meta.TagIDs,
		Categories:  meta.Categories,
		CategoryIDs: meta.CategoryIDs,
		Published:   meta.Published,
		HideInList:  meta.HideInList,
		Feature:     meta.Feature,
		IsTop:       meta.IsTop,
		Content:     postContent,
		Abstract:    abstract,
		FileName:    normalizeFileName(filename),
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
