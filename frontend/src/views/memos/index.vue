<template>
    <div class="h-full flex flex-col bg-background">
        <!-- Header Tools -->
        <div
class="flex-shrink-0 flex justify-between items-center px-4 h-12 bg-background sticky top-0 z-50 border-b border-border backdrop-blur-sm bg-opacity-90 select-none"
            style="--wails-draggable: drag">
            <div class="flex-1"></div>
            <div class="flex items-center gap-2" style="--wails-draggable: no-drag">
                <SearchInput v-model="memoStore.keyword" :placeholder="t('common.search')" />

                <span class="text-xs text-muted-foreground ml-2">
                    {{ t('memo.totalMemos') }}: {{ memoStore.totalMemos }}
                </span>
            </div>
        </div>

        <!-- Main Content -->
        <div class="flex-1 flex overflow-hidden">
            <!-- Left Sidebar -->
            <aside class="w-68 flex-shrink-0 border-r border-border flex flex-col overflow-hidden">
                <!-- Top Section: Stats & Graph -->
                <div class="p-4 space-y-6 flex-shrink-0">
                    <ContributionGraph
:data="memoStore.heatmapData" :label="t('memo.heatmap')"
                        @day-click="handleDayClick" />

                    <!-- Stats -->
                    <div class="space-y-1">
                        <div
class="flex items-center justify-between text-sm cursor-pointer p-2 rounded-md transition-colors hover:bg-primary/15"
                            :class="[
                                memoStore.timeFilter === 'all'
                                    ? 'bg-primary/10 text-primary font-medium'
                                    : 'text-muted-foreground hover:text-foreground'
                            ]" @click="memoStore.setTimeFilter('all')">
                            <span>{{ t('memo.totalMemos') }}</span>
                            <span class="text-xs font-medium opacity-80">{{ memoStore.totalMemos }}</span>
                        </div>
                        <div
class="flex items-center justify-between text-sm cursor-pointer p-2 rounded-md transition-colors hover:bg-primary/15"
                            :class="[
                                memoStore.timeFilter === 'today'
                                    ? 'bg-primary/10 text-primary font-medium'
                                    : 'text-muted-foreground hover:text-foreground'
                            ]" @click="memoStore.setTimeFilter('today')">
                            <span>{{ t('memo.todayMemos') }}</span>
                            <span class="text-xs font-medium opacity-80">{{ memoStore.todayMemos }}</span>
                        </div>
                        <div
class="flex items-center justify-between text-sm cursor-pointer p-2 rounded-md transition-colors hover:bg-primary/15"
                            :class="[
                                memoStore.timeFilter === 'month'
                                    ? 'bg-primary/10 text-primary font-medium'
                                    : 'text-muted-foreground hover:text-foreground'
                            ]" @click="memoStore.setTimeFilter('month')">
                            <span>{{ t('memo.monthMemos') }}</span>
                            <span class="text-xs font-medium opacity-80">{{ memoStore.monthMemos }}</span>
                        </div>
                    </div>
                </div>

                <!-- Tags Section (Scrollable) -->
                <div
v-if="memoStore.tagStats.length > 0"
                    class="flex-1 overflow-y-auto p-4 pt-0 min-h-0 overscroll-none">
                    <div class="space-y-1">
                        <div
                            class="flex items-center justify-between px-2 py-2 text-xs text-muted-foreground sticky top-0 bg-background/95 backdrop-blur-sm z-10">
                            <span>{{ t('memo.tags') }}</span>
                            <span class="text-xs font-medium opacity-80">{{ memoStore.totalTags }}</span>
                        </div>
                        <div class="flex flex-wrap gap-1.5 px-2">
                            <button
                                class="inline-flex items-center px-2.5 py-1 text-[11px] rounded-full transition-all duration-200 cursor-pointer border"
                                :class="[
                                    memoStore.selectedTag === null
                                        ? 'bg-primary/10 text-primary border-primary/30'
                                        : 'bg-muted/30 text-muted-foreground border-transparent hover:bg-muted/50 hover:text-foreground'
                                ]" @click="memoStore.setSelectedTag(null)">
                                {{ t('memo.allTags') }}
                            </button>
                            <button
v-for="tag in memoStore.tagStats" :key="tag.name"
                                class="inline-flex items-center px-2.5 py-1 text-[11px] rounded-full transition-all duration-200 cursor-pointer border"
                                :class="[
                                    memoStore.selectedTag === tag.name
                                        ? 'bg-primary/10 text-primary border-primary/30'
                                        : 'text-muted-foreground border-transparent hover:bg-muted/50 hover:text-foreground'
                                ]" @click="memoStore.setSelectedTag(tag.name)">
                                #{{ tag.name }} <span class="ml-1 opacity-60">{{ tag.count }}</span>
                            </button>
                        </div>
                    </div>
                </div>
            </aside>

            <!-- Right Content -->
            <main class="flex-1 flex flex-col min-w-0 bg-background">
                <!-- Input Fixed Top -->
                <div class="p-6 pb-0 z-10 bg-background">
                    <div class="max-w-2xl mx-auto mb-6">
                        <MemoInput
ref="memoInputRef" :placeholder="t('memo.inputPlaceholder')"
                            @submit="handleSaveMemo" />
                    </div>
                </div>

                <!-- Scrollable List -->
                <div class="flex-1 overflow-y-auto p-6 pt-0">
                    <div class="max-w-2xl mx-auto">

                        <!-- Date Header -->
                        <div
