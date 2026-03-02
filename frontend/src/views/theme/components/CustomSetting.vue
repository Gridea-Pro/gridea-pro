<template>
  <div class="pb-24 pt-4 pl-32">
    <div v-if="currentThemeConfig.length > 0">
      <div class="flex flex-col md:flex-row gap-8">
        <!-- Sidebar -->
        <aside v-if="groups.length > 0" class="w-full md:w-48 flex-shrink-0 md:border-r md:border-border md:pr-6">
          <nav class="space-y-1 sticky top-0">
            <button v-for="group in groups" :key="group" :class="[
              'w-full h-12 text-left px-3 py-2 text-sm rounded-md transition-colors',
              activeGroup === group
                ? 'text-primary bg-primary/10 font-medium cursor-pointer'
                : 'text-muted-foreground hover:bg-primary/10 hover:text-foreground cursor-pointer'
            ]" @click="activeGroup = group">
              {{ group }}
            </button>
          </nav>
        </aside>

        <!-- Right Content -->
        <div class="flex-1 min-w-0">
          <div class="space-y-6 m-0">
            <div v-for="(item, index1) in currentThemeConfig" :key="index1">
              <div v-if="item && item.group === activeGroup" class="space-y-2">
                <div class="flex justify-between items-center"
                  :class="{ 'max-w-sm': item.type === 'switch' || item.type === 'toggle' }">
                  <label class="text-sm font-medium text-foreground">{{ item.label }}</label>
                  <div v-if="item.type === 'switch' || item.type === 'toggle'">
                    <Switch v-model:checked="form[item.name]" size="sm" />
                  </div>
                </div>

                <div v-if="item.note" class="text-xs text-muted-foreground mb-2">{{ item.note }}</div>

                <!-- Input -->
                <div v-if="item.type === 'input' && !item.card" class="max-w-sm">
                  <Input v-model="form[item.name]" />
                </div>

                <!-- Color Input -->
                <div v-if="item.type === 'input' && item.card === 'color'" class="relative max-w-sm">
                  <Popover>
                    <PopoverTrigger as-child>
                      <button
                        class="flex items-center w-full px-3 py-2 border border-input rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-ring text-left">
                        <div v-if="form[item.name]" class="w-4 h-4 rounded-full mr-2 border border-border"
                          :style="{ backgroundColor: form[item.name] }"></div>
                        <span v-else class="text-muted-foreground">Select color</span>
                        <span class="flex-1">{{ form[item.name] }}</span>
                      </button>
                    </PopoverTrigger>
                    <PopoverContent class="w-auto p-3">
                      <color-card @change="handleColorChange($event, item.name)"></color-card>
                    </PopoverContent>
                  </Popover>
                </div>

                <!-- Post Input -->
                <div v-if="item.type === 'input' && item.card === 'post'" class="max-w-sm">
                  <Popover>
                    <PopoverTrigger as-child>
                      <button
                        class="w-full text-left px-3 py-2 border border-input rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-ring">
                        {{ form[item.name] || 'Select Post' }}
                      </button>
                    </PopoverTrigger>
                    <PopoverContent class="w-80 p-0">
                      <div class="max-h-96 overflow-auto">
                        <article-select-card :posts="postsWithLink"
                          @select="handlePostSelected($event, item.name);"></article-select-card>
                      </div>
                    </PopoverContent>
                  </Popover>
                  <div v-if="form[item.name]" class="text-xs text-muted-foreground mt-1">{{
                    getPostTitleByLink(form[item.name]) }}</div>
                </div>

                <!-- Select -->
                <div v-if="item.type === 'select' && item.options" class="max-w-sm">
                  <Select :model-value="String(form[item.name] || '')" @update:model-value="(v) => form[item.name] = v">
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem v-for="(option, index2) in item.options" :key="String(option.value)"
                        :value="String(option.value)">
                        {{ option.label }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <!-- Select (was Radio) -->
                <div v-if="item.type === 'radio' && item.options">
                  <div class="w-full max-w-sm">
                    <Select :model-value="String(form[item.name] || '')"
                      @update:model-value="(v) => form[item.name] = v">
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem v-for="(option, index2) in item.options" :key="String(option.value)"
                          :value="String(option.value)">
                          {{ option.label }}
                        </SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                <!-- Textarea -->
                <div v-if="item.type === 'textarea'" class="max-w-sm">
                  <Textarea v-model="form[item.name]" rows="4" />
                </div>

                <!-- Picture Upload -->
                <div v-if="['picture-upload', 'picture', 'image'].includes(item.type)" class="space-y-3">
                  <Input v-model="form[item.name]" placeholder="输入在线图片链接或点击下方虚线框上传" class="max-w-sm" />
                  <div class="flex items-start gap-4">
                    <div
                      class="w-24 h-24 border border-dashed border-input rounded-lg flex items-center justify-center cursor-pointer hover:border-primary transition-colors relative overflow-hidden bg-background shrink-0"
                      @mouseenter="($event.currentTarget as HTMLElement).querySelector('.delete-btn')?.classList.remove('hidden')"
                      @mouseleave="($event.currentTarget as HTMLElement).querySelector('.delete-btn')?.classList.add('hidden')"
                      @click="handleImageUpload(item.name)">
                      <img v-if="form[item.name]" :src="getImageUrl(form[item.name])"
                        class="w-full h-full object-cover" />
                      <div v-else class="flex flex-col items-center text-muted-foreground">
                        <i class="ri-upload-2-line text-2xl mb-1"></i>
                      </div>
                      <!-- 悬浮删除/重置按钮 -->
                      <div v-if="form[item.name]"
                        class="delete-btn hidden absolute top-1 right-1 bg-red-500 hover:bg-red-600 text-white rounded-full w-5 h-5 flex items-center justify-center z-10 shadow-sm border border-white transition-colors cursor-pointer"
                        title="移除图片" @click.stop="resetFormItem(item.name)">
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
                          class="w-3.5 h-3.5">
                          <path
                            d="M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z" />
                        </svg>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Markdown -->
                <div v-if="item.type === 'markdown'" class="border border-input rounded-lg overflow-hidden shadow-sm">
                  <monaco-markdown-editor ref="monacoMarkdownEditor"
                    v-model:value="form[item.name]"></monaco-markdown-editor>
                </div>

                <!-- Array -->
                <div v-if="item.type === 'array'" class="space-y-4">
                  <div v-for="(configItem, configItemIndex) in form[item.name]" :key="configItemIndex"
                    class="p-4 border border-input rounded-lg bg-card relative group">
                    <div class="absolute top-2 right-2 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                      <Button size="icon" variant="ghost"
                        class="h-6 w-6 text-blue-600 hover:text-blue-700 hover:bg-blue-100"
                        @click="addConfigItem(item.name, Number(configItemIndex), item.arrayItems)">
                        <i class="ri-add-line"></i>
                      </Button>
                      <Button size="icon" variant="ghost"
                        class="h-6 w-6 text-destructive hover:text-destructive hover:bg-destructive/10"
                        @click="deleteConfigItem(form[item.name], Number(configItemIndex))">
                        <i class="ri-subtract-line"></i>
                      </Button>
                    </div>

                    <div v-for="(field, fieldIndex) in item.arrayItems" :key="fieldIndex" class="mb-4 last:mb-0">
                      <!-- Array Item Switch -->
                      <div v-if="field.type === 'switch' || field.type === 'toggle'"
                        class="flex items-center justify-between max-w-sm">
                        <Label class="text-xs font-medium text-foreground">{{ field.label }}</Label>
                        <Switch v-model:checked="configItem[field.name]" size="sm" />
                      </div>

                      <Label v-else class="block text-xs font-medium text-muted-foreground mb-1">{{ field.label
                      }}</Label>

                      <!-- Array Item Input -->
                      <div v-if="field.type === 'input' && !field.card" class="max-w-sm">
                        <Input v-model="configItem[field.name]" />
                      </div>

                      <!-- Array Item Select -->
                      <div v-if="field.type === 'select'" class="max-w-sm">
                        <Select :model-value="String(configItem[field.name] || '')"
                          @update:model-value="(v) => configItem[field.name] = v">
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem v-for="opt in field.options" :key="String(opt.value)"
                              :value="String(opt.value)">
                              {{ opt.label }}
                            </SelectItem>
                          </SelectContent>
                        </Select>
                      </div>

                      <!-- Array Item Picture -->
                      <div v-if="['picture-upload', 'picture', 'image'].includes(field.type)" class="space-y-3">
                        <Input v-model="configItem[field.name]" placeholder="输入在线图片链接或点击下方虚线框上传" class="max-w-sm" />
                        <div class="flex items-center gap-2">
                          <div
                            class="relative w-full h-32 border-2 border-dashed border-gray-300 dark:border-zinc-700 rounded-lg overflow-hidden flex items-center justify-center cursor-pointer hover:border-gray-400 dark:hover:border-zinc-500 transition-colors"
                            @mouseenter="($event.currentTarget as HTMLElement).querySelector('.delete-btn')?.classList.remove('hidden')"
                            @mouseleave="($event.currentTarget as HTMLElement).querySelector('.delete-btn')?.classList.add('hidden')"
                            @click="handleImageUpload(item.name, field.name, Number(configItemIndex))">
                            <img v-if="configItem[field.name]" :src="getImageUrl(configItem[field.name])"
                              class="w-full h-full object-cover" />
                            <i v-else class="ri-add-line text-muted-foreground"></i>

                            <!-- 悬浮删除/重置按钮 -->
                            <div v-if="configItem[field.name]"
                              class="delete-btn hidden absolute top-2 right-2 bg-red-500 hover:bg-red-600 text-white rounded-full w-5 h-5 flex items-center justify-center z-10 shadow-sm border border-white transition-colors cursor-pointer"
                              title="移除图片" @click.stop="resetFormItem(item.name, field.name, Number(configItemIndex))">
                              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
                                class="w-3.5 h-3.5">
                                <path
                                  d="M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z" />
                              </svg>
                            </div>
                          </div>
                        </div>
                      </div>

                    </div>
                  </div>
                  <Button v-if="!form[item.name] || form[item.name].length === 0" variant="outline"
                    class="w-full border-dashed" @click="addConfigItem(item.name, -1, item.arrayItems)">
                    <i class="ri-add-line mr-2"></i> Add Item
                  </Button>
                </div>

              </div>
            </div>
          </div>
        </div>
      </div>

      <footer-box>
        <div class="flex justify-between w-full">
          <Button variant="ghost" size="icon" title="Reset to defaults" @click="resetThemeCustomConfig">
            <i class="ri-arrow-go-back-line text-lg"></i>
          </Button>
          <Button variant="default"
            class="w-18 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
            @click="saveThemeCustomConfig">
            {{ t('common.save') }}
          </Button>
        </div>
      </footer-box>
    </div>

    <div v-else class="flex flex-col items-center justify-center py-20 text-muted-foreground">
      <img class="w-32 h-32 mb-4 opacity-50" src="@/assets/images/graphic-empty-box.svg" alt="">
      <div class="text-lg">{{ t('settings.theme.noCustomConfigTip') }}</div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useSiteStore } from '@/stores/site'
