<template>
  <div class="p-2">
    <div class="mb-8">
      <div class="text-sm font-medium text-[var(--text-secondary)] mb-4">{{ t('preferences.themeMode') }}</div>
      <div class="grid grid-cols-3 gap-4">
        <div
          v-for="item in modeOptions"
          :key="item.value"
          class="flex flex-col items-center justify-center p-4 bg-[var(--bg-card)] border border-[var(--color-border)] rounded-lg cursor-pointer transition-all duration-200 hover:border-[var(--color-primary)] hover:bg-[var(--bg-sidebar)]"
          :class="{ 'border-[var(--color-primary)] bg-[var(--bg-sidebar)]': mode === item.value }"
          @click="mode = item.value"
        >
          <component
            :is="item.icon"
            class="w-6 h-6 mb-2 text-[var(--text-secondary)]"
            :class="{ '!text-[var(--color-primary)]': mode === item.value }"
          />
          <span
            class="text-[13px] text-[var(--text-primary)]"
            :class="{ '!text-[var(--color-primary)]': mode === item.value }"
          >{{ item.label }}</span>
        </div>
      </div>
    </div>

    <div class="mb-8">
      <div class="text-sm font-medium text-[var(--text-secondary)] mb-4">{{ t('preferences.themeColor') }}</div>
      <div class="grid grid-cols-[repeat(auto-fill,minmax(80px,1fr))] gap-4">
        <div
          v-for="color in themeColors"
          :key="color.value"
          class="flex flex-col items-center cursor-pointer group"
          @click="theme = color.value"
        >
          <div
            class="w-12 h-12 rounded-full mb-2 flex items-center justify-center transition-transform duration-200 shadow-sm group-hover:scale-110"
            :style="{ background: color.primary }"
          >
            <CheckIcon v-if="theme === color.value" class="w-6 h-6 text-white" />
          </div>
          <span
            class="text-xs text-[var(--text-secondary)] group-hover:text-[var(--text-primary)]"
            :class="{ '!text-[var(--color-primary)] font-medium': theme === color.value }"
          >{{ color.label }}</span>
        </div>
      </div>
    </div>

    <div>
      <div class="text-sm font-medium text-[var(--text-secondary)] mb-4">{{ t('preferences.editorFontFamily') }}</div>
      <Input
        v-model="editorFontFamily"
        :placeholder="t('preferences.editorFontFamilyPlaceholder')"
        class="max-w-md font-mono text-sm"
      />
      <div class="text-xs text-[var(--text-secondary)] mt-2">{{ t('preferences.editorFontFamilyHint') }}</div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useThemeStore, ThemeMode, ThemeColor } from '@/stores/theme'
import { SunIcon, MoonIcon, ComputerDesktopIcon, CheckIcon } from '@heroicons/vue/24/outline'
import { Input } from '@/components/ui/input'

const { t } = useI18n()
const themeStore = useThemeStore()

const mode = computed({
  get: () => themeStore.mode,
  set: (val) => themeStore.setMode(val)
})

const theme = computed({
  get: () => themeStore.theme,
  set: (val) => themeStore.setTheme(val)
})

const editorFontFamily = computed({
  get: () => themeStore.editorFontFamily,
  set: (val) => themeStore.setEditorFontFamily(val),
})

const modeOptions = computed(() => [
  { label: t('preferences.light'), value: 'light' as ThemeMode, icon: SunIcon },
  { label: t('preferences.dark'), value: 'dark' as ThemeMode, icon: MoonIcon },
  { label: t('preferences.followSystem'), value: 'system' as ThemeMode, icon: ComputerDesktopIcon },
])

const themeColors = computed(() => [
  { label: t('preferences.colorDefault'), value: 'default' as ThemeColor, primary: '#ffffff' },
  { label: t('preferences.colorBlue'), value: 'blue' as ThemeColor, primary: '#096dd9' },
  { label: t('preferences.colorWarm'), value: 'warm' as ThemeColor, primary: '#D47B4A' },
  { label: t('preferences.colorSakura'), value: 'sakura' as ThemeColor, primary: '#FF77A9' },
  { label: t('preferences.colorTwilight'), value: 'twilight' as ThemeColor, primary: '#722ED1' },
  { label: t('preferences.colorGlass'), value: 'glass' as ThemeColor, primary: '#E0EAFC' },
])
</script>
