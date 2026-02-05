package domain

import "context"

// Category 分类结构
type Category struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

// CategoryRepository 定义分类存储接口
type CategoryRepository interface {
	GetAll(ctx context.Context) ([]Category, error)
	SaveAll(ctx context.Context, categories []Category) error
}
