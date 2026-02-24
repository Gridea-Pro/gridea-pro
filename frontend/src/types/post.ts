/**
 * 文章相关类型定义
 */

import type { Dayjs } from 'dayjs'

export interface PostForm {
    title: string
    fileName: string
    tags: string[]
    date: Dayjs
    content: string
    published: boolean
    hideInList: boolean
    isTop: boolean
    featureImage: {
        path: string
        name: string
        type: string
    }
    featureImagePath: string
    deleteFileName: string
}

export interface PostFormData {
    title: string
    fileName: string
    tags: string[]
    date: string // ISO 格式字符串
    content: string
    published: boolean
    hideInList: boolean
    isTop: boolean
    featureImage: {
        path: string
        name: string
        type: string
    }
    featureImagePath: string
    deleteFileName: string
}

export interface ValidationResult {
    valid: boolean
    error?: string
}

export interface ArticleStats {
    wordsNumber: number
    formatTime: string
}
