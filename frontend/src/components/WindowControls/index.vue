<template>
  <div v-if="showControls" class="traffic-lights" style="--wails-draggable: no-drag"
    @mouseenter="hovered = true" @mouseleave="hovered = false">
    <!-- 关闭 -->
    <button class="light light-close" @click="close" :title="t('window.close')">
      <svg v-if="hovered" width="8" height="8" viewBox="0 0 8 8">
        <line stroke="currentColor" stroke-width="1.2" x1="1" y1="1" x2="7" y2="7" />
        <line stroke="currentColor" stroke-width="1.2" x1="7" y1="1" x2="1" y2="7" />
      </svg>
    </button>
    <!-- 最小化 -->
    <button class="light light-minimize" @click="minimize" :title="t('window.minimize')">
      <svg v-if="hovered" width="8" height="2" viewBox="0 0 8 2">
        <line stroke="currentColor" stroke-width="1.5" x1="1" y1="1" x2="7" y2="1" />
      </svg>
    </button>
    <!-- 最大化/还原 -->
    <button class="light light-maximize" @click="toggleMaximize" :title="t('window.zoom')">
      <svg v-if="hovered && !isMaximized" width="8" height="8" viewBox="0 0 8 8">
        <polygon fill="currentColor" points="1,1 1,7 7,7 7,1" stroke="currentColor" stroke-width="0.5" fill-opacity="0" />
      </svg>
      <svg v-if="hovered && isMaximized" width="8" height="8" viewBox="0 0 10 10">
        <polygon fill="none" stroke="currentColor" stroke-width="1.2" points="3,1 9,1 9,7 3,7" />
        <polygon fill="none" stroke="currentColor" stroke-width="1.2" points="1,3 7,3 7,9 1,9" />
      </svg>
    </button>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  WindowMinimise,
  WindowToggleMaximise,
  Quit,
  EventsOn,
} from '@/wailsjs/runtime'

const { t } = useI18n()
const showControls = ref(false)
const isMaximized = ref(false)
const hovered = ref(false)

onMounted(async () => {
  showControls.value = false

  EventsOn('wails:window-maximised', () => {
    isMaximized.value = true
  })
  EventsOn('wails:window-restored', () => {
    isMaximized.value = false
  })
})

const minimize = () => WindowMinimise()
const toggleMaximize = () => WindowToggleMaximise()
const close = () => Quit()
</script>

<style scoped>
.traffic-lights {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-left: 14px;
  -webkit-app-region: no-drag;
}

.light {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  transition: opacity 0.15s ease;
  color: rgba(0, 0, 0, 0.5);
  outline: none;
}

.light-close {
  background-color: #ff5f57;
}

.light-close:hover {
  background-color: #e5453d;
}

.light-minimize {
  background-color: #febc2e;
}

.light-minimize:hover {
  background-color: #e5a81f;
}

.light-maximize {
  background-color: #28c840;
}

.light-maximize:hover {
  background-color: #1fad32;
}

/* 窗口未聚焦时变灰（可选，未来可监听 focus 事件） */
.light:active {
  opacity: 0.7;
}
</style>
