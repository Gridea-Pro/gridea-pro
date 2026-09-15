/**
 * 扩展注册表。
 * 导入风格与 editor-vue（tiptap 3.15，已验证可用）一致；StarterKit v3 已内置
 * link / underline / code / codeBlock / lists / horizontalRule，故禁用 link、codeBlock 后
 * 用自定义/Lowlight 版本替换。
 */
import type { Extensions } from '@tiptap/core'
import StarterKit from '@tiptap/starter-kit'
import { Markdown } from '@tiptap/markdown'
import { createImage, ImageHtmlParser } from './ResizableImage'
import { CharacterCount } from '@tiptap/extension-character-count'
import Link from '@tiptap/extension-link'
import TextAlign from '@tiptap/extension-text-align'
import Placeholder from '@tiptap/extension-placeholder'
import { Table } from '@tiptap/extension-table'
import { TableRow } from '@tiptap/extension-table-row'
import { TableCell } from '@tiptap/extension-table-cell'
import { TableHeader } from '@tiptap/extension-table-header'
import TaskList from '@tiptap/extension-task-list'
import TaskItem from '@tiptap/extension-task-item'

import { AlignedParagraph, AlignedHeading } from './Align'
import { MarkdownEscape } from './MarkdownEscape'
import { CustomSubscript, CustomSuperscript } from './Script'
import { CustomTextStyle, GradientColor, FontSize } from './RichTextStyle'
import { CustomHighlight } from './Highlight'
import { createInlineMath, createBlockMath } from './Math'
import { CustomEmoji } from './Emoji'
import { createFootnoteRef, createFootnoteDef } from './Footnote'
import { MoreBreak } from './MoreBreak'
import { RawHtml } from './RawHtml'
import { createMermaid } from './Mermaid'
import { createCodeBlock } from './CodeBlock'
import { createDetails } from './Details'
import { createAiWriting } from './AiWriting'
import type { BuildEditorOptions } from '../types'

// 危险协议判定：与 Link.configure.isAllowedUri 共用，保证 HTML 与 Markdown 两条路径口径一致。
// 前导空白/大小写/制表符都规整掉再匹配，避免 "  JavaScript:" / "java\tscript:" 之类绕过。
function isSafeHref(url: string): boolean {
  // eslint-disable-next-line no-control-regex
  const normalized = (url || '').replace(/[\u0000-\u0020]+/g, '').toLowerCase()
  return !/^(javascript|vbscript|data|file):/.test(normalized)
}

// SafeLink 覆盖内置 Markdown 往返：内置 parseMarkdown/renderMarkdown 不校验 href，
// javascript: 链接会原样吃进/写回 .md，发布后成存储型 XSS。此处非法 href 一律降级为纯文本。
const SafeLink = Link.extend({
  parseMarkdown(this: { parent?: any }, token: any, helpers: any) {
    if (!isSafeHref(token?.href ?? '')) {
      // 非法协议：丢弃 link 标记，仅保留可见文字，绝不让危险 href 进入文档模型。
      return helpers.parseInline(token.tokens || [])
    }
    return helpers.applyMark('link', helpers.parseInline(token.tokens || []), {
      href: token.href,
      title: token.title || null,
    })
  },
  renderMarkdown(node: any, h: any) {
    const href = node?.attrs?.href ?? ''
    const text = h.renderChildren(node)
    // 双保险：即便非法 href 混进了文档模型，序列化落盘时也不写成链接。
    if (!isSafeHref(href)) return text
    const title = node?.attrs?.title ?? ''
    return title ? `[${text}](${href} "${title}")` : `[${text}](${href})`
  },
})

