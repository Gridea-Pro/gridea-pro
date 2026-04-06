<template>
  <div class="pb-20 max-w-4xl mx-auto pt-4">
    <div class="space-y-6">
      <!-- Platform -->
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.platform') }}</label>
        <!-- // TODO: Check i18n key -->
        <div class="w-full max-w-sm">
          <Select :model-value="String(form.platform || '')" @update:model-value="(v) => form.platform = v as any">
            <SelectTrigger>
              <SelectValue :placeholder="t('settings.network.platform')" /> <!-- // TODO: Check i18n key -->
            </SelectTrigger>
            <SelectContent>
              <SelectItem
v-for="p in ['github', 'netlify', 'vercel', 'coding', 'gitee', 'sftp']" :key="String(p)"
                :value="String(p)">
                {{ getPlatformLabel(p) }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <!-- Domain (non-SFTP platforms) -->
      <div v-if="form.platform !== 'sftp'" class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.domain') }}</label>
        <div class="relative max-w-sm">
          <span class="absolute left-3 top-2.5 text-muted-foreground text-sm">https://</span>
          <Input v-model="form.domain" placeholder="mydomain.com" class="pl-[4.5rem]" />
        </div>
      </div>

      <!-- Netlify -->
      <template v-if="form.platform === 'netlify'">
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.siteId') }}</label>
          <div class="max-w-sm">
            <Input v-model="form.netlifySiteId" class="" />
          </div>
        </div>
        <div v-if="remoteType === 'password'" class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.accessToken') }}</label>
          <div class="relative max-w-sm">
            <Input v-model="form.netlifyAccessToken" :type="passVisible ? 'text' : 'password'" class="pr-8" />
            <component
:is="passVisible ? EyeIcon : EyeSlashIcon"
              class="absolute right-2.5 top-3 w-4 h-4 cursor-pointer text-muted-foreground/70 hover:text-foreground transition-colors"
              @click="passVisible = !passVisible" />
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <div></div>
          <div>
            <a
href="https://gridea.pro/netlify" target="_blank"
              class="text-primary hover:underline text-sm opacity-80 decoration-primary/50 underline-offset-4">{{ t('settings.network.howToConfigure') }}</a>
          </div>
        </div>
      </template>

      <!-- Vercel -->
      <template v-if="form.platform === 'vercel'">
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.projectName') }}</label>
          <div class="max-w-sm">
            <Input v-model="form.repository" placeholder="my-vercel-project" class="" />
            <div class="text-xs text-muted-foreground mt-1.5">{{ t('settings.network.vercelProjectDesc') }}</div>
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.accessToken') }}</label>
          <div class="relative max-w-sm">
            <Input v-model="form.token" :type="passVisible ? 'text' : 'password'" class="pr-8" />
            <component
:is="passVisible ? EyeIcon : EyeSlashIcon"
              class="absolute right-2.5 top-3 w-4 h-4 cursor-pointer text-muted-foreground/70 hover:text-foreground transition-colors"
              @click="passVisible = !passVisible" />
            <div class="text-xs text-muted-foreground mt-1.5">{{ t('settings.network.vercelTokenDesc') }}</div>
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.customDomain') }}</label>
          <div class="max-w-sm">
            <Input v-model="form.cname" placeholder="mydomain.com" class="" />
            <div class="text-xs text-muted-foreground mt-1.5">{{ t('settings.network.vercelDomainTip') }}</div>
          </div>
        </div>
      </template>

      <!-- Git Platforms -->
      <template v-if="['github', 'coding', 'gitee'].includes(form.platform)">
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.repository')
          }}</label>
          <div class="max-w-sm">
            <Input v-model="form.repository" class="" />
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.branch') }}</label>
          <div class="max-w-sm">
            <Input v-model="form.branch" :placeholder="form.platform === 'github' ? 'main' : 'master'" class="" />
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.username')
          }}</label>
          <div class="max-w-sm">
            <Input v-model="form.username" class="" />
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.email') }}</label>
          <div class="max-w-sm">
            <Input v-model="form.email" class="" />
          </div>
        </div>
        <div v-if="form.platform === 'coding'" class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.tokenUsername')
          }}</label>
          <div class="max-w-sm">
            <Input v-model="form.tokenUsername" class="" />
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.token') }}</label>
          <div class="relative max-w-sm">
            <Input v-model="form.token" :type="passVisible ? 'text' : 'password'" class="pr-8" />
            <component
