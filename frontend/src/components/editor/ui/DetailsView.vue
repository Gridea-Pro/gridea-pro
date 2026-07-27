<template>
  <NodeViewWrapper as="div" class="details-node" :class="{ 'is-open': open, 'is-selected': selected }">
    <div class="details-summary-row" contenteditable="false">
      <button type="button" class="details-toggle" :aria-expanded="open" @click="toggleOpen">
        <ChevronRight class="size-4 chevron" />
      </button>
      <input
        ref="summaryInput"
        class="details-summary-input"
        :value="summary"
        :placeholder="t('editor.details.summary')"
        :readonly="!editor?.isEditable"
        @input="onSummaryInput"
        @keydown.enter.prevent="focusBody"
        @mousedown.stop
      />
    </div>
    <NodeViewContent v-show="open" class="details-body" />
  </NodeViewWrapper>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { NodeViewWrapper, NodeViewContent, nodeViewProps } from '@tiptap/vue-3'
import { ChevronRight } from 'lucide-vue-next'

const props = defineProps(nodeViewProps)
const { t } = useI18n()

const summary = computed<string>(() => (props.node.attrs.summary as string) || '')
const open = computed<boolean>(() => props.node.attrs.open !== false)
const summaryInput = ref<HTMLInputElement | null>(null)

function onSummaryInput(e: Event) {
  props.updateAttributes({ summary: (e.target as HTMLInputElement).value })
}
function toggleOpen() {
  props.updateAttributes({ open: !open.value })
}
function focusBody() {
  // 回车从摘要跳到正文：把光标移到节点内容起始
  if (typeof props.getPos === 'function') {
    const pos = props.getPos()
    if (typeof pos === 'number') {
      props.editor.chain().focus(pos + 1).run()
    }
  }
}
</script>

<style scoped>
.details-node {
  margin: 0.6em 0;
  border: 1px solid var(--editor-border);
  border-radius: var(--editor-radius);
  background: var(--editor-secondary);
  overflow: hidden;
}
.details-node.is-selected {
  outline: 2px solid var(--editor-accent);
  outline-offset: -1px;
}
.details-summary-row {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 10px;
  user-select: none;
}
.details-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  flex-shrink: 0;
  border: none;
  border-radius: 5px;
  background: transparent;
  color: var(--editor-muted);
  cursor: pointer;
}
.details-toggle:hover {
  background: var(--editor-bg);
  color: var(--editor-fg);
}
.chevron {
  transition: transform 0.15s ease;
}
.details-node.is-open .chevron {
  transform: rotate(90deg);
}
.details-summary-input {
  flex: 1;
  min-width: 0;
  border: none;
  outline: none;
  background: transparent;
  color: var(--editor-fg);
  font-weight: 600;
  font-size: 0.95em;
}
.details-body {
  padding: 2px 14px 10px 36px;
  background: var(--editor-bg);
}
</style>
