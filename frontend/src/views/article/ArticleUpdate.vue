<template>
  <div class="article-update-page" :class="{ 'is-entering': entering }" v-if="visible" @mousemove="handlePageMousemove">
    <!-- Header -->
    <div class="page-title" ref="pageTitle">
      <div class="flex justify-end gap-2">
        <Button variant="ghost" size="sm"
          class="rounded-full text-muted-foreground hover:bg-primary/10 hover:text-foreground h-8 w-12 p-0"
          @click="close" :title="$t('common.back')">
          <ArrowLeftIcon class="size-3" />
        </Button>

        <Button variant="ghost" size="sm"
          class="rounded-full text-muted-foreground hover:bg-primary/10 hover:text-foreground h-8 w-12 p-0"
          :disabled="!canSubmit" @click="saveDraft" :title="$t('article.saveDraft')">
          <CheckIcon class="size-3" />
        </Button>

        <Button variant="ghost" size="sm"
          class="rounded-full text-primary hover:bg-primary/10 hover:text-primary h-8 w-12 p-0" @click="publishPost"
          :title="$t('article.publish')">
          <PaperAirplaneIcon class="size-3 -rotate-45" />
        </Button>
      </div>
    </div>

    <!-- Right Tools -->
    <div class="right-tool-container">
      <!-- Info Popover -->
      <Popover>
        <PopoverTrigger as-child>
          <Button variant="ghost" size="sm"
            class="rounded-full text-muted-foreground hover:text-foreground hover:bg-primary/10 h-8 w-8 p-0">
            <InformationCircleIcon class="size-4" />
          </Button>
        </PopoverTrigger>
        <PopoverContent side="left" align="start" class="w-48 p-4 bg-primary/10 transition-colors duration-200">
          <div class="post-stats">
            <div class="item">
              <h4>{{ $t('article.words') }}</h4>
              <div class="number text-foreground">{{ postStats.wordsNumber }}</div>
            </div>
            <div class="item">
              <h4>{{ $t('article.readingTime') }}</h4>
              <div class="number text-foreground">{{ postStats.formatTime }}</div>
            </div>
          </div>
        </PopoverContent>
      </Popover>

      <!-- Emoji Popover -->
      <Popover>
        <PopoverTrigger as-child>
          <Button variant="ghost" size="sm"
            class="rounded-full text-muted-foreground hover:text-foreground hover:bg-secondary h-8 w-8 p-0">
            <FaceSmileIcon class="size-4" />
          </Button>
        </PopoverTrigger>
        <PopoverContent side="left" align="start" class="w-[320px] p-0 overflow-hidden" :side-offset="10">
          <EmojiCard @select="handleEmojiSelect" />
        </PopoverContent>
      </Popover>

      <Button variant="ghost" size="sm"
        class="rounded-full text-muted-foreground hover:text-foreground hover:bg-secondary h-8 w-8 p-0"
        @click="insertImage" :title="$t('article.insertImage')">
        <PhotoIcon class="size-4" />
      </Button>

      <Button variant="ghost" size="sm"
        class="rounded-full text-muted-foreground hover:text-foreground hover:bg-secondary h-8 w-8 p-0"
        @click="insertMore" :title="$t('article.insertMore')">
        <EllipsisHorizontalIcon class="size-4" />
      </Button>

      <Button variant="ghost" size="sm"
        class="rounded-full text-muted-foreground hover:text-foreground hover:bg-secondary h-8 w-8 p-0"
        @click="handlePostSettingClick" :title="$t('article.settings')">
        <Cog6ToothIcon class="size-4" />
      </Button>

      <Button variant="ghost" size="sm"
        class="rounded-full text-muted-foreground hover:text-foreground hover:bg-secondary h-8 w-8 p-0"
        @click="previewPost" :title="`${$t('nav.preview')} [Ctrl + P]`">
        <EyeIcon class="size-4" />
      </Button>
    </div>

    <div class="right-bottom-tool-container">
      <Popover>
        <PopoverTrigger as-child>
          <Button variant="ghost" size="sm"
            class="rounded-full text-muted-foreground hover:text-foreground hover:bg-secondary h-8 w-8 p-0">
            <i class="ri-keyboard-line"></i>
          </Button>
        </PopoverTrigger>
        <PopoverContent side="left" align="end" class="w-64 p-4 max-h-[400px] overflow-y-auto">
          <div class="keyboard-tip mb-2">
            💁‍♂️ 编辑区域右键能弹出快捷菜单哦
          </div>
          <div class="keyboard-container w-full">
            <div class="item" v-for="(item, index) in shortcutKeys" :key="index">
              <div class="keyboard-group-title text-xs font-bold text-muted-foreground my-2 border-b pb-1">{{ item.name
                }}</div>
              <div class="list">
                <div class="list-item" v-for="(listItem, listIndex) in item.list" :key="listIndex">
                  <div class="list-item-title text-foreground">{{ listItem.title }}</div>
                  <div class="text-muted-foreground">
                    <span v-for="(keyCode, keyIndex) in listItem.keyboard" :key="keyIndex">
                      <code>{{ keyCode }}</code> <span v-if="keyIndex !== listItem.keyboard.length - 1"> + </span>
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </div>

    <!-- Content -->
    <div class="page-content">
      <div class="editor-wrapper">
        <input
          class="post-title py-2 border-none mt-4 mb-4 bg-transparent text-2xl focus:outline-none focus:ring-0 text-foreground placeholder:text-muted-foreground font-bold"
          :placeholder="$t('article.title')" v-model="form.title" @change="handleTitleChange"
          @keydown="handleInputKeydown" />

        <monaco-markdown-editor ref="monacoMarkdownEditor" v-model:value="form.content" @keydown="handleInputKeydown"
          :isPostPage="true" class="post-editor"></monaco-markdown-editor>

        <div class="footer-info">
          {{ $t('article.writingIn') }} <a @click.prevent="openPage('https://gridea.pro')"
            class="link cursor-pointer">Gridea Pro</a>
        </div>
      </div>

      <!-- Preview Drawer (Sheet) -->
      <Sheet v-model:open="previewVisible">
        <SheetContent side="right" class="w-screen max-w-4xl sm:max-w-4xl p-0">
          <div class="flex h-full flex-col overflow-y-scroll bg-background py-6 shadow-xl">
            <div class="px-4 sm:px-6">
              <div class="flex items-start justify-between">
                <SheetTitle class="text-lg font-medium text-foreground"></SheetTitle>
              </div>
            </div>
            <div class="relative mt-6 flex-1 px-4 sm:px-6">

              <h1 class="preview-title text-foreground">{{ form.title }}</h1>
              <div class="preview-date">{{ dayjs(form.date).format(siteStore.site.themeConfig.dateFormat) }}</div>
              <div class="preview-tags">
                <span class="tag" v-for="(tag, index) in form.tags" :key="index">
                  {{ tag }}
                </span>
              </div>
              <div class="preview-container" ref="previewContainerRef"></div>
            </div>
          </div>
        </SheetContent>
      </Sheet>

      <!-- Settings Drawer (Sheet) -->
      <Sheet v-model:open="postSettingsVisible">
        <SheetContent side="right" class="w-[400px] sm:max-w-md p-0 gap-0 flex flex-col">
          <SheetHeader class="px-6 py-6 border-b">
            <SheetTitle>{{ $t('article.settings') }}</SheetTitle>
          </SheetHeader>

          <div class="relative flex-1 px-6 py-6 space-y-6 overflow-y-auto">
            <!-- URL -->
            <div class="space-y-2">
              <Label>URL</Label>
              <Input v-model="form.fileName" @change="handleFileNameChange" />
            </div>


            <!-- Categories -->
            <div class="space-y-2">
              <Label>{{ $t('nav.category') }}</Label>
              <Select v-model="form.category">
                <SelectTrigger class="w-full">
                  <SelectValue :placeholder="$t('selectCategory')" /> <!-- // TODO: Check i18n key -->
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="_none_">{{ $t('none') }}</SelectItem> <!-- // TODO: Check i18n key -->
                  <SelectItem v-for="c in availableCategories" :key="c" :value="c">{{ c }}</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <!-- Tags -->
            <div class="space-y-2">
              <Label>{{ $t('nav.tag') }}</Label>
              <div>
                <div class="flex flex-wrap gap-2 p-2 border rounded-md bg-background min-h-[32px] mb-2">
                  <span v-for="tag in form.tags" :key="tag"
                    class="inline-flex items-center px-2 py-0.5 rounded-full bg-primary/10 border border-primary/20 text-xs text-primary/80">
                    {{ tag }}
                    <button @click="removeTag(tag)" class="ml-1 text-primary/60 hover:text-destructive">
                      <XMarkIcon class="size-3" />
                    </button>
                  </span>
                  <input v-model="tagInput" @keydown.enter.prevent="addTag"
                    class="flex-1 min-w-[80px] bg-transparent outline-none text-foreground text-sm px-1"
                    placeholder="Add tag..." />
                </div>
                <div class="flex flex-wrap gap-2 max-h-[120px] overflow-y-auto p-1 border rounded-md">
                  <span v-for="t in availableTags" :key="t" @click="selectTag(t)"
                    class="cursor-pointer text-xs px-2 py-1 rounded-full bg-primary/5 hover:bg-primary/15 border border-primary/10 transition-colors select-none text-muted-foreground">
                    {{ t }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Date -->
            <div class="space-y-2">
              <Label>{{ $t('article.createAt') }}</Label>
              <Popover>
                <PopoverTrigger as-child>
                  <Button variant="outline" :class="cn(
                    'w-full justify-start text-left font-normal hover:bg-primary/5 hover:text-primary border-primary/20 cursor-pointer',
                    !dateValue && 'text-muted-foreground',
                  )">
                    <CalendarIcon class="mr-2 h-4 w-4" />
                    {{ (form.date && form.date.isValid()) ? form.date.format('YYYY-MM-DD HH:mm:ss') : $t('pickDate') }}
                    <!-- // TODO: Check i18n key -->
                  </Button>
                </PopoverTrigger>
                <PopoverContent class="w-auto p-0" align="start">
                  <Calendar v-model="dateValue" show-week-number />
                  <div class="border-t p-3">
                    <Label class="text-xs text-muted-foreground mb-2 block capitalize">{{ $t('time') }}</Label>
                    <!-- // TODO: Check i18n key -->
                    <div class="relative">
                      <ClockIcon class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground z-10" />
                      <Input type="time" step="1" v-model="timeValue"
                        class="h-9 pl-9 accent-primary selection:bg-primary selection:text-primary-foreground" />
                    </div>
                  </div>
                </PopoverContent>
              </Popover>
            </div>

            <!-- Feature Image -->
            <div class="space-y-2">
              <Label>{{ $t('article.featureImage') }}</Label>
              <div class="space-y-2">
                <Input v-model="featureDisplayValue"
                  :placeholder="$t('article.featureImagePlaceholder') || 'Image URL or Local Path'" />

                <div
                  class="feature-uploader cursor-pointer border border-dashed rounded-md p-4 text-center hover:border-primary transition-colors bg-background"
                  @click="selectFeatureImage">
                  <div v-if="featureImagePreviewSrc">
                    <img class="feature-image mx-auto max-h-[150px] object-cover rounded-md"
                      :src="featureImagePreviewSrc" />
                  </div>
                  <div v-else>
                    <img src="@/assets/images/image_upload.svg" class="upload-img mx-auto w-20">
                    <i class="ri-upload-2-line upload-icon text-lg mt-2 block text-muted-foreground"></i>
                    <div class="text-xs text-muted-foreground mt-2">点击选择本地图片</div>
                  </div>
                </div>

                <Button v-if="featureDisplayValue" variant="destructive" size="sm" class="mt-2 w-full"
                  @click.stop="clearFeatureImage">
                  <template #icon>
                    <TrashIcon class="size-4 mr-2" />
                  </template>
                  {{ $t('common.delete') }}
                </Button>
              </div>
            </div>

            <!-- Hide in List -->
            <div class="flex items-center justify-between">
              <Label>{{ $t('article.hideInList') }}</Label>
              <Switch size="sm" v-model:checked="form.hideInList" />
            </div>

            <!-- Top Article -->
            <div class="flex items-center justify-between">
              <Label>{{ $t('article.top') }}</Label>
              <Switch size="sm" v-model:checked="form.isTop" />
            </div>
          </div>

          <SheetFooter class="flex-shrink-0 px-6 py-4 border-t gap-3">
            <Button variant="outline"
              class="w-18 h-8 text-xs justify-center rounded-full border border-primary/20 text-primary/80 hover:bg-primary/5 hover:text-primary cursor-pointer"
              @click="postSettingsVisible = false">
              {{ t('common.cancel') }}
            </Button>
            <Button variant="default"
              class="w-18 h-8 text-xs justify-center rounded-full bg-primary text-background hover:bg-primary/90 cursor-pointer"
              @click="handleConfirmPublish">
              {{ t('article.publish') }}
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      <!-- Unsaved Changes Dialog -->
      <Dialog v-model:open="showUnsavedDialog">
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{{ t('common.warning') }}</DialogTitle>
          </DialogHeader>
          <div class="flex items-center gap-3 text-destructive">
            <ExclamationTriangleIcon class="size-6" />
            <p class="text-sm text-foreground">{{ t('article.unsavedWarning') }}</p>
          </div>
          <DialogFooter>
            <Button variant="outline" @click="showUnsavedDialog = false">{{ t('common.cancel') }}</Button>
            <Button @click="confirmClose">{{ t('article.noSaveAndBack') }}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <!-- 编辑器点击图片上传用 -->
      <input ref="uploadInputRef" class="upload-input hidden" type="file" accept="image/*" @change="fileChangeHandler">

        <span class="save-tip">{{ postStatusTip }}</span>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, computed, onMounted, watch, onUnmounted, toRaw } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { useImageUrl } from '@/composables/useImageUrl'
