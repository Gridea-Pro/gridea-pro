import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { generateAvatar } from '@/utils/avatarGenerator'
import { parseDate } from '@/utils/date'
import type { Comment, CommentSettings } from '@/types/comment'

export const useCommentStore = defineStore('comment', () => {
  const settings = ref<CommentSettings>({
    enable: false,
    platform: 'Valine',
    platformConfigs: {},
  })

  const comments = ref<Comment[]>([])
  const loading = ref(false)

  // Pagination State
  const page = ref(1)
  const pageSize = ref(50)
  const total = ref(0)
  const totalPages = ref(1)

  // 默认 lastReadTime 为 1970 年，确保首次安装能看到所有未读（或由用户点击消除）
  const lastReadTime = ref(localStorage.getItem('comment:lastReadTime') || new Date(0).toISOString())

  const unreadCount = computed(() => {
    // 简单的客户端未读数计算：基于当前页获取到的最新评论时间判断
    // 注意：如果是分页加载，这里只计算当前页的新评论
    const lastRead = new Date(lastReadTime.value)
    return comments.value.filter(c => parseDate(c.createdAt) > lastRead).length
  })

  // 标记所有为已读 (更新 lastReadTime 为最新一条评论的时间或当前时间)
  const markAllAsRead = () => {
    if (comments.value.length > 0) {
      // 列表默认按时间倒序，直接取第一条为最新
      const latest = comments.value[0].createdAt
      if (parseDate(latest) > new Date(lastReadTime.value)) {
        // Store as ISO string for consistency, or keep original format?
        // Let's store as new Date(latest).toISOString() to assume standard format in local storage
        lastReadTime.value = parseDate(latest).toISOString()
      }
    } else {
      lastReadTime.value = new Date().toISOString()
    }
    localStorage.setItem('comment:lastReadTime', lastReadTime.value)
  }

  // 从后端加载评论设置
  const loadSettings = async () => {
    try {
      // @ts-ignore - Wails 绑定
      if (window.go?.facade?.CommentFacade?.GetSettings) {
        const result = await window.go.facade.CommentFacade.GetSettings()
        if (result) {
          settings.value = {
            enable: result.enable || false,
            platform: result.platform || 'Valine',
            platformConfigs: result.platformConfigs || {},
          }
        }
      }
    } catch (error) {
      console.error('加载评论设置失败:', error)
    }
  }

  // 保存评论设置到后端
  const saveSettings = async (newSettings: CommentSettings) => {
    try {
      // @ts-ignore - Wails 绑定
      if (window.go?.facade?.CommentFacade?.SaveSettings) {
        await window.go.facade.CommentFacade.SaveSettings(newSettings)
      }
      settings.value = newSettings
      console.log('Settings saved:', newSettings)
    } catch (error) {
      console.error('保存评论设置失败:', error)
      throw error
    }
  }

  const deleteComment = async (id: string) => {
    try {
      // @ts-ignore - Wails 绑定
      if (window.go?.facade?.CommentFacade?.DeleteComment) {
        await window.go.facade.CommentFacade.DeleteComment(id)
        // 成功后从本地移除
        comments.value = comments.value.filter(c => c.id !== id)
        total.value-- // 更新总数
      }
    } catch (error) {
      console.error('删除评论失败:', error)
      throw error
    }
  }

  // 获取评论
  const fetchComments = async (p = 1, s = pageSize.value) => {
    loading.value = true
    try {
      // 尝试从后端获取
      // @ts-ignore - Wails 绑定
      if (window.go?.facade?.CommentFacade?.FetchComments) {
        // Backend now returns { comments, total, page, pageSize, totalPages }
        const result = await window.go.facade.CommentFacade.FetchComments(p, s)

        if (result) {
          comments.value = result.comments || []
          total.value = result.total || 0
          page.value = result.page || p
          pageSize.value = result.pageSize || s
          totalPages.value = result.totalPages || 1
        } else {
          comments.value = []
          total.value = 0
        }

        // 自动计算 isNew 方便前端展示高亮
        comments.value.forEach(c => {
          c.isNew = parseDate(c.createdAt) > new Date(lastReadTime.value)
        })
      }
    } catch (error) {
      console.error('从后端获取评论失败:', error)
      comments.value = []
    } finally {
      loading.value = false
    }
  }

  // 回复评论
  const replyComment = async (parentId: string, content: string, articleId: string) => {
    try {
      // @ts-ignore - Wails 绑定
      if (window.go?.facade?.CommentFacade?.ReplyComment) {
        await window.go.facade.CommentFacade.ReplyComment(parentId, content, articleId)
        await fetchComments(page.value, pageSize.value) // 刷新当前列表
        markAllAsRead() // 回复后自动标记为已读？或者保持未读？通常回复后算已读比较合理
      }
    } catch (error) {
      console.error('回复评论失败:', error)
      throw error
    }
  }

  return {
    settings,
    comments,
    loading,
    page,
    pageSize,
    total,
    totalPages,
    unreadCount,
    loadSettings,
    saveSettings,
    fetchComments,
    markAllAsRead,
    deleteComment,
    replyComment,
  }
})

