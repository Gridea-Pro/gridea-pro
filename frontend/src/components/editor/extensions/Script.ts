/**
 * 上标 / 下标的 Markdown 方言（^x^ / ~x~），对齐 Gridea 的 markdown-it-sup / -sub。
 * 钩子移植自 @ctzhian/tiptap（node/Script.tsx），@tiptap/markdown 3.24 原生 API。
 * 关键：单 ~ 的 tokenizer 通过负向先行 (?!~) 避开 GFM 删除线 ~~。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import Subscript from '@tiptap/extension-subscript'
import Superscript from '@tiptap/extension-superscript'

export const CustomSubscript = Subscript.extend({
  markdownTokenName: 'sub',
  renderMarkdown: (node: any, helpers: any) => `~${helpers.renderChildren(node)}~`,
  parseMarkdown: (token: any, helpers: any) => {
    const content = helpers.parseInline(token.tokens || [])
    if (!content.length && token.text) content.push(helpers.createTextNode(token.text))
    return helpers.applyMark('subscript', content)
  },
  markdownTokenizer: {
    name: 'sub',
    level: 'inline',
    start: (src: string) => src.indexOf('~'),
    tokenize(src: string, _tokens: any, helpers: any) {
      const match = /^~(?!~)([^~\s\n]+?)~(?!~)/.exec(src) // 内容禁含空白，对齐 markdown-it-sub
      if (!match) return
      return { type: 'sub', raw: match[0], text: match[1], tokens: helpers.inlineTokens(match[1]) }
    },
  },
})

export const CustomSuperscript = Superscript.extend({
  markdownTokenName: 'sup',
  renderMarkdown: (node: any, helpers: any) => `^${helpers.renderChildren(node)}^`,
  parseMarkdown: (token: any, helpers: any) => {
    const content = helpers.parseInline(token.tokens || [])
    if (!content.length && token.text) content.push(helpers.createTextNode(token.text))
    return helpers.applyMark('superscript', content)
  },
  markdownTokenizer: {
    name: 'sup',
    level: 'inline',
    start: (src: string) => src.indexOf('^'),
    tokenize(src: string, _tokens: any, helpers: any) {
      const match = /^\^(?!\^)([^^\s\n]+?)\^(?!\^)/.exec(src) // 内容禁含空白，对齐 markdown-it-sup
      if (!match) return
      return { type: 'sup', raw: match[0], text: match[1], tokens: helpers.inlineTokens(match[1]) }
    },
  },
})
