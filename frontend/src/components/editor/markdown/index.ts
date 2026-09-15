/**
 * Markdown 往返入口。基于官方 @tiptap/markdown：
 * - editor.getMarkdown() 序列化
 * - editor.commands.setContent(md, { contentType: 'markdown' }) 反序列化
 */
import type { Editor } from '@tiptap/core'
import { canonicalizeMarkdown } from './canonicalize'
import { foldDetailsContent, foldAlignContent } from './foldDetails'
import { toast } from '@/helpers/toast'
import i18n from '@/locales'

// 记录每个编辑器上一次成功序列化的 Markdown，用于序列化失败时兜底。
const lastGoodMarkdown = new WeakMap<Editor, string>()
// 当前处于「序列化失败」状态的编辑器。getMarkdown 每次输入都会调用，若每次失败都 toast 会刷屏，
// 故只在「首次进入失败」时提示一次；恢复成功后清除，下次再失败可再提示。
const serializationFailing = new WeakSet<Editor>()

/** 导出 Markdown（规范化为幂等形式：表格最小内边距等，保证 once-saved-then-stable） */
export function getMarkdown(editor: Editor): string {
  try {
    const md = canonicalizeMarkdown(editor.getMarkdown())
    lastGoodMarkdown.set(editor, md)
    serializationFailing.delete(editor)
    return md
  } catch (e) {
    // 关键：序列化失败绝不能返回 ''。onUpdate 会把返回值写回 model.value（即 form.content），
    // 返回空串会导致正文被静默清空、一旦保存就落盘为空文件、整篇文章内容丢失。
    // 改为返回上一次成功的 Markdown（保留已知良好内容）；无历史时退回纯文本，也好过空串。
    console.error('[editor] getMarkdown 序列化失败，返回上一次成功结果以避免清空正文:', e)
    // 用户可见提示（去重：一次失败episode只提示一次），否则内容持续被兜底冻结却毫无察觉。
    if (!serializationFailing.has(editor)) {
      serializationFailing.add(editor)
      try {
        toast.error(i18n.global.t('editor.serializeFailed'))
      } catch {
        /* i18n 未就绪时静默，不影响兜底逻辑 */
      }
    }
    const last = lastGoodMarkdown.get(editor)
    if (last !== undefined) return last
    try {
      return editor.getText()
    } catch {
      return ''
    }
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
