<template>
  <NodeViewWrapper
    as="div"
    class="image-view"
    :class="{ 'is-selected': selected }"
    :style="{ textAlign: node.attrs.textAlign || undefined }"
  >
    <div class="image-box" :style="boxStyle">
      <img :src="node.attrs.src" :alt="node.attrs.alt || ''" :title="node.attrs.title || ''" draggable="false" />
      <!-- 透明罩层：拦截 WebKit Live Text（点击图片时 OCR 选中图中文字），点击即选中节点，双击预览 -->
      <div class="image-shield" @mousedown.prevent="selectSelf" @dblclick.prevent="previewSelf" />
      <!-- 选中时四角拖拽手柄 -->
      <template v-if="selected && editor?.isEditable">
        <span
          v-for="h in handles"
          :key="h"
          class="image-handle"
          :class="`handle-${h}`"
          @mousedown.prevent.stop="startResize($event, h)"
        />
      </template>
    </div>
  </NodeViewWrapper>
</template>

<script setup lang="ts">
import { ref, computed, onBeforeUnmount } from 'vue'
import { NodeViewWrapper, nodeViewProps } from '@tiptap/vue-3'
import { dispatchEditorAction } from '../extensions/slash/data'

const props = defineProps(nodeViewProps)

function previewSelf() {
  selectSelf()
  // 经编辑器 DOM 事件通知 index.vue 打开预览（与 link/image 弹窗同机制）
  dispatchEditorAction(props.editor, 'image-preview')
}

const handles = ['nw', 'ne', 'sw', 'se'] as const
type Handle = (typeof handles)[number]

// 拖拽中的实时宽度（本地态，松手才提交 transaction，避免每像素一个 undo 步）
const liveWidth = ref<number | null>(null)

const boxStyle = computed(() => {
  const w = liveWidth.value ?? (props.node.attrs.width as number | null)
  return w ? { width: `${w}px` } : {}
})

function selectSelf() {
  if (typeof props.getPos === 'function') {
    const pos = props.getPos()
    if (typeof pos === 'number') props.editor.commands.setNodeSelection(pos)
  }
}

let startX = 0
let startW = 0
let activeHandle: Handle = 'se'

function startResize(e: MouseEvent, h: Handle) {
  selectSelf()
  activeHandle = h
  startX = e.clientX
  const box = (e.target as HTMLElement).closest('.image-box') as HTMLElement | null
  startW = box?.getBoundingClientRect().width || 0
  window.addEventListener('mousemove', onMove)
  window.addEventListener('mouseup', onUp)
}
function onMove(e: MouseEvent) {
  const dx = e.clientX - startX
  // 左侧手柄向左拖为放大
  const delta = activeHandle === 'nw' || activeHandle === 'sw' ? -dx : dx
  liveWidth.value = Math.max(80, Math.round(startW + delta))
}
function onUp() {
  window.removeEventListener('mousemove', onMove)
  window.removeEventListener('mouseup', onUp)
  if (liveWidth.value) {
    props.updateAttributes({ width: liveWidth.value })
    liveWidth.value = null
  }
}
onBeforeUnmount(() => {
  window.removeEventListener('mousemove', onMove)
  window.removeEventListener('mouseup', onUp)
})
</script>

<style scoped>
.image-view {
  margin: 0.6em 0;
  line-height: 0;
}
.image-box {
  position: relative;
  display: inline-block;
  max-width: 100%;
}
.image-box img {
  display: block;
  width: 100%;
  height: auto;
  border-radius: var(--editor-radius);
}
.image-shield {
  position: absolute;
  inset: 0;
  /* 透明罩层只为拦截 Live Text 与原生选中，不做视觉 */
  background: transparent;
}
.image-view.is-selected .image-box {
  outline: 2px solid var(--editor-accent);
  outline-offset: 1px;
  border-radius: var(--editor-radius);
}
.image-handle {
  position: absolute;
  width: 10px;
  height: 10px;
  background: var(--editor-bg);
  border: 1.5px solid var(--editor-accent);
  border-radius: 50%;
  z-index: 2;
}
.handle-nw { top: -5px; left: -5px; cursor: nwse-resize; }
.handle-ne { top: -5px; right: -5px; cursor: nesw-resize; }
.handle-sw { bottom: -5px; left: -5px; cursor: nesw-resize; }
.handle-se { bottom: -5px; right: -5px; cursor: nwse-resize; }
</style>
