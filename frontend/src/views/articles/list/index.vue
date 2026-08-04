<template>
    <div class="h-full flex flex-col bg-background text-foreground">
        <!-- Header Tools -->
        <ListHeader
:keyword="keyword" :selected-count="selectedPost.length" @update:keyword="keyword = $event"
            @delete-selected="deleteModalVisible = true" @new-article="$emit('newArticle')" />

        <!-- Main Content -->
        <div class="flex-1 flex overflow-hidden">
            <!-- Left Sidebar -->
            <aside class="w-68 flex-shrink-0 border-r border-border flex flex-col overflow-hidden">
                <div class="p-4 space-y-6 flex-shrink-0">
                    <ContributionGraph
:data="heatmapData" :label="t('article.createAt')"
                        @day-click="setSelectedDate" />

                    <!-- Stats -->
                    <div class="space-y-1">
                        <div
class="flex items-center justify-between text-sm cursor-pointer p-2 rounded-md transition-colors hover:bg-primary/15"
                            :class="[
                                timeFilter === 'all'
                                    ? 'bg-primary/10 text-primary font-medium'
                                    : 'text-muted-foreground hover:text-foreground'
                            ]" @click="setTimeFilter('all')">
                            <span>{{ t('nav.article') }}</span>
                            <span class="text-xs font-medium opacity-80">{{ totalPosts }}</span>
                        </div>
                        <div
class="flex items-center justify-between text-sm cursor-pointer p-2 rounded-md transition-colors hover:bg-primary/15"
                            :class="[
                                timeFilter === 'today'
                                    ? 'bg-primary/10 text-primary font-medium'
                                    : 'text-muted-foreground hover:text-foreground'
                            ]" @click="setTimeFilter('today')">
                            <span>{{ t('article.todayArticles') }}</span>
                            <span class="text-xs font-medium opacity-80">{{ todayPosts }}</span>
                        </div>
                        <div
class="flex items-center justify-between text-sm cursor-pointer p-2 rounded-md transition-colors hover:bg-primary/15"
                            :class="[
                                timeFilter === 'month'
                                    ? 'bg-primary/10 text-primary font-medium'
                                    : 'text-muted-foreground hover:text-foreground'
                            ]" @click="setTimeFilter('month')">
                            <span>{{ t('article.monthArticles') }}</span>
                            <span class="text-xs font-medium opacity-80">{{ monthPosts }}</span>
                        </div>
                    </div>
                </div>

                <!-- Categories Section -->
                <div
v-if="categoryStats.length > 0"
                    class="flex-1 overflow-y-auto p-4 pt-0 min-h-0 overscroll-none">
                    <div class="space-y-1">
                        <div
                            class="flex items-center justify-between px-2 py-2 text-xs text-muted-foreground sticky top-0 bg-background/95 backdrop-blur-sm z-10">
                            <span>{{ t('nav.category') }}</span>
                        </div>
                        <div class="flex flex-wrap gap-1.5 px-2">
                            <button
class="inline-flex items-center px-2.5 py-1 text-[11px] rounded-full transition-all duration-200 cursor-pointer border"
                                :class="[
                                    selectedCategory === null
                                        ? 'bg-primary/10 text-primary border-primary/30'
                                        : 'bg-muted/30 text-muted-foreground border-transparent hover:bg-muted/50 hover:text-foreground'
                                ]" @click="setSelectedCategory(null)">
                                {{ t('article.allCategories') }}
                            </button>
                            <button
v-for="cat in categoryStats" :key="cat.name"
class="inline-flex items-center px-2.5 py-1 text-[11px] rounded-full transition-all duration-200 cursor-pointer border"
                                :class="[
                                    selectedCategory === cat.name
                                        ? 'bg-primary/10 text-primary border-primary/30'
                                        : 'text-muted-foreground border-transparent hover:bg-muted/50 hover:text-foreground'
                                ]" @click="setSelectedCategory(cat.name)">
                                {{ cat.name }} <span class="ml-1 opacity-60">{{ cat.count }}</span>
                            </button>
                        </div>
                    </div>
                </div>

                <!-- Tags Section -->
                <div
