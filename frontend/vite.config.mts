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
          'editor': ['monaco-editor', 'monaco-markdown'],
          'rich-editor': [
            '@tiptap/core',
            '@tiptap/vue-3',
            '@tiptap/starter-kit',
            '@tiptap/markdown',
            '@tiptap/extensions',
            '@tiptap/extension-table',
            '@tiptap/extension-list',
            '@tiptap/extension-image',
          ],
        },
      },
    },
    chunkSizeWarningLimit: 1000,
  },
  optimizeDeps: {
    include: [
      'vue',
      'vue-router',
      'pinia',
      'monaco-editor',
      'monaco-markdown',
      '@tiptap/core',
      '@tiptap/vue-3',
      '@tiptap/starter-kit',
      '@tiptap/markdown',
    ],
  }
})
