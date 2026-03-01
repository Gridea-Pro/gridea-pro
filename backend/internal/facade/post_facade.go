package facade

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
	"gridea-pro/backend/internal/utils"
	"time"
)

// PostFacade wraps PostService
type PostFacade struct {
	internal *service.PostService
}

func NewPostFacade(s *service.PostService) *PostFacade {
	return &PostFacade{internal: s}
}

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
	return f.internal.LoadPosts(ctx)
}

func (f *PostFacade) LoadTags() ([]domain.Tag, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.LoadTags(ctx)
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

	return f.internal.SavePost(ctx, post)
}

func (f *PostFacade) DeletePost(fileName string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.DeletePost(ctx, fileName)
}

func (f *PostFacade) UploadImages(files []domain.UploadedFile) ([]string, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.UploadImages(ctx, files)
}

// SavePostFromFrontend handles post saving from the frontend
func (f *PostFacade) SavePostFromFrontend(form PostForm) (*PostDashboardDTO, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}

	post, err := f.mapFormToPost(form)
	if err != nil {
		return nil, err
	}

	if err := f.internal.SavePost(ctx, post); err != nil {
		return nil, err
	}

	posts, err := f.internal.LoadPosts(ctx)
	if err != nil {
		return nil, err
	}
	tags, err := f.internal.LoadTags(ctx)
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

	if err := f.internal.DeletePost(ctx, fileName); err != nil {
		return nil, err
	}
	return f.internal.LoadPosts(ctx)
}

// UploadImagesFromFrontend handles image uploading from the frontend
func (f *PostFacade) UploadImagesFromFrontend(files []domain.UploadedFile) ([]string, error) {
	// reuse wrapper with context logic
	return f.UploadImages(files)
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
		// Feature field might need logic if it comes from FeatureImagePath or FeatureImage?
		// In previous logic, Feature was derived.
		// Here we map Feature from FeatureImagePath as default if Feature is empty in form (Form doesn't have Feature field, relying on Path)
		Feature: form.FeatureImagePath,
	}, nil
}
