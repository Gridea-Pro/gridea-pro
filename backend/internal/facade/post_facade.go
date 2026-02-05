package facade

import (
	"context"
	"encoding/json"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// PostFacade wraps PostService
type PostFacade struct {
	internal *service.PostService
}

func NewPostFacade(s *service.PostService) *PostFacade {
	return &PostFacade{internal: s}
}

func (f *PostFacade) LoadPosts() ([]domain.Post, error) {
	return f.internal.LoadPosts(context.TODO())
}

func (f *PostFacade) LoadTags() ([]domain.Tag, error) {
	return f.internal.LoadTags(context.TODO())
}

func (f *PostFacade) SavePost(input domain.PostInput) error {
	return f.internal.SavePost(context.TODO(), &input)
}

func (f *PostFacade) DeletePost(fileName string) error {
	return f.internal.DeletePost(context.TODO(), fileName)
}

func (f *PostFacade) UploadImages(files []domain.UploadedFile) ([]string, error) {
	return f.internal.UploadImages(context.TODO(), files)
}

// RegisterEvents registers event listeners for post operations
func (f *PostFacade) RegisterEvents(ctx context.Context) {
	registerPostSaveEvent(ctx, f)
	registerPostListEvent(ctx, f)
	registerPostDeleteEvent(ctx, f)
	registerPostListDeleteEvent(ctx, f)
	registerImageUploadEvent(ctx, f)
}

func registerPostSaveEvent(ctx context.Context, facade *PostFacade) {
	runtime.EventsOn(ctx, "app-post-create", func(data ...interface{}) {
		if len(data) == 0 {
			runtime.EventsEmit(ctx, "app-post-created", map[string]interface{}{
				"success": false,
				"posts":   []domain.Post{},
			})
			return
		}

		postMap, ok := data[0].(map[string]interface{})
		if !ok {
			runtime.EventsEmit(ctx, "app-post-created", map[string]interface{}{
				"success": false,
				"posts":   []domain.Post{},
			})
			return
		}

		jsonBytes, err := json.Marshal(postMap)
		if err != nil {
			runtime.EventsEmit(ctx, "app-post-created", map[string]interface{}{
				"success": false,
				"posts":   []domain.Post{},
			})
			return
		}

		var input domain.PostInput
		if err := json.Unmarshal(jsonBytes, &input); err != nil {
			runtime.EventsEmit(ctx, "app-post-created", map[string]interface{}{
				"success": false,
				"posts":   []domain.Post{},
			})
			return
		}

		if err := facade.SavePost(input); err != nil {
			runtime.EventsEmit(ctx, "app-post-created", map[string]interface{}{
				"success": false,
				"posts":   []domain.Post{},
				"tags":    []domain.Tag{},
			})
			return
		}

		posts, _ := facade.LoadPosts()
		tags, _ := facade.LoadTags()
		runtime.EventsEmit(ctx, "app-post-created", map[string]interface{}{
			"success": true,
			"posts":   posts,
			"tags":    tags,
		})
	})
}

func registerPostListEvent(ctx context.Context, facade *PostFacade) {
	runtime.EventsOn(ctx, "app-post-list", func(data ...interface{}) {
		posts, err := facade.LoadPosts()
		runtime.EventsEmit(ctx, "app-post-list", map[string]interface{}{
			"success": err == nil,
			"posts":   posts,
		})
	})
}

func registerPostDeleteEvent(ctx context.Context, facade *PostFacade) {
	runtime.EventsOn(ctx, "app-post-delete", func(data ...interface{}) {
		if len(data) == 0 {
			return
		}
		fileName, ok := data[0].(string)
		if !ok {
			return
		}
		if err := facade.DeletePost(fileName); err != nil {
			runtime.EventsEmit(ctx, "app-post-deleted", map[string]interface{}{
				"success": false,
				"posts":   []domain.Post{},
			})
			return
		}
		posts, _ := facade.LoadPosts()
		runtime.EventsEmit(ctx, "app-post-deleted", map[string]interface{}{
			"success": true,
			"posts":   posts,
		})
	})
}

func registerPostListDeleteEvent(ctx context.Context, facade *PostFacade) {
	runtime.EventsOn(ctx, "app-post-list-delete", func(data ...interface{}) {
		if len(data) == 0 {
			return
		}

		jsonBytes, err := json.Marshal(data[0])
		if err != nil {
			return
		}

		var posts []domain.Post
		if err := json.Unmarshal(jsonBytes, &posts); err != nil {
			return
		}

		for _, post := range posts {
			_ = facade.DeletePost(post.FileName)
		}

		postsData, _ := facade.LoadPosts()
		runtime.EventsEmit(ctx, "app-post-list-deleted", map[string]interface{}{
			"success": true,
			"posts":   postsData,
		})
	})
}

func registerImageUploadEvent(ctx context.Context, facade *PostFacade) {
	runtime.EventsOn(ctx, "image-upload", func(data ...interface{}) {
		if len(data) == 0 {
			return
		}

		filesData, ok := data[0].([]interface{})
		if !ok {
			return
		}

		jsonBytes, err := json.Marshal(filesData)
		if err != nil {
			return
		}

		var files []domain.UploadedFile
		if err := json.Unmarshal(jsonBytes, &files); err != nil {
			return
		}

		paths, err := facade.UploadImages(files)
		if err != nil {
			runtime.EventsEmit(ctx, "image-uploaded", []string{})
			return
		}

		runtime.EventsEmit(ctx, "image-uploaded", paths)
	})
}
