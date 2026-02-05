/**
 * 文章编辑器 Composable（主入口）
 * 
 * 整合所有编辑器相关的 Composables，提供统一的接口
 */

import { computed } from 'vue'
import { usePostForm } from './usePostForm'
import { usePostActions } from './usePostActions'
import { usePostStats } from './usePostStats'
import { useFileUpload } from './useFileUpload'
import type { PostFormData } from '@/types/post'

export interface UseArticleEditorOptions {
    /**
     * 文章文件名（编辑现有文章时提供）
     */
    articleFileName?: string

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

/**
 * 文章编辑器主 Composable
 * 
 * 使用示例:
 * ```typescript
 * const {
 *   form,
 *   canSubmit,
 *   stats,
 *   saveDraft,
 *   publish,
 *   uploadImage
 * } = useArticleEditor({ articleFileName: 'my-post' })
 * ```
 */
export function useArticleEditor(options: UseArticleEditorOptions = {}) {
    // 表单管理
    const postForm = usePostForm({
        articleFileName: options.articleFileName
    })

    // 文件上传
    const fileUpload = useFileUpload()

    // 文章统计
    const postStats = usePostStats(() => postForm.form.content)

    // 文章操作
    const postActions = usePostActions({
        getFormData: (published?: boolean): PostFormData => {
            const validation = postForm.validate()
            if (!validation.valid) {
                throw new Error(validation.error)
            }
            return postForm.toFormData(published)
        },
        validate: postForm.validate,
        onSaveSuccess: options.onSaveSuccess,
        onPublishSuccess: options.onPublishSuccess,
        onError: options.onError
    })

    return {
        // 表单相关
        form: postForm.form,
        canSubmit: postForm.canSubmit,
        validate: postForm.validate,
        reset: postForm.reset,
        autoGenerateFileName: postForm.autoGenerateFileName,
        setFileName: postForm.setFileName,

        // 统计信息
        stats: postStats.stats,
        wordsNumber: postStats.wordsNumber,
        readingTime: postStats.readingTime,

        // 文件上传
        uploadImage: fileUpload.uploadImage,
        uploadImages: fileUpload.uploadImages,
        processFeatureImage: fileUpload.processFeatureImage,

        // 操作
        saveDraft: postActions.saveDraft,
        publish: postActions.publish,
        save: postActions.save,
        saving: postActions.saving,
        publishing: postActions.publishing,
        lastSavedAt: postActions.lastSavedAt
    }
}
