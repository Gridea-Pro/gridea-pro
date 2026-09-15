/**
 * 对比度契约校验。
 *
 * 解析 src/assets/styles/tokens.css，对「外观 × 强调色 × 明暗」的每一种组合
 * 求出语义 token 的实际色值，断言 WCAG 对比度：正文 ≥ 4.5，UI/大字 ≥ 3.0。
 *
 * 自己实现 OKLCH 求值而不是拉起浏览器，是为了让这个检查能无依赖地进 CI。
 * 求值范围刻意只覆盖 tokens.css 实际用到的语法子集。
 *
 *   npm run check:contrast
 */
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

const HERE = dirname(fileURLToPath(import.meta.url))
const TOKENS_CSS = resolve(HERE, '../src/assets/styles/tokens.css')

const SURFACES = ['pure', 'paper', 'glass'] as const
const ACCENTS = ['green', 'rose', 'sakura', 'sunset', 'amber', 'cyan', 'blue', 'purple', 'ink'] as const
const MODES = ['light', 'dark'] as const

/** [说明, 前景 token, 背景 token, 最低对比度] */
const CONTRACT: Array<[string, string, string, number]> = [
  ['正文文字', 'foreground', 'background', 4.5],
  ['正文文字（卡片上）', 'card-foreground', 'card', 4.5],
  ['次要文字', 'muted-foreground', 'background', 4.5],
  ['次要文字（卡片上）', 'muted-foreground', 'card', 4.5],
  ['次要文字（侧栏上）', 'muted-foreground', 'sidebar', 4.5],
  ['主按钮文字', 'primary-foreground', 'primary', 3.0],
  ['链接 / 选中态', 'primary-text', 'background', 3.0],
  ['侧栏上的强调色', 'primary-text', 'sidebar', 3.0],
  ['洗色底上的强调色', 'primary-text', 'accent', 3.0],
  ['成功状态', 'success', 'background', 3.0],
  ['警告状态', 'warning', 'background', 3.0],
  ['危险状态', 'destructive', 'background', 3.0],
  ['信息状态', 'info', 'background', 3.0],
  ['焦点环', 'ring', 'background', 3.0],
  ['代码 · 关键字', 'code-keyword', 'background', 3.0],
  ['代码 · 字符串', 'code-string', 'background', 3.0],
  ['代码 · 注释', 'code-comment', 'background', 3.0],
]

// ── 色彩空间 ──────────────────────────────────────────────────
type RGB = [number, number, number] // 0..1 线性前的 sRGB 分量
type OKLCH = { l: number; c: number; h: number; alpha: number }

const clamp01 = (v: number) => Math.min(1, Math.max(0, v))
const srgbToLinear = (v: number) => (v <= 0.04045 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4)
const linearToSrgb = (v: number) => (v <= 0.0031308 ? v * 12.92 : 1.055 * v ** (1 / 2.4) - 0.055)

function oklchToRgb({ l, c, h }: OKLCH): RGB {
  const hr = (h * Math.PI) / 180
  const a = c * Math.cos(hr)
  const b = c * Math.sin(hr)
  const l_ = (l + 0.3963377774 * a + 0.2158037573 * b) ** 3
  const m_ = (l - 0.1055613458 * a - 0.0638541728 * b) ** 3
  const s_ = (l - 0.0894841775 * a - 1.291485548 * b) ** 3
  const lr = 4.0767416621 * l_ - 3.3077115913 * m_ + 0.2309699292 * s_
  const lg = -1.2684380046 * l_ + 2.6097574011 * m_ - 0.3413193965 * s_
  const lb = -0.0041960863 * l_ - 0.7034186147 * m_ + 1.707614701 * s_
  return [clamp01(linearToSrgb(lr)), clamp01(linearToSrgb(lg)), clamp01(linearToSrgb(lb))]
}

function rgbToOklch([r, g, b]: RGB, alpha = 1): OKLCH {
  const lr = srgbToLinear(r)
  const lg = srgbToLinear(g)
  const lb = srgbToLinear(b)
  const l_ = Math.cbrt(0.4122214708 * lr + 0.5363325363 * lg + 0.0514459929 * lb)
  const m_ = Math.cbrt(0.2119034982 * lr + 0.6806995451 * lg + 0.1073969566 * lb)
  const s_ = Math.cbrt(0.0883024619 * lr + 0.2817188376 * lg + 0.6299787005 * lb)
  const L = 0.2104542553 * l_ + 0.793617785 * m_ - 0.0040720468 * s_
  const A = 1.9779984951 * l_ - 2.428592205 * m_ + 0.4505937099 * s_
  const B = 0.0259040371 * l_ + 0.7827717662 * m_ - 0.808675766 * s_
  const c = Math.hypot(A, B)
  let h = (Math.atan2(B, A) * 180) / Math.PI
  if (h < 0) h += 360
  return { l: L, c, h, alpha }
}

