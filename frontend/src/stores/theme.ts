import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  PROSE_FONTS,
  PROSE_SIZES,
  PROSE_SIZE_DEFAULT,
  applyTypeScale,
  applyTypeFamily,
} from '@/helpers/typography'

export type ThemeMode = 'light' | 'dark' | 'system'
/** 外观主题：管底色质感 */
export type ThemeSurface = 'pure' | 'paper' | 'glass'
/** 强调色：管主题色 */
export type ThemeAccent = 'green' | 'rose' | 'sakura' | 'sunset' | 'amber' | 'cyan' | 'blue' | 'purple'

export const STORAGE_KEYS = {
  mode: 'app_theme_mode',
  surface: 'app_theme_surface',
  accent: 'app_theme_accent',
  /** v1 的一维主题色，仅用于一次性迁移 */
  legacyColor: 'app_theme_color',
  /** 正文基准字号；行距、段距、标题字阶全部由它派生 */
  proseSize: 'app_prose_font_size',
  /** 正文字体，取值为 PROSE_FONTS 的 id */
  proseFont: 'app_prose_font_family',
} as const

export const DEFAULT_SURFACE: ThemeSurface = 'pure'
export const DEFAULT_ACCENT: ThemeAccent = 'green'

/**
 * v1 的六个捆绑主题 → v2 的（外观, 强调色）。
 * index.html 的防闪脚本内联了同一份映射，改此处务必同步。
 */
export const LEGACY_THEME_MAP: Record<string, [ThemeSurface, ThemeAccent]> = {
  default: ['pure', 'green'],
  blue: ['pure', 'blue'],
  warm: ['paper', 'sunset'],
  sakura: ['pure', 'sakura'],
  twilight: ['pure', 'purple'],
  glass: ['glass', 'cyan'],
}

export const SURFACE_REGISTRY: ReadonlyArray<{ value: ThemeSurface; labelKey: string }> = [
  { value: 'pure', labelKey: 'preferences.surfacePure' },
  { value: 'paper', labelKey: 'preferences.surfacePaper' },
  { value: 'glass', labelKey: 'preferences.surfaceGlass' },
]

/** 色卡预览色不在此声明——模板给元素挂 data-accent 后直接取 CSS 的 --seed-accent */
export const ACCENT_REGISTRY: ReadonlyArray<{ value: ThemeAccent; labelKey: string }> = [
  { value: 'green', labelKey: 'preferences.accentGreen' },
  { value: 'rose', labelKey: 'preferences.accentRose' },
  { value: 'sakura', labelKey: 'preferences.accentSakura' },
  { value: 'sunset', labelKey: 'preferences.accentSunset' },
  { value: 'amber', labelKey: 'preferences.accentAmber' },
  { value: 'cyan', labelKey: 'preferences.accentCyan' },
  { value: 'blue', labelKey: 'preferences.accentBlue' },
  { value: 'purple', labelKey: 'preferences.accentPurple' },
]

const SURFACES = SURFACE_REGISTRY.map((s) => s.value)
const ACCENTS = ACCENT_REGISTRY.map((a) => a.value)

/** 读取正文字号，非法值一律回落到默认，避免脏数据把正文撑成天文数字 */
function readProseSize(): number {
  const raw = Number(localStorage.getItem(STORAGE_KEYS.proseSize))
  return (PROSE_SIZES as readonly number[]).includes(raw) ? raw : PROSE_SIZE_DEFAULT
}

function readProseFont(): string {
  const id = localStorage.getItem(STORAGE_KEYS.proseFont) || 'system'
  return PROSE_FONTS.some((f) => f.id === id) ? id : 'system'
}

/** 读取外观 + 强调色，必要时从 v1 的单一主题色迁移 */
function readSurfaceAndAccent(): [ThemeSurface, ThemeAccent] {
  const storedSurface = localStorage.getItem(STORAGE_KEYS.surface) as ThemeSurface | null
  const storedAccent = localStorage.getItem(STORAGE_KEYS.accent) as ThemeAccent | null
  if (storedSurface && storedAccent && SURFACES.includes(storedSurface) && ACCENTS.includes(storedAccent)) {
    return [storedSurface, storedAccent]
  }

  const legacy = localStorage.getItem(STORAGE_KEYS.legacyColor)
  const migrated = legacy ? LEGACY_THEME_MAP[legacy] : undefined
  const [surface, accent] = migrated ?? [DEFAULT_SURFACE, DEFAULT_ACCENT]

  localStorage.setItem(STORAGE_KEYS.surface, surface)
  localStorage.setItem(STORAGE_KEYS.accent, accent)
  localStorage.removeItem(STORAGE_KEYS.legacyColor)
  return [surface, accent]
}

export const useThemeStore = defineStore('theme', () => {
  const [initialSurface, initialAccent] = readSurfaceAndAccent()

  const mode = ref<ThemeMode>((localStorage.getItem(STORAGE_KEYS.mode) as ThemeMode) || 'system')
  const surface = ref<ThemeSurface>(initialSurface)
  const accent = ref<ThemeAccent>(initialAccent)
  const systemIsDark = ref(window.matchMedia('(prefers-color-scheme: dark)').matches)
  const proseFontSize = ref<number>(readProseSize())
  const proseFontId = ref<string>(readProseFont())

  const isDark = computed(() => (mode.value === 'system' ? systemIsDark.value : mode.value === 'dark'))

  function applyTheme() {
    const html = document.documentElement
    html.classList.toggle('dark', isDark.value)
    html.setAttribute('data-surface', surface.value)
    html.setAttribute('data-accent', accent.value)
  }

  function setMode(newMode: ThemeMode) {
    mode.value = newMode
    localStorage.setItem(STORAGE_KEYS.mode, newMode)
    applyTheme()
  }

  function setSurface(newSurface: ThemeSurface) {
    surface.value = newSurface
    localStorage.setItem(STORAGE_KEYS.surface, newSurface)
    applyTheme()
  }

  function setAccent(newAccent: ThemeAccent) {
    accent.value = newAccent
    localStorage.setItem(STORAGE_KEYS.accent, newAccent)
    applyTheme()
  }

  /** 应用排版偏好。字号一变，行距 / 段距 / 标题字阶全部按比例联动。 */
  function applyTypography() {
    applyTypeScale(proseFontSize.value)
    applyTypeFamily(PROSE_FONTS.find((f) => f.id === proseFontId.value)?.stack ?? '')
  }

  function setProseFontSize(size: number) {
    if (!(PROSE_SIZES as readonly number[]).includes(size)) return
    proseFontSize.value = size
    localStorage.setItem(STORAGE_KEYS.proseSize, String(size))
    applyTypography()
  }

  function setProseFontId(id: string) {
    if (!PROSE_FONTS.some((f) => f.id === id)) return
    proseFontId.value = id
    localStorage.setItem(STORAGE_KEYS.proseFont, id)
    applyTypography()
  }

  function initTheme() {
    applyTheme()
    applyTypography()
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    systemIsDark.value = mediaQuery.matches
    mediaQuery.addEventListener('change', (e) => {
      systemIsDark.value = e.matches
      if (mode.value === 'system') {
        applyTheme()
      }
    })
  }

  return {
    mode,
    surface,
    accent,
    systemIsDark,
    proseFontSize,
    proseFontId,
    isDark,
    setMode,
    setSurface,
    setAccent,
    setProseFontSize,
    setProseFontId,
    applyTheme,
    applyTypography,
    initTheme,
  }
})