v-if="tagStats.length > 0"
                    class="flex-1 overflow-y-auto p-4 pt-0 min-h-0 overscroll-none">
                    <div class="space-y-1">
                        <div
                            class="flex items-center justify-between px-2 py-2 text-xs text-muted-foreground sticky top-0 bg-background/95 backdrop-blur-sm z-10">
                            <span>{{ t('nav.tag') }}</span>
                            <span class="text-xs font-medium opacity-80">{{ tagStats.length }}</span>
                        </div>
                        <div class="flex flex-wrap gap-1.5 px-2">
                            <button
class="inline-flex items-center px-2.5 py-1 text-[11px] rounded-full transition-all duration-200 cursor-pointer border"
                                :class="[
                                    selectedTag === null
                                        ? 'bg-primary/10 text-primary border-primary/30'
                                        : 'bg-muted/30 text-muted-foreground border-transparent hover:bg-muted/50 hover:text-foreground'
                                ]" @click="setSelectedTag(null)">
                                {{ t('article.allArticles') }}
                            </button>
                            <button
v-for="tag in tagStats" :key="tag.name"
class="inline-flex items-center px-2.5 py-1 text-[11px] rounded-full transition-all duration-200 cursor-pointer border"
                                :class="[
                                    selectedTag === tag.name
                                        ? 'bg-primary/10 text-primary border-primary/30'
                                        : 'text-muted-foreground border-transparent hover:bg-muted/50 hover:text-foreground'
                                ]" @click="setSelectedTag(tag.name)">
                                #{{ tag.name }} <span class="ml-1 opacity-60">{{ tag.count }}</span>
                            </button>
                        </div>
                    </div>
                </div>
            </aside>

            <!-- Right Content -->
            <main class="flex-1 flex flex-col min-w-0 bg-background">
                <!-- Active Filter Header -->
                <div v-if="selectedDate || selectedTag || selectedCategory || timeFilter !== 'all'"
                    class="px-4 pt-4 flex items-center gap-2 flex-wrap">
                    <div
v-if="selectedDate" class="flex items-center gap-1.5 px-2.5 py-1 bg-primary/10 text-primary rounded-full text-xs border border-primary/20">
                        <CalendarIcon class="size-3" />
                        {{ selectedDate }}
                        <button class="ml-1 hover:text-destructive cursor-pointer" @click="setSelectedDate(null)">
                            <XMarkIcon class="size-3" />
                        </button>
                    </div>
                    <div
v-if="selectedCategory" class="flex items-center gap-1.5 px-2.5 py-1 bg-primary/10 text-primary rounded-full text-xs border border-primary/20">
                        <FolderIcon class="size-3" />
                        {{ selectedCategory }}
                        <button class="ml-1 hover:text-destructive cursor-pointer" @click="setSelectedCategory(null)">
                            <XMarkIcon class="size-3" />
                        </button>
                    </div>
                    <div
v-if="selectedTag" class="flex items-center gap-1.5 px-2.5 py-1 bg-primary/10 text-primary rounded-full text-xs border border-primary/20">
                        <TagIcon class="size-3" />
                        #{{ selectedTag }}
                        <button class="ml-1 hover:text-destructive cursor-pointer" @click="setSelectedTag(null)">
                            <XMarkIcon class="size-3" />
                        </button>
                    </div>
                    <div
v-if="timeFilter !== 'all' && !selectedDate" class="flex items-center gap-1.5 px-2.5 py-1 bg-primary/10 text-primary rounded-full text-xs border border-primary/20">
                        <ClockIcon class="size-3" />
                        {{ timeFilter === 'today' ? t('article.todayArticles') : t('article.monthArticles') }}
                        <button class="ml-1 hover:text-destructive cursor-pointer" @click="setTimeFilter('all')">
                            <XMarkIcon class="size-3" />
                        </button>
                    </div>
                    <span class="text-xs text-muted-foreground ml-auto">{{ postList.length }} {{ t('nav.article') }}</span>
                </div>

                <!-- Scrollable Content -->
                <div class="flex-1 overflow-y-auto px-4 py-4 pb-20">
                    <div class="space-y-3">
                        <ArticleCard
