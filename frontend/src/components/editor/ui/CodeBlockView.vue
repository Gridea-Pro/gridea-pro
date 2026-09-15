<template>
  <NodeViewWrapper class="code-block-node" :class="{ 'is-selected': selected }">
    <div class="code-block-header" contenteditable="false">
      <select
        class="code-lang-select"
        :value="language"
        :disabled="!editor?.isEditable"
        @change="onLangChange"
      >
        <option value="">{{ t('editor.codeBlockView.plain') }}</option>
        <option v-for="l in languages" :key="l" :value="l">{{ l }}</option>
      </select>
      <button
        type="button"
        class="code-copy-btn"
        :title="t('editor.codeBlockView.copy')"
        @click="copy"
      >
        <Check v-if="copied" class="size-3.5" />
        <Copy v-else class="size-3.5" />
        <span>{{ copied ? t('editor.codeBlockView.copied') : t('editor.codeBlockView.copy') }}</span>
      </button>
    </div>
    <pre><NodeViewContent as="code" :class="language ? `language-${language}` : ''" /></pre>
  </NodeViewWrapper>
</template>

<script setup lang="ts">
import { computed, ref, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { NodeViewWrapper, NodeViewContent, nodeViewProps } from '@tiptap/vue-3'
import { IconCopy as Copy, IconCheck as Check } from '@tabler/icons-vue'
import { CODE_LANGUAGES } from '../extensions/CodeBlock'

const props = defineProps(nodeViewProps)
const { t } = useI18n()

const language = computed<string>(() => (props.node.attrs.language as string) || '')

// 节点当前语言若不在候选列表里，补进去（保证下拉能显示自定义语言）
const languages = computed(() => {
  const l = language.value
  if (l && !CODE_LANGUAGES.includes(l)) return [l, ...CODE_LANGUAGES]
  return CODE_LANGUAGES
})

function onLangChange(e: Event) {
  const value = (e.target as HTMLSelectElement).value
  props.updateAttributes({ language: value || null })
}

const copied = ref(false)
let copyTimer: ReturnType<typeof setTimeout> | null = null

async function copy() {
  const text = props.node.textContent || ''
  try {
    await navigator.clipboard.writeText(text)
  } catch {
    // 回退：execCommand（部分 WebView 无 clipboard API）
    const ta = document.createElement('textarea')
    ta.value = text
    ta.style.position = 'fixed'
    ta.style.opacity = '0'
    document.body.appendChild(ta)
    ta.select()
    try {
      document.execCommand('copy')
    } catch {
      /* ignore */
    }
    ta.remove()
  }
  copied.value = true
  if (copyTimer) clearTimeout(copyTimer)
  copyTimer = setTimeout(() => (copied.value = false), 1800)
}

onBeforeUnmount(() => {
  if (copyTimer) clearTimeout(copyTimer)
})
</script>

<style scoped>
.code-block-node {
  position: relative;
  margin: 0.6em 0;
}
.code-block-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 4px 8px 4px 10px;
  background: var(--editor-secondary);
  border: 1px solid var(--editor-border);
  border-bottom: none;
  border-radius: var(--editor-radius) var(--editor-radius) 0 0;
  user-select: none;
}
.code-lang-select {
  appearance: none;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 6px;
  padding: 2px 8px;
  font-family: var(--editor-mono);
  font-size: 12px;
  color: var(--editor-muted);
  cursor: pointer;
}
.code-lang-select:hover:not(:disabled) {
  border-color: var(--editor-border);
  color: var(--editor-fg);
}
.code-lang-select:disabled {
  cursor: default;
}
.code-copy-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  border: 1px solid var(--editor-border);
  border-radius: 6px;
  background: var(--editor-bg);
  color: var(--editor-muted);
  font-size: 12px;
  cursor: pointer;
  transition: color 0.12s, border-color 0.12s;
}
.code-copy-btn:hover {
  color: var(--editor-fg);
  border-color: var(--editor-accent);
}
.code-block-node :deep(pre) {
  margin: 0;
  border: 1px solid var(--editor-border);
  border-radius: 0 0 var(--editor-radius) var(--editor-radius);
}
.code-block-node.is-selected :deep(pre) {
  outline: 2px solid var(--editor-accent);
  outline-offset: -1px;
}
</style>