import { toast } from '@/helpers/toast'
import urlJoin from 'url-join'
import MonacoMarkdownEditor from '@/components/MonacoMarkdownEditor/index.vue'
import FooterBox from '@/components/FooterBox/index.vue'
import ColorCard from '@/components/ColorCard/index.vue'
import ArticleSelectCard from '@/views/articles/list/components/ArticleSelectCard.vue'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { EventsEmit, EventsOnce, ResolveFilePaths } from '@/wailsjs/runtime'
import { SaveThemeCustomConfigFromFrontend, UploadThemeCustomConfigImage } from '@/wailsjs/go/facade/ThemeFacade'

// Modal logic replacement
const confirmReset = (callback: () => void) => {
  if (confirm('此操作将会使该主题配置恢复到初始状态，确认重置吗？')) {
    callback()
  }
}

const { t } = useI18n()
const router = useRouter()
const siteStore = useSiteStore()

const form = reactive<Record<string, any>>({})

const currentThemeConfig = computed<any[]>(() => {
  return (siteStore.site.currentThemeConfig || []) as unknown as any[]
})

const groups = computed(() => {
  if (!currentThemeConfig.value) return []
  let list = currentThemeConfig.value.map((item: any) => item.group)
  list = list.filter((g: any) => g) // filter undefined or null
  list = [...new Set(list)]
  return list
})

