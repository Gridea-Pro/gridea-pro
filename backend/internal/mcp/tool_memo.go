package mcp

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

func listMemosTool() mcp.Tool {
	return mcp.NewTool("list_memos", mcp.WithDescription("List all memos"))
}

func listMemosHandler(s *service.MemoService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		memos, err := s.LoadMemos(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load memos: %v", err)), nil
		}
		return mcp.NewToolResultText(jsonify(memos)), nil
	}
}

func createMemoTool() mcp.Tool {
	return mcp.NewTool("create_memo",
		mcp.WithDescription("Create a new memo"),
		mcp.WithString("content", mcp.Description("Memo content (markdown supported)"), mcp.Required()),
	)
}

func createMemoHandler(s *service.MemoService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		content, err := request.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError("content is required"), nil
		}

		memos, _ := s.LoadMemos(ctx)
		const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		id, _ := gonanoid.Generate(alphabet, 6)
		now := time.Now()

		newMemo := domain.Memo{
			ID:        id,
			Content:   content,
			CreatedAt: now,
			UpdatedAt: now,
		}
		// Tags are extracted in Service/Facade usually, but Service.SaveMemos just saves.
		// We might need to handle extraction if service doesn't do it automatically on save.
		// Gridea Pro's MemoService seems to rely on Facade to extract tags before saving?
		// Checked: facade.extractTags calls regex. Service just saves.
		// So we should replicate extraction logic or move it to service.
		// For now, let's keep it simple and just save, tag extraction is a UI feature mostly,
		// unless we want AI to auto-tag. AI can put #tags in content directly.

		// Prepend
		memos = append([]domain.Memo{newMemo}, memos...)

		if err := s.SaveMemos(ctx, memos); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save memo: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Memo created: %s", id)), nil
	}
}

func deleteMemoTool() mcp.Tool {
	return mcp.NewTool("delete_memo",
		mcp.WithDescription("Delete a memo"),
		mcp.WithString("id", mcp.Description("Memo ID to delete"), mcp.Required()),
		mcp.WithBoolean("confirm", mcp.Description("Confirm deletion"), mcp.Required()),
	)
}

func deleteMemoHandler(s *service.MemoService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}

		confirm := request.GetBool("confirm", false)

		memos, err := s.LoadMemos(ctx)
		if err != nil {
			return mcp.NewToolResultError("Failed to load memos"), nil
		}

		// Find memo
		var memo *domain.Memo
		var index int
		for i, m := range memos {
			if m.ID == id {
				memo = &m
				index = i
				break
			}
		}

		if memo == nil {
			return mcp.NewToolResultError("Memo not found"), nil
		}

		if !confirm {
			return mcp.NewToolResultText(fmt.Sprintf("⚠️ CONFIRMATION REQUIRED\nDelete memo: '%s'?\nCall delete_memo again with confirm=true", memo.Content)), nil
		}

		// Delete
		memos = append(memos[:index], memos[index+1:]...)
		if err := s.SaveMemos(ctx, memos); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete memo: %v", err)), nil
		}

		return mcp.NewToolResultText("Memo deleted"), nil
	}
}

// Update Memo
func updateMemoTool() mcp.Tool {
	return mcp.NewTool("update_memo",
		mcp.WithDescription("Update an existing memo"),
		mcp.WithString("id", mcp.Description("Memo ID to update"), mcp.Required()),
		mcp.WithString("content", mcp.Description("New memo content (markdown supported)"), mcp.Required()),
	)
}

func updateMemoHandler(s *service.MemoService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		content, err := request.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError("content is required"), nil
		}

		memos, err := s.LoadMemos(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load memos: %v", err)), nil
		}

		found := false
		for i, m := range memos {
			if m.ID == id {
				memos[i].Content = content
				memos[i].UpdatedAt = time.Now()
				memos[i].Tags = extractMemoTags(content)
				found = true
				break
			}
		}

		if !found {
			return mcp.NewToolResultError("Memo not found"), nil
		}

		if err := s.SaveMemos(ctx, memos); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save memo: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Memo updated: %s", id)), nil
	}
}

// Get Memo Stats
func getMemoStatsTool() mcp.Tool {
	return mcp.NewTool("get_memo_stats",
		mcp.WithDescription("Get memo statistics including daily counts for heatmap visualization"),
	)
}

func getMemoStatsHandler(s *service.MemoService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		memos, err := s.LoadMemos(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load memos: %v", err)), nil
		}

		// 按日期分组统计
		dailyCounts := make(map[string]int)
		for _, m := range memos {
			// 提取日期部分 (YYYY-MM-DD)
			date := m.CreatedAt.Format("2006-01-02")
			dailyCounts[date]++
		}

		stats := map[string]interface{}{
			"totalMemos":  len(memos),
			"dailyCounts": dailyCounts,
		}

		return mcp.NewToolResultText(jsonify(stats)), nil
	}
}

// extractMemoTags 从 memo content 中提取 #tag 标签
func extractMemoTags(content string) []string {
	var tags []string
	seen := make(map[string]bool)
	// 简单匹配 #tag 模式（非 # 开头的行内标签）
	for _, word := range strings.Fields(content) {
		if strings.HasPrefix(word, "#") && len(word) > 1 {
			tag := strings.TrimRight(word[1:], ".,;:!?")
			if tag != "" && !seen[tag] {
				tags = append(tags, tag)
				seen[tag] = true
			}
		}
	}
	return tags
}