import { usePostStats } from '@/composables/usePostStats'
import shortid from 'shortid'
import dayjs from 'dayjs'
import * as monaco from 'monaco-editor'
import Prism from 'prismjs'
import markdown from '@/helpers/markdown'
import MonacoMarkdownEditor from '@/components/MonacoMarkdownEditor/Index.vue'
import EmojiCard from '@/components/EmojiCard/Index.vue'
import slug from '@/helpers/slug'
import { IPost } from '@/interfaces/post'
import { UrlFormats } from '@/helpers/enums'
import shortcutKeys from '@/helpers/shortcut-keys'
import ga from '@/helpers/analytics'
import {
  ArrowLeftIcon,
  CheckIcon,
  InformationCircleIcon,
  FaceSmileIcon,
  PhotoIcon,
  Cog6ToothIcon,
  EyeIcon,
  TrashIcon,
  PaperAirplaneIcon,
  XMarkIcon,
  ExclamationTriangleIcon,
} from '@heroicons/vue/24/outline'
import { EllipsisHorizontalIcon } from '@heroicons/vue/24/solid'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Calendar } from '@/components/ui/calendar'
import { CalendarIcon, ClockIcon } from '@heroicons/vue/24/outline'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Sheet, SheetContent, SheetTitle, SheetHeader, SheetFooter } from '@/components/ui/sheet'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { toast } from '@/helpers/toast'
import { ITag } from '@/interfaces/tag'
import { EventsEmit, EventsOnce, EventsOn, EventsOff, BrowserOpenURL } from '@/wailsjs/runtime'
import { SavePostFromFrontend, UploadImagesFromFrontend } from '@/wailsjs/go/facade/PostFacade'
import { domain } from '@/wailsjs/go/models'

