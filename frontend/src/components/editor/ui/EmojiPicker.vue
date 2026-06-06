<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <!-- 直接 button 子元素：as-child 包 slot 间接层会导致触发 props 丢失（历史教训） -->
      <button
        type="button"
        class="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground ed-ctl"
        :title="t('editor.emoji.title')"
        :aria-label="t('editor.emoji.title')"
        @mousedown.prevent
      >
        <Smile class="h-[18px] w-[18px]" />
      </button>
    </PopoverTrigger>
    <PopoverContent class="w-[324px] p-0" align="start" @open-auto-focus.prevent>
      <!-- 面板移植自 editor-vue（搜索 + 分类 + 网格 + 键盘导航），并修复其缺陷：
           ① 分类改用数据自带 group（原版用关键词模糊搜索，不准且乱序）；
           ② ↑↓ 改为按行（8 格）移动（原版只动 1 格）；
           ③ 新增「最近使用」；④ 仅渲染原生 emoji，不依赖 GitHub CDN 回退图。 -->
      <div class="emoji-panel" @keydown="onKeydown">
        <div class="emoji-search">
          <input
            ref="searchInput"
            v-model="query"
            type="text"
            class="emoji-search-input"
            :placeholder="t('editor.emojiPicker.search')"
            spellcheck="false"
          />
        </div>
        <div v-if="!hasQuery" class="emoji-tabs">
          <button
            v-for="(tab, i) in visibleTabs"
            :key="tab.key"
            type="button"
            :class="['emoji-tab', 'ed-ctl', { 'is-active': activeTab === i }]"
            @click="switchTab(i)"
          >
            {{ tab.label() }}
          </button>
        </div>
        <div ref="gridWrap" class="emoji-grid-wrap">
          <div v-if="!shown.length" class="emoji-empty">😕 {{ t('editor.emojiPicker.empty') }}</div>
          <div v-else class="emoji-grid">
            <button
              v-for="(it, idx) in shown"
              :key="`${it.name}-${idx}`"
              type="button"
              :data-emoji-index="idx"
              :class="['emoji-cell', { 'is-selected': idx === selectedIndex }]"
              :title="it.name"
              @mouseenter="selectedIndex = idx"
              @click="pick(it)"
            >
              {{ it.emoji }}
            </button>
          </div>
        </div>
      </div>
    </PopoverContent>
  </Popover>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { IconMoodSmile as Smile } from '@tabler/icons-vue'
import * as emojiModule from '@tiptap/extension-emoji'
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover'

const emit = defineEmits<{ select: [emoji: string] }>()
const { t } = useI18n()

const open = ref(false)
const query = ref('')
const activeTab = ref(0)
const selectedIndex = ref(0)
const searchInput = ref<HTMLInputElement | null>(null)
const gridWrap = ref<HTMLElement | null>(null)

const PER_ROW = 8

interface EmojiItem {
  name: string
  emoji?: string
  shortcodes?: string[]
  tags?: string[]
  group?: string
}

// 仅原生 emoji（排除肤色修饰符 components 与仅有 CDN 回退图的 GitHub 自定义项）
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const RAW: EmojiItem[] = ((emojiModule as any).gitHubEmojis || (emojiModule as any).emojis || []) as EmojiItem[]
const ALL: EmojiItem[] = RAW.filter(
  (e) => !!e.emoji && e.group !== 'GitHub' && e.group !== 'components',
)
const byName = new Map(ALL.map((e) => [e.name, e]))

// ── 最近使用 ─────────────────────────────────────────
const RECENT_KEY = 'gridea-recent-emojis'
const recentNames = ref<string[]>([])
try {
  const stored = localStorage.getItem(RECENT_KEY)
  if (stored) {
    const parsed = JSON.parse(stored)
    if (Array.isArray(parsed)) recentNames.value = parsed.filter((x) => typeof x === 'string')
  }
} catch {
  /* ignore */
}
function addRecent(name: string) {
  const idx = recentNames.value.indexOf(name)
  if (idx > -1) recentNames.value.splice(idx, 1)
  recentNames.value.unshift(name)
  if (recentNames.value.length > 24) recentNames.value.length = 24
  try {
    localStorage.setItem(RECENT_KEY, JSON.stringify(recentNames.value))
  } catch {
    /* ignore */
  }
}
const recentItems = computed(() =>
  recentNames.value.map((n) => byName.get(n)).filter((x): x is EmojiItem => !!x),
)

// ── 分类（按数据自带 group，精准且有序）─────────────────
interface Tab {
  key: string
  group: string | null // null = 最近
  label: () => string
}
const tabs: Tab[] = [
  { key: 'recent', group: null, label: () => t('editor.emojiPicker.recent') },
  { key: 'smileys', group: '', label: () => t('editor.emojiPicker.smileys') },
  { key: 'people', group: 'people & body', label: () => t('editor.emojiPicker.people') },
  { key: 'nature', group: 'animals & nature', label: () => t('editor.emojiPicker.nature') },
  { key: 'food', group: 'food & drink', label: () => t('editor.emojiPicker.food') },
  { key: 'travel', group: 'travel & places', label: () => t('editor.emojiPicker.travel') },
  { key: 'activity', group: 'activities', label: () => t('editor.emojiPicker.activity') },
  { key: 'objects', group: 'objects', label: () => t('editor.emojiPicker.objects') },
  { key: 'symbols', group: 'symbols', label: () => t('editor.emojiPicker.symbols') },
  { key: 'flags', group: 'flags', label: () => t('editor.emojiPicker.flags') },
]
const visibleTabs = computed(() => (recentItems.value.length ? tabs : tabs.slice(1)))

