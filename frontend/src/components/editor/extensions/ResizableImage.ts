/**
 * 图片：宽度属性 + 拖拽调整大小（NodeView 注入）+ Markdown 双形态往返。
 *
 * - 无宽度：保持纯 markdown `![alt](src "title")`；
 * - 有宽度：序列化为 `<img src alt title width>`（goldmark WithUnsafe 原生渲染，
 *   markdown 的 ![]() 语法装不下宽度，imsize 方言 goldmark 不认）。
 * - 解析：独立成行的 <img> 是块级 html token——本扩展以 markdownTokenName:'html'
 *   先于 RawHtml 拦截（非 <img> 返回 null 走兜底，与 Mermaid/CodeBlock 同模式）。
 * NodeView 经工厂注入，测试台不传 → renderHTML，模块无 .vue 静态依赖。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { Extension } from '@tiptap/core'
import Image from '@tiptap/extension-image'
import { VueNodeViewRenderer } from '@tiptap/vue-3'
import { wrapAlign } from './Align'

function escAttr(s: string): string {
  return String(s).replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/</g, '&lt;')
}

export function createImage(nodeView?: unknown) {
  const config: any = {
    addAttributes() {
      return {
        ...this.parent?.(),
        /** 显示宽度（px 整数）；null = 原始大小 */
        width: {
          default: null,
          parseHTML: (el: HTMLElement) => {
            const w = el.getAttribute('width') || (el.style?.width || '').replace('px', '')
            const n = parseInt(w, 10)
            return Number.isFinite(n) && n > 0 ? n : null
          },
          renderHTML: (attrs: Record<string, any>) => (attrs.width ? { width: attrs.width } : {}),
        },
        /** 对齐（left/center/right）；序列化复用 align 包裹，解析由 foldAlignContent 折回 */
        textAlign: {
          default: null,
          parseHTML: (el: HTMLElement) => el.getAttribute('data-align') || null,
          renderHTML: (attrs: Record<string, any>) => (attrs.textAlign ? { 'data-align': attrs.textAlign } : {}),
        },
      }
    },

    renderMarkdown(node: any) {
      const { src = '', alt, title, width, textAlign } = node.attrs || {}
      let md: string
      if (width) {
        const altAttr = ` alt="${alt ? escAttr(alt) : ''}"`
        const titleAttr = title ? ` title="${escAttr(title)}"` : ''
        md = `<img src="${escAttr(src)}"${altAttr}${titleAttr} width="${width}">`
      } else {
        const t = title ? ` "${title}"` : ''
        md = `![${alt || ''}](${src}${t})`
      }
      return wrapAlign(md, textAlign)
    },
  }
  if (nodeView) config.addNodeView = () => VueNodeViewRenderer(nodeView as any)
  return Image.extend(config).configure({ inline: false, allowBase64: false })
}

/**
 * 独立解析器：把「整行只有一个 <img>」的块级 html token 解析为 image 节点。
 * 必须独立于 Image 扩展——若直接改 Image 的 markdownTokenName 为 'html'，
 * 会丢掉对标准 ![]() image token 的解析（已踩过）。注册须早于 RawHtml（null 走兜底）。
 */
export const ImageHtmlParser = Extension.create({
  name: 'imageHtmlParser',
  markdownTokenName: 'html',
  parseMarkdown(token: any, helpers: any) {
    const raw = String(token.raw ?? token.text ?? '').trim()
    if (!/^<img\s[^>]*\/?>$/i.test(raw)) return null
    const host = document.createElement('div')
    host.innerHTML = raw
    const img = host.querySelector('img')
    if (!img || !img.getAttribute('src')) return null
    const w = parseInt(img.getAttribute('width') || '', 10)
    return helpers.createNode('image', {
      src: img.getAttribute('src') || '',
      alt: img.getAttribute('alt') || null,
      title: img.getAttribute('title') || null,
      width: Number.isFinite(w) && w > 0 ? w : null,
    })
  },
} as any)
