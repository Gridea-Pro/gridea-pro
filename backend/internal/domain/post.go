package domain

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Post represents a blog post
type Post struct {
	// Metadata
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	CreatedAt        time.Time `json:"createdAt" ts_type:"string"` // 创建时间（取代老版的 date）
	UpdatedAt        time.Time `json:"updatedAt" ts_type:"string"` // 最后修改时间（每次保存自动更新）
	Tags             []string  `json:"tags"`
	TagIDs           []string  `json:"tagIds"`
	Categories       []string  `json:"categories"`  // 分类名称（降级兜底）
	CategoryIDs      []string  `json:"categoryIds"` // 分类 Slug（主键，优先使用）
	Published        bool      `json:"published"`
	HideInList       bool      `json:"hideInList"`
	IsTop            bool      `json:"isTop"`
	Feature          string    `json:"feature"`
	FeatureImagePath string    `json:"featureImagePath"`
	FeatureImage     FileInfo  `json:"featureImage"`

	// Content
	Content        string `json:"content"`
	FileName       string `json:"fileName"`
	DeleteFileName string `json:"deleteFileName"` // 用于重命名/删除
	Abstract       string `json:"abstract"`
}

// IsSafeFileName 校验文章文件名是否安全：不得包含路径分隔符或 ".." 穿越片段。
// 文件名会被拼进 posts/ 与 post-images/ 目录，未经校验的 "../" 可写/删站点目录外的任意文件。
func IsSafeFileName(name string) bool {
	if name == "" {
		return true // 空值由各自的必填校验负责
	}
	if strings.ContainsAny(name, `/\`) {
		return false
	}
	if name == "." || name == ".." || strings.Contains(name, "..") {
		return false
	}
	return true
}

// Validate validates the post
func (p *Post) Validate() error {
	if strings.TrimSpace(p.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(p.FileName) == "" {
		return errors.New("filename is required")
	}
	if !IsSafeFileName(p.FileName) {
		return errors.New("invalid filename: must not contain path separators or '..'")
	}
	if !IsSafeFileName(p.DeleteFileName) {
		return errors.New("invalid deleteFileName: must not contain path separators or '..'")
	}
	return nil
}

// UnmarshalJSON 自定义反序列化逻辑，专治历史技术债
func (p *Post) UnmarshalJSON(data []byte) error {
	// 1. 定义一个别名类型，避免递归调用 UnmarshalJSON 导致死循环
	type postAlias Post

	// 2. 定义一个匿名结构体，它继承了 Post 的所有字段，并额外“撒网”捕获老版 date 字段
	aux := &struct {
		OldDate time.Time `json:"date"` // 专门用来接住老数据的 date
		*postAlias
	}{
		postAlias: (*postAlias)(p),
	}

	// 3. 解析 JSON 数据
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 4. 兼容转换逻辑：
	// 如果新的 createdAt 是空的（说明是老数据），并且老 date 有值，就转移数据
	if p.CreatedAt.IsZero() && !aux.OldDate.IsZero() {
		p.CreatedAt = aux.OldDate
	}

	// 容错：如果 updatedAt 也是空的，可以用 createdAt 兜底
	if p.UpdatedAt.IsZero() && !p.CreatedAt.IsZero() {
		p.UpdatedAt = p.CreatedAt
	}

	return nil
}

// PostRepository defines the interface for post persistence
type PostRepository interface {
	// Standard CRUD
	Create(ctx context.Context, post *Post) error
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, fileName string) error
	GetByFileName(ctx context.Context, fileName string) (*Post, error)

	// List returns paginated posts.
	// page is 1-based. size is items per page.
	// Returns posts, total count, and error.
	// 注意：若需要取全部文章，请使用 GetAll，不要用 List(ctx, 1, 大数字) 的惯用法。
	List(ctx context.Context, page, size int) ([]Post, int64, error)

	// GetAll returns all posts (sorted by IsTop desc, then CreatedAt desc).
	// 用于渲染、级联修改、数据迁移等需要全量遍历的场景。
	GetAll(ctx context.Context) ([]Post, error)

	// RewritePostTags 在持有仓库写锁的前提下，对每篇文章调用 mutate 就地改写标签字段并即时落盘。
	// 整个"读-改-写"全程持锁，不会与并发 SavePost 交错——用于标签/分类级联改名或删除，
	// 避免级联用旧快照整篇覆盖、抹掉用户此刻刚保存的正文。mutate 返回 true 表示该文章有改动需落盘。
	RewritePostTags(ctx context.Context, mutate func(p *Post) bool) error

	// Reload forces a rescan of the post directory
	Reload(ctx context.Context) error
}
