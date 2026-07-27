<template>
  <div class="pb-24 pt-4 pl-32">
    <div class="flex flex-col md:flex-row gap-8">
      <!-- Sidebar -->
      <aside class="w-full md:w-48 flex-shrink-0 md:border-r md:border-border md:pr-6">
        <nav class="space-y-1 sticky top-0">
          <button v-for="g in groups" :key="g.key" :class="[
            'w-full h-12 text-left px-3 py-2 text-sm rounded-md transition-colors',
            activeGroup === g.key
              ? 'text-primary bg-primary/10 font-medium cursor-pointer'
              : 'text-muted-foreground hover:bg-primary/10 hover:text-foreground cursor-pointer'
          ]" @click="activeGroup = g.key">
            {{ g.label }}
          </button>
        </nav>
      </aside>

      <!-- Right Content -->
      <div class="flex-1 min-w-0">
        <div class="space-y-6 max-w-2xl">
          <div v-for="field in activeFields" :key="field.name" class="space-y-2">
            <!-- Switch -->
            <template v-if="field.type === 'switch'">
              <div class="flex justify-between items-center max-w-sm">
                <label class="text-sm font-medium text-foreground">{{ field.label }}</label>
                <Switch :checked="!!form[field.name]" @update:checked="(v: boolean) => form[field.name] = v" size="sm" />
              </div>
              <div v-if="field.desc" class="text-xs text-muted-foreground">{{ field.desc }}</div>
            </template>

            <!-- Input -->
            <template v-else-if="field.type === 'input'">
              <label class="block text-sm font-medium text-foreground">{{ field.label }}</label>
              <div class="max-w-sm">
                <Input v-model="form[field.name]" :placeholder="field.placeholder" />
              </div>
              <div v-if="field.desc" class="text-xs text-muted-foreground">{{ field.desc }}</div>
            </template>

            <!-- Textarea -->
            <template v-else-if="field.type === 'textarea'">
              <label class="block text-sm font-medium text-foreground">{{ field.label }}</label>
              <div class="max-w-sm">
                <Textarea v-model="form[field.name]" :placeholder="field.placeholder" :rows="field.rows || 4" />
              </div>
              <div v-if="field.desc" class="text-xs text-muted-foreground">{{ field.desc }}</div>
            </template>

            <!-- Picture Upload -->
            <template v-else-if="field.type === 'picture-upload'">
              <label class="block text-sm font-medium text-foreground">{{ field.label }}</label>
              <Input v-model="form[field.name]" :placeholder="field.placeholder" class="max-w-sm" />
              <div class="flex items-start gap-4">
                <div
                  class="w-24 h-24 border border-dashed border-input rounded-lg flex items-center justify-center cursor-pointer hover:border-primary transition-colors relative overflow-hidden bg-background shrink-0"
                  @mouseenter="($event.currentTarget as HTMLElement).querySelector('.delete-btn')?.classList.remove('hidden')"
                  @mouseleave="($event.currentTarget as HTMLElement).querySelector('.delete-btn')?.classList.add('hidden')"
                  @click="handleImageUpload(field.name)">
                  <img v-if="form[field.name]" :src="getImageUrl(form[field.name])"
                    class="w-full h-full object-cover" />
                  <div v-else class="flex flex-col items-center text-muted-foreground">
                    <i class="ri-upload-2-line text-2xl mb-1"></i>
                  </div>
                  <div v-if="form[field.name]"
                    class="delete-btn hidden absolute top-1 right-1 bg-red-500 hover:bg-red-600 text-white rounded-full w-5 h-5 flex items-center justify-center z-10 shadow-sm border border-white transition-colors cursor-pointer"
                    :title="t('settings.theme.removeImage')" @click.stop="form[field.name] = ''">
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-3.5 h-3.5">
                      <path d="M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z" />
                    </svg>
                  </div>
                </div>
              </div>
              <div v-if="field.desc" class="text-xs text-muted-foreground">{{ field.desc }}</div>
            </template>
          </div>
        </div>

        <footer-box>
          <div class="flex justify-end items-center w-full">
            <Button
              variant="default"
              :disabled="loadFailed"
              class="w-18 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
              @click="submit">
              {{ t('common.save') }}
            </Button>
          </div>
        </footer-box>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { reactive, ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { toast } from '@/helpers/toast'
import FooterBox from '@/components/FooterBox/index.vue'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { GetSeoSetting, SaveSeoSettingFromFrontend } from '@/wailsjs/go/facade/SeoSettingFacade'
import { UploadThemeCustomConfigImage } from '@/wailsjs/go/facade/ThemeFacade'
import { domain } from '@/wailsjs/go/models'
import { useSiteStore } from '@/stores/site'

const { t } = useI18n()
const siteStore = useSiteStore()

type FieldType = 'switch' | 'input' | 'textarea' | 'picture-upload'
interface FieldDef {
  group: string
  name: string
  type: FieldType
  label: string
  placeholder?: string
  desc?: string
  rows?: number
}

const groups = computed(() => [
  { key: 'basic', label: t('settings.seo.groupBasic') },
  { key: 'social', label: t('settings.seo.groupSocial') },
  { key: 'jsonld', label: t('settings.seo.groupJsonLd') },
  { key: 'verification', label: t('settings.seo.groupVerification') },
  { key: 'analytics', label: t('settings.seo.groupAnalytics') },
  { key: 'indexing', label: t('settings.seo.groupIndexing') },
  { key: 'custom', label: t('settings.seo.groupCustom') },
])

const fields = computed<FieldDef[]>(() => [
  // —— 基础 SEO ——
  { group: 'basic', name: 'metaKeywords', type: 'input',
    label: t('settings.seo.metaKeywords'), placeholder: t('settings.seo.metaKeywordsPlaceholder'),
    desc: t('settings.seo.metaKeywordsDesc') },
  { group: 'basic', name: 'enableCanonicalURL', type: 'switch',
    label: t('settings.seo.canonicalURL'), desc: t('settings.seo.canonicalURLDesc') },
  { group: 'basic', name: 'ogDefaultImage', type: 'picture-upload',
    label: t('settings.seo.ogDefaultImage'), placeholder: t('settings.seo.ogDefaultImagePlaceholder'),
    desc: t('settings.seo.ogDefaultImageDesc') },

  // —— 社交分享 ——
  { group: 'social', name: 'enableOpenGraph', type: 'switch',
    label: t('settings.seo.openGraph'), desc: t('settings.seo.openGraphDesc') },
  { group: 'social', name: 'twitterSite', type: 'input',
    label: t('settings.seo.twitterSite'), placeholder: '@yoursite', desc: t('settings.seo.twitterSiteDesc') },
  { group: 'social', name: 'twitterCreator', type: 'input',
    label: t('settings.seo.twitterCreator'), placeholder: '@author', desc: t('settings.seo.twitterCreatorDesc') },

  // —— 结构化数据 ——
  { group: 'jsonld', name: 'enableJsonLD', type: 'switch',
    label: t('settings.seo.jsonLD'), desc: t('settings.seo.jsonLDDesc') },

  // —— 站长平台验证 ——
  { group: 'verification', name: 'googleSearchConsoleCode', type: 'input',
    label: t('settings.seo.googleSearchConsole'), placeholder: t('settings.seo.googleSearchConsolePlaceholder'),
    desc: t('settings.seo.googleSearchConsoleDesc') },
  { group: 'verification', name: 'bingVerificationCode', type: 'input',
    label: t('settings.seo.bingVerification'), desc: t('settings.seo.bingVerificationDesc') },
  { group: 'verification', name: 'baiduVerificationCode', type: 'input',
    label: t('settings.seo.baiduVerification'), desc: t('settings.seo.baiduVerificationDesc') },
  { group: 'verification', name: '360VerificationCode', type: 'input',
    label: t('settings.seo.so360Verification'), desc: t('settings.seo.so360VerificationDesc') },
  { group: 'verification', name: 'yandexVerificationCode', type: 'input',
    label: t('settings.seo.yandexVerification'), desc: t('settings.seo.yandexVerificationDesc') },

  // —— 网站分析统计 ——
  { group: 'analytics', name: 'googleAnalyticsId', type: 'input',
    label: t('settings.seo.googleAnalytics'), placeholder: 'G-XXXXXXXXXX',
    desc: t('settings.seo.googleAnalyticsDesc') },
  { group: 'analytics', name: 'baiduAnalyticsId', type: 'input',
    label: t('settings.seo.baiduAnalytics'), placeholder: t('settings.seo.baiduAnalyticsPlaceholder'),
    desc: t('settings.seo.baiduAnalyticsDesc') },
  { group: 'analytics', name: 'plausibleDomain', type: 'input',
    label: t('settings.seo.plausibleDomain'), placeholder: 'example.com',
    desc: t('settings.seo.plausibleDomainDesc') },
  { group: 'analytics', name: 'umamiWebsiteId', type: 'input',
    label: t('settings.seo.umamiWebsiteId'), desc: t('settings.seo.umamiWebsiteIdDesc') },
  { group: 'analytics', name: 'umamiScriptUrl', type: 'input',
    label: t('settings.seo.umamiScriptUrl'), placeholder: 'https://analytics.example.com/script.js',
    desc: t('settings.seo.umamiScriptUrlDesc') },
  { group: 'analytics', name: 'cloudflareWebAnalyticsToken', type: 'input',
    label: t('settings.seo.cloudflareWebAnalytics'), desc: t('settings.seo.cloudflareWebAnalyticsDesc') },

  // —— 站点索引 ——
  { group: 'indexing', name: 'sitemapEnabled', type: 'switch',
    label: t('settings.seo.sitemapEnabled'), desc: t('settings.seo.sitemapEnabledDesc') },
  { group: 'indexing', name: 'robotsEnabled', type: 'switch',
    label: t('settings.seo.robotsEnabled'), desc: t('settings.seo.robotsEnabledDesc') },
  { group: 'indexing', name: 'robotsCustom', type: 'textarea', rows: 6,
    label: t('settings.seo.robotsCustom'), placeholder: t('settings.seo.robotsCustomPlaceholder'),
    desc: t('settings.seo.robotsCustomDesc') },

  // —— 自定义代码 ——
  { group: 'custom', name: 'customHeadCode', type: 'textarea',
    label: t('settings.seo.customHeadCode'), placeholder: t('settings.seo.customHeadCodePlaceholder'),
    desc: t('settings.seo.customHeadCodeDesc') },
  { group: 'custom', name: 'customBodyStartCode', type: 'textarea',
    label: t('settings.seo.customBodyStartCode'), placeholder: t('settings.seo.customBodyStartCodePlaceholder'),
    desc: t('settings.seo.customBodyStartCodeDesc') },
  { group: 'custom', name: 'customBodyEndCode', type: 'textarea',
    label: t('settings.seo.customBodyEndCode'), placeholder: t('settings.seo.customBodyEndCodePlaceholder'),
    desc: t('settings.seo.customBodyEndCodeDesc') },
])

const activeGroup = ref('basic')
const activeFields = computed(() => fields.value.filter(f => f.group === activeGroup.value))

const form = reactive<Record<string, any>>({
  metaKeywords: '',
  enableCanonicalURL: false,
  ogDefaultImage: '',
  enableOpenGraph: false,
  twitterSite: '',
  twitterCreator: '',
  enableJsonLD: false,
  googleSearchConsoleCode: '',
  bingVerificationCode: '',
  baiduVerificationCode: '',
  '360VerificationCode': '',
  yandexVerificationCode: '',
  googleAnalyticsId: '',
  baiduAnalyticsId: '',
  plausibleDomain: '',
  umamiWebsiteId: '',
  umamiScriptUrl: '',
  cloudflareWebAnalyticsToken: '',
  sitemapEnabled: false,
  robotsEnabled: false,
  robotsCustom: '',
  customHeadCode: '',
  customBodyStartCode: '',
  customBodyEndCode: '',
})

// 加载失败标志：后端读设置若因文件损坏/IO 失败而报错，表单会是空的。
// 此时必须禁止保存——否则空表单一提交就把磁盘上损坏但可抢救的配置覆盖成空值。
const loadFailed = ref(false)

onMounted(async () => {
  try {
    const setting = await GetSeoSetting() as Record<string, any>
    if (setting) {
      Object.keys(form).forEach((k) => {
        if (setting[k] !== undefined && setting[k] !== null) {
          form[k] = setting[k]
        }
      })
    }
  } catch (e) {
    console.error('Failed to load SEO settings', e)
    loadFailed.value = true
    toast.error(t('settings.loadFailedNoSave'))
  }
})

const submit = async () => {
  if (loadFailed.value) {
    toast.error(t('settings.loadFailedNoSave'))
    return
  }
  try {
    const settingDomain = new domain.SeoSetting(form)
    await SaveSeoSettingFromFrontend(settingDomain)
    toast.success(t('settings.seo.saveSuccess'))
  } catch (e) {
    console.error(e)
    toast.error(t('settings.seo.saveFailed'))
  }
}

// Picture upload helpers（参照 CustomSetting.vue 的实现）
const getImageUrl = (path: string) => {
  if (!path) return ''
  if (path.startsWith('http') || path.startsWith('data:')) return path
  let fullPath = path
  if (path.startsWith('/media/')) {
    fullPath = `${siteStore.site.appDir}/themes/${siteStore.site.themeConfig.themeName}/assets${path}`
  } else if (path.startsWith('/images/')) {
    fullPath = `${siteStore.site.appDir}${path}`
  }
  return `/local-file?path=${encodeURIComponent(fullPath)}`
}

const handleImageUpload = async (fieldName: string) => {
  try {
    const filePath = await (window as any).go.app.App.OpenImageDialog()
    if (!filePath) return
    const uploadedUrl = await UploadThemeCustomConfigImage(filePath)
    form[fieldName] = uploadedUrl
  } catch (error) {
    console.error('Image upload error', error)
    toast.error(`${t('settings.theme.uploadFailed')}: ${error}`)
  }
}
</script>
