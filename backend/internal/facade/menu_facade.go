package facade

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

// MenuFacade wraps MenuService
type MenuFacade struct {
	internal *service.MenuService
}

func NewMenuFacade(s *service.MenuService) *MenuFacade {
	return &MenuFacade{internal: s}
}

func (f *MenuFacade) LoadMenus() ([]domain.Menu, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.LoadMenus(ctx)
}

func (f *MenuFacade) SaveMenus(menus []domain.Menu) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SaveMenus(ctx, menus)
}

// MenuForm for frontend usage
type MenuForm struct {
	Name     string      `json:"name"`
	OpenType string      `json:"openType"` // "Internal" or "External"
	Link     string      `json:"link"`
	Index    interface{} `json:"index"` // int or null/undefined, handled as interface{}
}

// SaveMenuFromFrontend accepts a MenuForm directly from frontend
func (f *MenuFacade) SaveMenuFromFrontend(form MenuForm) ([]domain.Menu, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}

	// 加载现有菜单列表
	menus, err := f.internal.LoadMenus(ctx)
	if err != nil {
		menus = []domain.Menu{}
	}

	// 判断是新增还是更新
	var index int = -1
	if form.Index != nil {
		switch v := form.Index.(type) {
		case float64:
			index = int(v)
		case int:
			index = v
		}
	}

	newMenu := domain.Menu{
		Name:     form.Name,
		OpenType: form.OpenType,
		Link:     form.Link,
	}

	if index >= 0 && index < len(menus) {
		// 更新现有菜单，保留子菜单
		menus[index].Name = newMenu.Name
		menus[index].OpenType = newMenu.OpenType
		menus[index].Link = newMenu.Link
	} else {
		// 添加新菜单
		menus = append(menus, newMenu)
	}

	// 保存菜单列表
	if err := f.internal.SaveMenus(ctx, menus); err != nil {
		return nil, err
	}

	// 返回更新后的列表
	return menus, nil
}

// DeleteMenuFromFrontend accepts an index and returns updated list
func (f *MenuFacade) DeleteMenuFromFrontend(index int) ([]domain.Menu, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}

	// 加载现有菜单列表
	menus, err := f.internal.LoadMenus(ctx)
	if err != nil {
		menus = []domain.Menu{}
	}

	// 删除指定索引的菜单
	if index >= 0 && index < len(menus) {
		menus = append(menus[:index], menus[index+1:]...)
	} else {
		return nil, fmt.Errorf("invalid menu index: %d", index)
	}

	// 保存更新后的菜单列表
	if err := f.internal.SaveMenus(ctx, menus); err != nil {
		return nil, err
	}

	// 返回更新后的列表
	return menus, nil
}

// RegisterEvents 注册菜单相关事件监听器
func (f *MenuFacade) RegisterEvents(ctx context.Context) {
	// Events are no longer used for Save/Delete
	// Sort event logic is usually handled by saving the reordered list directly,
	// so frontend can just call SaveMenus (or we can expose a SaveAll method).
	// Actually, MenuService.SaveMenus takes []domain.Menu, which is exactly what Sort needs.
	// So frontend can just call SaveMenus directly if we expose it (which it is exposed).
}
