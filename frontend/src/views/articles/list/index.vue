<template>
    <div class="h-full flex flex-col bg-background text-foreground">
        <!-- Header Tools -->
        <ListHeader
:keyword="keyword" :selected-count="selectedPost.length" @update:keyword="keyword = $event"
            @delete-selected="deleteModalVisible = true" @new-article="$emit('newArticle')" />

        <!-- Content -->
        <div class="flex-1 overflow-y-auto px-4 py-6 pb-20">
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

        <!-- Delete Confirm Dialog -->
        <DeleteDialog v-model:open="deleteModalVisible" @confirm="confirmDelete" />
    </div>
</template>

<script lang="ts" setup>
import { useSiteStore } from '@/stores/site'
import { BrowserOpenURL } from '@/wailsjs/runtime'
import { GetPreviewURL } from '@/wailsjs/go/facade/PreviewFacade'
import { toast } from 'vue-sonner'

import {
    Pagination,
    PaginationContent,
    PaginationEllipsis,
    PaginationItem,
    PaginationLink,
    PaginationNext,
    PaginationPrevious,
} from '@/components/ui/pagination'

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

const siteStore = useSiteStore()

// ── Composables ─────────────────────────────────────────
const {
    keyword,
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

// ── 预览 ────────────────────────────────────────────────
const previewPost = async (post: IPost) => {
    const { postPath } = siteStore.themeConfig
    try {
        const serverUrl = await GetPreviewURL()
        const url = `${serverUrl}/${postPath}/${post.fileName}/`
        BrowserOpenURL(url)
    } catch (e) {
        console.error(e)
        toast.error('预览服务获取失败')
    }
}
</script>
