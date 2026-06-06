<script setup lang="ts">
/**
 * 文字颜色 / 荧光笔 选色弹层，1:1 对齐 editor-vue（PandaWiki）的 Ui/ColorPicker：
 * 默认(清除)行 + 艺术色卡(10×7) + 渐变色(仅文字色) + 最近使用(localStorage, 上限10) + 其他颜色(取色器)。
 * type='text' 用于文字颜色（支持渐变），type='highlight' 用于荧光笔（无渐变行）。
 * select 事件：颜色字符串（hex 或 CSS gradient）；null 表示恢复默认/清除。
 */
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { IconLetterA, IconHighlight, IconChevronDown, IconCheck, IconBan, IconPalette, IconChevronRight } from '@tabler/icons-vue'
import { Chrome } from '@ckpack/vue-color'
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover'
import { normalizeCssColor } from '../extensions/RichTextStyle'

const props = withDefaults(defineProps<{ type?: 'text' | 'highlight'; modelValue?: string }>(), {
  type: 'text',
  modelValue: '',
})

const emit = defineEmits<{ select: [color: string | null] }>()

const { t } = useI18n()

const open = ref(false)
const customOpen = ref(false)

// ── 色值（原样照搬 editor-vue / PandaWiki）─────────────────
const paletteColors = [
  '#000000', '#262626', '#595959', '#8C8C8C', '#BFBFBF', '#D9D9D9', '#E9E9E9', '#F5F5F5', '#FAFAFA', '#FFFFFF',
  '#F5222D', '#FA541C', '#FA8C16', '#FADB14', '#52C41A', '#13C2C2', '#1890FF', '#2F54EB', '#722ED1', '#EB2F96',
  '#FFE8E6', '#FFECE0', '#FFF7E6', '#FEFFE6', '#E6FFFB', '#E6F7FF', '#F0F5FF', '#F9F0FF', '#FFF0F6', '#FFFFFF',
  '#FFA39E', '#FFBB96', '#FFD591', '#FFFF00', '#B7EB8F', '#87E8DE', '#91D5FF', '#ADC6FF', '#D3ADF7', '#FFADD2',
  '#FF4D4F', '#FF7A45', '#FFA940', '#FFEC3D', '#73D13D', '#36CFC9', '#4096FF', '#597EF7', '#9254DE', '#F759AB',
  '#CF1322', '#D4380D', '#D46B08', '#D4B106', '#389E0D', '#08979C', '#096DD9', '#1D39C4', '#531DAB', '#C41D7F',
  '#820014', '#871400', '#873800', '#614700', '#135200', '#00474F', '#003A8C', '#061178', '#22075E', '#780650',
]
const gradientColors = [
  'linear-gradient(132deg, rgb(36, 73, 254) 0%, rgb(202, 75, 167) 100%)',
  'linear-gradient(132deg, rgb(255, 65, 108) 0%, rgb(255, 75, 43) 100%)',
  'linear-gradient(132deg, rgb(255, 113, 0) 0%, rgb(243, 0, 173) 100%)',
  'linear-gradient(132deg, rgb(198, 118, 255) 0%, rgb(101, 76, 255) 41%, rgb(64, 94, 255) 75%, rgb(0, 127, 255) 99%)',
  'linear-gradient(132deg, rgb(0, 201, 255) 0%, rgb(146, 254, 157) 100%)',
  'linear-gradient(132deg, rgb(252, 70, 107) 0%, rgb(63, 94, 251) 100%)',
  'linear-gradient(132deg, rgb(63, 43, 150) 0%, rgb(168, 192, 255) 100%)',
  'linear-gradient(132deg, rgb(253, 187, 45) 0%, rgb(34, 193, 195) 100%)',
  'linear-gradient(132deg, rgb(213, 51, 105) 0%, rgb(218, 174, 81) 100%)',
  'linear-gradient(132deg, rgb(152, 144, 227) 0%, rgb(177, 244, 207) 100%)',
]

