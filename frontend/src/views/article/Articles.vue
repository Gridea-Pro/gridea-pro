<template>
  <div class="h-full flex flex-col bg-background text-foreground">
    <!-- Header Tools -->
    <div
      class="flex-shrink-0 flex justify-between items-center px-4 h-12 bg-background border-b border-border select-none"
      style="--wails-draggable: drag">
      <div class="flex items-center min-h-[32px]" style="--wails-draggable: no-drag">
        <Transition enter-active-class="transition ease-out duration-200"
          enter-from-class="opacity-0 translate-x-[-10px]" enter-to-class="opacity-100 translate-x-0"
          leave-active-class="transition ease-in duration-150" leave-from-class="opacity-100 translate-x-0"
          leave-to-class="opacity-0 translate-x-[-10px]">
          <div v-if="selectedPost.length > 0" @click="deleteModalVisible = true"
            class="flex items-center py-1.5 px-3 bg-destructive/10 text-destructive cursor-pointer hover:bg-destructive/20 rounded-md text-xs transition-colors border border-destructive/20">
            <TrashIcon class="size-3 mr-2 mb-0.5" />
            <span>{{ t('common.delete') }} {{ selectedPost.length }}</span>
          </div>
        </Transition>
      </div>
      <div class="flex items-center gap-3" style="--wails-draggable: no-drag">
        <div v-if="searchInputVisible" class="relative">
          <Input v-model="keyword" class="w-[200px] h-8 pl-8 text-xs rounded-full" :placeholder="t('article.search')"
            @blur="handleSearchInputBlur" ref="searchInputRef" autofocus />
          <MagnifyingGlassIcon class="absolute left-2.5 top-2 size-4 text-muted-foreground" />
        </div>
        <div v-else
          class="flex items-center justify-center w-8 h-8 rounded-full hover:bg-primary/10 cursor-pointer transition-colors text-muted-foreground hover:text-foreground"
          @click="showSearchInput" :title="t('article.search')">
          <MagnifyingGlassIcon class="size-4" />
        </div>
        <div
          class="flex items-center justify-center w-8 h-8 rounded-full hover:bg-primary/10 cursor-pointer transition-colors text-muted-foreground hover:text-foreground"
          @click="newArticle" :title="t('article.new')">
          <PlusIcon class="size-4" />
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto px-4 py-6 pb-20">
      <div class="space-y-3">
        <div v-for="post in currentPostList" :key="post.fileName"
          class="group relative flex rounded-xl relative cursor-pointer transition-all duration-200 bg-primary/2 border border-primary/10 hover:bg-primary/10 hover:shadow-xs hover:-translate-y-0.5"
          @click="editPost(post)">
          <div class="p-3 flex-1 flex flex-col sm:flex-row">
            <!-- Checkbox & Info -->
            <div class="flex flex-1 min-w-0">
              <div class="flex flex-shrink-0 items-center p-3" @click.stop>
                <Checkbox :checked="selectedPost.some(p => p.fileName === post.fileName)"
                  @update:checked="onSelectChange(post)"
                  class="data-[state=checked]:bg-primary data-[state=checked]:border-primary" />
              </div>
              <div class="flex-1 h-14 min-w-0">
                <div
                  class="text-[13px] font-medium mt-1.5 mb-2 text-foreground truncate pr-16 group-hover:text-primary transition-colors max-w-[600px]">
                  {{ post.data.title }}</div>
                <div class="flex items-center text-xs text-muted-foreground gap-3 flex-wrap">
                  <div class="flex items-center text-[10px]">
                    <div class="w-1.5 h-1.5 rounded-full mr-1.5"
                      :class="post.data.published ? 'bg-green-500' : 'bg-gray-300'"></div>
                    {{ post.data.published ? t('article.published') : t('article.draft') }}
                  </div>
                  <div class="w-px h-3 bg-primary/30"></div>
                  <div class="flex items-center text-[10px]">
                    <CalendarIcon class="size-3 mr-1 text-muted-foreground/70 translate-y-[-0.5px]" />
                    {{ dayjs(post.data.date).format('YYYY-MM-DD') }}
                  </div>
                  <template v-if="(post.data.categories || []).length > 0">
                    <div class="w-px h-3 bg-primary/30"></div>
                    <div class="flex items-center text-[10px] text-muted-foreground/70">
                      <FolderIcon class="size-3 mr-1" />
                      {{ (post.data.categories || [])[0] }}
                    </div>
                  </template>
                  <template v-if="(post.data.tags || []).length > 0">
                    <div class="w-px h-3 bg-primary/30"></div>
                    <div class="flex items-center flex-wrap gap-1 text-[10px]">
                      <!-- <TagIcon class="size-3 text-muted-foreground/70" /> -->
                      <span v-for="(tag, index) in (post.data.tags || []).slice(0, 3)" :key="index"
                        class="px-2 py-0.5 bg-primary/10 border border-primary/20 rounded-full text-[10px] text-primary/80">
                        {{ tag }}
                      </span>
                      <span v-if="(post.data.tags || []).length > 3" class="text-[10px]">...</span>
                    </div>
                  </template>
                </div>
              </div>
            </div>
            <!-- Actions -->
            <div class="absolute right-4 top-1/2 -translate-y-1/2 hidden group-hover:flex items-center gap-2 p-1 z-20">
              <div
                class="p-1.5 hover:bg-primary/10 rounded-md cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                @click.stop="previewPost(post)" title="预览">
                <EyeIcon class="size-3" />
              </div>
              <div
                class="p-1.5 hover:bg-primary/10 rounded-md cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                @click.stop="editPost(post)" title="编辑">
                <PencilIcon class="size-3" />
              </div>
              <div
                class="p-1.5 hover:bg-destructive/10 hover:text-destructive rounded-md cursor-pointer text-muted-foreground transition-colors"
                @click.stop="deleteSinglePost(post)" title="删除">
                <TrashIcon class="size-3" />
              </div>
            </div>
          </div>

          <!-- Feature Image -->
          <div v-if="post.data.feature"
            class="w-[100px] hidden sm:block relative overflow-hidden rounded-r-xl transition-opacity duration-200 group-hover:opacity-0">
            <img :src="getFeatureUrl(post.data.feature)" class="absolute inset-0 w-full h-full object-cover" />
          </div>

          <!-- Status Badges -->
          <div class="absolute top-0 right-0 flex pointer-events-none">
            <div v-if="post.data.hideInList"
              class="px-2 py-0.5 text-[10px] font-bold bg-foreground text-background rounded-bl-lg z-10 shadow-sm">HIDE
            </div>
            <div v-if="post.data.isTop"
              class="px-2 py-0.5 text-[10px] font-bold bg-yellow-400 text-yellow-900 rounded-bl-lg ml-[-4px] z-10 shadow-sm">
              TOP
            </div>
          </div>
        </div>
      </div>

    </div>

    <!-- Pagination -->
    <div class="h-12 py-3 px-4 border-t border-border flex justify-center bg-background" v-if="totalPages > 1">
      <Pagination :total="postList.length" :items-per-page="PAGE_SIZE" :page="currentPage" :sibling-count="2">
        <PaginationContent>
          <PaginationItem>
            <PaginationPrevious @click="currentPage > 1 && handlePageChanged(currentPage - 1)"
              :class="{ 'pointer-events-none opacity-50': currentPage === 1 }" />
          </PaginationItem>

          <template v-for="page in visiblePages" :key="page">
            <PaginationItem v-if="page === -1">
              <PaginationEllipsis />
            </PaginationItem>
            <PaginationLink v-else :value="page" :isActive="currentPage === page" @click="handlePageChanged(page)">
              {{ page }}
            </PaginationLink>
          </template>

          <PaginationItem>
            <PaginationNext @click="currentPage < totalPages && handlePageChanged(currentPage + 1)"
              :class="{ 'pointer-events-none opacity-50': currentPage === totalPages }" />
          </PaginationItem>
        </PaginationContent>
      </Pagination>
    </div>

    <ArticleUpdate v-if="articleUpdateVisible" :visible="articleUpdateVisible" :articleFileName="currentArticleFileName"
      @close="close" @fetchData="reloadSite"></ArticleUpdate>

    <DeleteConfirmDialog v-model:open="deleteModalVisible" :confirm-text="t('common.delete')"
      @confirm="confirmDelete" />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { useImageUrl } from '@/composables/useImageUrl'
