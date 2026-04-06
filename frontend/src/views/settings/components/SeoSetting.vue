<template>
  <div class="pb-20 max-w-4xl mx-auto pt-4">
    <div class="space-y-6">
      <!-- 结构化数据 -->
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.seo.jsonLD') }}</label>
        <div class="flex items-center gap-3">
          <Switch :checked="form.enableJsonLD" @update:checked="(v: boolean) => form.enableJsonLD = v" size="sm" />
          <span class="text-xs text-muted-foreground">{{ t('settings.seo.jsonLDDesc') }}</span>
        </div>
      </div>

      <!-- 社交分享 -->
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.seo.openGraph') }}</label>
        <div class="flex items-center gap-3">
          <Switch :checked="form.enableOpenGraph" @update:checked="(v: boolean) => form.enableOpenGraph = v" size="sm" />
          <span class="text-xs text-muted-foreground">{{ t('settings.seo.openGraphDesc') }}</span>
        </div>
      </div>

      <!-- Canonical URL -->
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.seo.canonicalURL') }}</label>
        <div class="flex items-center gap-3">
          <Switch :checked="form.enableCanonicalURL" @update:checked="(v: boolean) => form.enableCanonicalURL = v" size="sm" />
          <span class="text-xs text-muted-foreground">{{ t('settings.seo.canonicalURLDesc') }}</span>
        </div>
      </div>

      <!-- Meta Keywords -->
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.seo.metaKeywords') }}</label>
        <div class="max-w-sm">
          <Input v-model="form.metaKeywords" :placeholder="t('settings.seo.metaKeywordsPlaceholder')" />
        </div>
      </div>

      <!-- Google Analytics -->
      <div class="grid grid-cols-[180px_1fr] items-start gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground pt-2">{{ t('settings.seo.googleAnalytics') }}</label>
        <div class="max-w-sm">
          <Input v-model="form.googleAnalyticsId" placeholder="G-XXXXXXXXXX" />
          <div class="text-xs text-muted-foreground mt-1.5">{{ t('settings.seo.googleAnalyticsDesc') }}</div>
        </div>
      </div>

      <!-- Google Search Console -->
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.seo.googleSearchConsole') }}</label>
        <div class="max-w-sm">
          <Input v-model="form.googleSearchConsoleCode" :placeholder="t('settings.seo.googleSearchConsolePlaceholder')" />
        </div>
      </div>

      <!-- 百度统计 -->
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.seo.baiduAnalytics') }}</label>
        <div class="max-w-sm">
          <Input v-model="form.baiduAnalyticsId" :placeholder="t('settings.seo.baiduAnalyticsPlaceholder')" />
        </div>
      </div>

      <!-- 自定义 Head 代码 -->
      <div class="grid grid-cols-[180px_1fr] items-start gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground pt-2">{{ t('settings.seo.customHeadCode') }}</label>
        <div class="max-w-sm">
          <Textarea v-model="form.customHeadCode" :placeholder="t('settings.seo.customHeadCodePlaceholder')" rows="4" />
          <div class="text-xs text-muted-foreground mt-1.5">{{ t('settings.seo.customHeadCodeDesc') }}</div>
        </div>
      </div>
    </div>

    <footer-box>
      <div class="flex justify-end items-center w-full">
        <Button
          variant="default"
          class="w-18 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
          @click="submit">
          {{ t('common.save') }}
        </Button>
      </div>
    </footer-box>
  </div>
</template>

<script lang="ts" setup>
import { reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { toast } from '@/helpers/toast'
import FooterBox from '@/components/FooterBox/index.vue'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { GetSeoSetting, SaveSeoSettingFromFrontend } from '@/wailsjs/go/facade/SeoSettingFacade'
import { domain } from '@/wailsjs/go/models'

const { t } = useI18n()

const form = reactive({
  enableJsonLD: false,
  enableOpenGraph: false,
  enableCanonicalURL: false,
  metaKeywords: '',
  googleAnalyticsId: '',
  googleSearchConsoleCode: '',
  baiduAnalyticsId: '',
  customHeadCode: '',
})

onMounted(async () => {
  try {
    const setting = await GetSeoSetting()
    if (setting) {
      form.enableJsonLD = setting.enableJsonLD || false
      form.enableOpenGraph = setting.enableOpenGraph || false
      form.enableCanonicalURL = setting.enableCanonicalURL || false
      form.metaKeywords = setting.metaKeywords || ''
      form.googleAnalyticsId = setting.googleAnalyticsId || ''
      form.googleSearchConsoleCode = setting.googleSearchConsoleCode || ''
      form.baiduAnalyticsId = setting.baiduAnalyticsId || ''
      form.customHeadCode = setting.customHeadCode || ''
    }
  } catch (e) {
    console.error('Failed to load SEO settings', e)
  }
})

const submit = async () => {
  try {
    const settingDomain = new domain.SeoSetting(form)
    await SaveSeoSettingFromFrontend(settingDomain)
    toast.success(t('settings.seo.saveSuccess'))
  } catch (e) {
    console.error(e)
    toast.error(t('settings.seo.saveFailed'))
  }
}
</script>
