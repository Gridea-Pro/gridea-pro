/**
 * 文章模块专用 — 图片 URL 处理
 *
 * 封装 feature 图片和内容图片的 URL 生成逻辑，
 * 统一处理本地路径、远程 URL、Wails 协议的转换。
 *
 * 底层复用全局 `useImageUrl` composable。
 */

import { ref, watch } from 'vue'
import { useSiteStore } from '@/stores/site'
import { useImageUrl as useGlobalImageUrl } from '@/composables/useImageUrl'

export function useArticleImageUrl() {
    const siteStore = useSiteStore()
    const { getImageUrl, isExternalUrl, isDataUrl, isLocalFile } = useGlobalImageUrl()

    /**
     * 列表版本号，用于强制刷新图片缓存。
     * 当 siteStore.posts 变化时自动更新。
     */
    const listVersion = ref(Date.now())

    watch(() => siteStore.posts, () => {
        listVersion.value = Date.now()
    })

    /**
     * 获取文章列表卡片中的 feature 图片 URL
     * - 外部链接 / Data URL → 直接返回
     * - `/post-images/xxx` 相对路径 → 拼接 appDir 后通过 Wails AssetServer 中间件访问
     * - 带时间戳避免浏览器缓存
     */
    const getFeatureUrl = (path: string): string => {
        if (!path) return ''
        if (path.startsWith('http') || path.startsWith('data:')) return path

        let fullPath = path
        if (path.startsWith('/post-images/')) {
            fullPath = `${siteStore.site.appDir}${path}`
        }

        return `${getImageUrl(fullPath)}&t=${listVersion.value}`
    }

    /**
     * 获取编辑器中 feature 图片的预览 URL
     * @param featureImagePath - 绝对本地路径或外部 URL
     * @param featureImageLocalPath - 通过对话框选择的本地绝对路径
     * @param timestamp - 预览时间戳，用于强制刷新
     */
    const getFeaturePreviewUrl = (
        featureImageLocalPath: string,
        featureImagePath: string,
        timestamp: number,
    ): string => {
        if (featureImageLocalPath) {
            const url = getImageUrl(featureImageLocalPath)
            return `${url}&t=${timestamp}`
        }

        // 处理 /post-images/ 相对路径
        if (featureImagePath && featureImagePath.startsWith('/post-images/')) {
            const fullPath = `${siteStore.site.appDir}${featureImagePath}`
            return `${getImageUrl(fullPath)}&t=${timestamp}`
        }

        // 外部 URL 直接返回
        return featureImagePath
    }

    return {
        listVersion,
        getFeatureUrl,
        getFeaturePreviewUrl,
        // 透传全局方法
        getImageUrl,
        isExternalUrl,
        isDataUrl,
        isLocalFile,
    }
}
