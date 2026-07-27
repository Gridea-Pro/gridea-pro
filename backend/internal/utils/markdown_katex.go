package utils

// 自家 KaTeX goldmark 扩展。
//
// 为什么不直接用 github.com/FurqanSoftware/goldmark-katex：
//
//  1. 字面美元符号识别太松：`It costs $10 and $20.` 会被错当公式起点（issue 108）。
//     这里实现的行内 parser 加了 Pandoc / markdown-it-katex 同款严格规则：
//       - `$` 开后不能紧跟空白
//       - `$` 闭前不能紧跟空白
//       - `$` 开前不能是数字（防止 `$10`）
//       - `$` 闭后不能是数字（防止 `and $20.`）
//
//  2. 块级 `$$...$$` 走的是行内 parser，输出会被外层 `<p>` 包住产生 `<p><div>...</div></p>`
//     这种 HTML 不合法的嵌套。这里把它升级成真正的 goldmark 块级 parser，多行块级公式自然成
//     为顶级 block，不会被段落包住。
//
//  3. 每渲染一条公式都新建 QuickJS VM 并 eval 整个 katex.min.js（~400KB JS），单条 ~50ms。
//     这里改成全局单例 VM + sync.Mutex，初始化一次后每条公式只调 katex.renderToString，
//     耗时降到 ~1-3ms 级别。VM 出错（panic / Eval 失败）时会自动重建。
//
// 整体仍然遵循 goldmark 标准插件接口，不 fork 上游。

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"modernc.org/quickjs"

	"gridea-pro/backend/internal/utils/katexjs"
)

// ---------- AST 节点 ----------

var (
	kindInlineMath = ast.NewNodeKind("InlineMath")
	kindBlockMath  = ast.NewNodeKind("BlockMath")
)

// inlineMath 表示行内公式 `$...$` 或同行块级公式 `$$...$$`。
// 同行块级使用 displayMode=true 渲染（KaTeX 输出 <span class="katex-display">），
// 嵌在 <p> 里仍然是合法 HTML（span 是 inline 元素）。
type inlineMath struct {
	ast.BaseInline
	Equation []byte
	Display  bool
}

func (n *inlineMath) Kind() ast.NodeKind { return kindInlineMath }
func (n *inlineMath) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Equation": string(n.Equation),
	}, nil)
}

// blockMath 表示多行块级公式：`$$` 单独成行开 → 内容 → `$$` 收尾。
// 作为真正的 goldmark 块级节点，不会嵌进 <p>。
type blockMath struct {
	ast.BaseBlock
	Equation []byte

	// accumulating 在 parser 阶段缓存正在累积的多行内容；Close 时合并到 Equation。
	accumulating *bytes.Buffer
}

func (n *blockMath) Kind() ast.NodeKind { return kindBlockMath }
func (n *blockMath) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Equation": string(n.Equation),
	}, nil)
}

// ---------- 块级 parser ----------

type katexBlockParser struct{}

func (p *katexBlockParser) Trigger() []byte { return []byte{'$'} }

func (p *katexBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	pos := pc.BlockOffset()
	if pos < 0 || pos+1 >= len(line) {
		return nil, parser.NoChildren
	}
	if line[pos] != '$' || line[pos+1] != '$' {
		return nil, parser.NoChildren
	}

	rest := line[pos+2:]
	// 去掉行尾换行（如果有）
	rest = bytes.TrimRight(rest, "\r\n")

	node := &blockMath{}

	// 同一行就有闭合 `$$`：单行块级
	if before, _, found := bytes.Cut(rest, []byte("$$")); found {
		node.Equation = bytes.TrimSpace(before)
		// 整行消化掉
		reader.Advance(segment.Len())
		return node, parser.NoChildren | parser.Close
	}

	// 多行：把第一行剩余内容（如果有）作为公式开头
	node.accumulating = &bytes.Buffer{}
	if len(rest) > 0 {
		node.accumulating.Write(rest)
		node.accumulating.WriteByte('\n')
	}
	reader.Advance(segment.Len())
	return node, parser.NoChildren
}

