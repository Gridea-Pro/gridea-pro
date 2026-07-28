/**
 * 读取语义 token 的实际颜色值。
 *
 * 给需要「具体色值」而非 CSS 变量的第三方组件用（Monaco / Mermaid / CodeMirror
 * 都只接受 hex 或 rgb 字面量，无法消费 var()）。tokens.css 用 OKLCH 与
 * color-mix 派生，浏览器解析后的形式不固定，故统一走 canvas 光栅化取真值。
 */

let ctx: CanvasRenderingContext2D | null = null
let probe: HTMLElement | null = null

function getCtx(): CanvasRenderingContext2D | null {
  if (ctx) return ctx
  const canvas = document.createElement('canvas')
  canvas.width = canvas.height = 1
  ctx = canvas.getContext('2d', { willReadFrequently: true })
  return ctx
}

function getProbe(): HTMLElement {
  if (probe && probe.isConnected) return probe
  probe = document.createElement('div')
  probe.style.display = 'none'
  document.body.appendChild(probe)
  return probe
}

const toHex = (n: number) => n.toString(16).padStart(2, '0')

/**
 * @param token   变量名，不含 `--` 前缀
 * @param base    半透明 token 的合成底色（如 glass 外观的 --card），默认白
 * @returns       `#rrggbb`
 */
export function readTokenHex(token: string, base = '#ffffff'): string {
  const c = getCtx()
  if (!c) return base
  const el = getProbe()
  el.style.color = ''
  el.style.color = `var(--${token})`
  const resolved = getComputedStyle(el).color
  if (!resolved) return base

  c.fillStyle = base
  c.fillRect(0, 0, 1, 1)
  c.fillStyle = resolved
  c.fillRect(0, 0, 1, 1)
  const [r, g, b] = c.getImageData(0, 0, 1, 1).data
  return `#${toHex(r)}${toHex(g)}${toHex(b)}`
}

/** 同 readTokenHex，但返回不带 `#` 的六位值（Monaco 的 token rules 要求此格式） */
export function readTokenRaw(token: string, base = '#ffffff'): string {
  return readTokenHex(token, base).slice(1)
}

/** 读取一组 token，返回 { token: '#rrggbb' } */
export function readTokens(tokens: string[], base?: string): Record<string, string> {
  const bg = base ?? readTokenHex('background')
  return Object.fromEntries(tokens.map((t) => [t, readTokenHex(t, bg)]))
}

/** 在 hex 上叠加透明度，产出 Monaco / CodeMirror 可用的 8 位色值 */
export function withAlpha(hex: string, alpha: number): string {
  const a = Math.round(Math.min(Math.max(alpha, 0), 1) * 255)
  return `${hex}${toHex(a)}`
}
