/**
 * 文章表单数据模型 Composable
 *
 * 职责：
 * - 表单 reactive 数据模型定义
 * - 编辑/新建时的表单初始化 (buildCurrentForm)
 * - 文件名生成、URL 校验
 * - Tag / Category 管理
 * - 日期 computed getter/setter（与 @internationalized/date 桥接）
 * - 特色图片选择 / 显示 / 清除
 * - formatForm → facade.PostForm 构建
 *
 * 从 ArticleUpdate.vue 精确迁移，零回归。
 */

import { ref, reactive, computed, toRaw } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { useArticleStats } from './useArticleStats'
import { useArticleImageUrl } from '../../shared/useImageUrl'
import { generateId } from '@/utils/id'
import dayjs from 'dayjs'
import slug from '@/helpers/slug'
import ga from '@/helpers/analytics'
import { toast } from '@/helpers/toast'
import { UrlFormats } from '@/helpers/enums'
import { facade } from '@/wailsjs/go/models'
import type { IPost } from '@/interfaces/post'
import {
    type DateValue,
    getLocalTimeZone,
    fromDate,
    CalendarDate,
} from '@internationalized/date'

/** 特色图片内部数据结构 */
export interface FeatureImageData {
    path: string
    name: string
    type: string
}

/** 表单 reactive 类型 */
export interface ArticleFormState {
    id: string
    title: string
    fileName: string
    tags: string[]
    category: string     // 分类名称（显示用）
    categoryId: string   // 分类 Slug（存储主键）
    categories: string[]
    createdAt: dayjs.Dayjs
    content: string
    published: boolean
    hideInList: boolean
    isTop: boolean
    featureImage: FeatureImageData
    featureImagePath: string
    deleteFileName: string
}