func (p *katexBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	bm := node.(*blockMath)

	// 找闭合 `$$`。允许 `equation $$` 或单独一行 `$$`。
	if before, _, found := bytes.Cut(line, []byte("$$")); found {
		if len(before) > 0 {
			bm.accumulating.Write(before)
		}
		bm.Equation = bytes.TrimSpace(bm.accumulating.Bytes())
		bm.accumulating = nil
		reader.Advance(segment.Len())
		return parser.Close
	}

	bm.accumulating.Write(line)
	reader.Advance(segment.Len())
	return parser.Continue | parser.NoChildren
}

func (p *katexBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	// 文档结束但未见闭合 `$$`：兜底把累积内容当公式输出。
	bm := node.(*blockMath)
	if bm.accumulating != nil {
		bm.Equation = bytes.TrimSpace(bm.accumulating.Bytes())
		bm.accumulating = nil
	}
}

func (p *katexBlockParser) CanInterruptParagraph() bool { return true }
func (p *katexBlockParser) CanAcceptIndentedLine() bool { return false }

// ---------- 行内 parser ----------

type katexInlineParser struct{}

func (p *katexInlineParser) Trigger() []byte { return []byte{'$'} }

func (p *katexInlineParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) == 0 || line[0] != '$' {
		return nil
	}
	src := block.Source()
	_, pos := block.Position()
	startAbs := pos.Start

	// `$$...$$`（同行）：display 模式
	if len(line) >= 2 && line[1] == '$' {
		// 找同行的闭合 `$$`
		body := line[2:]
		idx := bytes.Index(body, []byte("$$"))
		if idx < 0 {
			return nil
		}
		// 跨行不接（碰到换行就停）
		if nl := bytes.IndexAny(body[:idx], "\r\n"); nl >= 0 {
			return nil
		}
		eq := bytes.TrimSpace(body[:idx])
		if len(eq) == 0 {
			return nil
		}
		block.Advance(2 + idx + 2)
		return &inlineMath{Equation: append([]byte(nil), eq...), Display: true}
	}

	// `$...$`：严格规则
	if len(line) < 3 {
		return nil
	}
	// 开 `$` 后不能紧跟空白
	if isSpaceLike(line[1]) {
		return nil
	}
	// 开 `$` 前不能是数字（防止 `$10`）
	if startAbs > 0 {
		prev := src[startAbs-1]
		if prev >= '0' && prev <= '9' {
			return nil
		}
	}

	// 扫描闭合 `$`：本行内、不能紧跟空白、不能紧接数字
	endIdx := -1
	for i := 1; i < len(line); i++ {
		c := line[i]
		if c == '\\' {
			i++
			continue
		}
		if c == '\r' || c == '\n' {
			break
		}
		if c == '$' {
			// 闭 `$` 前不能紧跟空白
			if isSpaceLike(line[i-1]) {
				continue
			}
			// 闭 `$` 后不能紧接数字（防止 `and $20.` 把 20 吞掉）
			if i+1 < len(line) {
				next := line[i+1]
				if next >= '0' && next <= '9' {
					return nil
				}
			}
			endIdx = i
			break
		}
	}
	if endIdx < 0 {
		return nil
	}

	eq := line[1:endIdx]
	if len(eq) == 0 {
		return nil
	}

	block.Advance(endIdx + 1)
	return &inlineMath{Equation: append([]byte(nil), eq...), Display: false}
}

func isSpaceLike(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// ---------- renderer ----------

type katexHTMLRenderer struct {
	cacheInline gcache.Cache
	cacheBlock  gcache.Cache
}

func newKatexHTMLRenderer() *katexHTMLRenderer {
	return &katexHTMLRenderer{
		cacheInline: gcache.New(5000).ARC().Build(),
		cacheBlock:  gcache.New(5000).ARC().Build(),
	}
}

func (r *katexHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindInlineMath, r.renderInline)
	reg.Register(kindBlockMath, r.renderBlock)
}

