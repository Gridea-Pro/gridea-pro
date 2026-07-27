/**
 * 脚注，对齐 Gridea 的 markdown-it-footnote：
 *   - 引用 `[^id]`（行内 sup）
 *   - 定义 `[^id]: 文本`（块级，文末）
 * 基础版：定义文本以纯文本属性保存，保证往返。富文本定义/回跳交互可后续增强。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { Node } from '@tiptap/core'

export const FootnoteRef = Node.create({
  name: 'footnoteRef',
  group: 'inline',
  inline: true,
  atom: true,

  addAttributes() {
    return { id: { default: '' } }
  },
  parseHTML() {
    return [{ tag: 'sup[data-footnote-ref]' }]
  },
  renderHTML({ node }: any) {
    return ['sup', { 'data-footnote-ref': '', class: 'footnote-ref' }, `[${node.attrs.id}]`]
  },

  markdownTokenName: 'footnoteRef',
  markdownTokenizer: {
    name: 'footnoteRef',
    level: 'inline',
    start: (src: string) => {
      const i = src.indexOf('[^')
      return i < 0 ? undefined : i
    },
    tokenize(src: string) {
      const m = /^\[\^([^\]\s]+)\](?!:)/.exec(src)
      if (!m) return
      return { type: 'footnoteRef', raw: m[0], id: m[1] }
    },
  },
  parseMarkdown: (token: any, helpers: any) => helpers.createNode('footnoteRef', { id: token.id }),
  renderMarkdown: (node: any) => `[^${node.attrs?.id ?? ''}]`,
} as any)

export const FootnoteDef = Node.create({
  name: 'footnoteDef',
  group: 'block',
  atom: true,

  addAttributes() {
    return { id: { default: '' }, text: { default: '' } }
  },
  parseHTML() {
    return [{ tag: 'div[data-footnote-def]' }]
  },
  renderHTML({ node }: any) {
    return ['div', { 'data-footnote-def': '', class: 'footnote-def' }, `[^${node.attrs.id}]: ${node.attrs.text}`]
  },

  markdownTokenName: 'footnoteDef',
  markdownTokenizer: {
    name: 'footnoteDef',
    level: 'block',
    start: (src: string) => {
      const m = src.match(/(^|\n)\[\^[^\]\s]+\]:/)
      return m ? m.index : undefined
    },
    tokenize(src: string) {
      const m = /^\[\^([^\]\s]+)\]:[ \t]*(.*)/.exec(src)
      if (!m) return
      return { type: 'footnoteDef', raw: m[0], id: m[1], text: m[2] }
    },
  },
  parseMarkdown: (token: any, helpers: any) =>
    helpers.createNode('footnoteDef', { id: token.id, text: token.text }),
  renderMarkdown: (node: any) => `[^${node.attrs?.id ?? ''}]: ${node.attrs?.text ?? ''}`,
} as any)
