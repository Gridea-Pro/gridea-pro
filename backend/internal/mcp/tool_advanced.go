package mcp

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/deploy"
	"gridea-pro/backend/internal/engine"
	"gridea-pro/backend/internal/service"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Render Site
func renderSiteTool() mcp.Tool {
	return mcp.NewTool("render_site", mcp.WithDescription("Render the static site. Call this after making changes to posts or settings."))
}

func renderSiteHandler(s *engine.Engine) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := s.RenderAll(ctx); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Render failed: %v", err)), nil
		}
		return mcp.NewToolResultText("Site rendered successfully"), nil
	}
}

// Deploy Site
func deploySiteTool() mcp.Tool {
	return mcp.NewTool("deploy_site",
		mcp.WithDescription("Deploy the static site to the configured platform (GitHub, Gitee, Vercel, etc.). The site will be rendered before deployment. Requires DEPLOY_ENABLED=true."),
		mcp.WithBoolean("confirm", mcp.Description("Set to true to confirm deployment"), mcp.Required()),
	)
}

func deploySiteHandler(settingService *service.SettingService, renderer *engine.Engine, appDir string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		confirm := request.GetBool("confirm", false)

		setting, err := settingService.GetSetting(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load settings: %v", err)), nil
		}

		if !confirm {
			return mcp.NewToolResultText(fmt.Sprintf(
				"⚠️ CONFIRMATION REQUIRED\nDeploy site to platform '%s' (domain: %s)?\nCall deploy_site again with confirm=true to proceed.",
				setting.Platform, setting.Domain(),
			)), nil
		}

		// 1. Render site
		slog.Info("Rendering site before deployment...")
		if err := renderer.RenderAll(ctx); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Render failed: %v", err)), nil
		}

		// 2. Prepare output directory
		outputDir := filepath.Join(appDir, "output")
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			_ = os.MkdirAll(outputDir, 0755)
		}

		// 3. Select deploy provider
		var provider deploy.Provider
		switch setting.Platform {
		case "github", "gitee", "coding":
			provider = deploy.NewGitProvider()
		case "vercel":
			proxyURL := ""
			if setting.ProxyEnabled {
				proxyURL = setting.ProxyURL
			}
			provider = deploy.NewVercelProvider(proxyURL)
		case "netlify":
			proxyURL := ""
			if setting.ProxyEnabled {
				proxyURL = setting.ProxyURL
			}
			provider = deploy.NewNetlifyProvider(proxyURL)
		case "sftp":
			provider = deploy.NewSftpProvider()
		default:
			provider = deploy.NewGitProvider()
		}

		// 4. Deploy
		logger := func(msg string) {
			slog.Info("deploy", "msg", msg)
		}

		if err := provider.Deploy(ctx, outputDir, &setting, logger); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Deployment failed: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Site deployed successfully to %s (domain: %s)", setting.Platform, setting.Domain())), nil
	}
}