// ── 最近使用（与 editor-vue 同 key / 同上限 / 同去重）────────
const RECENT_KEY = 'panda-recent-colors'
const recentColors = ref<string[]>([])

onMounted(() => {
  try {
    const stored = localStorage.getItem(RECENT_KEY)
    if (stored) {
      const parsed = JSON.parse(stored)
      // 防坏数据：非数组（或被写入字符串）一律回退空
      if (Array.isArray(parsed)) recentColors.value = parsed.filter((x) => typeof x === 'string')
    }
  } catch {
    /* ignore */
  }
})

function addToRecent(color: string) {
  if (!color) return
  const idx = recentColors.value.indexOf(color)
  if (idx > -1) recentColors.value.splice(idx, 1)
  recentColors.value.unshift(color)
  if (recentColors.value.length > 10) recentColors.value.pop()
  try {
    localStorage.setItem(RECENT_KEY, JSON.stringify(recentColors.value))
  } catch {
    /* ignore */
  }
}

// ── 当前态 ───────────────────────────────────────────
const current = computed(() => props.modelValue || '')
const isGradient = (v: string) => v.includes('gradient')
// 浏览器会把 hex 归一成 rgb(...)：比较前两侧都过 normalizeCssColor，否则选中态圈不亮
const isActive = (v: string) => {
  if (!current.value) return false
  const norm = (s: string) => normalizeCssColor(s).replace(/\s/g, '').toLowerCase()
  return norm(current.value) === norm(v)
}

const barStyle = computed(() => {
  if (!current.value) return { backgroundColor: props.type === 'highlight' ? '#FFFF00' : 'transparent' }
  return isGradient(current.value)
    ? { backgroundImage: current.value }
    : { backgroundColor: current.value }
})

function swatchStyle(v: string) {
  return isGradient(v) ? { backgroundImage: v } : { backgroundColor: v }
}

// ── 行为 ────────────────────────────────────────────
function selectColor(color: string | null) {
  emit('select', color)
  if (color) addToRecent(color)
  // 点了色卡/默认就作废取色器的中间值，避免 onOpenChange 把拖动残值塞进「最近使用」
  lastCustom = ''
  open.value = false
  customOpen.value = false
}

// 取色器拖动防抖（150ms）：否则每像素一个 transaction，撤销栈被灌爆
let lastCustom = ''
let pickerTimer: ReturnType<typeof setTimeout> | null = null
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onPickerChange(payload: any) {
  if (payload && typeof payload.hex === 'string') {
    lastCustom = payload.hex
    if (pickerTimer) clearTimeout(pickerTimer)
    pickerTimer = setTimeout(() => emit('select', lastCustom), 150)
  }
}
onBeforeUnmount(() => {
  if (pickerTimer) clearTimeout(pickerTimer)
})
function onOpenChange(v: boolean) {
  if (!v && lastCustom) {
    addToRecent(lastCustom)
    lastCustom = ''
  }
  if (!v) customOpen.value = false
}
</script>

