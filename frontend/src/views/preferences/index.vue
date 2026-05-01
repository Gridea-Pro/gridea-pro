<template>
  <div class="flex h-full min-h-[400px] w-full max-w-full bg-background rounded-xl overflow-hidden text-foreground">
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
    <div class="flex-1 min-w-0 max-w-full p-8 overflow-y-auto overflow-x-hidden">
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

      <!-- 站点管理 -->
      <div v-if="activeTab === 'sites'" class="animate-fade-in flex flex-col min-w-0">
        <div class="flex justify-between items-center mb-6 flex-shrink-0">
          <h2 class="text-xl font-semibold text-foreground">{{ t('settings.sites.title') }}</h2>
          <Button variant="outline" size="sm" class="h-8 px-3 text-xs rounded-full" @click="handleAddSite">
            <PlusIcon class="size-4 mr-1" />
            {{ t('settings.sites.add') }}
          </Button>
        </div>

        <draggable v-model="sites" handle=".handle" item-key="path" class="w-full min-w-0 flex-1 overflow-y-auto overflow-x-hidden" @change="handleSiteSort">
          <template #item="{ element: site }">
            <div
              class="group flex w-full min-w-0 max-w-full items-center gap-3 px-4 py-3 mb-2 rounded-lg border transition-all duration-200 overflow-hidden"
              :class="site.active
                ? 'bg-primary/5 border-primary/30'
                : 'bg-card/50 border-border/50 hover:bg-primary/2 hover:border-primary/20'">
              <!-- 拖拽手柄 -->
              <div class="handle flex-shrink-0 cursor-move text-muted-foreground/40 hover:text-muted-foreground">
                <Bars3Icon class="size-3.5" />
              </div>

              <!-- 站点信息 -->
              <div class="flex-1 min-w-0 overflow-hidden">
                <div class="text-sm font-medium text-foreground truncate block w-full">{{ site.name }}</div>
                <div class="text-xs text-muted-foreground/60 truncate mt-0.5 block w-full">{{ site.path }}</div>
              </div>

              <!-- 编辑按钮 -->
              <button
                class="flex-shrink-0 text-muted-foreground/30 hover:text-foreground transition-colors cursor-pointer opacity-0 group-hover:opacity-100"
                @click="handleEditSite(site)">
                <PencilIcon class="size-3.5" />
              </button>

              <!-- 删除按钮 -->
              <button
                class="flex-shrink-0 text-muted-foreground/30 hover:text-destructive transition-colors cursor-pointer opacity-0 group-hover:opacity-100"
                @click="handleDeleteSite(site)">
                <TrashIcon class="size-3.5" />
              </button>

              <!-- Switch 开关 -->
              <div class="flex-shrink-0">
                <Switch
                  :checked="site.active" size="sm"
                  @update:checked="(v: boolean) => handleSwitchSite(site, v)" />
              </div>
            </div>
          </template>
        </draggable>

        <!-- 空状态 -->
        <div v-if="sites.length === 0" class="text-center py-12 text-muted-foreground">
          <div class="text-sm">{{ t('settings.sites.empty') }}</div>
        </div>
      </div>

      <!-- AI 配置 -->
      <div v-if="activeTab === 'ai'" class="animate-fade-in">
        <h2 class="text-xl font-semibold mb-6 text-foreground">{{ t('settings.ai.title') }}</h2>

        <RadioGroup v-model="aiForm.mode" class="space-y-4">
          <!-- 内置模型 -->
          <label
            class="flex items-start gap-3 p-4 rounded-lg border border-border cursor-pointer hover:border-primary/40 transition-colors"
            :class="{ 'border-primary/60 bg-primary/5': aiForm.mode === 'builtin' }">
            <RadioGroupItem value="builtin" class="mt-0.5" />
            <div class="flex-1">
              <div class="text-sm font-medium text-foreground">
                {{ t('settings.ai.modeBuiltIn') }}
                <span class="ml-2 text-xs text-muted-foreground">{{ t('settings.ai.modeBuiltInBadge') }}</span>
              </div>
              <div class="text-xs text-muted-foreground mt-1">
                {{ t('settings.ai.modeBuiltInDesc') }}
              </div>
              <div v-if="builtInModels.length > 0" class="text-xs text-muted-foreground mt-1.5">
                {{ t('settings.ai.builtInModelsLabel') }}
                <code class="px-1 py-0.5 mx-0.5 rounded bg-muted text-foreground/80" v-for="m in builtInModels"
                  :key="m">{{ m }}</code>
              </div>
              <div class="text-xs text-muted-foreground mt-1.5">
                {{ t('settings.ai.builtInLimit', { daily: 20, minute: 5 }) }}
              </div>
            </div>
          </label>

          <!-- 自定义模型 -->
          <label
            class="flex items-start gap-3 p-4 rounded-lg border border-border cursor-pointer hover:border-primary/40 transition-colors"
            :class="{ 'border-primary/60 bg-primary/5': aiForm.mode === 'custom' }">
            <RadioGroupItem value="custom" class="mt-0.5" />
            <div class="flex-1">
              <div class="text-sm font-medium text-foreground">{{ t('settings.ai.modeCustom') }}</div>
              <div class="text-xs text-muted-foreground mt-1">{{ t('settings.ai.modeCustomDesc') }}</div>

              <!-- 自定义模型表单 -->
              <div v-if="aiForm.mode === 'custom'" class="mt-4 space-y-4" @click.stop>
                <!-- 厂商 -->
                <div class="flex items-center gap-3">
                  <Label class="w-20 text-xs shrink-0">{{ t('settings.ai.provider') }}</Label>
                  <Select :model-value="aiForm.activeProvider" @update:model-value="onProviderChange">
                    <SelectTrigger class="flex-1">
                      <SelectValue :placeholder="t('settings.ai.selectProvider')" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem v-for="p in providerRegistry" :key="p.id" :value="p.id">{{ p.name }}</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <!-- Base URL -->
                <div v-if="aiForm.activeProvider && isCustomOpenAIProvider" class="flex items-start gap-3">
                  <Label class="w-20 text-xs shrink-0 pt-2">{{ t('settings.ai.baseURL') }}</Label>
                  <div class="flex-1 space-y-1.5">
                    <Input v-model="currentBaseURL" placeholder="https://api.example.com/v1" />
                    <div class="text-xs text-muted-foreground">{{ t('settings.ai.baseURLDesc') }}</div>
                  </div>
                </div>

                <!-- 模型 -->
                <div v-if="aiForm.activeProvider" class="flex items-start gap-3">
                  <Label class="w-20 text-xs shrink-0 pt-2">{{ t('settings.ai.model') }}</Label>
                  <div class="flex-1 space-y-2">
                    <div class="flex items-center gap-2">
                      <Input v-if="isCustomOpenAIProvider" v-model="currentModel" class="flex-1"
                        :placeholder="t('settings.ai.customModelPlaceholder')" />
                      <Select v-else-if="!useCustomModelId" v-model="currentModel">
                        <SelectTrigger class="flex-1">
                          <SelectValue :placeholder="t('settings.ai.selectModel')" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem v-for="m in currentModelOptions" :key="m" :value="m">{{ m }}</SelectItem>
                        </SelectContent>
                      </Select>
                      <Input v-else v-model="customModelInput" class="flex-1"
                        :placeholder="t('settings.ai.customModelPlaceholder')"
                        @input="handleCustomModelInputChange" />
                      <Button variant="outline" size="icon" type="button" class="size-9 shrink-0"
                        :disabled="refreshingModels" :title="t('settings.ai.refreshModels')"
                        @click="handleRefreshModels">
                        <ArrowPathIcon class="size-4" :class="{ 'animate-spin': refreshingModels }" />
                      </Button>
                    </div>
                    <label v-if="!isCustomOpenAIProvider" class="flex items-center gap-1.5 text-xs text-muted-foreground cursor-pointer w-fit">
                      <Checkbox :model-value="useCustomModelId" class="size-3"
                        @update:model-value="(v: any) => onUseCustomModelToggle(!!v)">
                        <Check class="size-2.5" />
                      </Checkbox>
                      {{ t('settings.ai.useCustomModelId') }}
                    </label>
                  </div>
                </div>

                <!-- API Key -->
                <div v-if="aiForm.activeProvider" class="flex items-start gap-3">
                  <Label class="w-20 text-xs shrink-0 pt-2">{{ t('settings.ai.apiKey') }}</Label>
                  <div class="flex-1 space-y-1.5">
                    <Input v-model="currentApiKey" type="password" placeholder="sk-..." />
                    <div v-if="currentProviderInfo?.apiKeyURL" class="text-xs">
                      <a class="inline-flex items-center gap-1 text-primary hover:underline cursor-pointer"
                        @click.prevent="openApiKeyURL(currentProviderInfo!.apiKeyURL)">
                        {{ t('settings.ai.getApiKey') }}
                        <ArrowTopRightOnSquareIcon class="size-3" />
                      </a>
                    </div>
                  </div>
                </div>

                <!-- 测试连接 -->
                <div v-if="aiForm.activeProvider" class="flex items-center gap-3">
                  <Label class="w-20 text-xs shrink-0">&nbsp;</Label>
                  <Button variant="outline" type="button"
                    class="h-8 px-4 text-xs rounded-full border-primary/30 text-primary hover:bg-primary/5 cursor-pointer"
                    :disabled="testingConnection" @click="handleTestConnection">
                    {{ testingConnection ? t('settings.ai.testing') : t('settings.ai.testConnection') }}
                  </Button>
                </div>
              </div>
            </div>
          </label>
        </RadioGroup>

        <div class="flex justify-end mt-6">
          <Button variant="default"
            class="h-8 px-5 text-xs rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
            @click="saveAISetting">
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

    <!-- 添加/编辑站点对话框 -->
    <Dialog v-model:open="showAddDialog">
      <DialogContent class="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>{{ editingSite ? t('settings.sites.edit') : t('settings.sites.add') }}</DialogTitle>
        </DialogHeader>
        <div class="space-y-4 py-4">
          <div>
            <Label class="mb-1.5 block text-sm">{{ t('settings.sites.name') }}</Label>
            <Input v-model="newSiteName" :placeholder="t('settings.sites.namePlaceholder')" />
          </div>
          <div>
            <Label class="mb-1.5 block text-sm">{{ t('settings.sites.path') }}</Label>
            <div class="flex gap-2">
              <Input v-model="newSitePath" class="flex-1" readonly :placeholder="t('settings.sites.selectPath')" />
              <Button variant="outline" size="icon" @click="selectNewSitePath">
                <FolderOpenIcon class="size-5" />
              </Button>
            </div>
          </div>
        </div>
        <DialogFooter class="gap-3">
          <Button
