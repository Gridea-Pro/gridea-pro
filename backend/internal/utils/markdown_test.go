package utils

import (
	"strings"
	"testing"
)

func TestToHTML_Basic(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     string
	}{
		{
			name:     "空字符串",
			markdown: "",
			want:     "",
		},
		{
			name:     "纯文本",
			markdown: "Hello, World!",
			want:     "<p>Hello, World!</p>\n",
		},
		{
			name:     "标题",
			markdown: "# Heading 1\n## Heading 2",
			want:     "<h1 id=\"heading-1\">Heading 1</h1>\n<h2 id=\"heading-2\">Heading 2</h2>\n",
		},
		{
			name:     "粗体和斜体",
			markdown: "**bold** and *italic*",
			want:     "<p><strong>bold</strong> and <em>italic</em></p>\n",
		},
		{
			name:     "链接",
			markdown: "[Google](https://google.com)",
			want:     "<p><a href=\"https://google.com\">Google</a></p>\n",
		},
		{
			name:     "图片",
			markdown: "![alt text](image.jpg)",
			want:     "<p><img src=\"image.jpg\" alt=\"alt text\" /></p>\n",
		},
		{
			name:     "无序列表",
			markdown: "- Item 1\n- Item 2",
			want:     "<ul>\n<li>Item 1</li>\n<li>Item 2</li>\n</ul>\n",
		},
		{
			name:     "有序列表",
			markdown: "1. First\n2. Second",
			want:     "<ol>\n<li>First</li>\n<li>Second</li>\n</ol>\n",
		},
		{
			name:     "代码块",
			markdown: "```go\nfunc main() {}\n```",
			want:     "<pre><code class=\"language-go\">func main() {}\n</code></pre>\n",
		},
		{
			name:     "行内代码",
			markdown: "Use `code` here",
			want:     "<p>Use <code>code</code> here</p>\n",
		},
		{
			name:     "引用",
			markdown: "> This is a quote",
			want:     "<blockquote>\n<p>This is a quote</p>\n</blockquote>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToHTML(tt.markdown)
			if got != tt.want {
				t.Errorf("ToHTML() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestToHTML_GFM(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		contains []string
	}{
		{
			name: "表格",
			markdown: `| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |`,
			contains: []string{"<table>", "<thead>", "<tbody>", "<th>", "<td>"},
		},
		{
			name:     "删除线",
			markdown: "~~deleted~~",
			contains: []string{"<del>deleted</del>"},
		},
		{
			name:     "任务列表",
			markdown: "- [x] Done\n- [ ] Todo",
			contains: []string{"<input", "checked", "disabled", "type=\"checkbox\""},
		},
		{
			name:     "自动链接",
			markdown: "https://example.com",
			contains: []string{"<a href=\"https://example.com\">https://example.com</a>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToHTML(tt.markdown)
			for _, substr := range tt.contains {
				if !strings.Contains(got, substr) {
					t.Errorf("ToHTML() missing expected substring:\ngot: %q\nwant to contain: %q", got, substr)
				}
			}
		})
	}
}

func TestToHTML_ChineseContent(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     string
	}{
		{
			name:     "中文文本",
			markdown: "你好，世界！",
			want:     "<p>你好，世界！</p>\n",
		},
		{
			name:     "中文标题",
			markdown: "# 中文标题",
			want:     "<h1 id=\"中文标题\">中文标题</h1>\n",
		},
		{
			name:     "混合内容",
			markdown: "这是**粗体**和*斜体*的中文文本",
			want:     "<p>这是<strong>粗体</strong>和<em>斜体</em>的中文文本</p>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToHTML(tt.markdown)
			if got != tt.want {
				t.Errorf("ToHTML() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestToHTMLUnsafe(t *testing.T) {
	markdown := "<script>alert('xss')</script>\n\n**bold**"
	result := ToHTMLUnsafe(markdown)

	// 应该包含原始 HTML
	if !strings.Contains(result, "<script>") {
		t.Error("ToHTMLUnsafe() should preserve raw HTML")
	}

	// 也应该处理 Markdown
	if !strings.Contains(result, "<strong>bold</strong>") {
		t.Error("ToHTMLUnsafe() should process Markdown")
	}
}

func TestToHTML_SafeByDefault(t *testing.T) {
	markdown := "<script>alert('xss')</script>\n\n**bold**"
	result := ToHTML(markdown)

	// 默认应该转义 HTML
	if strings.Contains(result, "<script>alert('xss')</script>") {
		t.Error("ToHTML() should escape raw HTML by default")
	}

	// 应该包含转义后的内容
	if !strings.Contains(result, "&lt;script&gt;") {
		t.Error("ToHTML() should contain escaped HTML")
	}

	// 仍然应该处理 Markdown
	if !strings.Contains(result, "<strong>bold</strong>") {
		t.Error("ToHTML() should process Markdown")
	}
}
