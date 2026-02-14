package mcp

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/service"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Render Site
func renderSiteTool() mcp.Tool {
	return mcp.NewTool("render_site", mcp.WithDescription("Render the static site. Call this after making changes to posts or settings."))
}

func renderSiteHandler(s *service.RendererService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := s.RenderAll(ctx); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Render failed: %v", err)), nil
		}
		return mcp.NewToolResultText("Site rendered successfully"), nil
	}
}
