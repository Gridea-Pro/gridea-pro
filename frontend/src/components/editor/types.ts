/**
 * 编辑器共享类型
 */
import type { ShallowRef, Ref } from 'vue'
import type { Editor } from '@tiptap/vue-3'
import type { Component } from 'vue'

/** 目录项 */
export interface TocItem {
  id: string
  level: number
  text: string
}

/** 编辑器显示模式 */
export type EditorMode = 'rich' | 'source' | 'split'

/** 斜杠命令项 */
export interface SlashCommandItem {
  /** 唯一标识 */
  id: string
  /** i18n key（editor.* 下） */
  title: string
  /** 关键词，用于搜索过滤（含中英文/拼音首字母） */
  keywords?: string[]
  /** 分组标签 */
  group?: string
  /** 图标组件 */
  icon?: Component
  /** 执行命令 */
  command: (ctx: { editor: Editor; range: { from: number; to: number } }) => void
}

/** 工具栏按钮描述 */
export interface ToolbarAction {
  id: string
  title: string
  icon?: Component
  isActive?: (editor: Editor) => boolean
  isDisabled?: (editor: Editor) => boolean
  run: (editor: Editor) => void
}

/** TiptapEditor 对父组件暴露的命令式 API */
export interface EditorExposed {
  editor: ShallowRef<Editor | null>
  getMarkdown: () => string
  setMarkdown: (md: string) => void
  insertImageByPath: (path: string) => void
  insertImageBytes: (file: File) => Promise<void>
  insertMore: () => void
  insertEmoji: (emoji: string) => void
  focus: () => void
  mode: Ref<EditorMode>
}

/** 上传处理器：返回可用于 Markdown 的相对路径（如 /post-images/x.png） */
export type UploadHandler = (file: File) => Promise<string>

/** useEditor 组装选项 */
export interface BuildEditorOptions {
  /** 初始 markdown 内容 */
  content: string
  /** 占位符 */
  placeholder?: string
  /** 图片上传（粘贴/拖拽/选择统一入口） */
  upload: UploadHandler
  /** AI 续写：给定前后文返回补全 */
  aiComplete?: (prefix: string, suffix: string) => Promise<string>
  /** AI 润色：给定文本返回润色结果 */
  aiPolish?: (text: string) => Promise<string>
  /** 内容更新回调（markdown） */
  onUpdate?: (markdown: string) => void
  /** 目录更新回调 */
  onTocUpdate?: (toc: TocItem[]) => void
  /** 错误回调 */
  onError?: (err: Error, context: string) => void
  /**
   * 节点 NodeView 组件（仅富文本环境注入；纯 Markdown 管线/测试台不传，节点退回 renderHTML）。
   * key 为节点名：mermaid / codeBlock 等；值为 Vue 组件。
   */
  nodeViews?: Record<string, unknown>
}
