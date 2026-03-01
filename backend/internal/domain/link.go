package domain

import (
	"context"
	"errors"
)

// Link 友链实体 (Pure Entity)
// Added json tags for frontend compatibility.
type Link struct {
	ID          string `json:"_id"` // NanoID or UUID
	Name        string `json:"name"`
	Url         string `json:"url"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
}

// Validate 校验友链数据
func (l *Link) Validate() error {
	if l.Name == "" {
		return errors.New("link name cannot be empty")
	}
	if l.Url == "" {
		return errors.New("link url cannot be empty")
	}
	return nil
}

// LinkRepository 定义友链存储接口 (Standard CRUD)
type LinkRepository interface {
	// Create 创建友链
	Create(ctx context.Context, link *Link) error

	// Update 更新友链
	Update(ctx context.Context, id string, link *Link) error

	// Delete 删除友链
	Delete(ctx context.Context, id string) error

	// GetByID 获取单个友链
	GetByID(ctx context.Context, id string) (*Link, error)

	// List 获取友链列表
	List(ctx context.Context) ([]Link, error)
	SaveAll(ctx context.Context, links []Link) error
}

// GetID implements Identifiable interface
func (l Link) GetID() string {
	return l.ID
}