variant="outline"
            class="w-18 h-8 text-xs justify-center rounded-full border border-primary/20 text-primary/80 hover:bg-primary/5 hover:text-primary cursor-pointer"
            @click="showAddDialog = false">{{ t('common.cancel') }}</Button>
          <Button
variant="default"
            class="w-18 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
            :disabled="!newSiteName || !newSitePath" @click="confirmAddSite">{{ t('common.save') }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- 删除确认对话框 -->
    <DeleteConfirmDialog v-model:open="showDeleteDialog" :confirm-text="t('common.delete')" @confirm="confirmDeleteSite">
      <template #description>
        {{ t('settings.sites.confirmDelete') }}
      </template>
    </DeleteConfirmDialog>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { toast } from '@/helpers/toast'
import { GetVersion } from '@/wailsjs/go/app/App'
import Draggable from 'vuedraggable'
import AppSetting from './components/AppSetting.vue'
import DeleteConfirmDialog from '@/components/ConfirmDialog/DeleteConfirmDialog.vue'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Checkbox } from '@/components/ui/checkbox'
import { Check } from 'lucide-vue-next'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import {
  SwatchIcon,
  LanguageIcon,
  GlobeAltIcon,
  InformationCircleIcon,
  FolderOpenIcon,
  PlusIcon,
  Bars3Icon,
  TrashIcon,
  PencilIcon,
  SparklesIcon,
  ArrowPathIcon,
  ArrowTopRightOnSquareIcon,
} from '@heroicons/vue/24/outline'
import { EventsEmit, BrowserOpenURL } from '@/wailsjs/runtime'
import { OpenFolderDialog, GetSites, AddSite, RemoveSite, UpdateSites, SwitchSite } from '@/wailsjs/go/app/App'
import {
  GetAISetting,
  SaveAISettingFromFrontend,
  GetProviderRegistry,
  GetBuiltInModels,
  ListProviderModels,
  ListProviderModelsWithBaseURL,
  TestConnection,
  TestConnectionWithBaseURL,
} from '@/wailsjs/go/facade/AIFacade'
import { domain, ai as aiNS } from '@/wailsjs/go/models'
import { setI18nLanguage, type LocaleType } from '@/locales'

