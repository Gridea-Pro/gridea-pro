<template>
  <div ref="el" class="source-editor"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { EditorView, keymap, lineNumbers, highlightActiveLine } from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { markdown } from '@codemirror/lang-markdown'
import { syntaxHighlighting, defaultHighlightStyle } from '@codemirror/language'

const model = defineModel<string>('value', { required: true })

const el = ref<HTMLElement | null>(null)
let view: EditorView | null = null
let updatingFromModel = false

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
        }),
      ],
    }),
  })
})

watch(model, (val) => {
  if (!view) return
  const current = view.state.doc.toString()
  if (val === current) return
  updatingFromModel = true
  view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: val ?? '' } })
  updatingFromModel = false
})

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