const props = defineProps<{
  visible: boolean
  articleFileName: string
}>()

const emit = defineEmits<{
  close: []
  fetchData: []
}>()

const { t } = useI18n()
const siteStore = useSiteStore()
const { getImageUrl } = useImageUrl()

const postSettingsVisible = ref(false)
const previewVisible = ref(false)
const changedAfterLastSave = ref(false)
const entering = ref(false)
const postStatusTip = ref('')
const previewTimestamp = ref(Date.now())

const featureDisplayValue = computed({
  get: () => {
    if (form.featureImage.path) {
      // Display relative path for better UX if possible
      const postImagesIndex = form.featureImage.path.indexOf('/post-images/')
      if (postImagesIndex !== -1) {
        return form.featureImage.path.substring(postImagesIndex)
      }
      return form.featureImage.path
    }
    return form.featureImagePath
  },
  set: (val: string) => {
    if (form.featureImage.path && val !== form.featureImage.path) {
      // Check if user is just editing the relative path view of the same absolute path
      const postImagesIndex = form.featureImage.path.indexOf('/post-images/')
      if (postImagesIndex !== -1) {
        const relativePath = form.featureImage.path.substring(postImagesIndex)
        if (val === relativePath) return // No change
      }

      // If truly changed, clear object and treat as string
      form.featureImage = { path: '', name: '', type: '' }
    }
    form.featureImagePath = val
  }
})

