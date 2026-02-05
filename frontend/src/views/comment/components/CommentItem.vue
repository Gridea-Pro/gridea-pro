<template>
    <div
        class="rounded-xl relative transition-all duration-200 bg-primary/2 border border-primary/10 hover:bg-primary/5 hover:shadow-xs hover:-translate-y-0.5 p-5">
        <!-- Top: User Info -->
        <div class="flex justify-between items-start mb-3">
            <div class="flex items-center gap-3">
                <a :href="comment.url || 'javascript:void(0)'" class="flex items-center gap-3 group"
                    :class="{ 'cursor-default': !comment.url, 'cursor-pointer': comment.url }"
                    @click="openLink($event, comment.url)">
                    <img :src="comment.avatar || generateAvatar(comment.nickname || comment.id)"
                        class="w-10 h-10 rounded-full bg-secondary object-cover border border-border" alt="avatar" />
                    <div class="flex items-center gap-2 flex-wrap">
                        <span class="font-semibold text-foreground"
                            :class="{ 'group-hover:text-primary transition-colors': comment.url }">
                            {{ comment.nickname }}
                        </span>
                    </div>
                </a>

                <span class="text-xs text-gray-400">{{ formatTime(comment.createdAt) }}</span>
            </div>

            <div class="flex items-center gap-1">
                <button
                    class="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
                    :title="t('comment.reply')" @click="toggleReplyForm">
                    <ChatBubbleOvalLeftEllipsisIcon class="size-3" />
                </button>
                <button
                    class="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
                    :title="t('comment.delete')" @click="$emit('delete', comment)">
                    <TrashIcon class="size-3" />
                </button>
            </div>
        </div>

        <!-- Middle: Content -->
        <div class="relative group/content">
            <MarkdownContent ref="contentRef" :class="{ 'max-h-[240px] overflow-hidden': !isExpanded && isOverflowing }"
                @click="handleContentClick">
                <span v-if="comment.parentNick" class="text-muted-foreground me-1">
                    {{ t('comment.reply') }} @{{ comment.parentNick }} :
                </span>
                <span v-html="comment.content"></span>
            </MarkdownContent>

            <!-- Expand Button Overlay -->
            <div v-if="isOverflowing && !isExpanded"
                class="absolute bottom-0 w-full h-16 bg-gradient-to-t from-background via-background/60 to-transparent flex justify-center items-end pb-0 cursor-pointer pl-[52px] rounded-b-xl"
                @click.stop="toggleExpand">
                <div class="text-primary/80 hover:text-primary transition-colors hover:scale-110 animate-bounce">
                    <ChevronDoubleDownIcon class="size-4" />
                </div>
            </div>

            <!-- Collapse Button -->
            <div v-if="isExpanded && isOverflowing"
                class="w-full flex justify-center items-center -mt-2 mb-2 cursor-pointer" @click.stop="toggleExpand">
                <div class="text-primary/80 hover:text-primary transition-colors hover:scale-110">
                    <ChevronDoubleUpIcon class="size-4" />
                </div>
            </div>
        </div>

        <!-- Reply Form -->
        <div v-if="isReplying"
            class="ml-[52px] mb-4 bg-secondary/30 p-4 rounded-xl border border-border animate-fade-in-down">
            <textarea v-model="replyContent" :placeholder="`${t('comment.reply')} @${comment.nickname}：`" rows="3"
                class="w-full px-3 py-2 bg-background border border-border rounded-lg text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-primary resize-y min-h-[80px]"></textarea>
            <div class="flex justify-end gap-2 mt-3">
                <Button variant="outline"
                    class="text-primary/80 text-xs px-4 h-8 border border-primary/10 rounded-full cursor-pointer hover:bg-primary/10 hover:text-primary transition-colors"
                    @click="cancelReply">{{ t('comment.cancel') }}</Button>
                <Button variant="default" class="text-xs px-4 h-8 border border-primary/10 rounded-full cursor-pointer"
                    @click="submitReply">
                    {{ t('comment.reply') }}
                </Button>
            </div>
        </div>

        <!-- Bottom: Article Reference -->
        <div
            class="ml-[52px] bg-secondary/50 rounded-lg px-3 py-2 text-xs text-muted-foreground flex items-center gap-2 border border-border/50 hover:text-primary">
            <DocumentTextIcon class="size-4 flex-shrink-0" />
            <a :href="comment.articleUrl || comment.articleId || 'javascript:void(0)'"
                class="truncate transition-colors cursor-pointer"
                @click="openLink($event, comment.articleUrl || comment.articleId)">
                <span>{{ comment.articleTitle }}</span>
            </a>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Comment } from '@/types/comment'
