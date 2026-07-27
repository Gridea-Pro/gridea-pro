/**
 * 文字颜色（含渐变）/ 字号：textStyle 的全局属性 + Markdown 序列化。
 *
 * 往返策略：解析方向 @tiptap/markdown 已能把内联 <span style="..."> 经 parseHTML 还原成
 * textStyle mark（含下方注册的全局属性）；序列化方向官方默认丢弃 textStyle —— 这里覆盖
 * CustomTextStyle.renderMarkdown，把 color/fontSize（含渐变 CSS）序列化为内联 <span style>，
 * goldmark(WithUnsafe) 发布时原生渲染，编辑器重开可无损还原。
 *
 * 渐变与命令实现移植自 editor-vue 的 GradientExtensions.ts / FontSize.ts（PandaWiki 同款）。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { Extension } from '@tiptap/core'
import { TextStyle } from '@tiptap/extension-text-style'

declare module '@tiptap/core' {
  interface Commands<ReturnType> {
    gradientColor: {
      setColor: (color: string | null) => ReturnType
      unsetColor: () => ReturnType
    }
    fontSize: {
      setFontSize: (size: string) => ReturnType
      unsetFontSize: () => ReturnType
    }
  }
}

function isGradient(v: unknown): boolean {
  return typeof v === 'string' && v.includes('gradient')
}

/** DOM style getter 会把 hex 归一成 rgb(...)；序列化统一回小写 hex，保证 .md 形态稳定 */
export function normalizeCssColor(v: string): string {
  const m = /^rgb\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*\)$/.exec(String(v).trim())
  if (!m) return String(v)
  return `#${[m[1], m[2], m[3]].map((n) => (+n).toString(16).padStart(2, '0')).join('')}`
}

/** textStyle：补 Markdown 序列化（color/fontSize → 内联 span；无属性时原样透传子内容） */
export const CustomTextStyle = TextStyle.extend({
  renderMarkdown(node: any, helpers: any) {
    const inner = helpers.renderChildren(node)
    const { color, fontSize } = node.attrs || {}
    const styles: string[] = []
    if (color && isGradient(color)) {
      styles.push(
        `background-image: ${color}`,
        '-webkit-background-clip: text',
        'background-clip: text',
        '-webkit-text-fill-color: transparent',
        'color: transparent',
      )
    } else if (color) {
      styles.push(`color: ${normalizeCssColor(color)}`)
    }
    if (fontSize) styles.push(`font-size: ${fontSize}`)
    if (!styles.length) return inner
    return `<span style="${styles.join('; ')}">${inner}</span>`
  },
} as any)

/** 文字颜色（支持纯色 + 渐变），替代官方 Color 扩展（其 renderHTML 不支持渐变） */
export const GradientColor = Extension.create({
  name: 'gradientColor',

  addGlobalAttributes() {
    return [
      {
        types: ['textStyle'],
        attributes: {
          color: {
            default: null,
            parseHTML: (element: HTMLElement) => {
              const bgImage = element.style.backgroundImage
              if (bgImage && bgImage.includes('gradient')) return bgImage
              return element.style.color || null
            },
            renderHTML: (attributes: Record<string, any>) => {
              if (!attributes.color) return {}
              const colorValue = attributes.color
              if (isGradient(colorValue)) {
                return {
                  style: `background-image: ${colorValue}; -webkit-background-clip: text; background-clip: text; -webkit-text-fill-color: transparent; color: transparent;`,
                }
              }
              return { style: `color: ${normalizeCssColor(colorValue)}` }
            },
          },
        },
      },
    ]
  },

  addCommands() {
    return {
      setColor:
        (color: string | null) =>
        ({ chain }: any) => {
          if (!color) {
            return chain().setMark('textStyle', { color: null }).removeEmptyTextStyle().run()
          }
          return chain().setMark('textStyle', { color }).run()
        },
      unsetColor:
        () =>
        ({ chain }: any) =>
          chain().setMark('textStyle', { color: null }).removeEmptyTextStyle().run(),
    }
  },
})

/** 字号（textStyle 属性），移植自 editor-vue FontSize.ts */
export const FontSize = Extension.create({
  name: 'fontSize',

  addOptions() {
    return { types: ['textStyle'] }
  },

  addGlobalAttributes() {
    return [
      {
        types: this.options.types,
        attributes: {
          fontSize: {
            default: null,
            parseHTML: (element: HTMLElement) => (element.style?.fontSize || '').replace(/['"]+/g, '') || null,
            renderHTML: (attributes: Record<string, any>) => {
              if (!attributes.fontSize) return {}
              return { style: `font-size: ${attributes.fontSize}` }
            },
          },
        },
      },
    ]
  },

  addCommands() {
    return {
      setFontSize:
        (fontSize: string) =>
        ({ chain }: any) =>
          chain().setMark('textStyle', { fontSize }).run(),
      unsetFontSize:
        () =>
        ({ chain }: any) =>
          chain().setMark('textStyle', { fontSize: null }).removeEmptyTextStyle().run(),
    }
  },
})
