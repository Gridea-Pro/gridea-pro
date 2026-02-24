<template>
  <div class="flex h-screen w-full overflow-hidden bg-background text-foreground">
    <!-- Sidebar -->
    <aside v-if="sidebarVisible"
      class="w-[200px] flex-shrink-0 flex flex-col bg-sidebar border-r border-border z-10 select-none">
      <!-- Draggable Area -->
      <div class="h-10 w-full flex-shrink-0 header-spacer"></div>

      <div class="flex-1 flex flex-col overflow-hidden">
        <!-- Logo -->
        <div class="h-16 mb-[18px] flex justify-center items-center">
          <img class="h-16 rounded-full" src="@/assets/logo.png" alt="Logo">
        </div>

        <!-- Menu -->
        <nav class="flex-1 overflow-y-auto px-3 py-4 scrollbar-hide">
          <ul class="space-y-1">
            <li v-for="menu in sideMenus" :key="menu.router">
              <Button variant="ghost"
                class="w-full justify-start px-3 py-2.5 h-12 font-normal hover:bg-primary/15 cursor-pointer transition-colors"
                :class="[
                  currentRouter === menu.router
                    ? 'bg-primary/10 text-primary font-medium hover:bg-primary/15'
                    : 'text-muted-foreground hover:text-foreground'
                ]" @click="clickMenu(menu)">
                <div class="flex items-center w-full">
                  <component :is="menu.icon" v-if="menu.icon" class="mr-3 size-4 transition-colors duration-200"
                    :class="currentRouter === menu.router ? 'text-primary' : 'text-muted-foreground group-hover:text-primary'" />
                  <span class="text-xs flex-1 text-left">{{ menu.text }}</span>
                  <span v-if="menu.router === '/comments' && commentStore.unreadCount > 0"
                    class="ml-auto bg-red-500 text-white text-[9px] font-bold px-1 min-w-[14px] h-[14px] flex items-center justify-center rounded-full leading-none">
                    {{ commentStore.unreadCount > 99 ? '99+' : commentStore.unreadCount }}
                  </span>
                  <span v-if="(menu.count || 0) > 0" class="text-xs transition-colors duration-200 ml-2"
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
        <Button variant="outline"
          class="w-36 h-8 text-xs justify-center rounded-full border-primary/20 hover:bg-primary/5 cursor-pointer"
          @click="preview">
          <EyeIcon class="size-3 mr-2" />
          {{ t('nav.preview') }}
        </Button>

        <Button variant="default"
          class="w-36 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
          :disabled="publishLoading" @click="publish">
          <template v-if="publishLoading">
            <svg class="animate-spin -ml-1 mr-2 h-4 w-4 justify-center text-primary-foreground"
              xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
              </path>
            </svg>
          </template>
          <template v-else>
            <RocketLaunchIcon class="size-3 mr-2" />
            {{ t('nav.sync') }}
          </template>
        </Button>

        <div class="flex items-center justify-center gap-6 text-muted-foreground w-[80%] relative"
          :class="{ 'after:content-[\'\'] after:block after:w-1.5 after:h-1.5 after:bg-[#fa5c5c] after:rounded-full after:absolute after:top-0 after:right-5': hasUpdate }">
          <GlobeAltIcon class="size-4 cursor-pointer hover:text-primary transition-colors duration-300" @click="goWeb"
            v-if="siteStore?.setting?.domain" />

          <CogIcon class="size-4 cursor-pointer hover:text-primary transition-colors duration-300"
            @click="openPreferences" title="设置" />

          <div class="relative group" title="Star Support">
            <svg viewBox="0 0 24 24" aria-hidden="true"
              class="size-4 cursor-pointer hover:text-primary transition-colors duration-300 fill-current"
              @click="handleGithubClick">
              <path fill-rule="evenodd" clip-rule="evenodd"
                d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.335 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
            </svg>
          </div>
        </div>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 flex flex-col min-h-0 overflow-hidden bg-background">
      <!-- Draggable Area -->
      <div class="flex-1 h-8 w-full flex-shrink-0 overflow-y-auto p-0">
        <router-view v-slot="{ Component }">
          <keep-alive exclude="Loading,Theme">
            <component :is="Component" />
          </keep-alive>
        </router-view>
      </div>
    </main>

    <!-- Dialogs -->
    <Dialog v-model:open="syncErrorModalVisible">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('dashboard.syncError1') }}</DialogTitle>
        </DialogHeader>
        <div class="text-sm text-muted-foreground">
          <a href="#" class="text-primary hover:underline" @click.prevent="openInBrowser('https://gridea.dev/')">FAQ</a>
          {{
            t('common.or') }} <a href="#" class="text-primary hover:underline"
            @click.prevent="openInBrowser('https://github.com/getgridea/gridea/issues')">Issues</a> {{
              t('dashboard.syncError2') }}
        </div>
        <DialogFooter>
          <Button @click="syncErrorModalVisible = false">Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="updateModalVisible">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>🔥 New Version</DialogTitle>
        </DialogHeader>
        <div class="download-container mb-4 text-center text-sm">
          👉 <a href="#" class="text-primary hover:underline"
            @click.prevent="openInBrowser('https://gridea.dev')">Gridea
            Homepage</a> | <a href="#" class="text-primary hover:underline"
            @click.prevent="openInBrowser('https://github.com/getgridea/gridea/releases')">Github Releases</a> 👈
        </div>
        <h2 class="text-xl font-bold mb-2">{{ newVersion }}</h2>
        <div class="version-info prose prose-sm max-w-none dark:prose-invert" v-html="updateContent"></div>
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
      <DialogContent class="max-w-[800px]">
        <DialogHeader>
          <DialogTitle>{{ t('settings.basic.title') }}</DialogTitle>
        </DialogHeader>
        <div class="h-[600px]">
          <app-system />
        </div>
      </DialogContent>
    </Dialog>

  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useCommentStore } from '@/stores/comment'
