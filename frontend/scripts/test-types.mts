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

// Check if setMermaid is accessible through the extensions' type system
const mermaidExt = exts.find((e: any) => e.name === 'mermaid')
console.log('Mermaid extension found:', !!mermaidExt)
console.log('Mermaid addCommands:', typeof (mermaidExt as any).addCommands)

if ((mermaidExt as any).addCommands) {
  const commands = (mermaidExt as any).addCommands()
  console.log('Available commands:', Object.keys(commands))
}
