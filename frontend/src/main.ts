import { createApp } from 'vue'
import { createPinia } from 'pinia'
import 'katex/dist/katex.min.css'
import '@fontsource/noto-serif/index.css'
import './assets/styles/tailwind.css'
import './assets/styles/main.less'
import Prism from 'prismjs'
import type { App as VueApp } from 'vue'
import i18n from './locales'
import App from './App.vue'
import router from './router/index'
import { safeEventsEmit, isWailsEnvironment } from '@/helpers/wailsRuntime'



function setupApp(app: VueApp): void {
  app.use(createPinia())
  app.use(router)
  app.use(i18n)
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
