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

		if err := s.SaveTag(ctx, tag, ""); err != nil { // empty originalName means create
			return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
		}
		return mcp.NewToolResultText("Tag created"), nil
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

		if err := s.SaveCategory(ctx, cat, ""); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
		}
		return mcp.NewToolResultText("Category created"), nil
	}
}

func deleteCategoryTool() mcp.Tool {
	return mcp.NewTool("delete_category",
		mcp.WithDescription("Delete a category"),
		mcp.WithString("slug", mcp.Description("Category slug"), mcp.Required()),
		mcp.WithBoolean("confirm", mcp.Description("Confirm deletion"), mcp.Required()),
	)
}

func deleteCategoryHandler(s *service.CategoryService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug, err := request.RequireString("slug")
		if err != nil {
			return mcp.NewToolResultError("slug is required"), nil
		}
		confirm := request.GetBool("confirm", false)

		if !confirm {
			return mcp.NewToolResultText(fmt.Sprintf("⚠️ Confirm delete category '%s'?", slug)), nil
		}

		if err := s.DeleteCategory(ctx, slug); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
		}
		return mcp.NewToolResultText("Category deleted"), nil
	}
}
