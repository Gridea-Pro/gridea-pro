/**
 * `<!-- more -->` 摘要分隔节点。Gridea 用它作为「阅读更多」截断点（后端按此正则清理）。
 * markdown: 拦截内容为 more 的 HTML 注释 token；序列化回 `<!-- more -->`。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { Node } from '@tiptap/core'

const MORE_RE = /^<!--\s*more\s*-->$/

export const MoreBreak = Node.create({
  name: 'moreBreak',
  group: 'block',
  atom: true,
  selectable: true,

  parseHTML() {
    return [{ tag: 'div[data-more-break]' }]
  },
  renderHTML() {
    return ['div', { 'data-more-break': '', class: 'more-break', contenteditable: 'false' }]
  },

  markdownTokenName: 'html',
  parseMarkdown(token: any, helpers: any) {
    const raw = String(token.raw || token.text || '').trim()
    if (!MORE_RE.test(raw)) return null // 非 more 注释 → 交给 RawHtml
    return helpers.createNode('moreBreak')
  },
  renderMarkdown() {
    return '<!-- more -->'
  },
} as any)