:is="passVisible ? EyeIcon : EyeSlashIcon"
              class="absolute right-2.5 top-3 w-4 h-4 cursor-pointer text-muted-foreground/70 hover:text-foreground transition-colors"
              @click="passVisible = !passVisible" />
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">CNAME</label>
          <div class="max-w-sm">
            <Input v-model="form.cname" placeholder="mydomain.com" class="" />
          </div>
        </div>
      </template>

      <!-- SFTP -->
      <template v-if="form.platform === 'sftp'">
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.server') }}</label>
          <div class="max-w-sm">
            <Input v-model="form.server" placeholder="192.168.1.100" class="" />
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.port') }}</label>
          <div class="max-w-sm">
            <Input v-model="form.port" type="number" placeholder="22" class="" />
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.connectType') }}</label>
          <div class="w-full max-w-sm">
            <Select :model-value="String(remoteType || '')" @update:model-value="(v) => remoteType = v">
              <SelectTrigger>
                <SelectValue :placeholder="t('settings.network.connectType')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="password">Password</SelectItem>
                <SelectItem value="key">SSH Key</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.sftpUsername') }}</label>
          <div class="max-w-sm">
            <Input v-model="form.username" class="" />
          </div>
        </div>
        <div v-if="remoteType === 'password'" class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.password') }}</label>
          <div class="relative max-w-sm">
            <Input v-model="form.password" :type="passVisible ? 'text' : 'password'" class="pr-8" />
            <component
:is="passVisible ? EyeIcon : EyeSlashIcon"
              class="absolute right-2.5 top-3 w-4 h-4 cursor-pointer text-muted-foreground/70 hover:text-foreground transition-colors"
              @click="passVisible = !passVisible" />
          </div>
        </div>
        <div v-else class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.privateKeyPath') }}</label>
          <div class="max-w-sm">
            <div class="flex gap-2">
              <Input v-model="form.privateKey" class="flex-1" readonly :placeholder="t('settings.network.selectKeyFile')" />
              <Button variant="outline" size="icon" @click="selectKeyFile">
                <FolderOpenIcon class="size-5" />
              </Button>
            </div>
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.remotePath') }}</label>
          <div class="max-w-sm">
            <Input v-model="form.remotePath" class="" />
            <div class="text-xs text-muted-foreground mt-1.5">{{ t('settings.network.remotePathTip') }}</div>
          </div>
        </div>
        <!-- SFTP: Domain at the end -->
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.domain') }}</label>
          <div class="relative max-w-sm">
            <span class="absolute left-3 top-2.5 text-muted-foreground text-sm">https://</span>
            <Input v-model="form.domain" placeholder="myblog.com" class="pl-[4.5rem]" />
          </div>
        </div>
      </template>

      <!-- Proxy Settings -->
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.proxyEnabled') }}</label>
        <div class="flex items-center gap-3">
          <Switch :checked="form.proxyEnabled" @update:checked="(v: boolean) => form.proxyEnabled = v" size="sm" />
          <span class="text-xs text-muted-foreground">{{ t('settings.network.proxyEnabledDesc') }}</span>
        </div>
      </div>
      <div class="grid grid-cols-[180px_1fr] items-start gap-4" v-if="form.proxyEnabled">
        <label class="text-sm font-medium text-right text-muted-foreground pt-2">{{ t('settings.network.proxyURL') }}</label>
        <div class="max-w-sm">
          <Input v-model="form.proxyURL" placeholder="http://127.0.0.1:7890" />
          <div class="text-xs text-muted-foreground mt-1.5">{{ t('settings.network.proxyURLDesc') }}</div>
          <div v-if="proxyURLError" class="text-xs text-destructive mt-1">{{ proxyURLError }}</div>
        </div>
      </div>

    </div>

    <footer-box>
      <div class="flex justify-between items-center w-full">
        <div><!-- Optional left content --></div>
        <div class="flex gap-4">
          <Button
variant="outline" :disabled="detectLoading || !canSubmit"
            class="w-auto h-8 text-xs justify-center rounded-full border border-primary/20 text-primary/80 hover:bg-primary/5 hover:text-primary cursor-pointer"
            @click="remoteDetect">
            {{ detectLoading ? t('settings.network.checking') : t('settings.network.testConnection') }}
          </Button>
          <Button
variant="default" :disabled="!canSubmit"
            class="w-18 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
            @click="submit">
            {{ t('common.save') }}
          </Button>
        </div>
      </div>
    </footer-box>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { toast } from '@/helpers/toast'
