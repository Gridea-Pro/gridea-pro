/**
 * 文章表单管理 Composable
 * 
 * 管理文章表单状态、验证和数据转换
 */

import { reactive, computed, onMounted } from 'vue'
import { useSiteStore } from '@/stores/site'
import { useI18n } from 'vue-i18n'
import dayjs from 'dayjs'
import shortid from 'shortid'
import slug from '@/helpers/slug'
import type { IPost } from '@/interfaces/post'
import type { PostForm, PostFormData, ValidationResult } from '@/types/post'
import { UrlFormats } from '@/helpers/enums'

export interface UsePostFormOptions {
    articleFileName?: string
    onLoad?: (form: PostForm) => void
}

export function usePostForm(options: UsePostFormOptions = {}) {
    const siteStore = useSiteStore()
    const { t } = useI18n()

    // 表单状态
    const form = reactive<PostForm>({
        title: '',
        fileName: '',
        tags: [],
        date: dayjs(),
        content: '',
        published: false,
        hideInList: false,
        isTop: false,
        featureImage: {
            path: '',
            name: '',
            type: ''
        },
        featureImagePath: '',
        deleteFileName: ''
    })

    // 用于跟踪文件名是否已被用户手动修改
    let fileNameChanged = false
    let originalFileName = ''
    let currentPostIndex = -1

    /**
     * 从已存在的文章加载数据
     */
    const loadExistingPost = () => {
        const { articleFileName } = options

        if (!articleFileName) {
            // 新文章：根据URL格式生成默认文件名
            if (siteStore.themeConfig.postUrlFormat === UrlFormats.ShortId) {
                form.fileName = shortid.generate()
            }
            return
        }

        // 编辑现有文章
        fileNameChanged = true // 编辑时不自动更新URL
        currentPostIndex = siteStore.posts.findIndex(
            (item: IPost) => item.fileName === articleFileName
        )

        if (currentPostIndex === -1) return

        const currentPost = siteStore.posts[currentPostIndex]
        originalFileName = currentPost.fileName

        // 填充表单数据
        form.title = currentPost.data.title
        form.fileName = currentPost.fileName
        form.tags = currentPost.data.tags || []
        form.date = dayjs(currentPost.data.date).isValid()
            ? dayjs(currentPost.data.date)
            : dayjs()
        form.content = currentPost.content
        form.published = currentPost.data.published
        form.hideInList = currentPost.data.hideInList || false
        form.isTop = currentPost.data.isTop || false

        // 处理特色图片
        if (currentPost.data.feature) {
            if (currentPost.data.feature.includes('http')) {
                form.featureImagePath = currentPost.data.feature
            } else {
                // 移除 'images/' 前缀
                form.featureImage.path = currentPost.data.feature.substring(7) || ''
                form.featureImage.name = form.featureImage.path.replace(/^.*[\\/]/, '')
            }
        }

        // 调用回调
        options.onLoad?.(form)
    }

    /**
     * 自动生成文件名（基于标题）
     */
    const autoGenerateFileName = () => {
        if (fileNameChanged || !form.title) return

        if (siteStore.themeConfig.postUrlFormat === UrlFormats.Slug) {
            form.fileName = slug(form.title)
        }
    }

    /**
     * 手动设置文件名
     */
    const setFileName = (fileName: string) => {
        form.fileName = fileName
        fileNameChanged = true
    }

    /**
     * 检查文章URL是否有效（无冲突）
     */
    const checkArticleUrlValid = (): boolean => {
        const restPosts = [...siteStore.posts]
        const foundPostIndex = restPosts.findIndex(
            (post: IPost) => post.fileName === form.fileName
        )

        if (foundPostIndex === -1) {
            return true // 没有找到重复，有效
        }

        // 如果是新文章且找到重复，无效
        if (currentPostIndex === -1) {
            return false
        }

        // 如果是编辑文章，排除自己后检查
        restPosts.splice(currentPostIndex, 1)
        const duplicateIndex = restPosts.findIndex(
            (post: IPost) => post.fileName === form.fileName
        )

        return duplicateIndex === -1
    }

    /**
     * 验证表单
     */
    const validate = (): ValidationResult => {
        if (!form.title.trim()) {
            return { valid: false, error: t('titleRequired') || '标题不能为空' }
        }

        if (!form.content.trim()) {
            return { valid: false, error: t('contentRequired') || '内容不能为空' }
        }

        // 确保有文件名
        if (!form.fileName) {
            if (siteStore.themeConfig.postUrlFormat === UrlFormats.Slug) {
                form.fileName = slug(form.title)
            } else {
                form.fileName = shortid.generate()
            }
        }

        // 检查URL有效性
        if (!checkArticleUrlValid()) {
            return { valid: false, error: t('postUrlRepeatTip') || 'URL已存在' }
        }

        // 检查文件名格式
        if (form.fileName.includes('/')) {
            return { valid: false, error: t('postUrlIncludeTip') || 'URL不能包含斜杠' }
        }

        return { valid: true }
    }

    /**
     * 转换为提交数据格式
     */
    const toFormData = (published?: boolean): PostFormData => {
        // 如果文件名改变，标记删除旧文件
        if (form.fileName.toLowerCase() !== originalFileName.toLowerCase()) {
            form.deleteFileName = originalFileName
        }

        return {
            title: form.title,
            fileName: form.fileName,
            tags: [...form.tags],
            date: form.date.format('YYYY-MM-DD HH:mm:ss'),
            content: form.content,
            published: typeof published === 'boolean' ? published : form.published,
            hideInList: form.hideInList,
            isTop: form.isTop,
            featureImage: form.featureImagePath
                ? { path: '', name: '', type: '' }
                : {
                    path: form.featureImage.path || '',
                    name: form.featureImage.name || '',
                    type: form.featureImage.type || ''
                },
            featureImagePath: form.featureImagePath || '',
            deleteFileName: form.deleteFileName || ''
        }
    }

    /**
     * 重置表单
     */
    const reset = () => {
        Object.assign(form, {
            title: '',
            fileName: '',
            tags: [],
            date: dayjs(),
            content: '',
            published: false,
            hideInList: false,
            isTop: false,
            featureImage: { path: '', name: '', type: '' },
            featureImagePath: '',
            deleteFileName: ''
        })
        fileNameChanged = false
        originalFileName = ''
        currentPostIndex = -1
    }

    // Computed
    const canSubmit = computed(() => {
        return Boolean(form.title && form.content)
    })

    // 生命周期
    onMounted(() => {
        loadExistingPost()
    })

    return {
        form,
        canSubmit,
        validate,
        toFormData,
        reset,
        autoGenerateFileName,
        setFileName,
        checkArticleUrlValid
    }
}