// ── 求值 ──────────────────────────────────────────────────────
/** 颜色统一以 OKLCH + alpha 承载；alpha < 1 表示尚未与背景合成 */
type Color = OKLCH

function parseHex(s: string): Color | null {
  const m = /^#([0-9a-f]{3,8})$/i.exec(s.trim())
  if (!m) return null
  let h = m[1]
  if (h.length === 3 || h.length === 4) h = [...h].map((ch) => ch + ch).join('')
  const num = (i: number) => parseInt(h.slice(i, i + 2), 16) / 255
  const alpha = h.length === 8 ? num(6) : 1
  return rgbToOklch([num(0), num(2), num(4)], alpha)
}

/** 求 `calc(...)`、`min(a, b)`、`50%`、`0.5`、以及相对色关键字 l/c/h */
function evalComponent(expr: string, base: Color | null, kind: 'l' | 'c' | 'h'): number {
  const src = expr.trim()
  const min = /^min\((.+),(.+)\)$/is.exec(src)
  if (min) return Math.min(evalComponent(min[1], base, kind), evalComponent(min[2], base, kind))
  const calc = /^calc\((.*)\)$/is.exec(src)
  const body = calc ? calc[1] : src

  const resolveAtom = (atom: string): number => {
    const a = atom.trim()
    if (a === 'l') return base ? base.l : 0
    if (a === 'c') return base ? base.c : 0
    if (a === 'h') return base ? base.h : 0
    if (a.endsWith('%')) {
      const v = parseFloat(a)
      return kind === 'l' ? v / 100 : kind === 'c' ? v / 100 : v
    }
    return parseFloat(a)
  }

  // tokens.css 的算式只出现 `x * k`、`x + k`、`x - k` 三种形式；min() 在上面单独处理
  const mul = /^(.+?)\s*\*\s*(.+)$/.exec(body)
  if (mul) return resolveAtom(mul[1]) * resolveAtom(mul[2])
  const add = /^(.+?)\s*\+\s*(.+)$/.exec(body)
  if (add) return resolveAtom(add[1]) + resolveAtom(add[2])
  const sub = /^(.+?)\s+-\s+(.+)$/.exec(body)
  if (sub) return resolveAtom(sub[1]) - resolveAtom(sub[2])
  return resolveAtom(body)
}

/** 拆顶层逗号 / 空格分隔，忽略括号内部 */
function splitTop(s: string, sep: RegExp): string[] {
  const out: string[] = []
  let depth = 0
  let cur = ''
  for (const ch of s) {
    if (ch === '(') depth++
    if (ch === ')') depth--
    if (depth === 0 && sep.test(ch)) {
      if (cur.trim()) out.push(cur.trim())
      cur = ''
    } else cur += ch
  }
  if (cur.trim()) out.push(cur.trim())
  return out
}

class Resolver {
  private cache = new Map<string, Color | null>()
  constructor(private vars: Map<string, string>) {}

  get(name: string): Color | null {
    if (this.cache.has(name)) return this.cache.get(name)!
    this.cache.set(name, null) // 断环
    const raw = this.vars.get(name)
    const v = raw == null ? null : this.evalColor(raw)
    this.cache.set(name, v)
    return v
  }

  /**
   * 先把 var() 全部替换成字面文本再求值。
   * 变量不只承载颜色——`--neutral-c` 这类是纯数字，出现在 oklch() 的分量位上，
   * 只按颜色去解析会得到 NaN。
   */
  private expand(s: string, depth = 0): string {
    if (depth > 16) return s
    return s.replace(
      /var\(\s*--([\w-]+)\s*(?:,([^()]*(?:\([^()]*\)[^()]*)*))?\)/g,
      (m, name: string, fb?: string) => {
        const raw = this.vars.get(name)
        if (raw != null) return this.expand(raw, depth + 1)
        return fb != null ? this.expand(fb, depth + 1) : m
      }
    )
  }

