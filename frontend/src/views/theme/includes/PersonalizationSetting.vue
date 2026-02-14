<template>
  <div class="pb-20 max-w-4xl mx-auto pt-4">

    <div class="space-y-6">
      <!-- Show Feature Image -->
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ $t('settings.basic.showFeatureImage')
          }}</label>
        <div>
          <Switch size="sm" v-model:checked="form.showFeatureImage" />
        </div>
      </div>

      <!-- Articles Per Page -->
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ $t('settings.basic.articlesPerPage')
          }}</label>
        <div class="max-w-sm flex items-center gap-4">
          <Slider v-model="postPageSizeArray" :min="0" :max="50" :step="5" class="flex-1" />
          <span class="text-sm w-8 text-right">{{ form.postPageSize }}</span>
        </div>
      </div>

      <!-- Archives Per Page -->
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ $t('settings.basic.archivesPerPage')
          }}</label>
        <div class="max-w-sm flex items-center gap-4">
          <Slider v-model="archivesPageSizeArray" :min="0" :max="100" :step="10" class="flex-1" />
          <span class="text-sm w-8 text-right">{{ form.archivesPageSize }}</span>
        </div>
      </div>

      <!-- URL Formats -->
      <div class="grid grid-cols-[180px_1fr] items-start gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground pt-2">{{ $t('article.defaultUrl') }}</label>
        <div class="w-full max-w-sm">
          <Select :model-value="String(form.postUrlFormat || '')"
            @update:model-value="(v: string) => form.postUrlFormat = v">
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="item in urlFormats" :key="String(item.value)" :value="String(item.value)">
                {{ item.text }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div class="grid grid-cols-[180px_1fr] items-start gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground pt-2">{{ $t('tag.defaultUrl') }}</label>
        <div class="w-full max-w-sm">
          <Select :model-value="String(form.tagUrlFormat || '')"
            @update:model-value="(v: string) => form.tagUrlFormat = v">
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="item in urlFormats" :key="String(item.value)" :value="String(item.value)">
                {{ item.text }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <!-- Paths -->
      <div class="grid grid-cols-[180px_1fr] items-start gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground pt-2">{{ $t('article.urlPath') }}</label>
        <div class="w-full max-w-sm">
          <Select v-model="postPathSelectValue">
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="post" title="example.com/post/xxx">{{ $t('settings.theme.default') }}</SelectItem>
              <!-- // TODO: Check if 'default' exists or string literal -->
              <SelectItem value="__empty__" title="example.com/xxx">{{ $t('article.concise') }}</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div class="grid grid-cols-[180px_1fr] items-start gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground pt-2">{{ $t('tag.urlPath') }}</label>
        <div class="w-full max-w-sm">
          <Select v-model="tagPathSelectValue">
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="tag" title="example.com/tag/xxx">{{ $t('settings.theme.default') }}</SelectItem>
              <!-- // TODO: Check key -->
              <SelectItem value="__empty__" title="example.com/xxx">{{ $t('article.concise') }}</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ $t('settings.basic.archivesPathPrefix')
          }}</label> <!-- // TODO: Check key -->
        <div class="max-w-sm">
          <Input v-model="form.archivesPath" :class="{ 'border-destructive': !form.archivesPath }" />
          <div v-if="!form.archivesPath" class="text-xs text-destructive mt-1">Required</div>
        </div>
      </div>

      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ $t('settings.basic.dateFormat')
          }}</label>
        <div class="max-w-sm">
          <Input v-model="form.dateFormat" />
          <div class="text-xs mt-1"><a href="#"
              @click.prevent="openPage('http://momentjs.cn/docs/#/displaying/format/')"
              class="text-primary hover:underline">Momentjs Format</a></div>
        </div>
      </div>

      <div class="grid grid-cols-[180px_1fr] items-start gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground pt-2">RSS/Feed</label>
        <div class="w-full max-w-sm">
          <Select :model-value="String(feedFullTextStr || '')" @update:model-value="(v: string) => feedFullTextStr = v">
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="true">{{ $t('article.showFullText') }}</SelectItem>
              <SelectItem value="false">{{ $t('article.showAbstract') }}</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ $t('settings.basic.rssArticles')
          }}</label>
        <div class="max-w-sm flex items-center gap-4">
          <Slider v-model="feedCountArray" :min="0" :max="50" :step="5" class="flex-1" />
          <span class="text-sm w-8 text-right">{{ form.feedCount }}</span>
        </div>
      </div>

    </div>

    <footer-box>
      <div class="flex justify-end w-full">
        <Button variant="default"
          class="w-18 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
          @click="saveTheme" :disabled="!form.archivesPath">
          {{ $t('common.save') }}
        </Button>
      </div>
    </footer-box>
  </div>
