package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

// List Posts
func listPostsTool() mcp.Tool {
	return mcp.NewTool("list_posts", mcp.WithDescription("List all posts with metadata"))
}

func listPostsHandler(s *service.PostService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		posts, err := s.LoadPosts(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load posts: %v", err)), nil
		}

		// Simplify output to save tokens, or return full? Let's return full for now but maybe trim content
		// Actually, standard list should probably be metadata only
		var simplified []map[string]interface{}
		for _, p := range posts {
			simplified = append(simplified, map[string]interface{}{
				"fileName":  p.FileName,
				"title":     p.Title,
				"date":      p.CreatedAt,
				"tags":      p.Tags,
				"published": p.Published,
			})
		}

		return mcp.NewToolResultText(jsonify(simplified)), nil
	}
}

// Get Post
func getPostTool() mcp.Tool {
	return mcp.NewTool("get_post",
		mcp.WithDescription("Get full content of a post by filename"),
		mcp.WithString("filename", mcp.Description("The filename of the post (e.g. 'hello-world.md')"), mcp.Required()),
	)
}

func getPostHandler(s *service.PostService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filename, err := request.RequireString("filename")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("filename is required: %v", err)), nil
		}

		post, err := s.GetByFileName(ctx, filename)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Post not found: %v", err)), nil
		}

		return mcp.NewToolResultText(jsonify(post)), nil
	}
}

// Create Post
func createPostTool() mcp.Tool {
	return mcp.NewTool("create_post",
		mcp.WithDescription("Create a new post"),
		mcp.WithString("title", mcp.Description("Post title"), mcp.Required()),
		mcp.WithString("content", mcp.Description("Post markdown content"), mcp.Required()),
		mcp.WithString("date", mcp.Description("Publish date (YYYY-MM-DD HH:mm:ss)")),
		mcp.WithString("fileName", mcp.Description("Custom filename (optional)")),
		mcp.WithString("tags", mcp.Description("Comma separated tags or JSON array string")), // Simplification: use string or deal with array complexity
		// Actually mcp-go supports explicit types. Let's use array if possible or just string for simplicity in handlers
		mcp.WithString("category", mcp.Description("Category name (one category per post)")),
		mcp.WithBoolean("published", mcp.Description("Whether to publish immediately")),
	)
}

func createPostHandler(s *service.PostService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		title, err := request.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError("title is required"), nil
		}
		content, err := request.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError("content is required"), nil
		}

		post := domain.Post{
			Title:   title,
			Content: content,
		}

		dateStr := request.GetString("date", "")
		if dateStr != "" {
			parsed, _ := time.ParseInLocation(domain.TimeLayout, dateStr, time.Local)
			if parsed.IsZero() {
				parsed, _ = time.ParseInLocation(domain.DateLayout, dateStr, time.Local)
			}
			post.CreatedAt = parsed
		} else {
			post.CreatedAt = time.Now()
		}

		post.FileName = request.GetString("fileName", "")
		post.Published = request.GetBool("published", true)

		if v := request.GetString("tags", ""); v != "" {
			post.Tags = strings.Split(v, ",")
			for i := range post.Tags {
				post.Tags[i] = strings.TrimSpace(post.Tags[i])
			}
		}
		if v := request.GetString("category", ""); v != "" {
			post.Categories = []string{strings.TrimSpace(v)}
		}

		// 如果未提供 fileName，从标题自动生成 slug
		if post.FileName == "" {
			post.FileName = generateSlug(title)
		}

		if err := s.SavePost(ctx, &post); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create post: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Post created successfully: %s", post.Title)), nil
	}
}

// Update Post
func updatePostTool() mcp.Tool {
	return mcp.NewTool("update_post",
		mcp.WithDescription("Update an existing post"),
		mcp.WithString("filename", mcp.Description("Filename of the post to update"), mcp.Required()),
		mcp.WithString("title", mcp.Description("New title")),
		mcp.WithString("content", mcp.Description("New content")),
		mcp.WithString("tags", mcp.Description("New tags (comma separated)")),
		mcp.WithBoolean("published", mcp.Description("Update publish status")),
	)
}

