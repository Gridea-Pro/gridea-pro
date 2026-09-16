package domain

import (
	"context"
	"errors"
)

// Category 分类实体 (Pure Entity)
// Added json tags for frontend compatibility.
type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	// Cover 分类封面图，站点内相对路径（如 /post-images/xxx.webp）或外链 URL。
	// 存在站点目录而不是主题目录，换主题不会丢。
	Cover string `json:"cover,omitempty"`
}

// Validate 校验分类数据
func (c *Category) Validate() error {
	if c.Name == "" {
		return errors.New("category name cannot be empty")
	}
	if c.Slug == "" {
		return errors.New("category slug cannot be empty")
	}
	return nil
}

type CategoryRepository interface {
	// Create 创建分类（若 ID 为空则自动生成 UUID）
	Create(ctx context.Context, category *Category) error

	// Update 更新分类（按 ID 查找）
	Update(ctx context.Context, id string, category *Category) error

	// Delete 删除分类（按 ID 删除）
	Delete(ctx context.Context, id string) error

	// GetByID 根据不可变 ID 获取分类
	GetByID(ctx context.Context, id string) (*Category, error)

	// GetBySlug 根据 Slug 获取分类（向后兼容）
	GetBySlug(ctx context.Context, slug string) (*Category, error)

	// List 获取分类列表
	List(ctx context.Context) ([]Category, error)
	SaveAll(ctx context.Context, categories []Category) error
}

// GetID implements Identifiable interface
// 返回不可变 UUID——BaseJSONRepository 的所有 CRUD 都基于此字段
func (c Category) GetID() string {
	return c.ID
}
