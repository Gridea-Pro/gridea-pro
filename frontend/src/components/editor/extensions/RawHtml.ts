/**
 * 原始 HTML 透传（块级）。Gridea 的 markdown-it 配置 html:true，正文允许内嵌 HTML；
 * 编辑器需原样保留这些 HTML 块（如 <div>、<details>、iframe 嵌入等），往返零漂移。
 * markdown: 兜底拦截 html token，保存 token.raw；序列化回原文。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { Node } from '@tiptap/core'

export const RawHtml = Node.create({
  name: 'rawHtml',
  group: 'block',
  atom: true,
  selectable: true,

  addAttributes() {
    return { html: { default: '' } }
  },

  parseHTML() {
    return [{ tag: 'div[data-raw-html]' }]
  },
  renderHTML({ node }: any) {
    return ['div', { 'data-raw-html': '', class: 'raw-html-block' }, String(node.attrs.html ?? '')]
  },

  markdownTokenName: 'html',
  parseMarkdown(token: any, helpers: any) {
    const raw = String(token.raw ?? token.text ?? '')
    if (!raw.trim()) return null
    return helpers.createNode('rawHtml', { html: raw.replace(/\n+$/, '') })
  },
  renderMarkdown(node: any) {
    return String(node.attrs?.html ?? '')
  },
} as any)
