package domain

import "context"

// Link 友链结构
type Link struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Url         string `json:"url"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
}

// LinkRepository 定义友链存储接口
type LinkRepository interface {
	GetAll(ctx context.Context) ([]Link, error)
	SaveAll(ctx context.Context, links []Link) error
}