interface SiteEntry {
  name: string
  path: string
  active: boolean
}

const { t, locale } = useI18n()
const siteStore = useSiteStore()

const activeTab = ref('appearance')
const currentLanguage = ref<LocaleType>(locale.value as LocaleType)
const version = ref('')
GetVersion().then((v) => { version.value = v })

// 站点管理
const sites = ref<SiteEntry[]>([])
const showAddDialog = ref(false)
const newSiteName = ref('')
const newSitePath = ref('')
const showDeleteDialog = ref(false)
const siteToDelete = ref<SiteEntry | null>(null)
const editingSite = ref<SiteEntry | null>(null)
const switching = ref(false)

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
  { key: 'sites', icon: GlobeAltIcon, label: t('settings.sites.title') },
  { key: 'ai', icon: SparklesIcon, label: t('settings.ai.title') },
  { key: 'about', icon: InformationCircleIcon, label: t('nav.about') },
])

// ─── AI 配置 ───────────────────────────────────────────
const aiForm = ref<domain.AISetting>(
  new domain.AISetting({
    mode: 'builtin',
    activeProvider: '',
    customs: {},
  }),
)
const builtInModels = ref<string[]>([])
const providerRegistry = ref<aiNS.ProviderInfo[]>([])
const currentProviderInfo = computed<aiNS.ProviderInfo | undefined>(() =>
  providerRegistry.value.find((p) => p.id === aiForm.value.activeProvider),
)
const isCustomOpenAIProvider = computed(() => aiForm.value.activeProvider === 'custom-openai')
// 当前选中厂商的模型列表（默认或刷新后），按厂商隔离
const modelOptionsByProvider = ref<Record<string, string[]>>({})
const currentModelOptions = computed<string[]>(() => {
  const provider = aiForm.value.activeProvider
  if (!provider) return []
  return modelOptionsByProvider.value[provider] || currentProviderInfo.value?.defaultModels || []
})
const useCustomModelId = ref(false) // 是否手动输入模型 ID（UI 状态，不持久化）
const customModelInput = ref('')
const refreshingModels = ref(false)
const testingConnection = ref(false)

