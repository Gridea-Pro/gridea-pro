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

const manager = new (MarkdownManager as unknown as {
  new (o: { extensions: unknown[] }): { 
    parse(md: string): unknown
    serialize(json: unknown): string
    instance: any
  }
})({ extensions: buildExtensions({ content: '', placeholder: '', upload: async () => '' }) })

// Test mermaid vs regular code
const md1 = '```mermaid\ngraph TD\n  A --> B\n```'
const md2 = '```js\ncode\n```'

console.log('=== Parsing Mermaid vs JS ===\n')

console.log('Input 1 (mermaid):', JSON.stringify(md1))
const json1 = manager.parse(md1)
console.log('Parsed node type:', (json1 as any).content?.[0]?.type)
console.log('Node attrs:', (json1 as any).content?.[0]?.attrs)

console.log('\nInput 2 (js):', JSON.stringify(md2))
const json2 = manager.parse(md2)
console.log('Parsed node type:', (json2 as any).content?.[0]?.type)
console.log('Node attrs:', (json2 as any).content?.[0]?.attrs)

// Roundtrip
console.log('\n=== Roundtrip Test ===\n')
const out1 = manager.serialize(json1)
const out2 = manager.serialize(json2)
console.log('Mermaid roundtrip output:')
console.log(JSON.stringify(out1))
console.log('\nJS roundtrip output:')
console.log(JSON.stringify(out2))
