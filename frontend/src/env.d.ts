/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  // eslint-disable-next-line @typescript-eslint/no-explicit-any, @typescript-eslint/ban-types
  const component: DefineComponent<{}, {}, any>
  export default component
}

declare module 'slug'

// Markdown-it plugins
declare module 'markdown-it-sub'
declare module 'markdown-it-sup'
declare module 'markdown-it-footnote'
declare module 'markdown-it-abbr'
declare module 'markdown-it-emoji'
declare module 'markdown-it-task-lists'
declare module 'markdown-it-toc-and-anchor'
declare module 'markdown-it-mark'
declare module 'markdown-it-katex'
declare module 'markdown-it-imsize'
declare module 'markdown-it-image-lazy-loading'
declare module 'markdown-it-implicit-figures'
declare module '@iktakahiro/markdown-it-katex'
