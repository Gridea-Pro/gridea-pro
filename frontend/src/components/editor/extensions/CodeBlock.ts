/**
 * 代码块（lowlight 语法高亮）。Markdown 往返沿用官方 CodeBlockLowlight 的 GFM 围栏，
 * 本工厂只在富文本环境额外挂一个 NodeView（语言下拉 + 复制按钮）。
 * NodeView 经工厂参数注入，纯 Markdown 测试台不传 → 退回默认 <pre><code> 渲染，
 * 故本模块对 .vue 无静态依赖，可安全进入 buildExtensions / tsx 测试台。
 *
 * 注意：注册需早于 Mermaid 之后（Mermaid 先拦截 ```mermaid，其余 code 交给本扩展）。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { CodeBlockLowlight } from '@tiptap/extension-code-block-lowlight'
import { VueNodeViewRenderer } from '@tiptap/vue-3'
import { common, createLowlight } from 'lowlight'

export const lowlight = createLowlight(common)

/** 语言下拉候选（常用子集；节点已有但不在列表中的语言会被运行时补入） */
export const CODE_LANGUAGES = [
  'bash', 'c', 'cpp', 'csharp', 'css', 'diff', 'go', 'graphql', 'html', 'java',
  'javascript', 'json', 'kotlin', 'less', 'lua', 'makefile', 'markdown', 'objectivec',
  'php', 'python', 'r', 'ruby', 'rust', 'scss', 'shell', 'sql', 'swift', 'toml',
  'typescript', 'xml', 'yaml',
]

export function createCodeBlock(nodeView?: unknown) {
  const ext = nodeView
    ? CodeBlockLowlight.extend({
        addNodeView() {
          return VueNodeViewRenderer(nodeView as any)
        },
      })
    : CodeBlockLowlight
  return ext.configure({ lowlight })
}
