<template>
  <div class="pb-24 pt-4 pl-32">
    <div v-if="currentThemeConfig.length > 0">
      <div class="flex flex-col md:flex-row gap-8">
        <!-- Sidebar -->
        <aside class="w-full md:w-48 flex-shrink-0 md:border-r md:border-border md:pr-6" v-if="groups.length > 0">
          <nav class="space-y-1 sticky top-0">
            <button v-for="group in groups" :key="group" @click="activeGroup = group" :class="[
              'w-full text-left px-3 py-2 text-sm rounded-md transition-colors',
              activeGroup === group
                ? 'bg-primary text-primary-foreground font-medium'
                : 'text-muted-foreground hover:bg-muted hover:text-foreground'
            ]">
              {{ group }}
            </button>
          </nav>
        </aside>

        <!-- Right Content -->
        <div class="flex-1 min-w-0">
          <div class="space-y-6 m-0">
            <div v-for="(item, index1) in currentThemeConfig" :key="index1">
              <div v-if="item.group === activeGroup" class="space-y-2">
                <div class="flex justify-between items-center">
                  <label class="text-sm font-medium text-foreground">{{ item.label }}</label>
                </div>

                <div class="text-xs text-muted-foreground mb-2" v-if="item.note">{{ item.note }}</div>

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
                        <div class="w-4 h-4 rounded-full mr-2 border border-border" v-if="form[item.name]"
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
                        <posts-card :posts="postsWithLink"
                          @select="handlePostSelected($event, item.name);"></posts-card>
                      </div>
                    </PopoverContent>
                  </Popover>
                  <div class="text-xs text-muted-foreground mt-1" v-if="form[item.name]">{{
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

                <!-- Switch -->
                <div v-if="item.type === 'switch'">
                  <Switch v-model:checked="form[item.name]" />
                </div>

                <!-- Textarea -->
                <div v-if="item.type === 'textarea'" class="max-w-sm">
                  <Textarea v-model="form[item.name]" rows="4" />
                </div>

                <!-- Picture Upload -->
                <div v-if="item.type === 'picture-upload'" class="flex items-start gap-4">
                  <div
                    class="w-32 h-32 border-2 border-dashed border-input rounded-lg flex items-center justify-center cursor-pointer hover:border-primary transition-colors relative overflow-hidden bg-background flex-shrink-0"
                    @click="triggerFileInput(`fileInput-${index1}`)">
                    <img v-if="form[item.name]" :src="getImageUrl(form[item.name])"
                      class="w-full h-full object-cover" />
                    <div v-else class="flex flex-col items-center text-muted-foreground">
                      <i class="ri-upload-2-line text-2xl mb-1"></i>
                    </div>
                    <input type="file" :ref="(el) => setFileInputRef(el, `fileInput-${index1}`)" class="hidden"
                      accept="image/*" @change="(e) => handleImageUpload(e, item.name)" />
                  </div>
                  <Button variant="ghost" size="icon" v-if="form[item.name]" @click="resetFormItem(item.name)"
                    class="mt-2" title="Reset">
                    <i class="ri-arrow-go-back-line"></i>
                  </Button>
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
                      <Label class="block text-xs font-medium text-muted-foreground mb-1">{{ field.label }}</Label>

                      <!-- Array Item Input -->
                      <div class="max-w-sm" v-if="field.type === 'input' && !field.card">
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

                      <!-- Array Item Switch -->
                      <div v-if="field.type === 'switch'">
                        <Switch v-model:checked="configItem[field.name]" />
                      </div>

                      <!-- Array Item Picture -->
                      <div v-if="field.type === 'picture-upload'" class="flex items-center gap-2">
                        <div
                          class="w-16 h-16 border border-dashed border-input rounded flex items-center justify-center cursor-pointer overflow-hidden relative"
                          @click="triggerFileInput(`fileInput-${index1}-${configItemIndex}-${fieldIndex}`)">
                          <img v-if="configItem[field.name]" :src="getImageUrl(configItem[field.name])"
                            class="w-full h-full object-cover" />
                          <i v-else class="ri-add-line text-muted-foreground"></i>
                          <input type="file"
                            :ref="(el) => setFileInputRef(el, `fileInput-${index1}-${configItemIndex}-${fieldIndex}`)"
                            class="hidden" accept="image/*"
                            @change="(e) => handleImageUpload(e, item.name, field.name, Number(configItemIndex))" />
                        </div>
                        <button v-if="configItem[field.name]"
                          @click="resetFormItem(item.name, field.name, Number(configItemIndex))"
                          class="text-xs text-muted-foreground hover:text-destructive">Reset</button>
                      </div>

                    </div>
                  </div>
                  <Button variant="outline" class="w-full border-dashed"
                    v-if="!form[item.name] || form[item.name].length === 0"
                    @click="addConfigItem(item.name, -1, item.arrayItems)">
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
          <Button variant="ghost" size="icon" @click="resetThemeCustomConfig" title="Reset to defaults">
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
import MonacoMarkdownEditor from '@/components/MonacoMarkdownEditor/Index.vue'
import FooterBox from '@/components/FooterBox/Index.vue'
import ColorCard from '@/components/ColorCard/Index.vue'
import PostsCard from '@/components/PostsCard/Index.vue'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { EventsEmit, EventsOnce, ResolveFilePaths } from 'wailsjs/runtime'

// Modal logic replacement
const confirmReset = (callback: () => void) => {
  if (confirm('此操作将会使该主题配置恢复到初始状态，确认重置吗？')) {
    callback()
  }
}

const { t } = useI18n()
const router = useRouter()
const siteStore = useSiteStore()

const form = reactive<any>({})
const fileInputRefs = ref<Map<string, HTMLInputElement>>(new Map())

const setFileInputRef = (el: any, key: string) => {
  if (el) {
    fileInputRefs.value.set(key, el as HTMLInputElement)
  }
}

const triggerFileInput = (key: string) => {
  const input = fileInputRefs.value.get(key)
  input?.click()
}

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
  const list = siteStore.site.posts.map((post: any) => {
    return {
      ...post,
      link: urlJoin(siteStore.site.setting.domain, siteStore.site.themeConfig.postPath, post.fileName, '/'),
    }
  }).filter((post: any) => post.data.published)

  return list
})