// 确保 customs[provider] 存在
const ensureCustomEntry = (provider: string) => {
  if (!provider) return
  if (!aiForm.value.customs) aiForm.value.customs = {}
  if (!aiForm.value.customs[provider]) {
    aiForm.value.customs[provider] = { model: '', apiKey: '', baseURL: '' } as any
  }
}

// 当前激活厂商的 model
const currentModel = computed<string>({
  get: () => aiForm.value.customs?.[aiForm.value.activeProvider]?.model || '',
  set: (val) => {
    const provider = aiForm.value.activeProvider
    if (!provider) return
    ensureCustomEntry(provider)
    aiForm.value.customs[provider].model = val
  },
})

// 当前激活厂商的 apiKey
const currentApiKey = computed<string>({
  get: () => aiForm.value.customs?.[aiForm.value.activeProvider]?.apiKey || '',
  set: (val) => {
    const provider = aiForm.value.activeProvider
    if (!provider) return
    ensureCustomEntry(provider)
    aiForm.value.customs[provider].apiKey = val
  },
})

// 当前激活厂商的 Base URL（用于自定义 OpenAI 兼容 / 中转站）
const currentBaseURL = computed<string>({
  get: () => (aiForm.value.customs?.[aiForm.value.activeProvider] as any)?.baseURL || '',
  set: (val) => {
    const provider = aiForm.value.activeProvider
    if (!provider) return
    ensureCustomEntry(provider)
    ;(aiForm.value.customs[provider] as any).baseURL = val.trim()
  },
})