import FooterBox from '@/components/FooterBox/index.vue'
import ga from '@/helpers/analytics'
import type { ISettingForm } from '@/interfaces/setting'
import { EyeIcon, EyeSlashIcon, FolderOpenIcon } from '@heroicons/vue/24/outline'
import { Switch } from '@/components/ui/switch'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { EventsEmit, EventsOnce } from '@/wailsjs/runtime'
import { SaveSettingFromFrontend, RemoteDetectFromFrontend } from '@/wailsjs/go/facade/SettingFacade'
import { OpenKeyFileDialog } from '@/wailsjs/go/app/App'
import { domain } from '@/wailsjs/go/models'

const { t } = useI18n()
const siteStore = useSiteStore()

const passVisible = ref(false)
const detectLoading = ref(false)
const remoteType = ref('password')

const selectKeyFile = async () => {
  const filePath = await OpenKeyFileDialog()
  if (filePath) {
    form.privateKey = filePath
  }
}

// 每个平台的专属字段（切换时独立保存/恢复）
const platformFields: Record<string, string[]> = {
  github: ['domain', 'repository', 'branch', 'username', 'email', 'tokenUsername', 'token', 'cname'],
  gitee: ['domain', 'repository', 'branch', 'username', 'email', 'tokenUsername', 'token', 'cname'],
  coding: ['domain', 'repository', 'branch', 'username', 'email', 'tokenUsername', 'token', 'cname'],
  netlify: ['domain', 'netlifySiteId', 'netlifyAccessToken'],
  vercel: ['domain', 'repository', 'token', 'cname'],
  sftp: ['domain', 'server', 'port', 'username', 'password', 'privateKey', 'remotePath'],
}

// 平台配置缓存
const platformConfigs = ref<Record<string, Record<string, any>>>({})

const form = reactive<ISettingForm>({
  platform: 'github',
  domain: '',
  repository: '',
  branch: '',
  username: '',
  email: '',
  tokenUsername: '',
  token: '',
  cname: '',
  port: '22',
  server: '',
  password: '',
  privateKey: '',
  remotePath: '',
  netlifyAccessToken: '',
  netlifySiteId: '',
  proxyEnabled: false,
  proxyURL: '',
})

// 将当前表单的平台专属字段保存到 platformConfigs
const savePlatformConfig = (platform: string) => {
  const fields = platformFields[platform] || []
  const config: Record<string, any> = {}
  for (const field of fields) {
    if (field === 'domain') {
      // domain 保存时固定 https:// 前缀，去掉尾部斜杠
      const d = form.domain ? form.domain.replace(/\/+$/, '') : ''
      config[field] = d ? `https://${d}` : ''
    } else {
      config[field] = form[field] || ''
    }
  }
  platformConfigs.value[platform] = config
}

// 从 platformConfigs 恢复平台专属字段到表单
const restorePlatformConfig = (platform: string) => {
  // 先清空所有平台专属字段
  const allPlatformFields = new Set<string>()
  for (const fields of Object.values(platformFields)) {
    for (const f of fields) allPlatformFields.add(f)
  }
  for (const field of allPlatformFields) {
    ;(form as any)[field] = ''
  }
  // 恢复 SFTP 默认端口
  if (platform === 'sftp') {
    form.port = '22'
  }

  // 恢复目标平台保存的配置
  const config = platformConfigs.value[platform]
  if (config) {
    for (const [key, val] of Object.entries(config)) {
      if (key === 'domain') {
        // domain 去掉协议前缀
        const domainVal = val || ''
        const idx = domainVal.indexOf('://')
        form.domain = idx !== -1 ? domainVal.substring(idx + 3) : domainVal
      } else {
        ;(form as any)[key] = val || ''
      }
    }
  }
}

// 监听平台切换
let skipWatch = false
watch(() => form.platform, (newPlatform, oldPlatform) => {
  if (skipWatch || !oldPlatform || newPlatform === oldPlatform) return
  savePlatformConfig(oldPlatform)
  restorePlatformConfig(newPlatform)

  // 重置密码可见性和认证类型
  passVisible.value = false
  if (newPlatform === 'sftp') {
    const config = platformConfigs.value[newPlatform]
    remoteType.value = config?.privateKey ? 'key' : 'password'
  }
})

const getPlatformLabel = (p: string) => {
  const labels: Record<string, string> = {
    github: 'Github Pages',
    netlify: 'Netlify',
    vercel: 'Vercel',
    coding: 'Coding Pages',
    gitee: 'Gitee Pages',
    sftp: 'SFTP'
  }
  return labels[p] || p
}