import ArticleUpdate from './ArticleUpdate.vue'
import DeleteConfirmDialog from '@/components/ConfirmDialog/DeleteConfirmDialog.vue'
import dayjs from 'dayjs'
import { toast } from 'vue-sonner'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination'
import { PAGINATION } from '@/constants/editor'

import type { IPost } from '@/interfaces/post'
import {
  TrashIcon,
  MagnifyingGlassIcon,
  PlusIcon,
  CalendarIcon,
  TagIcon,
  EyeIcon,
  PencilIcon,
  FolderIcon,
} from '@heroicons/vue/24/outline'
import { EventsEmit, EventsOnce, EventsOff, BrowserOpenURL } from 'wailsjs/runtime'

const { t } = useI18n()
const siteStore = useSiteStore()
const { getImageUrl } = useImageUrl()

const articleUpdateVisible = ref<boolean>(false)
const currentArticleFileName = ref<string>('')
const selectedPost = ref<IPost[]>([])
const keyword = ref<string>('')
const searchInputVisible = ref<boolean>(false)
const currentPage = ref<number>(1)
const PAGE_SIZE = PAGINATION.DEFAULT_PAGE_SIZE
const searchInputRef = ref()
const deleteModalVisible = ref(false)

const listVersion = ref(Date.now())

