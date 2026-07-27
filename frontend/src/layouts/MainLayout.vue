<template>
  <div class="flex h-screen w-full overflow-hidden bg-background text-foreground">
    <!-- Sidebar -->
    <aside
v-if="sidebarVisible"
      class="w-[200px] flex-shrink-0 flex flex-col bg-sidebar border-r border-border z-10 select-none">
      <!-- Draggable Area + Window Controls (Windows/Linux: traffic lights) -->
      <div class="h-10 w-full flex-shrink-0 header-spacer flex items-center">
        <WindowControls />
      </div>

      <div class="flex-1 flex flex-col overflow-hidden">
        <!-- Logo -->
        <div class="h-16 mb-[18px] flex justify-center items-center">
          <img class="h-16 rounded-full" src="@/assets/logo.png" alt="Logo">
        </div>

        <!-- Menu -->
        <nav class="flex-1 overflow-y-auto px-3 py-4 scrollbar-hide">
          <ul class="space-y-1">
            <li v-for="menu in sideMenus" :key="menu.router">
              <Button
variant="ghost"
                class="w-full justify-start px-3 py-2.5 h-12 font-normal hover:bg-primary/15 cursor-pointer transition-colors"
                :class="[
                  currentRouter === menu.router
                    ? 'bg-primary/10 text-primary font-medium hover:bg-primary/15'
                    : 'text-muted-foreground hover:text-foreground'
                ]" @click="clickMenu(menu)">
                <div class="flex items-center w-full">
                  <component
:is="menu.icon" v-if="menu.icon" class="mr-3 size-4 transition-colors duration-200"
                    :class="currentRouter === menu.router ? 'text-primary' : 'text-muted-foreground group-hover:text-primary'" />
                  <span class="text-xs flex-1 text-left">{{ menu.text }}</span>
                  <span
v-if="menu.router === '/comments' && commentStore.unreadCount > 0"
                    class="ml-auto bg-red-500 text-white text-[9px] font-bold px-1 min-w-[14px] h-[14px] flex items-center justify-center rounded-full leading-none">
                    {{ commentStore.unreadCount > 99 ? '99+' : commentStore.unreadCount }}
                  </span>
                  <span
v-if="(menu.count || 0) > 0" class="text-xs transition-colors duration-200 ml-2"
                    :class="currentRouter === menu.router ? 'text-primary opacity-80' : 'text-muted-foreground opacity-50'">
                    {{ menu.count }}
                  </span>
                </div>
              </Button>
            </li>
          </ul>
        </nav>
      </div>

      <!-- Bottom Actions -->
      <div class="p-4 bg-sidebar border-r border-border flex flex-col items-center gap-3 z-50">
        <Button
variant="outline"
          class="w-36 h-8 text-xs justify-center rounded-full border-primary/20 hover:bg-primary/5 cursor-pointer"
          @click="preview">
          <EyeIcon class="size-3 mr-2" />
          {{ t('nav.preview') }}
        </Button>

        <!-- 发布 / 取消按钮（#42）：部署中点击切换为取消；不再需要 disabled -->
        <Button
variant="default"
          class="w-36 h-8 text-xs justify-center rounded-full cursor-pointer"
          :class="publishLoading
            ? 'bg-destructive text-destructive-foreground hover:bg-destructive/90'
            : 'bg-primary text-background hover:bg-primary/90'"
          @click="publishLoading ? cancelPublish() : publish()">
          <template v-if="publishLoading && !cancelling">
            <svg
class="animate-spin h-4 w-4 text-destructive-foreground"
              xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path
class="opacity-75" fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
              </path>
            </svg>
            <span class="ml-2">取消发布</span>
          </template>
          <template v-else-if="cancelling">
            <span class="text-xs">取消中...</span>
          </template>
          <template v-else>
            <RocketLaunchIcon class="size-3 mr-2" />
            {{ t('nav.sync') }}
          </template>
        </Button>

        <div class="flex items-center justify-center gap-6 text-muted-foreground w-[80%]">
          <GlobeAltIcon
v-if="siteStore.currentDomain" class="size-4 cursor-pointer hover:text-primary transition-colors duration-300"
            @click="goWeb" />

          <CogIcon
class="size-4 cursor-pointer hover:text-primary transition-colors duration-300"
            title="设置" @click="openPreferences" />

          <div class="relative group" title="Star Support">
            <svg
viewBox="0 0 24 24" aria-hidden="true"
              class="size-4 cursor-pointer hover:text-primary transition-colors duration-300 fill-current"
              @click="handleGithubClick">
              <path