const activeGroup = ref('')

watch(groups, (newVal) => {
  if (newVal.length > 0 && !activeGroup.value) {
    activeGroup.value = newVal[0]
  }
}, { immediate: true })

const postsWithLink = computed(() => {
  if (!siteStore.site.posts) return []
  const list = siteStore.site.posts.map((post: any) => {
    return {
      ...post,
      link: urlJoin(
        siteStore.site.setting?.domain || '',
        siteStore.site.themeConfig?.postPath || '',
        post.fileName || '',
        '/'
      ),
    }
  }).filter((post: any) => post.data && post.data.published)

  return list
})

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

const loadCustomConfig = () => {
  const keys = Object.keys(siteStore.site.themeCustomConfig || {})
  keys.forEach((key: string) => {
    form[key] = siteStore.site.themeCustomConfig[key]
  })
  currentThemeConfig.value.forEach((item: any) => {
    if (form[item.name] === undefined) {
      form[item.name] = item.value
    }
  })
}

onMounted(() => {
  loadCustomConfig()
})

const getPostTitleByLink = (link: string) => {
  const foundPost = postsWithLink.value.find((post: any) => post.link === link)
  return (foundPost && foundPost.data && foundPost.data.title) || ''
}

const saveThemeCustomConfig = async () => {
  console.log('this.form', form)
  try {
    await SaveThemeCustomConfigFromFrontend(form)
    siteStore.site.themeCustomConfig = { ...form }
    toast.success(t('settings.theme.configSaved'))
  } catch (e) {
    console.error(e)
    toast.error(t('settings.theme.saveFailed'))
  }
}

