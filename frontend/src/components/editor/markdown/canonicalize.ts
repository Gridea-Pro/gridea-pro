/**
 * 把 Markdown 规整为「幂等的规范形式」：
 * - 表格：单元格最小内边距 `| a | b |`、分隔行 `---`、单元格内字面 `|` 转义为 `\|`，
 *   消除 marked 在表格列宽（尤其 CJK）上的非幂等抖动；
 * - HTML 实体：marked 序列化会把正文里的 `& < > " '` 编码成实体，这里在「非代码」区域反解回字面，
 *   避免改写用户源文件；
 * - 去行尾空白、去首尾空行。
 *
 * 全程**围栏代码块感知**（``` / ~~~）：围栏内一律原样，不做表格识别与实体反解。
 * 单元格/行内的代码跨度（`...`）也受保护。
 */

function decodeBasicEntities(s: string): string {
  // 安全：刻意不反解 &lt; / &gt;。@tiptap/markdown 会把正文里字面的 < > 序列化成 &lt; &gt;，
  // 若在此无差别反解回 < >，会把作者刻意用 &lt;script&gt; 安全展示的代码样例变回可执行标签，
  // 而后端发布走 goldmark.WithUnsafe() 会当真标签渲染 → 存储型 XSS（仅"加载→保存"就能触发）。
  // 只反解不构成可执行标签的实体（引号、&amp;），接受"字面 < > 以 &lt; &gt; 形式往返存储"这一
  // 更安全的默认——发布时 goldmark 会把 &lt; 渲染为字面 <，显示正确且不可执行。
  return s
    .replace(/&quot;/g, '"')
    .replace(/&#0?39;/g, "'")
    .replace(/&apos;/g, "'")
    .replace(/&amp;/g, '&') // 必须最后，避免把 &amp;lt; 误并成 &lt; 之外的形态
}

/** 仅在非行内代码（`...`/``..``）段反解实体 */
function decodeEntitiesOutsideCode(line: string): string {
  const parts = line.split(/(`+)/)
  let fence = ''
  let result = ''
  for (const p of parts) {
    if (/^`+$/.test(p)) {
      if (!fence) fence = p
      else if (p === fence) fence = ''
      result += p
      continue
    }
    result += fence ? p : decodeBasicEntities(p)
  }
  return result
}

/** 按未转义、且不在行内代码跨度内的 `|` 切分单元格（支持多反引号 run） */
function splitRow(row: string): string[] {
  const s = row.trim().replace(/^\|/, '').replace(/\|$/, '')
  const cells: string[] = []
  let cur = ''
  let fence = ''
  let i = 0
  while (i < s.length) {
    const ch = s[i]
    if (ch === '\\' && i + 1 < s.length) {
      cur += ch + s[i + 1]
      i += 2
      continue
    }
    if (ch === '`') {
      let j = i
      while (j < s.length && s[j] === '`') j++
      const run = s.slice(i, j)
      if (!fence) fence = run
      else if (run === fence) fence = ''
      cur += run
      i = j
      continue
    }
    if (ch === '|' && !fence) {
      cells.push(cur.trim())
      cur = ''
      i++
      continue
    }
    cur += ch
    i++
  }
  cells.push(cur.trim())
  return cells
}

function isSeparatorRow(row: string): boolean {
  const cells = splitRow(row)
  return cells.length > 0 && cells.every((c) => /^:?-{1,}:?$/.test(c))
}

/** 单元格内字面 | 统一转义为 \|（GFM 表格安全），幂等 */
function escapeCellPipes(c: string): string {
  return c.replace(/\\\|/g, '|').replace(/\|/g, '\\|')
}

function isTableRow(s: string | undefined): boolean {
  // 仅认定顶层（最多 3 空格缩进、无 > 前缀）的表格，避免误伤引用/列表内的伪表格
  return !!s && /^\s{0,3}\|.*\|/.test(s.trimEnd())
}

// 折叠面板 raw-html 标记行（<details>/<summary>…</summary>/</details>）：
// 这些是 HTML 而非正文，summary 里的实体是 Details.ts 刻意转义的（防 XSS / 防 <、& 破坏结构），
// 不可像正文那样反解回字面，否则会重新注入可执行 HTML。
function isRawHtmlMarkerLine(t: string): boolean {
  return /^<\/?details\b/i.test(t) || /^<summary>/i.test(t) || /<\/summary>\s*$/i.test(t)
}

// 任何含原始 HTML 标签的行都豁免实体反解。原因：反解引号实体（&quot; / &#39;）在 HTML 属性值里
// 会把转义的引号变回真引号，闭合属性并注入事件处理器，例如
// `<div title="&quot; onmouseover=alert(1) x=&quot;">` 经反解后 title 提前闭合、onmouseover 生效，
// 后端 goldmark.WithUnsafe() 当真属性渲染 → 存储型 XSS（仅"加载→保存"即触发）。
// 对 HTML 行保留实体字面即可，goldmark 会正确显示，且不可执行。
function containsRawHtmlTag(t: string): boolean {
  return /<\/?[a-zA-Z][\w-]*(?:\s|>|\/|$)/.test(t)
}

function normalizeTable(block: string[]): string[] {
  if (block.length < 2 || !isSeparatorRow(block[1])) return block
  const rows = block.map(splitRow)
  const cols = rows[0].length
  // 分隔行列数须与表头一致，否则不认定为可规整表格
  if (rows[1].length !== cols) return block
  const sep = rows[1].map((c) => `${c.startsWith(':') ? ':' : ''}---${c.endsWith(':') ? ':' : ''}`)
  const fmt = (cells: string[]) =>
    `| ${cells.map((c) => escapeCellPipes(decodeEntitiesOutsideCode(c))).join(' | ')} |`
  return [fmt(rows[0]), fmt(sep), ...rows.slice(2).map(fmt)]
}

export function canonicalizeMarkdown(md: string): string {
  const lines = md.replace(/\r\n/g, '\n').split('\n')
  const out: string[] = []
  let inFence = false
  let fenceChar = ''
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]
    const fenceMatch = line.match(/^\s{0,3}(```+|~~~+)/)
    if (fenceMatch) {
      const ch = fenceMatch[1][0]
      if (!inFence) {
        inFence = true
        fenceChar = ch
        out.push(line.replace(/[ \t]+$/g, ''))
        continue
      } else if (ch === fenceChar) {
        inFence = false
        out.push(line.replace(/[ \t]+$/g, ''))
        continue
      }
    }
    if (inFence) {
      out.push(line) // 围栏内原样
      continue
    }
    // 表格块：当前行是 | 行，下一行是分隔行
    if (isTableRow(line) && i + 1 < lines.length && isSeparatorRow(lines[i + 1])) {
      const block: string[] = []
      let j = i
      while (j < lines.length && isTableRow(lines[j])) {
        block.push(lines[j])
        j++
      }
      const normalized = normalizeTable(block)
      if (normalized !== block) {
        out.push(...normalized)
        i = j - 1
        continue
      }
    }
    const stripped = line.replace(/[ \t]+$/g, '')
    const t = stripped.trim()
    // details/summary 标记行，或任何含 HTML 标签的行：一律不反解实体，防止引号实体反解注入属性。
    const skipDecode = isRawHtmlMarkerLine(t) || containsRawHtmlTag(t)
    out.push(skipDecode ? stripped : decodeEntitiesOutsideCode(stripped))
  }
  return out
    .join('\n')
    .replace(/\n{3,}/g, '\n\n')
    .replace(/^(?:[ \t]*\n)+/, '')
    .replace(/\s+$/g, '') + '\n'
}
