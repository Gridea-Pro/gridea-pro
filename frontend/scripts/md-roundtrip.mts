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
const { foldDetailsContent, foldAlignContent } = await import('../src/components/editor/markdown/foldDetails.ts')

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
  { name: 'footnote-dup-ref', md: 'see[^1] and again[^1] here\n\n[^1]: shared note' },
  { name: 'task-list', md: '- [ ] todo\n- [x] done' },
  { name: 'bullet-list', md: '- one\n- two\n  - nested' },
  { name: 'ordered-list', md: '1. one\n2. two' },
  { name: 'blockquote', md: '> quoted line' },
  { name: 'code-block', md: '```js\nconst a = 1\n```' },
  { name: 'mermaid', md: '```mermaid\ngraph TD\n  A --> B\n```' },
  { name: 'link', md: 'see [Gridea](https://gridea.pro)' },
  { name: 'image', md: '![alt](/post-images/x.png)' },
  // 带宽度图片：序列化为 <img>（goldmark unsafe 可发布），无宽度保持 ![]()
  { name: 'image-width', md: '<img src="/post-images/x.png" alt="截图" width="640">' },
  // 居中图片：对齐包裹（manager 直通路径走 raw-html 三段式）
  { name: 'image-center', md: '<div style="text-align: center">\n\n![图](/post-images/x.png)\n\n</div>' },
  { name: 'hr', md: 'a\n\n---\n\nb' },
  { name: 'table', md: '| a | b |\n| --- | --- |\n| 1 | 2 |' },
  { name: 'table-pipe', md: '| a | `x \\| y` | b |\n| --- | --- | --- |\n| 1 | 2 | 3 |' },
  { name: 'soft-break', md: 'line one\nline two' },
  { name: 'raw-html', md: '<div class="x">raw block</div>' },
  // 颜色/字号/带色高亮：内联 HTML 往返（textStyle/highlight 序列化层）
  { name: 'color-span', md: '带 <span style="color: #f5222d">红色</span> 文本' },
  { name: 'fontsize-span', md: '字号 <span style="font-size: 20px">大字</span> 文本' },
  { name: 'color-fontsize-span', md: '组合 <span style="color: #1890ff; font-size: 18px">蓝大</span> 文本' },
  { name: 'mark-color', md: '高亮 <mark style="background-color: #ffff00">黄底</mark> 文本' },
  { name: 'gradient-span', md: '渐变 <span style="background-image: linear-gradient(132deg, rgb(36, 73, 254) 0%, rgb(202, 75, 167) 100%); -webkit-background-clip: text; background-clip: text; -webkit-text-fill-color: transparent; color: transparent">文字</span> 后' },
  { name: 'gradient-mark', md: '渐变高亮 <mark style="background-image: linear-gradient(132deg, rgb(255, 65, 108) 0%, rgb(255, 75, 43) 100%);">文字</mark> 后' },
  // 对齐包裹（manager 直通路径走 raw-html 三段式）
  { name: 'align-center', md: '<div style="text-align: center">\n\n居中 **段落** 内容\n\n</div>' },
  { name: 'align-right-heading', md: '<div style="text-align: right">\n\n## 右对齐标题\n\n</div>' },
  // details 在 manager 直通路径（无折叠）下的 raw-html 三段式往返
  { name: 'details-raw', md: '<details>\n<summary>T</summary>\n\nbody\n\n</details>' },
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

