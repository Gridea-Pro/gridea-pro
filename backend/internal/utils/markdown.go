package utils

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var (
	// md 是预配置的 goldmark 实例
	md goldmark.Markdown
)

func init() {
	// 配置 goldmark，启用常用扩展
	md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,         // GitHub Flavored Markdown (表格、删除线、任务列表等)
			extension.Typographer, // 智能标点符号
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // 自动生成标题 ID
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(), // 硬换行
			html.WithXHTML(),     // 输出 XHTML 兼容的 HTML
		),
	)
}

// ToHTML 将 Markdown 文本转换为 HTML
// 参数:
//
//	markdown: 要转换的 Markdown 文本
//
// 返回:
//
//	转换后的 HTML 字符串
func ToHTML(markdown string) string {
	if markdown == "" {
		return ""
	}

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		// 如果转换失败，返回原始文本（带 <p> 标签包裹）
		return "<p>" + markdown + "</p>"
	}

	return buf.String()
}

// ToHTMLUnsafe 将 Markdown 文本转换为 HTML（允许原始 HTML）
// 警告: 此函数允许 Markdown 中的原始 HTML，可能存在 XSS 风险
// 仅在完全信任输入内容时使用
func ToHTMLUnsafe(markdown string) string {
	if markdown == "" {
		return ""
	}

	// 创建允许原始 HTML 的实例
	unsafeMd := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Typographer,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(), // 允许原始 HTML
		),
	)

	var buf bytes.Buffer
	if err := unsafeMd.Convert([]byte(markdown), &buf); err != nil {
		return "<p>" + markdown + "</p>"
	}

	return buf.String()
}
