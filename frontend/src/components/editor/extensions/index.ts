/**
 * 扩展注册表。
 * 导入风格与 editor-vue（tiptap 3.15，已验证可用）一致；StarterKit v3 已内置
 * link / underline / code / codeBlock / lists / horizontalRule，故禁用 link、codeBlock 后
 * 用自定义/Lowlight 版本替换。
 */
import type { Extensions } from '@tiptap/core'
import StarterKit from '@tiptap/starter-kit'
import { Markdown } from '@tiptap/markdown'
import { Image } from '@tiptap/extension-image'
import { CharacterCount } from '@tiptap/extension-character-count'
import Link from '@tiptap/extension-link'
import TextAlign from '@tiptap/extension-text-align'
import { TextStyle } from '@tiptap/extension-text-style'
import { Color } from '@tiptap/extension-color'
import Placeholder from '@tiptap/extension-placeholder'
import { Table } from '@tiptap/extension-table'
import { TableRow } from '@tiptap/extension-table-row'
import { TableCell } from '@tiptap/extension-table-cell'
import { TableHeader } from '@tiptap/extension-table-header'
import TaskList from '@tiptap/extension-task-list'
import TaskItem from '@tiptap/extension-task-item'
import { CodeBlockLowlight } from '@tiptap/extension-code-block-lowlight'
import { common, createLowlight } from 'lowlight'

import { MarkdownEscape } from './MarkdownEscape'
import { CustomSubscript, CustomSuperscript } from './Script'
import { CustomHighlight } from './Highlight'
import { CustomInlineMath, CustomBlockMath } from './Math'
import { CustomEmoji } from './Emoji'
import { FootnoteRef, FootnoteDef } from './Footnote'
import { MoreBreak } from './MoreBreak'
import { RawHtml } from './RawHtml'
import type { BuildEditorOptions } from '../types'

const lowlight = createLowlight(common)

export function buildExtensions(options: BuildEditorOptions): Extensions {
  return [
    StarterKit.configure({
      // 用自定义/增强版替换
      link: false,
      codeBlock: false,
      dropcursor: { color: 'var(--editor-accent)', width: 2 },
    }),

    // 链接（自定义配置）
    Link.configure({
      openOnClick: false,
      autolink: true,
      HTMLAttributes: { class: 'editor-link' },
    }),

    // 反斜杠转义保真（必须，避免 \$ \* \| 等被静默丢弃）
    MarkdownEscape,

    // 文本格式（== / ^ / ~ 方言钩子）
    CustomSubscript,
    CustomSuperscript,
    TextStyle,
    Color,
    CustomHighlight,

    // 对齐
    TextAlign.configure({ types: ['heading', 'paragraph'] }),

    // 图片
    Image.configure({ inline: false, allowBase64: false }),

    // 表格
    Table.configure({ resizable: true }),
    TableRow,
    TableHeader,
    TableCell,

    // 任务列表
    TaskList,
    TaskItem.configure({ nested: true }),

    // 代码块高亮
    CodeBlockLowlight.configure({ lowlight }),

    // 行内/块级公式（$ / $$，规则对齐 markdown-it-katex）
    CustomInlineMath,
    CustomBlockMath,

    // Emoji 短码 :name: 与脚注 [^id]
    CustomEmoji,
    FootnoteRef,
    FootnoteDef,

    // <!-- more --> 与原始 HTML 透传（MoreBreak 优先拦截 more 注释，其余交 RawHtml 兜底）
    MoreBreak,
    RawHtml,

    // 工具类
    CharacterCount.configure({ limit: null }),
    Placeholder.configure({
      placeholder: options.placeholder || '输入 / 唤起命令，或直接开始写作…',
      showOnlyWhenEditable: true,
    }),

    // Markdown 往返（必须在最后，读取各扩展的 renderMarkdown/parseMarkdown）
    // breaks:true 对齐 Gridea 的 helpers/markdown（markdown-it breaks:true），避免段内单换行漂移
    Markdown.configure({
      indentation: { style: 'space', size: 2 },
      markedOptions: { gfm: true, breaks: true },
    }),
  ]
}
