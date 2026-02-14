import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { resolve } from 'path'

export default defineConfig({
  plugins: [
    tailwindcss(),
    vue({
      script: {
        defineModel: true,
        propsDestructure: true,
      },
    }),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      'wailsjs': resolve(__dirname, 'src/wailsjs'),
    },
  },
  css: {
    preprocessorOptions: {
      less: {
        javascriptEnabled: true,
        additionalData: `@import "${resolve(__dirname, 'src/assets/styles/var.less')}";`,
      },
    },
  },
  server: {
    port: 5173,
    host: '127.0.0.1',
    strictPort: true,
  },
  build: {
    target: 'es2020',
    cssCodeSplit: true,
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor': ['vue', 'vue-router', 'pinia'],
          'markdown': ['markdown-it'],
          'editor': ['monaco-markdown'],
        },
      },
    },
    chunkSizeWarningLimit: 1000,
  },
  optimizeDeps: {
    include: ['vue', 'vue-router', 'pinia'],
  },
})