const featureImagePreviewSrc = computed(() => {
  const ts = previewTimestamp.value

  if (form.featureImage.path) {
    const url = getImageUrl(form.featureImage.path)
    console.log('featureImagePreviewSrc: local image', {
      path: form.featureImage.path,
      url: url,
      ts
    })
    return `${url}&t=${ts}`
  }

  // Handle relative path /post-images/ manually entered or loaded
  if (form.featureImagePath && form.featureImagePath.startsWith('/post-images/')) {
    const fullPath = `${siteStore.site.appDir}${form.featureImagePath}`
    return `${getImageUrl(fullPath)}&t=${ts}`
  }

  console.log('featureImagePreviewSrc: external URL', form.featureImagePath)
  return form.featureImagePath
})

const uploadInputRef = ref<HTMLInputElement | null>(null)
const featureUploadInputRef = ref<HTMLInputElement | null>(null)
const monacoMarkdownEditor = ref<{ editor: import('monaco-editor').editor.IStandaloneCodeEditor | null } | null>(null)
const previewContainerRef = ref<HTMLElement | null>(null)
const pageTitle = ref<HTMLElement | null>(null)
const showUnsavedDialog = ref(false)
const tagInput = ref('')

const form = reactive({
  title: '',
  fileName: '',
  tags: [] as string[],

  category: '',
  categories: [] as string[], // Keep for compatibility or remove if unused locally
  date: dayjs(),
  content: '',
  published: false,
  hideInList: false,
  isTop: false,
  featureImage: {
    path: '',
    name: '',
    type: '',
  },
  featureImagePath: '',
  deleteFileName: '',
})

// 编辑文章时，当前文章的索引
let currentPostIndex = -1
let originalFileName = ''
let fileNameChanged = false

// 使用 Composables
const { stats: postStats } = usePostStats(() => form.content)

const canSubmit = computed(() => {
  return form.title && form.content
})

const availableTags = computed(() => {
  return siteStore.tags.map((tag) => tag.name)
})

const availableCategories = computed(() => {
  return siteStore.categories.map((category) => category.name)
})

// Tags Logic
const addTag = () => {
  const val = tagInput.value.trim()
  if (val && !form.tags.includes(val)) {
    form.tags.push(val)
  }
  tagInput.value = ''
}

const removeTag = (tag: string) => {
  form.tags = form.tags.filter(t => t !== tag)
}

const selectTag = (tag: string) => {
  if (!form.tags.includes(tag)) {
    form.tags.push(tag)
  }
}

// Categories Logic
const categoryInput = ref('')

const addCategory = () => {
  const val = categoryInput.value.trim()
  if (val && !form.categories.includes(val)) {
    form.categories.push(val)
  }
  categoryInput.value = ''
}

const removeCategory = (category: string) => {
  form.categories = form.categories.filter(c => c !== category)
}

const selectCategory = (category: string) => {
  if (!form.categories.includes(category)) {
    form.categories.push(category)
  }
}

// Date Logic
import { cn } from '@/lib/utils'
import {
  DateFormatter,
  type DateValue,
  getLocalTimeZone,
  fromDate,
  CalendarDate,
} from '@internationalized/date'

const dateValue = computed({
  get: () => {
    let dVal = form.date
    if (!dayjs.isDayjs(dVal) || !dVal.isValid()) {
      dVal = dayjs()
    }
    try {
      const d = dVal.toDate()
      // Convert to CalendarDate via ZonedDateTime to ensure correct timezone handling
      const zdt = fromDate(d, getLocalTimeZone())
      // Return only the date part to the Calendar to avoid time constraints
      return new CalendarDate(zdt.year, zdt.month, zdt.day)
    } catch (e) {
      console.error('Failed to convert date', e)
      const now = new Date()
      return new CalendarDate(now.getFullYear(), now.getMonth() + 1, now.getDate())
    }
  },
  set: (val: any) => {
    if (!val) return

    // val should be a CalendarDate or similar object with year, month, day
    // We update form.date (dayjs) preserving the original time
    const current = dayjs.isDayjs(form.date) && form.date.isValid() ? form.date : dayjs()

    // val.month is 1-indexed (from @internationalized/date), dayjs is 0-indexed
    const newDate = current
      .year(val.year)
      .month(val.month - 1)
      .date(val.day)

    form.date = newDate
  }
})

const timeValue = computed({
  get: () => {
    return (form.date && dayjs.isDayjs(form.date) && form.date.isValid())
      ? form.date.format('HH:mm:ss')
      : '00:00:00'
  },
  set: (val: string) => {
    if (!val) return
    const [hours, minutes, seconds] = val.split(':').map(Number)
    if (isNaN(hours) || isNaN(minutes)) return

    let current = form.date
    if (!dayjs.isDayjs(current) || !current.isValid()) {
      current = dayjs()
    }

    // Update time portion
    form.date = current
      .hour(hours)
      .minute(minutes)
      .second(seconds || 0)
  }
})

const dateStr = computed(() => {
  return form.date.format('YYYY-MM-DDTHH:mm')
})

const updateDate = (e: Event) => {
  const val = (e.target as HTMLInputElement).value
  form.date = dayjs(val)
}

