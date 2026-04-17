<template>
  <div class="tiptap-editor-shell" :style="{ '--editor-font-family': themeStore.editorFontFamily }">
    <div class="editor-surface" @click="focusEditor">
      <BubbleMenu
        v-if="editor"
        :editor="editor"
        class="bubble-menu"
        :should-show="shouldShowBubbleMenu"
      >
        <button type="button" class="bubble-btn" :class="{ 'is-active': isMarkActive('bold') }" @mousedown.prevent @click="toggleBold">B</button>
        <button type="button" class="bubble-btn" :class="{ 'is-active': isMarkActive('italic') }" @mousedown.prevent @click="toggleItalic">I</button>
        <button type="button" class="bubble-btn" :class="{ 'is-active': isMarkActive('strike') }" @mousedown.prevent @click="toggleStrike">S</button>
        <button type="button" class="bubble-btn" :class="{ 'is-active': isMarkActive('code') }" @mousedown.prevent @click="toggleInlineCode">&lt;/&gt;</button>
        <button type="button" class="bubble-btn" :class="{ 'is-active': isNodeActive('blockquote') }" @mousedown.prevent @click="toggleBlockquote">"</button>
      </BubbleMenu>

      <EditorContent :editor="editor" class="tiptap-editor-content" />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { EditorContent, useEditor } from '@tiptap/vue-3'
import { BubbleMenu } from '@tiptap/vue-3/menus'
import type { Editor } from '@tiptap/core'

import { useThemeStore } from '@/stores/theme'

import { createArticleEditorExtensions } from './extensions'
import { preprocessMarkdownForTiptap } from './markdown-passthrough'

interface Props {
  placeholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: '',
})

const modelValue = defineModel<string>('value', { required: true })

const emit = defineEmits<{
  keydown: [event: KeyboardEvent]
  focus: []
}>()

const themeStore = useThemeStore()
const isSyncing = ref(false)

const normalizeMarkdown = (markdown: string) => markdown.replace(/\r\n?/g, '\n')
const serializeMarkdown = (editor: Editor) => normalizeMarkdown(editor.getMarkdown() || '')

const editor = useEditor({
  extensions: createArticleEditorExtensions(props.placeholder),
  content: preprocessMarkdownForTiptap(modelValue.value || ''),
  contentType: 'markdown',
  editorProps: {
    attributes: {
      class: 'gridea-rich-editor',
      'data-tiptap-root': 'true',
      spellcheck: 'false',
    },
    handleKeyDown: (_, event) => {
      emit('keydown', event)
      return false
    },
  },
  onFocus: () => {
    emit('focus')
  },
  onUpdate: ({ editor: currentEditor }) => {
    if (isSyncing.value) {
      return
    }

    const markdown = serializeMarkdown(currentEditor)
    if (markdown !== modelValue.value) {
      modelValue.value = markdown
    }
  },
})

const focusEditor = () => {
  editor.value?.commands.focus()
}

const isMarkActive = (name: string) => editor.value?.isActive(name) ?? false
const isNodeActive = (name: string, attrs?: Record<string, unknown>) => editor.value?.isActive(name, attrs) ?? false

const run = (callback: (currentEditor: Editor) => boolean) => {
  if (!editor.value) {
    return
  }

  callback(editor.value)
}

const toggleHeading = (level: 1 | 2 | 3 | 4 | 5 | 6) => run((currentEditor) => currentEditor.chain().focus().toggleHeading({ level }).run())
const toggleBold = () => run((currentEditor) => currentEditor.chain().focus().toggleBold().run())
const toggleItalic = () => run((currentEditor) => currentEditor.chain().focus().toggleItalic().run())
const toggleStrike = () => run((currentEditor) => currentEditor.chain().focus().toggleStrike().run())
const toggleInlineCode = () => run((currentEditor) => currentEditor.chain().focus().toggleCode().run())
const toggleBulletList = () => run((currentEditor) => currentEditor.chain().focus().toggleBulletList().run())
const toggleOrderedList = () => run((currentEditor) => currentEditor.chain().focus().toggleOrderedList().run())
const toggleTaskList = () => run((currentEditor) => currentEditor.chain().focus().toggleTaskList().run())
const toggleBlockquote = () => run((currentEditor) => currentEditor.chain().focus().toggleBlockquote().run())
const toggleCodeBlock = () => run((currentEditor) => currentEditor.chain().focus().toggleCodeBlock().run())