  evalColor(input: string): Color | null {
    const s = this.expand(input).trim()
    if (!s || s === 'transparent') return { l: 0, c: 0, h: 0, alpha: 0 }

    const hex = parseHex(s)
    if (hex) return hex

    const mix = /^color-mix\(\s*in\s+([\w-]+)\s*,([\s\S]*)\)$/i.exec(s)
    if (mix) {
      const [aRaw, bRaw] = splitTop(mix[2], /,/)
      const parse = (part: string) => {
        const pm = /^([\s\S]+?)\s+([\d.]+)%$/.exec(part.trim())
        return pm
          ? { color: this.evalColor(pm[1]), pct: parseFloat(pm[2]) / 100 }
          : { color: this.evalColor(part), pct: null as number | null }
      }
      const A = parse(aRaw)
      const B = parse(bRaw ?? 'transparent')
      const pa = A.pct ?? (B.pct != null ? 1 - B.pct : 0.5)
      if (!A.color || !B.color) return null
      // 与 transparent 混合等价于降 alpha；其余情况按分量线性插值
      const lerp = (x: number, y: number) => x * pa + y * (1 - pa)
      if (B.color.alpha === 0) return { ...A.color, alpha: A.color.alpha * pa }
      if (A.color.alpha === 0) return { ...B.color, alpha: B.color.alpha * (1 - pa) }
      if (mix[1].toLowerCase() === 'oklch') {
        return {
          l: lerp(A.color.l, B.color.l),
          c: lerp(A.color.c, B.color.c),
          h: lerp(A.color.h, B.color.h),
          alpha: lerp(A.color.alpha, B.color.alpha),
        }
      }
      const ra = oklchToRgb(A.color)
      const rb = oklchToRgb(B.color)
      return rgbToOklch(
        [lerp(ra[0], rb[0]), lerp(ra[1], rb[1]), lerp(ra[2], rb[2])],
        lerp(A.color.alpha, B.color.alpha)
      )
    }

    const ok = /^oklch\(([\s\S]*)\)$/i.exec(s)
    if (ok) {
      let body = ok[1].trim()
      let base: Color | null = null
      const from = /^from\s+([\s\S]+)$/i.exec(body)
      if (from) {
        const parts = splitTop(from[1], /\s/)
        base = this.evalColor(parts[0])
        body = parts.slice(1).join(' ')
      }
      const comps = splitTop(body, /\s/)
      if (comps.length < 3) return null
      return {
        l: evalComponent(comps[0], base, 'l'),
        c: evalComponent(comps[1], base, 'c'),
        h: evalComponent(comps[2], base, 'h'),
        alpha: base?.alpha ?? 1,
      }
    }
    return null
  }
}

/** 把 alpha < 1 的颜色合成到不透明背景上 */
function flatten(fg: Color, bg: Color): Color {
  if (fg.alpha >= 1) return fg
  const a = oklchToRgb(fg)
  const b = oklchToRgb(bg)
  const k = fg.alpha
  return rgbToOklch([a[0] * k + b[0] * (1 - k), a[1] * k + b[1] * (1 - k), a[2] * k + b[2] * (1 - k)])
}

function relLum(c: Color): number {
  const [r, g, b] = oklchToRgb(c)
  return 0.2126 * srgbToLinear(r) + 0.7152 * srgbToLinear(g) + 0.0722 * srgbToLinear(b)
}

function contrast(a: Color, b: Color): number {
  const x = relLum(a)
  const y = relLum(b)
  return (Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05)
}

const hexOf = (c: Color) =>
  '#' + oklchToRgb(c).map((v) => Math.round(v * 255).toString(16).padStart(2, '0')).join('')

// ── 解析 tokens.css ───────────────────────────────────────────
type Rule = { selector: string; decls: Array<[string, string]> }

function parseRules(css: string): Rule[] {
  const stripped = css.replace(/\/\*[\s\S]*?\*\//g, '')
  const rules: Rule[] = []
  const re = /([^{}]+)\{([^{}]*)\}/g
  let m: RegExpExecArray | null
  while ((m = re.exec(stripped))) {
    const selector = m[1].trim()
    const decls: Array<[string, string]> = []
    for (const d of m[2].split(';')) {
      const i = d.indexOf(':')
      if (i < 0) continue
      const prop = d.slice(0, i).trim()
      if (!prop.startsWith('--')) continue
      decls.push([prop.slice(2), d.slice(i + 1).trim()])
    }
    if (decls.length) rules.push({ selector, decls })
  }
  return rules
}

/**
 * 按 CSS 层叠顺序（特异性升序，同特异性按源码先后）挑出适用于当前组合的规则。
 * tokens.css 的选择器形态有限，这里显式枚举，避免引入完整的选择器引擎。
 */
function selectorApplies(sel: string, surface: string, accent: string, dark: boolean): boolean {
  return sel.split(',').some((oneRaw) => {
    const one = oneRaw.trim()
    if (one.endsWith(' body')) return false // body 上的氛围光不参与 token 求值
    if (/\.dark/.test(one) !== dark && /\.dark/.test(one)) return false
    const surfM = /\[data-surface='([\w-]+)'\]/.exec(one)
    if (surfM && surfM[1] !== surface) return false
    const accM = /\[data-accent='([\w-]+)'\]/.exec(one)
    if (accM && accM[1] !== accent) return false
    if (/\[data-accent\]/.test(one) && !accM && false) return false
    return true
  })
}

function specificity(sel: string): number {
  // 伪类 / 类 / 属性选择器同属特异性第二档，各计 10。
  // `:root` 与 `[data-accent='x']` 因此同分——同分时由源码顺序决胜，
  // 这一点必须与浏览器一致，否则脚本会漏判真实存在的层叠覆盖。
  const one = sel.split(',')[0].trim()
  let n = 0
  if (/:root/.test(one)) n += 10
  n += (one.match(/\.[a-z-]+/gi)?.length ?? 0) * 10
  n += (one.match(/\[[^\]]+\]/g)?.length ?? 0) * 10
  return n
}

