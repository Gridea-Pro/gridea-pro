<template>
  <div class="pb-20 pt-4 px-4 w-full">
    <div class="space-y-10">

    <!-- ── 当前平台 ───────────────────────────────────────────── -->
    <div v-if="activePlatformData" class="space-y-4">
      <h2 class="text-sm text-primary font-medium border-l-[3px] border-primary pl-3 flex items-center h-4">
        {{ t('settings.network.currentPlatform') }}
      </h2>
      <div class="border border-primary/20 rounded-xl overflow-hidden">
        <!-- 顶部：基本信息 + 操作 -->
        <div class="flex items-center gap-5 px-6 py-5">
          <div class="size-11 rounded-lg flex items-center justify-center text-white flex-shrink-0 shadow-sm"
            :style="{ background: activePlatformData.color }">
            <component :is="activePlatformData.icon" class="size-5.5" />
          </div>

          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2.5">
              <span class="text-base font-bold text-foreground">{{ activePlatformData.name }}</span>
              <!-- 状态 chip -->
              <template v-if="activeStatus?.connected">
                <span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-green-500/10 text-green-600 dark:text-green-400 text-[11px] font-medium">
                  <span class="size-1.5 rounded-full bg-green-500 inline-block"></span>
                  {{ t('settings.network.connected') }}
                </span>
              </template>
              <template v-else>
                <span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-muted text-muted-foreground text-[11px] font-medium">
                  <span class="size-1.5 rounded-full bg-muted-foreground/40 inline-block"></span>
                  {{ t('settings.network.notConnected') }}
                </span>
              </template>
            </div>
            <div class="text-xs text-muted-foreground mt-0.5">{{ activePlatformData.description }}</div>
          </div>

          <!-- 已连接时的操作按钮 -->
          <div v-if="activeStatus?.connected" class="flex items-center gap-2 flex-shrink-0">
            <Button variant="outline" size="sm" class="h-8 text-xs rounded-full px-4 text-primary hover:bg-primary/10 hover:text-primary border-primary/20"
              @click="openDrawer(activePlatformData.id)">
              <Cog6ToothIcon class="size-3.5 mr-1.5" />
              {{ t('settings.network.editConfig') }}
            </Button>
            <Button variant="ghost" size="sm"
              class="h-8 text-xs rounded-full px-3 text-destructive hover:text-destructive hover:bg-destructive/10"
              @click="handleRevoke(activePlatformData.id)">
              {{ t('settings.network.disconnect') }}
            </Button>
          </div>
        </div>

        <!-- 已连接：用户信息 + 配置信息展示 -->
        <div v-if="activeStatus?.connected"
          class="px-6 pb-5 -mt-1">
          <div class="flex flex-wrap items-center gap-x-5 gap-y-2 px-4 py-3 bg-muted/40 rounded-lg">
            <!-- 用户头像 + 用户名（可点击跳转主页） -->
            <a v-if="activeStatus?.username"
              class="flex items-center gap-2 text-xs cursor-pointer"
              @click="openUserProfile(activePlatformData.id, activeStatus.username)">
              <img v-if="activeStatus?.avatarUrl" :src="activeStatus.avatarUrl"
                class="size-5 rounded-full flex-shrink-0" alt="" />
              <span class="font-semibold text-foreground hover:text-primary transition-colors">{{ activeStatus.username }}</span>
            </a>
            <!-- 分隔点 -->
            <span v-if="activeStatus?.username && activeConfigItems.length > 0" class="text-muted-foreground/30">·</span>
            <div v-for="item in activeConfigItems" :key="item.value"
              class="flex items-center gap-1.5 text-xs text-muted-foreground">
              <component :is="item.icon" class="size-3.5 flex-shrink-0 opacity-50" />
              <span class="truncate max-w-[200px]">{{ item.value }}</span>
            </div>
          </div>
        </div>

        <!-- 未连接：连接方式 -->
        <div v-if="!activeStatus?.connected" class="px-6 pb-5 -mt-1">
          <div class="border-t border-border/40 pt-4">
            <div class="flex items-center gap-3">
              <!-- OAuth 授权（主要入口，支持 OAuth 的平台始终显示） -->
              <template v-if="activePlatformData.hasOAuth">
                <Button v-if="!oauthLoading[activePlatformData.id]"
                  variant="default" size="sm" class="h-9 text-xs rounded-full px-5"
                  @click="handleOAuth(activePlatformData.id)">
                  <KeyIcon class="size-3.5 mr-1.5" />
                  {{ t('settings.network.connectViaOAuth') }}
                </Button>
                <Button v-else
                  variant="default" size="sm" class="h-9 text-xs rounded-full px-5" disabled>
                  <ArrowPathIcon class="size-3.5 animate-spin mr-1.5" />
                  {{ t('settings.network.waitingAuth') }}
                </Button>
              </template>
              <!-- 手动配置（备选方案，有 OAuth 时为 outline，无 OAuth 时为 default） -->
              <Button :variant="activePlatformData.hasOAuth ? 'outline' : 'default'"
                size="sm" class="h-9 text-xs rounded-full px-5"
                @click="openDrawer(activePlatformData.id)">
                <Cog6ToothIcon class="size-3.5 mr-1.5" />
                {{ t('settings.network.connectManual') }}
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ── 其他平台 ───────────────────────────────────────────── -->
    <div v-if="otherPlatforms.length > 0" class="space-y-4">
      <h2 class="text-sm text-primary font-medium border-l-[3px] border-primary pl-3 flex items-center h-4">
        {{ t('settings.network.otherPlatforms') }}
      </h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <div v-for="p in otherPlatforms" :key="p.id"
          class="group flex flex-col p-4 rounded-xl relative transition-all duration-200 bg-primary/2 border border-primary/20 hover:bg-primary/5 hover:shadow-xs hover:-translate-y-0.5">

          <!-- 顶部：图标 + 名称 + 状态 -->
          <div class="flex items-start gap-3 mb-2">
            <div class="size-9 rounded-lg flex items-center justify-center text-white flex-shrink-0 shadow-sm"
              :style="{ background: p.color }">
              <component :is="p.icon" class="size-4.5" />
            </div>
            <div class="flex-1 min-w-0 pt-0.5">
              <div class="flex items-center gap-2">
                <span class="text-sm font-semibold text-foreground leading-tight">{{ p.name }}</span>
                <span v-if="statuses[p.id]?.connected"
                  class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-full bg-green-500/10 text-green-600 dark:text-green-400 text-[10px] font-medium">
                  <span class="size-1.5 rounded-full bg-green-500 inline-block"></span>
                  {{ t('settings.network.connected') }}
                </span>
                <span v-else
                  class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-full bg-muted text-muted-foreground text-[10px] font-medium">
                  <span class="size-1.5 rounded-full bg-muted-foreground/40 inline-block"></span>
                  {{ t('settings.network.notConnected') }}
                </span>
              </div>
              <div class="text-xs text-muted-foreground mt-0.5 line-clamp-1">{{ p.description }}</div>
            </div>
          </div>

          <!-- 已连接：用户信息摘要 -->
          <div v-if="statuses[p.id]?.connected && statuses[p.id]?.username" class="mt-1 mb-3">
            <div class="flex flex-wrap items-center gap-x-3 gap-y-1.5 px-3 py-2 bg-muted/40 rounded-lg text-[11px] text-muted-foreground">
              <div class="flex items-center gap-1.5">
                <img v-if="statuses[p.id]?.avatarUrl" :src="statuses[p.id].avatarUrl"
                  class="size-4 rounded-full flex-shrink-0" alt="" />
                <span class="font-medium text-foreground/70">{{ statuses[p.id].username }}</span>
              </div>
              <template v-for="item in getCardItems(p.id)" :key="item.value">
                <div class="flex items-center gap-1">
                  <component :is="item.icon" class="size-3 opacity-50" />
                  <span>{{ item.value }}</span>
                </div>
              </template>
            </div>
          </div>

          <!-- 底部操作 -->
          <div class="flex items-center justify-between mt-auto pt-2.5 border-t border-border/50">
            <button
              class="p-1.5 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-md transition-colors cursor-pointer"
              :title="t('settings.network.configure')"
              @click.stop="openDrawer(p.id)">
              <Cog6ToothIcon class="size-3.5" />
            </button>
            <Button
              size="sm" variant="secondary"
              class="h-7 text-[10px] rounded-full px-3 bg-primary/5 border border-primary/10 text-primary hover:bg-primary hover:text-white transition-colors cursor-pointer"
              @click.stop="setActive(p.id)">
              {{ t('settings.network.setAsActive') }}
            </Button>
          </div>
        </div>
      </div>
    </div>

    </div>

    <!-- ── 断开连接确认弹窗 ──────────────────────────────────── -->
    <DeleteConfirmDialog
      v-model:open="revokeDialogOpen"
      :title="t('settings.network.disconnect')"
      :content="t('settings.network.revokeConfirm')"
      :confirm-text="t('settings.network.disconnect')"
      @confirm="confirmRevoke" />

    <!-- ── 手动配置抽屉 ───────────────────────────────────────── -->
    <Transition name="drawer">
      <div v-if="drawerOpen" class="fixed inset-0 z-50 flex justify-end" @click.self="closeDrawer">
        <div class="absolute inset-0 bg-black/30 backdrop-blur-[2px]" @click="closeDrawer" />
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

            <!-- 已连接用户信息 -->
            <div v-if="statuses[drawerPlatform]?.connected && statuses[drawerPlatform]?.username"
              class="flex items-center gap-3 px-4 py-3 bg-green-500/5 border border-green-500/15 rounded-lg">
              <img v-if="statuses[drawerPlatform]?.avatarUrl"
                :src="statuses[drawerPlatform].avatarUrl"
                class="size-9 rounded-full flex-shrink-0" alt="" />
              <div class="size-9 rounded-full bg-green-500/10 flex items-center justify-center flex-shrink-0 text-green-600" v-else>
                <UserIcon class="size-4.5" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="text-sm font-semibold text-foreground">{{ statuses[drawerPlatform].username }}</div>
                <div class="text-[11px] text-green-600 dark:text-green-400">
                  {{ statuses[drawerPlatform].connectedVia === 'oauth' ? 'OAuth' : t('settings.network.manualConfigTitle') }}
                  · {{ t('settings.network.connected') }}
                </div>
              </div>
            </div>

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
                <Input v-model="drawerForm.cname" placeholder="mydomain.com" />
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
                <Input v-model="drawerForm.cname" placeholder="mydomain.com" />
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

            <!-- ─ 代理设置（所有平台通用） ─ -->
            <div class="mt-2 pt-4 border-t border-border/50">
              <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-3">
                {{ t('settings.network.proxy') }}
              </div>
              <div class="space-y-3">
                <div class="flex items-center gap-3">
                  <Switch :checked="drawerForm.proxyEnabled" @update:checked="drawerForm.proxyEnabled = $event" size="sm" />
                  <span class="text-xs text-muted-foreground">{{ t('settings.network.proxyEnabled') }}</span>
                </div>
                <template v-if="drawerForm.proxyEnabled">
                  <FormField :label="t('settings.network.proxyURL')">
                    <Input v-model="drawerForm.proxyURL" placeholder="http://127.0.0.1:7890" />
                    <template #hint>{{ t('settings.network.proxyURLDesc') }}</template>
                  </FormField>
                </template>
              </div>
            </div>

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
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { toast } from '@/helpers/toast'
import { FolderOpenIcon, Cog6ToothIcon, ArrowPathIcon, XMarkIcon, InformationCircleIcon, KeyIcon, GlobeAltIcon, CodeBracketIcon, ServerStackIcon, UserIcon, LinkIcon } from '@heroicons/vue/24/outline'
import DeleteConfirmDialog from '@/components/ConfirmDialog/DeleteConfirmDialog.vue'
import { Switch } from '@/components/ui/switch'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { EventsEmit, EventsOn, BrowserOpenURL } from '@/wailsjs/runtime'
import { SaveSettingFromFrontend, RemoteDetectFromFrontend } from '@/wailsjs/go/facade/SettingFacade'
import { GetAllStatuses, StartOAuthFlow, RevokeToken, HasCredential } from '@/wailsjs/go/facade/OAuthFacade'
import { OpenKeyFileDialog } from '@/wailsjs/go/app/App'
import { domain } from '@/wailsjs/go/models'
import type { service } from '@/wailsjs/go/models'