fill-rule="evenodd" clip-rule="evenodd"
                d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.335 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
            </svg>
          </div>

          <ArrowUpCircleIcon
            v-if="hasUpdate"
            :title="t('update.title')"
            class="update-indicator size-4 cursor-pointer text-primary hover:text-primary/80 transition-colors duration-300"
            @click="openUpdateDialog" />
        </div>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 flex flex-col min-h-0 overflow-hidden bg-background select-none">
      <div class="flex-1 w-full overflow-y-auto overflow-x-hidden p-0">
        <router-view v-slot="{ Component }">
          <keep-alive exclude="Loading,Theme">
            <component :is="Component" />
          </keep-alive>
        </router-view>
      </div>
    </main>

    <!--
      部署面板（#43，重新设计）：push 布局的 flex 兄弟节点，而非 fixed 浮层。
      收起时 width=0（被 overflow-hidden 截断）、展开时 width=380px；主内容通过
      flex 自动收窄给它腾空间。融入 app 自身调色盘，和左栏视觉对称。
    -->
    <aside
      class="flex-shrink-0 overflow-hidden bg-sidebar border-l border-border transition-[width] duration-300 ease-out"
      :class="deployPanelVisible ? 'w-[380px]' : 'w-0'">
      <!-- 内层固定 380px 宽，父层 width 过渡期间不会挤压内容 -->
      <div class="w-[380px] h-full flex flex-col">
        <!-- Header -->
        <div class="px-5 py-4 border-b border-border/60 flex items-start justify-between gap-3">
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 mb-1.5">
              <span
                v-if="publishLoading"
                class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-primary/10 text-primary text-[11px] font-medium">
                <span class="size-1.5 rounded-full bg-primary animate-pulse"></span>
                部署中
              </span>
              <span
                v-else-if="deployOutcome === 'success'"
                class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-green-500/10 text-green-600 dark:text-green-400 text-[11px] font-medium">
                <span class="size-1.5 rounded-full bg-green-500"></span>
                部署完成
              </span>
              <span
                v-else-if="deployOutcome === 'failed'"
                class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-destructive/10 text-destructive text-[11px] font-medium">
                <span class="size-1.5 rounded-full bg-destructive"></span>
                部署失败
              </span>
              <span
                v-else
                class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-muted text-muted-foreground text-[11px] font-medium">
                <span class="size-1.5 rounded-full bg-muted-foreground/40"></span>
                已取消
              </span>
            </div>
            <div class="text-xs text-muted-foreground truncate">
              {{ deploySummary }}
            </div>
          </div>
          <button
            class="size-7 flex-shrink-0 grid place-items-center rounded-full text-muted-foreground hover:text-primary hover:bg-primary/5 transition-colors cursor-pointer focus:outline-none focus-visible:ring-2 focus-visible:ring-primary/30"
            @click="deployPanelVisible = false">
            <XMarkIcon class="size-4" />
          </button>
        </div>

        <!-- 进度条：只有日志里解析到数字才显示 -->
        <div v-if="deployProgress.total > 0" class="px-5 py-3 border-b border-border/60">
          <div class="flex items-center justify-between text-[11px] text-muted-foreground mb-2">
            <span>{{ deployProgress.label }}</span>
            <span class="font-mono tabular-nums">{{ deployProgress.done }} / {{ deployProgress.total }}</span>
          </div>
          <div class="h-1 bg-muted rounded-full overflow-hidden">
            <div
              class="h-full bg-primary transition-all duration-300 ease-out"
              :style="{ width: `${deployProgress.pct}%` }"></div>
          </div>
          <div v-if="deployProgress.failed > 0" class="mt-2 flex items-center gap-3 text-[10px]">
            <span class="text-green-600 dark:text-green-400">
              ✓ 成功 {{ deployProgress.done - deployProgress.failed }}
            </span>
            <span class="text-destructive">✗ 失败 {{ deployProgress.failed }}</span>
          </div>
        </div>

        <!-- 日志流 -->
        <div ref="logScrollEl" class="flex-1 overflow-y-auto py-3 px-4 space-y-1">
          <div
            v-for="(entry, idx) in deployLogs"
            :key="idx"
            class="flex gap-2.5 text-[11px] leading-5"
            :class="logLineClass(entry)">
            <span class="flex-shrink-0 w-3.5 text-center font-mono">{{ logLineIcon(entry) }}</span>
            <span class="flex-1 break-all font-mono">{{ logLineText(entry) }}</span>
          </div>
          <div
            v-if="publishLoading && deployLogs.length === 0"
            class="text-[11px] text-muted-foreground italic px-1">
            等待后端输出...
          </div>
        </div>

        <!-- Footer -->
        <div class="px-4 py-3 border-t border-border/60 bg-muted/30 flex items-center justify-between gap-2">
          <span class="text-[10px] text-muted-foreground font-mono tabular-nums">
            {{ deployLogs.length }} 行
          </span>
          <div class="flex items-center gap-1">
            <button
              class="h-7 px-3 text-[11px] rounded-full text-muted-foreground hover:text-primary hover:bg-primary/5 transition-colors cursor-pointer"
              @click="copyDeployLogs">
              复制日志
            </button>
            <button
              v-if="!publishLoading && deployLogs.length > 0"
              class="h-7 px-3 text-[11px] rounded-full text-muted-foreground hover:text-primary hover:bg-primary/5 transition-colors cursor-pointer"
              @click="clearDeployLogs">
              清空
            </button>
          </div>
        </div>
      </div>
    </aside>

    <!-- Dialogs -->
    <Dialog v-model:open="updateModalVisible">
      <DialogContent class="update-dialog p-0 max-w-[420px] overflow-hidden border-0 shadow-2xl">
        <DialogTitle class="sr-only">{{ t('update.title') }}</DialogTitle>

        <!-- Hero -->
        <div class="relative px-6 pt-8 pb-5 bg-gradient-to-br from-primary/15 via-primary/5 to-transparent">
          <div class="flex items-start gap-4">
            <img
              src="@/assets/logo-pro.png" alt="Gridea Pro"
              class="size-14 rounded-lg shadow-sm flex-shrink-0 object-cover" />
            <div class="flex-1 min-w-0 flex flex-col justify-center gap-1.5">
              <h2 class="text-lg font-semibold text-foreground leading-tight">
                {{ t('update.title') }}
              </h2>
              <div class="inline-flex items-center gap-1.5">
                <span class="inline-flex items-center justify-center h-4 px-2 rounded-full bg-muted/60 text-[10px] text-muted-foreground border border-border/60 font-mono">
                  v{{ currentVersion }}
                </span>
                <ArrowRightIcon class="size-3 text-muted-foreground" />
                <span class="inline-flex items-center justify-center h-4 px-2 rounded-full bg-primary/10 text-[10px] text-primary/80 border border-primary/20 font-mono">
                  v{{ newVersion }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Release notes -->
        <div class="px-8 py-5 max-h-[320px] overflow-y-auto border-t border-border/60">
          <div class="release-notes text-xs text-foreground/90 leading-relaxed" v-html="updateContent"></div>
        </div>

        <!-- Progress (downloading / ready / error) -->
        <div
          v-if="updateState !== 'idle'"
          class="px-6 pt-3 pb-1 border-t border-border/60">
          <div v-if="updateState === 'downloading'" class="space-y-1.5">
            <div class="flex items-center justify-between text-[11px] text-muted-foreground font-mono">
              <span>{{ formatBytes(downloadReceived) }} / {{ formatBytes(downloadTotal) }}</span>
              <span>{{ downloadPercent.toFixed(1) }}%</span>
            </div>
            <div class="h-1.5 w-full bg-muted rounded-full overflow-hidden">
              <div
                class="h-full bg-primary rounded-full transition-[width] duration-150 ease-out"
                :style="{ width: downloadPercent + '%' }"></div>
            </div>
          </div>
          <div v-else-if="updateState === 'ready'" class="flex items-center gap-2 text-xs text-primary">
            <CheckCircleIcon class="size-4" />
            <span>{{ t('update.readyToRestart') }}</span>
          </div>
          <div v-else-if="updateState === 'error'" class="flex items-start gap-2 text-xs text-destructive">
            <ExclamationCircleIcon class="size-4 flex-shrink-0 mt-0.5" />
            <div class="space-y-0.5">
              <div class="font-medium">{{ t(`update.error.${updateErrorKind}.title`) }}</div>
              <div class="text-muted-foreground text-[11px]">{{ t(`update.error.${updateErrorKind}.hint`) }}</div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-6 py-4 border-t border-border/60 bg-muted/30 flex items-center justify-between gap-3">
          <div class="flex items-center gap-3">
            <button
              :title="t('update.viewOnGithub')"
              class="size-7 grid place-items-center rounded-full text-muted-foreground hover:text-primary hover:bg-primary/5 transition-colors cursor-pointer focus:outline-none focus-visible:ring-2 focus-visible:ring-primary/30"
              @click="openInBrowser('https://github.com/Gridea-Pro/gridea-pro/releases')">
              <svg viewBox="0 0 24 24" aria-hidden="true" class="size-4 fill-current">
                <path fill-rule="evenodd" clip-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.335 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
              </svg>
            </button>
            <button
              v-if="updateState === 'idle'"
              class="text-xs text-muted-foreground hover:text-primary transition-colors cursor-pointer focus:outline-none"
              @click="skipThisVersion">
              {{ t('update.skip') }}
            </button>
          </div>
          <div class="flex items-center gap-3">
            <!-- idle: 稍后再说 + 立即更新 -->
            <template v-if="updateState === 'idle'">
              <Button
                variant="outline"
                class="w-20 h-8 text-xs justify-center rounded-full border border-primary/20 text-primary/80 hover:bg-primary/5 hover:text-primary cursor-pointer"
                @click="updateModalVisible = false">
                {{ t('update.later') }}
              </Button>
              <Button
                variant="default"
                class="w-24 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
                @click="startUpdate">
                {{ t('update.install') }}
              </Button>
            </template>
            <!-- downloading: 取消下载 -->
            <template v-else-if="updateState === 'downloading'">
              <Button
                variant="outline"
                class="w-20 h-8 text-xs justify-center rounded-full border border-primary/20 text-primary/80 hover:bg-primary/5 hover:text-primary cursor-pointer"
                @click="cancelUpdate">
                {{ t('common.cancel') }}
              </Button>
            </template>
            <!-- ready: 稍后 + 立即重启 -->
            <template v-else-if="updateState === 'ready'">
              <Button
                variant="outline"
                class="w-20 h-8 text-xs justify-center rounded-full border border-primary/20 text-primary/80 hover:bg-primary/5 hover:text-primary cursor-pointer"
                @click="updateModalVisible = false">
                {{ t('update.later') }}
              </Button>
              <Button
                variant="default"
                class="w-24 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
                @click="applyUpdate">
                {{ t('update.restart') }}
              </Button>
            </template>
            <!-- error: 关闭 + 重试 -->
            <template v-else-if="updateState === 'error'">
              <Button
                variant="outline"
                class="w-20 h-8 text-xs justify-center rounded-full border border-primary/20 text-primary/80 hover:bg-primary/5 hover:text-primary cursor-pointer"
                @click="updateModalVisible = false">
                {{ t('update.later') }}
              </Button>
              <Button
                variant="default"
                class="w-24 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
                @click="startUpdate">
                {{ t('update.retry') }}
              </Button>
            </template>
          </div>
        </div>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="logModalVisible">
      <DialogContent class="max-w-[900px]">
        <DialogHeader>
          <DialogTitle>{{ log.type }}</DialogTitle>
        </DialogHeader>
        <pre class="whitespace-pre-wrap text-xs bg-muted p-4 rounded-md max-h-[60vh] overflow-auto font-mono">{{ log.message
        }}</pre>
      </DialogContent>
    </Dialog>


    <Dialog v-model:open="systemModalVisible">
      <DialogContent class="max-w-[800px] overflow-hidden">
        <DialogHeader>
          <DialogTitle>{{ t('settings.basic.title') }}</DialogTitle>
        </DialogHeader>
        <div class="h-[600px] overflow-hidden">
          <app-system />
        </div>
      </DialogContent>
    </Dialog>

  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useCommentStore } from '@/stores/comment'
