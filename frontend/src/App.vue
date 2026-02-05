<template>
  <div id="app">
    <Toaster 
      position="top-center" 
      :expand="false"
      richColors
      :duration="2000"
      :closeButton="false"
      :toastOptions="{
        unstyled: false,
        classes: {
          toast: 'transform scale-75 !min-w-[160px] !max-w-[280px] !px-4 !py-2.5 !gap-2 !text-xs !shadow-lg !rounded-full !flex !flex-row !items-center',
          title: '!text-sm !font-medium',
          description: '!text-[11px]',
        }
      }"
    />



    <div 
      v-if="error" 
      class="fixed inset-0 z-[99999] overflow-auto bg-white p-5 text-red-600 whitespace-pre-wrap"
    >
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

<style lang="less">
@import './assets/styles/main.less';
@import './assets/styles/custom.less';
</style>


<style>

  /* Global CSS */
  body {
    background: var(--bg-body);
    color: var(--text-primary);
  }
  
  html, body {
  /* 1. 锁死整个窗口的滚动 */
  overflow: hidden; 
  /* 2. 禁止 macOS 的橡皮筋回弹效果 */
  overscroll-behavior: none; 
  /* 3. 确保占满屏幕 */
  height: 100vh;
  width: 100vw;
  margin: 0;
  padding: 0;
  /* 4. 设置背景色，防止加载时闪白 */
  background-color: var(--bg-body); 
  /* 5. 禁用文字选中 */
  /* -webkit-user-select: none;
  user-select: none; */
}

#app {
  height: 100%;
  width: 100%;
  overflow: hidden; /* 确保 app 容器也不滚动 */
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

  /* sonner toast 背景色 */
  [data-sonner-toaster] [data-sonner-toast][data-type="success"] {
    background-color: #000000 !important;
    color: #ffffff !important;
    border: 1px solid #000000 !important;
  }
  /* Success 图标颜色 */
  [data-sonner-toaster] [data-sonner-toast][data-type="success"] [data-icon] {
    color: #4ade80 !important; /* using green-400 for visibility on black */
  }

  [data-sonner-toaster] [data-sonner-toast][data-type="error"] {
    background-color: #FA5C5C !important;
    color: #ffffff !important;
    border: 1px solid #FA5C5C !important;
  }
  [data-sonner-toaster] [data-sonner-toast][data-type="error"] [data-icon] {
    color: #ffffff !important;
  }

  [data-sonner-toaster] [data-sonner-toast][data-type="warning"] {
    background-color: #FFC107 !important;
    color: #ffffff !important;
    border: 1px solid #FFC107 !important;
  }
  [data-sonner-toaster] [data-sonner-toast][data-type="warning"] [data-icon] {
    color: #ffffff !important;
  }

  [data-sonner-toaster] [data-sonner-toast][data-type="info"] {
    background-color: #000000 !important;
    color: #ffffff !important;
    border: 1px solid #000000 !important;
  }
  [data-sonner-toaster] [data-sonner-toast][data-type="info"] [data-icon] {
    color: #3b82f6 !important; /* blue-500 for info icon on black */
  }



  ::-webkit-scrollbar{
    display: none;
  }
  /*滚动条两端的箭头*/
  ::-webkit-scrollbar-button{
    display: none;
  }
  ::-webkit-scroll-track{
    display: none;
  }
  ::-webkit-scrollbar-track-piece {
    display: none;
  }
  ::-webkit-scrollbar-thumb{
    display: none;
  }
  ::-webkit-scrollbar-corner {
    display: none;
  }
  ::-webkit-resizer{
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

</style>
