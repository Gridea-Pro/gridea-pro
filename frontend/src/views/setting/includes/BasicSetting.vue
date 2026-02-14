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
              <SelectItem v-for="p in ['github', 'netlify', 'coding', 'gitee', 'sftp']" :key="String(p)"
                :value="String(p)">
                {{ getPlatformLabel(p) }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <!-- Domain -->
      <div class="grid grid-cols-[180px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.domain') }}</label>
        <div class="flex gap-2 max-w-sm">
          <div class="w-28">
            <Select :model-value="String(protocol || '')" @update:model-value="(v) => protocol = v">
              <SelectTrigger class="">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="https://">https://</SelectItem>
                <SelectItem value="http://">http://</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Input v-model="form.domain" placeholder="mydomain.com" class="flex-1" />
        </div>
      </div>

      <!-- Netlify -->
      <template v-if="form.platform === 'netlify'">
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">Site ID</label>
          <div class="max-w-sm">
            <Input v-model="form.netlifySiteId" class="" />
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4" v-if="remoteType === 'password'">
          <label class="text-sm font-medium text-right text-muted-foreground">Access Token</label>
          <div class="relative max-w-sm">
            <Input v-model="form.netlifyAccessToken" :type="passVisible ? 'text' : 'password'" class="pr-8" />
            <component :is="passVisible ? EyeIcon : EyeSlashIcon"
              class="absolute right-2.5 top-3 w-4 h-4 cursor-pointer text-muted-foreground/70 hover:text-foreground transition-colors"
              @click="passVisible = !passVisible" />
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <div></div>
          <div>
            <a href="https://gridea.dev/netlify" target="_blank"
              class="text-primary hover:underline text-sm opacity-80 decoration-primary/50 underline-offset-4">如何配置？</a>
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
            <Input v-model="form.branch" placeholder="master" class="" />
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
        <div class="grid grid-cols-[180px_1fr] items-center gap-4" v-if="form.platform === 'coding'">
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
            <component :is="passVisible ? EyeIcon : EyeSlashIcon"
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
          <label class="text-sm font-medium text-right text-muted-foreground">Port</label>
          <div class="max-w-sm">
            <Input v-model="form.port" type="number" class="" />
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">Server</label>
          <div class="max-w-sm">
            <Input v-model="form.server" class="" />
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">Username</label>
          <div class="max-w-sm">
            <Input v-model="form.username" class="" />
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">Connect Type</label>
          <div class="w-full max-w-sm">
            <Select :model-value="String(remoteType || '')" @update:model-value="(v) => remoteType = v">
              <SelectTrigger>
                <SelectValue placeholder="Select Connect Type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="password">Password</SelectItem>
                <SelectItem value="key">SSH Key</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4" v-if="remoteType === 'password'">
          <label class="text-sm font-medium text-right text-muted-foreground">Password</label>
          <div class="relative max-w-sm">
            <Input v-model="form.password" :type="passVisible ? 'text' : 'password'" class="pr-8" />
            <component :is="passVisible ? EyeIcon : EyeSlashIcon"
              class="absolute right-2.5 top-3 w-4 h-4 cursor-pointer text-muted-foreground/70 hover:text-foreground transition-colors"
              @click="passVisible = !passVisible" />
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4" v-else>
          <label class="text-sm font-medium text-right text-muted-foreground">Private Key Path</label>
          <div class="max-w-sm">
            <Input v-model="form.privateKey" class="" />
            <div class="text-xs text-muted-foreground mt-1.5">{{ t('settings.network.privateKeyTip') }}</div>
          </div>
        </div>
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">Remote Path</label>
          <div class="max-w-sm">
            <Input v-model="form.remotePath" class="" />
            <div class="text-xs text-muted-foreground mt-1.5">{{ t('settings.network.remotePathTip') }}</div>
          </div>
        </div>
      </template>

      <!-- Proxy -->
      <template v-if="form.platform !== 'sftp'">
        <div class="grid grid-cols-[180px_1fr] items-center gap-4">
          <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.proxy') }}</label>
          <div class="w-full max-w-sm">
            <Select :model-value="String(form.enabledProxy || '')"
              @update:model-value="(v) => form.enabledProxy = v as any">
              <SelectTrigger>
                <SelectValue :placeholder="t('settings.network.proxy')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="direct">Direct</SelectItem>
                <SelectItem value="proxy">Proxy</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <template v-if="form.enabledProxy === 'proxy'">
          <div class="grid grid-cols-[180px_1fr] items-center gap-4">
            <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.proxyAddress')
            }}</label>
            <div class="max-w-sm">
              <Input v-model="form.proxyPath" class="" />
            </div>
          </div>
          <div class="grid grid-cols-[180px_1fr] items-center gap-4">
            <label class="text-sm font-medium text-right text-muted-foreground">{{ t('settings.network.proxyPort')
            }}</label>
            <div class="max-w-sm">
              <Input v-model="form.proxyPort" class="" />
            </div>
          </div>
        </template>
      </template>

    </div>

    <footer-box>
      <div class="flex justify-between items-center w-full">
        <div><!-- Optional left content --></div>
        <div class="flex gap-4">
          <Button variant="secondary" :disabled="detectLoading || !canSubmit"
            class="w-auto h-8 text-xs justify-center rounded-full border-primary/20 hover:bg-primary/5 cursor-pointer"
            @click="remoteDetect">
            <span v-if="detectLoading" class="mr-2">Checking...</span>
            {{ t('settings.network.testConnection') }}
          </Button>
          <Button variant="default" :disabled="!canSubmit"
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
import FooterBox from '@/components/FooterBox/Index.vue'
import ga from '@/helpers/analytics'
import { ISetting } from '@/interfaces/setting'
import { EyeIcon, EyeSlashIcon } from '@heroicons/vue/24/outline'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { EventsEmit, EventsOnce } from '@/wailsjs/runtime'
import { SaveSettingFromFrontend, RemoteDetectFromFrontend } from '@/wailsjs/go/facade/SettingFacade'
import { domain } from '@/wailsjs/go/models'

