<template>
  <div ref="el" class="source-editor"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { EditorView, keymap, lineNumbers, highlightActiveLine } from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { defaultKeymap, history, historyKeymap, undoDepth, redoDepth } from '@codemirror/commands'
import { markdown } from '@codemirror/lang-markdown'
import { syntaxHighlighting, defaultHighlightStyle } from '@codemirror/language'
import { createSourceCommands } from './sourceCommands'

const model = defineModel<string>('value', { required: true })

const el = ref<HTMLElement | null>(null)
let view: EditorView | null = null
let updatingFromModel = false

// 源码模式工具栏所需：撤销/重做可用态 + 命令 API（见 sourceCommands.ts）
const canUndo = ref(false)
const canRedo = ref(false)
const cmd = createSourceCommands(() => view)

onMounted(() => {
  if (!el.value) return
  view = new EditorView({
    parent: el.value,
    state: EditorState.create({
      doc: model.value ?? '',
      extensions: [
        lineNumbers(),
        highlightActiveLine(),
        history(),
        // 格式快捷键与富文本对齐（放 defaultKeymap 前，避免被其吞掉）
        keymap.of([
          { key: 'Mod-b', run: () => (cmd.wrapInline('**'), true) },
          { key: 'Mod-i', run: () => (cmd.wrapInline('*'), true) },
          { key: 'Mod-e', run: () => (cmd.wrapInline('`'), true) },
          { key: 'Mod-Shift-s', run: () => (cmd.wrapInline('~~'), true) },
        ]),
        keymap.of([...defaultKeymap, ...historyKeymap]),
        markdown(),
        syntaxHighlighting(defaultHighlightStyle),
        EditorView.lineWrapping,
        EditorView.theme({
          '&': { backgroundColor: 'transparent', color: 'var(--editor-fg)', height: '100%' },
          '.cm-content': {
            fontFamily: 'var(--editor-mono)',
            fontSize: '14px',
            lineHeight: '1.7',
            caretColor: 'var(--editor-accent)',
          },
          '.cm-gutters': {
            backgroundColor: 'transparent',
            color: 'var(--editor-muted)',
            border: 'none',
          },
          '.cm-activeLine': { backgroundColor: 'var(--editor-hover)' },
          '.cm-activeLineGutter': { backgroundColor: 'transparent' },
          '&.cm-focused': { outline: 'none' },
          '.cm-selectionBackground, ::selection': { backgroundColor: 'var(--editor-selection)' },
        }),
        EditorView.updateListener.of((u) => {
          if (u.docChanged && !updatingFromModel) {
            model.value = u.state.doc.toString()
          }
          canUndo.value = undoDepth(u.state) > 0
          canRedo.value = redoDepth(u.state) > 0
        }),
      ],
    }),
  })
})

watch(model, (val) => {
  if (!view) return
  const current = view.state.doc.toString()
  const next = val ?? ''
  if (next === current) return
  // 最小 diff 替换（公共前后缀外的区间），避免整文档替换把光标弹回开头（分栏模式在富文本侧打字时）
  let start = 0
  const minLen = Math.min(current.length, next.length)
  while (start < minLen && current[start] === next[start]) start++
  let endCur = current.length
  let endNext = next.length
  while (endCur > start && endNext > start && current[endCur - 1] === next[endNext - 1]) {
    endCur--
    endNext--
  }
  updatingFromModel = true
  view.dispatch({ changes: { from: start, to: endCur, insert: next.slice(start, endNext) } })
  updatingFromModel = false
})

defineExpose({ cmd, canUndo, canRedo, focus: () => view?.focus() })

onBeforeUnmount(() => {
  view?.destroy()
  view = null
})
</script>

<style scoped>
.source-editor {
  width: 100%;
  height: 100%;
  overflow: auto;
}
</style>