v-for="post in currentPostList" :key="post.fileName" :post="post"
                            :selected="selectedPost.some(p => p.fileName === post.fileName)" @edit="$emit('editPost', $event)"
                            @select="onSelectChange" @preview="previewPost" @delete="deleteSinglePost" />
                    </div>
                </div>

                <!-- Pagination -->
                <div v-if="totalPages > 1" class="h-12 py-3 px-4 border-t border-border flex justify-center bg-background">
                    <Pagination :total="postList.length" :items-per-page="PAGE_SIZE" :page="currentPage" :sibling-count="2">
                        <PaginationContent>
                            <PaginationItem>
                                <PaginationPrevious
:class="{ 'pointer-events-none opacity-50': currentPage === 1 }"
                                    @click="currentPage > 1 && handlePageChanged(currentPage - 1)" />
                            </PaginationItem>

                            <template v-for="page in visiblePages" :key="page">
                                <PaginationItem v-if="page === -1">
                                    <PaginationEllipsis />
                                </PaginationItem>
                                <PaginationLink
v-else :value="page" :is-active="currentPage === page"
                                    @click="handlePageChanged(page)">
                                    {{ page }}
                                </PaginationLink>
                            </template>

                            <PaginationItem>
                                <PaginationNext
:class="{ 'pointer-events-none opacity-50': currentPage === totalPages }"
                                    @click="currentPage < totalPages && handlePageChanged(currentPage + 1)" />
                            </PaginationItem>
                        </PaginationContent>
                    </Pagination>
                </div>
            </main>
        </div>

        <!-- Delete Confirm Dialog -->
        <DeleteDialog v-model:open="deleteModalVisible" @confirm="confirmDelete" />
    </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { BrowserOpenURL } from '@/wailsjs/runtime'
import { GetPreviewURL } from '@/wailsjs/go/facade/PreviewFacade'
import { toast } from 'vue-sonner'
import { CalendarIcon, FolderIcon, TagIcon, ClockIcon, XMarkIcon } from '@heroicons/vue/24/outline'

import {
    Pagination,
    PaginationContent,
    PaginationEllipsis,
    PaginationItem,
    PaginationLink,
    PaginationNext,
    PaginationPrevious,
} from '@/components/ui/pagination'

import ContributionGraph from '@/views/memos/components/ContributionGraph.vue'
import ListHeader from './components/ListHeader.vue'
import ArticleCard from './components/ArticleCard.vue'
import DeleteDialog from './components/DeleteDialog.vue'

import { useArticleList } from './composables/useArticleList'
import { useSelection } from './composables/useSelection'

import type { IPost } from '@/interfaces/post'

defineEmits<{
    newArticle: []
    editPost: [post: IPost]
}>()

const { t } = useI18n()
const siteStore = useSiteStore()

const {
    keyword,
    selectedTag,
    selectedCategory,
    selectedDate,
    timeFilter,
    setTimeFilter,
    setSelectedTag,
    setSelectedCategory,
    setSelectedDate,
    heatmapData,
    tagStats,
    categoryStats,
    totalPosts,
    todayPosts,
    monthPosts,
    currentPage,
    PAGE_SIZE,
    totalPages,
    currentPostList,
    visiblePages,
    handlePageChanged,
    postList,
} = useArticleList()

const {
    selectedPost,
    deleteModalVisible,
    onSelectChange,
    deleteSinglePost,
    confirmDelete,
} = useSelection()

const previewPost = async (post: IPost) => {
    const { postPath } = siteStore.themeConfig
    try {
        const serverUrl = await GetPreviewURL()
        const url = `${serverUrl}/${postPath}/${post.fileName}/`
        BrowserOpenURL(url)
    } catch (e) {
        console.error(e)
        toast.error(t('article.previewServiceFailed'))
    }
}
</script>