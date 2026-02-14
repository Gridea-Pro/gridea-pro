package facade

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
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

func (f *PostFacade) SavePost(input domain.PostInput) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SavePost(ctx, &input)
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
func (f *PostFacade) SavePostFromFrontend(input domain.PostInput) (*PostDashboardDTO, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}

	if err := f.internal.SavePost(ctx, &input); err != nil {
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