const hasQuery = computed(() => query.value.trim().length > 0)

// ── 搜索（移植 editor-vue 评分：精确=3 / 前缀=2 / 包含=1，多词累加）──
const shown = computed<EmojiItem[]>(() => {
  if (hasQuery.value) {
    const keywords = query.value.trim().toLowerCase().split(/\s+/).filter(Boolean)
    return ALL.map((item) => {
      const haystack = [item.name, ...(item.shortcodes || []), ...(item.tags || [])].map((s) =>
        s.toLowerCase(),
      )
      let score = 0
      for (const kw of keywords) {
        if (haystack.some((h) => h === kw)) score += 3
        else if (haystack.some((h) => h.startsWith(kw))) score += 2
        else if (haystack.some((h) => h.includes(kw))) score += 1
      }
      return { item, score }
    })
      .filter(({ score }) => score > 0)
      .sort((a, b) => b.score - a.score)
      .slice(0, 160)
      .map(({ item }) => item)
  }
  const tab = visibleTabs.value[activeTab.value]
  if (!tab) return []
  if (tab.group === null) return recentItems.value
  return ALL.filter((e) => e.group === tab.group)
})

function switchTab(i: number) {
  activeTab.value = i
  selectedIndex.value = 0
}

watch([query, () => shown.value.length], () => {
  selectedIndex.value = 0
})

// 面板打开：重置 + 聚焦搜索框
watch(open, (o) => {
  if (!o) return
  query.value = ''
  activeTab.value = 0 // visibleTabs 首个（有最近则为最近，否则为表情）
  selectedIndex.value = 0
  void nextTick(() => searchInput.value?.focus())
})

// 选中项滚动跟随
watch(selectedIndex, () => {
  void nextTick(() => {
    gridWrap.value
      ?.querySelector(`[data-emoji-index="${selectedIndex.value}"]`)
      ?.scrollIntoView({ block: 'nearest' })
  })
})

function pick(it: EmojiItem) {
  if (!it.emoji) return
  emit('select', it.emoji)
  addRecent(it.name)
  open.value = false
}

function onKeydown(e: KeyboardEvent) {
  const len = shown.value.length
  if (!len) return
  switch (e.key) {
    case 'ArrowRight':
      e.preventDefault()
      selectedIndex.value = Math.min(len - 1, selectedIndex.value + 1)
      break
    case 'ArrowLeft':
      e.preventDefault()
      selectedIndex.value = Math.max(0, selectedIndex.value - 1)
      break
    case 'ArrowDown':
      // 修复：按行移动（原版只 +1）
      e.preventDefault()
      selectedIndex.value = Math.min(len - 1, selectedIndex.value + PER_ROW)
      break
    case 'ArrowUp':
      e.preventDefault()
      selectedIndex.value = Math.max(0, selectedIndex.value - PER_ROW)
      break
    case 'Enter': {
      e.preventDefault()
      const it = shown.value[selectedIndex.value]
      if (it) pick(it)
      break
    }
  }
}
</script>

<style scoped>
.emoji-panel {
  width: 100%;
  user-select: none;
}
.emoji-search {
  padding: 8px 8px 6px;
}
.emoji-search-input {
  width: 100%;
  height: 30px;
  padding: 0 10px;
  border: 1px solid var(--editor-border);
  border-radius: 6px;
  background: var(--editor-secondary);
  color: var(--editor-fg);
  font-size: 13px;
  outline: none;
}
.emoji-search-input:focus {
  border-color: var(--editor-accent);
}
.emoji-tabs {
  display: flex;
  gap: 2px;
  padding: 0 8px 6px;
  overflow-x: auto;
  scrollbar-width: none;
}
.emoji-tabs::-webkit-scrollbar {
  display: none;
}
.emoji-tab {
  flex-shrink: 0;
  padding: 3px 8px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--editor-muted);
  font-size: 12px;
  white-space: nowrap;
}
.emoji-tab.is-active {
  background: color-mix(in srgb, var(--editor-accent) 16%, transparent);
  color: var(--editor-accent);
}
.emoji-grid-wrap {
  height: 264px;
  overflow-y: auto;
  padding: 0 8px 8px;
}
.emoji-grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 2px;
}
.emoji-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 34px;
  border: none;
  border-radius: 6px;
  background: transparent;
  font-size: 20px;
  line-height: 1;
  cursor: pointer;
}
.emoji-cell.is-selected {
  background: var(--editor-hover);
}
.emoji-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 200px;
  color: var(--editor-muted);
  font-size: 13px;
}
</style>
