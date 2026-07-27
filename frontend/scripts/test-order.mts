import { JSDOM } from 'jsdom'

const dom = new JSDOM('<!doctype html><html><body></body></html>', { url: 'http://localhost' })
const win = dom.window
for (const k of [
  'window', 'document', 'navigator', 'HTMLElement', 'Element', 'Node', 'DocumentFragment',
  'Text', 'Comment', 'DOMParser', 'XMLSerializer', 'getComputedStyle', 'NodeFilter', 'Range',
]) {
  const value = (win as unknown as Record<string, unknown>)[k]
  if (value === undefined) continue
  try {
    Object.defineProperty(globalThis, k, { value, configurable: true, writable: true })
  } catch {}
}

const { MarkdownManager } = await import('@tiptap/markdown')
const { buildExtensions } = await import('../src/components/editor/extensions/index.ts')

const exts = buildExtensions({ content: '', placeholder: '', upload: async () => '' })

// 按注册顺序展示扩展
console.log('Extension registration order:')
for (let i = 0; i < exts.length; i++) {
  const ext = exts[i]
  const name = (ext as any).name
  const priority = (ext as any).options?.priority || 'default'
  console.log(`${i}: ${name} (priority: ${priority})`)
}

// 查找Mermaid和CodeBlock的位置
const mermaidIdx = exts.findIndex((ext: any) => ext.name === 'mermaid')
const codeBlockIdx = exts.findIndex((ext: any) => ext.name === 'codeBlock')
console.log(`\nMermaid index: ${mermaidIdx}`)
console.log(`CodeBlock index: ${codeBlockIdx}`)
console.log(`Mermaid is ${mermaidIdx < codeBlockIdx ? 'BEFORE' : 'AFTER'} CodeBlock`)
