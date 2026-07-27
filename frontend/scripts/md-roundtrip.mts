/**
 * 无头 Markdown 往返测试台。
 *
 * 走的是 @tiptap/markdown 的 MarkdownManager —— 与生产环境
 * editor.getMarkdown() / setContent(md,'markdown') 完全同一条 parse/serialize 管线、
 * 同一套扩展 markdown 钩子，因此结论对生产忠实；但不创建 EditorView/插件，规避 jsdom 视图问题。
 *
 * 用法：
 *   npx tsx scripts/md-roundtrip.mts            # 合成方言样例
 *   npx tsx scripts/md-roundtrip.mts --posts    # 叠加真实 posts/*.md
 *   npx tsx scripts/md-roundtrip.mts --posts --verbose
 *
 * 通过判定：md -> parse -> serialize 与原文一致（按白名单规范化后）。
 */
import { JSDOM } from 'jsdom'
import fs from 'node:fs'
import path from 'node:path'

// ── jsdom 全局（DOMParser 等），必须在引入 @tiptap 之前 ──────────
const dom = new JSDOM('<!doctype html><html><body></body></html>', { url: 'http://localhost' })
const win = dom.window
for (const k of [
  'window', 'document', 'navigator', 'HTMLElement', 'Element', 'Node', 'DocumentFragment',
  'Text', 'Comment', 'DOMParser', 'XMLSerializer', 'getComputedStyle', 'NodeFilter', 'Range',
]) {
  const value = (win as unknown as Record<string, unknown>)[k]
  if (value === undefined) continue
  try {
    Object.defineProperty(globalThis, k, { value, configurable: true, writable: true })
  } catch {
    /* getter-only & non-configurable — skip */
  }
}

const { MarkdownManager } = await import('@tiptap/markdown')
const { buildExtensions } = await import('../src/components/editor/extensions/index.ts')
const { canonicalizeMarkdown } = await import('../src/components/editor/markdown/canonicalize.ts')

const manager = new (MarkdownManager as unknown as {
  new (o: { extensions: unknown[] }): { parse(md: string): unknown; serialize(json: unknown): string }
})({ extensions: buildExtensions({ content: '', placeholder: '', upload: async () => '' }) })

// 与生产 getMarkdown 同一条规范化管线
function roundtrip(md: string): string {
  const json = manager.parse(md)
  return canonicalizeMarkdown(manager.serialize(json))
}

/**
 * 白名单规范化（canonical form）——吸收 marked 序列化器确定性、幂等的良性改写，
 * 使比较只反映「语义漂移」。白名单项：尾空白、空行折叠、表格单元格内边距与分隔行、
 * 有序列表标记间距、无序列表标记 *|+ → -、主题分割线 ***|___ → ---。
 */
function canon(s: string): string {
  return s
    .replace(/\r\n/g, '\n')
    .split('\n')
    .map((line) => {
      let l = line.replace(/[ \t]+$/g, '')
      // 主题分割线
      if (/^\s*([*_-])[ \t]*(\1[ \t]*){2,}$/.test(l)) return '---'
      // 表格行：裁剪单元格内边距 + 规范分隔行 dash 数
      if (/^\s*\|.*\|\s*$/.test(l)) {
        const cells = l.trim().replace(/^\||\|$/g, '').split('|').map((c) => {
          const t = c.trim()
          const m = t.match(/^(:?)-+(:?)$/)
          return m ? `${m[1]}---${m[2]}` : t
        })
        return `|${cells.join('|')}|`
      }
      // 有序列表标记间距：N.   x -> N. x
      l = l.replace(/^(\s*\d+[.)])[ \t]+/, '$1 ')
      // 无序列表标记统一为 -
      l = l.replace(/^(\s*)[*+][ \t]+/, '$1- ').replace(/^(\s*)-[ \t]+/, '$1- ')
      return l
    })
    .join('\n')
    .replace(/\n{3,}/g, '\n\n')
    .replace(/\n+$/g, '')
    .trim()
}
const norm = canon

