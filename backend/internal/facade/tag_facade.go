package facade

import (
	"context"
	"encoding/json"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// TagFacade wraps TagService
type TagFacade struct {
	internal *service.TagService
}

func NewTagFacade(s *service.TagService) *TagFacade {
	return &TagFacade{internal: s}
}

func (f *TagFacade) LoadTags() ([]domain.Tag, error) {
	return f.internal.LoadTags(context.TODO())
}

func (f *TagFacade) SaveTag(tag domain.Tag, originalName string) error {
	return f.internal.SaveTag(context.TODO(), tag, originalName)
}

func (f *TagFacade) DeleteTag(name string) error {
	return f.internal.DeleteTag(context.TODO(), name)
}

// RegisterEvents 注册标签相关事件监听器
func (f *TagFacade) RegisterEvents(ctx context.Context) {
	registerTagSaveEvent(ctx, f)
	registerTagDeleteEvent(ctx, f)
	registerTagSortEvent(ctx)
}

// registerTagSaveEvent 注册标签保存事件
func registerTagSaveEvent(ctx context.Context, facade *TagFacade) {
	runtime.EventsOn(ctx, "tag-save", func(data ...interface{}) {
		if len(data) == 0 {
			runtime.EventsEmit(ctx, "tag-saved", map[string]interface{}{
				"success": false,
				"tags":    []domain.Tag{},
			})
			return
		}

		// 将前端传来的 map 转换为 JSON，再解析为 Tag
		tagMap, ok := data[0].(map[string]interface{})
		if !ok {
			runtime.EventsEmit(ctx, "tag-saved", map[string]interface{}{
				"success": false,
				"tags":    []domain.Tag{},
			})
			return
		}

		// 将 map 转换为 JSON bytes
		jsonBytes, err := json.Marshal(tagMap)
		if err != nil {
			runtime.EventsEmit(ctx, "tag-saved", map[string]interface{}{
				"success": false,
				"tags":    []domain.Tag{},
			})
			return
		}

		// 解析为 Tag form
		var tagForm struct {
			Name         string `json:"name"`
			Slug         string `json:"slug"`
			Color        string `json:"color"`
			OriginalName string `json:"originalName"`
		}
		if err := json.Unmarshal(jsonBytes, &tagForm); err != nil {
			runtime.EventsEmit(ctx, "tag-saved", map[string]interface{}{
				"success": false,
				"tags":    []domain.Tag{},
			})
			return
		}

		newTag := domain.Tag{
			Name:  tagForm.Name,
			Slug:  tagForm.Slug,
			Color: tagForm.Color,
		}

		// 调用 SaveTag 方法（会处理新增和更新）
		if err := facade.SaveTag(newTag, tagForm.OriginalName); err != nil {
			runtime.EventsEmit(ctx, "tag-saved", map[string]interface{}{
				"success": false,
				"tags":    []domain.Tag{},
			})
			return
		}

		// 重新加载标签列表
		tags, err := facade.LoadTags()
		if err != nil {
			tags = []domain.Tag{}
		}

		// 成功
		runtime.EventsEmit(ctx, "tag-saved", map[string]interface{}{
			"success": true,
			"tags":    tags,
		})
		fmt.Printf("标签保存成功: %s\n", tagForm.Name)
	})
}

// registerTagDeleteEvent 注册标签删除事件
func registerTagDeleteEvent(ctx context.Context, facade *TagFacade) {
	runtime.EventsOn(ctx, "tag-delete", func(data ...interface{}) {
		if len(data) == 0 {
			runtime.EventsEmit(ctx, "tag-deleted", map[string]interface{}{
				"success": false,
				"tags":    []domain.Tag{},
			})
			return
		}

		// 获取要删除的标签 slug
		slug, ok := data[0].(string)
		if !ok {
			runtime.EventsEmit(ctx, "tag-deleted", map[string]interface{}{
				"success": false,
				"tags":    []domain.Tag{},
			})
			return
		}

		// 加载现有标签列表
		tags, err := facade.LoadTags()
		if err != nil {
			tags = []domain.Tag{}
		}

		// 根据 slug 找到对应的标签名称
		var tagName string
		for _, tag := range tags {
			if tag.Slug == slug {
				tagName = tag.Name
				break
			}
		}

		if tagName == "" {
			runtime.EventsEmit(ctx, "tag-deleted", map[string]interface{}{
				"success": false,
				"tags":    tags,
			})
			return
		}

		// 调用 DeleteTag 方法（使用标签名称）
		if err := facade.DeleteTag(tagName); err != nil {
			runtime.EventsEmit(ctx, "tag-deleted", map[string]interface{}{
				"success": false,
				"tags":    tags,
			})
			return
		}

		// 重新加载标签列表
		tags, err = facade.LoadTags()
		if err != nil {
			tags = []domain.Tag{}
		}

		// 成功
		runtime.EventsEmit(ctx, "tag-deleted", map[string]interface{}{
			"success": true,
			"tags":    tags,
		})
		fmt.Printf("标签删除成功: %s\n", tagName)
	})
}

// registerTagSortEvent 注册标签排序事件
func registerTagSortEvent(ctx context.Context) {
	runtime.EventsOn(ctx, "tag-sort", func(data ...interface{}) {
		if len(data) == 0 {
			return
		}

		// 前端发送的是完整标签列表，但我们只需要名称顺序
		// TagService 没有 SaveTags 方法，标签排序通过 posts.json 中的 tags 字段来体现
		// 这里我们暂时不处理，因为标签排序逻辑不同于分类和菜单
		fmt.Println("标签排序功能暂不支持")
	})
}