const loadAISetting = async () => {
  try {
    const [setting, models, registry] = await Promise.all([
      GetAISetting(),
      GetBuiltInModels(),
      GetProviderRegistry(),
    ])
    builtInModels.value = models || []
    providerRegistry.value = registry || []
    aiForm.value = new domain.AISetting({
      mode: setting.mode || 'builtin',
      activeProvider: setting.activeProvider || '',
      customs: setting.customs || {},
    })
    // 同步当前厂商的 UI 状态
    refreshUIStateForCurrentProvider()
  } catch (e) {
    console.error('Failed to load AI setting:', e)
  }
}

// 切换厂商或加载后，根据 currentModel 决定 useCustomModelId 状态
const refreshUIStateForCurrentProvider = () => {
  const model = currentModel.value
  if (isCustomOpenAIProvider.value) {
    useCustomModelId.value = true
    customModelInput.value = model
    return
  }
  const defaults = currentProviderInfo.value?.defaultModels || []
  const refreshed = modelOptionsByProvider.value[aiForm.value.activeProvider] || []
  const knownModels = refreshed.length > 0 ? refreshed : defaults
  if (model && knownModels.length > 0 && !knownModels.includes(model)) {
    useCustomModelId.value = true
    customModelInput.value = model
  } else {
    useCustomModelId.value = false
    customModelInput.value = ''
  }
}

const onProviderChange = (newProvider: any) => {
  aiForm.value.activeProvider = String(newProvider || '')
  // 不清除任何已存在的 customs 数据，只切换显示的对象
  ensureCustomEntry(aiForm.value.activeProvider)
  refreshUIStateForCurrentProvider()
}

const onUseCustomModelToggle = (val: boolean) => {
  useCustomModelId.value = val
  if (val) {
    customModelInput.value = currentModel.value
  } else {
    // 切回下拉时，如果当前 model 不在已知列表里就清空
    const known = currentModelOptions.value
    if (!known.includes(currentModel.value)) {
      currentModel.value = ''
    }
  }
}

const handleCustomModelInputChange = () => {
  currentModel.value = customModelInput.value.trim()
}

const handleRefreshModels = async () => {
  const provider = aiForm.value.activeProvider
  if (!provider) {
    toast.warning(t('settings.ai.selectProviderFirst'))
    return
  }
  if (!currentApiKey.value) {
    toast.warning(t('settings.ai.fillApiKeyFirst'))
    return
  }
  if (isCustomOpenAIProvider.value && !currentBaseURL.value) {
    toast.warning(t('settings.ai.fillBaseURLFirst'))
    return
  }
  refreshingModels.value = true
  try {
    const models = isCustomOpenAIProvider.value
      ? await ListProviderModelsWithBaseURL(provider, currentApiKey.value, currentBaseURL.value)
      : await ListProviderModels(provider, currentApiKey.value)
    if (models && models.length > 0) {
      modelOptionsByProvider.value[provider] = models
      toast.success(t('settings.ai.refreshSuccess'))
    } else {
      toast.warning(t('settings.ai.refreshEmpty'))
    }
  } catch (e: any) {
    console.error('Refresh models failed:', e)
    toast.error(t('settings.ai.refreshFailed'))
  } finally {
    refreshingModels.value = false
  }
}

const handleTestConnection = async () => {
  const provider = aiForm.value.activeProvider
  if (!provider || !currentModel.value || !currentApiKey.value || (isCustomOpenAIProvider.value && !currentBaseURL.value)) {
    toast.warning(t('settings.ai.testIncomplete'))
    return
  }
  testingConnection.value = true
  try {
    if (isCustomOpenAIProvider.value) {
      await TestConnectionWithBaseURL(provider, currentModel.value, currentApiKey.value, currentBaseURL.value)
    } else {
      await TestConnection(provider, currentModel.value, currentApiKey.value)
    }
    toast.success(t('settings.ai.testSuccess'))
  } catch (e: any) {
    console.error('Test connection failed:', e)
    toast.error(t('settings.ai.testFailed'))
  } finally {
    testingConnection.value = false
  }
}

