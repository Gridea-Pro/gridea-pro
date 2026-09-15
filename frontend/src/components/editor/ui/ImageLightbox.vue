<script setup lang="ts">
/**
 * 图片预览（lightbox）：遮罩 + 居中原图，滚轮缩放，点击遮罩/Esc 关闭。
 * 渲染在编辑器子树内（--editor-* 变量可解析），fixed 全屏覆盖。
 */
import { ref, watch, onBeforeUnmount } from 'vue'

const props = defineProps<{ src: string | null }>()
const emit = defineEmits<{ close: [] }>()

const scale = ref(1)

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}
watch(
  () => props.src,
  (s) => {
    scale.value = 1
    if (s) window.addEventListener('keydown', onKey)
    else window.removeEventListener('keydown', onKey)
  },
)
onBeforeUnmount(() => window.removeEventListener('keydown', onKey))

function onWheel(e: WheelEvent) {
  e.preventDefault()
  const next = scale.value * (e.deltaY < 0 ? 1.1 : 0.9)
  scale.value = Math.min(6, Math.max(0.2, next))
}
</script>

<template>
  <div v-if="src" class="lightbox-overlay" @click.self="emit('close')" @wheel="onWheel">
    <img class="lightbox-img" :src="src" :style="{ transform: `scale(${scale})` }" @dblclick="scale = 1" />
    <button type="button" class="lightbox-close" @click="emit('close')">×</button>
    <span class="lightbox-zoom">{{ Math.round(scale * 100) }}%</span>
  </div>
</template>

<style scoped>
.lightbox-overlay {
  position: fixed;
  inset: 0;
  z-index: 70;
  background: rgba(0, 0, 0, 0.82);
  display: flex;
  align-items: center;
  justify-content: center;
  overscroll-behavior: contain;
}
.lightbox-img {
  max-width: 92vw;
  max-height: 88vh;
  border-radius: 6px;
  transition: transform 0.08s ease-out;
  user-select: none;
  -webkit-user-drag: none;
}
.lightbox-close {
  position: absolute;
  top: 14px;
  right: 18px;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
  font-size: 22px;
  line-height: 1;
  cursor: pointer;
}
.lightbox-close:hover {
  background: rgba(255, 255, 255, 0.26);
}
.lightbox-zoom {
  position: absolute;
  bottom: 16px;
  left: 50%;
  transform: translateX(-50%);
  color: rgba(255, 255, 255, 0.85);
  font-size: 12px;
  background: rgba(0, 0, 0, 0.45);
  padding: 2px 10px;
  border-radius: 999px;
}
</style>
