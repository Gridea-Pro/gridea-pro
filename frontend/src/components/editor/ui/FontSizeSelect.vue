<script setup lang="ts">
/**
 * 字号下拉，对齐 editor-vue：候选 10-48 共 15 档（3 列网格），
 * 触发器显示当前字号（无显式字号时按标题级别推断 26/22/20/18，正文 16）。
 */
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Editor } from '@tiptap/vue-3'
import { IconChevronDown } from '@tabler/icons-vue'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
} from '@/components/ui/dropdown-menu'

const props = defineProps<{ editor: Editor | null | undefined; disabled?: boolean }>()
const { t } = useI18n()

const sizes = ['10', '12', '14', '16', '18', '20', '22', '24', '26', '28', '30', '32', '36', '40', '48']

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

const current = computed(() => {
  void tick.value
  const e = props.editor
  if (!e) return '16'
  const fs = e.getAttributes('textStyle').fontSize as string | undefined
  if (fs) return fs.replace('px', '').replace('pt', '')
  if (e.isActive('heading', { level: 1 })) return '26'
  if (e.isActive('heading', { level: 2 })) return '22'
  if (e.isActive('heading', { level: 3 })) return '20'
  if (e.isActive('heading', { level: 4 })) return '18'
  return '16'
})

function select(size: string) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ;(props.editor?.chain().focus() as any)?.setFontSize(`${size}px`).run()
}
function reset() {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ;(props.editor?.chain().focus() as any)?.unsetFontSize().run()
}
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <button
        type="button"
        class="inline-flex h-8 items-center gap-1 rounded-md px-2 text-sm tabular-nums text-foreground ed-ctl disabled:cursor-not-allowed disabled:opacity-40"
        :title="t('editor.fontSize.title')"
        :disabled="disabled"
        @mousedown.prevent
      >
        <span class="min-w-5 text-center">{{ current }}</span>
        <IconChevronDown class="h-3 w-3 shrink-0 opacity-60" />
      </button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="start" class="w-36 p-1.5">
      <button
        type="button"
        class="mb-1 w-full rounded px-2 py-1 text-left text-sm ed-ctl"
        @click="reset"
      >
        {{ t('editor.fontSize.default') }}
      </button>
      <div class="grid grid-cols-3 gap-0.5">
        <button
          v-for="s in sizes"
          :key="s"
          type="button"
          class="rounded px-0 py-1 text-center text-[13px] tabular-nums ed-ctl"
          :class="{ 'ed-active font-medium': current === s }"
          @click="select(s)"
        >
          {{ s }}
        </button>
      </div>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
