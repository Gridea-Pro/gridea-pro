package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerResources() {
	// Site Info Resource
	s.mcpServer.AddResource(
		mcp.NewResource(
			"gridea://site/info",
			"Site Information",
			mcp.WithResourceDescription("Basic information about the Gridea Pro site"),
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			// Helper to fetch data
			config, err := s.services.Theme.LoadThemeConfig(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to load theme config: %v", err)
			}

			settings, err := s.services.Setting.GetSetting(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to load settings: %v", err)
			}

			info := map[string]interface{}{
				"siteName":    config.SiteName,
				"description": config.SiteDescription,
				"author":      config.SiteAuthor,
				"domain":      settings.Domain,
				"theme":       config.ThemeName,
			}

			jsonBytes, _ := json.MarshalIndent(info, "", "  ")

			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "application/json",
					Text:     string(jsonBytes),
				},
			}, nil
		},
	)

	// Posts Summary Resource
	s.mcpServer.AddResource(
		mcp.NewResource(
			"gridea://posts/summary",
			"Posts Summary",
			mcp.WithResourceDescription("Summary list of all posts"),
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			posts, err := s.services.Post.LoadPosts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to load posts: %v", err)
			}

			var summary []map[string]interface{}
			for _, p := range posts {
				summary = append(summary, map[string]interface{}{
					"title":     p.Title,
					"date":      p.CreatedAt,
					"fileName":  p.FileName,
					"tags":      p.Tags,
					"published": p.Published,
				})
			}

			jsonBytes, _ := json.MarshalIndent(summary, "", "  ")

			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "application/json",
					Text:     string(jsonBytes),
				},
			}, nil
		},
	)

	// Recent Memos Resource
	s.mcpServer.AddResource(
		mcp.NewResource(
			"gridea://memos/recent",
			"Recent Memos",
			mcp.WithResourceDescription("Most recent memos (up to 20)"),
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			memos, err := s.services.Memo.LoadMemos(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to load memos: %v", err)
			}

			// 最多返回 20 条
			limit := 20
			if len(memos) < limit {
				limit = len(memos)
			}
			recent := memos[:limit]

			jsonBytes, _ := json.MarshalIndent(recent, "", "  ")

			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "application/json",
					Text:     string(jsonBytes),
				},
			}, nil
		},
	)
}