const buildCurrentForm = () => {
  const { articleFileName } = props
  previewTimestamp.value = Date.now()
  if (articleFileName) {
    fileNameChanged = true // 编辑文章标题时 URL 不跟随其变化
    currentPostIndex = siteStore.posts.findIndex((item: IPost) => item.fileName === articleFileName)
    if (currentPostIndex !== -1) {
      const currentPost = siteStore.posts[currentPostIndex]
      originalFileName = currentPost.fileName

      form.title = currentPost.data.title
      form.fileName = currentPost.fileName
      form.tags = currentPost.data.tags || []
      form.category = (currentPost.data.categories && currentPost.data.categories.length > 0) ? currentPost.data.categories[0] : ''
      form.categories = currentPost.data.categories || []
      form.date = dayjs(currentPost.data.date).isValid() ? dayjs(currentPost.data.date) : dayjs()
      form.content = currentPost.content
      form.published = currentPost.data.published
      form.hideInList = currentPost.data.hideInList
      form.isTop = currentPost.data.isTop

      if (currentPost.data.feature && currentPost.data.feature.includes('http')) {
        // External URL
        form.featureImagePath = currentPost.data.feature
        form.featureImage = { path: '', name: '', type: '' }
      } else if (currentPost.data.feature && currentPost.data.feature.startsWith('/post-images/')) {
        // Local image saved in post-images directory
        // Convert relative path to absolute path using appDir
        const fileName = currentPost.data.feature.substring(13) // Remove '/post-images/'
        const absolutePath = `${siteStore.site.appDir}/post-images/${fileName}`
        form.featureImage.path = absolutePath
        form.featureImage.name = fileName
        form.featureImagePath = ''
      } else {
        // No feature image
        form.featureImage = { path: '', name: '', type: '' }
        form.featureImagePath = ''
      }
    }
  } else if (siteStore.site.themeConfig.postUrlFormat === UrlFormats.ShortId) {
    form.fileName = shortid.generate()
  }
}

const selectFeatureImage = async () => {
  try {
    // Use Wails file dialog to get absolute path
    const filePath = await (window as any).go.app.App.OpenImageDialog()

    if (!filePath) {
      console.log('selectFeatureImage: User cancelled')
      return
    }

    console.log('selectFeatureImage: selected file path', filePath)

    // Extract filename from path
    const fileName = filePath.split(/[\\/]/).pop() || ''

    // Determine MIME type from extension
    const ext = fileName.split('.').pop()?.toLowerCase() || ''
    const mimeTypes: Record<string, string> = {
      'jpg': 'image/jpeg',
      'jpeg': 'image/jpeg',
      'png': 'image/png',
      'gif': 'image/gif',
      'webp': 'image/webp'
    }
    const mimeType = mimeTypes[ext] || 'image/jpeg'

    // Set form data
    form.featureImage = {
      name: fileName,
      path: filePath,
      type: mimeType
    }
    form.featureImagePath = ''

    console.log('selectFeatureImage: featureImage set to', form.featureImage)

    ga('Post', 'Post - set-local-feature-image', '')
  } catch (error) {
    console.error('selectFeatureImage: error', error)
    toast.error('选择图片失败')
  }
}

const clearFeatureImage = () => {
  form.featureImage = { path: '', name: '', type: '' }
  form.featureImagePath = ''
}

const close = () => {
  // Close editor
  if (changedAfterLastSave.value) {
    showUnsavedDialog.value = true
    return
  }
  emit('close')
}

const confirmClose = () => {
  showUnsavedDialog.value = false
  emit('close')
}

const updatePostSavedStatus = () => {
  postStatusTip.value = `${t('savedIn')} ${dayjs().format('HH:mm:ss')}` // TODO: Check i18n key
  changedAfterLastSave.value = false
}

const handleTitleChange = (val: any) => {
  // Note: v-model handles the update, this is just for side effects
  if (!fileNameChanged && siteStore.site.themeConfig.postUrlFormat === UrlFormats.Slug) {
    form.fileName = slug(form.title)
  }
}

const handleFileNameChange = (val: any) => {
  fileNameChanged = !!val
}

const openPage = (url: string) => {
  BrowserOpenURL(url)
}

const buildFileName = () => {
  if (form.fileName !== '') {
    return
  }

  form.fileName = siteStore.site.themeConfig.postUrlFormat === UrlFormats.Slug
    ? slug(form.title)
    : shortid.generate()
}

const checkArticleUrlValid = () => {
  const restPosts = JSON.parse(JSON.stringify(siteStore.posts))
  const foundPostIndex = restPosts.findIndex((post: IPost) => post.fileName === form.fileName)

  if (foundPostIndex !== -1) {
    if (currentPostIndex === -1) {
      // 新增文章时文件名和其他文章文件名冲突
      return false
    }
    restPosts.splice(currentPostIndex, 1)
    const index = restPosts.findIndex((post: IPost) => post.fileName === form.fileName)
    if (index !== -1) {
      return false
    }
  }

  currentPostIndex = currentPostIndex === -1 ? 0 : currentPostIndex

  return true
}

const formatForm = (published?: boolean) => {
  buildFileName()
  const valid = checkArticleUrlValid()
  if (!valid) {
    toast.error(t('postUrlRepeatTip'))
    return
  }

  if (form.fileName.includes('/')) {
    toast.error(t('postUrlIncludeTip'))
    return
  }

  // 文件名改变之后，删除原来文件
  if (form.fileName.toLowerCase() !== originalFileName.toLowerCase()) {
    form.deleteFileName = originalFileName
  }

  console.log('Format form data', JSON.parse(JSON.stringify(form)))
  const rawForm = toRaw(form)

  const formData = {
    title: rawForm.title,
    fileName: rawForm.fileName,
    tags: [...rawForm.tags],
    categories: (rawForm.category && rawForm.category !== '_none_') ? [rawForm.category] : [],
    date: rawForm.date.format('YYYY-MM-DD HH:mm:ss'),
    content: rawForm.content,
    published: typeof published === 'boolean' ? published : rawForm.published,
    hideInList: rawForm.hideInList,
    isTop: rawForm.isTop,
    featureImage: rawForm.featureImage.path ? {
      path: rawForm.featureImage.path || '',
      name: rawForm.featureImage.name || '',
      type: rawForm.featureImage.type || '',
    } : {
      path: '',
      name: '',
      type: '',
    },
    featureImagePath: !rawForm.featureImage.path && rawForm.featureImagePath ? (rawForm.featureImagePath || '') : '',
    deleteFileName: rawForm.deleteFileName || '',
    tagIds: [],
  }

  return new domain.PostInput(formData)
}

