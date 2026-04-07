<template>
  <div class="flex h-full min-h-[400px] bg-background rounded-xl overflow-hidden text-foreground">
    <!-- 左侧导航 -->
    <div class="w-[200px] bg-secondary/20 py-4 px-2 border-r border-border">
      <div class="flex flex-col gap-1">
        <div
v-for="item in navItems" :key="item.key"
          class="flex items-center gap-2.5 px-3.5 py-2.5 rounded-lg cursor-pointer transition-all duration-200 text-sm text-muted-foreground hover:bg-secondary/50"
          :class="{ 'bg-secondary text-foreground font-medium': activeTab === item.key }" @click="activeTab = item.key">
          <component :is="item.icon" class="size-5" />
          <span>{{ item.label }}</span>
        </div>
      </div>
    </div>

    <!-- 右侧内容 -->
    <div class="flex-1 p-8 overflow-y-auto">
      <!-- 外观设置 -->
      <div v-if="activeTab === 'appearance'" class="animate-fade-in">
        <h2 class="text-xl font-semibold mb-6 text-foreground">{{ t('settings.theme.appearance') }}</h2>
        <app-setting />
      </div>

      <!-- 语言设置 -->
      <div v-if="activeTab === 'language'" class="animate-fade-in">
        <h2 class="text-xl font-semibold mb-6 text-foreground">{{ t('common.language') }}</h2>
        <div class="flex justify-between items-center py-4 border-b border-border">
          <div class="flex-1">
            <div class="text-sm font-medium text-foreground mb-1">{{ t('settings.system.displayLanguage') }}</div>
          </div>
          <div class="w-[180px]">
            <Select :model-value="currentLanguage" @update:model-value="saveLanguage">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="option in languageOptions" :key="option.value" :value="option.value">
                  {{ option.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>

      <!-- 站点管理 -->
      <div v-if="activeTab === 'sites'" class="animate-fade-in">
        <div class="flex justify-between items-center mb-6">
          <h2 class="text-xl font-semibold text-foreground">{{ t('settings.sites.title') }}</h2>
          <Button variant="outline" size="sm" class="h-8 px-3 text-xs rounded-full" @click="handleAddSite">
            <PlusIcon class="size-4 mr-1" />
            {{ t('settings.sites.add') }}
          </Button>
        </div>

        <draggable v-model="sites" handle=".handle" item-key="path" @change="handleSiteSort">
          <template #item="{ element: site }">
            <div
              class="group flex items-center gap-3 px-4 py-3 mb-2 rounded-xl border transition-all duration-200"
              :class="site.active
                ? 'bg-primary/5 border-primary/30'
                : 'bg-card/50 border-border/50 hover:bg-primary/2 hover:border-primary/20'">
              <!-- 拖拽手柄 -->
              <div class="handle cursor-move text-muted-foreground/40 hover:text-muted-foreground">
                <Bars3Icon class="size-3.5" />
              </div>

              <!-- 站点信息 -->
              <div class="flex-1 min-w-0">
                <div class="text-sm font-medium text-foreground truncate">{{ site.name }}</div>
                <div class="text-xs text-muted-foreground/60 truncate mt-0.5">{{ site.path }}</div>
              </div>

              <!-- Switch 开关 -->
              <Switch
                :checked="site.active" size="sm"
                @update:checked="(v: boolean) => handleSwitchSite(site, v)" />

              <!-- 删除按钮 -->
              <button
                class="text-muted-foreground/30 hover:text-destructive transition-colors cursor-pointer opacity-0 group-hover:opacity-100"
                @click="handleDeleteSite(site)">
                <TrashIcon class="size-4" />
              </button>
            </div>
          </template>
        </draggable>

        <!-- 空状态 -->
        <div v-if="sites.length === 0" class="text-center py-12 text-muted-foreground">
          <div class="text-sm">{{ t('settings.sites.empty') }}</div>
        </div>
      </div>

      <!-- 关于 -->
      <div v-if="activeTab === 'about'" class="animate-fade-in">
        <h2 class="text-xl font-semibold mb-6 text-foreground">{{ t('nav.about') }}</h2>
        <div class="flex flex-col items-center py-10">
          <img src="@/assets/logo.png" class="w-20 h-20 mb-4" />
          <div class="text-xl font-semibold text-foreground mb-1">Gridea Pro</div>
          <div class="text-sm text-muted-foreground mb-4">{{ t('settings.system.version') }} {{ version }}</div>
          <a class="text-primary cursor-pointer hover:underline" @click.prevent="openWebsite">gridea.pro</a>
        </div>
      </div>
    </div>

    <!-- 添加站点对话框 -->
    <Dialog v-model:open="showAddDialog">
      <DialogContent class="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>{{ t('settings.sites.add') }}</DialogTitle>
        </DialogHeader>
        <div class="space-y-4 py-4">
          <div>
            <Label class="mb-1.5 block text-sm">{{ t('settings.sites.name') }}</Label>
            <Input v-model="newSiteName" :placeholder="t('settings.sites.namePlaceholder')" />
          </div>
          <div>
            <Label class="mb-1.5 block text-sm">{{ t('settings.sites.path') }}</Label>
            <div class="flex gap-2">
              <Input v-model="newSitePath" class="flex-1" readonly :placeholder="t('settings.sites.selectPath')" />
              <Button variant="outline" size="icon" @click="selectNewSitePath">
                <FolderOpenIcon class="size-5" />
              </Button>
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showAddDialog = false">{{ t('common.cancel') }}</Button>
          <Button :disabled="!newSiteName || !newSitePath" @click="confirmAddSite">{{ t('common.confirm') }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- 删除确认对话框 -->
    <DeleteConfirmDialog v-model:open="showDeleteDialog" :confirm-text="t('common.delete')" @confirm="confirmDeleteSite">
      <template #description>
        {{ t('settings.sites.confirmDelete') }}
      </template>
    </DeleteConfirmDialog>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { toast } from '@/helpers/toast'
import pkg from '../../../package.json'
import Draggable from 'vuedraggable'
import AppSetting from './components/AppSetting.vue'
import DeleteConfirmDialog from '@/components/ConfirmDialog/DeleteConfirmDialog.vue'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import {
  SwatchIcon,
  LanguageIcon,
  GlobeAltIcon,
  InformationCircleIcon,
  FolderOpenIcon,
  PlusIcon,
  Bars3Icon,
  TrashIcon,
} from '@heroicons/vue/24/outline'
import { EventsEmit, BrowserOpenURL } from '@/wailsjs/runtime'
import { OpenFolderDialog, GetSites, AddSite, RemoveSite, UpdateSites, SwitchSite } from '@/wailsjs/go/app/App'
import { setI18nLanguage, type LocaleType } from '@/locales'

interface SiteEntry {
  name: string
  path: string
  active: boolean
}

const { t, locale } = useI18n()
const siteStore = useSiteStore()

const activeTab = ref('appearance')
const currentLanguage = ref<LocaleType>(locale.value as LocaleType)
const version = pkg.version

// 站点管理
const sites = ref<SiteEntry[]>([])
const showAddDialog = ref(false)
const newSiteName = ref('')
const newSitePath = ref('')
const showDeleteDialog = ref(false)
const siteToDelete = ref<SiteEntry | null>(null)
const switching = ref(false)

const languageOptions: { value: LocaleType; label: string }[] = [
  { value: 'zh-CN', label: '简体中文' },
  { value: 'zh-TW', label: '繁體中文' },
  { value: 'ja-JP', label: '日本語' },
  { value: 'ko', label: '한국어' },
  { value: 'en', label: 'English' },
  { value: 'fr-FR', label: 'Français' },
  { value: 'de', label: 'Deutsch' },
  { value: 'ru', label: 'Русский' },
  { value: 'es', label: 'Español' },
  { value: 'it', label: 'Italiano' },
  { value: 'pt-BR', label: 'Português' },
]

const navItems = computed(() => [
  { key: 'appearance', icon: SwatchIcon, label: t('settings.theme.appearance') },
  { key: 'language', icon: LanguageIcon, label: t('common.language') },
  { key: 'sites', icon: GlobeAltIcon, label: t('settings.sites.title') },
  { key: 'about', icon: InformationCircleIcon, label: t('nav.about') },
])

watch(locale, (val) => {
  currentLanguage.value = val as LocaleType
})

onMounted(async () => {
  currentLanguage.value = locale.value as LocaleType
  await loadSites()
})

const loadSites = async () => {
  try {
    const result = await GetSites()
    sites.value = result || []
  } catch (e) {
    console.error('Failed to load sites:', e)
  }
}

const saveLanguage = async (val: string) => {
  const newLocale = val as LocaleType
  await setI18nLanguage(newLocale)
  EventsEmit('app:change-locale', newLocale)
  toast.success(t('common.saved'))
}

// 站点管理
const handleAddSite = () => {
  newSiteName.value = ''
  newSitePath.value = ''
  showAddDialog.value = true
}

const selectNewSitePath = async () => {
  const filePaths = await OpenFolderDialog()
  if (filePaths && filePaths.length > 0) {
    newSitePath.value = filePaths[0].replace(/\\/g, '/')
    // 如果名称为空，用文件夹名作为默认名称
    if (!newSiteName.value) {
      const parts = newSitePath.value.split('/')
      newSiteName.value = parts[parts.length - 1] || ''
    }
  }
}

const confirmAddSite = async () => {
  try {
    const result = await AddSite(newSiteName.value, newSitePath.value)
    sites.value = result || []
    showAddDialog.value = false
    toast.success(t('settings.sites.added'))
  } catch (e: any) {
    toast.error(e.message || 'Failed to add site')
  }
}

const handleSwitchSite = async (site: SiteEntry, checked: boolean) => {
  if (!checked) {
    // 尝试关闭 — 检查是否是最后一个
    const activeCount = sites.value.filter(s => s.active).length
    if (activeCount <= 1) {
      toast.warning(t('settings.sites.cannotDisableLast'))
      return
    }
  }

  if (checked && switching.value) return
  if (!checked) return // 不允许手动关闭，只能通过打开另一个来关闭

  switching.value = true
  try {
    await SwitchSite(site.path)
    // 更新本地状态
    sites.value.forEach(s => {
      s.active = s.path === site.path
    })
    toast.success(t('settings.sites.switched'))
  } catch (e: any) {
    toast.error(e.message || 'Failed to switch site')
  } finally {
    switching.value = false
  }
}

const handleDeleteSite = (site: SiteEntry) => {
  if (site.active) {
    toast.warning(t('settings.sites.cannotDeleteActive'))
    return
  }
  siteToDelete.value = site
  showDeleteDialog.value = true
}

const confirmDeleteSite = async () => {
  if (!siteToDelete.value) return
  try {
    const result = await RemoveSite(siteToDelete.value.path)
    sites.value = result || []
    toast.success(t('settings.sites.deleted'))
  } catch (e: any) {
    toast.error(e.message || 'Failed to delete site')
  }
  siteToDelete.value = null
}

const handleSiteSort = async () => {
  try {
    await UpdateSites(sites.value)
  } catch (e: any) {
    toast.error(e.message || 'Failed to save order')
  }
}

const openWebsite = () => {
  BrowserOpenURL('https://gridea.pro')
}
</script>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(5px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