export function buildExtensions(options: BuildEditorOptions): Extensions {
  return [
    StarterKit.configure({
      // 用自定义/增强版替换
      link: false,
      codeBlock: false,
      // 段落/标题用带对齐序列化的版本（见 Align.ts）
      paragraph: false,
      heading: false,
      dropcursor: { color: 'var(--editor-accent)', width: 2 },
    }),
    AlignedParagraph,
    AlignedHeading,

    // 链接（自定义配置 + Markdown 往返协议校验）
    // 安全：只允许安全协议，阻断 javascript:/vbscript:/data:/file: 等——否则
    // [文字](javascript:...) 这类链接会随 Markdown 存回、发布后在站点页面点击即触发存储型 XSS。
    // isAllowedUri/protocols 只作用于 HTML/粘贴/autolink；Markdown 往返（.md 加载→序列化落盘）
    // 走内置 parseMarkdown/renderMarkdown，二者默认不校验 href，故此处 .extend 覆盖，非法链接降级为纯文本。
    SafeLink.configure({
      openOnClick: false,
      autolink: true,
      HTMLAttributes: { class: 'editor-link' },
      protocols: ['http', 'https', 'mailto', 'tel'],
      isAllowedUri: (url: string, ctx: { defaultValidate: (u: string) => boolean }) => {
        if (!isSafeHref(url)) return false
        return ctx.defaultValidate(url)
      },
    }),

    // 反斜杠转义保真（必须，避免 \$ \* \| 等被静默丢弃）
    MarkdownEscape,

    // 文本格式（== / ^ / ~ 方言钩子）
    CustomSubscript,
    CustomSuperscript,
    // textStyle 带 Markdown 序列化（颜色/字号 → 内联 span）；GradientColor 替代官方 Color
    CustomTextStyle,
    GradientColor,
    FontSize,
    CustomHighlight,

    // 对齐
    TextAlign.configure({ types: ['heading', 'paragraph'] }),

    // 图片（width 属性 + 拖拽调宽 NodeView；带宽序列化为 <img>，无宽保持 ![]()）
    createImage(options.nodeViews?.image),
    // <img> html token → image 节点（独立解析器，须早于 RawHtml）
    ImageHtmlParser,

    // 表格
    Table.configure({ resizable: true }),
    TableRow,
    TableHeader,
    TableCell,

    // 任务列表
    TaskList,
    TaskItem.configure({ nested: true }),

    // Mermaid（须早于 CodeBlock 注册，以拦截 ```mermaid 围栏）
    createMermaid(options.nodeViews?.mermaid),

    // 代码块高亮（NodeView：语言下拉 + 复制，仅富文本环境注入）
    createCodeBlock(options.nodeViews?.codeBlock),

    // 行内/块级公式（$ / $$，规则对齐 markdown-it-katex）
    // onClick 仅在富文本环境注入 → 点击打开 LaTeX 编辑弹层；测试台不传 → 仅渲染
    createInlineMath(options.onMathEdit),
    createBlockMath(options.onMathEdit),

    // Emoji 短码 :name: 与脚注 [^id]（脚注 NodeView 仅富文本环境注入）
    CustomEmoji,
    createFootnoteRef(options.nodeViews?.footnoteRef),
    createFootnoteDef(options.nodeViews?.footnoteDef),

    // 折叠面板（details 节点；解析侧由 setMarkdown 的 foldDetailsContent 折叠 raw-html 三段式，
    // 此处仅提供 schema/renderMarkdown/NodeView，不直接吃 markdown token）
    createDetails(options.nodeViews?.details),

    // <!-- more --> 与原始 HTML 透传（MoreBreak 优先拦截 more 注释，其余交 RawHtml 兜底）
    MoreBreak,
    RawHtml,

    // 工具类
    CharacterCount.configure({ limit: null }),
    Placeholder.configure({
      placeholder: options.placeholder || '输入 / 唤起命令，或直接开始写作…',
      showOnlyWhenEditable: true,
    }),

    // 行内 AI 续写（仅在提供 aiComplete 时生效；纯 ProseMirror，无 .vue 依赖）
    createAiWriting(
      options.aiComplete ? (ctx) => options.aiComplete!(ctx.prefix, ctx.suffix) : undefined,
    ),

    // Markdown 往返（必须在最后，读取各扩展的 renderMarkdown/parseMarkdown）
    // breaks:true 对齐 Gridea 的 helpers/markdown（markdown-it breaks:true），避免段内单换行漂移
    Markdown.configure({
      indentation: { style: 'space', size: 2 },
      markedOptions: { gfm: true, breaks: true },
    }),
  ]
}
