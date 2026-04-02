<template>
  <div class="pb-20 max-w-4xl mx-auto pt-4">
    <div class="space-y-6">
      <!-- 启用 PWA -->
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.pwa.enable') }}</label>
        <div class="flex items-center gap-3">
          <Switch :checked="form.enabled" @update:checked="(v: boolean) => form.enabled = v" size="sm" />
          <span class="text-xs text-muted-foreground">{{ t('settings.pwa.enableDesc') }}</span>
        </div>
      </div>

      <template v-if="form.enabled">
        <!-- 应用名称 -->
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.pwa.appName') }}</label>
          <div class="max-w-sm">
            <Input v-model="form.appName" :placeholder="t('settings.pwa.appNamePlaceholder')" />
          </div>
        </div>

        <!-- 短名称 -->
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.pwa.shortName') }}</label>
          <div class="max-w-sm">
            <Input v-model="form.shortName" :placeholder="t('settings.pwa.shortNamePlaceholder')" />
          </div>
        </div>

        <!-- 应用描述 -->
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.pwa.description') }}</label>
          <div class="max-w-sm">
            <Input v-model="form.description" :placeholder="t('settings.pwa.descriptionPlaceholder')" />
          </div>
        </div>

        <!-- 主题颜色 -->
        <div class="grid grid-cols-[180px_1fr] items-start gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground pt-2">{{ t('settings.pwa.themeColor') }}</label>
          <div class="max-w-sm">
            <div class="flex items-center gap-2">
              <input
                type="color"
                :value="form.themeColor"
                @input="(e: Event) => form.themeColor = (e.target as HTMLInputElement).value"
                class="w-8 h-8 rounded cursor-pointer bg-transparent p-0" />
              <Input v-model="form.themeColor" placeholder="#ffffff" class="flex-1" />
            </div>
            <div class="text-xs text-muted-foreground mt-1.5">{{ t('settings.pwa.themeColorDesc') }}</div>
          </div>
        </div>

        <!-- 背景颜色 -->
        <div class="grid grid-cols-[180px_1fr] items-start gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground pt-2">{{ t('settings.pwa.backgroundColor') }}</label>
          <div class="max-w-sm">
            <div class="flex items-center gap-2">
              <input
                type="color"
                :value="form.backgroundColor"
                @input="(e: Event) => form.backgroundColor = (e.target as HTMLInputElement).value"
                class="w-8 h-8 rounded cursor-pointer bg-transparent p-0" />
              <Input v-model="form.backgroundColor" placeholder="#ffffff" class="flex-1" />
            </div>
            <div class="text-xs text-muted-foreground mt-1.5">{{ t('settings.pwa.backgroundColorDesc') }}</div>
          </div>
        </div>

        <!-- 屏幕方向 -->
        <div class="grid grid-cols-[180px_1fr] items-start gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground pt-2">{{ t('settings.pwa.orientation') }}</label>
          <div class="w-full max-w-sm">
            <Select :model-value="form.orientation" @update:model-value="(v) => form.orientation = v as string">
              <SelectTrigger>
                <SelectValue :placeholder="t('settings.pwa.orientation')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="any">{{ t('settings.pwa.orientationAny') }}</SelectItem>
                <SelectItem value="portrait">{{ t('settings.pwa.orientationPortrait') }}</SelectItem>
                <SelectItem value="landscape">{{ t('settings.pwa.orientationLandscape') }}</SelectItem>
              </SelectContent>
            </Select>
            <div class="text-xs text-muted-foreground mt-1.5">{{ t('settings.pwa.orientationDesc') }}</div>
          </div>
        </div>

        <!-- PWA 图标 -->
        <div class="grid grid-cols-[180px_1fr] items-start gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground pt-2">{{ t('settings.pwa.icon') }}</label>
          <div class="max-w-sm">
            <div class="flex items-end gap-3">
              <div
                class="w-24 h-24 border-1 border-dashed border-input rounded-lg flex items-center justify-center cursor-pointer hover:border-primary transition-colors relative overflow-hidden bg-background"
                @click="pickIcon">
                <img v-if="iconPath" :src="iconPath" class="w-full h-full object-cover" />
                <div v-else class="flex flex-col items-center text-muted-foreground">
                  <i class="ri-add-line text-2xl mb-1"></i>
                  <span class="text-xs">Upload</span>
                </div>
              </div>
              <button
                v-if="form.customIcon"
                class="text-xs text-muted-foreground hover:text-destructive cursor-pointer transition-colors mb-1"
                @click="removeIcon">
                {{ t('settings.pwa.useAvatar') }}
              </button>
            </div>
            <div class="text-xs text-muted-foreground mt-1.5">{{ t('settings.pwa.iconDesc') }}</div>
          </div>
        </div>
      </template>
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
import { reactive, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { toast } from '@/helpers/toast'
import FooterBox from '@/components/FooterBox/index.vue'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { GetPwaSetting, SavePwaSettingFromFrontend, SavePwaIcon, HasCustomPwaIcon, RemovePwaIcon } from '@/wailsjs/go/facade/PwaSettingFacade'
import { domain } from '@/wailsjs/go/models'

const { t } = useI18n()
const siteStore = useSiteStore()

const iconPath = ref('')
const toFileUrl = (p: string) => `/local-file?path=${encodeURIComponent(p)}&t=${Date.now()}`

const form = reactive({
  enabled: false,
  appName: '',
  shortName: '',
  description: '',
  themeColor: '#ffffff',
  backgroundColor: '#ffffff',
  orientation: 'any',
  customIcon: false,
})

const updateIconPath = () => {
  if (form.customIcon) {
    iconPath.value = toFileUrl(`${siteStore.site.appDir}/images/pwa-icon.png`)
  } else {
    iconPath.value = toFileUrl(`${siteStore.site.appDir}/images/avatar.png`)
  }
}

const pickIcon = async () => {
  try {
    const selectedPath = await (window as any).go.app.App.OpenImageDialog()
    if (!selectedPath) return

    await SavePwaIcon(selectedPath)
    form.customIcon = true
    updateIconPath()
    toast.success(t('settings.pwa.iconSaved'))
  } catch (e) {
    console.error(e)
    toast.error(t('settings.pwa.iconUploadFailed'))
  }
}

const removeIcon = async () => {
  try {
    await RemovePwaIcon()
    form.customIcon = false
    updateIconPath()
  } catch (e) {
    console.error(e)
  }
}

onMounted(async () => {
  try {
    const setting = await GetPwaSetting()
    if (setting) {
      form.enabled = setting.enabled || false
      form.appName = setting.appName || ''
      form.shortName = setting.shortName || ''
      form.description = setting.description || ''
      form.themeColor = setting.themeColor || '#ffffff'
      form.backgroundColor = setting.backgroundColor || '#ffffff'
      form.orientation = setting.orientation || 'any'
      form.customIcon = setting.customIcon || false
    }
    // 检查自定义图标文件是否存在
    const hasIcon = await HasCustomPwaIcon()
    if (!hasIcon && form.customIcon) {
      form.customIcon = false
    }
    updateIconPath()
  } catch (e) {
    console.error('Failed to load PWA settings', e)
  }
})

const submit = async () => {
  try {
    const settingDomain = new domain.PwaSetting(form)
    await SavePwaSettingFromFrontend(settingDomain)
    toast.success(t('settings.pwa.saveSuccess'))
  } catch (e) {
    console.error(e)
    toast.error(t('settings.pwa.saveFailed'))
  }
}
</script>
