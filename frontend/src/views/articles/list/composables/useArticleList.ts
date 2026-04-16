import { ref, computed, watch } from 'vue'
import { useSiteStore } from '@/stores/site'
import { PAGINATION } from '@/constants/editor'
import dayjs from 'dayjs'
import type { IPost } from '@/interfaces/post'

export type TimeFilter = 'all' | 'today' | 'month'

export function useArticleList() {
    const siteStore = useSiteStore()

    const keyword = ref<string>('')
    const selectedTag = ref<string | null>(null)
    const selectedCategory = ref<string | null>(null)
    const selectedDate = ref<string | null>(null)
    const timeFilter = ref<TimeFilter>('all')

    const currentPage = ref<number>(1)
    const PAGE_SIZE = PAGINATION.DEFAULT_PAGE_SIZE

    const resetFilters = () => {
        currentPage.value = 1
    }

    watch(keyword, resetFilters)
    watch(selectedTag, resetFilters)
    watch(selectedCategory, resetFilters)
    watch(selectedDate, resetFilters)
    watch(timeFilter, resetFilters)

    const setTimeFilter = (filter: TimeFilter) => {
        timeFilter.value = filter
        if (filter !== 'all') {
            selectedTag.value = null
            selectedCategory.value = null
            selectedDate.value = null
        }
    }

    const setSelectedTag = (tag: string | null) => {
        selectedTag.value = tag
        if (tag) {
            timeFilter.value = 'all'
            selectedDate.value = null
            selectedCategory.value = null
        }
    }

    const setSelectedCategory = (category: string | null) => {
        selectedCategory.value = category
        if (category) {
            timeFilter.value = 'all'
            selectedDate.value = null
            selectedTag.value = null
        }
    }

    const setSelectedDate = (date: string | null) => {
        selectedDate.value = date
        if (date) {
            timeFilter.value = 'all'
            selectedTag.value = null
            selectedCategory.value = null
        }
    }

    const heatmapData = computed(() => {
        const map: Record<string, number> = {}
        siteStore.posts.forEach((post: IPost) => {
            const dateStr = dayjs(post.createdAt).format('YYYY-MM-DD')
            map[dateStr] = (map[dateStr] || 0) + 1
        })
        return map
    })

    const tagStats = computed(() => {
        const map: Record<string, number> = {}
        siteStore.posts.forEach((post: IPost) => {
            (post.tags || []).forEach((tag: string) => {
                map[tag] = (map[tag] || 0) + 1
            })
        })
        return Object.entries(map)
            .map(([name, count]) => ({ name, count }))
            .sort((a, b) => b.count - a.count)
    })

    const categoryStats = computed(() => {
        const map: Record<string, number> = {}
        siteStore.posts.forEach((post: IPost) => {
            (post.categories || []).forEach((cat: string) => {
                map[cat] = (map[cat] || 0) + 1
            })
        })
        return Object.entries(map)
            .map(([name, count]) => ({ name, count }))
            .sort((a, b) => b.count - a.count)
    })

    const totalPosts = computed(() => siteStore.posts.length)
    const todayPosts = computed(() => {
        const startOfDay = dayjs().startOf('day').valueOf()
        return siteStore.posts.filter((p: IPost) => dayjs(p.createdAt).valueOf() >= startOfDay).length
    })
    const monthPosts = computed(() => {
        const startOfMonth = dayjs().startOf('month').valueOf()
        return siteStore.posts.filter((p: IPost) => dayjs(p.createdAt).valueOf() >= startOfMonth).length
    })

    const postList = computed<IPost[]>(() => {
        const search = keyword.value.toLowerCase().trim()
        let posts = [...siteStore.posts]

        if (search) {
            posts = posts.filter((post: IPost) =>
                post.title.toLowerCase().includes(search),
            )
        }

        if (selectedDate.value) {
            posts = posts.filter((post: IPost) =>
                dayjs(post.createdAt).format('YYYY-MM-DD') === selectedDate.value
            )
        } else if (timeFilter.value === 'today') {
            const startOfDay = dayjs().startOf('day').valueOf()
            posts = posts.filter((p: IPost) => dayjs(p.createdAt).valueOf() >= startOfDay)
        } else if (timeFilter.value === 'month') {
            const startOfMonth = dayjs().startOf('month').valueOf()
            posts = posts.filter((p: IPost) => dayjs(p.createdAt).valueOf() >= startOfMonth)
        }

        if (selectedTag.value) {
            posts = posts.filter((post: IPost) =>
                (post.tags || []).includes(selectedTag.value!)
            )
        }

        if (selectedCategory.value) {
            posts = posts.filter((post: IPost) =>
                (post.categories || []).includes(selectedCategory.value!)
            )
        }

        return posts.sort((a, b) => {
            const aTop = a.isTop ? 1 : 0
            const bTop = b.isTop ? 1 : 0
            if (aTop !== bTop) return bTop - aTop
            return dayjs(b.createdAt).valueOf() - dayjs(a.createdAt).valueOf()
        })
    })

    const totalPages = computed(() => Math.ceil(postList.value.length / PAGE_SIZE))

    const currentPostList = computed<IPost[]>(() => {
        const start = (currentPage.value - 1) * PAGE_SIZE
        const end = currentPage.value * PAGE_SIZE
        return postList.value.slice(start, end)
    })

    const visiblePages = computed<number[]>(() => {
        const total = totalPages.value
        const current = currentPage.value
        const delta = 2
        const range: number[] = []
        const rangeWithDots: number[] = []
        let l: number | undefined

        range.push(1)
        for (let i = current - delta; i <= current + delta; i++) {
            if (i < total && i > 1) range.push(i)
        }
        range.push(total)

        for (const i of range) {
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

    const handlePageChanged = (page: number) => {
        currentPage.value = page
        window.scrollTo({ top: 0, behavior: 'smooth' })
    }

    return {
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
    }
}