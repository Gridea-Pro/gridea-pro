<template>
  <BubbleMenu
    v-if="editor"
    :editor="editor"
    :should-show="shouldShow"
  >
    <div
      class="flex items-center gap-0.5 p-1 bg-popover text-popover-foreground border border-border rounded-md shadow-md"
    >
      <button
        type="button"
        :class="btnClass('bold')"
        :title="t('editor.bold')"
        @mousedown.prevent
        @click="run((c) => c.toggleBold())"
      >
        <Bold />
      </button>
      <button
        type="button"
        :class="btnClass('italic')"
        :title="t('editor.italic')"
        @mousedown.prevent
        @click="run((c) => c.toggleItalic())"
      >
        <Italic />
      </button>
      <button
        type="button"
        :class="btnClass('strike')"
        :title="t('editor.strike')"
        @mousedown.prevent
        @click="run((c) => c.toggleStrike())"
      >
        <Strikethrough />
      </button>
      <button
        type="button"
        :class="btnClass('code')"
        :title="t('editor.code')"
        @mousedown.prevent
        @click="run((c) => c.toggleCode())"
      >
        <Code />
      </button>
      <button
        type="button"
        :class="btnClass('highlight')"
        :title="t('editor.highlight')"
        @mousedown.prevent
        @click="run((c) => c.toggleHighlight())"
      >
        <Highlighter />
      </button>

      <div class="w-px h-[18px] mx-0.5 bg-border" />

      <button
        type="button"
        :class="btnClass('link')"
        :title="t('editor.link')"
        @mousedown.prevent
        @click="emit('link')"
      >
        <LinkIcon />
      </button>
      <button
        type="button"
        :class="baseBtn"
        :title="t('editor.aiPolish')"
        @mousedown.prevent
        @click="emit('polish')"
      >
        <Sparkles />
      </button>
    </div>
  </BubbleMenu>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Editor, ChainedCommands } from '@tiptap/vue-3'
import { BubbleMenu } from '@tiptap/vue-3/menus'
import type { EditorState } from '@tiptap/pm/state'
import {
  Bold,
  Italic,
  Strikethrough,
  Code,
  Highlighter,
  Link as LinkIcon,
  Sparkles,
} from 'lucide-vue-next'

const props = defineProps<{ editor: Editor | null }>()
const emit = defineEmits<{
  link: []
  polish: []
}>()

const { t } = useI18n()

// 让 isActive 随选区实时刷新。editor 在 onMounted 时可能仍为 undefined（useEditor 异步就绪），
// 用 watch 待其出现再绑定，否则监听器永不注册、高亮态不更新。
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

// tiptap 的 shouldShow 回调参数形态较细，此处用宽松类型
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function shouldShow(p: any): boolean {
  const state = p.state as EditorState
  const ed = p.editor as Editor
  const view = p.view as { hasFocus?: () => boolean } | undefined
  if (view?.hasFocus && !view.hasFocus()) return false // 失焦（如切到源码栏）时隐藏
  return ed.isEditable && state.selection.from !== state.selection.to && !ed.isActive('codeBlock')
}

const baseBtn =
  'inline-flex items-center justify-center w-7 h-7 rounded-md cursor-pointer transition-colors hover:bg-accent hover:text-accent-foreground'

function isActive(name: string): boolean {
  // 读取 tick 触发依赖，保证选区变化时模板重算
  void tick.value
  return props.editor?.isActive(name) ?? false
}

function btnClass(name: string): string {
  return isActive(name) ? `${baseBtn} text-primary bg-accent` : baseBtn
}

function run(fn: (c: ChainedCommands) => ChainedCommands): void {
  const e = props.editor
  if (!e) return
  fn(e.chain().focus()).run()
}
</script>

<style scoped>
button {
  border: none;
  background: transparent;
}
button :deep(svg) {
  width: 17px;
  height: 17px;
}
</style>
