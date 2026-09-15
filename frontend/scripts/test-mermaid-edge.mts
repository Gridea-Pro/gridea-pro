/**
 * Mermaid边界情况测试
 */
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
const { canonicalizeMarkdown } = await import('../src/components/editor/markdown/canonicalize.ts')

const manager = new (MarkdownManager as unknown as {
  new (o: { extensions: unknown[] }): { parse(md: string): unknown; serialize(json: unknown): string }
})({ extensions: buildExtensions({ content: '', placeholder: '', upload: async () => '' }) })

function test(name: string, md: string, expectPass = true) {
  try {
    const json = manager.parse(md)
    const out = canonicalizeMarkdown(manager.serialize(json))
    const pass = out === md
    if (pass === expectPass) {
      console.log(`✓ ${name}`)
    } else {
      console.log(`✗ ${name}`)
      console.log(`  Input:  ${JSON.stringify(md)}`)
      console.log(`  Output: ${JSON.stringify(out)}`)
    }
  } catch (e) {
    if (!expectPass) {
      console.log(`✓ ${name} (threw as expected)`)
    } else {
      console.log(`✗ ${name}: ${(e as Error).message}`)
    }
  }
}

console.log('=== Mermaid Edge Cases ===\n')
test('simple', '```mermaid\ngraph TD\n  A --> B\n```')
test('empty', '```mermaid\n\n```')
test('no trailing newline', '```mermaid\ngraph TD\n  A --> B```')
test('double trailing newline', '```mermaid\ngraph TD\n  A --> B\n\n```')
test('leading newline', '```mermaid\n\ngraph TD\n  A --> B\n```')
test('vs code block', '```js\nconst a = 1\n```')
test('mermaid vs js', '```mermaid\ngraph\n```\n\n```js\ncode\n```')
