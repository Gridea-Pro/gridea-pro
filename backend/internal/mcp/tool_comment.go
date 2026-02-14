package mcp

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/service"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func listCommentsTool() mcp.Tool {
	return mcp.NewTool("list_comments",
		mcp.WithDescription("List recent comments"),
		mcp.WithNumber("page", mcp.Description("Page number (default 1)")),
		mcp.WithNumber("pageSize", mcp.Description("Page size (default 20)")),
	)
}

func listCommentsHandler(s *service.CommentService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		page := request.GetInt("page", 1)
		pageSize := request.GetInt("pageSize", 20)

		result, err := s.FetchComments(ctx, page, pageSize)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch comments: %v", err)), nil
		}

		return mcp.NewToolResultText(jsonify(result)), nil
	}
}

func replyCommentTool() mcp.Tool {
	return mcp.NewTool("reply_comment",
		mcp.WithDescription("Reply to a comment"),
		mcp.WithString("parentInfo", mcp.Description("Parent Comment ID or Info (Internal ID)")),
		// Wait, FetchComments returns domain.PaginatedComments -> Comments.
		// domain.Comment has ID? `backend/internal/domain/comment.go` needs checking if ID is exposed as what.
		// Service `ReplyComment` takes `parentID`, `content`, `articleID`.
		// Let's assume parentID is the ID of the comment we reply to.
		// We need articleID too?
		// `ReplyComment(ctx, parentID, content, articleID)`
		mcp.WithString("parentId", mcp.Description("Parent Comment ID"), mcp.Required()),
		mcp.WithString("articleId", mcp.Description("Article ID (or path)"), mcp.Required()),
		mcp.WithString("content", mcp.Description("Reply content"), mcp.Required()),
	)
}

func replyCommentHandler(s *service.CommentService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		parentId, err := request.RequireString("parentId")
		if err != nil {
			return mcp.NewToolResultError("parentId is required"), nil
		}
		articleId, err := request.RequireString("articleId")
		if err != nil {
			return mcp.NewToolResultError("articleId is required"), nil
		}
		content, err := request.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError("content is required"), nil
		}

		if err := s.ReplyComment(ctx, parentId, content, articleId); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to reply: %v", err)), nil
		}

		return mcp.NewToolResultText("Reply sent"), nil
	}
}

func deleteCommentTool() mcp.Tool {
	return mcp.NewTool("delete_comment",
		mcp.WithDescription("Delete a comment"),
		mcp.WithString("id", mcp.Description("Comment ID"), mcp.Required()),
		mcp.WithBoolean("confirm", mcp.Description("Confirm deletion"), mcp.Required()),
	)
}

func deleteCommentHandler(s *service.CommentService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		confirm := request.GetBool("confirm", false)

		if !confirm {
			return mcp.NewToolResultText(fmt.Sprintf("⚠️ Confirm delete comment '%s'?", id)), nil
		}

		if err := s.DeleteComment(ctx, id); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete comment: %v", err)), nil
		}

		return mcp.NewToolResultText("Comment deleted"), nil
	}
}
