<template>
  <div id="app">
    <Toaster
position="top-center" :expand="false" rich-colors :duration="2000" :close-button="false" :toast-options="{
      unstyled: false,
      classes: {
        toast: 'transform scale-75 !min-w-[160px] !max-w-[280px] !px-4 !py-2.5 !gap-2 !text-xs !shadow-lg !rounded-full !flex !flex-row !items-center',
        title: '!text-sm !font-medium',
        description: '!text-[11px]',
      }
    }" />



    <div v-if="error" class="fixed inset-0 z-[99999] overflow-auto bg-background p-5 text-destructive whitespace-pre-wrap">
      <h1 class="text-2xl font-bold mb-4">Runtime Error</h1>
      <pre class="text-sm">{{ error }}</pre>
    </div>
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { ref, onErrorCaptured, onMounted, onUnmounted } from 'vue'
import { useThemeStore } from '@/stores/theme'
import { Toaster } from 'vue-sonner'
import { safeEventsEmit, safeWindowShow } from '@/helpers/wailsRuntime'
import { setupToastListeners } from '@/helpers/toast'

const themeStore = useThemeStore()
const error = ref<string>('')
const isDev = import.meta.env.DEV

const globalClickListener = (e: MouseEvent) => {
  if (!isDev) return

  const target = e.target as HTMLElement
  console.log('Global Click:', target.tagName, target.className)

  safeEventsEmit('renderer-log', `Click: ${target.tagName}.${target.className}`)
}

onMounted(() => {
  themeStore.initTheme()
  setupToastListeners()
  safeWindowShow()
  if (isDev) {
    window.addEventListener('click', globalClickListener, true)
  }
})

onUnmounted(() => {
  if (isDev) {
    window.removeEventListener('click', globalClickListener, true)
  }
})

onErrorCaptured((err) => {
  const errorMessage = err instanceof Error ? err.stack || err.message : String(err)
  error.value = errorMessage
  console.error('❌ [App] Error captured:', err)

  safeEventsEmit('renderer-error', errorMessage)

  return false
})
</script>

<style>
/* Global CSS */
body {
  background: var(--background);
  color: var(--foreground);
}

html,
body {
  overflow: hidden;
  overscroll-behavior: none;
  height: 100vh;
  width: 100vw;
  margin: 0;
  padding: 0;
  background-color: var(--background);
}

#app {
  height: 100%;
  width: 100%;
  overflow: hidden;
  /* 确保 app 容器也不滚动 */
}

/* Sonner Toast 样式修复 */
[data-sonner-toaster] {
  position: fixed !important;
  top: 0 !important;
  left: 50% !important;
  transform: translateX(-50%) !important;
  z-index: 99999 !important;
  pointer-events: none !important;
  padding-top: 16px !important;
}

[data-sonner-toaster] [data-sonner-toast] {
  pointer-events: auto !important;
  display: flex !important;
  flex-direction: row !important;
  align-items: center !important;
  border-radius: 9999px !important;
  gap: 8px !important;
}

[data-sonner-toaster] [data-sonner-toast] [data-icon] {
  flex-shrink: 0 !important;
  display: flex !important;
  align-items: center !important;
}

[data-sonner-toaster] [data-sonner-toast] [data-content] {
  flex: 1 !important;
  display: flex !important;
  align-items: center !important;
}

/* Toast 用中性弹层表面，紧急程度只由图标的状态色表达——不整条染色。
   见「UI 设计规范 v1 · 用色纪律」 */
[data-sonner-toaster] [data-sonner-toast] {
  background-color: var(--popover) !important;
  color: var(--popover-foreground) !important;
  border: 1px solid var(--border) !important;
  box-shadow: var(--shadow-2) !important;
}

[data-sonner-toaster] [data-sonner-toast][data-type="success"] [data-icon] {
  color: var(--success) !important;
}

[data-sonner-toaster] [data-sonner-toast][data-type="error"] [data-icon] {
  color: var(--destructive) !important;
}

[data-sonner-toaster] [data-sonner-toast][data-type="warning"] [data-icon] {
  color: var(--warning) !important;
}

[data-sonner-toaster] [data-sonner-toast][data-type="info"] [data-icon] {
  color: var(--info) !important;
}



::-webkit-scrollbar {
  display: none;
}

/*滚动条两端的箭头*/
::-webkit-scrollbar-button {
  display: none;
}

::-webkit-scroll-track {
  display: none;
}

::-webkit-scrollbar-track-piece {
  display: none;
}

::-webkit-scrollbar-thumb {
  display: none;
}

::-webkit-scrollbar-corner {
  display: none;
}

::-webkit-resizer {
  display: none;
}

.github {
  font-size: 16px;
  margin-left: 16px;
  cursor: pointer;
}

.logo {
  user-select: none;
}

.application {
  /* font-family: PingFang SC,-apple-system,SF UI Text,Lucida Grande,STheiti,Microsoft YaHei,sans-serif !important; */
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
}

/* 恢复输入框和编辑器的系统原生选中文本机制 */
input,
textarea,
[contenteditable="true"],
.monaco-editor,
.monaco-editor *,
.monaco-editor-wrapper,
.monaco-editor-container {
  -webkit-user-select: auto !important;
  user-select: auto !important;
}
</style>
