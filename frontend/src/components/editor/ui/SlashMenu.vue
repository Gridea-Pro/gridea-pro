<template>
  <div
    ref="rootEl"
    class="slash-menu w-[260px] max-h-[320px] overflow-auto rounded-md border border-border bg-popover p-1 text-popover-foreground shadow-md"
  >
    <div v-if="items.length === 0" class="px-2 py-3 text-center text-sm text-muted-foreground">
      {{ t('editor.slash.empty') }}
    </div>

    <template v-else>
      <div v-for="grp in grouped" :key="grp.group" class="mb-1 last:mb-0">
        <div class="px-2 pb-1 pt-1.5 text-xs font-medium text-muted-foreground">
          {{ t('editor.slash.group.' + grp.group) }}
        </div>
        <button
          v-for="entry in grp.entries"
          :key="entry.item.key"
          :ref="(el) => setItemRef(el, entry.index)"
          type="button"
          class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm text-foreground transition-colors cursor-pointer"
          :class="entry.index === selected ? 'ed-active' : 'ed-hover'"
          @mouseenter="selected = entry.index"
          @mousedown.prevent
          @click="pick(entry.item)"
        >
          <component :is="entry.item.icon" class="h-4 w-4 shrink-0" />
          <span class="truncate">{{ t('editor.slash.' + entry.item.key) }}</span>
        </button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SlashItem } from '../extensions/slash/data'

const props = defineProps<{
  items: SlashItem[]
  command: (item: SlashItem) => void
}>()

const { t } = useI18n()

const selected = ref(0)
const rootEl = ref<HTMLElement | null>(null)
const itemEls = ref<(HTMLElement | null)[]>([])

const GROUP_ORDER: SlashItem['group'][] = ['basic', 'block', 'insert']

interface GroupEntry {
  item: SlashItem
  index: number
}
interface Grouped {
  group: SlashItem['group']
  entries: GroupEntry[]
}

/** 按 group 分组，并为每项分配一个在「扁平可见列表」中的全局索引（高亮以此为准）。 */
const grouped = computed<Grouped[]>(() => {
  const out: Grouped[] = []
  let index = 0
  for (const group of GROUP_ORDER) {
    const entries: GroupEntry[] = []
    for (const item of props.items) {
      if (item.group !== group) continue
      entries.push({ item, index })
      index += 1
    }
    if (entries.length > 0) out.push({ group, entries })
  }
  return out
})

/** 与 grouped 中的全局索引顺序一致的扁平列表，供键盘导航/执行使用。 */
const flatItems = computed<SlashItem[]>(() => grouped.value.flatMap((g) => g.entries.map((e) => e.item)))

function setItemRef(el: Element | { $el: Element } | null, index: number): void {
  itemEls.value[index] = (el as HTMLElement) ?? null
}

function clampSelected(): void {
  const max = flatItems.value.length - 1
  if (max < 0) {
    selected.value = 0
    return
  }
  if (selected.value > max) selected.value = max
  if (selected.value < 0) selected.value = 0
}

function scrollSelectedIntoView(): void {
  nextTick(() => {
    itemEls.value[selected.value]?.scrollIntoView({ block: 'nearest' })
  })
}

function move(delta: number): void {
  const len = flatItems.value.length
  if (len === 0) return
  selected.value = (selected.value + delta + len) % len
  scrollSelectedIntoView()
}

function pick(item: SlashItem): void {
  props.command(item)
}

function runSelected(): boolean {
  const item = flatItems.value[selected.value]
  if (!item) return false
  pick(item)
  return true
}

function onKeyDown(event: KeyboardEvent): boolean {
  switch (event.key) {
    case 'ArrowUp':
      move(-1)
      return true
    case 'ArrowDown':
      move(1)
      return true
    case 'Enter':
    case 'Tab':
      return runSelected()
    default:
      return false
  }
}

watch(
  () => props.items,
  () => {
    selected.value = 0
    itemEls.value = []
    clampSelected()
    scrollSelectedIntoView()
  },
)

defineExpose({ onKeyDown })
</script>