const { t } = useI18n()
const siteStore = useSiteStore()

// ── 平台定义 ──────────────────────────────────────────────────────────────

const GitHubIcon = { template: `<svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"/></svg>` }
const VercelIcon = { template: `<svg viewBox="0 0 512 512" fill="currentColor"><path d="M256 48L496 464H16L256 48z"/></svg>` }
const NetlifyIcon = { template: `<svg viewBox="0 0 256 256" fill="currentColor"><path d="M134.4 4.8L255.2 125.6l-50.4 12-32.8-32.8-3.2 0-14 14 26.8 26.8-16.4 4-17.6-17.6-14 14L160 172.4l-16 3.6-10.8-10.8-14 14 4.4 4.4-16.4 4L76 156.4 4.8 134.4 0 128 128 0l6.4 4.8zM90 170.4l14-14L77.6 130l-14 14L90 170.4zm28-28l14-14-26.4-26.4-14 14L118 142.4zm28-28l14-14-26.4-26.4-14 14L146 114.4z"/></svg>` }
const GiteeIcon = { template: `<svg viewBox="0 0 24 24" fill="currentColor"><path d="M11.984 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 12 0a12 12 0 0 0-.016 0zm6.09 5.333c.328 0 .593.266.592.593v1.482a.594.594 0 0 1-.593.592H9.777c-.982 0-1.778.796-1.778 1.778v5.926c0 .327.266.592.593.592h5.926c.982 0 1.778-.796 1.778-1.778v-.296a.593.593 0 0 0-.592-.593h-4.15a.592.592 0 0 1-.592-.592v-1.482a.593.593 0 0 1 .593-.592h6.814c.328 0 .593.265.593.592v3.408a4 4 0 0 1-4 4H6.518a.593.593 0 0 1-.593-.593V8.333a4 4 0 0 1 4-3H18.074z"/></svg>` }
const CodingIcon = { template: `<svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm4.5 16.5h-9a1.5 1.5 0 0 1 0-3h9a1.5 1.5 0 0 1 0 3zm0-4.5h-9a1.5 1.5 0 0 1 0-3h9a1.5 1.5 0 0 1 0 3z"/></svg>` }
const ServerIcon = { template: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="5" rx="1"/><rect x="2" y="10" width="20" height="5" rx="1"/><rect x="2" y="17" width="20" height="5" rx="1"/><circle cx="6" cy="5.5" r=".8" fill="currentColor"/><circle cx="6" cy="12.5" r=".8" fill="currentColor"/></svg>` }
const BranchIcon = { template: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="6" y1="3" x2="6" y2="15"/><circle cx="18" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><path d="M18 9a9 9 0 0 1-9 9"/></svg>` }