<template>
  <Popover v-model:open="open" @update:open="onOpenChange">
    <PopoverTrigger as-child>
      <button
        type="button"
        class="inline-flex h-8 items-center gap-0.5 rounded-md px-1 text-foreground ed-ctl"
        :title="type === 'text' ? t('editor.color.textTitle') : t('editor.color.highlightTitle')"
        @mousedown.prevent
      >
        <span class="flex flex-col items-center gap-0.5">
          <IconLetterA v-if="type === 'text'" class="h-4 w-4" />
          <IconHighlight v-else class="h-4 w-4" />
          <span class="h-1 w-4 rounded-sm border border-border" :style="barStyle" />
        </span>
        <IconChevronDown class="h-3 w-3 opacity-60" />
      </button>
    </PopoverTrigger>

    <PopoverContent class="w-[280px] p-3 select-none" align="start" @mousedown.prevent.stop>
      <!-- 默认 / 无颜色 -->
      <button
        type="button"
        class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm ed-ctl"
        @click="selectColor(null)"
      >
        <span
          class="flex h-5 w-5 items-center justify-center rounded border border-border"
          :class="{ 'bg-foreground text-background': !current }"
        >
          <IconCheck v-if="!current" class="h-3.5 w-3.5" />
          <IconBan v-else class="h-3.5 w-3.5 text-muted-foreground" />
        </span>
        {{ type === 'text' ? t('editor.color.default') : t('editor.color.noColor') }}
      </button>

      <div class="my-2 border-t border-border" />

      <!-- 艺术色卡 -->
      <div class="mb-2 text-xs text-muted-foreground">{{ t('editor.color.palette') }}</div>
      <div class="grid grid-cols-10 gap-1">
        <button
          v-for="(c, i) in paletteColors"
          :key="`p${i}`"
          type="button"
          class="relative h-5 w-5 rounded-sm border border-black/5 cursor-pointer transition-transform hover:z-10 hover:scale-110 hover:shadow"
          :class="{ 'ring-1 ring-primary ring-offset-1 ring-offset-popover': isActive(c) }"
          :style="{ backgroundColor: c }"
          :title="c"
          @click="selectColor(c)"
        />
      </div>

      <!-- 渐变色（仅文字颜色） -->
      <template v-if="type === 'text'">
        <div class="mb-2 mt-3 text-xs text-muted-foreground">{{ t('editor.color.gradient') }}</div>
        <div class="grid grid-cols-10 gap-1">
          <button
            v-for="(g, i) in gradientColors"
            :key="`g${i}`"
            type="button"
            class="h-5 w-5 rounded-sm border border-black/5 cursor-pointer transition-transform hover:z-10 hover:scale-110 hover:shadow"
            :class="{ 'ring-1 ring-primary ring-offset-1 ring-offset-popover': isActive(g) }"
            :style="{ backgroundImage: g }"
            @click="selectColor(g)"
          />
        </div>
      </template>

      <!-- 最近使用 -->
      <template v-if="recentColors.length">
        <div class="mb-2 mt-3 border-t border-border pt-2 text-xs text-muted-foreground">
          {{ t('editor.color.recent') }}
        </div>
        <div class="grid grid-cols-10 gap-1">
          <button
            v-for="(c, i) in recentColors"
            :key="`r${i}`"
            type="button"
            class="h-5 w-5 rounded-sm border border-black/5 cursor-pointer transition-transform hover:z-10 hover:scale-110 hover:shadow"
            :class="{ 'ring-1 ring-primary ring-offset-1 ring-offset-popover': isActive(c) }"
            :style="swatchStyle(c)"
            :title="c"
            @click="selectColor(c)"
          />
        </div>
      </template>

      <!-- 其他颜色 -->
      <div class="mt-3 border-t border-border pt-2">
        <button
          type="button"
          class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm ed-ctl"
          @click="customOpen = !customOpen"
        >
          <IconPalette class="h-4 w-4 text-muted-foreground" />
          {{ t('editor.color.custom') }}
          <IconChevronRight class="ml-auto h-3.5 w-3.5 text-muted-foreground transition-transform" :class="{ 'rotate-90': customOpen }" />
        </button>
        <div v-if="customOpen" class="color-picker-chrome mt-2">
          <Chrome
            :model-value="!current || isGradient(current) ? '#000000' : current"
            disable-alpha
            @update:model-value="onPickerChange"
          />
        </div>
      </div>
    </PopoverContent>
  </Popover>
</template>

<style scoped>
/* Chrome 取色器贴合弹层主题 */
.color-picker-chrome :deep(.vc-chrome) {
  width: 100%;
  box-shadow: none;
  border: 1px solid var(--editor-border);
  border-radius: 8px;
  background: var(--editor-popover);
}
.color-picker-chrome :deep(.vc-chrome-body) {
  background: var(--editor-popover);
}
</style>
