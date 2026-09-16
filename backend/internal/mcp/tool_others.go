package mcp

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// --- Tags ---

func listTagsTool() mcp.Tool {
	return mcp.NewTool("list_tags", mcp.WithDescription("List all tags"))
}

func listTagsHandler(s *service.TagService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tags, err := s.LoadTags(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
		}
		return mcp.NewToolResultText(jsonify(tags)), nil
	}
}

func createTagTool() mcp.Tool {
	return mcp.NewTool("create_tag",
		mcp.WithDescription("Create a tag"),
		mcp.WithString("name", mcp.Description("Tag name"), mcp.Required()),
		mcp.WithString("slug", mcp.Description("Tag slug"), mcp.Required()),
		mcp.WithString("color", mcp.Description("Tag color code")),
	)
}

func createTagHandler(s *service.TagService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		slug, err := request.RequireString("slug")
		if err != nil {
			return mcp.NewToolResultError("slug is required"), nil
		}

		tag := domain.Tag{
			Name: name,
			Slug: slug,
		}
		tag.Color = request.GetString("color", "")
		if tag.Color == "" {
			// 未指定颜色时，根据名称哈希从预设颜色中选取
			hash := 0
			for _, c := range name {
				hash += int(c)
			}
			tag.Color = service.TagColors[hash%len(service.TagColors)]
		}

		if err := s.SaveTag(ctx, tag, ""); err != nil { // empty originalName means create
			return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
		}
		return mcp.NewToolResultText("Tag created"), nil
	}
}

func updateTagTool() mcp.Tool {
	return mcp.NewTool("update_tag",
		mcp.WithDescription("Update an existing tag by name"),
		mcp.WithString("name", mcp.Description("Current tag name (used to find the tag)"), mcp.Required()),
		mcp.WithString("newName", mcp.Description("New tag name (optional, omit to keep current)")),
		mcp.WithString("slug", mcp.Description("New slug (optional)")),
		mcp.WithString("color", mcp.Description("New color code (optional)")),
	)
}

func updateTagHandler(s *service.TagService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}

		// 先加载现有标签
		tags, err := s.LoadTags(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load tags: %v", err)), nil
		}

		var found *domain.Tag
		for _, t := range tags {
			if t.Name == name {
				found = &t
				break
			}
		}
		if found == nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tag '%s' not found", name)), nil
		}

		// 应用更新
		updated := *found
		if newName := request.GetString("newName", ""); newName != "" {
			updated.Name = newName
		}
		if slug := request.GetString("slug", ""); slug != "" {
			updated.Slug = slug
		}
		if color := request.GetString("color", ""); color != "" {
			updated.Color = color
		}

		if err := s.SaveTag(ctx, updated, name); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Tag '%s' updated", updated.Name)), nil
	}
}

func deleteTagTool() mcp.Tool {
	return mcp.NewTool("delete_tag",
		mcp.WithDescription("Delete a tag"),
		mcp.WithString("name", mcp.Description("Tag name"), mcp.Required()),
		mcp.WithBoolean("confirm", mcp.Description("Confirm deletion"), mcp.Required()),
	)
}

func deleteTagHandler(s *service.TagService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		confirm := request.GetBool("confirm", false)

		if !confirm {
			return mcp.NewToolResultText(fmt.Sprintf("⚠️ Confirm delete tag '%s'?", name)), nil
		}

		if err := s.DeleteTag(ctx, name); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
		}
		return mcp.NewToolResultText("Tag deleted"), nil
	}
}

// --- Categories ---

func listCategoriesTool() mcp.Tool {
	return mcp.NewTool("list_categories", mcp.WithDescription("List all categories"))
}

func listCategoriesHandler(s *service.CategoryService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		cats, err := s.LoadCategories(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
		}
		return mcp.NewToolResultText(jsonify(cats)), nil
	}
}

func createCategoryTool() mcp.Tool {
	return mcp.NewTool("create_category",
		mcp.WithDescription("Create a category"),
		mcp.WithString("name", mcp.Description("Category name"), mcp.Required()),
		mcp.WithString("slug", mcp.Description("Category slug"), mcp.Required()),
		mcp.WithString("description", mcp.Description("Description")),
		mcp.WithString("cover", mcp.Description("Cover image path or URL (optional)")),
	)
}

func createCategoryHandler(s *service.CategoryService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		slug, err := request.RequireString("slug")
		if err != nil {
			return mcp.NewToolResultError("slug is required"), nil
		}

		cat := domain.Category{
			Name: name,
			Slug: slug,
		}
		cat.Description = request.GetString("description", "")
		cat.Cover = request.GetString("cover", "")

		if err := s.SaveCategory(ctx, cat, ""); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
		}
		return mcp.NewToolResultText("Category created"), nil
	}
}

func updateCategoryTool() mcp.Tool {
	return mcp.NewTool("update_category",
		mcp.WithDescription("Update an existing category by ID"),
		mcp.WithString("id", mcp.Description("Category ID"), mcp.Required()),
		mcp.WithString("name", mcp.Description("New category name (optional)")),
		mcp.WithString("slug", mcp.Description("New slug (optional)")),
		mcp.WithString("description", mcp.Description("New description (optional)")),
		mcp.WithString("cover", mcp.Description("New cover image path or URL (optional)")),
	)
}

func updateCategoryHandler(s *service.CategoryService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}

		existing, err := s.GetByID(ctx, id)
		if err != nil || existing == nil {
			return mcp.NewToolResultError(fmt.Sprintf("Category '%s' not found", id)), nil
		}

		updated := *existing
		if name := request.GetString("name", ""); name != "" {
			updated.Name = name
		}
		if slug := request.GetString("slug", ""); slug != "" {
			updated.Slug = slug
		}
		if desc := request.GetString("description", ""); desc != "" {
			updated.Description = desc
		}
		// 用「参数是否出现」而不是「值是否为空」判断，否则传空串无法清除封面——
		// 界面上有「移除封面」按钮，两个入口的能力要对等。
		if _, ok := request.GetArguments()["cover"]; ok {
			updated.Cover = request.GetString("cover", "")
		}

		if err := s.SaveCategory(ctx, updated, id); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Category '%s' updated", updated.Name)), nil
	}
}

func deleteCategoryTool() mcp.Tool {
	return mcp.NewTool("delete_category",
		mcp.WithDescription("Delete a category"),
		mcp.WithString("id", mcp.Description("Category ID"), mcp.Required()),
		mcp.WithBoolean("confirm", mcp.Description("Confirm deletion"), mcp.Required()),
	)
}

func deleteCategoryHandler(s *service.CategoryService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		confirm := request.GetBool("confirm", false)

		if !confirm {
			return mcp.NewToolResultText(fmt.Sprintf("⚠️ Confirm delete category ID '%s'?", id)), nil
		}

		if err := s.DeleteCategory(ctx, id); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
		}
		return mcp.NewToolResultText("Category deleted"), nil
	}
}