const resetThemeCustomConfig = () => {
  confirmReset(async () => {
    try {
      await SaveThemeCustomConfigFromFrontend({})
      siteStore.site.themeCustomConfig = {}
      Object.keys(form).forEach(key => {
        delete form[key]
      })
      loadCustomConfig()
      toast.success(t('settings.theme.resetSuccess'))
    } catch (e) {
      console.error(e)
      toast.error(t('settings.theme.resetFailed'))
    }
  })
}

const handleColorChange = (color: string, name: string, arrayIndex?: number, fieldName?: string) => {
  if (arrayIndex === undefined) {
    form[name] = color
  } else if (arrayIndex !== undefined && fieldName !== undefined) {
    form[name][arrayIndex][fieldName] = color
  }
}

const handlePostSelected = (postUrl: string, name: string, arrayIndex?: number, fieldName?: string) => {
  console.log('postUrl', postUrl)
  if (arrayIndex === undefined) {
    form[name] = postUrl
  } else if (arrayIndex !== undefined && fieldName !== undefined) {
    form[name][arrayIndex][fieldName] = postUrl
  }
}

const handleImageUpload = async (formItemName: string, arrayFieldItemName?: string, configItemIndex?: number) => {
  try {
    const filePath = await (window as any).go.app.App.OpenImageDialog()
    if (!filePath) return

    const uploadedUrl = await UploadThemeCustomConfigImage(filePath)
    if (arrayFieldItemName && typeof configItemIndex === 'number') {
      form[formItemName][configItemIndex][arrayFieldItemName] = uploadedUrl
    } else {
      form[formItemName] = uploadedUrl
    }
  } catch (error) {
    console.error('UploadThemeCustomConfigImage error', error)
    toast.error(`Upload failed: ${error}`)
  }
}

const resetFormItem = (formItemName: string, arrayFieldItemName?: string, configItemIndex?: number) => {
  const originalItem = currentThemeConfig.value.find((item: any) => item.name === formItemName)
  if (arrayFieldItemName && typeof configItemIndex === 'number') {
    const foundItem = originalItem.arrayItems.find((item: any) => item.name === arrayFieldItemName)
    form[formItemName][configItemIndex][arrayFieldItemName] = foundItem.value
  } else {
    form[formItemName] = originalItem.value
  }
}

const deleteConfigItem = (formItem: any[], index: number) => {
  console.log('run...', formItem, index)
  formItem.splice(index, 1)
}

const addConfigItem = (name: string, index: number, arrayItems: any) => {
  if (!form[name]) {
    form[name] = []
  }
  const newValue = arrayItems.reduce((o: any, c: any) => {
    o[c.name] = c.value
    return o
  }, {})
  // index + 1 inserts after current item. If -1, inserts at 0.
  form[name].splice(index + 1, 0, newValue)
}
</script>