/**
 * Emoji 短码（:smile:），对齐 Gridea 的 markdown-it-emoji。钩子移植自 @ctzhian/tiptap。
 * tokenizer 仅匹配「已知」短码（shortcodeToEmoji 命中），故 8:30、http:// 等不会误判。
 * 选择器 UI 见 ui/EmojiPicker.vue（与本扩展同用 @tiptap/extension-emoji 数据源）。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import Emoji, { gitHubEmojis, shortcodeToEmoji } from '@tiptap/extension-emoji'

const SHORTCODE = /^:([a-zA-Z0-9_+-]+):/

export const CustomEmoji = Emoji.extend({
  markdownTokenName: 'emoji',
  markdownTokenizer: {
    name: 'emoji',
    level: 'inline',
    start: (src: string) => {
      const i = src.indexOf(':')
      return i < 0 ? undefined : i
    },
    tokenize(src: string) {
      const m = SHORTCODE.exec(src)
      if (!m) return
      const item = shortcodeToEmoji(m[1], gitHubEmojis)
      if (!item) return // 非已知短码 → 不匹配
      return { type: 'emoji', raw: m[0], name: item.name, shortcode: m[1] }
    },
  },
  parseMarkdown: (token: any, helpers: any) => helpers.createNode('emoji', { name: token.name }),
  renderMarkdown: (node: any) => {
    const name = node.attrs?.name
    return name ? `:${name}:` : ''
  },
} as any).configure({ emojis: gitHubEmojis, enableEmoticons: false })
