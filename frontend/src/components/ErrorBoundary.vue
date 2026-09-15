<template>
  <!-- 出错时只替换本边界包裹的区域，边界之外（侧栏、部署面板等）保持可用 -->
  <div v-if="failure" class="flex h-full w-full items-center justify-center p-8">
    <div class="w-full max-w-lg rounded-lg border border-border bg-card p-6 shadow-sm">
      <div class="flex items-start gap-3">
        <AlertTriangle class="mt-0.5 size-5 shrink-0 text-destructive" />
        <div class="min-w-0 flex-1">
          <h2 class="text-sm font-medium text-foreground">{{ t('error.boundary.title') }}</h2>
          <p class="mt-1 break-words text-xs text-muted-foreground">{{ failure.message }}</p>
        </div>
      </div>

      <div class="mt-4 flex items-center gap-2">
        <Button size="sm" @click="retry">
          <RefreshCw class="mr-1.5 size-3.5" />
          {{ t('error.boundary.retry') }}
        </Button>
        <Button size="sm" variant="outline" @click="copyDiagnostics">
          <Copy class="mr-1.5 size-3.5" />
          {{ copied ? t('error.boundary.copied') : t('error.boundary.copyDiagnostics') }}
        </Button>
        <button
          class="ml-auto flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
          @click="detailsOpen = !detailsOpen">
          {{ t('error.boundary.details') }}
          <ChevronDown class="size-3.5 transition-transform" :class="{ 'rotate-180': detailsOpen }" />
        </button>
      </div>

      <pre
        v-if="detailsOpen"
        class="mt-3 max-h-56 overflow-auto rounded bg-muted p-3 text-[11px] leading-relaxed text-muted-foreground"
        >{{ failure.stack || failure.message }}</pre
      >
    </div>
  </div>

  <!-- 重试时先摘掉再挂回，强制重建子树状态（key 放不到 template 上，只能靠 v-if 切换） -->
  <template v-else-if="mounted">
    <slot />
  </template>
</template>

<script setup lang="ts">
import { ref, nextTick, onErrorCaptured } from 'vue'
import { useI18n } from 'vue-i18n'
import { AlertTriangle, RefreshCw, Copy, ChevronDown } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { reportError, formatDiagnostics, type ReportedError } from '@/helpers/errorReporter'

const props = withDefaults(defineProps<{ label?: string }>(), { label: '' })

const { t } = useI18n()
const failure = ref<ReportedError | null>(null)
const detailsOpen = ref(false)
const copied = ref(false)
const mounted = ref(true)

onErrorCaptured((err, instance) => {
  const component = (instance?.$options as { name?: string } | undefined)?.name
  failure.value = reportError(err, 'boundary', [props.label, component].filter(Boolean).join(' / '))
  // 阻止继续向上冒泡：否则会一路传到 App 根部，退化成整窗口报错
  return false
})

async function retry() {
  failure.value = null
  detailsOpen.value = false
  copied.value = false
  mounted.value = false
  await nextTick()
  mounted.value = true
}

async function copyDiagnostics() {
  if (!failure.value) return
  await navigator.clipboard.writeText(formatDiagnostics(failure.value))
  copied.value = true
  setTimeout(() => (copied.value = false), 2000)
}
</script>
