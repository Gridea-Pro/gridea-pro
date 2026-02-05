package domain

import "context"

// Tag 标签结构
type Tag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Used  bool   `json:"used"`
	Color string `json:"color,omitempty"`
}

// TagRepository 定义标签存储接口
type TagRepository interface {
	GetAll(ctx context.Context) ([]Tag, error)
	// SaveAll 通常在从文章重新计算标签后使用
	SaveAll(ctx context.Context, tags []Tag) error
}
