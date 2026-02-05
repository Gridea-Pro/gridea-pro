package domain

import (
	"context"
)

// PostData 文章元数据
type PostData struct {
	Title      string   `json:"title" yaml:"title"`
	Date       string   `json:"date" yaml:"date"`
	Tags       []string `json:"tags" yaml:"tags"`
	TagIDs     []string `json:"tagIds" yaml:"tag_ids"` // Hidden field for robust linking
	Categories []string `json:"categories" yaml:"categories"`
	Published  bool     `json:"published" yaml:"published"`
	HideInList bool     `json:"hideInList" yaml:"hideInList"`
	Feature    string   `json:"feature" yaml:"feature"`
	IsTop      bool     `json:"isTop" yaml:"isTop"`
}

// Post 文章结构
type Post struct {
	Data     PostData `json:"data"`
	Content  string   `json:"content"`
	Abstract string   `json:"abstract"`
	FileName string   `json:"fileName"`
}

// PostInput 文章输入（DTO for Create/Update）
type PostInput struct {
	Title            string   `json:"title"`
	Date             string   `json:"date"`
	Tags             []string `json:"tags"`
	TagIDs           []string `json:"tagIds"`
	Categories       []string `json:"categories"`
	Published        bool     `json:"published"`
	HideInList       bool     `json:"hideInList"`
	IsTop            bool     `json:"isTop"`
	Content          string   `json:"content"`
	FileName         string   `json:"fileName"`
	DeleteFileName   string   `json:"deleteFileName"`
	FeatureImage     FileInfo `json:"featureImage"`
	FeatureImagePath string   `json:"featureImagePath"`
}

// PostRepository 定义文章存储接口
type PostRepository interface {
	// GetAll 获取所有文章
	GetAll(ctx context.Context) ([]Post, error)
	// Save 保存文章
	Save(ctx context.Context, post *PostInput) error
	// Delete 删除文章
	Delete(ctx context.Context, fileName string) error
	// GetByFileName 获取单篇文章
	GetByFileName(ctx context.Context, fileName string) (*Post, error)
}
