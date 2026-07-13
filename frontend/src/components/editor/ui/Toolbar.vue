<template>
  <div class="editor-toolbar">
    <div class="tb-group">
      <ToolbarButton :title="t('editor.undo')" :disabled="!can('undo')" @click="exec('undo')">
        <Undo2 />
      </ToolbarButton>
      <ToolbarButton :title="t('editor.redo')" :disabled="!can('redo')" @click="exec('redo')">
        <Redo2 />
      </ToolbarButton>
    </div>

    <div class="tb-sep" />

    <div class="tb-group">
      <HeadingSelect :editor="editor" :source-heading="isSource ? source?.cmd.setHeading : null" />
      <FontSizeSelect :editor="editor" :disabled="isSource" />
    </div>

    <div class="tb-sep" />

    <div class="tb-group">
      <ToolbarButton :title="t('editor.bold')" :active="active('bold')" @click="exec('bold')"><Bold /></ToolbarButton>
      <ToolbarButton :title="t('editor.italic')" :active="active('italic')" @click="exec('italic')"><Italic /></ToolbarButton>
      <ToolbarButton :title="t('editor.strike')" :active="active('strike')" @click="exec('strike')"><Strikethrough /></ToolbarButton>
      <ToolbarButton :title="t('editor.code')" :active="active('code')" @click="exec('code')"><Code /></ToolbarButton>
      <ColorPicker type="text" :model-value="currentColor" :disabled="isSource" @select="emit('color', $event)" />
      <ColorPicker type="highlight" :model-value="currentHighlight" :disabled="isSource" @select="emit('highlight', $event)" />
    </div>

    <div class="tb-sep" />

    <div class="tb-group">
      <ToolbarButton :title="t('editor.bulletList')" :active="active('bulletList')" @click="exec('bulletList')"><List /></ToolbarButton>
      <ToolbarButton :title="t('editor.orderedList')" :active="active('orderedList')" @click="exec('orderedList')"><ListOrdered /></ToolbarButton>
      <ToolbarButton :title="t('editor.taskList')" :active="active('taskList')" @click="exec('taskList')"><ListChecks /></ToolbarButton>
      <ToolbarButton :title="t('editor.blockquote')" :active="active('blockquote')" @click="exec('blockquote')"><Quote /></ToolbarButton>
      <ToolbarButton :title="t('editor.codeBlock')" :active="active('codeBlock')" @click="exec('codeBlock')"><Code2 /></ToolbarButton>
      <ToolbarButton :title="t('editor.hr')" @click="exec('hr')"><Minus /></ToolbarButton>
    </div>

    <div class="tb-sep" />

    <div class="tb-group">
      <ToolbarButton :title="t('editor.shortcuts.alignLeft')" :active="active({ textAlign: 'left' })" :disabled="isSource"
        @click="exec('alignLeft')"><AlignLeft /></ToolbarButton>
      <ToolbarButton :title="t('editor.shortcuts.alignCenter')" :active="active({ textAlign: 'center' })" :disabled="isSource"
        @click="exec('alignCenter')"><AlignCenter /></ToolbarButton>
      <ToolbarButton :title="t('editor.shortcuts.alignRight')" :active="active({ textAlign: 'right' })" :disabled="isSource"
        @click="exec('alignRight')"><AlignRight /></ToolbarButton>
    </div>

    <div class="tb-sep" />

    <div class="tb-group">
      <ToolbarButton :title="t('editor.link')" :active="active('link')" @click="emit('link')"><LinkIcon /></ToolbarButton>
      <ToolbarButton :title="t('editor.image')" @click="emit('image')"><ImageIcon /></ToolbarButton>
      <ToolbarButton :title="t('editor.table')" @click="exec('table')"><TableIcon /></ToolbarButton>
      <EmojiPicker @select="emit('emoji', $event)" />
      <ToolbarButton :title="t('editor.aiPolish')" @click="emit('polish')"><Sparkles /></ToolbarButton>

      <!-- ··· 收纳不常用功能 -->
      <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <button type="button" class="tb-more ed-ctl" :title="t('editor.more')" @mousedown.prevent>
            <Dots />
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" class="min-w-44">
          <DropdownMenuItem
            class="ed-menu-item"
            :class="{ 'ed-active': active('superscript') }"
            @select="exec('superscript')"
          >
            <SuperscriptIcon class="mr-2 size-4" />
            <span>{{ t('editor.sup') }}</span>
            <span class="ml-auto pl-3 text-xs text-muted-foreground">{{ modKey }} .</span>
          </DropdownMenuItem>
          <DropdownMenuItem
            class="ed-menu-item"
            :class="{ 'ed-active': active('subscript') }"
            @select="exec('subscript')"
          >
            <SubscriptIcon class="mr-2 size-4" />
            <span>{{ t('editor.sub') }}</span>
            <span class="ml-auto pl-3 text-xs text-muted-foreground">{{ modKey }} ,</span>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem class="ed-menu-item" @select="emit('summary')">
            <FileDescription class="mr-2 size-4" />
            <span>{{ t('editor.aiSummary') }}</span>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>

    <div class="tb-sep" />

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
import type { EditorMode, SourcePaneApi } from '../types'
import { TABLE_SKELETON, type SourceCommandApi } from '../sourceCommands'
import ToolbarButton from './ToolbarButton.vue'
import ColorPicker from './ColorPicker.vue'
import EmojiPicker from './EmojiPicker.vue'
import HeadingSelect from './HeadingSelect.vue'
import FontSizeSelect from './FontSizeSelect.vue'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import {
  IconArrowBackUp as Undo2, IconArrowForwardUp as Redo2, IconBold as Bold, IconItalic as Italic,
  IconStrikethrough as Strikethrough, IconCode as Code, IconSourceCode as Code2,
  IconSubscript as SubscriptIcon, IconSuperscript as SuperscriptIcon,
  IconList as List, IconListNumbers as ListOrdered, IconListCheck as ListChecks, IconBlockquote as Quote, IconMinus as Minus,
  IconAlignLeft as AlignLeft, IconAlignCenter as AlignCenter, IconAlignRight as AlignRight,
  IconLink as LinkIcon, IconPhoto as ImageIcon, IconTable as TableIcon,
  IconSparkles as Sparkles, IconEye as Eye, IconColumns as Columns2, IconFileCode as FileCode2,
  IconDots as Dots, IconFileDescription as FileDescription,
} from '@tabler/icons-vue'

