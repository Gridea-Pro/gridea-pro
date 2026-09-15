/**
 * 行内 AI 续写（ghost text）。纯 ProseMirror 插件（无 .vue 依赖，测试台安全）。
 * 光标在行尾、空选区、防抖后调用 getSuggestion({prefix,suffix})，把结果作为灰色幽灵文本 widget 显示；
 * Tab 接受（插入），Esc 或任何编辑/移动光标则清除。仅在提供 getSuggestion 时生效。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { Extension } from '@tiptap/core'
import { Plugin, PluginKey } from '@tiptap/pm/state'
import { Decoration, DecorationSet } from '@tiptap/pm/view'

export interface AiWritingOptions {
  minChars: number
  debounceMs: number
  getSuggestion?: (ctx: { prefix: string; suffix: string }) => Promise<string>
}

interface AiState {
  text: string
  pos: number | null
  decorations: DecorationSet
}

const aiKey = new PluginKey<AiState>('aiWriting')

function widget(text: string): HTMLElement {
  const span = document.createElement('span')
  span.className = 'ai-suggestion'
  span.textContent = text
  return span
}

type Cancelable<F> = F & { cancel: () => void }

function debounce<F extends (...a: any[]) => void>(fn: F, ms: number): Cancelable<F> {
  let timer: any = null
  const wrapped = ((...a: any[]) => {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => fn(...a), ms)
  }) as Cancelable<F>
  wrapped.cancel = () => {
    if (timer) clearTimeout(timer)
    timer = null
  }
  return wrapped
}

export function createAiWriting(getSuggestion?: AiWritingOptions['getSuggestion']) {
  return Extension.create<AiWritingOptions>({
    name: 'aiWriting',

    addOptions() {
      return { minChars: 2, debounceMs: 700, getSuggestion }
    },

    addKeyboardShortcuts() {
      return {
        Tab: () => {
          const st = aiKey.getState(this.editor.state)
          if (!st?.text || st.pos == null) return false
          const text = st.text
          return this.editor
            .chain()
            .command(({ tr, state, dispatch }) => {
              if (!dispatch) return true
              tr.insertText(text, state.selection.from)
              dispatch(tr.setMeta(aiKey, { clear: true }))
              return true
            })
            .focus()
            .run()
        },
        Escape: () => {
          const st = aiKey.getState(this.editor.state)
          if (!st?.text) return false
          this.editor.view.dispatch(this.editor.state.tr.setMeta(aiKey, { clear: true }))
          return true
        },
      }
    },

    addProseMirrorPlugins() {
      const opts = this.options
      const editor = this.editor
      let lastReq = 0

      const fetchSuggestion = debounce((view: any) => {
        const fn = opts.getSuggestion
        if (!fn) return
        const { selection, doc } = view.state
        if (!selection.empty) return
        const from = selection.from
        const suffixLine = doc.textBetween(from, doc.content.size, '\n', '\n') as string
        const nl = suffixLine.indexOf('\n')
        const afterCursor = nl >= 0 ? suffixLine.slice(0, nl) : suffixLine
        if (afterCursor.trim().length > 0) return // 仅在行尾触发
        const prefix = doc.textBetween(Math.max(0, from - 2000), from, '\n', '\n') as string
        const suffix = doc.textBetween(from, Math.min(doc.content.size, from + 2000), '\n', '\n') as string
        if (prefix.trim().length < (opts.minChars ?? 0)) return
        const reqId = ++lastReq
        fn({ prefix, suffix })
          .then((text) => {
            if (reqId !== lastReq || !text) return
            if (!editor.state.selection.empty) return
            // 光标自请求发出后已移动（如撤销/点击），丢弃陈旧建议，避免落到错误位置
            if (editor.state.selection.from !== from) return
            editor.view.dispatch(editor.state.tr.setMeta(aiKey, { text, pos: from }))
          })
          .catch(() => {})
      }, this.options.debounceMs)

      return [
        new Plugin<AiState>({
          key: aiKey,
          state: {
            init: () => ({ text: '', pos: null, decorations: DecorationSet.empty }),
            apply(tr, value, _old, newState) {
              const meta = tr.getMeta(aiKey)
              if (meta?.clear) return { text: '', pos: null, decorations: DecorationSet.empty }
              if (meta?.text != null) {
                const pos = meta.pos ?? newState.selection.from
                const deco = Decoration.widget(pos, () => widget(meta.text), { side: 1, ignoreSelection: true })
                return { text: meta.text, pos, decorations: DecorationSet.create(newState.doc, [deco]) }
              }
              if ((tr.docChanged || tr.selectionSet) && value.text) {
                return { text: '', pos: null, decorations: DecorationSet.empty }
              }
              return value
            },
          },
          props: {
            decorations(state) {
              return aiKey.getState(state)?.decorations
            },
          },
          view() {
            return {
              update(view, prev) {
                if (!view.state.doc.eq(prev.doc) || !view.state.selection.eq(prev.selection)) {
                  fetchSuggestion(view)
                }
              },
              destroy() {
                // 编辑器销毁时取消挂起的防抖请求，避免定时器泄漏
                fetchSuggestion.cancel()
              },
            }
          },
        }),
      ]
    },
  })
}