const saveDraft = async () => {
  console.log('Save draft clicked', canSubmit.value)
  // Save draft
  if (!canSubmit.value) return
  const formData = formatForm(false)
  console.log('Form data prepared', formData)
  // Submit form data
  if (!formData) return

  try {
    const data = await SavePostFromFrontend(formData)
    if (data && data.posts) siteStore.posts = data.posts as IPost[]
    if (data && data.tags) siteStore.tags = data.tags as ITag[]

    updatePostSavedStatus()
    toast.success(`🎉  ${t('draftSuccess')}`)
    emit('close')
  } catch (e) {
    console.error(e)
    toast.error('保存失败')
  }

  ga('Post', 'Post - click-save-draft', '')
}

const publishPost = () => {
  handlePostSettingClick()
}

const handleConfirmPublish = async () => {
  console.log('Confirm publish clicked', canSubmit.value)
  if (!canSubmit.value) return
  const formData = formatForm(true)
  console.log('Publish form data', formData)
  if (!formData) return

  try {
    const data = await SavePostFromFrontend(formData)
    if (data && data.posts) siteStore.posts = data.posts as IPost[]
    if (data && data.tags) siteStore.tags = data.tags as ITag[]

    updatePostSavedStatus()
    toast.success(`🎉  ${t('published')}`)
    postSettingsVisible.value = false
    emit('close')
  } catch (e) {
    console.error(e)
    toast.error('发布失败')
  }
}

const normalSavePost = async () => {
  if (!canSubmit.value) return
  const formData = formatForm()
  if (!formData) return

  try {
    await SavePostFromFrontend(formData)
    updatePostSavedStatus()
    emit('fetchData')
  } catch (e) {
    console.error(e)
  }
}

const insertImage = () => {
  uploadInputRef.value?.click()
  ga('Post', 'Post - click-insert-image', '')
}

const handlePostSettingClick = () => {
  console.log('Post settings clicked')
  postSettingsVisible.value = true
  ga('Post', 'Post - click-post-setting', '')
}

const handleInfoClick = () => {
  ga('Post', 'Post - click-post-info', '')
}

const handleEmojiClick = () => {
  ga('Post', 'Post - click-emoji-card', '')
}

const uploadImageFiles = async (files: domain.UploadedFile[]) => {
  try {
    const data = await UploadImagesFromFrontend(files)
    if (!monacoMarkdownEditor.value?.editor) {
      console.error('Monaco editor is not ready')
      return
    }

    const editor = monacoMarkdownEditor.value.editor
    for (const path of data) {
      // Use Wails AssetServer middleware to serve local files
      const url = `![](/local-file?path=${encodeURIComponent(path)})`

      const position = editor.getPosition()
      if (!position) return
      editor.executeEdits('', [{
        range: monaco.Range.fromPositions(position),
        text: url,
        forceMoveMarkers: true
      }])
    }
    editor.focus()
  } catch (e) {
    console.error(e)
    toast.error('上传图片失败')
  }
}

const insertMore = () => {
  if (!monacoMarkdownEditor.value?.editor) {
    console.error('Monaco editor is not ready')
    return
  }

  const editor = monacoMarkdownEditor.value.editor
  const position = editor.getPosition()
  if (!position) return
  editor.executeEdits('', [{
    range: monaco.Range.fromPositions(position),
    text: '\n<!-- more -->\n',
    forceMoveMarkers: true
  }])
  editor.focus()

  ga('Post', 'Post - click-add-more', '')
}

const handleEmojiSelect = (emoji: string) => {
  if (!monacoMarkdownEditor.value?.editor) {
    console.error('Monaco editor is not ready')
    return
  }

  const editor = monacoMarkdownEditor.value.editor
  const position = editor.getPosition()
  if (!position) return
  editor.executeEdits('', [{
    range: monaco.Range.fromPositions(position),
    text: emoji,
    forceMoveMarkers: true
  }])
  editor.focus()
}

const previewPost = () => {
  console.log('Preview post clicked')
  previewVisible.value = true
  setTimeout(() => {
    if (previewContainerRef.value) {
      previewContainerRef.value.innerHTML = markdown.render(form.content)
      Prism.highlightAll()
    }
  }, 1)

  ga('Post', 'Post - click-preview-post', '')
}

const fileChangeHandler = (e: any) => {
  const file = (e.target.files || e.dataTransfer)[0]
  if (!file) {
    return
  }
  const isImage = file.type.indexOf('image') !== -1
  if (!isImage) {
    return
  }
  if (file && isImage) {
    const uploadedFile = new domain.UploadedFile({
      name: file.name,
      path: file.path,
    })
    uploadImageFiles([uploadedFile])
  }
}

const handleInputKeydown = (e: KeyboardEvent) => {
  entering.value = true
  // Ctrl + P for preview
  if (e.ctrlKey && e.key === 'p') {
    e.preventDefault()
    previewPost()
  }
}

const handlePageMousemove = () => {
  entering.value = false
}