func (r *katexHTMLRenderer) renderInline(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	im := n.(*inlineMath)
	w.Write(r.lookup(im.Equation, im.Display))
	return ast.WalkContinue, nil
}

func (r *katexHTMLRenderer) renderBlock(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	bm := n.(*blockMath)
	w.Write(r.lookup(bm.Equation, true))
	return ast.WalkContinue, nil
}

func (r *katexHTMLRenderer) lookup(eq []byte, display bool) []byte {
	cache := r.cacheInline
	if display {
		cache = r.cacheBlock
	}
	if v, err := cache.Get(string(eq)); err == nil {
		return v.([]byte)
	}
	html := renderKatex(eq, display)
	cache.Set(string(eq), html)
	return html
}

// ---------- KaTeX 运行时（单例 QuickJS VM）----------

var (
	vmMu sync.Mutex
	vm   *quickjs.VM
)

// renderKatex 调 KaTeX 把公式渲染成 HTML。
// 内部用一个全局 QuickJS VM；多个 goroutine 调用时通过 vmMu 串行。
// VM 若出错（panic / Eval 失败）会被关闭，下一次调用自动重建。
func renderKatex(eq []byte, display bool) (out []byte) {
	vmMu.Lock()
	defer vmMu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			// VM 状态不可信：关掉，等下次调用再重建
			if vm != nil {
				vm.Close()
				vm = nil
			}
			out = fallbackHTML(eq, display)
		}
	}()

	if vm == nil {
		v, err := quickjs.NewVM()
		if err != nil {
			return fallbackHTML(eq, display)
		}
		// 防病态/超大公式（如超大 \begin{matrix}、深层嵌套 \frac）拖死渲染：
		// 限制单次 Eval 的执行时间与内存。超限时 Eval 返回错误，走下方 fallback 并重建 VM，
		// 避免一个坏公式永久挂住全局 vmMu、进而拖垮所有含公式文章的渲染。
		_ = v.SetEvalTimeout(5 * time.Second)
		v.SetMemoryLimit(256 << 20) // 256 MiB
		if _, err := v.Eval(katexjs.Script, quickjs.EvalGlobal); err != nil {
			v.Close()
			return fallbackHTML(eq, display)
		}
		vm = v
	}

	expr := fmt.Sprintf(`katex.renderToString(%q, {displayMode: %t, throwOnError: false})`,
		string(eq), display)
	result, err := vm.Eval(expr, quickjs.EvalGlobal)
	if err != nil {
		// 让 VM 重建：KaTeX 内部异常可能污染了 JS 全局态
		vm.Close()
		vm = nil
		return fallbackHTML(eq, display)
	}
	s, ok := result.(string)
	if !ok {
		return fallbackHTML(eq, display)
	}
	return []byte(s)
}

// fallbackHTML 在 KaTeX 渲染失败时，输出一段带原公式的占位 HTML，避免内容丢失。
func fallbackHTML(eq []byte, display bool) []byte {
	tag := "span"
	if display {
		tag = "div"
	}
	return fmt.Appendf(nil, `<%s class="katex-error" title="KaTeX render failed">%s</%s>`,
		tag, htmlEscape(eq), tag)
}

func htmlEscape(b []byte) string {
	var buf bytes.Buffer
	for _, c := range b {
		switch c {
		case '&':
			buf.WriteString("&amp;")
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '"':
			buf.WriteString("&quot;")
		default:
			buf.WriteByte(c)
		}
	}
	return buf.String()
}

// ---------- Extender ----------

// katexExtender 是给 goldmark.WithExtensions 用的入口。
type katexExtender struct{}

// KatexExtension 返回一个新的 KaTeX 扩展实例。挂到 goldmark.New 的 WithExtensions 里即可。
func KatexExtension() goldmark.Extender {
	return &katexExtender{}
}

func (e *katexExtender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(util.Prioritized(&katexBlockParser{}, 701)),
		parser.WithInlineParsers(util.Prioritized(&katexInlineParser{}, 150)),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(util.Prioritized(newKatexHTMLRenderer(), 100)),
	)
}