function buildVars(rules: Rule[], surface: string, accent: string, dark: boolean): Map<string, string> {
  const applicable = rules
    .map((r, i) => ({ r, i, spec: specificity(r.selector) }))
    .filter((x) => selectorApplies(x.r.selector, surface, accent, dark))
    .sort((a, b) => a.spec - b.spec || a.i - b.i)
  const vars = new Map<string, string>()
  for (const { r } of applicable) for (const [k, v] of r.decls) vars.set(k, v)
  return vars
}

// ── 主流程 ────────────────────────────────────────────────────
const css = readFileSync(TOKENS_CSS, 'utf8')
const rules = parseRules(css)

// `--dump <surface> <accent> <light|dark> [token...]`：输出解算色值，
// 用于与浏览器实际渲染结果交叉比对，确认上面这套求值器没有漂移。
const dumpIdx = process.argv.indexOf('--dump')
if (dumpIdx >= 0) {
  const [surface, accent, mode] = process.argv.slice(dumpIdx + 1, dumpIdx + 4)
  const names = process.argv.slice(dumpIdx + 4)
  const resolver = new Resolver(buildVars(rules, surface, accent, mode === 'dark'))
  const bg = resolver.get('background')!
  const wanted = names.length
    ? names
    : ['background', 'foreground', 'card', 'popover', 'primary', 'primary-foreground',
       'primary-text', 'secondary', 'muted', 'muted-foreground', 'accent', 'border', 'sidebar']
  for (const n of wanted) {
    const c = resolver.get(n)
    console.log(`${n}\t${c ? hexOf(flatten(c, bg)) : 'UNRESOLVED'}`)
  }
  process.exit(0)
}

let failures = 0
let checks = 0
const missing = new Set<string>()

for (const surface of SURFACES) {
  for (const accent of ACCENTS) {
    for (const mode of MODES) {
      const dark = mode === 'dark'
      const resolver = new Resolver(buildVars(rules, surface, accent, dark))
      const bg = resolver.get('background')
      if (!bg) {
        console.error(`✗ ${surface}/${accent}/${mode}: --background 无法求值`)
        failures++
        continue
      }
      for (const [label, fgName, bgName, min] of CONTRACT) {
        const fgRaw = resolver.get(fgName)
        const bgRaw = resolver.get(bgName)
        if (!fgRaw || !bgRaw) {
          missing.add(!fgRaw ? fgName : bgName)
          continue
        }
        const bgC = flatten(bgRaw, bg)
        const fgC = flatten(fgRaw, bgC)
        const ratio = contrast(fgC, bgC)
        checks++
        if (ratio < min) {
          failures++
          console.error(
            `✗ ${surface}/${accent}/${mode}  ${label}: ${ratio.toFixed(2)} < ${min}` +
              `  (--${fgName} ${hexOf(fgC)} on --${bgName} ${hexOf(bgC)})`
          )
        }
      }
    }
  }
}

if (missing.size) {
  console.error(`\n✗ 以下 token 未能求值：${[...missing].map((m) => '--' + m).join(', ')}`)
  failures += missing.size
}

const total = SURFACES.length * ACCENTS.length * MODES.length
if (failures === 0) {
  console.log(`✓ 对比度契约全部通过 — ${total} 种组合 × ${CONTRACT.length} 项 = ${checks} 次断言`)
  process.exit(0)
} else {
  console.error(`\n✗ ${failures} 项未达标（共 ${checks} 次断言，${total} 种组合）`)
  process.exit(1)
}
