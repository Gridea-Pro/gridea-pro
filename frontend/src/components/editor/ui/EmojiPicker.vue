<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <!-- 直接 button 子元素：as-child 包 slot 间接层会导致触发 props 丢失（点击无反应的根因） -->
      <button
        type="button"
        class="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-accent-foreground"
        :title="t('editor.emoji.title')"
        :aria-label="t('editor.emoji.title')"
        @mousedown.prevent
      >
        <Smile class="h-[18px] w-[18px]" />
      </button>
    </PopoverTrigger>
    <PopoverContent
      align="start"
      :side-offset="6"
      class="w-auto border-border p-0"
      @open-auto-focus.prevent
    >
      <EmojiPicker
        :native="true"
        :theme="pickerTheme"
        :hide-search="false"
        :disable-skin-tones="false"
        @select="onSelect"
      />
    </PopoverContent>
  </Popover>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { IconMoodSmile as Smile } from '@tabler/icons-vue'
import EmojiPicker, { type EmojiExt } from 'vue3-emoji-picker'
import 'vue3-emoji-picker/css'
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover'
import { useThemeStore } from '@/stores/theme'

const emit = defineEmits<{
  /** Emitted with the selected emoji as a unicode character. */
  select: [emoji: string]
}>()

const { t } = useI18n()
const themeStore = useThemeStore()

const open = ref(false)

const pickerTheme = computed<'light' | 'dark'>(() =>
  themeStore.isDark ? 'dark' : 'light',
)

function onSelect(e: EmojiExt) {
  // `i` is the emoji unicode character.
  emit('select', e.i)
  open.value = false
}
</script>

<style scoped>
/* Bridge the picker chrome to Gridea's theme tokens so it follows light/dark.
   vue3-emoji-picker exposes these --v3-picker-* custom properties. */
:deep(.v3-emoji-picker) {
  --v3-picker-bg: var(--editor-popover);
  --v3-picker-fg: var(--editor-popover-fg);
  --v3-picker-border: var(--editor-border);
  --v3-picker-emoji-hover: var(--editor-hover);
  --v3-picker-input-bg: var(--editor-secondary);
  --v3-picker-input-border: var(--editor-border);
  --v3-picker-input-focus-border: var(--editor-accent);
  border: none;
  box-shadow: none;
}
</style>
