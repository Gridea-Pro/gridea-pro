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
  return s
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#0?39;/g, "'")
    .replace(/&apos;/g, "'")
    .replace(/&amp;/g, '&') // 必须最后，避免 &amp;lt; 被误解码为 <
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
    out.push(isRawHtmlMarkerLine(stripped.trim()) ? stripped : decodeEntitiesOutsideCode(stripped))
  }
  return out
    .join('\n')
    .replace(/\n{3,}/g, '\n\n')
    .replace(/^(?:[ \t]*\n)+/, '')
    .replace(/\s+$/g, '') + '\n'
}
