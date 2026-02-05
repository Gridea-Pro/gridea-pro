package facade

import (
	"context"
	"encoding/json"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// LinkFacade wraps LinkService
type LinkFacade struct {
	internal *service.LinkService
}

func NewLinkFacade(s *service.LinkService) *LinkFacade {
	return &LinkFacade{internal: s}
}

func (f *LinkFacade) LoadLinks() ([]domain.Link, error) {
	links, err := f.internal.LoadLinks(context.TODO())
	if err != nil {
		return nil, err
	}

	// Check if any link has missing ID and patch it
	hasMissingID := false
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for i := range links {
		if links[i].ID == "" {
			id, err := gonanoid.Generate(alphabet, 6)
			if err == nil {
				links[i].ID = id
				hasMissingID = true
				fmt.Printf("Patched missing ID for link: %s -> %s\n", links[i].Name, id)
			}
		}
	}

	// Save back if patched
	if hasMissingID {
		if err := f.internal.SaveLinks(context.TODO(), links); err != nil {
			fmt.Printf("Failed to save patched links: %v\n", err)
		} else {
			fmt.Println("Successfully saved patched links with new IDs")
		}
	}

	return links, nil
}

func (f *LinkFacade) SaveLinks(links []domain.Link) error {
	return f.internal.SaveLinks(context.TODO(), links)
}

// RegisterEvents 注册友链相关事件监听器
func (f *LinkFacade) RegisterEvents(ctx context.Context) {
	registerLinkSaveEvent(ctx, f)
	registerLinkDeleteEvent(ctx, f)
	registerLinkSortEvent(ctx, f)
}

// registerLinkSaveEvent 注册友链保存事件
func registerLinkSaveEvent(ctx context.Context, facade *LinkFacade) {
	runtime.EventsOn(ctx, "link-save", func(data ...interface{}) {
		if len(data) == 0 {
			runtime.EventsEmit(ctx, "link-saved", map[string]interface{}{
				"success": false,
				"links":   []domain.Link{},
			})
			return
		}

		// 将前端传来的 map 转换为 JSON，再解析为 Link
		linkMap, ok := data[0].(map[string]interface{})
		if !ok {
			runtime.EventsEmit(ctx, "link-saved", map[string]interface{}{
				"success": false,
				"links":   []domain.Link{},
			})
			return
		}

		fmt.Printf("Received link save request: %+v\n", linkMap)

		// 尝试直接从 map 获取 ID，以防 JSON 解析问题
		var incomingID string
		if idVal, exists := linkMap["id"]; exists {
			if idStr, ok := idVal.(string); ok {
				incomingID = idStr
			}
		}
		fmt.Printf("Incoming Link ID: %s\n", incomingID)

		// 将 map 转换为 JSON bytes
		jsonBytes, err := json.Marshal(linkMap)
		if err != nil {
			runtime.EventsEmit(ctx, "link-saved", map[string]interface{}{
				"success": false,
				"links":   []domain.Link{},
			})
			return
		}

		// 解析为 Link
		var newLink domain.Link
		if err := json.Unmarshal(jsonBytes, &newLink); err != nil {
			runtime.EventsEmit(ctx, "link-saved", map[string]interface{}{
				"success": false,
				"links":   []domain.Link{},
			})
			return
		}

		// 确保 ID 一致性
		if newLink.ID == "" && incomingID != "" {
			newLink.ID = incomingID
		}

		// 加载现有友链列表
		links, err := facade.LoadLinks()
		if err != nil {
			links = []domain.Link{}
		}

		// 查找是否已存在该友链（通过 ID）
		found := false
		for i, link := range links {
			if link.ID == newLink.ID {
				// 更新现有友链
				links[i] = newLink
				found = true
				fmt.Printf("Found existing link, updating: %s\n", link.ID)
				break
			}
		}

		// 如果不存在，添加新友链
		if !found {
			fmt.Printf("Link not found, creating new one\n")
			links = append(links, newLink)
		}

		// 保存友链列表
		if err := facade.SaveLinks(links); err != nil {
			runtime.EventsEmit(ctx, "link-saved", map[string]interface{}{
				"success": false,
				"links":   links,
			})
			return
		}

		// 成功
		runtime.EventsEmit(ctx, "link-saved", map[string]interface{}{
			"success": true,
			"links":   links,
		})
		fmt.Printf("友链保存成功: %s\n", newLink.Name)
	})
}

// registerLinkDeleteEvent 注册友链删除事件
func registerLinkDeleteEvent(ctx context.Context, facade *LinkFacade) {
	runtime.EventsOn(ctx, "link-delete", func(data ...interface{}) {
		if len(data) == 0 {
			runtime.EventsEmit(ctx, "link-deleted", map[string]interface{}{
				"success": false,
				"links":   []domain.Link{},
			})
			return
		}

		// 获取要删除的友链 ID
		linkID, ok := data[0].(string)
		if !ok {
			runtime.EventsEmit(ctx, "link-deleted", map[string]interface{}{
				"success": false,
				"links":   []domain.Link{},
			})
			return
		}

		// 加载现有友链列表
		links, err := facade.LoadLinks()
		if err != nil {
			links = []domain.Link{}
		}

		// 过滤掉要删除的友链
		filteredLinks := make([]domain.Link, 0)
		for _, link := range links {
			if link.ID != linkID {
				filteredLinks = append(filteredLinks, link)
			}
		}

		// 保存更新后的友链列表
		if err := facade.SaveLinks(filteredLinks); err != nil {
			runtime.EventsEmit(ctx, "link-deleted", map[string]interface{}{
				"success": false,
				"links":   filteredLinks,
			})
			return
		}

		// 成功
		runtime.EventsEmit(ctx, "link-deleted", map[string]interface{}{
			"success": true,
			"links":   filteredLinks,
		})
		fmt.Printf("友链删除成功: %s\n", linkID)
	})
}

// registerLinkSortEvent 注册友链排序事件
func registerLinkSortEvent(ctx context.Context, facade *LinkFacade) {
	runtime.EventsOn(ctx, "link-sort", func(data ...interface{}) {
		if len(data) == 0 {
			return
		}

		// 将前端传来的数组转换为 JSON，再解析为 []Link
		linksData, ok := data[0].([]interface{})
		if !ok {
			return
		}

		// 将 []interface{} 转换为 JSON bytes
		jsonBytes, err := json.Marshal(linksData)
		if err != nil {
			return
		}

		// 解析为 []Link
		var links []domain.Link
		if err := json.Unmarshal(jsonBytes, &links); err != nil {
			return
		}

		// 保存排序后的友链列表
		if err := facade.SaveLinks(links); err != nil {
			fmt.Printf("友链排序保存失败: %v\n", err)
			return
		}

		fmt.Println("友链排序保存成功")
	})
}
