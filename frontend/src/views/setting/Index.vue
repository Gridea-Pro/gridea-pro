<template>
  <div class="h-full flex flex-col bg-background">
    <TabGroup :selectedIndex="selectedIndex" @change="onTabChange">
      <div
        class="flex-shrink-0 flex justify-between items-center px-4 h-12 bg-background sticky top-0 z-50 border-b border-border backdrop-blur-sm bg-opacity-90 select-none"
        style="--wails-draggable: drag">
        <div class="flex-1 overflow-x-auto no-scrollbar">
          <TabList as="nav" ref="tabListRef" class="relative flex items-center">
            <div ref="bgRef" class="tab-transition absolute h-[70%] pointer-events-none" style="width: 0; left: 0;">
            </div>
            <div ref="lineRef" class="tab-transition absolute bottom-0 h-[2px] bg-primary rounded-full z-10"
              style="width: 0; left: 0; opacity: 0;"></div>

            <Tab as="template" v-slot="{ selected }" v-for="(key, index) in tabs" :key="key">
              <button :ref="(el) => setBtnRef(el as any, index)"
                class="tab-btn relative z-20 px-4 py-2 text-[12px] font-medium tracking-wide transition-all duration-500 cursor-pointer whitespace-nowrap outline-none"
                :class="selected ? 'active text-primary' : 'text-muted-foreground hover:text-primary'"
                style="--wails-draggable: no-drag">
                <span class="text-glow">{{ t(`settings.${key}.title`) }}</span>
              </button>
            </Tab>
          </TabList>
        </div>
      </div>

      <div class="flex-1 overflow-y-auto px-4 py-6">
        <TabPanels class="h-full">
          <TabPanel class="focus:outline-none">
            <basic-setting></basic-setting>
          </TabPanel>
        </TabPanels>
      </div>
    </TabGroup>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, ref, watch, computed, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { TabGroup, TabList, Tab, TabPanels, TabPanel } from '@headlessui/vue'
import BasicSetting from './includes/BasicSetting.vue'

const { t } = useI18n()

const tabs = ['network']
const currentTab = ref('network')

const tabListRef = ref<any>(null)
const lineRef = ref<HTMLElement | null>(null)
const bgRef = ref<HTMLElement | null>(null)
const btnRefs = ref<Array<HTMLElement | null>>([])

let isUnmounted = false

const resolveDomEl = (maybeEl: any): HTMLElement | null => {
  const dom = maybeEl && maybeEl.$el ? maybeEl.$el : maybeEl
  return dom instanceof HTMLElement ? dom : null
}

const setBtnRef = (el: any, index: number) => {
  btnRefs.value[index] = resolveDomEl(el)
}

const selectedIndex = computed(() => {
  return tabs.indexOf(currentTab.value)
})

const onTabChange = (index: number) => {
  currentTab.value = tabs[index]
}

const applyIndicator = async (stretch = false) => {
  await nextTick()
  if (isUnmounted) return

  const line = lineRef.value
  const bg = bgRef.value
  const idx = selectedIndex.value
  const btn = btnRefs.value[idx]

  if (!line || !bg || !btn) {
    if (line) line.style.opacity = '0'
    if (bg) bg.style.width = '0'
    return
  }

  const lineWidth = btn.offsetWidth * 0.4
  line.style.opacity = '1'
  line.style.width = `${lineWidth}px`
  line.style.left = `${btn.offsetLeft + (btn.offsetWidth - lineWidth) / 2}px`

  bg.style.width = `${btn.offsetWidth}px`
  bg.style.left = `${btn.offsetLeft}px`

  if (stretch) {
    line.style.transform = 'scaleX(1.4)'
    window.setTimeout(() => {
      if (lineRef.value) lineRef.value.style.transform = 'scaleX(1)'
    }, 250)
  }
}

let resizeObserver: ResizeObserver | null = null

onMounted(async () => {
  await nextTick()
  if (isUnmounted) return

  applyIndicator(false)

  const listEl = resolveDomEl(tabListRef.value)
  if (!listEl) return

  resizeObserver = new ResizeObserver(() => {
    if (isUnmounted) return
    applyIndicator(false)
  })
  resizeObserver.observe(listEl)
})

watch(selectedIndex, (newIndex, oldIndex) => {
  applyIndicator(newIndex !== oldIndex)
})

onUnmounted(() => {
  isUnmounted = true
  resizeObserver?.disconnect()
  resizeObserver = null
})
</script>

<style scoped>
.tab-transition {
  transition: all 0.2s cubic-bezier(0.23, 1, 0.32, 1);
}

.text-glow {
  text-shadow: 0 0 15px rgba(99, 102, 241, 0);
  transition: all 0.2s ease;
}

.no-scrollbar::-webkit-scrollbar {
  display: none;
}
</style>
