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
      <HeadingSelect :editor="editor" />
      <FontSizeSelect :editor="editor" />
    </div>

    <div class="tb-sep" />

    <div class="tb-group">
      <ToolbarButton :title="t('editor.bold')" :active="active('bold')" @click="run((c) => c.toggleBold())"><Bold /></ToolbarButton>
      <ToolbarButton :title="t('editor.italic')" :active="active('italic')" @click="run((c) => c.toggleItalic())"><Italic /></ToolbarButton>
      <ToolbarButton :title="t('editor.strike')" :active="active('strike')" @click="run((c) => c.toggleStrike())"><Strikethrough /></ToolbarButton>
      <ToolbarButton :title="t('editor.code')" :active="active('code')" @click="run((c) => c.toggleCode())"><Code /></ToolbarButton>
      <ColorPicker type="text" :model-value="currentColor" @select="emit('color', $event)" />
      <ColorPicker type="highlight" :model-value="currentHighlight" @select="emit('highlight', $event)" />
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
      <ToolbarButton :title="t('editor.shortcuts.alignLeft')" :active="active({ textAlign: 'left' })"
        @click="run((c) => c.setTextAlign('left'))"><AlignLeft /></ToolbarButton>
      <ToolbarButton :title="t('editor.shortcuts.alignCenter')" :active="active({ textAlign: 'center' })"
        @click="run((c) => c.setTextAlign('center'))"><AlignCenter /></ToolbarButton>
      <ToolbarButton :title="t('editor.shortcuts.alignRight')" :active="active({ textAlign: 'right' })"
        @click="run((c) => c.setTextAlign('right'))"><AlignRight /></ToolbarButton>
    </div>

    <div class="tb-sep" />

    <div class="tb-group">
      <ToolbarButton :title="t('editor.link')" :active="active('link')" @click="emit('link')"><LinkIcon /></ToolbarButton>
      <ToolbarButton :title="t('editor.image')" @click="emit('image')"><ImageIcon /></ToolbarButton>
      <ToolbarButton :title="t('editor.table')" @click="run((c) => c.insertTable({ rows: 3, cols: 3, withHeaderRow: true }))"><TableIcon /></ToolbarButton>
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
            @select="run((c) => c.toggleSuperscript())"
          >
            <SuperscriptIcon class="mr-2 size-4" />
            <span>{{ t('editor.sup') }}</span>
            <span class="ml-auto pl-3 text-xs text-muted-foreground">{{ modKey }} .</span>
          </DropdownMenuItem>
          <DropdownMenuItem
            class="ed-menu-item"
            :class="{ 'ed-active': active('subscript') }"
            @select="run((c) => c.toggleSubscript())"
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
import type { EditorMode } from '../types'
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

const props = defineProps<{ editor: Editor | null | undefined; mode: EditorMode }>()
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