v-if="memoStore.selectedDate"
                            class="mb-6 pb-4 border-b border-border flex items-center justify-between group">
                            <div class="flex items-center gap-2">
                                <CalendarIcon class="size-4 text-primary" />
                                <span class="text-sm font-medium text-muted-foreground">{{ memoStore.selectedDate
                                    }}</span>
                            </div>
                            <div class="text-xs text-muted-foreground">
                                {{ memoStore.filteredMemos.length }} Memos
                            </div>
                        </div>

                        <!-- Tag Header -->
                        <div
v-if="memoStore.selectedTag"
                            class="mb-6 pb-4 border-b border-border flex items-center justify-between group">
                            <div class="flex items-center gap-2">
                                <template v-if="tagInputVisible">
                                    <span class="text-sm font-medium text-primary">#</span>
                                    <input
ref="tagInputRef"
                                        v-model="editingTagName"
                                        class="text-sm font-medium bg-transparent border-b border-primary focus:outline-none min-w-[100px]" @blur="handleRenameTag" @keydown.enter="handleRenameTag" />
                                </template>
                                <template v-else>
                                    <span class="text-sm font-medium text-muted-foreground">
                                        <span class="text-primary"># </span>
                                        {{ memoStore.selectedTag }}
                                    </span>
                                    <div class="hidden group-hover:flex items-center gap-1 ml-2">
                                        <button
                                            class="p-1 hover:bg-primary/10 rounded-md text-muted-foreground hover:text-primary transition-colors cursor-pointer"
                                            :title="t('nativeMenu.edit')" @click="handleEditTag">
                                            <PencilIcon class="size-3" />
                                        </button>
                                        <button
                                            class="p-1 hover:bg-destructive/10 rounded-md text-muted-foreground hover:text-destructive transition-colors cursor-pointer"
                                            :title="t('common.delete')" @click="handleDeleteTag">
                                            <TrashIcon class="size-3" />
                                        </button>
                                    </div>
                                </template>
                            </div>
                            <div class="text-xs text-muted-foreground">
                                {{ memoStore.filteredMemos.length }} Memos
                            </div>
                        </div>

                        <MemoList
:memos="memoStore.filteredMemos" :empty-text="t('memo.empty')"
                            @update="handleUpdateMemo" @delete="handleDeleteMemo"
                            @tag-click="memoStore.setSelectedTag" />
                    </div>
                </div>
            </main>
        </div>

        <!-- Delete Confirm -->
        <DeleteConfirmDialog v-model:open="deleteDialogVisible" @confirm="confirmDelete" />
    </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMemoStore } from '@/stores/memo'
import type { IMemo } from '@/interfaces/memo'
import { toast } from '@/helpers/toast'
import ContributionGraph from './components/ContributionGraph.vue'
import MemoInput from './components/MemoInput.vue'
import MemoList from './components/MemoList.vue'
import DeleteConfirmDialog from '@/components/ConfirmDialog/DeleteConfirmDialog.vue'
import SearchInput from '@/components/Base/SearchInput.vue'
import { MagnifyingGlassIcon, PencilIcon, TrashIcon, CalendarIcon } from '@heroicons/vue/24/outline'

const { t } = useI18n()
const memoStore = useMemoStore()

const memoInputRef = ref<InstanceType<typeof MemoInput> | null>(null)

const handleDayClick = (date: string) => {
    memoStore.setSelectedDate(date)
}

// Tag Management

// Tag Management
const tagInputVisible = ref(false)
const tagInputRef = ref<HTMLInputElement | null>(null)
const editingTagName = ref('')

const handleEditTag = async () => {
    if (!memoStore.selectedTag) return
    editingTagName.value = memoStore.selectedTag
    tagInputVisible.value = true
    await nextTick()
    tagInputRef.value?.focus()
}

const handleRenameTag = async () => {
    if (!memoStore.selectedTag || !editingTagName.value || editingTagName.value === memoStore.selectedTag) {
        tagInputVisible.value = false
        return
    }

    try {
        await memoStore.renameTag(memoStore.selectedTag, editingTagName.value)
        toast.success(t('memo.saved'))
        memoStore.setSelectedTag(editingTagName.value)
    } catch {
        toast.error('重命名失败')
    } finally {
        tagInputVisible.value = false
    }
}

const handleDeleteTag = async () => {
    if (!memoStore.selectedTag) return
    tagToDelete.value = memoStore.selectedTag
    deleteDialogVisible.value = true
}

// Delete state
const deleteDialogVisible = ref(false)
const memoToDelete = ref<string | null>(null)
const tagToDelete = ref<string | null>(null)

async function handleSaveMemo(content: string) {
    // Creating new memo
    try {
        await memoStore.saveMemo(content)
        toast.success(t('memo.published'))
    } catch {
        toast.error('保存失败')
    }
}

async function handleUpdateMemo(memo: IMemo) {
    try {
        await memoStore.updateMemo(memo)
        toast.success(t('memo.saved'))
    } catch {
        toast.error('更新失败')
    }
}

function handleDeleteMemo(id: string) {
    memoToDelete.value = id
    deleteDialogVisible.value = true
}

async function confirmDelete() {
    if (memoToDelete.value) {
        try {
            await memoStore.deleteMemo(memoToDelete.value)
            toast.success(t('memo.deleted'))
        } catch {
            toast.error('删除失败')
        } finally {
            deleteDialogVisible.value = false
            memoToDelete.value = null
        }
    } else if (tagToDelete.value) {
        try {
            await memoStore.deleteTag(tagToDelete.value)
            toast.success(t('memo.deleted'))
            memoStore.setSelectedTag(null)
        } catch {
            toast.error('删除失败')
        } finally {
            deleteDialogVisible.value = false
            tagToDelete.value = null
        }
    }
}



onMounted(() => {
    memoStore.fetchMemos()
})

onUnmounted(() => {
    memoStore.cleanup()
})
</script>