import { useMemoStore } from '@/stores/memo'
import { useSiteStore } from '@/stores/site'
import AppSystem from '@/views/preferences/index.vue'
import { Button } from '@/components/ui/button'
import { EventsEmit, EventsOn, BrowserOpenURL } from '@/wailsjs/runtime'
import { DeployToGit, CancelDeploy } from '@/wailsjs/go/facade/DeployFacade'
import {
  CheckUpdate,
  StartDownload,
  CancelDownload,
  ApplyUpdate,
} from '@/wailsjs/go/facade/UpdateFacade'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import WindowControls from '@/components/WindowControls/index.vue'
import {
  DocumentTextIcon,
  QueueListIcon,
  FolderIcon,
  TagIcon,
  LinkIcon,
  SwatchIcon,
  ServerIcon,
  EyeIcon,
  CloudArrowUpIcon,
  GlobeAltIcon,
  RocketLaunchIcon,
  ChatBubbleLeftRightIcon,
  CogIcon,
  LightBulbIcon,
  ArrowRightIcon,
  ArrowUpCircleIcon,
  CheckCircleIcon,
  ExclamationCircleIcon,
  XMarkIcon,
} from '@heroicons/vue/24/outline'
import pkg from '../../package.json'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const siteStore = useSiteStore()
const commentStore = useCommentStore()
const memoStore = useMemoStore()
console.log('MainLayout initialized, siteStore:', siteStore)