</template>

<script lang="ts" setup>
import { reactive, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useSiteStore } from '@/stores/site'
import { toast } from 'vue-sonner'
import FooterBox from '@/components/FooterBox/Index.vue'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Slider } from '@/components/ui/slider'
import {
  UrlFormats,
  DEFAULT_POST_PAGE_SIZE,
  DEFAULT_ARCHIVES_PAGE_SIZE,
  DEFAULT_FEED_COUNT,
  DEFAULT_ARCHIVES_PATH,
  DEFAULT_POST_PATH,
  DEFAULT_TAG_PATH,
} from '@/helpers/constants'
import ga from '@/helpers/analytics'
import { domain } from '@/wailsjs/go/models'
import { EventsEmit, EventsOnce, BrowserOpenURL } from '@/wailsjs/runtime'
import { SaveThemeConfigFromFrontend } from '@/wailsjs/go/facade/ThemeFacade'

const { t } = useI18n()
const router = useRouter()
const siteStore = useSiteStore()

const form = reactive({
  themeName: '',
  postPageSize: DEFAULT_POST_PAGE_SIZE,
  archivesPageSize: DEFAULT_ARCHIVES_PAGE_SIZE,
  showFeatureImage: true,
  postUrlFormat: 'SLUG',
  tagUrlFormat: 'SLUG',
  dateFormat: 'YYYY-MM-DD',
  feedFullText: true,
  feedCount: DEFAULT_FEED_COUNT,
  archivesPath: DEFAULT_ARCHIVES_PATH,
  postPath: DEFAULT_POST_PATH,
  tagPath: DEFAULT_TAG_PATH,
})

const postPageSizeArray = computed({
  get: () => [form.postPageSize],
  set: (val: number[]) => {
    form.postPageSize = val[0]
  }
})

const archivesPageSizeArray = computed({
  get: () => [form.archivesPageSize],
  set: (val: number[]) => {
    form.archivesPageSize = val[0]
  }
})

const feedCountArray = computed({
  get: () => [form.feedCount],
  set: (val: number[]) => {
    form.feedCount = val[0]
  }
})

const feedFullTextStr = computed({
  get: () => String(form.feedFullText),
  set: (val: string) => {
    form.feedFullText = val === 'true'
  }
})

const postPathSelectValue = computed({
  get: () => (form.postPath === '' ? '__empty__' : String(form.postPath || '')),
  set: (val: string) => {
    form.postPath = val === '__empty__' ? '' : val
  },
})

const tagPathSelectValue = computed({
  get: () => (form.tagPath === '' ? '__empty__' : String(form.tagPath || '')),
  set: (val: string) => {
    form.tagPath = val === '__empty__' ? '' : val
  },
})

const urlFormats = UrlFormats

const saveTheme = async () => {
  console.log('开始保存主题:', form.themeName)

  // Instantiate the class to ensure strict type safety
  const themeConfig = new domain.ThemeConfig({
    ...siteStore.site.themeConfig,
    ...form,
  })

  try {
    await SaveThemeConfigFromFrontend(themeConfig)

    toast.success(t('settings.theme.configSaved'))
    ga('Theme', 'Theme - save', form.themeName)

    // 重新加载站点数据
    EventsEmit('app-site-reload')
  } catch (e) {
    console.error('保存主题失败:', e)
    toast.error('主题保存失败')
  }
}

onMounted(() => {
  const config = siteStore.site.themeConfig

  form.themeName = config.themeName
  form.postPageSize = config.postPageSize || DEFAULT_POST_PAGE_SIZE
  form.archivesPageSize = config.archivesPageSize || DEFAULT_ARCHIVES_PAGE_SIZE
  form.showFeatureImage = config.showFeatureImage
  form.postUrlFormat = config.postUrlFormat
  form.tagUrlFormat = config.tagUrlFormat
  form.dateFormat = config.dateFormat
  form.feedFullText = config.feedFullText
  form.feedCount = config.feedCount || DEFAULT_FEED_COUNT
  form.archivesPath = config.archivesPath || DEFAULT_ARCHIVES_PATH
  form.postPath = config.postPath || DEFAULT_POST_PATH
  form.tagPath = config.tagPath || DEFAULT_TAG_PATH
})

const openPage = (url: string) => {
  BrowserOpenURL(url)
}
</script>
