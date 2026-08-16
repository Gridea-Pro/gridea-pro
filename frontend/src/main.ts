import { createApp } from 'vue'
import { createPinia } from 'pinia'
import 'katex/dist/katex.min.css'
import '@fontsource/noto-serif/index.css'
import './assets/styles/tailwind.css'
import Prism from 'prismjs'
import type { App as VueApp } from 'vue'
import i18n from './locales'
import App from './App.vue'
import router from './router/index'
import { safeEventsEmit, isWailsEnvironment } from '@/helpers/wailsRuntime'
import { reportError } from '@/helpers/errorReporter'

declare global {
  interface Window {
    /** 置位后 index.html 的启动期兜底不再接管错误，全部交给应用层 */
    __GRIDEA_APP_READY__?: boolean
  }
}

function setupApp(app: VueApp): void {
  app.use(createPinia())
  app.use(router)
  app.use(i18n)

  // 兜住没被任何 ErrorBoundary 拦下的 Vue 错误：记录 + 落盘，但不接管界面
  app.config.errorHandler = (err, _instance, info) => {
    reportError(err, 'vue', info)
  }
}

function initializeApp(): void {
  if (import.meta.env.DEV) {
    console.log('🚀 [Main] Initializing app')

    // 检测 Wails 环境
    if (!isWailsEnvironment()) {
      console.warn('⚠️ [Main] Wails Runtime 不可用。如需完整功能，请访问 Wails 调试地址（通常为 http://localhost:34115）')
    }
  }

  try {
    Prism.highlightAll()

    const app = createApp(App)
    setupApp(app)
    app.mount('#app')
    window.__GRIDEA_APP_READY__ = true

    // 挂载后的全局错误（异步回调、非 Vue 代码）走统一收口，不再触发启动期兜底
    window.addEventListener('error', (event) => {
      reportError(event.error ?? event.message, 'window', event.filename || '')
    })
    window.addEventListener('unhandledrejection', (event) => {
      reportError(event.reason, 'promise')
    })

    if (import.meta.env.DEV) {
      console.log('✅ [Main] App mounted successfully')
    }

    // 安全调用 Wails Runtime
    safeEventsEmit('renderer-log', 'App initialized')
  } catch (error) {
    console.error('❌ [Main] Failed to mount app:', error)

    // 安全调用 Wails Runtime
    safeEventsEmit('renderer-error', `Mount Error: ${String(error)}`)

    throw error
  }
}

initializeApp()
