package utils

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var (
	// 定义两个全局实例，复用以提升性能
	mdSafe   goldmark.Markdown
	mdUnsafe goldmark.Markdown
)

// dangerousURLScheme 判断链接目标是否为可执行/危险协议。
// 前导空白与控制字符先剥掉（浏览器解析 URL 时会忽略它们，"java\tscript:" 一样可执行），
// 再小写比对，堵住大小写/夹空白绕过。与前端 editor/extensions/index.ts 的 isSafeHref 口径一致。
var controlAndSpace = regexp.MustCompile(`[\x00-\x20]+`)

func isDangerousURL(dest string) bool {
	s := strings.ToLower(controlAndSpace.ReplaceAllString(dest, ""))
	switch {
	case strings.HasPrefix(s, "javascript:"),
		strings.HasPrefix(s, "vbscript:"),
		strings.HasPrefix(s, "file:"):
		return true
	case strings.HasPrefix(s, "data:"):
		// data:image/* 用于内联图片是常见且安全的；其余 data:（尤其 data:text/html）拒绝。
		return !strings.HasPrefix(s, "data:image/")
	}
	return false
}

// linkProtocolGuard 是 goldmark AST Transformer：渲染前把危险协议的链接目标清空，
// 使 [文字](javascript:...) 发布后不再产出可点击的 href。纵深防御——前端 SafeLink 已在
// Markdown 落盘前拦截，这里再兜一道，覆盖历史脏数据与非编辑器写入的 .md。
type linkProtocolGuard struct{}

func (linkProtocolGuard) Transform(node *ast.Document, reader text.Reader, _ parser.Context) {
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if link, ok := n.(*ast.Link); ok {
			if isDangerousURL(string(link.Destination)) {
				link.Destination = []byte("")
			}
		}
		if al, ok := n.(*ast.AutoLink); ok {
			if isDangerousURL(string(al.URL(reader.Source()))) {
				// AutoLink 的 URL 由源文本派生，无法就地改写为安全值，直接降级为纯文本节点。
				textNode := ast.NewString(al.Label(reader.Source()))
				n.Parent().ReplaceChild(n.Parent(), n, textNode)
			}
		}
		return ast.WalkContinue, nil
	})
}

func linkGuardOption() goldmark.Option {
	return goldmark.WithParserOptions(
		parser.WithASTTransformers(util.Prioritized(linkProtocolGuard{}, 999)),
	)
}

func init() {
	mdSafe = goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Typographer, extension.Footnote, KatexExtension()),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		linkGuardOption(),
		goldmark.WithRendererOptions(html.WithHardWraps(), html.WithXHTML()),
	)

	mdUnsafe = goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Typographer, extension.Footnote, KatexExtension()),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		linkGuardOption(),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
}

// ToHTML 将 Markdown 文本转换为 HTML
func ToHTML(markdown string) string {
	return convert(mdSafe, markdown)
}

// ToHTMLUnsafe 将 Markdown 文本转换为 HTML（允许原始 HTML）
// 警告: 此函数允许 Markdown 中的原始 HTML，可能存在 XSS 风险
func ToHTMLUnsafe(markdown string) string {
	return convert(mdUnsafe, markdown)
}

// 统一的内部转换逻辑
func convert(engine goldmark.Markdown, markdown string) string {
	if markdown == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := engine.Convert([]byte(markdown), &buf); err != nil {
		// fallback: simple wrapper
		return "<p>" + markdown + "</p>"
	}
	return buf.String()
}