const openApiKeyURL = (url: string) => {
  if (url) BrowserOpenURL(url)
}

const saveAISetting = async () => {
  try {
    await SaveAISettingFromFrontend(aiForm.value)
    toast.success(t('settings.ai.saveSuccess'))
  } catch (e: any) {
    toast.error(e.message || 'Failed to save AI setting')
  }
}

watch(locale, (val) => {
  currentLanguage.value = val as LocaleType
})

onMounted(async () => {
  currentLanguage.value = locale.value as LocaleType
  await loadSites()
  await loadAISetting()
})

const loadSites = async () => {
  try {
    const result = await GetSites()
    sites.value = result || []
  } catch (e) {
    console.error('Failed to load sites:', e)
  }
}

const saveLanguage = async (val: string) => {
  const newLocale = val as LocaleType
  await setI18nLanguage(newLocale)
  EventsEmit('app:change-locale', newLocale)
  toast.success(t('common.saved'))
}

// 站点管理
const handleAddSite = () => {
  editingSite.value = null
  newSiteName.value = ''
  newSitePath.value = ''
  showAddDialog.value = true
}

const handleEditSite = (site: SiteEntry) => {
  editingSite.value = site
  newSiteName.value = site.name
  newSitePath.value = site.path
  showAddDialog.value = true
}

const selectNewSitePath = async () => {
  const filePaths = await OpenFolderDialog()
  if (filePaths && filePaths.length > 0) {
    newSitePath.value = filePaths[0].replace(/\\/g, '/')
    // 如果名称为空，用文件夹名作为默认名称
    if (!newSiteName.value) {
      const parts = newSitePath.value.split('/')
      newSiteName.value = parts[parts.length - 1] || ''
    }
  }
}

const confirmAddSite = async () => {
  try {
    if (editingSite.value) {
      // 编辑模式：更新本地列表中的站点信息
      const idx = sites.value.findIndex(s => s.path === editingSite.value!.path)
      if (idx !== -1) {
        sites.value[idx].name = newSiteName.value
        sites.value[idx].path = newSitePath.value
      }
      await UpdateSites(sites.value)
      showAddDialog.value = false
      toast.success(t('common.saved'))
    } else {
      // 添加模式
      const result = await AddSite(newSiteName.value, newSitePath.value)
      sites.value = result || []
      showAddDialog.value = false
      toast.success(t('settings.sites.added'))
    }
  } catch (e: any) {
    toast.error(e.message || 'Failed to save site')
  }
}

const handleSwitchSite = async (site: SiteEntry, checked: boolean) => {
  if (!checked) {
    // 尝试关闭 — 检查是否是最后一个
    const activeCount = sites.value.filter(s => s.active).length
    if (activeCount <= 1) {
      toast.warning(t('settings.sites.cannotDisableLast'))
      return
    }
  }

  if (checked && switching.value) return
  if (!checked) return // 不允许手动关闭，只能通过打开另一个来关闭

  switching.value = true
  try {
    await SwitchSite(site.path)
    // 更新本地状态
    sites.value.forEach(s => {
      s.active = s.path === site.path
    })
    toast.success(t('settings.sites.switched'))
  } catch (e: any) {
    toast.error(e.message || 'Failed to switch site')
  } finally {
    switching.value = false
  }
}

const handleDeleteSite = (site: SiteEntry) => {
  if (site.active) {
    toast.warning(t('settings.sites.cannotDeleteActive'))
    return
  }
  siteToDelete.value = site
  showDeleteDialog.value = true
}

const confirmDeleteSite = async () => {
  if (!siteToDelete.value) return
  try {
    const result = await RemoveSite(siteToDelete.value.path)
    sites.value = result || []
    toast.success(t('settings.sites.deleted'))
  } catch (e: any) {
    toast.error(e.message || 'Failed to delete site')
  }
  siteToDelete.value = null
}

const handleSiteSort = async () => {
  try {
    await UpdateSites(sites.value)
  } catch (e: any) {
    toast.error(e.message || 'Failed to save order')
  }
}

const openWebsite = () => {
  BrowserOpenURL('https://gridea.pro')
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
