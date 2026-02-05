package domain

import "context"

// Menu 菜单结构
type Menu struct {
	Name     string `json:"name"`
	Link     string `json:"link"`
	OpenType string `json:"openType"`
}

// MenuRepository 定义菜单存储接口
type MenuRepository interface {
	GetAll(ctx context.Context) ([]Menu, error)
	SaveAll(ctx context.Context, menus []Menu) error
}
