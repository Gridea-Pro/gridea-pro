/**
 * 文章列表核心逻辑 Composable
 *
 * 职责：搜索过滤、排序（置顶优先 + 日期降序）、分页计算、页码省略号算法。
 * 从 Articles.vue 中精确迁移，零回归。
 */

import { ref, computed, watch } from 'vue'
import { useSiteStore } from '@/stores/site'
import { PAGINATION } from '@/constants/editor'
import dayjs from 'dayjs'
import type { IPost } from '@/interfaces/post'

export function useArticleList() {
    const siteStore = useSiteStore()

    // ── 搜索 ──────────────────────────────────────────────

    const keyword = ref<string>('')

    // ── 分页 ──────────────────────────────────────────────

    const currentPage = ref<number>(1)
    const PAGE_SIZE = PAGINATION.DEFAULT_PAGE_SIZE

    // 搜索关键词变化时重置到第一页
    watch(keyword, () => {
        currentPage.value = 1
    })

    // ── 排序 + 过滤后的完整列表 ────────────────────────────

    const postList = computed<IPost[]>(() => {
        const search = keyword.value.toLowerCase().trim()
        let posts: IPost[]

        if (!search) {
            posts = [...siteStore.posts]
        } else {
            posts = siteStore.posts.filter((post: IPost) =>
                post.title.toLowerCase().includes(search),
            )
        }

        return posts.sort((a, b) => {
            // 置顶优先
            const aTop = a.isTop ? 1 : 0
            const bTop = b.isTop ? 1 : 0
            if (aTop !== bTop) {
                return bTop - aTop
            }
            // 日期降序
            return dayjs(b.date).valueOf() - dayjs(a.date).valueOf()
        })
    })

    // ── 分页计算 ──────────────────────────────────────────

    const totalPages = computed(() => Math.ceil(postList.value.length / PAGE_SIZE))

    const currentPostList = computed<IPost[]>(() => {
        const start = (currentPage.value - 1) * PAGE_SIZE
        const end = currentPage.value * PAGE_SIZE
        return postList.value.slice(start, end)
    })

    /**
     * 页码省略号算法
     * - 始终显示首尾页
     * - 当前页 ± delta 范围内的页码全部显示
     * - 超出范围用 -1 (ellipsis) 代替
     */
    const visiblePages = computed<number[]>(() => {
        const total = totalPages.value
        const current = currentPage.value
        const delta = 2
        const range: number[] = []
        const rangeWithDots: number[] = []
        let l: number | undefined

        range.push(1)
        for (let i = current - delta; i <= current + delta; i++) {
            if (i < total && i > 1) {
                range.push(i)
            }
        }
        range.push(total)

        for (const i of range) {
            if (l) {
                if (i - l === 2) {
                    rangeWithDots.push(l + 1)
                } else if (i - l !== 1) {
                    rangeWithDots.push(-1) // ellipsis sentinel
                }
            }
            rangeWithDots.push(i)
            l = i
        }
        return rangeWithDots
    })

    // ── 分页操作 ──────────────────────────────────────────

    const handlePageChanged = (page: number) => {
        currentPage.value = page
        window.scrollTo({ top: 0, behavior: 'smooth' })
    }

    return {
        // 搜索
        keyword,
        // 分页
        currentPage,
        PAGE_SIZE,
        totalPages,
        currentPostList,
        visiblePages,
        handlePageChanged,
        // 完整排序列表（供删除等操作使用）
        postList,
    }
}