const canSubmit = computed(() => {
  const baseValid = form.domain
    && form.repository
    && form.branch
    && form.username
    && form.token
  const pagesPlatfomValid = baseValid && (form.platform === 'gitee' || form.platform === 'github' || (form.platform === 'coding' && form.tokenUsername))

  const sftpPlatformValid = ['sftp'].includes(form.platform)
    && form.port
    && form.server
    && form.username
    && form.remotePath
    && (form.password || form.privateKey)

  const netlifyPlatformValid = ['netlify'].includes(form.platform)
    && form.netlifyAccessToken
    && form.netlifySiteId

  const vercelPlatformValid = ['vercel'].includes(form.platform)
    && form.repository
    && form.token

  const proxyValid = !form.proxyEnabled || !form.proxyURL || proxyURLError.value === ''
  return (pagesPlatfomValid || sftpPlatformValid || netlifyPlatformValid || vercelPlatformValid) && proxyValid
})

const proxyURLError = computed(() => {
  if (!form.proxyEnabled || !form.proxyURL) return ''
  try {
    const u = new URL(form.proxyURL)
    const validSchemes = ['http:', 'https:', 'socks4:', 'socks4a:', 'socks5:', 'socks:']
    if (!validSchemes.includes(u.protocol)) {
      return t('settings.network.proxyURLInvalid')
    }
    return ''
  } catch {
    return t('settings.network.proxyURLInvalid')
  }
})

onMounted(() => {
  const setting = siteStore.site.setting
  skipWatch = true

  // 1. 恢复平台选择
  form.platform = setting.platform || 'github'

  // 2. 恢复平台配置
  if (setting.platformConfigs) {
    platformConfigs.value = JSON.parse(JSON.stringify(setting.platformConfigs))
  }

  // 3. 从 platformConfigs 恢复当前平台的专属字段到表单（包括 domain）
  restorePlatformConfig(form.platform)

  // 4. 恢复代理设置
  form.proxyEnabled = setting.proxyEnabled || false
  form.proxyURL = setting.proxyURL || ''

  // 5. domain 去掉协议前缀（restorePlatformConfig 已处理，这里兜底）
  const domainVal = form.domain || ''
  const protocolEndIndex = domainVal.indexOf('://')
  if (protocolEndIndex !== -1) {
    form.domain = domainVal.substring(protocolEndIndex + 3)
  }

  if (form.privateKey) {
    remoteType.value = 'key'
  }

  skipWatch = false
})

const buildFormData = () => {
  // 保存当前平台配置到 platformConfigs
  savePlatformConfig(form.platform)

  // SFTP 认证类型处理：清除未使用的凭据
  const configs = JSON.parse(JSON.stringify(platformConfigs.value))
  if (form.platform === 'sftp' && configs['sftp']) {
    if (remoteType.value === 'password') {
      configs['sftp'].privateKey = ''
    } else {
      configs['sftp'].password = ''
    }
  }

  return {
    platform: form.platform,
    platformConfigs: configs,
    proxyEnabled: form.proxyEnabled,
    proxyURL: form.proxyURL,
  }
}

const submit = async () => {
  try {
    const formData = buildFormData()
    const settingDomain = new domain.Setting(formData)
    await SaveSettingFromFrontend(settingDomain)
    EventsEmit('app-site-reload')
    toast.success(t('settings.basic.saveSuccess'))

    ga('Setting', 'Setting - save', form.platform)
  } catch (e) {
    console.error(e)
    toast.error(t('settings.network.saveFailed'))
  }
}

const remoteDetect = async () => {
  try {
    const formData = buildFormData()
    const settingDomain = new domain.Setting(formData)
    await SaveSettingFromFrontend(settingDomain)

    detectLoading.value = true
    ga('Setting', 'Setting - detect', form.platform)

    const result = await RemoteDetectFromFrontend(settingDomain)
    console.log('检测结果', result)
    detectLoading.value = false

    if (result && result.success) {
      toast.success(t('settings.network.connectSuccess'))
      ga('Setting', 'Setting - detect-success', form.platform)
    } else {
      toast.error(t('settings.network.connectFailed'))
      ga('Setting', 'Setting - detect-failed', form.platform)
    }

  } catch (e) {
    console.error(e)
    detectLoading.value = false
    toast.error(t('settings.network.detectFailed'))
    ga('Setting', 'Setting - detect-failed', form.platform)
  }
}

watch(() => form.token, (val) => {
  form.token = val.trim()
})
</script>
