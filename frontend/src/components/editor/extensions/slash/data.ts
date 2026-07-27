/**
 * 斜杠命令清单（对齐 @ctzhian/tiptap，裁剪为 Gridea 当前已支持的节点）。
 * UI 触发型命令（链接/图片/Emoji/公式）通过编辑器 DOM 派发 CustomEvent，由 index.vue 打开对应弹窗/选择器，
 * 从而与具体组件解耦。
 */
import type { Editor, Range } from '@tiptap/core'
import {
  Heading1, Heading2, Heading3, List, ListOrdered, ListChecks, Quote, Code2, Minus,
  Table as TableIcon, Image as ImageIcon, Link as LinkIcon, Sigma, FunctionSquare,
  SeparatorHorizontal,
} from 'lucide-vue-next'
import type { Component } from 'vue'

export interface SlashItem {
  /** i18n key（editor.slash.*） */
  key: string
  group: 'basic' | 'block' | 'insert'
  icon: Component
  /** 搜索关键词（英文 + 拼音首字母 + 中文） */
  keywords: string
  command: (ctx: { editor: Editor; range: Range }) => void
}

/** 在编辑器 DOM 上派发动作事件，供 index.vue 监听并打开弹窗/选择器 */
export function dispatchEditorAction(editor: Editor, action: string): void {
  editor.view.dom.dispatchEvent(
    new CustomEvent('gridea-editor:action', { detail: { action }, bubbles: true }),
  )
}

export const SLASH_ITEMS: SlashItem[] = [
  { key: 'h1', group: 'basic', icon: Heading1, keywords: 'h1 heading 标题 biaoti yiji',
    command: ({ editor, range }) => editor.chain().focus().deleteRange(range).setNode('heading', { level: 1 }).run() },
  { key: 'h2', group: 'basic', icon: Heading2, keywords: 'h2 heading 标题 biaoti erji',
    command: ({ editor, range }) => editor.chain().focus().deleteRange(range).setNode('heading', { level: 2 }).run() },
  { key: 'h3', group: 'basic', icon: Heading3, keywords: 'h3 heading 标题 biaoti sanji',
    command: ({ editor, range }) => editor.chain().focus().deleteRange(range).setNode('heading', { level: 3 }).run() },
  { key: 'bulletList', group: 'basic', icon: List, keywords: 'bullet list ul 无序列表 wuxuliebiao',
    command: ({ editor, range }) => editor.chain().focus().deleteRange(range).toggleBulletList().run() },
  { key: 'orderedList', group: 'basic', icon: ListOrdered, keywords: 'ordered list ol 有序列表 youxuliebiao',
    command: ({ editor, range }) => editor.chain().focus().deleteRange(range).toggleOrderedList().run() },
  { key: 'taskList', group: 'basic', icon: ListChecks, keywords: 'task todo 任务列表 renwuliebiao daiban',
    command: ({ editor, range }) => editor.chain().focus().deleteRange(range).toggleTaskList().run() },
  { key: 'blockquote', group: 'basic', icon: Quote, keywords: 'quote blockquote 引用 yinyong',
    command: ({ editor, range }) => editor.chain().focus().deleteRange(range).toggleBlockquote().run() },
  { key: 'codeBlock', group: 'block', icon: Code2, keywords: 'code block 代码块 daimakuai',
    command: ({ editor, range }) => editor.chain().focus().deleteRange(range).toggleCodeBlock().run() },
  { key: 'hr', group: 'block', icon: Minus, keywords: 'hr divider rule 分割线 fengexian',
    command: ({ editor, range }) => editor.chain().focus().deleteRange(range).setHorizontalRule().run() },
  { key: 'table', group: 'block', icon: TableIcon, keywords: 'table 表格 biaoge',
    command: ({ editor, range }) => editor.chain().focus().deleteRange(range).insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run() },
  { key: 'more', group: 'block', icon: SeparatorHorizontal, keywords: 'more readmore 摘要 阅读更多 yuedugengduo zhaiyao',
    command: ({ editor, range }) => editor.chain().focus().deleteRange(range).insertContent({ type: 'moreBreak' }).run() },
  { key: 'image', group: 'insert', icon: ImageIcon, keywords: 'image picture 图片 tupian',
    command: ({ editor, range }) => { editor.chain().focus().deleteRange(range).run(); dispatchEditorAction(editor, 'image') } },
  { key: 'link', group: 'insert', icon: LinkIcon, keywords: 'link url 链接 lianjie',
    command: ({ editor, range }) => { editor.chain().focus().deleteRange(range).run(); dispatchEditorAction(editor, 'link') } },
  { key: 'inlineMath', group: 'insert', icon: Sigma, keywords: 'math inline 行内公式 gongshi katex',
    command: ({ editor, range }) => editor.chain().focus().deleteRange(range).insertContent({ type: 'inlineMath', attrs: { latex: '' } }).run() },
  { key: 'blockMath', group: 'insert', icon: FunctionSquare, keywords: 'math block 块级公式 gongshi katex',
    command: ({ editor, range }) => editor.chain().focus().deleteRange(range).insertContent({ type: 'blockMath', attrs: { latex: '' } }).run() },
]

export function filterSlashItems(query: string): SlashItem[] {
  const q = query.trim().toLowerCase()
  if (!q) return SLASH_ITEMS
  return SLASH_ITEMS.filter((it) => it.key.toLowerCase().includes(q) || it.keywords.toLowerCase().includes(q))
}
