<template>
  <div class="h-full flex flex-col bg-background">
    <!-- Header -->
    <div
      class="flex-shrink-0 flex justify-between items-center px-4 h-12 bg-background sticky top-0 z-50 border-b border-border backdrop-blur-sm bg-opacity-90 select-none"
      style="--wails-draggable: drag">
      <div class="flex-1"></div>
      <div
        class="flex items-center justify-center w-8 h-8 rounded-full hover:bg-primary/10 cursor-pointer transition-colors text-muted-foreground hover:text-foreground"
        @click="settingsVisible = true" :title="t('settings.comment.title')" style="--wails-draggable: no-drag">
        <Cog6ToothIcon class="size-4" />
      </div>
    </div>

    <!-- Comment List -->
    <div class="flex-1 w-full overflow-y-auto px-4 py-6">
      <div v-if="loading" class="flex justify-center py-10">
        <svg class="animate-spin h-8 w-8 text-primary" xmlns="http://www.w3.org/2000/svg" fill="none"
          viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
          </path>
        </svg>
      </div>
      <div v-else-if="comments.length === 0" class="text-center text-muted-foreground py-10 flex flex-col items-center">
        <div class="text-sm">{{ t('comment.empty') }}</div>
      </div>
      <div v-else class="flex flex-col gap-6">
        <CommentItem v-for="comment in comments" :key="comment.id" :comment="comment" :preview-port="previewPort"
          @reply="handleReply" @delete="confirmDelete" @preview-image="handlePreviewImage" />
      </div>
    </div>

    <!-- Pagination -->
    <div class="h-12 py-3 px-4 border-t border-border flex justify-center bg-background" v-if="total > pageSize">
      <Pagination :total="total" :items-per-page="pageSize" :page="page" :sibling-count="2">
        <PaginationContent>
          <PaginationItem>
            <PaginationPrevious @click="page > 1 && handlePageChange(page - 1)"
              :class="{ 'pointer-events-none opacity-50': page === 1 }" />
          </PaginationItem>

          <template v-for="p in visiblePages" :key="p">
            <PaginationItem v-if="p === -1">
              <PaginationEllipsis />
            </PaginationItem>
            <PaginationLink v-else :value="p" :isActive="page === p" @click="handlePageChange(p)">
              {{ p }}
            </PaginationLink>
          </template>

          <PaginationItem>
            <PaginationNext @click="page < totalPages && handlePageChange(page + 1)"
              :class="{ 'pointer-events-none opacity-50': page === totalPages }" />
          </PaginationItem>
        </PaginationContent>
      </Pagination>
    </div>

    <!-- Image Preview Modal -->
    <ImagePreview v-model="previewImage" />

    <!-- Settings Drawer -->
    <CommentSettings :open="settingsVisible" @update:open="settingsVisible = $event" />

    <DeleteConfirmDialog v-model:open="deleteModalVisible" @confirm="handleDelete" />
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { toast } from '@/helpers/toast'
import { Cog6ToothIcon } from '@heroicons/vue/24/outline'
import CommentSettings from './components/CommentSettings.vue'
import CommentItem from './components/CommentItem.vue'
import ImagePreview from './components/ImagePreview.vue'
import DeleteConfirmDialog from '@/components/ConfirmDialog/DeleteConfirmDialog.vue'
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination'
import { useCommentStore } from '@/stores/comment'
import { Environment } from 'wailsjs/runtime'
import { parseEmoji } from '@/utils/emoji'
import markdownIt from '@/helpers/markdown'
import type { Comment } from '@/types/comment'

const { t } = useI18n()
const commentStore = useCommentStore()
const settingsVisible = ref(false)
const loading = ref(false)
const deleteModalVisible = ref(false)
const commentToDelete = ref<Comment | null>(null)
const previewImage = ref<string | null>(null)
const previewPort = ref(6060)

// Helper to unescape HTML entities
const decodeHTML = (html: string) => {
  const txt = document.createElement('textarea')
  txt.innerHTML = html
  return txt.value
}

const comments = computed(() => {
  return commentStore.comments.map(c => {
    let raw = decodeHTML(c.content)
    raw = raw.replace(/^>([^\s])/gm, '> $1')
    return {
      ...c,
      content: markdownIt.render(parseEmoji(raw))
    }
  })
})

const page = computed(() => commentStore.page)
const pageSize = computed(() => commentStore.pageSize)
const total = computed(() => commentStore.total)
const totalPages = computed(() => commentStore.totalPages)

const visiblePages = computed(() => {
  const total = totalPages.value
  const current = page.value
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
  // Filter duplicates
  return [...new Set(rangeWithDots)]
})


const handleReply = async (parentId: string, content: string, articleId: string) => {
  if (!content.trim()) {
    toast.warning(t('comment.emptyContent'))
    return
  }

  if (!commentStore.replyComment) {
    toast.error(t('comment.notAvailable'))
    return
  }

  try {
    await commentStore.replyComment(parentId, content, articleId)
    // Find nickname for success message if possible, though parentId is ID.
    // We can just say "Success".
    toast.success(t('comment.replySuccess', { nickname: '' }))
  } catch (e) {
    console.error(e)
    toast.error(t('comment.replyFail'))
  }
}

const confirmDelete = (comment: Comment) => {
  commentToDelete.value = comment
  deleteModalVisible.value = true
}

const handleDelete = async () => {
  if (commentToDelete.value) {
    if (commentStore.deleteComment) {
      try {
        await commentStore.deleteComment(commentToDelete.value.id)
        toast.success(t('comment.deleted'))
      } catch (e) {
        toast.error(t('comment.deleteFail'))
      }
    } else {
      toast.error(t('comment.notAvailable'))
    }
  }
  deleteModalVisible.value = false
  commentToDelete.value = null
}

const handlePreviewImage = (src: string) => {
  previewImage.value = src
}

const handlePageChange = async (newPage: number) => {
  if (newPage < 1 || newPage > totalPages.value) return
  loading.value = true
  await commentStore.fetchComments(newPage, pageSize.value)
  const list = document.querySelector('.overflow-y-auto')
  if (list) list.scrollTop = 0
  loading.value = false
}

let refreshInterval: any = null

onMounted(async () => {
  try {
    const env = await Environment()
    if (env.buildType !== 'production') {
      previewPort.value = 3367
    } else {
      previewPort.value = 6060
    }
    console.log('[Comment] Detected environment:', env.buildType, 'Using port:', previewPort.value)
  } catch (e) {
    console.warn('[Comment] Failed to get environment info, using default port 6060', e)
    previewPort.value = 6060
  }

  loading.value = true
  await commentStore.fetchComments(page.value, pageSize.value)
  commentStore.markAllAsRead()
  loading.value = false

  refreshInterval = setInterval(async () => {
    if (page.value === 1) {
      await commentStore.fetchComments(1, pageSize.value)
      commentStore.markAllAsRead()
    }
  }, 5000)
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>