const { t } = useI18n()
const siteStore = useSiteStore()

const passVisible = ref(false)
const detectLoading = ref(false)
const remoteType = ref('password')
const protocol = ref('https://')

const form = reactive<ISetting>({
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
  proxyPath: '',
  proxyPort: '',
  enabledProxy: 'direct',
  netlifyAccessToken: '',
  netlifySiteId: '',
})

const getPlatformLabel = (p: string) => {
  const labels: Record<string, string> = {
    github: 'Github Pages',
    netlify: 'Netlify',
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

  return pagesPlatfomValid || sftpPlatformValid || netlifyPlatformValid
})

onMounted(() => {
  const setting = siteStore.site.setting
  console.log('setting', setting)
  Object.keys(form).forEach((key: string) => {
    const k = key as keyof ISetting
    if (key === 'domain') {
      const protocolEndIndex = setting[k].indexOf('://')
      if (protocolEndIndex !== -1) {
        form[k] = setting[k].substring(protocolEndIndex + 3)
        protocol.value = setting[k].substring(0, protocolEndIndex + 3)
      }
    } else {
      // @ts-ignore
      form[k] = setting[k]
    }
  })

  if (form.privateKey) {
    remoteType.value = 'key'
  }
})

const submit = async () => {
  const formData = {
    ...form,
    domain: `${protocol.value}${form.domain}`,
  }

  if (remoteType.value === 'password') {
    formData.privateKey = ''
  } else {
    formData.password = ''
  }

  try {
    const settingDomain = new domain.Setting(formData)
    await SaveSettingFromFrontend(settingDomain)
    EventsEmit('app-site-reload')
    toast.success(t('settings.basic.saveSuccess'))

    ga('Setting', 'Setting - save', form.platform)
  } catch (e) {
    console.error(e)
    toast.error('保存失败')
  }
}

const remoteDetect = async () => {
  const formData = {
    ...form,
    domain: `${protocol.value}${form.domain}`,
  }

  // 先保存
  try {
    const settingDomain = new domain.Setting(formData)
    await SaveSettingFromFrontend(settingDomain)
    EventsEmit('app-site-reload')

    // Wait for reload or rely on backend processing
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
    toast.error('检测失败')
    ga('Setting', 'Setting - detect-failed', form.platform)
  }
}

watch(() => form.token, (val) => {
  form.token = val.trim()
})
</script>
