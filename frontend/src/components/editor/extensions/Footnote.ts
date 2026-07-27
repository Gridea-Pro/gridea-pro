/**
 * 脚注，对齐 Gridea 的 markdown-it-footnote：
 *   - 引用 `[^id]`（行内 sup）
 *   - 定义 `[^id]: 文本`（块级，文末）
 * Markdown 往返：定义文本以纯文本属性保存（attr-based），**与基础版完全一致、零漂移**。
 * 富文本环境额外注入 NodeView（自动编号 / 悬停预览 / 回跳 / 内联编辑），经工厂参数注入，
 * 测试台不传 → 退回 renderHTML，故本模块对 .vue 无静态依赖，可安全进入 buildExtensions / tsx 测试台。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { Node } from '@tiptap/core'
import { VueNodeViewRenderer } from '@tiptap/vue-3'

declare module '@tiptap/core' {
  interface Commands<ReturnType> {
    footnote: {
      /** 在光标处插入脚注引用，并在文末补一条空定义（若不存在） */
      insertFootnote: (id?: string) => ReturnType
    }
  }
}

export function createFootnoteRef(nodeView?: unknown) {
  const config: any = {
    name: 'footnoteRef',
    group: 'inline',
    inline: true,
    atom: true,

    addAttributes() {
      return { id: { default: '' } }
    },
    parseHTML() {
      return [{ tag: 'sup[data-footnote-ref]' }]
    },
    renderHTML({ node }: any) {
      return ['sup', { 'data-footnote-ref': '', class: 'footnote-ref' }, `[${node.attrs.id}]`]
    },

    markdownTokenName: 'footnoteRef',
    markdownTokenizer: {
      name: 'footnoteRef',
      level: 'inline',
      start: (src: string) => {
        const i = src.indexOf('[^')
        return i < 0 ? undefined : i
      },
      tokenize(src: string) {
        const m = /^\[\^([^\]\s]+)\](?!:)/.exec(src)
        if (!m) return
        return { type: 'footnoteRef', raw: m[0], id: m[1] }
      },
    },
    parseMarkdown: (token: any, helpers: any) => helpers.createNode('footnoteRef', { id: token.id }),
    renderMarkdown: (node: any) => `[^${node.attrs?.id ?? ''}]`,

    addCommands() {
      return {
        insertFootnote:
          (id?: string) =>
          ({ state, chain }: any) => {
            // 生成唯一 id：优先用传入；否则取「最小未被占用的正整数」字符串。
            // 收集所有现有 id（含字母 id），避免与 [^1]/[^ref1] 等混用时冲突。
            let fid = id
            if (!fid) {
              const used = new Set<string>()
              state.doc.descendants((n: any) => {
                if (n.type.name === 'footnoteRef' || n.type.name === 'footnoteDef') {
                  if (n.attrs.id) used.add(String(n.attrs.id))
                }
              })
              let k = 1
              while (used.has(String(k))) k++
              fid = String(k)
            }
            // 文末是否已有同 id 定义
            let hasDef = false
            state.doc.descendants((n: any) => {
              if (n.type.name === 'footnoteDef' && n.attrs.id === fid) hasDef = true
            })
            const c = chain().focus().insertContent({ type: 'footnoteRef', attrs: { id: fid } })
            if (!hasDef) {
              c.command(({ tr, dispatch }: any) => {
                if (!dispatch) return true
                const end = tr.doc.content.size
                const defNode = state.schema.nodes.footnoteDef.create({ id: fid, text: '' })
                tr.insert(end, defNode)
                return true
              })
            }
            return c.run()
          },
      }
    },
  }
  if (nodeView) config.addNodeView = () => VueNodeViewRenderer(nodeView as any)
  return Node.create(config)
}

export function createFootnoteDef(nodeView?: unknown) {
  const config: any = {
    name: 'footnoteDef',
    group: 'block',
    atom: true,

    addAttributes() {
      return { id: { default: '' }, text: { default: '' } }
    },
    parseHTML() {
      return [{ tag: 'div[data-footnote-def]' }]
    },
    renderHTML({ node }: any) {
      return [
        'div',
        { 'data-footnote-def': '', class: 'footnote-def' },
        `[^${node.attrs.id}]: ${node.attrs.text}`,
      ]
    },

    markdownTokenName: 'footnoteDef',
    markdownTokenizer: {
      name: 'footnoteDef',
      level: 'block',
      start: (src: string) => {
        const m = src.match(/(^|\n)\[\^[^\]\s]+\]:/)
        return m ? m.index : undefined
      },
      tokenize(src: string) {
        const m = /^\[\^([^\]\s]+)\]:[ \t]*(.*)/.exec(src)
        if (!m) return
        return { type: 'footnoteDef', raw: m[0], id: m[1], text: m[2] }
      },
    },
    parseMarkdown: (token: any, helpers: any) =>
      helpers.createNode('footnoteDef', { id: token.id, text: token.text }),
    renderMarkdown: (node: any) => `[^${node.attrs?.id ?? ''}]: ${node.attrs?.text ?? ''}`,
  }
  if (nodeView) config.addNodeView = () => VueNodeViewRenderer(nodeView as any)
  return Node.create(config)
}
