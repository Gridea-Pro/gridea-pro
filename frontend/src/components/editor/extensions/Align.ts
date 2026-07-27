/**
 * 段落/标题对齐的 Markdown 往返。
 *
 * markdown 本身无对齐语法；非左对齐的块序列化为
 *   <div style="text-align: center">\n\n…内容(markdown)…\n\n</div>
 * 空行分隔的块级 HTML 在 goldmark(WithUnsafe) 下内部 markdown 照常渲染、对齐生效。
 * 解析方向与 Details 同机制：marked 把三段拆成 rawHtml(开)+正文块+rawHtml(闭)，
 * 由 markdown/foldDetails.ts 的 foldAlignContent 在编辑器 I/O 边界折叠回 textAlign 属性。
 * 无头测试台直通 manager（不折叠）时保持 raw-html 三段式，无损且幂等。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import Paragraph from '@tiptap/extension-paragraph'
import Heading from '@tiptap/extension-heading'

function wrapAlign(md: string, align: string | null | undefined): string {
  if (!align || align === 'left') return md
  return `<div style="text-align: ${align}">\n\n${md}\n\n</div>`
}

export const AlignedParagraph = Paragraph.extend({
  renderMarkdown(node: any, helpers: any) {
    // 必须显式传 content 数组：空段落（如插入图片时自动产生）没有 content，
    // 直接传 node 会被 manager 重新丢回渲染管线 → 无限递归栈溢出
    return wrapAlign(helpers.renderChildren(node.content || []), node.attrs?.textAlign)
  },
} as any)

export const AlignedHeading = Heading.extend({
  renderMarkdown(node: any, helpers: any) {
    const level = Math.min(6, Math.max(1, Number(node.attrs?.level) || 1))
    const md = `${'#'.repeat(level)} ${helpers.renderChildren(node.content || [])}`
    return wrapAlign(md, node.attrs?.textAlign)
  },
} as any)
