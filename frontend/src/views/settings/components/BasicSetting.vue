<template>
  <div class="pb-20 max-w-3xl mx-auto pt-2">

    <!-- ── 平台列表 ───────────────────────────────────────────── -->
    <div class="space-y-2">
      <div
        v-for="p in platforms" :key="p.id"
        class="rounded-xl border transition-all duration-150"
        :class="activePlatform === p.id
          ? 'border-primary/40 shadow-[0_0_0_3px_rgba(var(--primary-rgb),0.08)]'
          : 'border-border/60 hover:border-border'">

        <!-- 主行 -->
        <div class="flex items-center gap-3 px-4 py-3.5">
          <!-- 图标 -->
          <div class="size-9 rounded-lg flex items-center justify-center flex-shrink-0 text-white text-sm font-bold"
            :style="{ background: p.color }">
            <component :is="p.icon" class="size-4.5" />
          </div>

          <!-- 名称 + 描述 -->
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="text-sm font-semibold text-foreground">{{ p.name }}</span>
              <span v-if="activePlatform === p.id"
                class="text-[10px] font-semibold px-1.5 py-0.5 rounded-full bg-primary/10 text-primary leading-none">
                {{ t('settings.network.activeLabel') }}
              </span>
            </div>
            <div class="text-xs text-muted-foreground mt-0.5 truncate">{{ p.description }}</div>
          </div>

          <!-- 状态 + 操作 -->
          <div class="flex items-center gap-2 flex-shrink-0">
            <!-- 已通过 OAuth 连接 -->
            <template v-if="statuses[p.id]?.connected && statuses[p.id]?.connectedVia === 'oauth'">
              <div class="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-green-500/10 text-green-600 dark:text-green-400 text-xs font-medium">
                <span class="size-1.5 rounded-full bg-green-500 inline-block"></span>
                {{ statuses[p.id].username || t('settings.network.connected') }}
              </div>
              <Button v-if="activePlatform !== p.id" variant="outline" size="sm"
                class="h-7 text-xs rounded-full px-3" @click="setActive(p.id)">
                {{ t('settings.network.setAsActive') }}
              </Button>
              <Button variant="outline" size="sm" class="h-7 text-xs rounded-full px-3"
                @click="openDrawer(p.id)">
                <Cog6ToothIcon class="size-3.5" />
              </Button>
              <Button variant="ghost" size="sm"
                class="h-7 text-xs rounded-full px-3 text-destructive hover:text-destructive"
                @click="handleRevoke(p.id)">
                {{ t('settings.network.disconnect') }}
              </Button>
            </template>

            <!-- 手动 Token 已保存 -->
            <template v-else-if="statuses[p.id]?.connected && statuses[p.id]?.connectedVia === 'manual'">
              <div class="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-amber-500/10 text-amber-600 dark:text-amber-400 text-xs font-medium">
                <span class="size-1.5 rounded-full bg-amber-500 inline-block"></span>
                {{ t('settings.network.tokenSaved') }}
              </div>
              <Button v-if="activePlatform !== p.id" variant="outline" size="sm"
                class="h-7 text-xs rounded-full px-3" @click="setActive(p.id)">
                {{ t('settings.network.setAsActive') }}
              </Button>
              <Button variant="outline" size="sm" class="h-7 text-xs rounded-full px-3"
                @click="openDrawer(p.id)">
                {{ t('settings.network.editConfig') }}
              </Button>
            </template>

            <!-- 未连接 -->
            <template v-else>
              <div class="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-muted text-muted-foreground text-xs font-medium">
                <span class="size-1.5 rounded-full bg-muted-foreground/40 inline-block"></span>
                {{ t('settings.network.notConnected') }}
              </div>
              <Button v-if="p.oauthAvailable && !oauthLoading[p.id]"
                variant="default" size="sm" class="h-7 text-xs rounded-full px-3"
                @click="handleOAuth(p.id)">
                {{ t('settings.network.authorizeWith') }}
              </Button>
              <Button v-if="p.oauthAvailable && oauthLoading[p.id]"
                variant="default" size="sm" class="h-7 text-xs rounded-full px-3" disabled>
                <ArrowPathIcon class="size-3.5 animate-spin mr-1" />
                {{ t('settings.network.waitingAuth') }}
              </Button>
              <Button :variant="p.oauthAvailable ? 'outline' : 'default'"
                size="sm" class="h-7 text-xs rounded-full px-3"
                @click="openDrawer(p.id)">
                {{ t('settings.network.configure') }}
              </Button>
            </template>
          </div>
        </div>
      </div>
    </div>

    <!-- ── 代理设置 ───────────────────────────────────────────── -->
    <div class="mt-6 rounded-xl border border-border/60 px-4 py-4 space-y-4">
      <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wide">
        {{ t('settings.network.proxy') }}
      </div>
      <div class="grid grid-cols-[160px_1fr] items-center gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground">
          {{ t('settings.network.proxyEnabled') }}
        </label>
        <div class="flex items-center gap-3">
          <Switch :checked="proxyEnabled" @update:checked="proxyEnabled = $event" size="sm" />
          <span class="text-xs text-muted-foreground">{{ t('settings.network.proxyEnabledDesc') }}</span>
        </div>
      </div>
      <div v-if="proxyEnabled" class="grid grid-cols-[160px_1fr] items-start gap-4">
        <label class="text-sm font-medium text-right text-muted-foreground pt-2">
          {{ t('settings.network.proxyURL') }}
        </label>
        <div class="max-w-sm">
          <Input v-model="proxyURL" placeholder="http://127.0.0.1:7890" />
          <div class="text-xs text-muted-foreground mt-1.5">{{ t('settings.network.proxyURLDesc') }}</div>
          <div v-if="proxyURLError" class="text-xs text-destructive mt-1">{{ proxyURLError }}</div>
        </div>
      </div>
    </div>

    <!-- ── 底部保存按钮（仅代理设置） ──────────────────────────── -->
    <footer-box>
      <div class="flex justify-end w-full">
        <Button variant="default"
          class="h-8 text-xs rounded-full bg-primary text-background hover:bg-primary/90 px-5"
          @click="saveProxy">
          {{ t('common.save') }}
        </Button>
      </div>
    </footer-box>

    <!-- ── 手动配置抽屉 ───────────────────────────────────────── -->
    <Transition name="drawer">
      <div v-if="drawerOpen" class="fixed inset-0 z-50 flex justify-end" @click.self="closeDrawer">
        <!-- 遮罩 -->
        <div class="absolute inset-0 bg-black/30 backdrop-blur-[2px]" @click="closeDrawer" />
        <!-- 抽屉面板 -->
        <div class="relative w-[420px] h-full bg-background border-l border-border shadow-2xl flex flex-col overflow-hidden">
          <!-- 抽屉头部 -->
          <div class="flex items-center justify-between px-5 py-4 border-b border-border flex-shrink-0">
            <div class="flex items-center gap-3">
              <div class="size-7 rounded-lg flex items-center justify-center text-white text-xs"
                :style="{ background: currentPlatform?.color }">
                <component v-if="currentPlatform?.icon" :is="currentPlatform.icon" class="size-3.5" />
              </div>
              <div>
                <div class="text-sm font-semibold">{{ currentPlatform?.name }}</div>
                <div class="text-xs text-muted-foreground">{{ t('settings.network.manualConfigTitle') }}</div>
              </div>
            </div>
            <button class="size-7 flex items-center justify-center rounded-lg hover:bg-muted transition-colors"
              @click="closeDrawer">
              <XMarkIcon class="size-4 text-muted-foreground" />
            </button>
          </div>

          <!-- 抽屉内容 -->
          <div class="flex-1 overflow-y-auto px-5 py-4 space-y-4">

            <!-- SFTP 提示 -->
            <div v-if="drawerPlatform === 'sftp'"
              class="flex items-start gap-2 text-xs text-muted-foreground bg-muted/50 rounded-lg px-3 py-2.5">
              <InformationCircleIcon class="size-4 flex-shrink-0 mt-0.5" />
              {{ t('settings.network.sftpNote') }}
            </div>

            <!-- ─ GitHub / Gitee / Coding ─ -->
            <template v-if="['github', 'gitee', 'coding'].includes(drawerPlatform)">
              <FormField :label="t('settings.network.domain')" prefix="https://">
                <Input v-model="drawerForm.domain" placeholder="mydomain.com" />
              </FormField>
              <FormField :label="t('settings.network.repository')">
                <Input v-model="drawerForm.repository" placeholder="username/repo" />
              </FormField>
              <FormField :label="t('settings.network.branch')">
                <Input v-model="drawerForm.branch" :placeholder="drawerPlatform === 'github' ? 'main' : 'master'" />
              </FormField>
              <FormField :label="t('settings.network.username')">
                <Input v-model="drawerForm.username" />
              </FormField>
              <FormField :label="t('settings.network.email')">
                <Input v-model="drawerForm.email" type="email" />
              </FormField>
              <FormField v-if="drawerPlatform === 'coding'" :label="t('settings.network.tokenUsername')">
                <Input v-model="drawerForm.tokenUsername" />
              </FormField>
              <FormField :label="t('settings.network.token')">
                <PasswordInput v-model="drawerForm.token"
                  :placeholder="hasExistingCredential(drawerPlatform, 'token') ? t('settings.network.tokenPlaceholder') : ''" />
              </FormField>
              <FormField label="CNAME">
                <Input v-model="drawerForm.cname" placeholder="mydomain.com（可选）" />
              </FormField>
            </template>

            <!-- ─ Netlify ─ -->
            <template v-if="drawerPlatform === 'netlify'">
              <FormField :label="t('settings.network.domain')" prefix="https://">
                <Input v-model="drawerForm.domain" placeholder="mydomain.com" />
              </FormField>
              <FormField :label="t('settings.network.siteId')">
                <Input v-model="drawerForm.netlifySiteId" />
              </FormField>
              <FormField :label="t('settings.network.accessToken')">
                <PasswordInput v-model="drawerForm.netlifyAccessToken"
                  :placeholder="hasExistingCredential('netlify', 'netlifyAccessToken') ? t('settings.network.tokenPlaceholder') : ''" />
                <template #hint>
                  <a href="https://gridea.pro/netlify" target="_blank"
                    class="text-primary/70 hover:text-primary text-xs">
                    {{ t('settings.network.netlifyGuide') }}
                  </a>
                </template>
              </FormField>
            </template>

            <!-- ─ Vercel ─ -->
            <template v-if="drawerPlatform === 'vercel'">
              <FormField :label="t('settings.network.domain')" prefix="https://">
                <Input v-model="drawerForm.domain" placeholder="mydomain.com" />
              </FormField>
              <FormField :label="t('settings.network.projectName')">
                <Input v-model="drawerForm.repository" placeholder="my-vercel-project" />
                <template #hint>{{ t('settings.network.vercelProjectDesc') }}</template>
              </FormField>
              <FormField :label="t('settings.network.accessToken')">
                <PasswordInput v-model="drawerForm.token"
                  :placeholder="hasExistingCredential('vercel', 'token') ? t('settings.network.tokenPlaceholder') : ''" />
                <template #hint>{{ t('settings.network.vercelTokenDesc') }}</template>
              </FormField>
              <FormField :label="t('settings.network.customDomain')">
                <Input v-model="drawerForm.cname" placeholder="mydomain.com（可选）" />
                <template #hint>{{ t('settings.network.vercelDomainTip') }}</template>
              </FormField>
            </template>

            <!-- ─ SFTP / FTP ─ -->
            <template v-if="drawerPlatform === 'sftp'">
              <FormField :label="t('settings.network.transferProtocol')">
                <Select :model-value="drawerForm.transferProtocol || 'sftp'"
                  @update:model-value="handleProtocolChange">
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="sftp">SFTP</SelectItem>
                    <SelectItem value="ftp">FTP</SelectItem>
                  </SelectContent>
                </Select>
              </FormField>
              <FormField :label="t('settings.network.server')">
                <Input v-model="drawerForm.server" placeholder="192.168.1.100" />
              </FormField>
              <FormField :label="t('settings.network.port')">
                <Input v-model="drawerForm.port" type="number"
                  :placeholder="drawerForm.transferProtocol === 'ftp' ? '21' : '22'" />
              </FormField>
              <FormField v-if="drawerForm.transferProtocol !== 'ftp'" :label="t('settings.network.connectType')">
                <Select :model-value="sftpAuthType" @update:model-value="sftpAuthType = $event">
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="password">{{ t('settings.network.password') }}</SelectItem>
                    <SelectItem value="key">SSH Key</SelectItem>
                  </SelectContent>
                </Select>
              </FormField>
              <FormField :label="t('settings.network.sftpUsername')">
                <Input v-model="drawerForm.username" />
              </FormField>
              <FormField v-if="drawerForm.transferProtocol === 'ftp' || sftpAuthType === 'password'"
                :label="t('settings.network.password')">
                <PasswordInput v-model="drawerForm.password"
                  :placeholder="hasExistingCredential('sftp', 'password') ? t('settings.network.tokenPlaceholder') : ''" />
              </FormField>
              <FormField v-if="drawerForm.transferProtocol !== 'ftp' && sftpAuthType === 'key'"
                :label="t('settings.network.privateKeyPath')">
                <div class="flex gap-2">
                  <Input v-model="drawerForm.privateKey" class="flex-1" readonly
                    :placeholder="t('settings.network.selectKeyFile')" />
                  <Button variant="outline" size="icon" @click="selectKeyFile">
                    <FolderOpenIcon class="size-4" />
                  </Button>
                </div>
              </FormField>
              <FormField :label="t('settings.network.remotePath')">
                <Input v-model="drawerForm.remotePath" />
                <template #hint>{{ t('settings.network.remotePathTip') }}</template>
              </FormField>
              <FormField :label="t('settings.network.domain')" prefix="https://">
                <Input v-model="drawerForm.domain" placeholder="myblog.com" />
              </FormField>
            </template>

          </div>

          <!-- 抽屉底部按钮 -->
          <div class="flex items-center justify-between gap-3 px-5 py-4 border-t border-border flex-shrink-0">
            <Button variant="outline" size="sm" class="h-8 text-xs rounded-full px-4"
              :disabled="detectLoading"
              @click="testConnection">
              {{ detectLoading ? t('settings.network.checking') : t('settings.network.testConnection') }}
            </Button>
            <div class="flex gap-2">
              <Button variant="ghost" size="sm" class="h-8 text-xs rounded-full px-4"
                @click="closeDrawer">
                {{ t('common.cancel') }}
              </Button>
              <Button variant="default" size="sm" class="h-8 text-xs rounded-full px-4"
                :disabled="saveLoading"
                @click="saveDrawer">
                {{ saveLoading ? '...' : t('settings.network.saveAndClose') }}
              </Button>
            </div>
          </div>
        </div>
      </div>
    </Transition>

  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { toast } from '@/helpers/toast'
