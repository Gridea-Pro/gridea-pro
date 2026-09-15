/**
 * 源码模式工具栏命令：对 CodeMirror 选区做 Markdown 文本变换。
 * 工具栏在 mode==='source' 时分发到这里，而不是隐藏的 Tiptap 实例
 * （Tiptap 在源码模式下选区/文档都是陈旧的，直接发命令会用旧文档覆盖 model）。
 *
 * 语法方言与本编辑器序列化器对齐：加粗 **、斜体 *、删除线 ~~、行内代码 `、
 * 上标 ^x^、下标 ~x~（见 extensions/Script.ts）、任务列表 - [ ]、more 分隔 <!-- more -->。
 */
import type { EditorView } from '@codemirror/view'
import { undo as cmUndo, redo as cmRedo } from '@codemirror/commands'

export type SourceLineKind = 'bullet' | 'ordered' | 'task' | 'quote'

export interface SourceCommandApi {
  /** 行内 mark 切换：选区被 marker 包裹（内含或紧邻）则解开，否则包裹；空选区插入一对并把光标置于其间 */
  wrapInline(marker: string): void
  /** 行前缀切换（无序/有序/任务/引用），作用于选区覆盖的所有非空行 */
  toggleLine(kind: SourceLineKind): void
  /** 标题级别：0 = 正文（仅去前缀），1-6 = 对应 # 数量 */
  setHeading(level: number): void
  /** 代码块围栏切换：选中行已被 ``` 围栏包住则拆围栏，否则加围栏 */
  toggleCodeBlock(): void
  /** 以独立段落插入块级内容（分割线、表格骨架、more 分隔等） */
  insertBlock(text: string): void
  /** 在光标处插入行内文本（emoji 等），有选区则替换 */
  insertText(text: string): void
  /** 用 text 替换当前选区（链接/图片 Markdown、AI 润色结果） */
  replaceSelection(text: string): void
  getSelectionText(): string
  undo(): void
  redo(): void
  focus(): void
}

/** 3×3 表格骨架（首行表头留空待填） */
export const TABLE_SKELETON = ['|  |  |  |', '| --- | --- | --- |', '|  |  |  |', '|  |  |  |'].join('\n')

