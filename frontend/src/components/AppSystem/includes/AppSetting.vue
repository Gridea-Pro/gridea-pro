<template>
  <div class="p-2">
    <div class="mb-8">
      <div class="text-sm font-medium text-[var(--text-secondary)] mb-4">主题模式</div>
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
      <div class="text-sm font-medium text-[var(--text-secondary)] mb-4">主题色</div>
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
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useThemeStore, ThemeMode, ThemeColor } from '@/stores/theme'
import { SunIcon, MoonIcon, ComputerDesktopIcon, CheckIcon } from '@heroicons/vue/24/outline'

const themeStore = useThemeStore()

const mode = computed({
  get: () => themeStore.mode,
  set: (val) => themeStore.setMode(val)
})

const theme = computed({
  get: () => themeStore.theme,
  set: (val) => themeStore.setTheme(val)
})

const modeOptions: { label: string, value: ThemeMode, icon: any }[] = [
  { label: '浅色', value: 'light', icon: SunIcon },
  { label: '深色', value: 'dark', icon: MoonIcon },
  { label: '跟随系统', value: 'system', icon: ComputerDesktopIcon },
]

const themeColors: { label: string, value: ThemeColor, primary: string }[] = [
  { label: '珍珠白', value: 'default', primary: '#ffffff' },
  { label: '极客蓝', value: 'blue', primary: '#096dd9' },
  { label: '日落暖', value: 'warm', primary: '#D47B4A' },
  { label: '樱花粉', value: 'sakura', primary: '#FF77A9' },
  { label: '暮山紫', value: 'twilight', primary: '#722ED1' },
  { label: '毛玻璃', value: 'glass', primary: '#E0EAFC' },
]
</script>
