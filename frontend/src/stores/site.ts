import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { IPost } from '@/interfaces/post'
import type { ITag } from '@/interfaces/tag'
import type { ITheme } from '@/interfaces/theme'
import type { ILink } from '@/interfaces/link'

export type { ITag, ILink }

import type { IMenu } from '@/interfaces/menu'
import type { ISetting, ICommentSetting } from '@/interfaces/setting'
import {
  DEFAULT_POST_PAGE_SIZE,
  DEFAULT_ARCHIVES_PAGE_SIZE,
  DEFAULT_FEED_COUNT,
  DEFAULT_ARCHIVES_PATH,
  DEFAULT_POST_PATH,
  DEFAULT_TAG_PATH,
} from '@/helpers/constants'

export interface ThemeInfo {
  folder: string
  name: string
  version: string
  description?: string
  author?: string
  repository?: string
}

export interface SiteConfig {
  [key: string]: unknown
}

export interface ThemeCustomConfig {
  [key: string]: unknown
}

export interface ICategory {
  id: string       // 不可变 UUID
  name: string
  slug: string
  description: string
}

export interface SiteState {
  appDir: string
  config: SiteConfig
  posts: IPost[]
  tags: ITag[]
  menus: IMenu[]
  categories: ICategory[]
  links: ILink[]
  themeConfig: ITheme
  themeCustomConfig: ThemeCustomConfig
  currentThemeConfig: ThemeCustomConfig
  themes: ThemeInfo[]
  setting: ISetting
  commentSetting: ICommentSetting
}

// 默认主题配置
const defaultThemeConfig: ITheme = {
  themeName: '',
  postPageSize: DEFAULT_POST_PAGE_SIZE,
  archivesPageSize: DEFAULT_ARCHIVES_PAGE_SIZE,
  siteName: '',
  siteAuthor: '',
  siteEmail: '',
  siteDescription: '',
  footerInfo: 'Powered by Gridea Pro',
  postUrlFormat: 'SLUG',
  tagUrlFormat: 'SLUG',
  dateFormat: 'YYYY-MM-DD',
  language: 'zh-cn',
  feedCount: DEFAULT_FEED_COUNT,
  feedFullText: true,
  postPath: DEFAULT_POST_PATH,
  tagPath: DEFAULT_TAG_PATH,
}

// 默认设置
const defaultSetting: ISetting = {
  platform: 'github',
  domain: '',
  repository: '',
  branch: '',
  username: '',
  email: '',
  tokenUsername: '',
  token: '',
  cname: '',
  port: '22',
  server: '',
  password: '',
  privateKey: '',
  remotePath: '',
  proxyPath: '',
  proxyPort: '',
  enabledProxy: 'direct',
  netlifySiteId: '',
  netlifyAccessToken: '',
}

// 默认评论设置
const defaultCommentSetting: ICommentSetting = {
  showComment: false,
  commentPlatform: 'gitalk',
  gitalkSetting: {
    clientId: '',
    clientSecret: '',
    repository: '',
    owner: '',
  },
  disqusSetting: {
    api: '',
    apikey: '',
    shortname: '',
  },
}