onMounted(() => {
  buildCurrentForm()
  EventsOff('click-menu-save')
  EventsOff('app-post-created')
  EventsOff('image-uploaded')

  EventsOn('click-menu-save', (data: any) => {
    normalSavePost()
  })

  watch(form, () => {
    changedAfterLastSave.value = true
  }, { deep: true })
})

onUnmounted(() => {
  EventsOff('click-menu-save')
  EventsOff('app-post-created')
  EventsOff('image-uploaded')
})
</script>

<style lang="less" scoped>
.article-update-page {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 40;
  background: var(--background);
  display: flex;
  flex-direction: column;

  .page-title {
    padding: 8px 16px;
    z-index: 41;
    background: var(--background);
    transition: opacity 700ms ease;
    border-bottom: 1px solid var(--border);
    --wails-draggable: drag;
  }

  .page-content {
    background: var(--background);
    flex: 1;
    display: flex;
    overflow: scroll;
  }

  &.is-entering {

    .page-title,
    .right-tool-container,
    .right-bottom-tool-container {
      opacity: 0;
    }
  }
}

.tip-text {
  margin: 8px 0;
  color: var(--muted-foreground);
}

#markdown-editor {
  :deep(.editor-toolbar) {
    position: fixed;
    top: 0px;
    z-index: 45;

    &:before {
      margin-bottom: 7px;
    }
  }
}

.editor-container {
  padding: 32px 89px 16px 24px;
  border: 1px solid var(--border);
  background: var(--card);
  box-shadow: 0 2px 8px rgba(115, 115, 115, 0.08);
}

.footer-info {
  text-align: center;
  color: var(--muted-foreground);
  font-size: 12px;
  font-weight: lighter;
  -webkit-font-smoothing: antialiased;
  padding-top: 6px;
  margin-top: 8px;
  border-top: 1px solid var(--border);

  .link {
    color: var(--muted-foreground);

    &:hover {
      color: var(--foreground);
    }
  }
}

.editor-wrapper {
  width: 100%;
  margin: 0 auto;
  position: relative;
  display: flex;
  flex-direction: column;

  .post-title {
    width: 728px;
    margin: 0 auto;
    display: block;
  }

  .post-editor {
    flex: 1;

    :deep(.monaco-markdown-editor) {
      width: 728px;
    }

    :deep(.monaco-editor) {
      .scrollbar {
        position: fixed !important;
        top: 110px !important;
      }
    }

    :deep(.monaco-editor),
    :deep(.monaco-editor-background) {
      background-color: transparent !important;
    }

    :deep(.monaco-editor .selected-text) {
      background-color: var(--primary) !important;
      opacity: 0.2;
    }

    :deep(.monaco-editor .view-lines) {
      user-select: text;
    }

    :deep(.monaco-editor .cursor) {
      visibility: inherit !important;
    }

    :deep(.monaco-editor .inputarea.ime-input) {
      z-index: 100 !important;
    }

    :deep(.monaco-editor .monaco-scrollable-element > .scrollbar > .slider) {
      background: rgba(121, 121, 121, 0.4) !important;

      &:hover {
        background: rgba(121, 121, 121, 0.6) !important;
      }
    }

    :deep(.monaco-editor .monaco-scrollable-element > .scrollbar.invisible.fade) {
      transition: opacity 0.3s cubic-bezier(0.3, 0.5, 0.5, 1);
    }
  }
}

.page-content {
  cursor: text;
}

.right-tool-container,
.right-bottom-tool-container {
  position: fixed;
  right: 12px;
  display: flex;
  flex-direction: column;
  color: var(--muted-foreground);
  transition: color 0.3s ease;
  transition: opacity 700ms ease;
  z-index: 45;
  pointer-events: none;

  &:hover {
    color: var(--foreground);
  }
}

.right-tool-container button,
.right-tool-container [role="button"],
.right-bottom-tool-container button,
.right-bottom-tool-container [role="button"] {
  pointer-events: auto;
}

.right-tool-container {
  bottom: 50%;
  transform: translateY(50%);
}

.right-bottom-tool-container {
  bottom: 2px;
}

.save-tip {
  padding: 4px 10px;
  line-height: 22px;
  font-size: 12px;
  color: var(--muted-foreground);
  position: fixed;
  left: 0;
  bottom: 0;
}

