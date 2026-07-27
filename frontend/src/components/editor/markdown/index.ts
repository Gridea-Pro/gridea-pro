/**
 * Markdown 往返入口。基于官方 @tiptap/markdown：
 * - editor.getMarkdown() 序列化
 * - editor.commands.setContent(md, { contentType: 'markdown' }) 反序列化
 */
import type { Editor } from '@tiptap/core'
import { canonicalizeMarkdown } from './canonicalize'

/** 导出 Markdown（规范化为幂等形式：表格最小内边距等，保证 once-saved-then-stable） */
export function getMarkdown(editor: Editor): string {
  try {
    return canonicalizeMarkdown(editor.getMarkdown())
  } catch (e) {
    console.error('[editor] getMarkdown failed:', e)
    return ''
  }
}

/**
 * 用 Markdown 设置内容。
 * @param emitUpdate 是否触发 onUpdate（外部回填时应为 false，避免回环）
 */
export function setMarkdown(editor: Editor, md: string, emitUpdate = false): void {
  try {
    editor.commands.setContent(md ?? '', { contentType: 'markdown', emitUpdate })
  } catch (e) {
    console.error('[editor] setContent(markdown) failed:', e)
  }
}