import FooterBox from '@/components/FooterBox/index.vue'
import { EyeIcon, EyeSlashIcon, FolderOpenIcon, Cog6ToothIcon, ArrowPathIcon, XMarkIcon, InformationCircleIcon } from '@heroicons/vue/24/outline'
import { Switch } from '@/components/ui/switch'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { EventsEmit, EventsOn, EventsOff } from '@/wailsjs/runtime'
import { SaveSettingFromFrontend, RemoteDetectFromFrontend } from '@/wailsjs/go/facade/SettingFacade'
import { GetAllStatuses, StartOAuthFlow, RevokeToken, HasCredential } from '@/wailsjs/go/facade/OAuthFacade'
import { OpenKeyFileDialog } from '@/wailsjs/go/app/App'
import { domain } from '@/wailsjs/go/models'
import type { service } from '@/wailsjs/go/models'

const { t } = useI18n()
const siteStore = useSiteStore()

// ── 平台定义 ──────────────────────────────────────────────────────────────

// 简单 SVG 图标组件（内联，避免额外依赖）
const GitHubIcon = { template: `<svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"/></svg>` }
const VercelIcon = { template: `<svg viewBox="0 0 512 512" fill="currentColor"><path d="M256 48L496 464H16L256 48z"/></svg>` }
const GenericIcon = { template: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18M9 21V9"/></svg>` }
const ServerIcon = { template: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="5" rx="1"/><rect x="2" y="10" width="20" height="5" rx="1"/><rect x="2" y="17" width="20" height="5" rx="1"/><circle cx="6" cy="5.5" r=".8" fill="currentColor"/><circle cx="6" cy="12.5" r=".8" fill="currentColor"/></svg>` }

const platforms = [
  { id: 'github',  name: 'GitHub Pages',    color: '#24292f', icon: GitHubIcon,  oauthAvailable: false, description: t('settings.network.github_desc') || '部署到 GitHub Pages，免费托管静态博客' },
  { id: 'netlify', name: 'Netlify',         color: '#00c7b7', icon: GenericIcon, oauthAvailable: false, description: '自动 CI/CD，支持自定义域名与 HTTPS' },
  { id: 'vercel',  name: 'Vercel',          color: '#000000', icon: VercelIcon,  oauthAvailable: false, description: '边缘网络加速，前端部署首选' },
  { id: 'gitee',   name: 'Gitee Pages',     color: '#c71d23', icon: GenericIcon, oauthAvailable: false, description: '国内访问速度更快，无需代理' },
  { id: 'coding',  name: 'Coding Pages',    color: '#0066ff', icon: GenericIcon, oauthAvailable: false, description: '腾讯旗下，国内稳定访问' },
  { id: 'sftp',    name: 'SFTP / FTP',      color: '#5856d6', icon: ServerIcon,  oauthAvailable: false, description: '直接部署到你自己的服务器' },
]

// ── 状态 ──────────────────────────────────────────────────────────────────

const statuses = ref<Record<string, service.PlatformStatus>>({})
const oauthLoading = ref<Record<string, boolean>>({})
const activePlatform = ref('github')
const proxyEnabled = ref(false)
const proxyURL = ref('')
const detectLoading = ref(false)
const saveLoading = ref(false)

// 抽屉状态
const drawerOpen = ref(false)
const drawerPlatform = ref('')
const sftpAuthType = ref('password')

const drawerForm = reactive<Record<string, any>>({
  domain: '',
  repository: '',
  branch: '',
  username: '',
  email: '',
  tokenUsername: '',
  token: '',
  cname: '',
  transferProtocol: 'sftp',
  port: '',
  server: '',
  password: '',
  privateKey: '',
  remotePath: '',
  netlifyAccessToken: '',
  netlifySiteId: '',
})

// ── 计算属性 ──────────────────────────────────────────────────────────────

const currentPlatform = computed(() => platforms.find(p => p.id === drawerPlatform.value))

const proxyURLError = computed(() => {
  if (!proxyEnabled.value || !proxyURL.value) return ''
  try {
    const u = new URL(proxyURL.value)
    const valid = ['http:', 'https:', 'socks4:', 'socks4a:', 'socks5:', 'socks:']
    return valid.includes(u.protocol) ? '' : t('settings.network.proxyURLInvalid')
  } catch {
    return t('settings.network.proxyURLInvalid')
  }
})

// ── 生命周期 ──────────────────────────────────────────────────────────────

onMounted(async () => {
  const setting = siteStore.site.setting
  activePlatform.value = setting.platform || 'github'
  proxyEnabled.value = setting.proxyEnabled || false
  proxyURL.value = setting.proxyURL || ''

  // 加载各平台连接状态（从 Keychain）
  await loadStatuses()

  // 更新 OAuth 可用性
  // await updateOAuthAvailability()  // 暂由 isOAuthAvailable 字段控制

  // 监听 OAuth 授权结果
  EventsOn('oauth:success', (data: any) => {
    const { provider, username, avatarUrl } = data
    oauthLoading.value[provider] = false
    statuses.value[provider] = {
      connected: true,
      connectedVia: 'oauth',
      username,
      avatarUrl,
      email: '',
    }
    toast.success(`${getPlatformName(provider)} ${t('settings.network.authSuccess')}`)
  })

  EventsOn('oauth:error', (data: any) => {
    const { provider, error } = data
    oauthLoading.value[provider] = false
    toast.error(`${t('settings.network.authFailed')}: ${error}`)
  })
})

// ── 方法 ──────────────────────────────────────────────────────────────────

async function loadStatuses() {
  try {
    const result = await GetAllStatuses()
    statuses.value = result
  } catch (e) {
    console.error('获取平台状态失败', e)
  }
}

async function handleOAuth(platformId: string) {
  oauthLoading.value[platformId] = true
  try {
    await StartOAuthFlow(platformId)
    // 等待 oauth:success 事件
  } catch (e: any) {
    oauthLoading.value[platformId] = false
    toast.error(e?.message || t('settings.network.authFailed'))
  }
}

async function handleRevoke(platformId: string) {
  if (!confirm(t('settings.network.revokeConfirm'))) return
  try {
    await RevokeToken(platformId)
    statuses.value[platformId] = { connected: false, connectedVia: '', username: '', avatarUrl: '', email: '' }
    toast.success(`${getPlatformName(platformId)} 已断开连接`)
    if (activePlatform.value === platformId) {
      activePlatform.value = ''
    }
  } catch (e: any) {
    toast.error(e?.message || '断开连接失败')
  }
}

function setActive(platformId: string) {
  activePlatform.value = platformId
  savePlatformSelection()
}

async function savePlatformSelection() {
  try {
    const setting = siteStore.site.setting
    const settingObj = new domain.Setting({
      platform: activePlatform.value,
      platformConfigs: setting.platformConfigs || {},
      proxyEnabled: proxyEnabled.value,
      proxyURL: proxyURL.value,
    })
    await SaveSettingFromFrontend(settingObj)
    EventsEmit('app-site-reload')
  } catch (e) {
    console.error(e)
  }
}

// 抽屉
function openDrawer(platformId: string) {
  drawerPlatform.value = platformId
  loadDrawerForm(platformId)
  drawerOpen.value = true
}

function closeDrawer() {
  drawerOpen.value = false
}

function loadDrawerForm(platformId: string) {
  const platformConfigs = siteStore.site.setting.platformConfigs || {}
  const cfg = platformConfigs[platformId] || {}

  // 重置
  Object.keys(drawerForm).forEach(k => { drawerForm[k] = '' })
  drawerForm.transferProtocol = 'sftp'
  drawerForm.port = platformId === 'sftp' ? '22' : ''

  // 填入已保存的非敏感字段
  for (const [k, v] of Object.entries(cfg)) {
    if (k === 'domain') {
      const d = String(v || '')
      const idx = d.indexOf('://')
      drawerForm.domain = idx !== -1 ? d.substring(idx + 3) : d
    } else {
      drawerForm[k] = v || ''
    }
  }

  // SFTP 认证类型
  if (platformId === 'sftp') {
    sftpAuthType.value = drawerForm.privateKey ? 'key' : 'password'
  }
}

function handleProtocolChange(v: string) {
  drawerForm.transferProtocol = v
  if (v === 'ftp') {
    if (!drawerForm.port || drawerForm.port === '22') drawerForm.port = '21'
    sftpAuthType.value = 'password'
  } else {
    if (!drawerForm.port || drawerForm.port === '21') drawerForm.port = '22'
  }
}

async function selectKeyFile() {
  const filePath = await OpenKeyFileDialog()
  if (filePath) drawerForm.privateKey = filePath
}

function hasExistingCredential(platform: string, field: string): boolean {
  return statuses.value[platform]?.connected ?? false
}

async function testConnection() {
  detectLoading.value = true
  try {
    const setting = buildSettingForPlatform(drawerPlatform.value)
    const settingDomain = new domain.Setting(setting)
    const result = await RemoteDetectFromFrontend(settingDomain)
    if (result?.success) {
      toast.success(t('settings.network.connectSuccess'))
    } else {
      toast.error(result?.message || t('settings.network.connectFailed'))
    }
  } catch (e: any) {
    toast.error(e?.message || t('settings.network.detectFailed'))
  } finally {
    detectLoading.value = false
  }
}

async function saveDrawer() {
  saveLoading.value = true
  try {
    const setting = buildSettingForPlatform(drawerPlatform.value)
    const settingDomain = new domain.Setting(setting)
    await SaveSettingFromFrontend(settingDomain)
    EventsEmit('app-site-reload')

    // 更新状态显示
    await loadStatuses()
    toast.success(t('settings.network.credentialSaved'))
    closeDrawer()
  } catch (e: any) {
    toast.error(e?.message || t('settings.network.saveFailed'))
  } finally {
    saveLoading.value = false
  }
}

async function saveProxy() {
  try {
    const setting = siteStore.site.setting
    const settingObj = new domain.Setting({
      platform: activePlatform.value,
      platformConfigs: setting.platformConfigs || {},
      proxyEnabled: proxyEnabled.value,
      proxyURL: proxyURL.value,
    })
    await SaveSettingFromFrontend(settingObj)
    EventsEmit('app-site-reload')
    toast.success(t('settings.basic.saveSuccess'))
  } catch (e) {
    toast.error(t('settings.network.saveFailed'))
  }
}

// ── 工具函数 ──────────────────────────────────────────────────────────────

function buildSettingForPlatform(platformId: string) {
  const existingConfigs = JSON.parse(JSON.stringify(siteStore.site.setting.platformConfigs || {}))

  const domain = drawerForm.domain ? `https://${drawerForm.domain.replace(/\/+$/, '')}` : ''

  const platformFieldMap: Record<string, string[]> = {
    github:  ['domain', 'repository', 'branch', 'username', 'email', 'tokenUsername', 'token', 'cname'],
    gitee:   ['domain', 'repository', 'branch', 'username', 'email', 'tokenUsername', 'token', 'cname'],
    coding:  ['domain', 'repository', 'branch', 'username', 'email', 'tokenUsername', 'token', 'cname'],
    netlify: ['domain', 'netlifySiteId', 'netlifyAccessToken'],
    vercel:  ['domain', 'repository', 'token', 'cname'],
    sftp:    ['domain', 'transferProtocol', 'server', 'port', 'username', 'password', 'privateKey', 'remotePath'],
  }

  const fields = platformFieldMap[platformId] || []
  const cfg: Record<string, any> = {}
  for (const f of fields) {
    if (f === 'domain') {
      cfg.domain = domain
    } else {
      cfg[f] = drawerForm[f] || ''
    }
  }

  // SFTP: 清除未使用的认证字段
  if (platformId === 'sftp') {
    if (sftpAuthType.value === 'password') cfg.privateKey = ''
    else cfg.password = ''
  }

  existingConfigs[platformId] = cfg

  return {
    platform: activePlatform.value,
    platformConfigs: existingConfigs,
    proxyEnabled: proxyEnabled.value,
    proxyURL: proxyURL.value,
  }
}

function getPlatformName(id: string) {
  return platforms.find(p => p.id === id)?.name || id
}
</script>

<!-- 小工具组件（行内定义，避免额外文件） -->
<script lang="ts">
// FormField 辅助组件
export const FormField = {
  props: ['label', 'prefix'],
  template: `
    <div class="space-y-1.5">
      <label class="text-xs font-medium text-muted-foreground">{{ label }}</label>
      <div class="relative">
        <span v-if="prefix" class="absolute left-3 top-2.5 text-muted-foreground text-sm z-10 pointer-events-none">{{ prefix }}</span>
        <div :class="prefix ? 'pl-[4.5rem]' : ''">
          <slot />
        </div>
      </div>
      <div v-if="$slots.hint" class="text-xs text-muted-foreground"><slot name="hint" /></div>
    </div>
  `
}

// PasswordInput 辅助组件
export const PasswordInput = {
  props: ['modelValue', 'placeholder'],
  emits: ['update:modelValue'],
  data() { return { visible: false } },
  template: `
    <div class="relative">
      <input :type="visible ? 'text' : 'password'"
        :value="modelValue" :placeholder="placeholder"
        @input="$emit('update:modelValue', $event.target.value)"
        class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring pr-9" />
      <button type="button" tabindex="-1"
        @click="visible = !visible"
        class="absolute right-2.5 top-2.5 text-muted-foreground/60 hover:text-muted-foreground transition-colors">
        <svg v-if="visible" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4"><path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.964-7.178Z"/><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"/></svg>
        <svg v-else xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4"><path stroke-linecap="round" stroke-linejoin="round" d="M3.98 8.223A10.477 10.477 0 0 0 1.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.451 10.451 0 0 1 12 4.5c4.756 0 8.773 3.162 10.065 7.498a10.522 10.522 0 0 1-4.293 5.774M6.228 6.228 3 3m3.228 3.228 3.65 3.65m7.894 7.894L21 21m-3.228-3.228-3.65-3.65m0 0a3 3 0 1 0-4.243-4.243m4.242 4.242L9.88 9.88"/></svg>
      </button>
    </div>
  `
}
</script>

<style scoped>
/* 抽屉动画 */
.drawer-enter-active,
.drawer-leave-active {
  transition: opacity 0.2s ease;
}
.drawer-enter-active .relative,
.drawer-leave-active .relative {
  transition: transform 0.25s cubic-bezier(0.32, 0.72, 0, 1);
}
.drawer-enter-from,
.drawer-leave-to {
  opacity: 0;
}
.drawer-enter-from .relative {
  transform: translateX(100%);
}
.drawer-leave-to .relative {
  transform: translateX(100%);
}
</style>
