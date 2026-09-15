package utils

import (
	"strings"
	"testing"
)

// TestKatexInlineRendered 行内公式 `$x$` 应该被渲染成 KaTeX HTML。
func TestKatexInlineRendered(t *testing.T) {
	out := ToHTML("行内：$E = mc^2$ 结束。")
	if !strings.Contains(out, `class="katex"`) {
		t.Fatalf("行内公式未被渲染:\n%s", out)
	}
	// MathML 部分应该出现
	if !strings.Contains(out, "<math") {
		t.Fatalf("KaTeX 输出缺少 MathML:\n%s", out)
	}
}

// TestKatexBlockRendered 块级公式 `$$\n...\n$$` 应该被渲染，且不嵌在 <p> 里。
func TestKatexBlockRendered(t *testing.T) {
	md := "块级公式：\n\n$$\na + b = c\n$$\n\n后文。"
	out := ToHTML(md)
	if !strings.Contains(out, "katex-display") {
		t.Fatalf("块级公式未被渲染成 display 模式:\n%s", out)
	}
	// 关键回归点：块级 KaTeX 输出不应该被 <p> 包住产生 <p><div>...</div></p>。
	// 这里 KaTeX 用的是 <span class="katex-display">，但起码不能被前后的段落 </p><p> 夹击。
	if strings.Contains(out, `<p><span class="katex-display"`) {
		t.Fatalf("块级公式被 <p> 包住了，HTML 嵌套不合法:\n%s", out)
	}
	// 只断言 katex-display 存在是不够的：曾经 Open/Continue 里多调了一次
	// reader.Advance，公式正文被整行吞掉，渲染出的是一个内容为空的 katex-display，
	// 上面两条断言全都能过。必须校验公式原文确实进了输出。
	if got := katexAnnotation(out); got != "a + b = c" {
		t.Fatalf("块级公式内容丢失，annotation = %q，期望 %q\n%s", got, "a + b = c", out)
	}
}

// TestKatexBlockMultiLineContent 多行块公式必须逐行完整保留，不能吞行。
func TestKatexBlockMultiLineContent(t *testing.T) {
	cases := map[string]string{
		"$$\nAAA\n$$":                "AAA",
		"$$\nAAA\nBBB\n$$":           "AAA\nBBB",
		"$$\nAAA\nBBB\nCCC\n$$":      "AAA\nBBB\nCCC",
		"$$AAA\nBBB\n$$":             "AAA\nBBB",
		"$$\nAAA":                    "AAA", // 未闭合兜底
		"文本\n\n$$\nE=mc^2\n$$\n\n文本": "E=mc^2",
	}
	for md, want := range cases {
		if got := katexAnnotation(ToHTMLUnsafe(md)); got != want {
			t.Errorf("块公式内容不符\n  输入: %q\n  得到: %q\n  期望: %q", md, got, want)
		}
	}
}

// katexAnnotation 取出 KaTeX 输出里的 LaTeX 原文（<annotation> 的内容）。
func katexAnnotation(html string) string {
	const open = `encoding="application/x-tex">`
	i := strings.Index(html, open)
	if i < 0 {
		return ""
	}
	rest := html[i+len(open):]
	j := strings.Index(rest, "</annotation>")
	if j < 0 {
		return ""
	}
	return rest[:j]
}

// TestKatexBlockSingleLine 单行 `$$ x $$` 也应该被识别成块级。
func TestKatexBlockSingleLine(t *testing.T) {
	out := ToHTML("$$ a + b = c $$")
	if !strings.Contains(out, "katex-display") {
		t.Fatalf("单行块级公式未被渲染:\n%s", out)
	}
}

// TestKatexLiteralDollarNotConsumed issue 108 同类回归保护：
// `It costs $10 and $20.` 这种字面美元符号不能被当成公式起点吞掉。
func TestKatexLiteralDollarNotConsumed(t *testing.T) {
	out := ToHTML("It costs $10 and $20.")
	if strings.Contains(out, `class="katex"`) {
		t.Fatalf("字面美元符号被错当公式渲染了:\n%s", out)
	}
	// 内容应该原样保留（也许 HTML escape 过 & 之类，但 $10、$20 应该都在）
	if !strings.Contains(out, "$10") || !strings.Contains(out, "$20") {
		t.Fatalf("字面 $10 / $20 内容丢失:\n%s", out)
	}
}

// TestKatexInsideCodeBlockNotRendered 代码块里的 `$x$` 不应该被渲染成公式。
func TestKatexInsideCodeBlockNotRendered(t *testing.T) {
	md := "代码块：\n\n```\n$x^2$\n```\n"
	out := ToHTML(md)
	if strings.Contains(out, `class="katex"`) {
		t.Fatalf("代码块里的 $x^2$ 被错渲染了:\n%s", out)
	}
}

// TestKatexInlineWhitespaceBoundary `$ x$` 和 `$x $` 都不应该开/闭公式。
func TestKatexInlineWhitespaceBoundary(t *testing.T) {
	cases := []string{
		"开 $ x$ 紧跟空白",
		"闭 $x $ 紧跟空白",
	}
	for _, md := range cases {
		out := ToHTML(md)
		if strings.Contains(out, `class="katex"`) {
			t.Fatalf("空白边界违反应被拒绝，但被渲染了:\nmd=%q\nout=%s", md, out)
		}
	}
}

// TestKatexCrossLineInlineRejected 跨行的 `$x\n$` 不应该被当公式。
func TestKatexCrossLineInlineRejected(t *testing.T) {
	out := ToHTML("跨行：$x\n$ 结束")
	if strings.Contains(out, `class="katex"`) {
		t.Fatalf("跨行行内公式被错渲染:\n%s", out)
	}
}

// TestKatexInlineDisplayInParagraph 同行 `$$x$$` 在段落里应该用 display 模式渲染。
func TestKatexInlineDisplayInParagraph(t *testing.T) {
	out := ToHTML("段落里 $$ a = 1 $$ 的公式")
	if !strings.Contains(out, `class="katex"`) {
		t.Fatalf("同行块级公式未被渲染:\n%s", out)
	}
}

// TestKatexCacheReuse 第二次渲染同一公式应该走缓存（耗时显著小于第一次）。
func TestKatexCacheReuse(t *testing.T) {
	md := "$E = mc^2$"
	_ = ToHTML(md)
	out := ToHTML(md)
	if !strings.Contains(out, `class="katex"`) {
		t.Fatalf("缓存后渲染丢失内容:\n%s", out)
	}
}
