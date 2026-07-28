import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

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

export const EDITOR_FONT_FAMILY_DEFAULT =
  'ui-monospace, Menlo, Monaco, "Cascadia Code", "Segoe UI Mono", Consolas, "Courier New", monospace'

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
  const editorFontFamily = ref<string>(
    localStorage.getItem('app_editor_font_family') || EDITOR_FONT_FAMILY_DEFAULT
  )

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

  function setEditorFontFamily(value: string) {
    editorFontFamily.value = value
    localStorage.setItem('app_editor_font_family', value)
  }

  function initTheme() {
    applyTheme()
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
    editorFontFamily,
    isDark,
    setMode,
    setSurface,
    setAccent,
    setEditorFontFamily,
    applyTheme,
    initTheme,
  }
})
