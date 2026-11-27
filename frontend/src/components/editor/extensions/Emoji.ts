/**
 * Emoji 扩展：Markdown 侧以原生 Unicode 输出（如 😄），兼容旧文中的 :shortcode: 写法。
 *
 * 为什么不输出 :shortcode:：
 *   1. 发布引擎（Go 端 goldmark）未内置 emoji 短码扩展，:smile: 发布后会被渲染为纯文本；
 *   2. 原生 Unicode emoji 跨主题、跨编辑器通用，在源码模式和发布 HTML 中均无需额外转换。
 *
 * 兼容与保留策略：
 *   - markdownTokenizer 仅匹配已知短码（shortcodeToEmoji 命中），旧文章里的 :smile: 仍可被
 *     解析为 emoji 节点，在富文本中正常显示为 😄，并在保存/切回源码时转为原生 Unicode；
 *   - renderMarkdown 优先输出对应 Unicode（item.emoji），仅无 Unicode 的 GitHub 自定义图标
 *     （如 :trollface:）回退为 :name:；
 *   - 覆盖 addProseMirrorPlugins() 返回空数组，阻止 @tiptap/extension-emoji 内置的
 *     appendTransaction 把正文中的原生 Unicode emoji 劫持替换为 emoji 节点（否则会丢失变体
 *     选择符如 ❤️ 的 \uFE0F，或触发回退 CDN 图片）。
 *   - 保留继承的 addInputRules / addPasteRules，富文本中输入或粘贴 :smile: 仍可即时转成表情。
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
    if (!name) return ''
    const item = shortcodeToEmoji(name, gitHubEmojis)
    return item?.emoji || `:${name}:`
  },
  addProseMirrorPlugins() {
    return []
  },
} as any).configure({ emojis: gitHubEmojis, enableEmoticons: false })
