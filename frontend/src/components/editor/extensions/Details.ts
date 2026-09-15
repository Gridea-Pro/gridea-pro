/**
 * 折叠面板（details/summary）。
 *
 * 往返策略（关键）：发布渲染器是 goldmark + WithUnsafe，原生渲染 <details> HTML，故本节点
 * 序列化为规范 <details> HTML。**解析侧不直接吃 markdown token**——marked 会在 summary 后的空行
 * 处把 <details> 块拆成「html(开标签+summary) + 正文块 + html(闭标签)」三段，正文因此已是富文本节点。
 * 编辑器 I/O 边界（markdown/index.ts 的 setMarkdown）用纯函数 foldDetailsContent 把这三段折叠回
 * 单个 details 节点；getMarkdown 经 renderMarkdown 再拆回 <details> 文本。无头测试台直接走 manager
 * 往返（raw-html 形态，已证幂等），不受影响。
 *
 * NodeView 经工厂参数注入，测试台不传 → 退回 renderHTML，本模块对 .vue 无静态依赖。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { Node } from '@tiptap/core'
import { VueNodeViewRenderer } from '@tiptap/vue-3'

declare module '@tiptap/core' {
  interface Commands<ReturnType> {
    details: {
      setDetails: (attrs?: { summary?: string; open?: boolean }) => ReturnType
      toggleDetailsOpen: () => ReturnType
    }
  }
}

// summary 来自用户输入，序列化进 <summary> 文本前必须转义——否则 goldmark(WithUnsafe)
// 会把 <script> 等当真 HTML 执行（已发布页面 XSS）。解析侧（foldDetails）对应解码，保证幂等。
const SUMMARY_ESC: Record<string, string> = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }
function escapeSummary(s: string): string {
  return s.replace(/[&<>"]/g, (ch) => SUMMARY_ESC[ch] || ch)
}

export function createDetails(nodeView?: unknown) {
  const config: any = {
    name: 'details',
    group: 'block',
    content: 'block+',
    defining: true,

    addAttributes() {
      return {
        summary: { default: '' },
        open: { default: true },
      }
    },

    parseHTML() {
      return [
        {
          tag: 'details',
          getAttrs: (el: HTMLElement) => ({
            open: el.hasAttribute('open'),
            summary: el.querySelector('summary')?.textContent || '',
          }),
          // 把 <summary> 从正文里剔除，仅余下正文进入 content
          contentElement: (el: HTMLElement) => {
            const clone = el.cloneNode(true) as HTMLElement
            clone.querySelector('summary')?.remove()
            return clone
          },
        },
      ]
    },
    renderHTML({ node }: any) {
      return [
        'details',
        { open: node.attrs.open ? '' : undefined, class: 'details-node' },
        ['summary', {}, node.attrs.summary || ''],
        ['div', { class: 'details-content' }, 0],
      ]
    },

    // 序列化为规范 <details> HTML（goldmark unsafe 原生渲染）
    renderMarkdown(node: any, helpers: any) {
      const open = node.attrs?.open ? ' open' : ''
      const summary = escapeSummary(node.attrs?.summary || '')
      const body = helpers.renderChildren(node, '\n\n')
      return `<details${open}>\n<summary>${summary}</summary>\n\n${body}\n\n</details>`
    },

    addCommands() {
      return {
        setDetails:
          (attrs: { summary?: string; open?: boolean } = {}) =>
          ({ commands }: any) =>
            commands.insertContent({
              type: 'details',
              attrs: { summary: attrs.summary || '', open: attrs.open ?? true },
              content: [{ type: 'paragraph' }],
            }),
        toggleDetailsOpen:
          () =>
          ({ state, commands }: any) => {
            const { $from } = state.selection
            for (let d = $from.depth; d > 0; d--) {
              const n = $from.node(d)
              if (n.type.name === 'details') {
                return commands.updateAttributes('details', { open: !n.attrs.open })
              }
            }
            return false
          },
      }
    },
  }
  if (nodeView) config.addNodeView = () => VueNodeViewRenderer(nodeView as any)
  return Node.create(config)
}