// ── Details 折叠/序列化幂等测试（编辑器 I/O 路径，无需 EditorView）──────────
// manager 往返本身走 raw-html 形态（在上面的 SAMPLES 里已覆盖）；这里额外验证
// 「折叠成 details 节点 → 序列化回 <details> → 再解析折叠」整条编辑器链路幂等且无损。
function testDetails() {
  const cases = [
    {
      name: 'details-basic',
      md: '<details open>\n<summary>注意</summary>\n\n正文段落\n\n- 项目 A\n- 项目 B\n\n</details>',
    },
    {
      name: 'details-closed',
      md: '<details>\n<summary>展开看代码</summary>\n\n```js\nconst a = 1\n```\n\n</details>',
    },
    // summary 含 HTML 实体（已转义形态）：折叠解码 → 再序列化转义 应稳定回到同一形态
    {
      name: 'details-summary-entities',
      md: '<details>\n<summary>Title with &lt;tag&gt; &amp; x</summary>\n\nbody\n\n</details>',
    },
    // 多行 summary（marked 仍作单个 html token，OPEN_RE 的 [\\s\\S]*? 跨行匹配）
    {
      name: 'details-summary-multiline',
      md: '<details>\n<summary>第一行\n第二行</summary>\n\nbody\n\n</details>',
    },
  ]
  for (const c of cases) {
    try {
      // 1) manager 解析 → rawHtml(开)+正文+rawHtml(闭)
      const doc1 = manager.parse(c.md) as any
      // 2) 折叠成 details 节点
      const { content, changed } = foldDetailsContent(doc1.content || [])
      if (!changed) {
        fails.push({ name: c.name, kind: 'DETAILS', detail: '  未折叠出 details 节点（开/闭标签匹配失败）' })
        continue
      }
      const detailsNode = content.find((n: any) => n.type === 'details')
      if (!detailsNode) {
        fails.push({ name: c.name, kind: 'DETAILS', detail: '  折叠结果中无 details 节点' })
        continue
      }
      // 3) 序列化折叠后的文档 → <details> 文本（用生产 canonicalizeMarkdown，忠实于 getMarkdown，
      //    以捕捉实体反解等规范化交互；harness 局部 canon 不解码实体，会漏掉该类问题）
      const out1 = canonicalizeMarkdown(manager.serialize({ type: 'doc', content }))
      // 4) 再解析 + 折叠 + 序列化 → 幂等
      const doc2 = manager.parse(out1) as any
      const fold2 = foldDetailsContent(doc2.content || [])
      const out2 = canonicalizeMarkdown(manager.serialize({ type: 'doc', content: fold2.content }))
      if (out1 !== out2) {
        fails.push({ name: c.name, kind: 'DETAILS', detail: firstDiff(out1, out2) })
        continue
      }
      // 5) 无损：原文词级 token 不得丢失
      const miss = missingWords(c.md, out1)
      if (miss.length) {
        fails.push({ name: c.name, kind: 'DETAILS', detail: `  missing(${miss.length}): ${miss.slice(0, 8).join(' · ')}` })
        continue
      }
      // 6) 必须仍是 <details> HTML（可被 goldmark unsafe 发布）
      if (!/<details[^>]*>[\s\S]*<\/details>/.test(out1)) {
        fails.push({ name: c.name, kind: 'DETAILS', detail: '  序列化结果不含 <details> 标签' })
        continue
      }
    } catch (e) {
      fails.push({ name: c.name, kind: 'DETAILS', detail: `  ${(e as Error).message}` })
    }
  }

  // XSS 防护断言：summary 属性里的脚本必须被转义，绝不能原样进入 <summary>
  try {
    const malicious = `<script>alert('xss')</script>`
    const docJson = {
      type: 'doc',
      content: [
        {
          type: 'details',
          attrs: { summary: malicious, open: true },
          content: [{ type: 'paragraph', content: [{ type: 'text', text: 'body' }] }],
        },
      ],
    }
    // 必须走生产 canonicalizeMarkdown：它会反解正文实体，但 summary 标记行须豁免，否则转义被撤销
    const serialized = canonicalizeMarkdown(manager.serialize(docJson))
    if (serialized.includes('<script>')) {
      fails.push({ name: 'details-xss-escape', kind: 'DETAILS', detail: '  canonicalize 后 summary 仍含可执行 <script>，存在已发布页面 XSS' })
    }
    if (!serialized.includes('&lt;script&gt;')) {
      fails.push({ name: 'details-xss-escape', kind: 'DETAILS', detail: '  summary 未保持 &lt;script&gt; 转义形态' })
    }
  } catch (e) {
    fails.push({ name: 'details-xss-escape', kind: 'DETAILS', detail: `  ${(e as Error).message}` })
  }
}
testDetails()

