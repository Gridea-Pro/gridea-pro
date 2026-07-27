/**
 * Mermaid 流程图节点。Markdown 往返：``` ```mermaid ``` ``` 围栏。
 * - 序列化 renderMarkdown → ```mermaid 围栏；
 * - 解析 parseMarkdown 拦截 lang==='mermaid' 的 code token（注册需早于 CodeBlockLowlight）。
 * NodeView 通过工厂参数注入（仅富文本环境），纯 Markdown 测试台不传 → 退回 renderHTML，
 * 故本模块对 .vue 无静态依赖，可安全进入 buildExtensions / tsx 测试台。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { Node } from '@tiptap/core'
import { VueNodeViewRenderer } from '@tiptap/vue-3'

declare module '@tiptap/core' {
  interface Commands<ReturnType> {
    mermaid: {
      setMermaid: (attrs?: { code?: string }) => ReturnType
      updateMermaid: (code: string) => ReturnType
    }
  }
}

export function createMermaid(nodeView?: unknown) {
  const config: any = {
    name: 'mermaid',
    group: 'block',
    atom: true,
    selectable: true,

    addAttributes() {
      return { code: { default: '' } }
    },
    parseHTML() {
      return [
        {
          tag: 'div[data-mermaid]',
          getAttrs: (el: HTMLElement) => ({ code: el.getAttribute('data-code') || '' }),
        },
      ]
    },
    renderHTML({ node }: any) {
      return [
        'div',
        { 'data-mermaid': '', 'data-code': node.attrs.code, class: 'mermaid-node' },
        node.attrs.code || '',
      ]
    },

    // 拦截 ```mermaid 围栏的 code token
    markdownTokenName: 'code',
    parseMarkdown(token: any, helpers: any) {
      if (String(token.lang || '').trim().toLowerCase() !== 'mermaid') return null
      return helpers.createNode('mermaid', { code: String(token.text || '').replace(/\n+$/, '') })
    },
    renderMarkdown(node: any) {
      return '```mermaid\n' + (node.attrs?.code || '') + '\n```'
    },

    addCommands() {
      return {
        setMermaid:
          (attrs: { code?: string } = {}) =>
          ({ commands }: any) =>
            commands.insertContent({ type: 'mermaid', attrs: { code: attrs.code || '' } }),
        updateMermaid:
          (code: string) =>
          ({ commands }: any) =>
            commands.updateAttributes('mermaid', { code }),
      }
    },
  }

  if (nodeView) {
    config.addNodeView = () => VueNodeViewRenderer(nodeView as any)
  }

  return Node.create(config)
}
