import { JSDOM } from 'jsdom'

const dom = new JSDOM('<!doctype html><html><body></body></html>', { url: 'http://localhost' })
const win = dom.window
for (const k of ['window', 'document', 'navigator', 'HTMLElement', 'Element', 'Node', 'DocumentFragment',
  'Text', 'Comment', 'DOMParser', 'XMLSerializer', 'getComputedStyle', 'NodeFilter', 'Range']) {
  const value = (win as unknown as Record<string, unknown>)[k]
  if (value === undefined) continue
  try {
    Object.defineProperty(globalThis, k, { value, configurable: true, writable: true })
  } catch {}
}

const { MarkdownManager } = await import('@tiptap/markdown')
const { buildExtensions } = await import('../src/components/editor/extensions/index.ts')
const { canonicalizeMarkdown } = await import('../src/components/editor/markdown/canonicalize.ts')

const manager = new (MarkdownManager as unknown as {
  new (o: { extensions: unknown[] }): { parse(md: string): unknown; serialize(json: unknown): string }
})({ extensions: buildExtensions({ content: '', placeholder: '', upload: async () => '' }) })

function roundtrip(md: string): string {
  const json = manager.parse(md)
  return canonicalizeMarkdown(manager.serialize(json))
}

const testCases = [
  '```mermaid\ngraph TD\n  A --> B\n```',
  '```mermaid\n\n```',
  '```mermaid\ngraph TD\n  A --> B\n\n```',
  '```js\nconst x = 1\n```',
  '# title\n\n```mermaid\ngraph\n```\n\ntext',
  '```mermaid\ngraph TD\n  A[A] --> B[B]\n```\n\n```js\ncode\n```',
]

let passed = 0
let failed = 0

for (const md of testCases) {
  const out = roundtrip(md)
  const pass = out === roundtrip(out) // 幂等性
  if (pass) {
    passed++
    console.log(`✓ ${JSON.stringify(md.split('\n')[0])}`)
  } else {
    failed++
    console.log(`✗ ${JSON.stringify(md.split('\n')[0])}`)
    console.log(`  2nd roundtrip differs`)
  }
}

console.log(`\nResult: ${passed}/${testCases.length} passed`)
process.exit(failed > 0 ? 1 : 0)
