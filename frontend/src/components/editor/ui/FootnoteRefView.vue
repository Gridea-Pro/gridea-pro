<template>
  <NodeViewWrapper
    as="sup"
    class="footnote-ref"
    :class="{ 'is-selected': selected }"
    :data-fn-ref="id"
    @click="jumpToDef"
    @mouseenter="hovering = true"
    @mouseleave="hovering = false"
  >
    [{{ num || id }}]
    <span v-if="hovering && defText" class="footnote-ref-tip" contenteditable="false">{{ defText }}</span>
  </NodeViewWrapper>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { NodeViewWrapper, nodeViewProps } from '@tiptap/vue-3'
import { footnoteNumber, footnoteDefText } from './footnoteNumber'

const props = defineProps(nodeViewProps)
const id = computed<string>(() => (props.node.attrs.id as string) || '')
const hovering = ref(false)

// 用 tick 触发重算：其它脚注增删会改变编号，故订阅 editor update
const tick = ref(0)
const num = computed(() => {
  void tick.value
  return footnoteNumber(props.editor, id.value)
})
const defText = computed(() => {
  void tick.value
  return footnoteDefText(props.editor, id.value)
})

function bump() {
  tick.value++
}
function jumpToDef() {
  const el = document.getElementById(`fn-def-${id.value}`)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    el.classList.add('footnote-flash')
    setTimeout(() => el.classList.remove('footnote-flash'), 1200)
  }
}

onMounted(() => props.editor?.on('update', bump))
onBeforeUnmount(() => props.editor?.off('update', bump))
</script>

<style scoped>
.footnote-ref {
  position: relative;
  cursor: pointer;
  color: var(--editor-accent);
  font-weight: 500;
  padding: 0 1px;
}
.footnote-ref.is-selected {
  outline: 2px solid var(--editor-accent);
  border-radius: 3px;
}
.footnote-ref-tip {
  position: absolute;
  left: 50%;
  bottom: 130%;
  transform: translateX(-50%);
  z-index: 30;
  max-width: 280px;
  width: max-content;
  padding: 6px 10px;
  background: var(--editor-fg);
  color: var(--editor-bg);
  border-radius: 6px;
  font-size: 12px;
  font-weight: 400;
  line-height: 1.5;
  white-space: normal;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.2);
  pointer-events: none;
}
</style>