// hasOAuth: 该平台是否支持 OAuth（始终显示授权按钮，不依赖后端 client ID 是否配置）
const platforms = [
  { id: 'github',  name: 'GitHub Pages',  color: '#24292f', icon: GitHubIcon,   hasOAuth: true,  profileUrl: 'https://github.com/',  description: t('settings.network.githubDesc') },
  { id: 'netlify', name: 'Netlify',       color: '#00c7b7', icon: NetlifyIcon,  hasOAuth: true,  profileUrl: '',                       description: t('settings.network.netlifyDesc') },
  { id: 'vercel',  name: 'Vercel',        color: '#000000', icon: VercelIcon,   hasOAuth: false, profileUrl: '',                       description: t('settings.network.vercelDesc') },
  { id: 'gitee',   name: 'Gitee Pages',   color: '#c71d23', icon: GiteeIcon,    hasOAuth: true,  profileUrl: 'https://gitee.com/',     description: t('settings.network.giteeDesc') },
  { id: 'coding',  name: 'Coding Pages',  color: '#0066ff', icon: CodingIcon,   hasOAuth: false, profileUrl: '',                       description: t('settings.network.codingDesc') },
  { id: 'sftp',    name: 'SFTP / FTP',    color: '#5856d6', icon: ServerIcon,    hasOAuth: false, profileUrl: '',                       description: t('settings.network.sftpDesc') },
]

