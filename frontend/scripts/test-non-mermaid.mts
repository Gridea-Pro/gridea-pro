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

const manager = new (MarkdownManager as unknown as {
  new (o: { extensions: unknown[] }): { parse(md: string): unknown; serialize(json: unknown): string }
})({ extensions: buildExtensions({ content: '', placeholder: '', upload: async () => '' }) })

function test(name: string, md: string, expectedType: string) {
  const json = manager.parse(md)
  const actual = (json as any).content?.[0]?.type
  const pass = actual === expectedType
  console.log(`${pass ? '✓' : '✗'} ${name}: expected ${expectedType}, got ${actual}`)
}

test('python', '```python\nprint("hello")\n```', 'codeBlock')
test('js', '```js\nconst x = 1\n```', 'codeBlock')
test('mermaid', '```mermaid\ngraph TD\n  A --> B\n```', 'mermaid')
test('mermaid uppercase', '```MERMAID\ngraph TD\n  A --> B\n```', 'mermaid')
test('mermaid mixed case', '```MeRmAiD\ngraph TD\n  A --> B\n```', 'mermaid')
test('no lang', '```\ncode\n```', 'codeBlock')
