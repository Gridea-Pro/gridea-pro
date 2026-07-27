<template>
  <NodeViewWrapper
    as="div"
    class="footnote-def"
    :class="{ 'is-selected': selected }"
    :id="`fn-def-${id}`"
    :data-footnote-id="id"
  >
    <span class="footnote-def-marker">{{ num || id }}.</span>
    <textarea
      ref="ta"
      class="footnote-def-text"
      :value="draft"
      rows="1"
      spellcheck="false"
      :placeholder="t('editor.footnote.placeholder')"
      :readonly="!editor?.isEditable"
      @input="onInput"
      @keydown="onKeydown"
    />
    <button
      type="button"
      class="footnote-def-back"
      :title="t('editor.footnote.back')"
      @click="jumpToRef"
    >
      <CornerLeftUp class="size-3.5" />
    </button>
  </NodeViewWrapper>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { NodeViewWrapper, nodeViewProps } from '@tiptap/vue-3'
import { CornerLeftUp } from 'lucide-vue-next'
import { footnoteNumber } from './footnoteNumber'

const props = defineProps(nodeViewProps)
const { t } = useI18n()
const id = computed<string>(() => (props.node.attrs.id as string) || '')
const ta = ref<HTMLTextAreaElement | null>(null)

const tick = ref(0)
const num = computed(() => {
  void tick.value
  return footnoteNumber(props.editor, id.value)
})

// draft 与 attr 解耦，避免每次按键 updateAttributes → 节点重渲染重置光标
const draft = ref<string>((props.node.attrs.text as string) || '')

function resize() {
  const el = ta.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = `${el.scrollHeight}px`
}
function onInput(e: Event) {
  draft.value = (e.target as HTMLTextAreaElement).value
  props.updateAttributes({ text: draft.value })
  resize()
}
function onKeydown(e: KeyboardEvent) {
  // 把正常输入键（含 Enter/Backspace）隔离在 textarea 内，避免 ProseMirror 抢去删节点/换块；
  // 但放行 Escape，让其冒泡触发编辑器层的退出/失焦
  if (e.key !== 'Escape') e.stopPropagation()
}
function jumpToRef() {
  // ref 用 data 属性而非 id（同一脚注可被多处引用，id 会重复违反 HTML 规范）；
  // 跳到第一处引用即可（标准脚注行为）
  const el = document.querySelector<HTMLElement>(`[data-fn-ref="${CSS.escape(id.value)}"]`)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    el.classList.add('footnote-flash')
    setTimeout(() => el.classList.remove('footnote-flash'), 1200)
  }
}

// 外部（如源码模式回灌）改了 text 且当前未聚焦 → 同步 draft
watch(
  () => props.node.attrs.text as string,
  (v) => {
    if (document.activeElement !== ta.value && v !== draft.value) {
      draft.value = v || ''
      nextTick(resize)
    }
  },
)

function bump() {
  tick.value++
}
onMounted(() => {
  props.editor?.on('update', bump)
  nextTick(resize)
})
onBeforeUnmount(() => props.editor?.off('update', bump))
</script>

<style scoped>
.footnote-def {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 4px 8px;
  margin: 2px 0;
  border-left: 2px solid var(--editor-border);
  font-size: 0.9em;
  color: var(--editor-muted);
}
.footnote-def.is-selected {
  border-left-color: var(--editor-accent);
  background: var(--editor-secondary);
}
.footnote-def-marker {
  flex-shrink: 0;
  font-variant-numeric: tabular-nums;
  color: var(--editor-accent);
  font-weight: 500;
  padding-top: 4px;
}
.footnote-def-text {
  flex: 1;
  min-width: 0;
  border: none;
  outline: none;
  resize: none;
  overflow: hidden;
  background: transparent;
  color: var(--editor-fg);
  font: inherit;
  line-height: 1.6;
  padding: 4px 0;
}
.footnote-def-back {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--editor-muted);
  cursor: pointer;
}
.footnote-def-back:hover {
  background: var(--editor-secondary);
  color: var(--editor-accent);
}
</style>