// ── 对齐折叠/序列化幂等测试（编辑器 I/O 路径）──────────
function testAlign() {
  const cases = [
    { name: 'align-fold-center', md: '<div style="text-align: center">\n\n居中段落 **加粗**\n\n</div>', align: 'center', type: 'paragraph' },
    { name: 'align-fold-heading', md: '<div style="text-align: right">\n\n## 右对齐标题\n\n</div>', align: 'right', type: 'heading' },
  ]
  for (const c of cases) {
    try {
      const doc1 = manager.parse(c.md) as any
      const { content, changed } = foldAlignContent(doc1.content || [])
      if (!changed) {
        fails.push({ name: c.name, kind: 'DETAILS', detail: '  未折叠出对齐属性（开/闭 div 匹配失败）' })
        continue
      }
      const block = content.find((n: any) => n.type === c.type)
      if (!block || block.attrs?.textAlign !== c.align) {
        fails.push({ name: c.name, kind: 'DETAILS', detail: `  折叠后 ${c.type}.textAlign=${block?.attrs?.textAlign}，期望 ${c.align}` })
        continue
      }
      const out1 = canonicalizeMarkdown(manager.serialize({ type: 'doc', content }))
      const doc2 = manager.parse(out1) as any
      const fold2 = foldAlignContent(doc2.content || [])
      const out2 = canonicalizeMarkdown(manager.serialize({ type: 'doc', content: fold2.content }))
      if (out1 !== out2) {
        fails.push({ name: c.name, kind: 'DETAILS', detail: firstDiff(out1, out2) })
        continue
      }
      const miss = missingWords(c.md, out1)
      if (miss.length) {
        fails.push({ name: c.name, kind: 'DETAILS', detail: `  missing: ${miss.slice(0, 8).join(' · ')}` })
        continue
      }
      if (!/text-align:\s*(center|right)/.test(out1)) {
        fails.push({ name: c.name, kind: 'DETAILS', detail: '  序列化结果未保留 text-align 包裹' })
      }
    } catch (e) {
      fails.push({ name: c.name, kind: 'DETAILS', detail: `  ${(e as Error).message}` })
    }
  }
}
testAlign()

// ── 编辑器交互产物的序列化健壮性（markdown 解析不会产生这些形态，必须单独构造）──
// 空段落曾触发 renderChildren(node) 无 content 时的无限递归（栈溢出 → getMarkdown 返回空）
try {
  const out = manager.serialize({
    type: 'doc',
    content: [
      { type: 'paragraph' },
      { type: 'image', attrs: { src: '/post-images/x.png', alt: null, title: null } },
      { type: 'paragraph' },
      { type: 'heading', attrs: { level: 2 } },
    ],
  })
  if (typeof out !== 'string') {
    fails.push({ name: 'empty-block-serialize', kind: 'DETAILS', detail: '  序列化返回非字符串' })
  }
} catch (e) {
  fails.push({ name: 'empty-block-serialize', kind: 'DETAILS', detail: `  ${(e as Error).message}` })
}

// 硬门槛：无 THREW/UNSTABLE/CONTENT-LOSS/ENTITY-ENCODED/DIALECT/DETAILS。1-pass/exact 仅信息性。
const GATING = new Set(['THREW', 'UNSTABLE', 'CONTENT-LOSS', 'ENTITY-ENCODED', 'DIALECT', 'DETAILS'])
const blocking = fails.filter((f) => GATING.has(f.kind))
console.log(
  `\n==== stable ${stable}/${cases.length} | lossless ${lossless}/${cases.length} | 1-pass-idempotent ${idem1}/${cases.length} | exact ${exactPass}/${cases.length} ====`,
)
for (const f of fails) console.log(`\n✗ [${f.kind}] ${f.name}\n${f.detail}`)
console.log(`\n${blocking.length ? `FAIL: ${blocking.length} blocking issue(s)` : 'PASS: no blocking issues'}`)
process.exit(blocking.length ? 1 : 0)
