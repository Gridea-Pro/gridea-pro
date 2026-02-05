package facade

import (
	"context"
	"encoding/json"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// CategoryFacade wraps CategoryService
type CategoryFacade struct {
	internal *service.CategoryService
}

func NewCategoryFacade(s *service.CategoryService) *CategoryFacade {
	return &CategoryFacade{internal: s}
}

func (f *CategoryFacade) LoadCategories() ([]domain.Category, error) {
	return f.internal.LoadCategories(context.TODO())
}

func (f *CategoryFacade) SaveCategories(categories []domain.Category) error {
	return f.internal.SaveCategories(context.TODO(), categories)
}

func (f *CategoryFacade) SaveCategory(category domain.Category, originalSlug string) error {
	return f.internal.SaveCategory(context.TODO(), category, originalSlug)
}

func (f *CategoryFacade) DeleteCategory(slug string) error {
	return f.internal.DeleteCategory(context.TODO(), slug)
}

// RegisterEvents 注册分类相关事件监听器
func (f *CategoryFacade) RegisterEvents(ctx context.Context) {
	registerCategorySaveEvent(ctx, f)
	registerCategoryDeleteEvent(ctx, f)
	registerCategorySortEvent(ctx, f)
}

// registerCategorySaveEvent 注册分类保存事件
func registerCategorySaveEvent(ctx context.Context, facade *CategoryFacade) {
	runtime.EventsOn(ctx, "category-save", func(data ...interface{}) {
		if len(data) == 0 {
			runtime.EventsEmit(ctx, "category-saved", map[string]interface{}{
				"success":    false,
				"categories": []domain.Category{},
			})
			return
		}

		// 将前端传来的 map 转换为 JSON，再解析为 Category
		categoryMap, ok := data[0].(map[string]interface{})
		if !ok {
			runtime.EventsEmit(ctx, "category-saved", map[string]interface{}{
				"success":    false,
				"categories": []domain.Category{},
			})
			return
		}

		// 将 map 转换为 JSON bytes
		jsonBytes, err := json.Marshal(categoryMap)
		if err != nil {
			runtime.EventsEmit(ctx, "category-saved", map[string]interface{}{
				"success":    false,
				"categories": []domain.Category{},
			})
			return
		}

		// 解析为 Category form
		var categoryForm struct {
			Name         string `json:"name"`
			Slug         string `json:"slug"`
			Description  string `json:"description"`
			OriginalSlug string `json:"originalSlug"`
		}
		if err := json.Unmarshal(jsonBytes, &categoryForm); err != nil {
			runtime.EventsEmit(ctx, "category-saved", map[string]interface{}{
				"success":    false,
				"categories": []domain.Category{},
			})
			return
		}

		newCategory := domain.Category{
			Name:        categoryForm.Name,
			Slug:        categoryForm.Slug,
			Description: categoryForm.Description,
		}

		// 调用 SaveCategory 方法（会处理新增和更新）
		if err := facade.SaveCategory(newCategory, categoryForm.OriginalSlug); err != nil {
			runtime.EventsEmit(ctx, "category-saved", map[string]interface{}{
				"success":    false,
				"categories": []domain.Category{},
			})
			return
		}

		// 重新加载分类列表
		categories, err := facade.LoadCategories()
		if err != nil {
			categories = []domain.Category{}
		}

		// 成功
		runtime.EventsEmit(ctx, "category-saved", map[string]interface{}{
			"success":    true,
			"categories": categories,
		})
		fmt.Printf("分类保存成功: %s\n", categoryForm.Name)
	})
}

// registerCategoryDeleteEvent 注册分类删除事件
func registerCategoryDeleteEvent(ctx context.Context, facade *CategoryFacade) {
	runtime.EventsOn(ctx, "category-delete", func(data ...interface{}) {
		if len(data) == 0 {
			runtime.EventsEmit(ctx, "category-deleted", map[string]interface{}{
				"success":    false,
				"categories": []domain.Category{},
			})
			return
		}

		// 获取要删除的分类 slug
		slug, ok := data[0].(string)
		if !ok {
			runtime.EventsEmit(ctx, "category-deleted", map[string]interface{}{
				"success":    false,
				"categories": []domain.Category{},
			})
			return
		}

		// 调用 DeleteCategory 方法
		if err := facade.DeleteCategory(slug); err != nil {
			runtime.EventsEmit(ctx, "category-deleted", map[string]interface{}{
				"success":    false,
				"categories": []domain.Category{},
			})
			return
		}

		// 重新加载分类列表
		categories, err := facade.LoadCategories()
		if err != nil {
			categories = []domain.Category{}
		}

		// 成功
		runtime.EventsEmit(ctx, "category-deleted", map[string]interface{}{
			"success":    true,
			"categories": categories,
		})
		fmt.Printf("分类删除成功: %s\n", slug)
	})
}

// registerCategorySortEvent 注册分类排序事件
func registerCategorySortEvent(ctx context.Context, facade *CategoryFacade) {
	runtime.EventsOn(ctx, "category-sort", func(data ...interface{}) {
		if len(data) == 0 {
			return
		}

		// 将前端传来的数组转换为 JSON，再解析为 []Category
		categoriesData, ok := data[0].([]interface{})
		if !ok {
			return
		}

		// 将 []interface{} 转换为 JSON bytes
		jsonBytes, err := json.Marshal(categoriesData)
		if err != nil {
			return
		}

		// 解析为 []Category
		var categories []domain.Category
		if err := json.Unmarshal(jsonBytes, &categories); err != nil {
			return
		}

		// 保存排序后的分类列表
		if err := facade.SaveCategories(categories); err != nil {
			fmt.Printf("分类排序保存失败: %v\n", err)
			return
		}

		fmt.Println("分类排序保存成功")
	})
}