const isMac = /mac/i.test(navigator.platform)
const modKey = isMac ? '⌘' : 'Ctrl'

const props = defineProps<{
  editor: Editor | null | undefined
  mode: EditorMode
  /** 源码栏 API：mode==='source' 时命令分发到 CodeMirror 而非隐藏的 Tiptap */
  source?: SourcePaneApi | null
}>()
const emit = defineEmits<{
  link: []
  image: []
  polish: []
  summary: []
  color: [color: string | null]
  highlight: [color: string | null]
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

const currentHighlight = computed(() => {
  void tick.value
  return (props.editor?.getAttributes('highlight').color as string) || ''
})

const isSource = computed(() => props.mode === 'source')

function run(fn: (c: ChainedCommands) => ChainedCommands) {
  // 源码模式下 Tiptap 文档是陈旧的，发命令会触发 onUpdate 用旧内容覆盖 model —— 一律拦截
  if (isSource.value) return
  const e = props.editor
  if (!e) return
  fn(e.chain().focus()).run()
}

// 同一按钮的双引擎语义：rich/split 走 Tiptap 命令，source 走 CodeMirror Markdown 文本变换
const richActions: Record<string, (c: ChainedCommands) => ChainedCommands> = {
  undo: (c) => c.undo(),
  redo: (c) => c.redo(),
  bold: (c) => c.toggleBold(),
  italic: (c) => c.toggleItalic(),
  strike: (c) => c.toggleStrike(),
  code: (c) => c.toggleCode(),
  superscript: (c) => c.toggleSuperscript(),
  subscript: (c) => c.toggleSubscript(),
  bulletList: (c) => c.toggleBulletList(),
  orderedList: (c) => c.toggleOrderedList(),
  taskList: (c) => c.toggleTaskList(),
  blockquote: (c) => c.toggleBlockquote(),
  codeBlock: (c) => c.toggleCodeBlock(),
  hr: (c) => c.setHorizontalRule(),
  alignLeft: (c) => c.setTextAlign('left'),
  alignCenter: (c) => c.setTextAlign('center'),
  alignRight: (c) => c.setTextAlign('right'),
  table: (c) => c.insertTable({ rows: 3, cols: 3, withHeaderRow: true }),
}
const sourceActions: Record<string, (s: SourceCommandApi) => void> = {
  undo: (s) => s.undo(),
  redo: (s) => s.redo(),
  bold: (s) => s.wrapInline('**'),
  italic: (s) => s.wrapInline('*'),
  strike: (s) => s.wrapInline('~~'),
  code: (s) => s.wrapInline('`'),
  // 上/下标方言 ^x^ / ~x~ 与 extensions/Script.ts 对齐
  superscript: (s) => s.wrapInline('^'),
  subscript: (s) => s.wrapInline('~'),
  bulletList: (s) => s.toggleLine('bullet'),
  orderedList: (s) => s.toggleLine('ordered'),
  taskList: (s) => s.toggleLine('task'),
  blockquote: (s) => s.toggleLine('quote'),
  codeBlock: (s) => s.toggleCodeBlock(),
  hr: (s) => s.insertBlock('---'),
  table: (s) => s.insertBlock(TABLE_SKELETON),
}

function exec(name: string) {
  if (isSource.value) {
    const s = props.source
    if (s) sourceActions[name]?.(s.cmd)
    return
  }
  const fn = richActions[name]
  if (fn) run(fn)
}

function can(name: 'undo' | 'redo'): boolean {
  void tick.value
  if (isSource.value) {
    const s = props.source
    if (!s) return false
    return name === 'undo' ? s.canUndo : s.canRedo
  }
  const e = props.editor
  if (!e) return false
  return name === 'undo' ? e.can().undo() : e.can().redo()
}

function active(nameOrAttrs: string | Record<string, unknown>, attrs?: Record<string, unknown>): boolean {
  void tick.value
  // 源码模式不做光标上下文解析，按钮一律不亮
  if (isSource.value) return false
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
  justify-content: center;
  flex-wrap: wrap;
  gap: 2px;
  padding: 6px 8px;
  /* 更淡的细分隔线（原 --editor-border 太实） */
  border-bottom: 1px solid color-mix(in srgb, var(--editor-border) 45%, transparent);
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
.tb-more {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--editor-muted);
}
.tb-more :deep(svg) {
  width: 17px;
  height: 17px;
}
</style>
