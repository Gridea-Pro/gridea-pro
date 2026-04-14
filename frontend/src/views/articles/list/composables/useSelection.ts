/**
 * 文章列表多选与批量操作 Composable
 *
 * 职责：多选切换、批量删除（含单条快捷删除）、删除确认弹窗状态。
 * 从 Articles.vue 中精确迁移，零回归。
 */

import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { toast } from 'vue-sonner'
import { EventsEmit } from '@/wailsjs/runtime'
import { DeletePostFromFrontend } from '@/wailsjs/go/facade/PostFacade'
import type { IPost } from '@/interfaces/post'

export function useSelection() {
    const { t } = useI18n()
    const siteStore = useSiteStore()

    // ── 选中状态 ──────────────────────────────────────────

    const selectedPost = ref<IPost[]>([])
    const deleteModalVisible = ref(false)

    /**
     * 切换单篇文章的选中状态
     */
    const onSelectChange = (post: IPost) => {
        const foundIndex = selectedPost.value.findIndex(
            (item) => item.fileName === post.fileName,
        )
        if (foundIndex !== -1) {
            selectedPost.value.splice(foundIndex, 1)
        } else {
            selectedPost.value.push(post)
        }
    }

    /**
     * 快捷删除单篇文章（选中后弹出确认框）
     */
    const deleteSinglePost = (post: IPost) => {
        selectedPost.value = [post]
        deleteModalVisible.value = true
    }

    /**
     * 确认删除（批量）
     * - 逐一调用后端 `DeletePostFromFrontend`
     * - 使用最后一次返回值更新 store
     * - 清空选中状态
     */
    const confirmDelete = async () => {
        deleteModalVisible.value = false
        const postsToDelete = JSON.parse(JSON.stringify(selectedPost.value))

        try {
            let updatedPosts: IPost[] = []

            for (const post of postsToDelete) {
                updatedPosts = (await DeletePostFromFrontend(post.fileName)) as IPost[]
            }

            if (updatedPosts) {
                siteStore.posts = updatedPosts
                EventsEmit('app-site-reload')
                toast.success(t('article.delete'))
                selectedPost.value = []
            }
        } catch (e) {
            console.error(e)
            toast.error('删除失败')
        }
    }

    /**
     * 触发全站数据重新加载
     */
    const reloadSite = () => {
        EventsEmit('app-site-reload')
    }

    return {
        selectedPost,
        deleteModalVisible,
        onSelectChange,
        deleteSinglePost,
        confirmDelete,
        reloadSite,
    }
}
