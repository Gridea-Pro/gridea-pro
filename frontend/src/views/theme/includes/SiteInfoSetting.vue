<template>
  <div class="pb-20 max-w-4xl mx-auto pt-4">
    <div class="space-y-6">
      <!-- Site Name -->
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ $t('settings.basic.siteName') }}</label>
        <div class="max-w-sm">
          <Input v-model="form.siteName" />
        </div>
      </div>

      <!-- Site Description -->
      <div class="grid grid-cols-[180px_1fr] items-start gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground pt-2">{{ $t('settings.basic.siteDescription')
          }}</label>
        <div class="max-w-sm">
          <Textarea v-model="form.siteDescription" rows="3" />
          <div class="text-xs text-muted-foreground mt-1">{{ $t('article.htmlSupport') }}</div>
        </div>
      </div>

      <!-- Site Author -->
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ $t('settings.basic.siteAuthor')
          }}</label>
        <div class="max-w-sm">
          <Input v-model="form.siteAuthor" />
        </div>
      </div>

      <!-- Site Email -->
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ $t('settings.basic.siteEmail') }}</label>
        <div class="max-w-sm">
          <Input v-model="form.siteEmail" />
        </div>
      </div>

      <!-- Footer Info -->
      <div class="grid grid-cols-[180px_1fr] items-start gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground pt-2">{{ $t('settings.basic.footerInfo')
          }}</label>
        <div class="max-w-sm">
          <Textarea v-model="form.footerInfo" rows="3" placeholder="Powered by Gridea Pro" />
          <div class="text-xs text-muted-foreground mt-1">{{ $t('htmlSupport') }}</div>
        </div>
      </div>

      <!-- Favicon -->
      <div class="grid grid-cols-[180px_1fr] items-start gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground pt-2">{{ $t('settings.basic.favicon')
          }}</label>
        <div class="max-w-sm">
          <div
            class="w-24 h-24 border-1 border-dashed border-input rounded-lg flex items-center justify-center cursor-pointer hover:border-primary transition-colors relative overflow-hidden bg-background"
            @click="pickFavicon">
            <img v-if="faviconPath" :src="faviconPath" class="w-full h-full object-cover" />
            <div v-else class="flex flex-col items-center text-muted-foreground">
              <i class="ri-add-line text-2xl mb-1"></i>
              <span class="text-xs">Upload</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Avatar -->
      <div class="grid grid-cols-[180px_1fr] items-start gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground pt-2">{{ $t('settings.basic.avatar')
          }}</label>
        <div class="max-w-sm">
          <div
            class="w-24 h-24 border-1 border-dashed border-input rounded-lg flex items-center justify-center cursor-pointer hover:border-primary transition-colors relative overflow-hidden bg-background"
            @click="pickAvatar">
            <img v-if="avatarPath" :src="avatarPath" class="w-full h-full object-cover" />
            <div v-else class="flex flex-col items-center text-muted-foreground">
              <i class="ri-add-line text-2xl mb-1"></i>
              <span class="text-xs">Upload</span>
            </div>
          </div>
        </div>
      </div>

    </div>

    <footer-box>
      <div class="flex justify-end w-full">
        <Button variant="default"
          class="w-18 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
          @click="saveTheme">
          {{ $t('common.save') }}
        </Button>
      </div>
    </footer-box>
  </div>
</template>

<script lang="ts" setup>
import { reactive, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { toast } from 'vue-sonner'
import FooterBox from '@/components/FooterBox/Index.vue'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import ga from '@/helpers/analytics'
import { EventsEmit, EventsOnce } from 'wailsjs/runtime'
import { SaveFavicon, SaveAvatar } from '../../../../wailsjs/wailsjs/go/facade/SettingFacade'

const { t } = useI18n()
const siteStore = useSiteStore()

// Favicon & Avatar Logic
const faviconPath = ref('')
const avatarPath = ref('')

const toFileUrl = (p: string) => `/local-file?path=${encodeURIComponent(p)}&t=${Date.now()}`

const updateMediaPaths = () => {
  faviconPath.value = toFileUrl(`${siteStore.site.appDir}/favicon.ico`)
  avatarPath.value = toFileUrl(`${siteStore.site.appDir}/images/avatar.png`)
}

const pickFavicon = async () => {
  try {
    const defaultPath = await (window as any).go.app.App.OpenImageDialog()
    if (!defaultPath) return

    await SaveFavicon(defaultPath)
    updateMediaPaths()
    toast.success(t('settings.basic.faviconSaved'))
  } catch (e) {
    console.error(e)
    toast.error(t('uploadFailed'))
  }
}

const pickAvatar = async () => {
  try {
    const defaultPath = await (window as any).go.app.App.OpenImageDialog()
    if (!defaultPath) return

    await SaveAvatar(defaultPath)
    updateMediaPaths()
    toast.success(t('settings.basic.avatarSaved'))
  } catch (e) {
    console.error(e)
    toast.error(t('uploadFailed'))
  }
}

// Form Logic
const form = reactive({
  siteName: '',
  siteAuthor: '',
  siteEmail: '',
  siteDescription: '',
  footerInfo: '',
})

const saveTheme = () => {
  console.log('开始保存站点信息')

  let timeoutId: any = null

  EventsOnce('theme-saved', async (result: any) => {
    console.log('收到 theme-saved 事件:', result)

    if (timeoutId) {
      clearTimeout(timeoutId)
      timeoutId = null
    }

    if (!result) {
      toast.error('主题保存失败')
      return
    }

    toast.success(t('settings.theme.configSaved'))
    ga('Theme', 'SiteInfo - save', form.siteName)

    EventsEmit('app-site-reload')
  })

  timeoutId = setTimeout(() => {
    console.error('保存主题超时')
    toast.error('保存主题超时，请重试')
  }, 10000)

  // Construct full config to save
  const fullConfig = { ...siteStore.site.themeConfig }
  fullConfig.siteName = form.siteName
  fullConfig.siteAuthor = form.siteAuthor
  fullConfig.siteEmail = form.siteEmail
  fullConfig.siteDescription = form.siteDescription
  fullConfig.footerInfo = form.footerInfo

  EventsEmit('theme-save', fullConfig)
  console.log('已发送 theme-save 事件')
}

onMounted(() => {
  const config = siteStore.site.themeConfig
  form.siteName = config.siteName
  form.siteAuthor = config.siteAuthor
  form.siteEmail = config.siteEmail || ''
  form.siteDescription = config.siteDescription
  form.footerInfo = config.footerInfo

  updateMediaPaths()
})
</script>
