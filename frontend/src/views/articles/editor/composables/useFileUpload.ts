/**
 * 文件上传 Composable
 * 
 * 封装图片上传相关逻辑
 */

import { EventsEmit, EventsOnce } from '@/wailsjs/runtime'
import { toast } from '@/helpers/toast'

export interface FeatureImage {
    path: string
    name: string
    type: string
}

export interface UploadOptions {
    maxSize?: number // 最大文件大小（字节）
    allowedTypes?: string[] // 允许的文件类型
}

const DEFAULT_OPTIONS: UploadOptions = {
    maxSize: 10 * 1024 * 1024, // 10MB
    allowedTypes: ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/svg+xml']
}

export function useFileUpload(options: UploadOptions = {}) {
    const mergedOptions = { ...DEFAULT_OPTIONS, ...options }

    /**
     * 验证文件
     */
    const validateFile = (file: File): { valid: boolean; error?: string } => {
        // 检查文件类型
        if (mergedOptions.allowedTypes && !mergedOptions.allowedTypes.includes(file.type)) {
            return {
                valid: false,
                error: `不支持的文件类型: ${file.type}。仅支持: ${mergedOptions.allowedTypes.join(', ')}`
            }
        }

        // 检查文件大小
        if (mergedOptions.maxSize && file.size > mergedOptions.maxSize) {
            const maxSizeMB = (mergedOptions.maxSize / 1024 / 1024).toFixed(2)
            return {
                valid: false,
                error: `文件过大。最大允许: ${maxSizeMB}MB`
            }
        }

        return { valid: true }
    }

    /**
     * 上传图片到本地
     * @param files - 文件数组
     * @returns Promise<上传后的文件路径数组>
     */
    const uploadImages = async (files: File[]): Promise<string[]> => {
        // 验证所有文件
        for (const file of files) {
            const validation = validateFile(file)
            if (!validation.valid) {
                toast.error(validation.error || '文件验证失败')
                throw new Error(validation.error)
            }
        }

        return new Promise((resolve, reject) => {
            EventsEmit('image-upload', files)
            EventsOnce('image-uploaded', (paths: string[]) => {
                if (paths && paths.length > 0) {
                    resolve(paths)
                } else {
                    reject(new Error('上传失败'))
                }
            })
        })
    }

    /**
     * 上传单张图片
     */
    const uploadImage = async (file: File): Promise<string> => {
        const paths = await uploadImages([file])
        return paths[0]
    }

    /**
     * 处理特色图片
     * 注意：特色图片只需验证，不需要实际上传（直接使用本地路径）
     */
    const processFeatureImage = (file: File): FeatureImage => {
        const validation = validateFile(file)
        if (!validation.valid) {
            toast.error(validation.error || '文件验证失败')
            throw new Error(validation.error)
        }

        return {
            name: file.name,
            path: (file as any).path || '', // Electron/Wails 环境下的文件路径
            type: file.type
        }
    }

    return {
        uploadImage,
        uploadImages,
        processFeatureImage,
        validateFile
    }
}
