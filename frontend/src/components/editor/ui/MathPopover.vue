<template>
  <div
    v-if="open"
    ref="root"
    class="math-popover"
    :style="{ left: `${x}px`, top: `${y}px` }"
    @keydown.stop
  >
    <div class="math-popover-preview" v-html="preview" />
    <textarea
      ref="ta"
      v-model="draft"
      class="math-popover-input"
      spellcheck="false"
      :placeholder="kind === 'block' ? '\\int_0^1 x\\,dx' : 'E = mc^2'"
      @keydown.escape.prevent.stop="cancel"
      @keydown.enter.exact.prevent="save"
      @keydown.meta.enter.prevent="save"
      @keydown.ctrl.enter.prevent="save"
    />
    <div class="math-popover-actions">
      <button type="button" class="mp-btn" @click="cancel">{{ t('editor.math.cancel') }}</button>
      <button type="button" class="mp-btn mp-primary" @click="save">{{ t('editor.math.save') }}</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, computed, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import katex from 'katex'

const props = defineProps<{
  open: boolean
  kind: 'inline' | 'block'
  latex: string
  x: number
  y: number
}>()
const emit = defineEmits<{ save: [latex: string]; cancel: [] }>()

const { t } = useI18n()
const root = ref<HTMLElement | null>(null)
const ta = ref<HTMLTextAreaElement | null>(null)
const draft = ref('')

const preview = computed(() => {
  const src = draft.value.trim()
  if (!src) return `<span class="math-popover-empty">${t('editor.math.empty')}</span>`
  try {
    return katex.renderToString(src, { throwOnError: false, displayMode: props.kind === 'block' })
  } catch (e) {
    return `<span class="math-popover-err">${(e as Error).message}</span>`
  }
})

function save() {
  emit('save', draft.value.trim())
}
function cancel() {
  emit('cancel')
}

function onDocPointer(e: MouseEvent) {
  if (root.value && !root.value.contains(e.target as Node)) cancel()
}

watch(
  () => props.open,
  (o) => {
    if (o) {
      draft.value = props.latex
      nextTick(() => {
        ta.value?.focus()
        ta.value?.select()
        // 延迟挂监听，避免触发本次打开的同一次点击
        setTimeout(() => document.addEventListener('mousedown', onDocPointer), 0)
      })
    } else {
      document.removeEventListener('mousedown', onDocPointer)
    }
  },
)

onBeforeUnmount(() => document.removeEventListener('mousedown', onDocPointer))
</script>

<style scoped>
.math-popover {
  position: fixed;
  z-index: 50;
  width: 320px;
  max-width: 90vw;
  background: var(--editor-bg);
  border: 1px solid var(--editor-border);
  border-radius: var(--editor-radius);
  box-shadow: 0 8px 28px rgba(0, 0, 0, 0.18);
  padding: 10px;
}
.math-popover-preview {
  min-height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px;
  margin-bottom: 8px;
  background: var(--editor-secondary);
  border-radius: 6px;
  overflow-x: auto;
}
.math-popover-input {
  width: 100%;
  min-height: 64px;
  padding: 8px 10px;
  border: 1px solid var(--editor-border);
  border-radius: 6px;
  outline: none;
  resize: vertical;
  background: var(--editor-code-bg);
  color: var(--editor-fg);
  font-family: var(--editor-mono);
  font-size: 13px;
  line-height: 1.6;
}
.math-popover-input:focus {
  border-color: var(--editor-accent);
}
.math-popover-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 8px;
}
.mp-btn {
  padding: 4px 12px;
  border: 1px solid var(--editor-border);
  border-radius: 6px;
  background: var(--editor-bg);
  color: var(--editor-fg);
  font-size: 13px;
  cursor: pointer;
}
.mp-primary {
  background: var(--editor-accent);
  color: var(--editor-accent-fg);
  border-color: var(--editor-accent);
}
.math-popover-preview :deep(.math-popover-empty),
.math-popover-preview :deep(.math-popover-err) {
  color: var(--editor-muted);
  font-size: 12px;
}
</style>
