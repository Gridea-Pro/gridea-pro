<template>
  <Popover :open="isOpen" @update:open="onOpenChange">
    <PopoverTrigger as-child>
      <slot name="trigger">
        <Button
          type="button"
          variant="ghost"
          size="icon"
          class="h-8 w-8 text-muted-foreground hover:text-foreground"
          :title="t('editor.emoji.title')"
          :aria-label="t('editor.emoji.title')"
          @mousedown.prevent
        >
          <Smile class="h-[18px] w-[18px]" />
        </Button>
      </slot>
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
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { IconMoodSmile as Smile } from '@tabler/icons-vue'
import EmojiPicker, { type EmojiExt } from 'vue3-emoji-picker'
import 'vue3-emoji-picker/css'
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover'
import { Button } from '@/components/ui/button'
import { useThemeStore } from '@/stores/theme'

const props = defineProps<{
  /** Optional controlled open state. Omit to let the Popover self-manage. */
  open?: boolean
}>()

const emit = defineEmits<{
  /** Emitted with the selected emoji as a unicode character. */
  select: [emoji: string]
  /** Two-way binding for the open state when `open` is provided. */
  'update:open': [value: boolean]
}>()

const { t } = useI18n()
const themeStore = useThemeStore()

// Self-managed open state used when `open` prop is not provided.
const internalOpen = ref(false)

const isOpen = computed(() =>
  props.open === undefined ? internalOpen.value : props.open,
)

watch(
  () => props.open,
  (value) => {
    if (value !== undefined) internalOpen.value = value
  },
)

const pickerTheme = computed<'light' | 'dark'>(() =>
  themeStore.isDark ? 'dark' : 'light',
)

function onOpenChange(value: boolean) {
  internalOpen.value = value
  emit('update:open', value)
}

function onSelect(e: EmojiExt) {
  // `i` is the emoji unicode character (EMOJI_EMOJI_KEY).
  emit('select', e.i)
  onOpenChange(false)
}
</script>

<style scoped>
/* Bridge the picker chrome to Gridea's theme tokens so it follows light/dark.
   vue3-emoji-picker exposes these --v3-picker-* custom properties. */
:deep(.v3-emoji-picker) {
  --v3-picker-bg: hsl(var(--popover));
  --v3-picker-fg: hsl(var(--popover-foreground));
  --v3-picker-border: hsl(var(--border));
  --v3-picker-emoji-hover: hsl(var(--accent));
  --v3-picker-input-bg: hsl(var(--secondary));
  --v3-picker-input-border: hsl(var(--border));
  --v3-picker-input-focus-border: hsl(var(--primary));
  border: none;
  box-shadow: none;
}
</style>
