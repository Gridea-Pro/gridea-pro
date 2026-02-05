/**
 * 图片 URL 处理 Composable
 * 
 * 统一处理本地文件路径和外部URL的转换逻辑
 * 消除在多个组件中重复的 getImageUrl 函数
 */

export function useImageUrl() {
    /**
     * 获取图片URL
     * @param path - 图片路径（本地路径、HTTP URL 或 Data URL）
     * @returns 处理后的 URL
     */
    const getImageUrl = (path: string): string => {
        if (!path) return ''

        // 如果已经是外部URL或Data URL，直接返回
        if (path.startsWith('http') || path.startsWith('data:')) {
            return path
        }

        // 本地文件通过 Wails AssetServer 中间件访问
        return `/local-file?path=${encodeURIComponent(path)}`
    }

    /**
     * 检查路径是否为外部URL
     */
    const isExternalUrl = (path: string): boolean => {
        return path.startsWith('http') || path.startsWith('https')
    }

    /**
     * 检查路径是否为Data URL
     */
    const isDataUrl = (path: string): boolean => {
        return path.startsWith('data:')
    }

    /**
     * 检查路径是否为本地文件
     */
    const isLocalFile = (path: string): boolean => {
        return !isExternalUrl(path) && !isDataUrl(path)
    }

    return {
        getImageUrl,
        isExternalUrl,
        isDataUrl,
        isLocalFile
    }
}