// ── 状态 ──────────────────────────────────────────────────────────────────

const statuses = ref<Record<string, service.PlatformStatus>>({})
const oauthLoading = ref<Record<string, boolean>>({})
const activePlatform = ref('github')
const detectLoading = ref(false)
const saveLoading = ref(false)
const revokeDialogOpen = ref(false)
const revokePlatformId = ref('')

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
  proxyEnabled: false,
  proxyURL: '',
})

// ── 计算属性 ──────────────────────────────────────────────────────────────

const currentPlatform = computed(() => platforms.find(p => p.id === drawerPlatform.value))
const activePlatformData = computed(() => platforms.find(p => p.id === activePlatform.value))
const otherPlatforms = computed(() => platforms.filter(p => p.id !== activePlatform.value))
const activeStatus = computed(() => statuses.value[activePlatform.value])

// 当前平台的配置信息条目（用于展示）
const activeConfigItems = computed(() => {
  const pid = activePlatform.value
  const cfg = (siteStore.site.setting.platformConfigs || {})[pid] || {}
  const items: { icon: any; value: string }[] = []

  // 域名：有 CNAME 显示 CNAME，否则显示 domain
  if (['github', 'gitee', 'coding'].includes(pid)) {
    const displayDomain = cfg.cname || (cfg.domain ? String(cfg.domain).replace(/^https?:\/\//, '') : '')
    if (displayDomain) items.push({ icon: GlobeAltIcon, value: displayDomain })
    if (cfg.repository) items.push({ icon: CodeBracketIcon, value: cfg.repository })
    if (cfg.branch) items.push({ icon: BranchIcon, value: cfg.branch })
  } else if (pid === 'netlify') {
    const d = cfg.domain ? String(cfg.domain).replace(/^https?:\/\//, '') : ''
    if (d) items.push({ icon: GlobeAltIcon, value: d })
    if (cfg.netlifySiteId) items.push({ icon: CodeBracketIcon, value: cfg.netlifySiteId })
  } else if (pid === 'vercel') {
    const displayDomain = cfg.cname || (cfg.domain ? String(cfg.domain).replace(/^https?:\/\//, '') : '')
    if (displayDomain) items.push({ icon: GlobeAltIcon, value: displayDomain })
    if (cfg.repository) items.push({ icon: CodeBracketIcon, value: cfg.repository })
  } else if (pid === 'sftp') {
    const d = cfg.domain ? String(cfg.domain).replace(/^https?:\/\//, '') : ''
    if (d) items.push({ icon: GlobeAltIcon, value: d })
    const addr = [cfg.server, cfg.port].filter(Boolean).join(':')
    if (addr) items.push({ icon: ServerStackIcon, value: addr })
  }

  return items
})

// ── 生命周期 ──────────────────────────────────────────────────────────────

onMounted(async () => {
  const setting = siteStore.site.setting
  activePlatform.value = setting.platform || 'github'

  await loadStatuses()

  // 监听 OAuth 授权结果
  EventsOn('oauth:success', async (data: any) => {
    const { provider, username, avatarUrl, email } = data
    oauthLoading.value[provider] = false
    statuses.value[provider] = {
      connected: true,
      connectedVia: 'oauth',
      username,
      avatarUrl,
      email: email || '',
    }
    toast.success(`${getPlatformName(provider)} ${t('settings.network.authSuccess')}`)

    // 自动填充默认配置并保存
    await autoFillAfterOAuth(provider, username, email || '')
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
  } catch (e: any) {
    oauthLoading.value[platformId] = false
    // Wails 返回的错误是字符串，不是 Error 对象
    const msg = typeof e === 'string' ? e : (e?.message || t('settings.network.authFailed'))
    toast.error(msg)
  }
}

function handleRevoke(platformId: string) {
  revokePlatformId.value = platformId
  revokeDialogOpen.value = true
}

async function confirmRevoke() {
  const platformId = revokePlatformId.value
  revokeDialogOpen.value = false
  try {
    await RevokeToken(platformId)
    await loadStatuses()
    EventsEmit('app-site-reload')
    toast.success(`${getPlatformName(platformId)} ${t('settings.network.disconnected')}`)
  } catch (e: any) {
    const msg = typeof e === 'string' ? e : (e?.message || t('settings.network.disconnectFailed'))
    toast.error(msg)
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
      proxyEnabled: setting.proxyEnabled || false,
      proxyURL: setting.proxyURL || '',
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

  Object.keys(drawerForm).forEach(k => { drawerForm[k] = '' })
  drawerForm.transferProtocol = 'sftp'
  drawerForm.port = platformId === 'sftp' ? '22' : ''

  for (const [k, v] of Object.entries(cfg)) {
    if (k === 'domain') {
      const d = String(v || '')
      const idx = d.indexOf('://')
      drawerForm.domain = idx !== -1 ? d.substring(idx + 3) : d
    } else {
      drawerForm[k] = v || ''
    }
  }

  if (platformId === 'sftp') {
    sftpAuthType.value = drawerForm.privateKey ? 'key' : 'password'
  }

  const setting = siteStore.site.setting
  drawerForm.proxyEnabled = setting.proxyEnabled || false
  drawerForm.proxyURL = setting.proxyURL || ''
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
    await loadStatuses()
    toast.success(t('settings.network.credentialSaved'))
    closeDrawer()
  } catch (e: any) {
    toast.error(e?.message || t('settings.network.saveFailed'))
  } finally {
    saveLoading.value = false
  }
}

// OAuth 成功后自动填充默认配置
async function autoFillAfterOAuth(platformId: string, username: string, email: string) {
  const existingConfigs = JSON.parse(JSON.stringify(siteStore.site.setting.platformConfigs || {}))
  const cfg = existingConfigs[platformId] || {}

  const lowerUsername = username.toLowerCase()

  if (platformId === 'github') {
    cfg.repository = `${lowerUsername}.github.io`
    cfg.branch = cfg.branch || 'main'
    cfg.username = username
    cfg.email = email || cfg.email || ''
    cfg.domain = `https://${lowerUsername}.github.io`
  } else if (platformId === 'gitee') {
    cfg.repository = lowerUsername
    cfg.branch = cfg.branch || 'master'
    cfg.username = username
    cfg.email = email || cfg.email || ''
    cfg.domain = `https://${lowerUsername}.gitee.io`
  } else if (platformId === 'netlify') {
    // Netlify 无法自动推断 site id，仅填充已知信息
  }

  existingConfigs[platformId] = cfg

  try {
    const settingObj = new domain.Setting({
      platform: activePlatform.value,
      platformConfigs: existingConfigs,
      proxyEnabled: siteStore.site.setting.proxyEnabled || false,
      proxyURL: siteStore.site.setting.proxyURL || '',
    })
    await SaveSettingFromFrontend(settingObj)
    EventsEmit('app-site-reload')
  } catch (e) {
    console.error('自动填充配置失败', e)
  }
}

// 获取平台配置摘要（用于其他平台卡片）
function getConfigSummary(platformId: string): string {
  const cfg = (siteStore.site.setting.platformConfigs || {})[platformId] || {}
  const parts: string[] = []
  if (cfg.domain) {
    parts.push(String(cfg.domain).replace(/^https?:\/\//, ''))
  }
  if (['github', 'gitee', 'coding'].includes(platformId)) {
    if (cfg.repository) parts.push(cfg.repository)
    if (cfg.branch) parts.push(cfg.branch)
  } else if (platformId === 'netlify' && cfg.netlifySiteId) {
    parts.push(cfg.netlifySiteId)
  } else if (platformId === 'vercel' && cfg.repository) {
    parts.push(cfg.repository)
  } else if (platformId === 'sftp' && cfg.server) {
    parts.push(`${cfg.server}${cfg.port ? ':' + cfg.port : ''}`)
  }
  return parts.join(' · ')
}

// 其他平台卡片的配置条目（带 icon）
function getCardItems(platformId: string): { icon: any; value: string }[] {
  const cfg = (siteStore.site.setting.platformConfigs || {})[platformId] || {}
  const items: { icon: any; value: string }[] = []
  if (['github', 'gitee', 'coding'].includes(platformId)) {
    if (cfg.repository) items.push({ icon: CodeBracketIcon, value: cfg.repository })
    if (cfg.branch) items.push({ icon: BranchIcon, value: cfg.branch })
  } else if (platformId === 'netlify' && cfg.netlifySiteId) {
    items.push({ icon: CodeBracketIcon, value: cfg.netlifySiteId })
  } else if (platformId === 'vercel' && cfg.repository) {
    items.push({ icon: CodeBracketIcon, value: cfg.repository })
  } else if (platformId === 'sftp' && cfg.server) {
    items.push({ icon: ServerStackIcon, value: `${cfg.server}${cfg.port ? ':' + cfg.port : ''}` })
  }
  return items
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

  if (platformId === 'sftp') {
    if (sftpAuthType.value === 'password') cfg.privateKey = ''
    else cfg.password = ''
  }

  existingConfigs[platformId] = cfg

  return {
    platform: activePlatform.value,
    platformConfigs: existingConfigs,
    proxyEnabled: drawerForm.proxyEnabled || false,
    proxyURL: drawerForm.proxyURL || '',
  }
}

function getPlatformName(id: string) {
  return platforms.find(p => p.id === id)?.name || id
}

function openUserProfile(platformId: string, username: string) {
  const platform = platforms.find(p => p.id === platformId)
  if (platform?.profileUrl && username) {
    BrowserOpenURL(platform.profileUrl + username)
  }
}
</script>

<!-- 小工具组件（行内定义，避免额外文件） -->
<script lang="ts">
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
