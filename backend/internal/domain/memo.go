package domain

import "context"

// Memo 闪念结构体
type Memo struct {
	ID        string   `json:"id"`        // NanoID (6字符)
	Content   string   `json:"content"`   // Markdown 内容
	Tags      []string `json:"tags"`      // 从内容中提取的标签
	Images    []string `json:"images"`    // 图片路径 (V2预留)
	CreatedAt string   `json:"createdAt"` // 创建时间
	UpdatedAt string   `json:"updatedAt"` // 更新时间
}

// MemoRepository 定义Memos存储接口
type MemoRepository interface {
	GetAll(ctx context.Context) ([]Memo, error)
	SaveAll(ctx context.Context, memos []Memo) error
}

// TagStat box struct for tag statistics
type TagStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// MemoStats struct for statistics
type MemoStats struct {
	Total   int            `json:"total"`
	Tags    []TagStat      `json:"tags"`
	Heatmap map[string]int `json:"heatmap"`
}

// MemoDashboardDTO for frontend dashboard
type MemoDashboardDTO struct {
	Memos []Memo    `json:"memos"`
	Stats MemoStats `json:"stats"`
}