import { generateAvatar } from '@/utils/avatarGenerator'
import { parseDate } from '@/utils/date'
import dayjs from 'dayjs'
import {
    ChatBubbleOvalLeftEllipsisIcon,
    TrashIcon,
    DocumentTextIcon,
    ChevronDoubleDownIcon,
    ChevronDoubleUpIcon
} from '@heroicons/vue/24/outline'
import { Button } from '@/components/ui/button'
import MarkdownContent from './MarkdownContent.vue'
import { useContentOverflow } from '@/composables/useContentOverflow'
import { BrowserOpenURL, Environment } from 'wailsjs/runtime'
import { useSiteStore } from '@/stores/site' // For openLink domain logic if needed

const props = defineProps<{
    comment: Comment
    previewPort?: number
}>()

const emit = defineEmits<{
    (e: 'reply', parentId: string, content: string, articleId: string): void
    (e: 'delete', comment: Comment): void
    (e: 'previewImage', src: string): void
}>()

const { t } = useI18n()
const siteStore = useSiteStore()

// Overflow logic
const { isExpanded, isOverflowing, contentRef, checkOverflow, toggleExpand } = useContentOverflow()

// Reply logic
const isReplying = ref(false)
const replyContent = ref('')

const toggleReplyForm = () => {
    isReplying.value = !isReplying.value
    if (!isReplying.value) {
        replyContent.value = ''
    }
}

const cancelReply = () => {
    isReplying.value = false
    replyContent.value = ''
}

const submitReply = () => {
    if (!replyContent.value.trim()) return
    emit('reply', props.comment.id, replyContent.value, props.comment.articleId)
    // Close form handled by parent usually? Or define here.
    // Ideally, wait for success. But for simple UI, we can clear it or let parent tell us.
    // Here we assume optimism or parent handles refresh.
    isReplying.value = false
    replyContent.value = ''
}

// Helper functions (could be utilities, but kept here for now)
const formatTime = (time: string) => {
    if (!time) return ''
    const date = parseDate(time)
    return dayjs(date).format('YYYY-MM-DD HH:mm')
}

const openLink = (e: Event, url?: string) => {
    if (!url || url === 'javascript:void(0)') return
    e.preventDefault()

    let targetUrl = url
    if (targetUrl.startsWith('/')) {
        const domain = siteStore.setting.domain || siteStore.setting.cname || ''
        if (domain) {
            const baseUrl = domain.startsWith('http') ? domain : `https://${domain}`
            targetUrl = `${baseUrl.replace(/\/$/, '')}${targetUrl}`
        } else {
            // Use fallback port passed from parent or default
            const port = props.previewPort || 6060
            targetUrl = `http://localhost:${port}${targetUrl}`
        }
    }

    if (targetUrl.startsWith('http')) {
        BrowserOpenURL(targetUrl)
    }
}

const handleContentClick = (e: MouseEvent) => {
    const target = e.target as HTMLElement
    if (target.tagName === 'IMG') {
        const img = target as HTMLImageElement
        emit('previewImage', img.src)
    }
}

onMounted(() => {
    checkOverflow()
})
</script>
