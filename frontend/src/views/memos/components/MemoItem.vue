<template>
    <div
        class="memo-item group p-4 rounded-xl border border-border/50 bg-card/50 hover:bg-primary/5 hover:shadow-xs hover:-translate-y-0.5 transition-all duration-200">
        <!-- Viewing Mode -->
        <div v-if="!isEditing">
            <div
class="memo-content text-sm text-foreground leading-7 tracking-wider mb-3 whitespace-pre-wrap break-words max-h-[160px] overflow-y-auto"
                v-html="highlightedContent" />

            <!-- 底部：时间和操作 -->
            <div class="flex items-center justify-between mt-3">
                <div class="flex items-center gap-3 text-xs text-muted-foreground">
                    <span class="flex items-center gap-1 opacity-80">
                        {{ formatTime(memo.createdAt) }}
                    </span>
                </div>

                <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    <button
                        class="p-1.5 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-md transition-colors cursor-pointer"
                        :title="t('common.copy')" @click="handleCopy">
                        <ClipboardDocumentIcon class="size-3.5" />
                    </button>
                    <button
                        class="p-1.5 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-md transition-colors cursor-pointer"
                        :title="t('common.edit')" @click="startEdit">
                        <PencilIcon class="size-3.5" />
                    </button>
                    <button
                        class="p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-md transition-colors cursor-pointer"
                        :title="t('common.delete')" @click="$emit('delete', memo.id)">
                        <TrashIcon class="size-3.5" />
                    </button>
                </div>
            </div>
        </div>

        <!-- Editing Mode -->
        <div v-else>
            <MemoInput
ref="editorRef" :placeholder="t('memo.inputPlaceholder')" :submit-text="t('common.save')"
                :is-editing="true" @submit="handleSave" @cancel="cancelEdit" />
        </div>
    </div>
</template>

<script lang="ts" setup>
import { computed, ref, nextTick, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ClipboardDocumentIcon, PencilIcon, TrashIcon } from '@heroicons/vue/24/outline'
import type { IMemo } from '@/interfaces/memo'
import { highlightTags } from '@/utils/tagParser'
import { toast } from '@/helpers/toast'
import MemoInput from './MemoInput.vue'

interface Props {
    memo: IMemo
}

const props = defineProps<Props>()

const emit = defineEmits<{
    delete: [id: string]
    tagClick: [tag: string]
    update: [memo: IMemo]
}>()

const { t } = useI18n()
const isEditing = ref(false)
const editorRef = ref<InstanceType<typeof MemoInput> | null>(null)

// Auto-update every 5 seconds for "just now" / "minutes ago"
const now = ref(new Date())
let timer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
    timer = setInterval(() => {
        now.value = new Date()
    }, 5000)
})

onUnmounted(() => {
    if (timer) clearInterval(timer)
})

const highlightedContent = computed(() => {
    return highlightTags(props.memo.content || '', 'text-primary font-medium cursor-pointer')
})

const startEdit = async () => {
    isEditing.value = true
    await nextTick()
    if (editorRef.value) {
        editorRef.value.setContent(props.memo.content)
    }
}

const cancelEdit = () => {
    isEditing.value = false
}

const handleSave = (content: string) => {
    const updatedMemo = { ...props.memo, content }
    emit('update', updatedMemo)
    isEditing.value = false
}

function formatTime(timestamp: string): string {
    const date = new Date(timestamp)
    const current = now.value
    const diff = current.getTime() - date.getTime()

    // 一分钟内
    if (diff < 60 * 1000) {
        return t('memo.justNow')
    }

    // 一小时内
    if (diff < 60 * 60 * 1000) {
        return t('memo.minutesAgo', { min: Math.floor(diff / (60 * 1000)) })
    }

    // 24小时内
    if (diff < 24 * 60 * 60 * 1000 && date.getDate() === current.getDate()) {
        return t('memo.hoursAgo', { hour: Math.floor(diff / (60 * 60 * 1000)) })
    }

    // 昨天
    const yesterday = new Date(current)
    yesterday.setDate(yesterday.getDate() - 1)
    if (date.toDateString() === yesterday.toDateString()) {
        return t('memo.yesterday', { time: date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' }) })
    }

    // 今年内
    if (date.getFullYear() === current.getFullYear()) {
        return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
    }

    // 其他
    return date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

async function handleCopy() {
    try {
        await navigator.clipboard.writeText(props.memo.content)
        toast.success(t('common.copied'))
    } catch {
        toast.error(t('common.copyFailed'))
    }
}
</script>
