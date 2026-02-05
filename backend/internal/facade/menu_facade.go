package facade

import (
	"context"
	"encoding/json"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// MenuFacade wraps MenuService
type MenuFacade struct {
	internal *service.MenuService
}

func NewMenuFacade(s *service.MenuService) *MenuFacade {
	return &MenuFacade{internal: s}
}

func (f *MenuFacade) LoadMenus() ([]domain.Menu, error) {
	return f.internal.LoadMenus(context.TODO())
}

func (f *MenuFacade) SaveMenus(menus []domain.Menu) error {
	return f.internal.SaveMenus(context.TODO(), menus)
}

// RegisterEvents 注册菜单相关事件监听器
func (f *MenuFacade) RegisterEvents(ctx context.Context) {
	registerMenuSaveEvent(ctx, f)
	registerMenuDeleteEvent(ctx, f)
	registerMenuSortEvent(ctx, f)
}

// registerMenuSaveEvent 注册菜单保存事件
func registerMenuSaveEvent(ctx context.Context, facade *MenuFacade) {
	runtime.EventsOn(ctx, "menu-save", func(data ...interface{}) {
		if len(data) == 0 {
			runtime.EventsEmit(ctx, "menu-saved", map[string]interface{}{
				"success": false,
				"menus":   []domain.Menu{},
			})
			return
		}

		// 将前端传来的 map 转换为 JSON，再解析为 Menu
		menuMap, ok := data[0].(map[string]interface{})
		if !ok {
			runtime.EventsEmit(ctx, "menu-saved", map[string]interface{}{
				"success": false,
				"menus":   []domain.Menu{},
			})
			return
		}

		// 将 map 转换为 JSON bytes
		jsonBytes, err := json.Marshal(menuMap)
		if err != nil {
			runtime.EventsEmit(ctx, "menu-saved", map[string]interface{}{
				"success": false,
				"menus":   []domain.Menu{},
			})
			return
		}

		// 解析为 Menu form
		var menuForm struct {
			Name     string      `json:"name"`
			OpenType string      `json:"openType"`
			Link     string      `json:"link"`
			Index    interface{} `json:"index"`
		}
		if err := json.Unmarshal(jsonBytes, &menuForm); err != nil {
			runtime.EventsEmit(ctx, "menu-saved", map[string]interface{}{
				"success": false,
				"menus":   []domain.Menu{},
			})
			return
		}

		// 加载现有菜单列表
		menus, err := facade.LoadMenus()
		if err != nil {
			menus = []domain.Menu{}
		}

		// 判断是新增还是更新
		var index int = -1
		if menuForm.Index != nil {
			switch v := menuForm.Index.(type) {
			case float64:
				index = int(v)
			case int:
				index = v
			}
		}

		newMenu := domain.Menu{
			Name:     menuForm.Name,
			OpenType: menuForm.OpenType,
			Link:     menuForm.Link,
		}

		if index >= 0 && index < len(menus) {
			// 更新现有菜单
			menus[index] = newMenu
		} else {
			// 添加新菜单
			menus = append(menus, newMenu)
		}

		// 保存菜单列表
		if err := facade.SaveMenus(menus); err != nil {
			runtime.EventsEmit(ctx, "menu-saved", map[string]interface{}{
				"success": false,
				"menus":   menus,
			})
			return
		}

		// 成功
		runtime.EventsEmit(ctx, "menu-saved", map[string]interface{}{
			"success": true,
			"menus":   menus,
		})
		fmt.Printf("菜单保存成功: %s\n", menuForm.Name)
	})
}

// registerMenuDeleteEvent 注册菜单删除事件
func registerMenuDeleteEvent(ctx context.Context, facade *MenuFacade) {
	runtime.EventsOn(ctx, "menu-delete", func(data ...interface{}) {
		if len(data) == 0 {
			runtime.EventsEmit(ctx, "menu-deleted", map[string]interface{}{
				"success": false,
				"menus":   []domain.Menu{},
			})
			return
		}

		// 获取要删除的菜单索引
		var index int
		switch v := data[0].(type) {
		case float64:
			index = int(v)
		case int:
			index = v
		default:
			runtime.EventsEmit(ctx, "menu-deleted", map[string]interface{}{
				"success": false,
				"menus":   []domain.Menu{},
			})
			return
		}

		// 加载现有菜单列表
		menus, err := facade.LoadMenus()
		if err != nil {
			menus = []domain.Menu{}
		}

		// 删除指定索引的菜单
		if index >= 0 && index < len(menus) {
			menus = append(menus[:index], menus[index+1:]...)
		}

		// 保存更新后的菜单列表
		if err := facade.SaveMenus(menus); err != nil {
			runtime.EventsEmit(ctx, "menu-deleted", map[string]interface{}{
				"success": false,
				"menus":   menus,
			})
			return
		}

		// 成功
		runtime.EventsEmit(ctx, "menu-deleted", map[string]interface{}{
			"success": true,
			"menus":   menus,
		})
		fmt.Printf("菜单删除成功，索引: %d\n", index)
	})
}

// registerMenuSortEvent 注册菜单排序事件
func registerMenuSortEvent(ctx context.Context, facade *MenuFacade) {
	runtime.EventsOn(ctx, "menu-sort", func(data ...interface{}) {
		if len(data) == 0 {
			return
		}

		// 将前端传来的数组转换为 JSON，再解析为 []Menu
		menusData, ok := data[0].([]interface{})
		if !ok {
			return
		}

		// 将 []interface{} 转换为 JSON bytes
		jsonBytes, err := json.Marshal(menusData)
		if err != nil {
			return
		}

		// 解析为 []Menu
		var menus []domain.Menu
		if err := json.Unmarshal(jsonBytes, &menus); err != nil {
			return
		}

		// 保存排序后的菜单列表
		if err := facade.SaveMenus(menus); err != nil {
			fmt.Printf("菜单排序保存失败: %v\n", err)
			return
		}

		fmt.Println("菜单排序保存成功")
	})
}