const showSearchInput = async () => {
  searchInputVisible.value = true
  await nextTick()
}

watch(() => siteStore.posts, () => {
  listVersion.value = Date.now()
})

const getFeatureUrl = (path: string) => {
  if (!path) return ''
  if (path.startsWith('http') || path.startsWith('data:')) return path

  let fullPath = path
  if (path.startsWith('/post-images/')) {
    fullPath = `${siteStore.site.appDir}${path}`
  }

  return `${getImageUrl(fullPath)}&t=${listVersion.value}`
}

const postList = computed(() => {
  const search = keyword.value.toLowerCase().trim()
  let posts = []

  if (!search) {
    posts = [...siteStore.posts]
  } else {
    posts = siteStore.posts.filter((post: IPost) =>
      post.data.title.toLowerCase().includes(search)
    )
  }

  return posts.sort((a, b) => {
    const aTop = a.data.isTop ? 1 : 0
    const bTop = b.data.isTop ? 1 : 0
    if (aTop !== bTop) {
      return bTop - aTop
    }
    return dayjs(b.data.date).valueOf() - dayjs(a.data.date).valueOf()
  })
})

const totalPages = computed(() => Math.ceil(postList.value.length / PAGE_SIZE))

const currentPostList = computed(() => {
  const start = (currentPage.value - 1) * PAGE_SIZE
  const end = currentPage.value * PAGE_SIZE
  return postList.value.slice(start, end)
})

const visiblePages = computed(() => {
  const total = totalPages.value
  const current = currentPage.value
  const delta = 2
  const range = []
  const rangeWithDots = []
  let l

  range.push(1)
  for (let i = current - delta; i <= current + delta; i++) {
    if (i < total && i > 1) {
      range.push(i)
    }
  }
  range.push(total)

  for (let i of range) {
    if (l) {
      if (i - l === 2) {
        rangeWithDots.push(l + 1)
      } else if (i - l !== 1) {
        rangeWithDots.push(-1)
      }
    }
    rangeWithDots.push(i)
    l = i
  }
  return rangeWithDots
})

const handleSearchInputBlur = () => {
  if (!keyword.value) {
    searchInputVisible.value = false
  }
}

const onSelectChange = (post: IPost) => {
  const foundIndex = selectedPost.value.findIndex((item) => item.fileName === post.fileName)
  if (foundIndex !== -1) {
    selectedPost.value.splice(foundIndex, 1)
  } else {
    selectedPost.value.push(post)
  }
}

const handlePageChanged = (page: number) => {
  currentPage.value = page
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

const reloadSite = () => {
  EventsEmit('app-site-reload')
}

const close = () => {
  articleUpdateVisible.value = false
  currentArticleFileName.value = ''
}

const newArticle = () => {
  articleUpdateVisible.value = true
  currentArticleFileName.value = ''
}

const editPost = (post: IPost) => {
  articleUpdateVisible.value = true
  currentArticleFileName.value = post.fileName
}

const previewPost = (post: IPost) => {
  const { postPath } = siteStore.themeConfig
  const url = `http://localhost:3367/${postPath}/${post.fileName}/`
  BrowserOpenURL(url)
}

const deleteSinglePost = (post: IPost) => {
  selectedPost.value = [post]
  deleteModalVisible.value = true
}

const confirmDelete = () => {
  deleteModalVisible.value = false
  const postsToDelete = JSON.parse(JSON.stringify(selectedPost.value))
  EventsEmit('app-post-list-delete', postsToDelete)
  EventsOnce('app-post-list-deleted', (data: { success: boolean; posts?: IPost[] }) => {
    if (data.success && data.posts) {
      siteStore.posts = data.posts
      toast.success(t('article.delete'))
      selectedPost.value = []
    }
  })
}

watch(keyword, () => {
  currentPage.value = 1
})

onMounted(() => {
  EventsOff('app-post-list-deleted')
})

onUnmounted(() => {
  EventsOff('app-post-list-deleted')
})
</script>