.preview-container {
  width: 100%;
  flex-shrink: 0;
  font-family: "Noto Serif", "PingFang SC", "Hiragino Sans GB", "Droid Sans Fallback", "Microsoft YaHei", sans-serif;
  font-size: 15px;
  color: var(--foreground);

  :deep(a) {
    color: var(--foreground);
    word-wrap: break-word;
    text-decoration: none;
    border-bottom: 1px solid var(--border);

    &:hover {
      color: var(--primary);
      border-bottom: 1px solid var(--primary);
    }
  }

  :deep(img) {
    display: block;
    max-width: 100%;
    border-radius: 2px;
    margin: 24px auto;
  }

  :deep(p) {
    line-height: 1.62;
    margin-bottom: 1.12em;
    font-size: 15px;
    letter-spacing: .05em;
    hyphens: auto;
  }

  :deep(p),
  :deep(li) {
    line-height: 1.62;

    code {
      font-family: 'Source Code Pro', Consolas, Menlo, Monaco, 'Courier New', monospace;
      line-height: initial;
      word-wrap: break-word;
      border-radius: 0;
      background-color: var(--secondary);
      color: var(--primary);
      padding: .2em .33333333em;
      font-size: .875rem;
      margin-left: .125em;
      margin-right: .125em;
    }
  }

  :deep(pre) {
    background: var(--secondary);
    padding: 16px;
    border-radius: 2px;

    code {
      color: var(--foreground);
      font-family: 'Source Code Pro', Consolas, Menlo, Monaco, 'Courier New', monospace;
    }
  }

  :deep(blockquote) {
    color: var(--muted-foreground);
    position: relative;
    padding: .4em 0 0 2.2em;
    font-size: .96em;

    &:before {
      position: absolute;
      top: -4px;
      left: 0;
      content: "\201c";
      font: 700 62px/1 serif;
      color: var(--border);
    }
  }

  :deep(table) {
    border-collapse: collapse;
    margin: 1rem 0;
    width: 100%;

    tr {
      border-top: 1px solid var(--border);

      &:nth-child(2n) {
        background-color: var(--secondary);
      }
    }

    td,
    th {
      border: 1px solid var(--border);
      padding: .6em 1em;
    }
  }

  :deep(ul),
  :deep(ol) {
    padding-left: 35px;
    line-height: 1.62;
    margin-bottom: 16px;
  }

  :deep(ol) {
    list-style: decimal !important;
  }

  :deep(ul) {
    list-style-type: square !important;
  }

  :deep(h1),
  h2,
  h3,
  h4,
  h5,
  h6 {
    margin: 16px 0;
    font-weight: 700;
    padding-top: 16px;
  }

  :deep(h1) {
    font-size: 1.8em;
  }

  :deep(h2) {
    font-size: 1.42em;
  }

  :deep(h3) {
    font-size: 1.17em;
  }

  :deep(h4) {
    font-size: 1em;
  }

  :deep(h5) {
    font-size: 1em;
  }

  :deep(h6) {
    font-size: 1em;
    font-weight: 500;
  }

  :deep(hr) {
    display: block;
    border: 0;
    margin: 2.24em auto 2.86em;

    &:before {
      color: rgba(0, 0, 0, .2);
      font-size: 1.1em;
      display: block;
      content: "* * *";
      text-align: center;
    }
  }

  :deep(.footnotes) {
    margin-left: auto;
    margin-right: auto;
    max-width: 760px;
    padding-left: 18px;
    padding-right: 18px;

    &:before {
      content: "";
      display: block;
      border-top: 4px solid rgba(0, 0, 0, .1);
      width: 50%;
      max-width: 100px;
      margin: 40px 0 20px;
    }
  }

  :deep(.contains-task-list) {
    list-style-type: none;
    padding-left: 30px;
  }

  :deep(.task-list-item) {
    position: relative;
  }

  :deep(.task-list-item-checkbox) {
    position: absolute;
    cursor: pointer;
    width: 16px;
    height: 16px;
    margin: 4px 0 0;
    top: -1px;
    left: -22px;
    transform-origin: center;
    transform: rotate(-90deg);
    transition: all .2s ease;

    &:checked {
      transform: rotate(0);

      &:before {
        border: transparent;
        background-color: #9AE6B4;
      }

      &:after {
        transform: rotate(-45deg) scale(1);
      }

      +.task-list-item-label {
        color: #999;
        text-decoration: line-through;
      }
    }

    &:before {
      content: "";
      width: 16px;
      height: 16px;
      box-sizing: border-box;
      display: inline-block;
      border: 1px solid #9AE6B4;
      border-radius: 2px;
      background-color: #fff;
      position: absolute;
      top: 0;
      left: 0;
      transition: all .2s ease;
    }

    &:after {
      content: "";
      transform: rotate(-45deg) scale(0);
      width: 9px;
      height: 5px;
      border: 1px solid #22543D;
      border-top: none;
      border-right: none;
      position: absolute;
      display: inline-block;
      top: 4px;
      left: 4px;
      transition: all .2s ease;
    }
  }

  :deep(.markdownIt-TOC) {
    list-style: none;
    background: #f7fafc;
    padding: 1.5rem;
    border-radius: 0.5rem;
    color: #4a5568;
  }

  :deep(.markdownIt-TOC ul) {
    list-style: none;
    padding-left: 16px;
  }

  :deep(mark) {
    background: #FAF089;
    color: #744210;
  }
}

.preview-title {
  font-size: 24px;
  font-weight: bold;
  font-family: "Noto Serif", "PingFang SC", "Hiragino Sans GB", "Droid Sans Fallback", "Microsoft YaHei", sans-serif;
}

.preview-date {
  font-size: 13px;
  color: var(--muted-foreground);
  margin-bottom: 16px;
}

.preview-tags {
  font-size: 12px;
  margin-bottom: 16px;

  .tag {
    display: inline-block;
    margin: 0 8px 8px 0;
    padding: 4px 8px;
    background: var(--secondary);
    color: var(--muted-foreground);
    border-radius: 20px;
  }
}

.preview-feature-image {
  max-width: 100%;
  margin-bottom: 16px;
  border-radius: 2px;
}

.keyboard-container {
  width: 200px;

  .keyboard-group-title {
    margin: 8px 0;
    font-size: 12px;
  }

  .list {
    .list-item {
      display: flex;
      justify-content: space-between;
      font-size: 12px;
      padding: 4px;
      border-radius: 2px;

      &:not(:last-child) {
        border-bottom: 1px solid var(--secondary);
      }

      &:hover {
        background: var(--secondary);
        color: var(--primary);
      }

      code {
        padding: 0px 4px;
        border-radius: 2px;
        background: var(--secondary);
      }
    }
  }
}

.post-stats {
  display: flex;

  .item {
    width: 50%;
    min-width: 80px;

    h4 {
      color: var(--muted-foreground);
      font-size: 12px;
      font-weight: normal;
    }

    .number {
      font-size: 18px;
      font-family: 'Noto Serif';
    }
  }
}

.keyboard-tip {
  font-size: 12px;
  color: var(--muted-foreground);
}
</style>
