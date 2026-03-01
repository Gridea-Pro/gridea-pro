package domain

import (
	"context"
	"errors"
	"time"
)

// Memo 闪念实体 (Pure Entity)
// Added json tags for frontend compatibility.
type Memo struct {
	ID        string    `json:"_id"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	Images    []string  `json:"images"`
	CreatedAt time.Time `json:"createdAt" ts_type:"string"`
	UpdatedAt time.Time `json:"updatedAt" ts_type:"string"`
}

// Validate 校验闪念数据
func (m *Memo) Validate() error {
	if m.Content == "" {
		return errors.New("memo content cannot be empty")
	}
	return nil
}

// MemoRepository 定义Memos存储接口 (Standard CRUD)
type MemoRepository interface {
	// Create 创建闪念
	Create(ctx context.Context, memo *Memo) error

	// Update 更新闪念
	Update(ctx context.Context, id string, memo *Memo) error

	// Delete 删除闪念
	Delete(ctx context.Context, id string) error

	// GetByID 获取单个闪念
	GetByID(ctx context.Context, id string) (*Memo, error)

	// List 获取闪念列表
	List(ctx context.Context) ([]Memo, error)
	// SaveAll 批量保存闪念
	SaveAll(ctx context.Context, memos []Memo) error
}

// TagStat 标签统计
type TagStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// MemoStats 闪念统计
type MemoStats struct {
	Total   int            `json:"total"`
	Tags    []TagStat      `json:"tags"`
	Heatmap map[string]int `json:"heatmap"`
}

type MemoDashboardDTO struct {
	Memos []Memo    `json:"memos"`
	Stats MemoStats `json:"stats"`
}