export function createSourceCommands(getView: () => EditorView | null): SourceCommandApi {
  function withView(fn: (view: EditorView) => void) {
    const view = getView()
    if (!view) return
    fn(view)
    view.focus()
  }

  function wrapInline(marker: string) {
    withView((view) => {
      const { state } = view
      const range = state.selection.main
      const m = marker
      if (range.empty) {
        view.dispatch({
          changes: { from: range.from, insert: m + m },
          selection: { anchor: range.from + m.length },
          userEvent: 'input',
        })
        return
      }
      const text = state.sliceDoc(range.from, range.to)
      const before = state.sliceDoc(Math.max(0, range.from - m.length), range.from)
      const after = state.sliceDoc(range.to, Math.min(state.doc.length, range.to + m.length))
      if (before === m && after === m) {
        // 选区紧邻外侧已有 marker → 删外侧
        view.dispatch({
          changes: [
            { from: range.from - m.length, to: range.from },
            { from: range.to, to: range.to + m.length },
          ],
          selection: { anchor: range.from - m.length, head: range.to - m.length },
          userEvent: 'delete',
        })
        return
      }
      if (text.startsWith(m) && text.endsWith(m) && text.length >= m.length * 2) {
        // 选区自身含 marker → 剥掉
        const inner = text.slice(m.length, text.length - m.length)
        view.dispatch({
          changes: { from: range.from, to: range.to, insert: inner },
          selection: { anchor: range.from, head: range.from + inner.length },
          userEvent: 'delete',
        })
        return
      }
      view.dispatch({
        changes: [
          { from: range.from, insert: m },
          { from: range.to, insert: m },
        ],
        selection: { anchor: range.from + m.length, head: range.to + m.length },
        userEvent: 'input',
      })
    })
  }

  // 行前缀识别：task 必须先于 bullet 判断（- [ ] 也满足 bullet 形态）
  const TASK_RE = /^(\s*)[-*+] \[[ xX]\] /
  const BULLET_RE = /^(\s*)[-*+] (?!\[[ xX]\] )/
  const ORDERED_RE = /^(\s*)\d+[.)] /
  const QUOTE_RE = /^(\s*)> ?/

  /** 剥掉行上已有的列表/任务前缀（保留缩进），返回 [缩进, 正文] */
  function stripListPrefix(text: string): [string, string] {
    for (const re of [TASK_RE, ORDERED_RE, BULLET_RE]) {
      const match = re.exec(text)
      if (match) return [match[1], text.slice(match[0].length)]
    }
    const indent = /^\s*/.exec(text)?.[0] ?? ''
    return [indent, text.slice(indent.length)]
  }

  /** 选区覆盖的行号区间 */
  function selectedLines(view: EditorView): { start: number; end: number } {
    const { state } = view
    const range = state.selection.main
    return { start: state.doc.lineAt(range.from).number, end: state.doc.lineAt(range.to).number }
  }

  function toggleLine(kind: SourceLineKind) {
    withView((view) => {
      const { state } = view
      const { start, end } = selectedLines(view)
      const matchRe = { bullet: BULLET_RE, ordered: ORDERED_RE, task: TASK_RE, quote: QUOTE_RE }[kind]

      const lines: { from: number; to: number; text: string }[] = []
      for (let n = start; n <= end; n++) {
        const line = state.doc.line(n)
        lines.push({ from: line.from, to: line.to, text: line.text })
      }
      const nonEmpty = lines.filter((l) => l.text.trim().length > 0)
      // 选区全是空行（或光标停在空行）时也要能"开启"该格式
      const targets = nonEmpty.length ? nonEmpty : lines
      const allMarked = nonEmpty.length > 0 && nonEmpty.every((l) => matchRe.test(l.text))

      let ordinal = 0
      const changes = targets.map((l) => {
        let next: string
        if (kind === 'quote') {
          next = allMarked ? l.text.replace(QUOTE_RE, '$1') : `> ${l.text}`
        } else {
          const [indent, body] = stripListPrefix(l.text)
          if (allMarked) {
            next = indent + body
          } else {
            ordinal++
            const prefix = kind === 'bullet' ? '- ' : kind === 'task' ? '- [ ] ' : `${ordinal}. `
            next = indent + prefix + body
          }
        }
        return { from: l.from, to: l.to, insert: next }
      })
      view.dispatch({ changes, userEvent: allMarked ? 'delete' : 'input' })
    })
  }

  function setHeading(level: number) {
    withView((view) => {
      const { state } = view
      const { start, end } = selectedLines(view)
      const changes = []
      for (let n = start; n <= end; n++) {
        const line = state.doc.line(n)
        const body = line.text.replace(/^\s*#{1,6}\s+/, '')
        const next = level > 0 ? `${'#'.repeat(level)} ${body}` : body
        if (next !== line.text) changes.push({ from: line.from, to: line.to, insert: next })
      }
      if (changes.length) view.dispatch({ changes, userEvent: 'input' })
    })
  }

  function toggleCodeBlock() {
    withView((view) => {
      const { state } = view
      const range = state.selection.main
      const firstLine = state.doc.lineAt(range.from)
      const lastLine = state.doc.lineAt(range.to)
      const isFence = (text: string) => /^\s*```/.test(text)

      // 已在围栏内（紧邻上下行是 ```）→ 删除两条围栏行（含换行符）
      const prev = firstLine.number > 1 ? state.doc.line(firstLine.number - 1) : null
      const next = lastLine.number < state.doc.lines ? state.doc.line(lastLine.number + 1) : null
      if (prev && next && isFence(prev.text) && isFence(next.text)) {
        view.dispatch({
          changes: [
            { from: prev.from, to: firstLine.from },
            { from: lastLine.to, to: next.to },
          ],
          userEvent: 'delete',
        })
        return
      }
      // 选区首尾行自身就是一对围栏 → 拆掉
      if (firstLine.number < lastLine.number && isFence(firstLine.text) && isFence(lastLine.text)) {
        view.dispatch({
          changes: [
            { from: firstLine.from, to: Math.min(firstLine.to + 1, state.doc.length) },
            { from: Math.max(lastLine.from - 1, 0), to: lastLine.to },
          ],
          userEvent: 'delete',
        })
        return
      }
      view.dispatch({
        changes: [
          { from: firstLine.from, insert: '```\n' },
          { from: lastLine.to, insert: '\n```' },
        ],
        // 光标落在开围栏的语言位，方便直接补语言名
        selection: { anchor: firstLine.from + 3 },
        userEvent: 'input',
      })
    })
  }

  function insertBlock(text: string) {
    withView((view) => {
      const { state } = view
      const range = state.selection.main
      const line = state.doc.lineAt(range.to)
      if (line.text.trim().length === 0) {
        view.dispatch({
          changes: { from: line.from, to: line.to, insert: text },
          selection: { anchor: line.from + text.length },
          userEvent: 'input',
        })
      } else {
        const insert = `\n\n${text}`
        view.dispatch({
          changes: { from: line.to, insert },
          selection: { anchor: line.to + insert.length },
          userEvent: 'input',
        })
      }
    })
  }

  function replaceSelection(text: string) {
    withView((view) => {
      const range = view.state.selection.main
      view.dispatch({
        changes: { from: range.from, to: range.to, insert: text },
        selection: { anchor: range.from + text.length },
        userEvent: 'input',
      })
    })
  }

  return {
    wrapInline,
    toggleLine,
    setHeading,
    toggleCodeBlock,
    insertBlock,
    insertText: replaceSelection,
    replaceSelection,
    getSelectionText() {
      const view = getView()
      if (!view) return ''
      const range = view.state.selection.main
      return view.state.sliceDoc(range.from, range.to)
    },
    undo() {
      withView((view) => cmUndo(view))
    },
    redo() {
      withView((view) => cmRedo(view))
    },
    focus() {
      getView()?.focus()
    },
  }
}
