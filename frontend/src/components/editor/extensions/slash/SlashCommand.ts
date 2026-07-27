/**
 * 斜杠命令扩展：`@tiptap/suggestion` + `VueRenderer` 挂载 ui/SlashMenu.vue。
 * 注意：本扩展由 index.vue 在创建编辑器时加入（不进 buildExtensions），故纯 Markdown 往返
 * 测试台（tsx 引入 buildExtensions）不会拉入此处的 .vue，可安全使用**静态** import SlashMenu，
 * 从而保证 VueRenderer.element / .ref 同步可用（异步组件会令二者为 null/指向包装器实例）。
 *
 * SlashMenu.vue 契约：
 *   props: { items: SlashItem[]; command: (item: SlashItem) => void }
 *   defineExpose({ onKeyDown(event: KeyboardEvent): boolean })
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { Extension } from '@tiptap/core'
import Suggestion from '@tiptap/suggestion'
import { VueRenderer } from '@tiptap/vue-3'
import SlashMenu from '../../ui/SlashMenu.vue'
import { filterSlashItems, type SlashItem } from './data'

export const SlashCommand = Extension.create({
  name: 'slashCommand',

  addProseMirrorPlugins() {
    return [
      Suggestion<SlashItem>({
        editor: this.editor,
        char: '/',
        allowSpaces: false,
        startOfLine: false,
        items: ({ query }) => filterSlashItems(query),
        command: ({ editor, range, props }) => {
          props.command({ editor, range })
        },
        render: () => {
          let renderer: VueRenderer | null = null
          let getRect: (() => DOMRect | null) | undefined
          let raf = 0

          const position = () => {
            const el = renderer?.element as HTMLElement | undefined
            if (!el) return
            const rect = getRect?.()
            if (!rect) {
              el.style.display = 'none'
              return
            }
            el.style.display = ''
            el.style.position = 'fixed'
            el.style.left = `${Math.round(rect.left)}px`
            el.style.top = `${Math.round(rect.bottom + 6)}px`
            el.style.zIndex = '10050'
            cancelAnimationFrame(raf)
            raf = requestAnimationFrame(() => {
              if (!el.isConnected) return
              const h = el.offsetHeight
              if (rect.bottom + 6 + h > window.innerHeight && rect.top - 6 - h > 0) {
                el.style.top = `${Math.round(rect.top - 6 - h)}px`
              }
            })
          }
          const onScrollResize = () => position()

          return {
            onStart: (p) => {
              getRect = p.clientRect ?? undefined
              renderer = new VueRenderer(SlashMenu, {
                props: { items: p.items, command: p.command },
                editor: p.editor,
              })
              const el = renderer.element as HTMLElement | null
              if (!el) return
              el.classList.add('slash-menu-portal')
              document.body.appendChild(el)
              position()
              window.addEventListener('scroll', onScrollResize, true)
              window.addEventListener('resize', onScrollResize)
            },
            onUpdate: (p) => {
              if (!renderer) return
              getRect = p.clientRect ?? undefined
              renderer.updateProps({ items: p.items, command: p.command })
              position()
            },
            onKeyDown: (p) => {
              if (p.event.key === 'Escape') return true
              return (renderer?.ref as any)?.onKeyDown?.(p.event) ?? false
            },
            onExit: () => {
              cancelAnimationFrame(raf)
              window.removeEventListener('scroll', onScrollResize, true)
              window.removeEventListener('resize', onScrollResize)
              ;(renderer?.element as HTMLElement | undefined)?.remove()
              renderer?.destroy()
              renderer = null
            },
          }
        },
      }),
    ]
  },
})
