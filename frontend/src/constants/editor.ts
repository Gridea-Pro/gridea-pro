/**
 * 编辑器相关常量定义
 */

/**
 * 编辑器配置常量
 */
export const EDITOR_CONSTANTS = {
    /** Monaco编辑器高度更新延迟（毫秒） */
    MONACO_HEIGHT_UPDATE_DELAY: 0,

    /** 自动保存间隔（毫秒） */
    AUTO_SAVE_INTERVAL: 30000, // 30秒

    /** 图片最大尺寸（字节） */
    IMAGE_MAX_SIZE: 10 * 1024 * 1024, // 10MB

    /** 防抖延迟（毫秒） */
    DEBOUNCE_DELAY: 300,
} as const

/**
 * Toast 提示持续时间（毫秒）
 */
export const TOAST_DURATION = {
    SUCCESS: 2000,
    ERROR: 3000,
    WARNING: 2500,
    INFO: 2000,
} as const

/**
 * 分页配置
 */
export const PAGINATION = {
    /** 默认每页条数 */
    DEFAULT_PAGE_SIZE: 20,

    /** 最大每页条数 */
    MAX_PAGE_SIZE: 100,

    /** 最小每页条数 */
    MIN_PAGE_SIZE: 10,
} as const

/**
 * 文件类型常量
 */
export const FILE_TYPES = {
    /** 允许的图片类型 */
    ALLOWED_IMAGE_TYPES: [
        'image/jpeg',
        'image/png',
        'image/gif',
        'image/webp',
        'image/svg+xml'
    ],

    /** 图片文件扩展名 */
    IMAGE_EXTENSIONS: ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.svg'],
} as const

/**
 * 特色图片类型
 */
export enum FeatureImageType {
    DEFAULT = 'DEFAULT',
    EXTERNAL = 'EXTERNAL'
}
