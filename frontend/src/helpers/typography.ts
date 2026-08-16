/**
 * 编辑器正文排版的字号真相源。
 *
 * 契约：正文与六级标题的字号由本模块统一派生，样式（--type-*）与工具栏字号下拉
 * 共用同一份返回值。此前两边各写各的（CSS 用 em 比例、下拉写死 26/22/20/18），
 * 四级标题的标注值与实际渲染值全部对不上——选中 H2 显示 22、实际 24.8，
 * 点一下"22"反而把标题改小。共用一份派生结果后，结构上不可能再分叉。
 *
 * 字号一律取整：设置面板允许改基准字号，比例派生会得出 27.648px 这类小数，
 * 而字号下拉必须显示整数。取整在派生时一次做掉，样式和 UI 拿到的是同一批整数。
 */

/**
 * 六级标题相对正文的比例。h5/h6 = 1 是有意的，见 tokens.css 中的说明。
 * index.html 的防闪脚本内联了同一份比例，由 scripts/check-typography-sync.mts 守护同步。
 */
export const HEADING_RATIOS = [1.75, 1.44, 1.2, 1.0625, 1, 1] as const
/** 文章标题（整篇的题目）必须压过正文 h1，否则层级倒挂。 */
export const TITLE_RATIO = 2

export interface TypeScale {
  base: number
  /** 文章标题，恒大于 headings[0] */
  title: number
  /** 下标 0-5 对应 h1-h6 */
  headings: number[]
}

/** 由基准字号派生整数字号阶。base 非法时回落到 16。 */
export function typeScale(base: number): TypeScale {
  const b = Number.isFinite(base) && base > 0 ? Math.round(base) : 16
  return {
    base: b,
    title: Math.round(b * TITLE_RATIO),
    headings: HEADING_RATIOS.map((r) => Math.round(b * r)),
  }
}

/** 取某一级标题的字号；level 越界时返回正文字号。 */
export function headingSize(scale: TypeScale, level: number): number {
  return scale.headings[level - 1] ?? scale.base
}

/**
 * 把字号阶写进 CSS 变量。设置面板改基准字号后调用，
 * 样式与工具栏会在同一帧拿到同一批新值。
 */
export function applyTypeScale(base: number, root: HTMLElement = document.documentElement): TypeScale {
  const scale = typeScale(base)
  root.style.setProperty('--type-base', `${scale.base}px`)
  root.style.setProperty('--type-title', `${scale.title}px`)
  scale.headings.forEach((size, i) => root.style.setProperty(`--type-h${i + 1}`, `${size}px`))
  return scale
}

/** 把正文字体写进 CSS 变量。空值表示跟随默认系统栈，此时清除内联覆盖。 */
export function applyTypeFamily(family: string, root: HTMLElement = document.documentElement): void {
  if (family) root.style.setProperty('--type-family', family)
  else root.style.removeProperty('--type-family')
}

/** 可选的正文基准字号。范围两端由「一屏能读」和「不至于糊成一片」定，不做无级调节。 */
export const PROSE_SIZES = [14, 15, 16, 17, 18, 20] as const
export const PROSE_SIZE_DEFAULT = 16

export interface ProseFontOption {
  id: string
  labelKey: string
  /** 完整字体栈；空串表示跟随界面字体（--font-sans） */
  stack: string
  /** 可用性探测用的首选字体名；空串表示无需探测（系统默认永远可用） */
  probe: string
}

/**
 * 正文字体候选。只列**系统自带**字体（决策 2：不打包字体文件），
 * 且每项都要能通过 isFontAvailable 探测——`--font-sans` 栈首那两个
 * 本机根本不存在的字体名就是反面教材，写了永远解析不到、白白掩盖真实渲染结果。
 */
export const PROSE_FONTS: ReadonlyArray<ProseFontOption> = [
  { id: 'system', labelKey: 'preferences.proseFontSystem', stack: '', probe: '' },
  {
    id: 'pingfang',
    labelKey: 'preferences.proseFontPingFang',
    stack: '"PingFang SC", "Hiragino Sans GB", sans-serif',
    probe: 'PingFang SC',
  },
  {
    id: 'yahei',
    labelKey: 'preferences.proseFontYaHei',
    stack: '"Microsoft YaHei", sans-serif',
    probe: 'Microsoft YaHei',
  },
  {
    id: 'songti',
    labelKey: 'preferences.proseFontSongti',
    stack: '"Songti SC", "SimSun", serif',
    probe: 'Songti SC',
  },
  {
    id: 'simsun',
    labelKey: 'preferences.proseFontSimSun',
    stack: '"SimSun", "Songti SC", serif',
    probe: 'SimSun',
  },
  {
    id: 'wenkai',
    labelKey: 'preferences.proseFontWenKai',
    stack: '"LXGW WenKai", "LXGW WenKai Screen", serif',
    probe: 'LXGW WenKai',
  },
]

/**
 * 探测某个字体本机是否真的存在。
 * 用 canvas 宽度对比而非 document.fonts.check —— 后者对未经 @font-face 声明的
 * 系统字体判定不可靠，装没装都可能返回 true。
 */
export function isFontAvailable(family: string): boolean {
  if (!family) return true
  if (typeof document === 'undefined') return true
  const ctx = document.createElement('canvas').getContext('2d')
  if (!ctx) return true
  // 中西文混排样本：只用西文的话，缺失的中文字体会因为西文回退一致而误判为可用
  const sample = 'MMMWWWlliI0Oo中文字体测试'
  return ['monospace', 'serif', 'sans-serif'].some((generic) => {
    ctx.font = `72px ${generic}`
    const base = ctx.measureText(sample).width
    ctx.font = `72px "${family}", ${generic}`
    return ctx.measureText(sample).width !== base
  })
}