function firstDiff(a: string, b: string): string {
  const la = a.split('\n')
  const lb = b.split('\n')
  const n = Math.max(la.length, lb.length)
  for (let i = 0; i < n; i++) {
    if (la[i] !== lb[i]) {
      return `  line ${i + 1}:\n    expect: ${JSON.stringify(la[i])}\n    actual: ${JSON.stringify(lb[i])}`
    }
  }
  return '  (only trailing differences)'
}

// expect 缺省时期望与 md 一致；给出 expect 表示「可接受的确定性规范化形态」
const SAMPLES: Array<{ name: string; md: string; expect?: string }> = [
  { name: 'heading', md: '# H1\n\n## H2\n\n### H3' },
  { name: 'emphasis', md: 'a **bold** and *italic* and `code` and ~~strike~~ text' },
  { name: 'highlight', md: 'a ==marked== word' },
  { name: 'highlight-spaces', md: 'a ==b and c== word' },
  { name: 'subsup', md: 'H~2~O and x^2^' },
  { name: 'inline-math', md: 'inline $E=mc^2$ here' },
  { name: 'block-math', md: '$$\n\\int_0^1 x dx\n$$' },
  { name: 'emoji', md: 'hello :smile: world' },
  { name: 'more-break', md: 'intro\n\n<!-- more -->\n\nbody' },
  { name: 'footnote', md: 'text with note[^1]\n\n[^1]: the note' },
  { name: 'task-list', md: '- [ ] todo\n- [x] done' },
  { name: 'bullet-list', md: '- one\n- two\n  - nested' },
  { name: 'ordered-list', md: '1. one\n2. two' },
  { name: 'blockquote', md: '> quoted line' },
  { name: 'code-block', md: '```js\nconst a = 1\n```' },
  { name: 'link', md: 'see [Gridea](https://gridea.pro)' },
  { name: 'image', md: '![alt](/post-images/x.png)' },
  { name: 'hr', md: 'a\n\n---\n\nb' },
  { name: 'table', md: '| a | b |\n| --- | --- |\n| 1 | 2 |' },
  { name: 'table-pipe', md: '| a | `x \\| y` | b |\n| --- | --- | --- |\n| 1 | 2 | 3 |' },
  { name: 'soft-break', md: 'line one\nline two' },
  { name: 'raw-html', md: '<div class="x">raw block</div>' },
  // 转义保真：转义字符不得丢失（反斜杠会规范化掉，字面字符保留）
  { name: 'escape-dollar', md: 'it costs \\$5 and \\$6 total', expect: 'it costs $5 and $6 total' },
  { name: 'escape-pipe', md: 'a \\| b \\| c', expect: 'a | b | c' },
  { name: 'escape-backslash', md: 'path C:\\\\temp\\\\x', expect: 'path C:\\temp\\x' },
  // HTML 实体：正文中的 < > & 不得被编码为实体
  { name: 'entities', md: 'if a < b && c > d then' },
  { name: 'entities-generic', md: 'List<T> and Map<K, V> generics' },
  { name: 'entity-in-code', md: 'use `a < b` inline code' },
  // flanking：成对比较运算符不得被误判为方言
  { name: 'eq-compare', md: 'assert a == b and c == d here' },
  // 单 ~ 含空格不再被误判为下标（flanking 修复）；注意 marked 会把它当删除线（见 spec-notes 已知边界），
  // 与 Gridea 的「字面」有分歧，但无损、稳定。
  { name: 'tilde-spaces', md: 'range ~a to b~ text', expect: 'range ~~a to b~~ text' },
  { name: 'caret-spaces', md: 'use ^a or b^ text' },
  // 围栏内的表格状文本不得被规整
  { name: 'fenced-table', md: '```\n| not | a | table |\n|-----|---|-------|\n```' },
]

const usePosts = process.argv.includes('--posts')
const verbose = process.argv.includes('--verbose')

const cases = [...SAMPLES]
if (usePosts) {
  const dir = '/Users/eric/Documents/Gridea Pro/posts'
  if (fs.existsSync(dir)) {
    for (const f of fs.readdirSync(dir).filter((x) => x.endsWith('.md'))) {
      let body = fs.readFileSync(path.join(dir, f), 'utf8')
      const m = body.match(/^---\n[\s\S]*?\n---\n?/)
      if (m) body = body.slice(m[0].length)
      cases.push({ name: `post:${f}`, md: body.trim() })
    }
  }
}

