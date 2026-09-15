package facade

import (
	"context"
	"encoding/base64"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
	"gridea-pro/backend/internal/utils"
	"sync/atomic"
	"time"
)

// PostFacade wraps PostService
type PostFacade struct {
	// internal 用原子读写保存：切站（UpdateAppDir）会热替换它，
	// 而 Wails 可能并发调用本 facade 的方法，原子读写避免"读到替换一半的状态"的 data race。
	internal atomic.Pointer[service.PostService]
}

func NewPostFacade(s *service.PostService) *PostFacade {
	f := &PostFacade{}
	f.internal.Store(s)
	return f
}

func (f *PostFacade) svc() *service.PostService { return f.internal.Load() }

// PostDashboardDTO defines the data structure for post dashboard
type PostDashboardDTO struct {
	Posts []domain.Post `json:"posts"`
	Tags  []domain.Tag  `json:"tags"`
}

// PostForm DTO for frontend input
type PostForm struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	CreatedAt        string          `json:"createdAt"`
	Tags             []string        `json:"tags"`
	TagIDs           []string        `json:"tagIds"`
	Categories       []string        `json:"categories"`
	CategoryIDs      []string        `json:"categoryIds"` // 分类 Slug 列表
	Published        bool            `json:"published"`
	HideInList       bool            `json:"hideInList"`
	IsTop            bool            `json:"isTop"`
	Content          string          `json:"content"`
	FileName         string          `json:"fileName"`
	DeleteFileName   string          `json:"deleteFileName"`
	FeatureImage     domain.FileInfo `json:"featureImage"`
	FeatureImagePath string          `json:"featureImagePath"`
}

func (f *PostFacade) LoadPosts() ([]domain.Post, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().LoadPosts(ctx)
}

func (f *PostFacade) LoadTags() ([]domain.Tag, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().LoadTags(ctx)
}

func (f *PostFacade) SavePost(form PostForm) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}

	post, err := f.mapFormToPost(form)
	if err != nil {
		return err
	}

	return f.svc().SavePost(ctx, post)
}

func (f *PostFacade) DeletePost(fileName string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().DeletePost(ctx, fileName)
}

func (f *PostFacade) UploadImages(files []domain.UploadedFile) ([]string, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().UploadImages(ctx, files)
}

// SavePostFromFrontend handles post saving from the frontend
func (f *PostFacade) SavePostFromFrontend(form PostForm) (*PostDashboardDTO, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}

	// 一次调用内只取一次 svc 快照，避免"保存到旧站、返回新站文章/标签列表"被切站劈成两半。
	svc := f.svc()

	post, err := f.mapFormToPost(form)
	if err != nil {
		return nil, err
	}

	if err := svc.SavePost(ctx, post); err != nil {
		return nil, err
	}

	posts, err := svc.LoadPosts(ctx)
	if err != nil {
		return nil, err
	}
	tags, err := svc.LoadTags(ctx)
	if err != nil {
		return nil, err
	}

	return &PostDashboardDTO{
		Posts: posts,
		Tags:  tags,
	}, nil
}

// DeletePostFromFrontend handles post deletion from the frontend
func (f *PostFacade) DeletePostFromFrontend(fileName string) ([]domain.Post, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}

	svc := f.svc()
	if err := svc.DeletePost(ctx, fileName); err != nil {
		return nil, err
	}
	return svc.LoadPosts(ctx)
}

// UploadImagesFromFrontend handles image uploading from the frontend
func (f *PostFacade) UploadImagesFromFrontend(files []domain.UploadedFile) ([]string, error) {
	// reuse wrapper with context logic
	return f.UploadImages(files)
}

// SaveImageBytesFromFrontend 保存粘贴 / 拖拽的图片字节（前端以 base64 传入），返回相对路径 /post-images/xxx
func (f *PostFacade) SaveImageBytesFromFrontend(name string, dataBase64 string) (string, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	data, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return "", fmt.Errorf("图片数据解码失败: %w", err)
	}
	return f.svc().SaveImageBytes(ctx, name, data)
}

// Helper to map Form to Domain Entity
func (f *PostFacade) mapFormToPost(form PostForm) (*domain.Post, error) {
	// Time Parsing with proper error handling
	var parsedDate time.Time
	if form.CreatedAt == "" {
		parsedDate = time.Now()
	} else {
		var err error
		parsedDate, err = utils.ParseTime(form.CreatedAt, time.Local)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
	}

	return &domain.Post{
		ID:               form.ID,
		Title:            form.Title,
		CreatedAt:        parsedDate,
		Tags:             form.Tags,
		TagIDs:           form.TagIDs,
		Categories:       form.Categories,
		CategoryIDs:      form.CategoryIDs,
		Published:        form.Published,
		HideInList:       form.HideInList,
		IsTop:            form.IsTop,
		Content:          form.Content,
		FileName:         form.FileName,
		DeleteFileName:   form.DeleteFileName,
		FeatureImage:     form.FeatureImage,
		FeatureImagePath: form.FeatureImagePath,
		// Feature 由 Repository 层根据 FeatureImage / FeatureImagePath 综合推导，此处不强制覆盖
	}, nil
}