func updatePostHandler(s *service.PostService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filename, err := request.RequireString("filename")
		if err != nil {
			return mcp.NewToolResultError("filename is required"), nil
		}

		// First load existing post
		existing, err := s.GetByFileName(ctx, filename)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Post not found: %s", filename)), nil
		}

		// Prepare input from existing data (Clone/Copy)
		post := *existing // Shallow copy is fine for fields we are modifying individually, but map/slice needs care if we were to modify in place without replacement.
		// Actually, standard usage here is overwriting fields.

		// Override with new values
		if v := request.GetString("title", ""); v != "" {
			post.Title = v
		}
		if v := request.GetString("content", ""); v != "" {
			post.Content = v
		}
		// mcp-go handling of boolean: returns default if missing.
		// We need to know if it was provided or not?
		// current lib doesn't support "check if present".
		// Assuming if explicit tool calls usually provide arguments they want to change.
		// BUT for booleans, false is valid.
		// Workaround: We can't easily detect "missing vs false" with this library helper without inspecting raw JSON if needed.
		// However, for typical "update" tools, users might send all fields or we accept logic "if not provided, keep existing".
		// request.GetBool defaults to false (or checks key existence? No, it takes default).
		// Let's assume if the user explicitly wants to change it they pass it.
		// Since we can't distinguish, we might need to check RawArguments if critical.
		// For now, let's assume published is NOT optional in update if we use Require/Get pattern indiscriminately.
		// OR: use a different pattern.
		// Let's look at `request.Params.Arguments` map directly if available in this library version.
		// It is available as `request.Arguments`.
		if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
			if _, ok := args["published"]; ok {
				post.Published = request.GetBool("published", post.Published)
			}
		}

		if v := request.GetString("tags", ""); v != "" {
			post.Tags = strings.Split(v, ",")
			for i := range post.Tags {
				post.Tags[i] = strings.TrimSpace(post.Tags[i])
			}
		}

		if err := s.SavePost(ctx, &post); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update post: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Post updated: %s", filename)), nil
	}
}

// Delete Post
func deletePostTool() mcp.Tool {
	return mcp.NewTool("delete_post",
		mcp.WithDescription("Delete a post"),
		mcp.WithString("filename", mcp.Description("Filename of the post to delete"), mcp.Required()),
		mcp.WithBoolean("confirm", mcp.Description("Set to true to confirm deletion")),
	)
}

func deletePostHandler(s *service.PostService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filename, err := request.RequireString("filename")
		if err != nil {
			return mcp.NewToolResultError("filename is required"), nil
		}

		confirm := request.GetBool("confirm", false)
		if !confirm {
			// Preview before delete
			post, err := s.GetByFileName(ctx, filename)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Post not found: %s", filename)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("⚠️ CONFIRMATION REQUIRED\nAre you sure you want to delete post '%s'?\nCall delete_post again with confirm=true to proceed.", post.Title)), nil
		}

		if err := s.DeletePost(ctx, filename); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete post: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Post deleted: %s", filename)), nil
	}
}

// Helper
func jsonify(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

// generateSlug 从标题生成 URL-safe 的文件名
func generateSlug(title string) string {
	// 提取 ASCII 部分用于 slug
	var slug []byte
	prevDash := false
	for _, c := range title {
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			slug = append(slug, byte(c))
			prevDash = false
		} else if c >= 'A' && c <= 'Z' {
			slug = append(slug, byte(c+32)) // toLower
			prevDash = false
		} else if (c == ' ' || c == '-' || c == '_') && !prevDash && len(slug) > 0 {
			slug = append(slug, '-')
			prevDash = true
		}
	}
	// 去掉尾部的 '-'
	result := strings.TrimRight(string(slug), "-")

	// 如果 ASCII 内容太少（如纯中文标题），用时间戳 + 随机 ID
	if len(result) < 3 {
		const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		id, _ := gonanoid.Generate(alphabet, 6)
		result = fmt.Sprintf("post-%s-%s", time.Now().Format("20060102"), id)
	}

	return result
}
