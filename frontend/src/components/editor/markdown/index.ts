/**
 * Markdown 往返入口。基于官方 @tiptap/markdown：
 * - editor.getMarkdown() 序列化
 * - editor.commands.setContent(md, { contentType: 'markdown' }) 反序列化
 */
import type { Editor } from '@tiptap/core'
import { canonicalizeMarkdown } from './canonicalize'
import { foldDetailsContent } from './foldDetails'

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
    // details 折叠需要先不触发 update，折叠后的二次回灌再按需 emit，避免两次 onUpdate
    const hasDetails = !!md && md.includes('<details')
    editor.commands.setContent(md ?? '', { contentType: 'markdown', emitUpdate: hasDetails ? false : emitUpdate })
    if (!hasDetails) return
    // 把 marked 拆出的 rawHtml(开)+正文+rawHtml(闭) 折叠回富文本 details 节点
    const json = editor.getJSON()
    const { content, changed } = foldDetailsContent((json.content as never[]) || [])
    if (changed) {
      editor.commands.setContent({ ...json, content }, { emitUpdate })
    } else if (emitUpdate) {
      // 无折叠但调用方要求 emit：补一次（不改内容）
      editor.commands.setContent(json, { emitUpdate })
    }
  } catch (e) {
    console.error('[editor] setContent(markdown) failed:', e)
  }
}
