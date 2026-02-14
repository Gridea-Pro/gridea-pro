package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerPrompts() {
	// 博客写作助手
	s.mcpServer.AddPrompt(
		mcp.NewPrompt(
			"blog_writing_assistant",
			mcp.WithPromptDescription("Helps write a new blog post with Gridea Pro style"),
			mcp.WithArgument("topic", mcp.ArgumentDescription("The topic of the blog post"), mcp.RequiredArgument()),
			mcp.WithArgument("tone", mcp.ArgumentDescription("The tone of the post (e.g. professional, casual). Default: casual")),
		),
		func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			topic := req.Params.Arguments["topic"]
			tone := req.Params.Arguments["tone"]
			if tone == "" {
				tone = "casual"
			}

			sysMsg := "You are an expert blogger using Gridea Pro. " +
				"Your task is to write a high-quality blog post about the given topic. " +
				"Please format the output as valid Markdown with Front Matter compatible with Gridea Pro.\n\n" +
				"Steps:\n" +
				"1. Call list_tags and list_categories to understand the current taxonomy\n" +
				"2. Draft an outline based on the topic\n" +
				"3. Write the full article with proper Front Matter\n" +
				"4. Use create_post to publish the article\n" +
				"5. Ask if the user wants to render_site\n\n" +
				"Front Matter Example:\n" +
				"---\n" +
				"title: [Title]\n" +
				"date: [YYYY-MM-DD HH:mm:ss]\n" +
				"tags: [Tag1, Tag2]\n" +
				"published: true\n" +
				"hideInList: false\n" +
				"isTop: false\n" +
				"---\n\n" +
				"Ensure the content is engaging and structured well."

			userMsg := fmt.Sprintf("Topic: %s\nTone: %s\n\nPlease write the blog post now.", topic, tone)

			return mcp.NewGetPromptResult(
				"Blog Writing Assistant",
				[]mcp.PromptMessage{
					mcp.NewPromptMessage(mcp.RoleAssistant, mcp.NewTextContent(sysMsg)),
					mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(userMsg)),
				},
			), nil
		},
	)

	// 闪念整理器：将闪念组织为博客文章
	s.mcpServer.AddPrompt(
		mcp.NewPrompt(
			"memo_to_post",
			mcp.WithPromptDescription("Organize multiple memos into a structured blog post"),
			mcp.WithArgument("topic", mcp.ArgumentDescription("Optional topic to focus on")),
		),
		func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			topic := req.Params.Arguments["topic"]

			sysMsg := "You are a content organizer for Gridea Pro. Your task is to:\n\n" +
				"1. Call list_memos to retrieve all memos\n" +
				"2. Analyze and group related memos by theme\n" +
				"3. Present the grouped memos to the user and suggest a blog post structure\n" +
				"4. After user approval, write a coherent blog post from the selected memos\n" +
				"5. Use create_post to publish the organized content\n" +
				"6. Optionally suggest cleaning up the source memos"

			userMsg := "Please organize my memos into a blog post."
			if topic != "" {
				userMsg = fmt.Sprintf("Please organize my memos related to '%s' into a blog post.", topic)
			}

			return mcp.NewGetPromptResult(
				"Memo to Post Organizer",
				[]mcp.PromptMessage{
					mcp.NewPromptMessage(mcp.RoleAssistant, mcp.NewTextContent(sysMsg)),
					mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(userMsg)),
				},
			), nil
		},
	)

	// 内容审查：检查 SEO、标签完整性等
	s.mcpServer.AddPrompt(
		mcp.NewPrompt(
			"content_review",
			mcp.WithPromptDescription("Review all posts for SEO, tag completeness, and content quality"),
		),
		func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			sysMsg := "You are a content quality reviewer for Gridea Pro. Your task is to:\n\n" +
				"1. Call list_posts to get all posts\n" +
				"2. For each post, check:\n" +
				"   - Has a descriptive title (not too short or generic)\n" +
				"   - Has at least one tag assigned\n" +
				"   - Has a proper publish date\n" +
				"   - Content length is reasonable (flag very short posts)\n" +
				"3. Call list_tags and list_categories to find:\n" +
				"   - Unused tags (no posts reference them)\n" +
				"   - Posts without categories\n" +
				"4. Present a summary report with actionable suggestions\n" +
				"5. Offer to fix issues (e.g., add missing tags) with user confirmation"

			return mcp.NewGetPromptResult(
				"Content Review",
				[]mcp.PromptMessage{
					mcp.NewPromptMessage(mcp.RoleAssistant, mcp.NewTextContent(sysMsg)),
					mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent("Please review my blog content and report any issues.")),
				},
			), nil
		},
	)

	// 站点健康检查
	s.mcpServer.AddPrompt(
		mcp.NewPrompt(
			"site_health_check",
			mcp.WithPromptDescription("Diagnose site health: empty tags, uncategorized posts, configuration issues"),
		),
		func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			sysMsg := "You are a site diagnostics tool for Gridea Pro. Perform a comprehensive health check:\n\n" +
				"1. Call list_posts — check for unpublished drafts, missing dates, empty content\n" +
				"2. Call list_tags — identify tags with no associated posts\n" +
				"3. Call list_categories — identify empty categories\n" +
				"4. Call list_links — check for links with missing URLs or descriptions\n" +
				"5. Call list_menus — verify menu links are valid\n" +
				"6. Call get_theme_config — check for missing site name/description/author\n" +
				"7. Call get_site_settings — verify deployment settings are configured\n\n" +
				"Present findings as a structured report with severity levels:\n" +
				"🔴 Critical | 🟡 Warning | 🟢 OK\n\n" +
				"Offer to fix issues with user confirmation."

			return mcp.NewGetPromptResult(
				"Site Health Check",
				[]mcp.PromptMessage{
					mcp.NewPromptMessage(mcp.RoleAssistant, mcp.NewTextContent(sysMsg)),
					mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent("Please run a full health check on my site.")),
				},
			), nil
		},
	)

	// 文章翻译
	s.mcpServer.AddPrompt(
		mcp.NewPrompt(
			"translate_post",
			mcp.WithPromptDescription("Translate a post to another language and create a new post"),
			mcp.WithArgument("filename", mcp.ArgumentDescription("Filename of the post to translate"), mcp.RequiredArgument()),
			mcp.WithArgument("language", mcp.ArgumentDescription("Target language (e.g. English, Japanese, French)"), mcp.RequiredArgument()),
		),
		func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			filename := req.Params.Arguments["filename"]
			language := req.Params.Arguments["language"]

			sysMsg := "You are a professional translator for Gridea Pro. Your task is to:\n\n" +
				"1. Call get_post with the given filename to retrieve the original content\n" +
				"2. Translate the title, content, and tags to the target language\n" +
				"3. Preserve all Markdown formatting and structure\n" +
				"4. Present the translation for user review\n" +
				"5. After approval, use create_post to save the translated version\n" +
				"   - Append language suffix to filename (e.g. hello-world-en.md)\n" +
				"   - Keep the same tags (translated) and categories\n" +
				"6. Ask if the user wants to render_site"

			userMsg := fmt.Sprintf("Please translate the post '%s' to %s.", filename, language)

			return mcp.NewGetPromptResult(
				"Post Translator",
				[]mcp.PromptMessage{
					mcp.NewPromptMessage(mcp.RoleAssistant, mcp.NewTextContent(sysMsg)),
					mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(userMsg)),
				},
			), nil
		},
	)
}
