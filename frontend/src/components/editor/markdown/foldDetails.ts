/**
 * <details> 折叠：把 manager 解析出的「rawHtml(开标签+summary) + 正文块 + rawHtml(闭标签)」三段序列
 * 折叠回单个 details 节点。纯 JSON 函数（不依赖 EditorView/PM 实例），供：
 *   - 编辑器 I/O 边界 setMarkdown（对 editor.getJSON() 的结果折叠后回灌）
 *   - 无头测试台 details 幂等性测试
 * 同时复用。
 */
/* eslint-disable @typescript-eslint/no-explicit-any */

export interface JNode {
  type: string
  attrs?: Record<string, any>
  content?: JNode[]
  [k: string]: any
}

// 开标签 rawHtml 形如：<details> / <details open>\n<summary>标题</summary>
const OPEN_RE = /^<details(\s+open(?:="[^"]*")?)?\s*>\s*<summary>([\s\S]*?)<\/summary>\s*$/i
const CLOSE_RE = /^<\/details>\s*$/i

// 与 Details.ts 的 escapeSummary 对偶：把 <summary> 文本里的实体解回字面，存入 summary 属性。
// 保证 attr → render(转义) → parse(解码) → attr 幂等。
const ENT: Record<string, string> = { amp: '&', lt: '<', gt: '>', quot: '"', '#39': "'", '#039': "'", apos: "'" }
function decodeEntities(s: string): string {
  return s.replace(/&(amp|lt|gt|quot|#0?39|apos);/g, (_, e) => ENT[e] ?? _)
}

function matchOpen(html: string): { open: boolean; summary: string } | null {
  const m = OPEN_RE.exec(html.trim())
  if (!m) return null
  return { open: !!m[1], summary: decodeEntities((m[2] || '').trim()) }
}

function isClose(html: string): boolean {
  return CLOSE_RE.test(html.trim())
}

/**
 * 折叠顶层节点序列中的 details 三段式。仅处理顶层（details 几乎总是顶层；嵌套 details 极罕见，
 * 不在此折叠，会退回 raw-html 形态，仍可发布、仍无损）。
 * @returns { content, changed } changed 为 false 时调用方应跳过回灌，避免多余 setContent。
 */
export function foldDetailsContent(nodes: JNode[]): { content: JNode[]; changed: boolean } {
  const out: JNode[] = []
  let changed = false
  let i = 0
  while (i < nodes.length) {
    const n = nodes[i]
    const openInfo = n?.type === 'rawHtml' ? matchOpen(String(n.attrs?.html ?? '')) : null
    if (openInfo) {
      const body: JNode[] = []
      let j = i + 1
      let closed = false
      while (j < nodes.length) {
        const m = nodes[j]
        const html = m?.type === 'rawHtml' ? String(m.attrs?.html ?? '') : ''
        if (html && isClose(html)) {
          closed = true
          break
        }
        // 遇到另一个开标签 → 不跨嵌套折叠，放弃本次
        if (html && matchOpen(html)) break
        body.push(m)
        j++
      }
      if (closed) {
        out.push({
          type: 'details',
          attrs: { summary: openInfo.summary, open: openInfo.open },
          content: body.length ? body : [{ type: 'paragraph' }],
        })
        changed = true
        i = j + 1
        continue
      }
    }
    out.push(n)
    i++
  }
  return { content: out, changed }
}
