<script setup lang="ts">
/**
 * 标题下拉（正文 / 标题1-6），对齐 editor-vue：项左侧 T/H1-H6 标记，右侧 ⌘⌥0-6 快捷键标注。
 * 快捷键本身由 TipTap Paragraph(Mod-Alt-0)/Heading(Mod-Alt-1..6) 内置注册，无需自定义。
 */
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Editor } from '@tiptap/vue-3'
import { IconChevronDown } from '@tabler/icons-vue'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/dropdown-menu'

const props = defineProps<{
  editor: Editor | null | undefined
  /** 源码模式回调：非空时选择项走 Markdown 行前缀（0=正文，1-6=# 级别），不再碰 Tiptap */
  sourceHeading?: ((level: number) => void) | null
}>()
const { t } = useI18n()

const isMac = /mac/i.test(navigator.platform)
const modAlt = isMac ? '⌘⌥' : 'Ctrl+Alt+'

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

interface HeadingOption {
  id: 'paragraph' | '1' | '2' | '3' | '4' | '5' | '6'
  icon: string
  label: () => string
  shortcut: string
}
const options: HeadingOption[] = [
  { id: 'paragraph', icon: 'T', label: () => t('editor.heading.paragraph'), shortcut: `${modAlt}0` },
  { id: '1', icon: 'H1', label: () => t('editor.heading.level', { n: 1 }), shortcut: `${modAlt}1` },
  { id: '2', icon: 'H2', label: () => t('editor.heading.level', { n: 2 }), shortcut: `${modAlt}2` },
  { id: '3', icon: 'H3', label: () => t('editor.heading.level', { n: 3 }), shortcut: `${modAlt}3` },
  { id: '4', icon: 'H4', label: () => t('editor.heading.level', { n: 4 }), shortcut: `${modAlt}4` },
  { id: '5', icon: 'H5', label: () => t('editor.heading.level', { n: 5 }), shortcut: `${modAlt}5` },
  { id: '6', icon: 'H6', label: () => t('editor.heading.level', { n: 6 }), shortcut: `${modAlt}6` },
]

const currentId = computed(() => {
  void tick.value
  // 源码模式不解析光标上下文，触发器恒显"正文"（Tiptap 的选区此时是陈旧的）
  if (props.sourceHeading) return 'paragraph'
  const e = props.editor
  if (!e) return 'paragraph'
  for (let i = 1; i <= 6; i++) {
    if (e.isActive('heading', { level: i })) return String(i)
  }
  return 'paragraph'
})

const currentLabel = computed(() => {
  const opt = options.find((o) => o.id === currentId.value)
  return opt ? opt.label() : t('editor.heading.paragraph')
})

function select(id: HeadingOption['id']) {
  if (props.sourceHeading) {
    props.sourceHeading(id === 'paragraph' ? 0 : parseInt(id, 10))
    return
  }
  const e = props.editor
  if (!e) return
  if (id === 'paragraph') {
    e.chain().focus().setParagraph().run()
  } else {
    e.chain().focus().toggleHeading({ level: parseInt(id, 10) as 1 | 2 | 3 | 4 | 5 | 6 }).run()
  }
}
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <button
        type="button"
        class="inline-flex h-8 min-w-16 items-center justify-between gap-1 rounded-md px-2 text-sm text-foreground ed-ctl"
        :title="t('editor.heading.title')"
        @mousedown.prevent
      >
        <span class="truncate">{{ currentLabel }}</span>
        <IconChevronDown class="h-3 w-3 shrink-0 opacity-60" />
      </button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="start" class="min-w-44">
      <DropdownMenuItem
        class="ed-menu-item"
        v-for="opt in options"
        :key="opt.id"
        :class="{ 'ed-active': currentId === opt.id }"
        @select="select(opt.id)"
      >
        <span class="mr-2 w-6 text-xs font-bold text-muted-foreground">{{ opt.icon }}</span>
        <span>{{ opt.label() }}</span>
        <span class="ml-auto pl-4 text-xs text-muted-foreground">{{ opt.shortcut }}</span>
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
