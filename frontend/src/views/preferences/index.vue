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

      <!-- 源文件夹设置 -->
      <div v-if="activeTab === 'folder'" class="animate-fade-in">
        <h2 class="text-xl font-semibold mb-6 text-foreground">{{ t('settings.basic.sitePath') }}</h2>
        <div class="flex flex-col gap-3 py-4 border-b border-border">
          <div>
            <div class="text-sm font-medium text-foreground mb-1">{{ t('settings.basic.sitePath') }}</div>
            <div class="text-xs text-muted-foreground">{{ t('settings.basic.sitePathDesc') }}</div>
          </div>
          <div class="flex gap-2 w-full">
            <Input v-model="currentFolderPath" readonly class="flex-1" />
            <Button variant="outline" size="icon" @click="handleFolderSelect">
              <FolderOpenIcon class="size-5" />
            </Button>
          </div>
          <Button class="mt-2 self-start" @click="saveFolder">
            {{ t('common.save') }}
          </Button>
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
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { toast } from 'vue-sonner'
import pkg from '../../../package.json'
import AppSetting from './components/AppSetting.vue'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  SwatchIcon,
  LanguageIcon,
  FolderIcon,
  InformationCircleIcon,
  FolderOpenIcon
} from '@heroicons/vue/24/outline'
import { EventsEmit, EventsOn, EventsOnce, BrowserOpenURL } from '@/wailsjs/runtime'
import { OpenFolderDialog } from '@/wailsjs/go/app/App'
import { setI18nLanguage, type LocaleType } from '@/locales'

const { t, locale } = useI18n()
const siteStore = useSiteStore()

const activeTab = ref('appearance')
const currentLanguage = ref<LocaleType>(locale.value as LocaleType)
const currentFolderPath = ref('-')
const themeMode = ref('system')
const version = pkg.version

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
  { key: 'folder', icon: FolderIcon, label: t('settings.basic.sitePath') },
  { key: 'about', icon: InformationCircleIcon, label: t('nav.about') },
])

// Sync with actual locale if it changes externally
watch(locale, (val) => {
  currentLanguage.value = val as LocaleType
})

onMounted(() => {
  currentFolderPath.value = siteStore.appDir
  // Initialize from global state, which is already correctly set by locales/index.ts
  currentLanguage.value = locale.value as LocaleType
})

const saveLanguage = async (val: string) => {
  const newLocale = val as LocaleType
  await setI18nLanguage(newLocale)
  // 通知后端重建原生菜单以匹配新语言
  EventsEmit('app:change-locale', newLocale)
  toast.success(t('common.saved'))
}

const handleFolderSelect = async () => {
  const filePaths = await OpenFolderDialog()
  if (filePaths && filePaths.length > 0) {
    currentFolderPath.value = filePaths[0].replace(/\\/g, '/')
  }
}

const saveFolder = () => {
  EventsEmit('app-source-folder-setting', currentFolderPath.value)
  EventsOnce('app-source-folder-set', (data: any) => {
    if (data) {
      toast.success(t('common.saved'))
      EventsEmit('app-site-reload')
    } else {
      toast.error(t('settings.basic.saveError')) // TODO: Check i18n key
    }
  })
}

const openWebsite = () => {
  BrowserOpenURL('https://gridea.dev')
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
