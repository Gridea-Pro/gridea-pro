// CJK 感知的双引号处理。
//
// goldmark 的 Typographer 判定闭引号要求 `CanClose && !CanOpen`——即引号右侧
// 必须是空白或标点。这条规则是为西文写的，中文里 `力"这` 两侧都是汉字，
// 两个标志同时为真，判定直接落空：结果是开引号被替换成弯引号、闭引号仍是直引号，
// 发布出去的中文引号一半弯一半直、根本不成对。
//
// 这里接管双引号：中文引号必然成对出现，用「开-闭交替」比 flanking 规则可靠，
// 且对英文同样成立（He said "hello" → He said “hello”）。
// 单引号仍留给 Typographer——英文撇号（it's）与闭单引号无法靠交替区分。
package utils

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var (
	leftDoubleQuote  = []byte("“") // “
	rightDoubleQuote = []byte("”") // ”
)

// quoteStateKey 记录当前块内双引号是否处于「已开启」状态。
var quoteStateKey = parser.NewContextKey()

type quoteState struct{ open bool }

func getQuoteState(pc parser.Context) *quoteState {
	if v := pc.Get(quoteStateKey); v != nil {
		if s, ok := v.(*quoteState); ok {
			return s
		}
	}
	s := &quoteState{}
	pc.Set(quoteStateKey, s)
	return s
}

type quoteParser struct{}

func (p *quoteParser) Trigger() []byte { return []byte{'"'} }

func (p *quoteParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) == 0 || line[0] != '"' {
		return nil
	}
	// 连续两个引号（如 21""）语义不明，原样放行，交给后续解析
	if len(line) > 1 && line[1] == '"' {
		return nil
	}

	s := getQuoteState(pc)
	repl := leftDoubleQuote
	if s.open {
		repl = rightDoubleQuote
	}
	s.open = !s.open

	node := ast.NewString(repl)
	node.SetCode(true) // 标记为已成形的输出，阻止后续 parser 再加工
	block.Advance(1)
	return node
}

// CloseBlock 重置状态：引号不跨块配对，否则一个未闭合的引号会污染后面所有段落。
func (p *quoteParser) CloseBlock(parent ast.Node, pc parser.Context) {
	getQuoteState(pc).open = false
}

type quoteExtender struct{}

// SmartQuoteExtension 返回 CJK 感知的双引号扩展。
// 必须与 extension.Typographer 一起挂载，且优先级高于它（9999）才能先拿到 `"`。
func SmartQuoteExtension() goldmark.Extender { return &quoteExtender{} }

func (e *quoteExtender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(&quoteParser{}, 500),
	))
}
