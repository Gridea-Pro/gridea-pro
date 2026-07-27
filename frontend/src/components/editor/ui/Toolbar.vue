<template>
  <div class="editor-toolbar">
    <div class="tb-group">
      <ToolbarButton :title="t('editor.undo')" :disabled="!can('undo')" @click="run((c) => c.undo())">
        <Undo2 />
      </ToolbarButton>
      <ToolbarButton :title="t('editor.redo')" :disabled="!can('redo')" @click="run((c) => c.redo())">
        <Redo2 />
      </ToolbarButton>
    </div>

    <div class="tb-sep" />

    <div class="tb-group">
      <ToolbarButton :title="t('editor.h1')" :active="active('heading', { level: 1 })"
        @click="run((c) => c.toggleHeading({ level: 1 }))"><Heading1 /></ToolbarButton>
      <ToolbarButton :title="t('editor.h2')" :active="active('heading', { level: 2 })"
        @click="run((c) => c.toggleHeading({ level: 2 }))"><Heading2 /></ToolbarButton>
      <ToolbarButton :title="t('editor.h3')" :active="active('heading', { level: 3 })"
        @click="run((c) => c.toggleHeading({ level: 3 }))"><Heading3 /></ToolbarButton>
    </div>

    <div class="tb-sep" />

    <div class="tb-group">
      <ToolbarButton :title="t('editor.bold')" :active="active('bold')" @click="run((c) => c.toggleBold())"><Bold /></ToolbarButton>
      <ToolbarButton :title="t('editor.italic')" :active="active('italic')" @click="run((c) => c.toggleItalic())"><Italic /></ToolbarButton>
      <ToolbarButton :title="t('editor.strike')" :active="active('strike')" @click="run((c) => c.toggleStrike())"><Strikethrough /></ToolbarButton>
      <ToolbarButton :title="t('editor.code')" :active="active('code')" @click="run((c) => c.toggleCode())"><Code /></ToolbarButton>
      <ToolbarButton :title="t('editor.highlight')" :active="active('highlight')" @click="run((c) => c.toggleHighlight())"><Highlighter /></ToolbarButton>
      <ToolbarButton :title="t('editor.sub')" :active="active('subscript')" @click="run((c) => c.toggleSubscript())"><SubscriptIcon /></ToolbarButton>
      <ToolbarButton :title="t('editor.sup')" :active="active('superscript')" @click="run((c) => c.toggleSuperscript())"><SuperscriptIcon /></ToolbarButton>
    </div>

    <div class="tb-sep" />

    <div class="tb-group">
      <ToolbarButton :title="t('editor.bulletList')" :active="active('bulletList')" @click="run((c) => c.toggleBulletList())"><List /></ToolbarButton>
      <ToolbarButton :title="t('editor.orderedList')" :active="active('orderedList')" @click="run((c) => c.toggleOrderedList())"><ListOrdered /></ToolbarButton>
      <ToolbarButton :title="t('editor.taskList')" :active="active('taskList')" @click="run((c) => c.toggleTaskList())"><ListChecks /></ToolbarButton>
      <ToolbarButton :title="t('editor.blockquote')" :active="active('blockquote')" @click="run((c) => c.toggleBlockquote())"><Quote /></ToolbarButton>
      <ToolbarButton :title="t('editor.codeBlock')" :active="active('codeBlock')" @click="run((c) => c.toggleCodeBlock())"><Code2 /></ToolbarButton>
      <ToolbarButton :title="t('editor.hr')" @click="run((c) => c.setHorizontalRule())"><Minus /></ToolbarButton>
    </div>

    <div class="tb-sep" />

    <div class="tb-group">
      <ToolbarButton :title="t('editor.link')" :active="active('link')" @click="emit('link')"><LinkIcon /></ToolbarButton>
      <ToolbarButton :title="t('editor.image')" @click="emit('image')"><ImageIcon /></ToolbarButton>
      <ToolbarButton :title="t('editor.table')" @click="run((c) => c.insertTable({ rows: 3, cols: 3, withHeaderRow: true }))"><TableIcon /></ToolbarButton>
      <ColorPicker :model-value="currentColor" @select="emit('color', $event)" />
      <EmojiPicker @select="emit('emoji', $event)" />
      <ToolbarButton :title="t('editor.aiPolish')" @click="emit('polish')"><Sparkles /></ToolbarButton>
    </div>

    <div class="tb-spacer" />

    <div class="tb-group">
      <ToolbarButton :title="t('editor.rich')" :active="mode === 'rich'" @click="emit('update:mode', 'rich')"><Eye /></ToolbarButton>
      <ToolbarButton :title="t('editor.split')" :active="mode === 'split'" @click="emit('update:mode', 'split')"><Columns2 /></ToolbarButton>
      <ToolbarButton :title="t('editor.source')" :active="mode === 'source'" @click="emit('update:mode', 'source')"><FileCode2 /></ToolbarButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Editor, ChainedCommands } from '@tiptap/vue-3'
