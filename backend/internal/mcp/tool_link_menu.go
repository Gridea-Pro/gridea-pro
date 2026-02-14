package mcp

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

// --- Link Tools ---

func listLinksTool() mcp.Tool {
	return mcp.NewTool("list_links", mcp.WithDescription("List all friend links"))
}

func listLinksHandler(s *service.LinkService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		links, err := s.LoadLinks(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load links: %v", err)), nil
		}
		return mcp.NewToolResultText(jsonify(links)), nil
	}
}

func createLinkTool() mcp.Tool {
	return mcp.NewTool("create_link",
		mcp.WithDescription("Create a new friend link"),
		mcp.WithString("name", mcp.Description("Site Name"), mcp.Required()),
		mcp.WithString("url", mcp.Description("Site URL"), mcp.Required()),
		mcp.WithString("avatar", mcp.Description("Avatar URL")),
		mcp.WithString("description", mcp.Description("Description")),
	)
}

func createLinkHandler(s *service.LinkService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		url, err := request.RequireString("url")
		if err != nil {
			return mcp.NewToolResultError("url is required"), nil
		}

		links, _ := s.LoadLinks(ctx)
		id, _ := gonanoid.New(9)

		newLink := domain.Link{
			ID:          id,
			Name:        name,
			Url:         url,
			Avatar:      request.GetString("avatar", ""),
			Description: request.GetString("description", ""),
		}

		links = append(links, newLink)

		if err := s.SaveLinks(ctx, links); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save link: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Link created: %s", name)), nil
	}
}

func deleteLinkTool() mcp.Tool {
	return mcp.NewTool("delete_link",
		mcp.WithDescription("Delete a friend link"),
		mcp.WithString("id", mcp.Description("Link ID"), mcp.Required()),
		mcp.WithBoolean("confirm", mcp.Description("Confirm deletion"), mcp.Required()),
	)
}

func deleteLinkHandler(s *service.LinkService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		confirm := request.GetBool("confirm", false)

		if !confirm {
			return mcp.NewToolResultText(fmt.Sprintf("⚠️ Confirm delete link ID '%s'?", id)), nil
		}

		links, err := s.LoadLinks(ctx)
		if err != nil {
			return mcp.NewToolResultError("Failed to load links"), nil
		}

		newLinks := []domain.Link{}
		found := false
		for _, l := range links {
			if l.ID == id {
				found = true
				continue
			}
			newLinks = append(newLinks, l)
		}

		if !found {
			return mcp.NewToolResultError("Link not found"), nil
		}

		if err := s.SaveLinks(ctx, newLinks); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete link: %v", err)), nil
		}

		return mcp.NewToolResultText("Link deleted"), nil
	}
}

// --- Menu Tools ---

func listMenusTool() mcp.Tool {
	return mcp.NewTool("list_menus", mcp.WithDescription("List all menus"))
}

func listMenusHandler(s *service.MenuService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		menus, err := s.LoadMenus(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load menus: %v", err)), nil
		}
		return mcp.NewToolResultText(jsonify(menus)), nil
	}
}

func createMenuTool() mcp.Tool {
	return mcp.NewTool("create_menu",
		mcp.WithDescription("Create a new menu item"),
		mcp.WithString("name", mcp.Description("Menu Name"), mcp.Required()),
		mcp.WithString("url", mcp.Description("Menu URL"), mcp.Required()),
		mcp.WithString("target", mcp.Description("Target (_blank, _self)")),
	)
}

func createMenuHandler(s *service.MenuService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		url, err := request.RequireString("url")
		if err != nil {
			return mcp.NewToolResultError("url is required"), nil
		}

		menus, _ := s.LoadMenus(ctx)

		newMenu := domain.Menu{
			Name:     name,
			Link:     url,
			OpenType: request.GetString("target", "_self"),
		}
		// Assuming domain.Menu doesn't have ID, relies on Index? Or Name/Link combination?
		// Let's check domain.Menu struct.

		menus = append(menus, newMenu)

		if err := s.SaveMenus(ctx, menus); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save menu: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Menu created: %s", name)), nil
	}
}

func deleteMenuTool() mcp.Tool {
	return mcp.NewTool("delete_menu",
		mcp.WithDescription("Delete a menu item"),
		mcp.WithString("name", mcp.Description("Menu Name"), mcp.Required()), // Assuming Name is unique enough for now?
		mcp.WithBoolean("confirm", mcp.Description("Confirm deletion"), mcp.Required()),
	)
}

func deleteMenuHandler(s *service.MenuService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		confirm := request.GetBool("confirm", false)

		if !confirm {
			return mcp.NewToolResultText(fmt.Sprintf("⚠️ Confirm delete menu '%s'?", name)), nil
		}

		menus, err := s.LoadMenus(ctx)
		if err != nil {
			return mcp.NewToolResultError("Failed to load menus"), nil
		}

		newMenus := []domain.Menu{}
		found := false
		for _, m := range menus {
			if m.Name == name { // Naive matching
				found = true
				continue
			}
			newMenus = append(newMenus, m)
		}

		if !found {
			return mcp.NewToolResultError("Menu not found"), nil
		}

		if err := s.SaveMenus(ctx, newMenus); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete menu: %v", err)), nil
		}

		return mcp.NewToolResultText("Menu deleted"), nil
	}
}
