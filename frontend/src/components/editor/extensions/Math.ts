/**
 * 行内/块级数学公式（KaTeX）。复用官方 @tiptap/extension-mathematics 的节点与渲染，
 * 但**重写行内 tokenizer**，严格对齐 Gridea 预览所用的 @iktakahiro/markdown-it-katex 规则：
 *   - 开界 `$`：其后不能是空白；
 *   - 闭界 `$`：其前不能是空白、其后不能是数字。
 * 由此避免把货币 `$11 亿`、环境变量 `$PATH`、模板 `${x}` 误判为公式（曾导致表格被吞）。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { InlineMath, BlockMath } from '@tiptap/extension-mathematics'

const inlineMathTokenizer = {
  name: 'inlineMath',
  level: 'inline' as const,
  start: (src: string) => {
    const i = src.indexOf('$')
    return i < 0 ? undefined : i
  },
  tokenize(src: string) {
    if (src[0] !== '$' || src[1] === '$') return // 块级 $$ 交给官方 block tokenizer
    const next = src[1]
    if (next === undefined || next === ' ' || next === '\t') return // 开界后不能是空白
    for (let i = 1; i < src.length; i++) {
      const ch = src[i]
      if (ch === '\\') {
        i++
        continue
      }
      if (ch === '\n') return // 行内不跨行
      if (ch === '$') {
        const prev = src[i - 1]
        const after = src[i + 1]
        if (prev === ' ' || prev === '\t') continue // 闭界前不能是空白
        if (after !== undefined && after >= '0' && after <= '9') continue // 闭界后不能是数字
        const latex = src.slice(1, i)
        if (!latex.trim()) return
        return { type: 'inlineMath', raw: src.slice(0, i + 1), latex, text: latex }
      }
    }
    return
  },
}

/** 点击数学公式时回调（仅富文本环境注入；测试台不传 → 公式仅渲染，不可编辑） */
export interface MathEditPayload {
  kind: 'inline' | 'block'
  pos: number
  latex: string
}
export type OnMathEdit = (p: MathEditPayload) => void

const InlineMathNode = InlineMath.extend({
  markdownTokenName: 'inlineMath',
  markdownTokenizer: inlineMathTokenizer,
  parseMarkdown: (token: any, helpers: any) =>
    helpers.createNode('inlineMath', { latex: token.latex ?? token.text ?? '' }),
  renderMarkdown: (node: any) => `$${node.attrs?.latex ?? ''}$`,
} as any)

const BlockMathNode = BlockMath.extend({
  markdownTokenName: 'blockMath',
  renderMarkdown: (node: any) => `$$\n${node.attrs?.latex ?? ''}\n$$`,
} as any)

export function createInlineMath(onEdit?: OnMathEdit) {
  return InlineMathNode.configure({
    katexOptions: { throwOnError: false },
    ...(onEdit
      ? { onClick: (node: any, pos: number) => onEdit({ kind: 'inline', pos, latex: node.attrs?.latex ?? '' }) }
      : {}),
  } as any)
}

export function createBlockMath(onEdit?: OnMathEdit) {
  return BlockMathNode.configure({
    katexOptions: { throwOnError: false, displayMode: true },
    ...(onEdit
      ? { onClick: (node: any, pos: number) => onEdit({ kind: 'block', pos, latex: node.attrs?.latex ?? '' }) }
      : {}),
  } as any)
}
