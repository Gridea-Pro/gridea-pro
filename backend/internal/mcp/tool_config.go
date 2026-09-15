package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"gridea-pro/backend/internal/service"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// --- Theme Tools ---

func listThemesTool() mcp.Tool {
	return mcp.NewTool("list_themes", mcp.WithDescription("List installed themes"))
}

func listThemesHandler(s *service.ThemeService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		themes, err := s.LoadThemes(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load themes: %v", err)), nil
		}
		return mcp.NewToolResultText(jsonify(themes)), nil
	}
}

func getThemeConfigTool() mcp.Tool {
	return mcp.NewTool("get_theme_config", mcp.WithDescription("Get current theme configuration"))
}

func getThemeConfigHandler(s *service.ThemeService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		config, err := s.LoadThemeConfig(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load theme config: %v", err)), nil
		}
		return mcp.NewToolResultText(jsonify(config)), nil
	}
}

func updateThemeConfigTool() mcp.Tool {
	return mcp.NewTool("update_theme_config",
		mcp.WithDescription("Update theme configuration. Provide specific fields to update. Requires confirmation."),
		mcp.WithString("siteName", mcp.Description("Site Name")),
		mcp.WithString("siteDescription", mcp.Description("Site Description")),
		mcp.WithString("siteAuthor", mcp.Description("Site Author")),
		mcp.WithString("siteEmail", mcp.Description("Site Email")),
		mcp.WithString("footerInfo", mcp.Description("Footer Info")),
		mcp.WithString("configJson", mcp.Description("Full or partial config JSON to merge (advanced)")),
		mcp.WithBoolean("confirm", mcp.Description("Set to true to confirm the update"), mcp.Required()),
	)
}

func updateThemeConfigHandler(s *service.ThemeService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		current, err := s.LoadThemeConfig(ctx)
		if err != nil {
			return mcp.NewToolResultError("Failed to load current config"), nil
		}

		confirm := request.GetBool("confirm", false)

		// 构建变更后的配置（不保存，只用于 diff 预览）
		updated := current // 值拷贝
		if v := request.GetString("siteName", ""); v != "" {
			updated.SiteName = v
		}
		if v := request.GetString("siteDescription", ""); v != "" {
			updated.SiteDescription = v
		}
		if v := request.GetString("siteAuthor", ""); v != "" {
			updated.SiteAuthor = v
		}
		if v := request.GetString("siteEmail", ""); v != "" {
			updated.SiteEmail = v
		}
		if v := request.GetString("footerInfo", ""); v != "" {
			updated.FooterInfo = v
		}

		// Handle JSON merge if provided
		if jsonStr := request.GetString("configJson", ""); jsonStr != "" {
			if err := json.Unmarshal([]byte(jsonStr), &updated); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid configJson: %v", err)), nil
			}
		}

		if !confirm {
			// 构建变更 diff 预览
			var changes []string
			if updated.SiteName != current.SiteName {
				changes = append(changes, fmt.Sprintf("  siteName: '%s' → '%s'", current.SiteName, updated.SiteName))
			}
			if updated.SiteDescription != current.SiteDescription {
				changes = append(changes, fmt.Sprintf("  siteDescription: '%s' → '%s'", current.SiteDescription, updated.SiteDescription))
			}
			if updated.SiteAuthor != current.SiteAuthor {
				changes = append(changes, fmt.Sprintf("  siteAuthor: '%s' → '%s'", current.SiteAuthor, updated.SiteAuthor))
			}
			if updated.SiteEmail != current.SiteEmail {
				changes = append(changes, fmt.Sprintf("  siteEmail: '%s' → '%s'", current.SiteEmail, updated.SiteEmail))
			}
			if updated.FooterInfo != current.FooterInfo {
				changes = append(changes, fmt.Sprintf("  footerInfo: '%s' → '%s'", current.FooterInfo, updated.FooterInfo))
			}

			if len(changes) == 0 {
				return mcp.NewToolResultText("No changes detected."), nil
			}

			diff := "⚠️ CONFIRMATION REQUIRED\nThe following changes will be applied:\n\n"
			for _, c := range changes {
				diff += c + "\n"
			}
			diff += "\nCall update_theme_config again with confirm=true to apply."
			return mcp.NewToolResultText(diff), nil
		}

		if err := s.SaveThemeConfig(ctx, updated); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save config: %v", err)), nil
		}

		return mcp.NewToolResultText("Theme config updated"), nil
	}
}

// --- Setting Tools ---

func getSettingsTool() mcp.Tool {
	return mcp.NewTool("get_site_settings", mcp.WithDescription("Get global site settings"))
}

func getSettingsHandler(s *service.SettingService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		setting, err := s.GetSetting(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
		}
		// Mask sensitive data in platformConfigs
		if setting.PlatformConfigs != nil {
			sensitiveKeys := []string{"token", "password", "privateKey", "netlifyAccessToken"}
			for _, m := range setting.PlatformConfigs {
				for _, key := range sensitiveKeys {
					if _, ok := m[key]; ok {
						m[key] = "***"
					}
				}
			}
		}

		return mcp.NewToolResultText(jsonify(setting)), nil
	}
}

func updateSettingsTool() mcp.Tool {
	return mcp.NewTool("update_site_settings",
		mcp.WithDescription("Update global site settings"),
		mcp.WithString("domain", mcp.Description("Site Domain")),
		mcp.WithString("repository", mcp.Description("Git Repository")),
		mcp.WithString("branch", mcp.Description("Git Branch")),
		mcp.WithString("username", mcp.Description("Git Username")),
		mcp.WithString("email", mcp.Description("Git Email")),
		// Sensitive fields not exposed via simple update to avoid accidental overwrite with ***
		// Use configJson for advanced updates if needed, or specific tools for sensitive data.
		mcp.WithString("configJson", mcp.Description("JSON to merge for advanced settings (careful with secrets)")),
	)
}

func updateSettingsHandler(s *service.SettingService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		current, err := s.GetSetting(ctx)
		if err != nil {
			return mcp.NewToolResultError("Failed to load settings"), nil
		}

		if v := request.GetString("domain", ""); v != "" {
			current.SetPlatformConfig(current.Platform, "domain", v)
		}
		if v := request.GetString("repository", ""); v != "" {
			current.SetPlatformConfig(current.Platform, "repository", v)
		}
		if v := request.GetString("branch", ""); v != "" {
			current.SetPlatformConfig(current.Platform, "branch", v)
		}
		if v := request.GetString("username", ""); v != "" {
			current.SetPlatformConfig(current.Platform, "username", v)
		}
		if v := request.GetString("email", ""); v != "" {
			current.SetPlatformConfig(current.Platform, "email", v)
		}

		if jsonStr := request.GetString("configJson", ""); jsonStr != "" {
			// Unmarshal into current to merge
			if err := json.Unmarshal([]byte(jsonStr), &current); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid configJson: %v", err)), nil
			}
		}

		// 安全：剥离 configJson 里可能夹带的敏感字段（token/password/privateKey），
		// 避免明文写入 setting.json 并绕过 Keychain。正常保存路径会把敏感字段转存 Keychain，
		// 但 MCP 进程不持有 Keychain 访问权，因此这里直接丢弃——与工具「不暴露敏感字段」的声明一致。
		_ = current.ExtractSensitiveFields()

		if err := s.SaveSetting(ctx, current); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save settings: %v", err)), nil
		}

		return mcp.NewToolResultText("Settings updated"), nil
	}
}
