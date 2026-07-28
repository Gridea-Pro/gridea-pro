<template>
  <div class="p-2">
    <div class="mb-8">
      <div class="text-sm font-medium text-muted-foreground mb-4">{{ t('preferences.themeMode') }}</div>
      <div class="grid grid-cols-3 gap-4">
        <button
          v-for="item in modeOptions"
          :key="item.value"
          type="button"
          class="flex flex-col items-center justify-center p-4 bg-card border rounded-xl cursor-pointer transition-colors hover:border-primary/50"
          :class="mode === item.value ? 'border-primary bg-accent' : 'border-border'"
          @click="mode = item.value"
        >
          <component
            :is="item.icon"
            class="w-6 h-6 mb-2"
            :class="mode === item.value ? 'text-primary-text' : 'text-muted-foreground'"
          />
          <span
            class="text-[13px]"
            :class="mode === item.value ? 'text-primary-text font-medium' : 'text-foreground'"
          >{{ item.label }}</span>
        </button>
      </div>
    </div>

    <div class="mb-8">
      <div class="text-sm font-medium text-muted-foreground mb-4">{{ t('preferences.themeSurface') }}</div>
      <div class="grid grid-cols-3 gap-4">
        <button
          v-for="item in surfaceOptions"
          :key="item.value"
          type="button"
          class="flex flex-col items-center justify-center p-4 bg-card border rounded-xl cursor-pointer transition-colors hover:border-primary/50"
          :class="surface === item.value ? 'border-primary bg-accent' : 'border-border'"
          @click="surface = item.value"
        >
          <component
            :is="item.icon"
            class="w-6 h-6 mb-2"
            :class="surface === item.value ? 'text-primary-text' : 'text-muted-foreground'"
          />
          <span
            class="text-[13px]"
            :class="surface === item.value ? 'text-primary-text font-medium' : 'text-foreground'"
          >{{ item.label }}</span>
        </button>
      </div>
    </div>

    <div class="mb-8">
      <div class="text-sm font-medium text-muted-foreground mb-4">{{ t('preferences.themeAccent') }}</div>
      <div class="grid grid-cols-[repeat(auto-fill,minmax(80px,1fr))] gap-4">
        <button
          v-for="item in accentOptions"
          :key="item.value"
          type="button"
          class="flex flex-col items-center cursor-pointer group bg-transparent border-0 p-0"
          @click="accent = item.value"
        >
          <!-- 色卡挂 data-accent 后直接取该强调色的种子，无需另存一份预览色值 -->
          <span
            :data-accent="item.value"
            class="w-12 h-12 rounded-full mb-2 flex items-center justify-center transition-transform duration-200 shadow-card group-hover:scale-110"
            :style="{ background: 'var(--seed-accent)' }"
          >
            <CheckIcon v-if="accent === item.value" class="w-6 h-6 text-white" />
          </span>
          <span
            class="text-xs"
            :class="accent === item.value ? 'text-primary-text font-medium' : 'text-muted-foreground group-hover:text-foreground'"
          >{{ item.label }}</span>
        </button>
      </div>
    </div>

    <div>
      <div class="text-sm font-medium text-muted-foreground mb-4">{{ t('preferences.editorFontFamily') }}</div>
      <Input
        v-model="editorFontFamily"
        :placeholder="t('preferences.editorFontFamilyPlaceholder')"
        class="max-w-md font-mono text-sm"
      />
      <div class="text-xs text-muted-foreground mt-2">{{ t('preferences.editorFontFamilyHint') }}</div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  useThemeStore,
  SURFACE_REGISTRY,
  ACCENT_REGISTRY,
  type ThemeMode,
  type ThemeSurface,
  type ThemeAccent,
} from '@/stores/theme'
import {
  SunIcon,
  MoonIcon,
  ComputerDesktopIcon,
  CheckIcon,
  Square2StackIcon,
  DocumentTextIcon,
  SparklesIcon,
} from '@heroicons/vue/24/outline'
import { Input } from '@/components/ui/input'

const { t } = useI18n()
const themeStore = useThemeStore()

const mode = computed({
  get: () => themeStore.mode,
  set: (val: ThemeMode) => themeStore.setMode(val),
})

const surface = computed({
  get: () => themeStore.surface,
  set: (val: ThemeSurface) => themeStore.setSurface(val),
})

const accent = computed({
  get: () => themeStore.accent,
  set: (val: ThemeAccent) => themeStore.setAccent(val),
})

const editorFontFamily = computed({
  get: () => themeStore.editorFontFamily,
  set: (val: string) => themeStore.setEditorFontFamily(val),
})

const modeOptions = computed(() => [
  { label: t('preferences.light'), value: 'light' as ThemeMode, icon: SunIcon },
  { label: t('preferences.dark'), value: 'dark' as ThemeMode, icon: MoonIcon },
  { label: t('preferences.followSystem'), value: 'system' as ThemeMode, icon: ComputerDesktopIcon },
])

const SURFACE_ICONS = {
  pure: Square2StackIcon,
  paper: DocumentTextIcon,
  glass: SparklesIcon,
} as const

const surfaceOptions = computed(() =>
  SURFACE_REGISTRY.map((s) => ({ value: s.value, label: t(s.labelKey), icon: SURFACE_ICONS[s.value] }))
)

const accentOptions = computed(() =>
  ACCENT_REGISTRY.map((a) => ({ value: a.value, label: t(a.labelKey) }))
)
</script>