import { useMemoStore } from '@/stores/memo'
import { useSiteStore } from '@/stores/site'
import AppSystem from '@/views/preferences/index.vue'
import { Button } from '@/components/ui/button'
import { EventsEmit, EventsOn, BrowserOpenURL } from '@/wailsjs/runtime'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
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
const syncErrorModalVisible = ref(false)
const updateModalVisible = ref(false)
const systemModalVisible = ref(false)
const updateContent = ref('')
const logModalVisible = ref(false)
const sidebarVisible = ref(true)
const log = ref<any>({})

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

const publish = () => {
  publishLoading.value = true
  EventsEmit('publish-site')
}

const goWeb = () => {
  if (siteStore?.setting?.domain) {
    openInBrowser(siteStore.setting.domain)
  }
}

const handleGithubClick = () => {
  openInBrowser('https://github.com/getgridea/gridea')
}

const openInBrowser = (url: string) => {
  BrowserOpenURL(url)
}

const checkUpdate = () => {
  // Implement update check logic or listen to main process
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

  // 编辑菜单
  EventsOn('menu:find', () => {
    EventsEmit('editor:find')
  })
  EventsOn('menu:replace', () => {
    EventsEmit('editor:replace')
  })
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
    checkUpdate()
  })

  // Initial site load request
  EventsEmit('app-ready')

  // 初始化加载评论并开启全局轮询（用于更新侧边栏红点）
  commentStore.fetchComments()
  memoStore.fetchMemos()
  const commentInterval = setInterval(() => {
    // 如果不在评论页面（避免与 Index.vue 的高频轮询重叠过多），则执行低频轮询
    if (route.path !== '/comments') {
      commentStore.fetchComments()
    }
  }, 10000)

  onUnmounted(() => {
    clearInterval(commentInterval)
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
</style>
