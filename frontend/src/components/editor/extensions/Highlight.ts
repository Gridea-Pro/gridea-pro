/**
 * 高亮的 Markdown 方言（==text==），对齐 Gridea 的 markdown-it-mark。
 * 官方 Highlight 默认序列化为 <mark> HTML；此处覆盖为 == 语法并提供 tokenizer。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import Highlight from '@tiptap/extension-highlight'

export const CustomHighlight = Highlight.extend({
  markdownTokenName: 'mark',
  renderMarkdown: (node: any, helpers: any) => `==${helpers.renderChildren(node)}==`,
  parseMarkdown: (token: any, helpers: any) => {
    const content = helpers.parseInline(token.tokens || [])
    if (!content.length && token.text) content.push(helpers.createTextNode(token.text))
    return helpers.applyMark('highlight', content)
  },
  markdownTokenizer: {
    name: 'mark',
    level: 'inline',
    start: (src: string) => src.indexOf('=='),
    tokenize(src: string, _tokens: any, helpers: any) {
      // flanking 对齐 markdown-it-mark：开界 == 后非空白、闭界 == 前非空白（内部可含空格）
      const match = /^==(?!=)(?=\S)([\s\S]*?\S)==(?!=)/.exec(src)
      if (!match) return
      return { type: 'mark', raw: match[0], text: match[1], tokens: helpers.inlineTokens(match[1]) }
    },
  },
}).configure({ multicolor: true })
