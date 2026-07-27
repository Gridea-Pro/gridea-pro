package mcp

import (
	"context"
	"fmt"
	"time"

	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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

		// MCP 创建闪念沿用当前时间（传零值）
		memo, err := s.CreateMemo(ctx, content, time.Time{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create memo: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Memo created: %s", memo.ID)), nil
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

		memo := domain.Memo{
			ID:      id,
			Content: content,
		}

		if err := s.UpdateMemo(ctx, memo); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update memo: %v", err)), nil
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
