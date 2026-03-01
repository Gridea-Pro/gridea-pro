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

// Validate validates the post
func (p *Post) Validate() error {
	if strings.TrimSpace(p.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(p.FileName) == "" {
		return errors.New("filename is required")
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
	List(ctx context.Context, page, size int) ([]Post, int64, error)

	// Deprecated: Use List instead
	GetAll(ctx context.Context) ([]Post, error)

	// Reload forces a rescan of the post directory
	Reload(ctx context.Context) error
}
