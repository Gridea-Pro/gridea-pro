package domain

import (
	"context"
	"errors"
)

// Menu 菜单实体 (Pure Entity)
// Added json tags for frontend compatibility.
type Menu struct {
	ID       string `json:"_id"` // NanoID or UUID, added for granular CRUD
	Name     string `json:"name"`
	Link     string `json:"link"`
	OpenType string `json:"openType"` // "_blank" or "_self"
}

// Validate 校验菜单数据
func (m *Menu) Validate() error {
	if m.Name == "" {
		return errors.New("menu name cannot be empty")
	}
	if m.Link == "" {
		return errors.New("menu link cannot be empty")
	}
	return nil
}

// MenuRepository 定义菜单存储接口 (Standard CRUD)
type MenuRepository interface {
	// Create 创建菜单
	Create(ctx context.Context, menu *Menu) error

	// Update 更新菜单
	Update(ctx context.Context, id string, menu *Menu) error

	// Delete 删除菜单
	Delete(ctx context.Context, id string) error

	// GetByID 获取单个菜单
	GetByID(ctx context.Context, id string) (*Menu, error)

	// List 获取菜单列表
	List(ctx context.Context) ([]Menu, error)
	SaveAll(ctx context.Context, menus []Menu) error
}

// GetID implements Identifiable interface
func (m Menu) GetID() string {
	return m.ID
}