const shouldShowBubbleMenu = computed(() => {
  return ({ editor: currentEditor }: { editor: Editor }) => {
    if (!currentEditor.isEditable || currentEditor.state.selection.empty) {
      return false
    }

    return !currentEditor.isActive('rawMarkdownInline')
  }
})

watch(
  modelValue,
  (nextValue) => {
    const currentEditor = editor.value
    if (!currentEditor) {
      return
    }

    const normalizedExternal = normalizeMarkdown(nextValue || '')
    if (serializeMarkdown(currentEditor) === normalizedExternal) {
      return
    }

    isSyncing.value = true
    currentEditor.commands.setContent(preprocessMarkdownForTiptap(normalizedExternal), {
      contentType: 'markdown',
    })
    isSyncing.value = false
  },
)

onBeforeUnmount(() => {
  editor.value?.destroy()
})

defineExpose({
  editor,
  focusEditor,
  toggleHeading,
  toggleBold,
  toggleItalic,
  toggleStrike,
  toggleInlineCode,
  toggleBulletList,
  toggleOrderedList,
  toggleTaskList,
  toggleBlockquote,
  toggleCodeBlock
})
</script>

<style lang="less" scoped>
.tiptap-editor-shell {
  --editor-font-family: ui-monospace, Menlo, Monaco, "Cascadia Code", "Segoe UI Mono", Consolas, "Courier New", monospace;
  min-height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.bubble-btn {
  border: none;
  border-radius: 999px;
  background: transparent;
  color: var(--muted-foreground);
  cursor: pointer;
  transition: background-color 0.18s ease, color 0.18s ease, transform 0.18s ease;
  min-width: 32px;
  height: 32px;
  padding: 0 10px;
  font-size: 12px;
  font-weight: 700;
}

.bubble-btn:hover,
.bubble-btn.is-active {
  color: var(--foreground);
  background: color-mix(in srgb, var(--primary) 12%, transparent);
}

.bubble-btn:active {
  transform: translateY(1px);
}

.editor-surface {
  flex: 1;
  min-height: 480px;
  border-radius: 28px;
  border: 1px solid color-mix(in srgb, var(--border) 82%, transparent);
  background:
    radial-gradient(circle at top left, color-mix(in srgb, var(--primary) 8%, transparent), transparent 32%),
    linear-gradient(180deg, color-mix(in srgb, var(--background) 96%, #fff 4%), var(--background));
  box-shadow: 0 18px 36px color-mix(in srgb, #000 5%, transparent);
  padding: 24px 28px 48px;
}

.bubble-menu {
  display: flex;
  gap: 6px;
  padding: 8px;
  border-radius: 14px;
  border: 1px solid color-mix(in srgb, var(--border) 72%, transparent);
  background: color-mix(in srgb, var(--background) 95%, transparent);
  box-shadow: 0 12px 30px color-mix(in srgb, #000 8%, transparent);
  backdrop-filter: blur(10px);
}

.tiptap-editor-content {
  min-height: 100%;
}

:deep(.gridea-rich-editor) {
  min-height: 440px;
  outline: none;
  color: var(--foreground);
  font-size: 16px;
  line-height: 1.9;
  font-family: var(--editor-font-family);
  white-space: pre-wrap;
}

:deep(.gridea-rich-editor > :first-child) {
  margin-top: 0;
}

:deep(.gridea-rich-editor p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  float: left;
  color: color-mix(in srgb, var(--muted-foreground) 78%, transparent);
  pointer-events: none;
  height: 0;
}

:deep(.gridea-rich-editor p) {
  margin: 0 0 1rem;
}

:deep(.gridea-rich-editor h1),
:deep(.gridea-rich-editor h2),
:deep(.gridea-rich-editor h3),
:deep(.gridea-rich-editor h4) {
  margin: 1.4rem 0 0.9rem;
  font-weight: 700;
  line-height: 1.3;
}

:deep(.gridea-rich-editor h2) {
  font-size: 1.45rem;
}

:deep(.gridea-rich-editor blockquote) {
  margin: 1.25rem 0;
  padding-left: 1rem;
  border-left: 3px solid color-mix(in srgb, var(--primary) 36%, transparent);
  color: var(--muted-foreground);
}

:deep(.gridea-rich-editor ul),
:deep(.gridea-rich-editor ol) {
  padding-left: 1.5rem;
  margin: 0 0 1rem;
}

:deep(.gridea-rich-editor code) {
  font-family: var(--editor-font-family);
  font-size: 0.9em;
  background: color-mix(in srgb, var(--secondary) 70%, transparent);
  border-radius: 8px;
  padding: 0.2rem 0.4rem;
}

:deep(.gridea-rich-editor pre) {
  margin: 1rem 0;
  border-radius: 18px;
  padding: 18px;
  background: color-mix(in srgb, var(--secondary) 76%, transparent);
  overflow-x: auto;
}

:deep(.gridea-rich-editor pre code) {
  padding: 0;
  background: transparent;
}

:deep(.gridea-rich-editor a) {
  color: var(--foreground);
  text-decoration: underline;
  text-decoration-color: color-mix(in srgb, var(--primary) 35%, transparent);
}

:deep(.gridea-rich-editor table) {
  width: 100%;
  margin: 1rem 0;
  border-collapse: collapse;
  overflow: hidden;
  border-radius: 14px;
}

:deep(.gridea-rich-editor th),
:deep(.gridea-rich-editor td) {
  border: 1px solid color-mix(in srgb, var(--border) 78%, transparent);
  padding: 10px 12px;
  vertical-align: top;
}

:deep(.gridea-rich-editor th) {
  background: color-mix(in srgb, var(--secondary) 72%, transparent);
}

:deep(.gridea-rich-editor img) {
  display: block;
  max-width: 100%;
  margin: 1rem auto;
  border-radius: 18px;
}

:deep([data-gridea-more]) {
  position: relative;
  margin: 1.6rem 0;
  padding: 12px 0;
  text-align: center;
}

:deep([data-gridea-more]::before) {
  content: "";
  position: absolute;
  left: 0;
  right: 0;
  top: 50%;
  height: 1px;
  background:
    linear-gradient(90deg, transparent, color-mix(in srgb, var(--primary) 34%, transparent), transparent);
}

:deep(.gridea-more__label) {
  position: relative;
  z-index: 1;
  display: inline-block;
  padding: 0 12px;
  background: var(--background);
  letter-spacing: 0.24em;
  font-size: 11px;
  font-weight: 700;
  color: var(--muted-foreground);
}

:deep([data-raw-html-block]),
:deep([data-raw-footnote-definition]) {
  margin: 1rem 0;
  padding: 14px 16px;
  border-radius: 18px;
  border: 1px dashed color-mix(in srgb, var(--border) 72%, transparent);
  background: color-mix(in srgb, var(--secondary) 40%, transparent);
}

:deep(.gridea-raw-block__label) {
  margin-bottom: 8px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--muted-foreground);
}

:deep(.gridea-raw-block__content) {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 12px;
  line-height: 1.6;
  font-family: var(--editor-font-family);
  color: var(--foreground);
}

:deep([data-raw-markdown-inline]) {
  display: inline-flex;
  align-items: center;
  max-width: 100%;
  margin: 0 0.2rem;
  padding: 0.12rem 0.5rem;
  border-radius: 999px;
  background: color-mix(in srgb, var(--secondary) 78%, transparent);
  color: var(--foreground);
  font-size: 0.82em;
  line-height: 1.4;
  white-space: break-spaces;
}

@media (max-width: 768px) {
  .editor-surface {
    padding: 18px 16px 32px;
    border-radius: 22px;
  }
}
</style>
