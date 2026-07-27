/**
 * 高亮的 Markdown 方言（==text==），对齐 Gridea 的 markdown-it-mark。
 * 官方 Highlight 默认序列化为 <mark> HTML；此处覆盖为 == 语法并提供 tokenizer。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import Highlight from '@tiptap/extension-highlight'
import { normalizeCssColor } from './RichTextStyle'

export const CustomHighlight = Highlight.extend({
  // 覆盖官方 color 属性：官方 parseHTML 只认 data-color/backgroundColor，
  // 这里补 backgroundImage（渐变荧光笔）。注意 mark 自身属性优先于全局属性，
  // 故必须在此覆盖而不能用 addGlobalAttributes。
  addAttributes() {
    return {
      ...this.parent?.(),
      color: {
        default: null,
        parseHTML: (el: HTMLElement) => {
          const bg = el.style.backgroundImage
          if (bg && bg.includes('gradient')) return bg
          return el.getAttribute('data-color') || el.style.backgroundColor || null
        },
        renderHTML: (attrs: Record<string, any>) => {
          if (!attrs.color) return {}
          if (String(attrs.color).includes('gradient')) {
            return { style: `background-image: ${attrs.color};` }
          }
          return { 'data-color': attrs.color, style: `background-color: ${attrs.color}` }
        },
      },
    }
  },

  markdownTokenName: 'mark',
  // 无色高亮保持 == 方言；带颜色（含渐变）序列化为内联 <mark style>（goldmark unsafe 可发布，
  // 解析方向经 parseHTML + GradientHighlight 全局属性无损还原）
  renderMarkdown: (node: any, helpers: any) => {
    const inner = helpers.renderChildren(node)
    const color = node.attrs?.color
    if (!color) return `==${inner}==`
    if (typeof color === 'string' && color.includes('gradient')) {
      return `<mark style="background-image: ${color};">${inner}</mark>`
    }
    return `<mark style="background-color: ${normalizeCssColor(color)}">${inner}</mark>`
  },
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