export function useArticleForm(articleFileName: () => string) {
    const { t } = useI18n()
    const siteStore = useSiteStore()
    const { getFeaturePreviewUrl, getImageUrl } = useArticleImageUrl()

    // ── 表单数据模型 ──────────────────────────────────────

    const form = reactive<ArticleFormState>({
        id: generateId(),
        title: '',
        fileName: '',
        tags: [],
        category: '',
        categoryId: '',
        categories: [],
        createdAt: dayjs(),
        content: '',
        published: false,
        hideInList: false,
        isTop: false,
        featureImage: { path: '', name: '', type: '' },
        featureImagePath: '',
        deleteFileName: '',
    })

    // ── 编辑状态追踪 ──────────────────────────────────────

    let currentPostIndex = -1
    let originalFileName = ''
    let fileNameChanged = false

    const previewTimestamp = ref(Date.now())
    const tagInput = ref('')
    const changedAfterLastSave = ref(false)
    const articleStatusTip = ref('')

    // ── 文章统计 ──────────────────────────────────────────

    const { stats: articleStats } = useArticleStats(() => form.content)

    // ── 计算属性 ──────────────────────────────────────────

    const canSubmit = computed(() => {
        return form.title && form.content
    })

    const availableTags = computed(() => {
        return siteStore.tags.map((tag) => tag.name)
    })

    // 返回 {name, slug, id} 对象数组，供 UI 显示 name、存储 UUID
    const availableCategories = computed(() => {
        return siteStore.categories.map((category) => ({
            name: category.name,
            slug: category.slug,
            id: category.id,
        }))
    })

    // ── Tag 操作 ──────────────────────────────────────────

    const addTag = () => {
        const val = tagInput.value.trim()
        if (val && !form.tags.includes(val)) {
            form.tags.push(val)
        }
        tagInput.value = ''
    }

    const removeTag = (tag: string) => {
        form.tags = form.tags.filter((t) => t !== tag)
    }

    const selectTag = (tag: string) => {
        if (!form.tags.includes(tag)) {
            form.tags.push(tag)
        }
    }

    // ── 日期桥接 (@internationalized/date ↔ dayjs) ───────

    const dateValue = computed<DateValue>({
        get: () => {
            let dVal = form.createdAt
            if (!dayjs.isDayjs(dVal) || !dVal.isValid()) {
                dVal = dayjs()
            }
            try {
                const d = dVal.toDate()
                const zdt = fromDate(d, getLocalTimeZone())
                return new CalendarDate(zdt.year, zdt.month, zdt.day)
            } catch (e) {
                console.error('Failed to convert date', e)
                const now = new Date()
                return new CalendarDate(now.getFullYear(), now.getMonth() + 1, now.getDate())
            }
        },
        set: (val: DateValue) => {
            if (!val) return
            const current = dayjs.isDayjs(form.createdAt) && form.createdAt.isValid() ? form.createdAt : dayjs()
            const newDate = current
                .year(val.year)
                .month(val.month - 1)
                .date(val.day)
            form.createdAt = newDate
        },
    })

    const timeValue = computed({
        get: () => {
            return form.createdAt && dayjs.isDayjs(form.createdAt) && form.createdAt.isValid()
                ? form.createdAt.format('HH:mm:ss')
                : '00:00:00'
        },
        set: (val: string) => {
            if (!val) return
            const [hours, minutes, seconds] = val.split(':').map(Number)
            if (isNaN(hours) || isNaN(minutes)) return

            let current = form.createdAt
            if (!dayjs.isDayjs(current) || !current.isValid()) {
                current = dayjs()
            }
            form.createdAt = current.hour(hours).minute(minutes).second(seconds || 0)
        },
    })

    // ── Feature Image ─────────────────────────────────────

    const featureDisplayValue = computed({
        get: () => {
            if (form.featureImage.path) {
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
                const postImagesIndex = form.featureImage.path.indexOf('/post-images/')
                if (postImagesIndex !== -1) {
                    const relativePath = form.featureImage.path.substring(postImagesIndex)
                    if (val === relativePath) return
                }
                form.featureImage = { path: '', name: '', type: '' }
            }
            form.featureImagePath = val
        },
    })

    const featureImagePreviewSrc = computed(() => {
        return getFeaturePreviewUrl(
            form.featureImage.path,
            form.featureImagePath,
            previewTimestamp.value,
        )
    })

    const selectFeatureImage = async () => {
        try {
            const filePath = await (window as any).go.app.App.OpenImageDialog()
            if (!filePath) return

            const fileName = filePath.split(/[\\/]/).pop() || ''
            const ext = fileName.split('.').pop()?.toLowerCase() || ''
            const mimeTypes: Record<string, string> = {
                jpg: 'image/jpeg',
                jpeg: 'image/jpeg',
                png: 'image/png',
                gif: 'image/gif',
                webp: 'image/webp',
            }
            const mimeType = mimeTypes[ext] || 'image/jpeg'

            form.featureImage = { name: fileName, path: filePath, type: mimeType }
            form.featureImagePath = ''

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

    // ── 表单初始化 ────────────────────────────────────────

    const buildCurrentForm = () => {
        const fileName = articleFileName()
        previewTimestamp.value = Date.now()

        if (fileName) {
            fileNameChanged = true
            currentPostIndex = siteStore.posts.findIndex(
                (item: IPost) => item.fileName === fileName,
            )
            if (currentPostIndex !== -1) {
                const currentPost = siteStore.posts[currentPostIndex]
                originalFileName = currentPost.fileName

                form.id = currentPost.id || generateId()
                form.title = currentPost.title
                form.fileName = currentPost.fileName
                form.tags = currentPost.tags || []

                // 优先用 categoryIds[0]（UUID）初始化，否则用分类名称和 availableCategories 反查
                const firstCategoryId =
                    currentPost.categoryIds && currentPost.categoryIds.length > 0
                        ? currentPost.categoryIds[0]
                        : ''
                if (firstCategoryId) {
                    const matched = siteStore.categories.find(
                        (c) => c.id === firstCategoryId,
                    )
                    form.category = matched ? matched.name : firstCategoryId
                    form.categoryId = firstCategoryId
                } else {
                    form.category =
                        currentPost.categories && currentPost.categories.length > 0
                            ? currentPost.categories[0]
                            : ''
                    // 尝试通过名称查找对应 id
                    const matchedByName = siteStore.categories.find(
                        (c) => c.name === form.category,
                    )
                    form.categoryId = matchedByName ? matchedByName.id : ''
                }
                form.categories = currentPost.categories || []
                // 兼容：优先读取 createdAt，若无则读可能遗留的 date (为了前端类型兼容性判断)
                const createTime = currentPost.createdAt || (currentPost as any).date
                form.createdAt = dayjs(createTime).isValid()
                    ? dayjs(createTime)
                    : dayjs()
                form.content = currentPost.content
                form.published = currentPost.published
                form.hideInList = currentPost.hideInList
                form.isTop = currentPost.isTop

                if (currentPost.feature && currentPost.feature.includes('http')) {
                    form.featureImagePath = currentPost.feature
                    form.featureImage = { path: '', name: '', type: '' }
                } else if (
                    currentPost.feature &&
                    currentPost.feature.startsWith('/post-images/')
                ) {
                    const fName = currentPost.feature.substring(13)
                    const absolutePath = `${siteStore.site.appDir}/post-images/${fName}`
                    form.featureImage.path = absolutePath
                    form.featureImage.name = fName
                    form.featureImagePath = ''
                } else {
                    form.featureImage = { path: '', name: '', type: '' }
                    form.featureImagePath = ''
                }
            }
        } else if (
            siteStore.site.themeConfig.postUrlFormat === UrlFormats.ShortId
        ) {
            form.fileName = generateId()
        }
    }

    // ── 标题 / 文件名变更处理 ─────────────────────────────

    const handleTitleChange = () => {
        if (
            !fileNameChanged &&
            siteStore.site.themeConfig.postUrlFormat === UrlFormats.Slug
        ) {
            form.fileName = slug(form.title)
        }
    }

    const handleFileNameChange = (val: any) => {
        fileNameChanged = !!val
    }

    // ── 文件名构建 & URL 校验 ──────────────────────────────

    const buildFileName = () => {
        if (form.fileName !== '') return

        form.fileName =
            siteStore.site.themeConfig.postUrlFormat === UrlFormats.Slug
                ? slug(form.title)
                : generateId()
    }

    const checkArticleUrlValid = (): boolean => {
        const restPosts = JSON.parse(JSON.stringify(siteStore.posts))
        const foundPostIndex = restPosts.findIndex(
            (post: IPost) => post.fileName === form.fileName,
        )

        if (foundPostIndex !== -1) {
            if (currentPostIndex === -1) {
                return false
            }
            restPosts.splice(currentPostIndex, 1)
            const index = restPosts.findIndex(
                (post: IPost) => post.fileName === form.fileName,
            )
            if (index !== -1) {
                return false
            }
        }

        currentPostIndex = currentPostIndex === -1 ? 0 : currentPostIndex
        return true
    }

    // ── formatForm → facade.PostForm ──────────────────────

    const formatForm = (published?: boolean): facade.PostForm | undefined => {
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

        if (form.fileName.toLowerCase() !== originalFileName.toLowerCase()) {
            form.deleteFileName = originalFileName
        }

        console.log('Format form data', JSON.parse(JSON.stringify(form)))
        const rawForm = toRaw(form)

        const formData = {
            id: rawForm.id || generateId(),
            title: rawForm.title,
            fileName: rawForm.fileName,
            tags: [...rawForm.tags],
            categories:
                rawForm.category && rawForm.category !== '_none_'
                    ? [rawForm.category]
                    : [],
            // 将已选分类的 slug 传入后端
            categoryIds:
                rawForm.categoryId && rawForm.categoryId !== '_none_'
                    ? [rawForm.categoryId]
                    : [],
            createdAt: rawForm.createdAt.format('YYYY-MM-DD HH:mm:ss'),
            content: rawForm.content,
            published:
                typeof published === 'boolean' ? published : rawForm.published,
            hideInList: rawForm.hideInList,
            isTop: rawForm.isTop,
            featureImage: rawForm.featureImage.path
                ? {
                    path: rawForm.featureImage.path || '',
                    name: rawForm.featureImage.name || '',
                    type: rawForm.featureImage.type || '',
                }
                : { path: '', name: '', type: '' },
            featureImagePath:
                !rawForm.featureImage.path && rawForm.featureImagePath
                    ? rawForm.featureImagePath || ''
                    : '',
            deleteFileName: rawForm.deleteFileName || '',
            tagIds: [],
        }

        return new facade.PostForm(formData)
    }

    // ── 保存状态更新 ──────────────────────────────────────

    const updateArticleSavedStatus = () => {
        articleStatusTip.value = `${t('savedIn')} ${dayjs().format('HH:mm:ss')}`
        changedAfterLastSave.value = false
    }

    return {
        // 表单数据
        form,
        tagInput,
        changedAfterLastSave,
        articleStatusTip,
        previewTimestamp,
        // 计算属性
        canSubmit,
        articleStats,
        availableTags,
        availableCategories,
        // 日期
        dateValue,
        timeValue,
        // Feature Image
        featureDisplayValue,
        featureImagePreviewSrc,
        selectFeatureImage,
        clearFeatureImage,
        // Tag 操作
        addTag,
        removeTag,
        selectTag,
        // 表单操作
        buildCurrentForm,
        handleTitleChange,
        handleFileNameChange,
        formatForm,
        updateArticleSavedStatus,
    }
}
