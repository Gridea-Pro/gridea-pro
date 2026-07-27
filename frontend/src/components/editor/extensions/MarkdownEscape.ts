/**
 * 反斜杠转义保真。marked 把 `\$ \* \[ \| \\` 等解析为 type:'escape' 的 inline token，
 * 而 @tiptap/markdown 无对应 handler → 静默丢弃，转义字符整体消失（跨文档内容丢失）。
 * 这里注册一个 'escape' handler，把转义出的字面字符保留为文本节点。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { Extension } from '@tiptap/core'

export const MarkdownEscape = Extension.create({
  name: 'markdownEscape',
  markdownTokenName: 'escape',
  parseMarkdown: (token: any, helpers: any) => helpers.createTextNode(token.text ?? token.raw ?? ''),
} as any)
