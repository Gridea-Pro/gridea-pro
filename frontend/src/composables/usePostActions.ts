/**
 * 文章操作 Composable
 * 
 * 处理文章的保存、发布等操作
 */

import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { EventsEmit, EventsOnce } from '@/wailsjs/runtime'
import { toast } from '@/helpers/toast'
import type { PostForm, PostFormData, ValidationResult } from '@/types/post'
import type { IPost } from '@/interfaces/post'
import ga from '@/helpers/analytics'

export interface UsePostActionsOptions {
    /**
     * 生成表单数据的函数
     */
    getFormData: (published?: boolean) => PostFormData

    /**
     * 验证函数
     */
    validate: () => ValidationResult

    /**
     * 保存成功回调
     */
    onSaveSuccess?: () => void

    /**
     * 发布成功回调
     */
    onPublishSuccess?: () => void

    /**
     * 操作失败回调
     */
    onError?: (error: Error) => void
}

export function usePostActions(options: UsePostActionsOptions) {
    const { t } = useI18n()
    const siteStore = useSiteStore()

    const saving = ref(false)
    const publishing = ref(false)
    const lastSavedAt = ref<Date | null>(null)

    /**
     * 保存草稿
     */
    const saveDraft = async (): Promise<void> => {
        // 验证表单
        const validation = options.validate()
        if (!validation.valid) {
            toast.error(validation.error || t('validationFailed'))
            return
        }

        saving.value = true

        try {
            const formData = options.getFormData(false)

            return new Promise((resolve, reject) => {
                EventsEmit('app-post-create', formData)

                EventsOnce('app-post-created', (data: { success: boolean; posts?: IPost[] }) => {
                    saving.value = false

                    if (data.success && data.posts) {
                        siteStore.posts = data.posts
                        lastSavedAt.value = new Date()
                        toast.success(`🎉 ${t('draftSuccess')}`)
                        options.onSaveSuccess?.()
                        resolve()
                    } else {
                        const error = new Error('保存失败')
                        toast.error(error.message)
                        options.onError?.(error)
                        reject(error)
                    }
                })
            })
        } catch (error) {
            saving.value = false
            const err = error instanceof Error ? error : new Error(String(error))
            toast.error(err.message)
            options.onError?.(err)
            throw err
        } finally {
            ga('Post', 'Post - click-save-draft', '')
        }
    }

    /**
     * 发布文章
     */
    const publish = async (): Promise<void> => {
        // 验证表单
        const validation = options.validate()
        if (!validation.valid) {
            toast.error(validation.error || t('validationFailed'))
            return
        }

        publishing.value = true

        try {
            const formData = options.getFormData(true)

            return new Promise((resolve, reject) => {
                EventsEmit('app-post-create', formData)

                EventsOnce('app-post-created', (data: { success: boolean; posts?: IPost[] }) => {
                    publishing.value = false

                    if (data.success && data.posts) {
                        siteStore.posts = data.posts
                        lastSavedAt.value = new Date()
                        toast.success(`🎉 ${t('published')}`)
                        options.onPublishSuccess?.()
                        resolve()
                    } else {
                        const error = new Error('发布失败')
                        toast.error(error.message)
                        options.onError?.(error)
                        reject(error)
                    }
                })
            })
        } catch (error) {
            publishing.value = false
            const err = error instanceof Error ? error : new Error(String(error))
            toast.error(err.message)
            options.onError?.(err)
            throw err
        } finally {
            ga('Post', 'Post - click-publish', '')
        }
    }

    /**
     * 普通保存（保持当前发布状态）
     */
    const save = async (): Promise<void> => {
        const validation = options.validate()
        if (!validation.valid) {
            toast.error(validation.error || t('validationFailed'))
            return
        }

        saving.value = true

        try {
            const formData = options.getFormData()

            return new Promise((resolve, reject) => {
                EventsEmit('app-post-create', formData)

                EventsOnce('app-post-created', (data: { success: boolean }) => {
                    saving.value = false

                    if (data.success) {
                        lastSavedAt.value = new Date()
                        options.onSaveSuccess?.()
                        resolve()
                    } else {
                        const error = new Error('保存失败')
                        options.onError?.(error)
                        reject(error)
                    }
                })
            })
        } catch (error) {
            saving.value = false
            const err = error instanceof Error ? error : new Error(String(error))
            options.onError?.(err)
            throw err
        }
    }

    return {
        saving,
        publishing,
        lastSavedAt,
        saveDraft,
        publish,
        save
    }
}