import type { EditorMode } from '../types'
import ToolbarButton from './ToolbarButton.vue'
import ColorPicker from './ColorPicker.vue'
import EmojiPicker from './EmojiPicker.vue'
import {
  Undo2, Redo2, Bold, Italic, Strikethrough, Code, Code2,
  Heading1, Heading2, Heading3, Highlighter,
  SubscriptIcon, SuperscriptIcon,
  List, ListOrdered, ListChecks, Quote, Minus,
  Link as LinkIcon, Image as ImageIcon, Table as TableIcon,
  Sparkles, Eye, Columns2, FileCode2,
} from 'lucide-vue-next'

const props = defineProps<{ editor: Editor | null | undefined; mode: EditorMode }>()
const emit = defineEmits<{
  link: []
  image: []
  polish: []
  color: [hex: string]
  emoji: [emoji: string]
  'update:mode': [mode: EditorMode]
}>()

const { t } = useI18n()

// 让 active/can/颜色 随选区与文档实时刷新；editor 异步就绪，用 watch 待其出现再绑定
const tick = ref(0)
function bump() {
  tick.value++
}
watch(
  () => props.editor,
  (ed, prev) => {
    prev?.off('transaction', bump)
    prev?.off('selectionUpdate', bump)
    ed?.on('transaction', bump)
    ed?.on('selectionUpdate', bump)
  },
  { immediate: true },
)
onBeforeUnmount(() => {
  props.editor?.off('transaction', bump)
  props.editor?.off('selectionUpdate', bump)
})

const currentColor = computed(() => {
  void tick.value
  return (props.editor?.getAttributes('textStyle').color as string) || ''
})

function run(fn: (c: ChainedCommands) => ChainedCommands) {
  const e = props.editor
  if (!e) return
  fn(e.chain().focus()).run()
}

function can(name: 'undo' | 'redo'): boolean {
  void tick.value
  const e = props.editor
  if (!e) return false
  return name === 'undo' ? e.can().undo() : e.can().redo()
}

function active(nameOrAttrs: string | Record<string, unknown>, attrs?: Record<string, unknown>): boolean {
  void tick.value
  const e = props.editor
  if (!e) return false
  return typeof nameOrAttrs === 'string'
    ? e.isActive(nameOrAttrs, attrs)
    : e.isActive(nameOrAttrs)
}
</script>

<style scoped>
.editor-toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 2px;
  padding: 6px 8px;
  border-bottom: 1px solid var(--editor-border);
  background: var(--editor-bg);
  position: sticky;
  top: 0;
  z-index: 5;
}
.tb-group {
  display: flex;
  align-items: center;
  gap: 1px;
}
.tb-sep {
  width: 1px;
  height: 18px;
  margin: 0 4px;
  background: var(--editor-border);
}
.tb-spacer {
  flex: 1;
}
</style>
