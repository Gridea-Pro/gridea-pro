/**
 * Markdown 往返入口。基于官方 @tiptap/markdown：
 * - editor.getMarkdown() 序列化
 * - editor.commands.setContent(md, { contentType: 'markdown' }) 反序列化
 */
import type { Editor } from '@tiptap/core'
import { canonicalizeMarkdown } from './canonicalize'
import { foldDetailsContent, foldAlignContent } from './foldDetails'

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
    // 折叠（details / 对齐包裹）需要先不触发 update，折叠后的二次回灌再按需 emit，避免两次 onUpdate
    const needsFold = !!md && (md.includes('<details') || md.includes('text-align:'))
    editor.commands.setContent(md ?? '', { contentType: 'markdown', emitUpdate: needsFold ? false : emitUpdate })
    if (!needsFold) return
    // 把 marked 拆出的 rawHtml(开)+正文+rawHtml(闭) 三段式折叠回富文本节点/属性
    // （details 体内的对齐包裹不再二次折叠，保持 raw 形态，仍无损可发布）
    const json = editor.getJSON()
    const d = foldDetailsContent((json.content as never[]) || [])
    const a = foldAlignContent(d.content)
    const changed = d.changed || a.changed
    if (changed) {
      editor.commands.setContent({ ...json, content: a.content }, { emitUpdate })
    } else if (emitUpdate) {
      // 无折叠但调用方要求 emit：补一次（不改内容）
      editor.commands.setContent(json, { emitUpdate })
    }
  } catch (e) {
    console.error('[editor] setContent(markdown) failed:', e)
  }
}
