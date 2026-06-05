<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { IconLetterA as Baseline, IconBan as Ban } from '@tabler/icons-vue'
import { Chrome } from '@ckpack/vue-color'
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover'

const props = defineProps<{ modelValue?: string }>()

const emit = defineEmits<{ select: [color: string] }>()

const { t } = useI18n()

const open = ref(false)

const presets = [
  '#000000',
  '#434343',
  '#999999',
  '#ffffff',
  '#e60000',
  '#ff9900',
  '#ffff00',
  '#008a00',
  '#0066cc',
  '#9933ff',
  '#ff5599',
  '#795548',
]

const currentColor = computed(() => props.modelValue ?? '')

function selectPreset(color: string) {
  emit('select', color)
  open.value = false
}

// @ckpack/vue-color 的 Payload 类型，此处只取 hex
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onPickerChange(payload: any) {
  if (payload && typeof payload.hex === 'string') {
    emit('select', payload.hex as string)
  }
}

function clearColor() {
  emit('select', '')
  open.value = false
}
</script>

<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <button
        type="button"
        class="inline-flex h-8 flex-col items-center justify-center gap-0.5 rounded-md px-1.5 text-foreground hover:bg-accent hover:text-accent-foreground"
        :title="t('editor.color.title')"
        @mousedown.prevent
      >
        <Baseline class="h-4 w-4" />
        <span
          class="h-1 w-4 rounded-sm border border-border"
          :style="{ backgroundColor: currentColor || 'transparent' }"
        />
      </button>
    </PopoverTrigger>

    <PopoverContent
      class="w-auto p-3"
      align="start"
      @mousedown.prevent.stop
    >
      <div class="flex flex-col gap-3">
        <div class="text-xs font-medium text-muted-foreground">
          {{ t('editor.color.presets') }}
        </div>

        <div class="grid grid-cols-6 gap-1.5">
          <button
            v-for="color in presets"
            :key="color"
            type="button"
            class="h-6 w-6 rounded-md border border-border transition-transform hover:scale-110"
            :class="{ 'ring-2 ring-primary ring-offset-1 ring-offset-popover': currentColor.toLowerCase() === color.toLowerCase() }"
            :style="{ backgroundColor: color }"
            :title="color"
            @click="selectPreset(color)"
          />
        </div>

        <div class="color-picker-chrome">
          <Chrome
            :model-value="currentColor || '#000000'"
            disable-alpha
            @update:model-value="onPickerChange"
          />
        </div>

        <button
          type="button"
          class="inline-flex items-center justify-center gap-1.5 rounded-md border border-border px-2 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground"
          @click="clearColor"
        >
          <Ban class="h-3.5 w-3.5" />
          {{ t('editor.color.clear') }}
        </button>
      </div>
    </PopoverContent>
  </Popover>
</template>

<style scoped>
.color-picker-chrome :deep(.vc-chrome) {
  width: 100%;
  box-shadow: none;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-family: inherit;
}
.color-picker-chrome :deep(.vc-chrome-body) {
  background: var(--popover);
}
</style>