const getImageUrl = (path: string) => {
  if (!path) return ''
  if (path.startsWith('http') || path.startsWith('data:')) return path

  let fullPath = path
  if (path.startsWith('/media/')) {
    fullPath = `${siteStore.site.appDir}/themes/${siteStore.site.themeConfig.themeName}/assets${path}`
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
  return (foundPost && foundPost.data.title) || ''
}

const saveThemeCustomConfig = () => {
  console.log('this.form', form)
  EventsEmit('theme-custom-config-save', { ...form })
  EventsOnce('theme-custom-config-saved', (result: any) => {
    if (result.success) {
      siteStore.site.themeCustomConfig = { ...form }
      toast.success(t('settings.theme.configSaved'))
    } else {
      toast.error(t('settings.theme.saveFailed')) // TODO: Check i18n key
    }
  })
}

const resetThemeCustomConfig = () => {
  confirmReset(() => {
    EventsEmit('theme-custom-config-save', {})
    EventsOnce('theme-custom-config-saved', async (result: any) => {
      if (result.success) {
        siteStore.site.themeCustomConfig = {}
        Object.keys(form).forEach(key => {
          delete form[key]
        })
        loadCustomConfig()
        toast.success(t('settings.theme.resetSuccess')) // TODO: Check i18n key
      } else {
        toast.error(t('settings.theme.resetFailed')) // TODO: Check i18n key
      }
    })
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

const handleImageUpload = async (event: Event, formItemName: string, arrayFieldItemName?: string, configItemIndex?: number) => {
  const files = (event.target as HTMLInputElement).files
  if (!files || files.length === 0) return

  const file = files[0]
  const isImage = file.type.indexOf('image') !== -1
  if (!isImage) return

  // Use ResolveFilePaths to get the actual file path
  const fileArray = Array.from(files)
  await ResolveFilePaths(fileArray as any)
  const filePath = (fileArray[0] as any).path

  if (arrayFieldItemName && typeof configItemIndex === 'number') {
    form[formItemName][configItemIndex][arrayFieldItemName] = filePath
  } else {
    form[formItemName] = filePath
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