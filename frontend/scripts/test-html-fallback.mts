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

const { getSchema } = await import('@tiptap/core')
const { buildExtensions } = await import('../src/components/editor/extensions/index.ts')

const schema = getSchema(buildExtensions({ content: '', placeholder: '', upload: async () => '' }))
const mermaidNodeType = schema.nodes.mermaid

// Test renderHTML
if (mermaidNodeType && mermaidNodeType.spec.toDOM) {
  const node = mermaidNodeType.create({ code: 'graph TD\n  A --> B' })
  const dom = mermaidNodeType.spec.toDOM?.(node)
  console.log('renderHTML output:', JSON.stringify(dom))
  console.log('Expected: ["div", { data-mermaid: "", data-code: "graph TD\\n  A --> B", class: "mermaid-node" }, "graph TD\\n  A --> B"]')
}