export const useSiteStore = defineStore('site', () => {
  // State
  const appDir = ref('')
  const config = ref<SiteConfig>({})
  const posts = ref<IPost[]>([])
  const tags = ref<ITag[]>([])
  const menus = ref<IMenu[]>([])
  const categories = ref<ICategory[]>([])
  const links = ref<ILink[]>([])
  const themeConfig = ref<ITheme>({ ...defaultThemeConfig })
  const themeCustomConfig = ref<ThemeCustomConfig>({})
  const currentThemeConfig = ref<ThemeCustomConfig>({})
  const themes = ref<ThemeInfo[]>([])
  const setting = ref<ISetting>({ ...defaultSetting })
  const commentSetting = ref<ICommentSetting>({ ...defaultCommentSetting })

  // Getters
  const site = computed(() => ({
    appDir: appDir.value,
    config: config.value,
    posts: posts.value,
    tags: tags.value,
    menus: menus.value,
    categories: categories.value,
    links: links.value,
    themeConfig: themeConfig.value,
    themeCustomConfig: themeCustomConfig.value,
    currentThemeConfig: currentThemeConfig.value,
    themes: themes.value,
    setting: setting.value,
    commentSetting: commentSetting.value,
  }))

  // Actions
  function updateSite(siteData: Partial<SiteState>) {
    try {
      if (import.meta.env.DEV) {
        console.log('🔍 [Store] Updating site data')
      }

      // 基础字段
      if (siteData.appDir !== undefined) {
        appDir.value = siteData.appDir
      }

      // 数组字段
      if (siteData.posts !== undefined) {
        posts.value = Array.isArray(siteData.posts) ? siteData.posts : []
      }

      if (siteData.tags !== undefined) {
        tags.value = Array.isArray(siteData.tags) ? siteData.tags : []
      }

      if (siteData.menus !== undefined) {
        menus.value = Array.isArray(siteData.menus) ? siteData.menus : []
      }

      if (siteData.categories !== undefined) {
        categories.value = Array.isArray(siteData.categories) ? siteData.categories : []
      }

      if (siteData.links !== undefined) {
        links.value = Array.isArray(siteData.links) ? siteData.links : []
      }

      if (siteData.themes !== undefined) {
        themes.value = Array.isArray(siteData.themes) ? siteData.themes : []
      }

      // 配置对象
      if (siteData.config !== undefined) {
        config.value = siteData.config || {}
      }

      // 主题配置
      if (siteData.themeConfig !== undefined) {
        themeConfig.value = {
          ...themeConfig.value,
          ...siteData.themeConfig,
          postUrlFormat: siteData.themeConfig?.postUrlFormat || themeConfig.value.postUrlFormat,
          tagUrlFormat: siteData.themeConfig?.tagUrlFormat || themeConfig.value.tagUrlFormat,
          dateFormat: siteData.themeConfig?.dateFormat || themeConfig.value.dateFormat,
          language: siteData.themeConfig?.language || themeConfig.value.language,
          postPageSize: siteData.themeConfig?.postPageSize || DEFAULT_POST_PAGE_SIZE,
          archivesPageSize: siteData.themeConfig?.archivesPageSize || DEFAULT_ARCHIVES_PAGE_SIZE,
          feedCount: siteData.themeConfig?.feedCount || DEFAULT_FEED_COUNT,
          postPath: siteData.themeConfig?.postPath || DEFAULT_POST_PATH,
          tagPath: siteData.themeConfig?.tagPath || DEFAULT_TAG_PATH,
          feedFullText: typeof siteData.themeConfig?.feedFullText === 'boolean'
            ? siteData.themeConfig.feedFullText
            : true,
        }
      }

      if (siteData.themeCustomConfig !== undefined) {
        themeCustomConfig.value = siteData.themeCustomConfig || {}
      }

      if (siteData.currentThemeConfig !== undefined) {
        currentThemeConfig.value = siteData.currentThemeConfig || {}
      }

      if (siteData.setting !== undefined) {
        setting.value = { ...setting.value, ...siteData.setting }
      }

      if (siteData.commentSetting !== undefined) {
        commentSetting.value = { ...commentSetting.value, ...siteData.commentSetting }
      }

      if (import.meta.env.DEV) {
        console.log('✅ [Store] Site data updated successfully')
      }
    } catch (error) {
      console.error('❌ [Store] Failed to update site data:', error)
      throw error
    }
  }

  function resetSite() {
    appDir.value = ''
    config.value = {}
    posts.value = []
    tags.value = []
    menus.value = []
    categories.value = []
    links.value = []
    themeConfig.value = { ...defaultThemeConfig }
    themeCustomConfig.value = {}
    currentThemeConfig.value = {}
    themes.value = []
    setting.value = { ...defaultSetting }
    commentSetting.value = { ...defaultCommentSetting }
  }

  return {
    // State
    appDir,
    config,
    posts,
    tags,
    menus,
    categories,
    links,
    themeConfig,
    themeCustomConfig,
    currentThemeConfig,
    themes,
    setting,
    commentSetting,
    // Getters
    site,
    // Actions
    updateSite,
    resetSite,
  }
})
