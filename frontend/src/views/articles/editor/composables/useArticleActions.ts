/**
 * 文章保存/发布业务操作 Composable
 *
 * 职责：
 * - saveDraft: 保存草稿 (published=false)
 * - handleConfirmPublish: 确认发布 (published=true)
 * - normalSavePost: 静默保存（菜单 Ctrl+S 触发）
 * - Wails API 调用 + Store 更新
 * - Wails Events 注册/注销
 *
 * 从 ArticleUpdate.vue 精确迁移，零回归。
 */

import { watch, onMounted, onUnmounted, type Ref, type ComputedRef } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { toast } from '@/helpers/toast'
import ga from '@/helpers/analytics'
import { EventsOn, EventsOff } from '@/wailsjs/runtime'
import { SavePostFromFrontend } from '@/wailsjs/go/facade/PostFacade'
import { facade } from '@/wailsjs/go/models'
import type { IPost } from '@/interfaces/post'
import type { ITag } from '@/interfaces/tag'
import type { ArticleFormState } from './useArticleForm'

interface UseArticleActionsOptions {
    form: ArticleFormState
    canSubmit: ComputedRef<string | false>
    changedAfterLastSave: Ref<boolean>
    articleSettingsVisible: Ref<boolean>
    formatForm: (published?: boolean) => facade.PostForm | undefined
    updateArticleSavedStatus: () => void
    onClose: () => void
    onFetchData: () => void
}

export function useArticleActions(options: UseArticleActionsOptions) {
    const {
        form,
        canSubmit,
        changedAfterLastSave,
        articleSettingsVisible,
        formatForm,
        updateArticleSavedStatus,
        onClose,
        onFetchData,
    } = options

    const { t } = useI18n()
    const siteStore = useSiteStore()

    // ── 保存草稿 ──────────────────────────────────────────

    const saveDraft = async () => {
        console.log('Save draft clicked', canSubmit.value)
        if (!canSubmit.value) return
        const formData = formatForm(false)
        console.log('Form data prepared', formData)
        if (!formData) return

        try {
            const data = await SavePostFromFrontend(formData)
            if (data && data.posts) siteStore.posts = data.posts as IPost[]
            if (data && data.tags) siteStore.tags = data.tags as ITag[]

            updateArticleSavedStatus()
            toast.success(`🎉  ${t('draftSuccess')}`)
            onClose()
        } catch (e) {
            console.error(e)
            toast.error('保存失败')
        }

        ga('Post', 'Post - click-save-draft', '')
    }

    // ── 发布 ──────────────────────────────────────────────

    const publishPost = () => {
        handleArticleSettingClick()
    }

    const handleConfirmPublish = async () => {
        console.log('Confirm publish clicked', canSubmit.value)
        if (!canSubmit.value) return
        const formData = formatForm(true)
        console.log('Publish form data', formData)
        if (!formData) return

        try {
            const data = await SavePostFromFrontend(formData)
            if (data && data.posts) siteStore.posts = data.posts as IPost[]
            if (data && data.tags) siteStore.tags = data.tags as ITag[]

            updateArticleSavedStatus()
            toast.success(`🎉  ${t('published')}`)
            articleSettingsVisible.value = false
            onClose()
        } catch (e) {
            console.error(e)
            toast.error('发布失败')
        }
    }

    // ── 静默保存 (Ctrl+S / 菜单) ──────────────────────────

    const normalSavePost = async () => {
        if (!canSubmit.value) return
        const formData = formatForm()
        if (!formData) return

        try {
            await SavePostFromFrontend(formData)
            updateArticleSavedStatus()
            onFetchData()
        } catch (e) {
            console.error(e)
        }
    }

    // ── 设置面板 ──────────────────────────────────────────

    const handleArticleSettingClick = () => {
        console.log('Post settings clicked')
        articleSettingsVisible.value = true
        ga('Post', 'Post - click-post-setting', '')
    }

    // ── Wails Events 生命周期 ──────────────────────────────

    const setupEvents = () => {
        EventsOff('click-menu-save')
        EventsOff('app-post-created')
        EventsOff('image-uploaded')

        EventsOn('click-menu-save', () => {
            normalSavePost()
        })

        // 监听表单变更
        watch(
            () => form,
            () => {
                changedAfterLastSave.value = true
            },
            { deep: true },
        )
    }

    const cleanupEvents = () => {
        EventsOff('click-menu-save')
        EventsOff('app-post-created')
        EventsOff('image-uploaded')
    }

    return {
        saveDraft,
        publishPost,
        handleConfirmPublish,
        normalSavePost,
        handleArticleSettingClick,
        setupEvents,
        cleanupEvents,
    }
}
