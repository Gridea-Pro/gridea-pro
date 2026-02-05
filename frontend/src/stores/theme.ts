import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export type ThemeMode = 'light' | 'dark' | 'system'
export type ThemeColor = 'default' | 'blue' | 'warm' | 'sakura' | 'twilight' | 'glass'

export const useThemeStore = defineStore('theme', () => {
  // State
  const mode = ref<ThemeMode>(
    (localStorage.getItem('app_theme_mode') as ThemeMode) || 'system'
  )
  const theme = ref<ThemeColor>(
    (localStorage.getItem('app_theme_color') as ThemeColor) || 'default'
  )
  const systemIsDark = ref(
    window.matchMedia('(prefers-color-scheme: dark)').matches
  )

  // Getters
  const isDark = computed(() => {
    if (mode.value === 'system') {
      return systemIsDark.value
    }
    return mode.value === 'dark'
  })

  const antDesignTheme = computed(() => {
    const colorMap: Record<ThemeColor, string> = {
      default: '#1b1b18',
      blue: '#096dd9',
      warm: '#D47B4A',
      sakura: '#FF77A9',
      twilight: '#722ED1',
      glass: '#E0EAFC'
    }
    return {
      token: {
        colorPrimary: colorMap[theme.value] || '#096dd9',
      }
    }
  })

  // Actions
  function applyTheme() {
    const html = document.documentElement

    if (isDark.value) {
      html.classList.add('dark')
    } else {
      html.classList.remove('dark')
    }

    html.setAttribute('data-theme', theme.value)
  }

  function setMode(newMode: ThemeMode) {
    mode.value = newMode
    localStorage.setItem('app_theme_mode', newMode)
    applyTheme()
  }

  function setTheme(newTheme: ThemeColor) {
    theme.value = newTheme
    localStorage.setItem('app_theme_color', newTheme)
    applyTheme()
  }

  function initTheme() {
    applyTheme()
    // Listen for system changes
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
    // State
    mode,
    theme,
    systemIsDark,
    // Getters
    isDark,
    antDesignTheme,
    // Actions
    setMode,
    setTheme,
    applyTheme,
    initTheme
  }
})