const version = pkg.version
const publishLoading = ref(false)
const hasUpdate = ref(false)
const newVersion = ref('')
const currentVersion = ref(pkg.version)

type UpdateState = 'idle' | 'downloading' | 'ready' | 'error'
const updateState = ref<UpdateState>('idle')
const downloadReceived = ref(0)
const downloadTotal = ref(0)
const downloadPercent = ref(0)
const updateErrorKind = ref('unknown')

const formatBytes = (n: number) => {
  if (!n || n <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let v = n
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i++
  }
  return `${v.toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}
const updateModalVisible = ref(false)
const systemModalVisible = ref(false)
const updateContent = ref('')
const logModalVisible = ref(false)
const sidebarVisible = ref(true)
const log = ref<any>({})
const cancelling = ref(false)

// 部署面板（#43，重新设计）：后端 runtime.EventsEmit("deploy-log", msg) 的行累积到
// deployLogs；一个从右侧滑入的抽屉面板呈现状态 + 进度 + 日志流，融入 app 自身调色盘。
const deployLogs = ref<string[]>([])
const deployPanelVisible = ref(false)
const logScrollEl = ref<HTMLElement | null>(null)

// 新日志到达时自动滚到底部（仅当用户没有手动向上滚查看历史时才跟随）
watch(() => deployLogs.value.length, async () => {
  await nextTick()
  const el = logScrollEl.value
  if (!el) return
  // 近底（32px 阈值）才跟随滚动；否则尊重用户当前位置
  const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 32
  if (atBottom) el.scrollTop = el.scrollHeight
})

// 从日志里解析进度：认"发现 N 个媒体文件" 作为总数，"上传 X 失败" 累计失败，
// "CDN 上传完成：成功 X / 总数 Y" 作为最终权威值。如果都没匹配上就不显示进度条。
const deployProgress = computed(() => {
  let total = 0
  let done = 0
  let failed = 0
  let label = ''
  for (const line of deployLogs.value) {
    const mStart = line.match(/发现\s+(\d+)\s+个媒体文件/)
    if (mStart) {
      total = parseInt(mStart[1], 10)
      label = 'CDN 媒体文件'
    }
    if (/^上传\s.+\s失败:/.test(line) || /^\s*✗\s/.test(line)) {
      failed++
    }
    const mEnd = line.match(/CDN 上传完成：成功\s+(\d+)\s+\/\s+总数\s+(\d+)（失败\s+(\d+)/)
    if (mEnd) {
      done = parseInt(mEnd[1], 10) + parseInt(mEnd[3], 10)
      total = parseInt(mEnd[2], 10)
      failed = parseInt(mEnd[3], 10)
    } else {
      const mEndSimple = line.match(/CDN 上传完成，共上传\s+(\d+)\s+个文件/)
      if (mEndSimple) {
        done = parseInt(mEndSimple[1], 10)
        if (total === 0) total = done
      }
    }
  }
  // 部署进行中：把已见过的"上传 X"成功/失败行加起来当作 done 的近似
  if (done === 0 && total > 0) {
    for (const line of deployLogs.value) {
      if (/^上传\s.+\s失败:/.test(line)) done++
    }
  }
  const pct = total === 0 ? 0 : Math.min(100, Math.round((done / total) * 100))
  return { total, done, failed, pct, label }
})

// 顶部 summary 文案：优先显示"XX · done/total"，否则显示最后一行日志摘要
const deploySummary = computed(() => {
  const p = deployProgress.value
  if (p.total > 0) return `${p.label} · ${p.done} / ${p.total}`
  const last = deployLogs.value[deployLogs.value.length - 1]
  if (!last) return '准备中...'
  return last.length > 48 ? last.slice(0, 48) + '…' : last
})

// 部署结果（部署中 / 成功 / 失败 / 取消）。非 running 时用来切换 chip 样式。
const deployOutcome = ref<'running' | 'success' | 'failed' | 'canceled'>('running')

// logLineIcon / logLineClass / logLineText：按日志前缀染色和图标化，不再是一团黑白文本
const logLineIcon = (entry: string): string => {
  if (/^上传\s.+\s失败/.test(entry) || /^\s*✗\s/.test(entry) || /❌/.test(entry)) return '✗'
  if (/^✅/.test(entry) || /成功/.test(entry)) return '✓'
  if (/⚠️|警告/.test(entry)) return '!'
  if (/^发现\s+\d+/.test(entry) || /开始上传/.test(entry) || /Starting|Building|Uploading|Deploying/.test(entry)) return '▸'
  return '·'
}

const logLineClass = (entry: string): string => {
  if (/^上传\s.+\s失败/.test(entry) || /^\s*✗\s/.test(entry) || /❌/.test(entry)) return 'text-destructive'
  if (/^✅/.test(entry) || /部署成功|上传完成，共上传/.test(entry)) return 'text-green-600 dark:text-green-400'
  if (/⚠️|警告/.test(entry)) return 'text-amber-600 dark:text-amber-400'
  return 'text-muted-foreground'
}

// 去掉行首的 emoji / 冗余前缀，让对齐更干净
const logLineText = (entry: string): string => entry.replace(/^(?:✅|❌|⚠️|✗|✓)\s*/, '').trim()

const clearDeployLogs = () => {
  deployLogs.value = []
  deployPanelVisible.value = false
  deployOutcome.value = 'running'
}

const currentRouter = computed(() => route.path)

const sideMenus = computed(() => {
  if (!siteStore) return []
  return [
    {
      icon: DocumentTextIcon,
      text: t('nav.article'),
      count: siteStore.posts?.length || 0,
      router: '/articles',
    },
    {
      icon: LightBulbIcon,
      text: t('nav.memo'),
      count: memoStore.totalMemos || 0,
      router: '/memos',
    },
    {
      icon: ChatBubbleLeftRightIcon,
      text: t('nav.comment'),
      count: commentStore.total || 0,
      router: '/comments',
    },
    {
      icon: QueueListIcon,
      text: t('nav.menu'),
      count: siteStore.menus?.length || 0,
      router: '/menu',
    },
    {
      icon: FolderIcon,
      text: t('nav.category'),
      count: siteStore.categories?.length || 0,
      router: '/categories',
    },
    {
      icon: TagIcon,
      text: t('nav.tag'),
      count: siteStore.tags?.length || 0,
      router: '/tags',
    },
    {
      icon: LinkIcon,
      text: t('nav.link'),
      count: siteStore.links?.length || 0,
      router: '/links',
    },
    {
      icon: SwatchIcon,
      text: t('nav.theme'),
      router: '/theme',
    },
    {
      icon: ServerIcon,
      text: t('nav.server'),
      router: '/settings',
    },
  ]
})

const clickMenu = (menu: any) => {
  router.push(menu.router)
}

const preview = () => {
  EventsEmit('preview-site')
}

const publish = async () => {
  if (publishLoading.value) return
  publishLoading.value = true
  // 每次部署重置状态：清空旧日志、切回 running 态。
  // 面板暂时不主动打开（UI 方案重新设计中），数据流和事件订阅保留。
  deployLogs.value = []
  deployOutcome.value = 'running'

  try {
    await DeployToGit()
    deployOutcome.value = 'success'
    EventsEmit('app:toast', {
      message: t('dashboard.syncSuccess'),
      type: 'success',
      duration: 3000,
    })
  } catch (error: any) {
    console.error('Deploy error:', error)
    const msg = error?.message || String(error)
    if (/canceled|cancelled|取消/i.test(msg)) {
      deployOutcome.value = 'canceled'
      EventsEmit('app:toast', { message: '已取消发布', type: 'info', duration: 2000 })
    } else {
      deployOutcome.value = 'failed'
      // 日志面板暂时隐藏，失败原因必须通过 toast 让用户看到。
      // 取 message 首行（git/HTTP 错误常带多行 stack），截断 200 字避免溢出。
      const firstLine = msg.split('\n')[0].trim()
      const shortMsg = firstLine.length > 200 ? firstLine.slice(0, 200) + '…' : firstLine
      EventsEmit('app:toast', {
        message: `部署失败：${shortMsg}`,
        type: 'error',
        duration: 8000,
      })
    }
  } finally {
    publishLoading.value = false
    cancelling.value = false
  }
}

const cancelPublish = async () => {
  if (!publishLoading.value || cancelling.value) return
  cancelling.value = true
  try {
    await CancelDeploy()
  } catch (e) {
    console.error('CancelDeploy error:', e)
    cancelling.value = false
  }
}

const copyDeployLogs = async () => {
  try {
    await navigator.clipboard.writeText(deployLogs.value.join('\n'))
    EventsEmit('app:toast', { message: '已复制部署日志', type: 'success', duration: 2000 })
  } catch {
    // 剪贴板权限缺失等场景，忽略
  }
}

const goWeb = () => {
  const domain = siteStore.currentDomain
  if (domain) {
    openInBrowser(domain)
  }
}

const handleGithubClick = () => {
  openInBrowser('https://github.com/Gridea-Pro/gridea-pro')
}

const openInBrowser = (url: string) => {
  BrowserOpenURL(url)
}

// 「跳过此版本」持久化：记录用户明确选择忽略的版本号
// 命中时：启动/轮询不亮红点、不弹窗；手动点菜单仍强制展示
const IGNORED_VERSION_KEY = 'gridea-pro:ignored-update-version'
const getIgnoredVersion = (): string => {
  try { return localStorage.getItem(IGNORED_VERSION_KEY) || '' } catch { return '' }
}
const setIgnoredVersion = (v: string) => {
  try { localStorage.setItem(IGNORED_VERSION_KEY, v) } catch (_) { /* noop */ }
}

const applyUpdateInfo = (info: any, { openDialog = false, manual = false } = {}) => {
  if (!info) return
  const ignored = getIgnoredVersion()
  const isIgnored = !manual && !!info.latestVersion && info.latestVersion === ignored

  newVersion.value = info.latestVersion || ''
  currentVersion.value = info.currentVersion || pkg.version
  updateContent.value = info.bodyHtml || ''
  hasUpdate.value = !!info.hasUpdate && !isIgnored

  if (openDialog && info.hasUpdate && !isIgnored) {
    // 打开弹窗时重置下载状态，避免上次残留
    resetDownloadState()
    updateModalVisible.value = true
  }
}

const skipThisVersion = () => {
  if (newVersion.value) setIgnoredVersion(newVersion.value)
  hasUpdate.value = false
  updateModalVisible.value = false
}

const openUpdateDialog = () => {
  resetDownloadState()
  updateModalVisible.value = true
}

const checkUpdate = async ({ manual = false, autoOpen = false } = {}) => {
  try {
    const info = await CheckUpdate()
    applyUpdateInfo(info, { openDialog: manual || autoOpen, manual })
    // 菜单手动触发、且后端确认无可用更新时，显式提示已是最新版，
    // 避免用户点击「检查更新」后完全没有反馈。
    if (manual && info && !info.hasUpdate) {
      EventsEmit('app:toast', {
        message: t('update.upToDate'),
        type: 'success',
        duration: 3000,
      })
    }
  } catch (err) {
    console.error('[checkUpdate] failed:', err)
    if (manual) {
      EventsEmit('app:toast', {
        message: String((err as any)?.message || err),
        type: 'error',
        duration: 3000,
      })
    }
  }
}

const resetDownloadState = () => {
  updateState.value = 'idle'
  downloadReceived.value = 0
  downloadTotal.value = 0
  downloadPercent.value = 0
  updateErrorKind.value = 'unknown'
}

const startUpdate = async () => {
  resetDownloadState()
  updateState.value = 'downloading'
  try {
    await StartDownload()
  } catch {
    updateState.value = 'error'
    updateErrorKind.value = 'unknown'
  }
}

const cancelUpdate = async () => {
  try { await CancelDownload() } catch (_) { /* noop */ }
  resetDownloadState()
}

const applyUpdate = async () => {
  try {
    await ApplyUpdate()
    // 后端会自行重启应用，前端不需要额外处理
  } catch {
    updateState.value = 'error'
    updateErrorKind.value = 'unknown'
  }
}

const reloadSite = () => {
  // Implement reload logic
  EventsEmit('app-site-reload')
}

const openPreferences = () => {
  EventsEmit('show-preferences')
}

onMounted(() => {
  // 启动时同步当前语言到后端，以便原生菜单使用正确的语言
  EventsEmit('app:change-locale', locale.value)

  // Listen to events
  EventsOn('app-site-loaded', (result: any) => {
    console.log('app-site-loaded', result)
    siteStore.updateSite(result)
  })

  EventsOn('log-error', (result: any) => {
    log.value = result
    logModalVisible.value = true
  })

  // 部署日志：后端 DeployService.log 会 emit 到 deploy-log。
  // publish() 开始时清空 deployLogs，这里持续 append，保证失败后用户仍可查看完整过程。
  EventsOn('deploy-log', (msg: any) => {
    if (typeof msg === 'string') {
      deployLogs.value.push(msg)
    } else if (msg && typeof msg.toString === 'function') {
      deployLogs.value.push(String(msg))
    }
  })

  // 监听首选项菜单事件
  EventsOn('show-preferences-dialog', () => {
    systemModalVisible.value = true
  })

  // ─── 原生菜单事件监听 ───

  // 文件菜单
  EventsOn('menu:new-post', () => {
    router.push('/articles?action=new')
  })
  EventsOn('menu:new-page', () => {
    router.push('/articles?action=new-page')
  })
  EventsOn('menu:save', () => {
    EventsEmit('editor:save')
  })
  EventsOn('menu:import', () => {
    console.log('[Menu] Import - TODO: 待实现')
  })
  EventsOn('menu:export', () => {
    console.log('[Menu] Export - TODO: 待实现')
  })

  // 编辑菜单：复制 HTML（find/replace 已从原生菜单移除——Tiptap 无内置查找替换）
  EventsOn('menu:copy-html', () => {
    EventsEmit('editor:copy-html')
  })

  // 视图菜单
  EventsOn('menu:toggle-sidebar', () => {
    sidebarVisible.value = !sidebarVisible.value
  })
  EventsOn('menu:toggle-preview', () => {
    EventsEmit('editor:toggle-preview')
  })
  EventsOn('menu:zoom-reset', () => {
    document.body.style.zoom = '1'
  })
  EventsOn('menu:zoom-in', () => {
    const current = parseFloat((document.body.style as any).zoom || '1')
      ; (document.body.style as any).zoom = String(Math.min(current + 0.1, 2.0))
  })
  EventsOn('menu:zoom-out', () => {
    const current = parseFloat((document.body.style as any).zoom || '1')
      ; (document.body.style as any).zoom = String(Math.max(current - 0.1, 0.5))
  })

  // 主题菜单
  EventsOn('menu:navigate', (path: string) => {
    router.push(path)
  })
  EventsOn('menu:refresh-themes', () => {
    EventsEmit('app-site-reload')
  })

  // 检查更新
  EventsOn('menu:check-update', () => {
    checkUpdate({ manual: true })
  })

  // 下载进度事件
  EventsOn('update:progress', (payload: any) => {
    downloadReceived.value = payload?.received || 0
    downloadTotal.value = payload?.total || 0
    downloadPercent.value = payload?.percent || 0
  })
  EventsOn('update:ready', () => {
    updateState.value = 'ready'
    downloadPercent.value = 100
  })
  EventsOn('update:error', (payload: any) => {
    updateState.value = 'error'
    updateErrorKind.value = payload?.kind || 'unknown'
  })

  // 原生菜单调用部署
  EventsOn('publish-site', () => {
    publish()
  })

  // Initial site load request
  EventsEmit('app-ready')

  // 启动后尝试检查更新：有新版本且不是已跳过的版本时自动弹窗
  checkUpdate({ autoOpen: true })

  // 初始化加载评论并开启全局轮询（用于更新侧边栏红点）
  commentStore.fetchComments()
  memoStore.fetchMemos()
  const commentInterval = setInterval(() => {
    // 如果不在评论页面（避免与 Index.vue 的高频轮询重叠过多），则执行低频轮询
    if (route.path !== '/comments') {
      commentStore.fetchComments()
    }
  }, 10000)

  // 每小时静默检查一次更新（跳过的版本不会弹窗）
  const updateInterval = setInterval(() => {
    checkUpdate()
  }, 60 * 60 * 1000)

  onUnmounted(() => {
    clearInterval(commentInterval)
    clearInterval(updateInterval)
  })
})

</script>

<style lang="less" scoped>
.header-spacer {
  height: 40px;
  --wails-draggable: drag;
}

/* Custom scrollbar for webkit */
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}

.scrollbar-hide {
  -ms-overflow-style: none;
  scrollbar-width: none;
}

/* Release notes typography — 适配 GitHub Release Markdown 输出 */
.release-notes :deep(h1),
.release-notes :deep(h2),
.release-notes :deep(h3),
.release-notes :deep(h4) {
  font-size: 0.8125rem;
  font-weight: 600;
  margin: 0.875rem 0 0.375rem;
  color: var(--foreground);
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.release-notes :deep(h1:first-child),
.release-notes :deep(h2:first-child),
.release-notes :deep(h3:first-child),
.release-notes :deep(h4:first-child) {
  margin-top: 0;
}

.release-notes :deep(p) {
  margin: 0.5rem 0;
}

.release-notes :deep(ul),
.release-notes :deep(ol) {
  margin: 0.5rem 0;
  padding-left: 1.25rem;
}

.release-notes :deep(li) {
  margin: 0.25rem 0;
  list-style-type: disc;
}

.release-notes :deep(li::marker) {
  color: var(--primary);
}

.release-notes :deep(a) {
  color: var(--primary);
  text-decoration: none;
}

.release-notes :deep(a:hover) {
  text-decoration: underline;
}

.release-notes :deep(code) {
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
  background: var(--muted);
  font-size: 0.75rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}

.release-notes :deep(pre) {
  padding: 0.75rem;
  border-radius: 0.375rem;
  background: var(--muted);
  overflow-x: auto;
  margin: 0.75rem 0;
}

.release-notes :deep(pre code) {
  padding: 0;
  background: transparent;
}

.release-notes :deep(hr) {
  margin: 1rem 0;
  border-color: var(--border);
}

/* 侧边栏「有新版本」呼吸指示器 —— 温和版 */
@keyframes update-breathe {
  0%,
  100% {
    transform: scale(1);
    opacity: 1;
  }
  50% {
    transform: scale(1.12);
    opacity: 0.75;
  }
}

.update-indicator {
  animation: update-breathe 1.6s ease-in-out infinite;
  transform-origin: center;
}

.update-indicator:hover {
  animation: none;
}

/* 自定义滚动条 */
.update-dialog :deep(*)::-webkit-scrollbar {
  width: 6px;
}

.update-dialog :deep(*)::-webkit-scrollbar-thumb {
  background: color-mix(in srgb, var(--primary) 20%, transparent);
  border-radius: 3px;
}

.update-dialog :deep(*)::-webkit-scrollbar-thumb:hover {
  background: color-mix(in srgb, var(--primary) 40%, transparent);
}
</style>
