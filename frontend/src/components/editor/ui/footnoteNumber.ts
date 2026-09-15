/**
 * 脚注显示编号：按文档顺序对去重后的 footnoteRef id 编号（1-based）。
 * 纯展示，Markdown 仍保存原始 id 字符串（往返不受影响）。仅被 .vue NodeView 引用。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import type { Editor } from '@tiptap/core'

export function footnoteOrder(editor: Editor | undefined | null): string[] {
  const order: string[] = []
  if (!editor) return order
  editor.state.doc.descendants((n: any) => {
    if (n.type.name === 'footnoteRef') {
      const fid = n.attrs.id as string
      if (fid && !order.includes(fid)) order.push(fid)
    }
  })
  return order
}

export function footnoteNumber(editor: Editor | undefined | null, id: string): number {
  const idx = footnoteOrder(editor).indexOf(id)
  return idx < 0 ? 0 : idx + 1
}

/** 查找某 id 的定义文本（用于引用悬停预览） */
export function footnoteDefText(editor: Editor | undefined | null, id: string): string {
  if (!editor) return ''
  let text = ''
  editor.state.doc.descendants((n: any) => {
    if (n.type.name === 'footnoteDef' && n.attrs.id === id) text = n.attrs.text || ''
  })
  return text
}