/** 词/URL 级 token（须含字母或数字），用于检测真实内容丢失（对重排/重格式化不敏感） */
function words(s: string): string[] {
  return (s.match(/[\p{L}\p{N}_./:@#$<>&-]{1,}/gu) || []).filter((w) => /[\p{L}\p{N}]/u.test(w))
}
const ENTITY_RE = /&(amp|lt|gt|quot|apos|#0?39);/g
function entityCount(s: string): number {
  return (s.match(ENTITY_RE) || []).length
}
function missingWords(from: string, to: string): string[] {
  const count = (arr: string[]) => {
    const m = new Map<string, number>()
    for (const w of arr) m.set(w, (m.get(w) || 0) + 1)
    return m
  }
  const a = count(words(from))
  const b = count(words(to))
  const miss: string[] = []
  for (const [w, c] of a) {
    const c2 = b.get(w) || 0
    for (let i = 0; i < c - c2; i++) miss.push(w)
  }
  return miss
}

let idem1 = 0 // 一次保存即稳定（out1===out2）
let stable = 0 // 收敛到不动点（out2===out3）——至多重格式化一次后永久稳定
let lossless = 0
let exactPass = 0
const fails: Array<{ name: string; kind: string; detail: string }> = []
for (const c of cases) {
  let out1: string
  let out2: string
  let out3: string
  try {
    out1 = roundtrip(c.md)
    out2 = roundtrip(out1)
    out3 = roundtrip(out2)
  } catch (e) {
    fails.push({ name: c.name, kind: 'THREW', detail: `  ${(e as Error).message}` })
    continue
  }
  const isSample = !c.name.startsWith('post:')
  const target = c.expect ?? c.md
  const onePass = out1 === out2
  const converged = out2 === out3
  const miss = missingWords(c.md, out1)
  const addedEntities = entityCount(out1) - entityCount(c.md)
  const exact = norm(out1) === norm(target)
  if (onePass) idem1++
  if (converged) stable++
  if (!miss.length && addedEntities <= 0) lossless++
  if (exact) exactPass++
  // 合成样例：要求精确匹配 expect（dialect 正确性）；真实 posts：要求无丢失/无新增实体 + 稳定
  if (!converged) fails.push({ name: c.name, kind: 'UNSTABLE', detail: firstDiff(out2, out3) })
  else if (miss.length) fails.push({ name: c.name, kind: 'CONTENT-LOSS', detail: `  missing(${miss.length}): ${miss.slice(0, 10).join(' · ')}` })
  else if (addedEntities > 0) fails.push({ name: c.name, kind: 'ENTITY-ENCODED', detail: `  +${addedEntities} HTML 实体被引入: ${(out1.match(ENTITY_RE) || []).slice(0, 6).join(' ')}` })
  else if (isSample && !exact) fails.push({ name: c.name, kind: 'DIALECT', detail: firstDiff(norm(target), norm(out1)) })
  else if (!onePass && verbose) fails.push({ name: c.name, kind: 'reformat-once', detail: firstDiff(out1, out2) })
  else if (!exact && verbose) fails.push({ name: c.name, kind: 'reformat', detail: firstDiff(norm(target), norm(out1)) })
}

// 硬门槛：无 THREW/UNSTABLE/CONTENT-LOSS/ENTITY-ENCODED/DIALECT。1-pass/exact 仅信息性。
const GATING = new Set(['THREW', 'UNSTABLE', 'CONTENT-LOSS', 'ENTITY-ENCODED', 'DIALECT'])
const blocking = fails.filter((f) => GATING.has(f.kind))
console.log(
  `\n==== stable ${stable}/${cases.length} | lossless ${lossless}/${cases.length} | 1-pass-idempotent ${idem1}/${cases.length} | exact ${exactPass}/${cases.length} ====`,
)
for (const f of fails) console.log(`\n✗ [${f.kind}] ${f.name}\n${f.detail}`)
console.log(`\n${blocking.length ? `FAIL: ${blocking.length} blocking issue(s)` : 'PASS: no blocking issues'}`)
process.exit(blocking.length ? 1 : 0)
