// Markdown 方言扩展：==高亮== / ^上标^ / ~下标~。
//
// 这三种语法编辑器侧早已支持（extensions/Highlight.ts、extensions/Script.ts，
// 对齐 Gridea 历史上的 markdown-it-mark / -sup / -sub），但后端 goldmark 没有对应扩展：
// 高亮发布出去变成字面 `==` 符号，下标还会被 GFM 删除线扩展吃掉渲染成 <del>
// （GFM 的判定是 OriginalLength > 2 才拒绝，长度 1 的 `~x~` 它照收）。
// 本文件补齐后端侧，判定规则与前端 tokenizer 逐条对齐，保证两边同进同出。
package utils

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var (
	kindMark = ast.NewNodeKind("Mark")
	kindSup  = ast.NewNodeKind("Sup")
	kindSub  = ast.NewNodeKind("Sub")
)

type dialectNode struct {
	ast.BaseInline
	kind ast.NodeKind
}

func (n *dialectNode) Kind() ast.NodeKind { return n.kind }
func (n *dialectNode) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, nil, nil)
}

// ---------- 解析 ----------

// dialectDelimiter 把「同字符成对包裹」的三种方言抽成一份参数化实现，
// 差异只有定界符、重复次数、内容是否允许空白。
type dialectDelimiter struct {
	char byte
	// 定界符的确切重复次数：== 为 2，^ / ~ 为 1。长度不符一律不接管，
	// 由此与 GFM 删除线（~~，长度 2）划清边界。
	length int
	// 内容是否允许含空白。markdown-it-sup/-sub 不允许，mark 允许。
	allowSpace bool
	kind       ast.NodeKind
}

func (d *dialectDelimiter) IsDelimiter(b byte) bool { return b == d.char }

func (d *dialectDelimiter) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (d *dialectDelimiter) OnMatch(consumes int) ast.Node {
	return &dialectNode{kind: d.kind}
}

type dialectParser struct{ d *dialectDelimiter }

func (p *dialectParser) Trigger() []byte { return []byte{p.d.char} }

func (p *dialectParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	before := block.PrecendingCharacter()
	// 紧邻同字符时不接管：位置 0 的 `===` 因长度不符被拒后，goldmark 会从下一个
	// 字符重试，那里恰好只剩两个 `=`，会把 `===x===` 误判成高亮。
	if before == rune(p.d.char) {
		return nil
	}
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, p.d.length, p.d)
	if node == nil {
		return nil
	}
	// 长度必须精确匹配：`===` 不是高亮，`~~x~~` 归 GFM 删除线
	if node.OriginalLength != p.d.length {
		return nil
	}
	// 内容禁含空白的方言（上下标）：先向前望一眼配对情况，
	// 避免把正文里孤立的 `^` / `~`（如 "25^C 到 30^C"）误判成标记。
	// 与前端 Script.ts 的正则同为单行判定。
	if !p.d.allowSpace && node.CanOpen {
		rest := line[node.OriginalLength:]
		idx := bytes.IndexByte(rest, p.d.char)
		if idx <= 0 || bytes.ContainsAny(rest[:idx], " \t") {
			return nil
		}
	}

	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

func (p *dialectParser) CloseBlock(parent ast.Node, pc parser.Context) {}

// ---------- 渲染 ----------

type dialectRenderer struct{ html.Config }

func (r *dialectRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindMark, tagRenderer("mark"))
	reg.Register(kindSup, tagRenderer("sup"))
	reg.Register(kindSub, tagRenderer("sub"))
}

func tagRenderer(tag string) renderer.NodeRendererFunc {
	open, close := "<"+tag+">", "</"+tag+">"
	return func(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			_, _ = w.WriteString(open)
		} else {
			_, _ = w.WriteString(close)
		}
		return ast.WalkContinue, nil
	}
}

// ---------- Extender ----------

type dialectExtender struct{}

// DialectExtension 返回 ==高亮== / ^上标^ / ~下标~ 的 goldmark 扩展。
func DialectExtension() goldmark.Extender { return &dialectExtender{} }

func (e *dialectExtender) Extend(m goldmark.Markdown) {
	mark := &dialectDelimiter{char: '=', length: 2, allowSpace: true, kind: kindMark}
	sup := &dialectDelimiter{char: '^', length: 1, kind: kindSup}
	sub := &dialectDelimiter{char: '~', length: 1, kind: kindSub}

	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(&dialectParser{d: mark}, 190),
		util.Prioritized(&dialectParser{d: sup}, 191),
		// 必须排在 GFM strikethrough（优先级 500）之前，否则单个 `~x~`
		// 会被删除线扩展先拿走，下标渲染成 <del>。
		util.Prioritized(&dialectParser{d: sub}, 192),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&dialectRenderer{Config: html.NewConfig()}, 190),
	))
}
