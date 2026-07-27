/**
 * 编辑器交互辅助 Composable（TipTap 版）
 *
 * 职责：
 * - 插入图片（OS 文件对话框 + Wails facade 上传 + TiptapEditor 插入）
 * - 插入更多分隔符 (<!-- more -->)
 * - Emoji 插入
 * - Markdown 预览 (markdown-it 渲染)
 * - 快捷键处理 / 外部链接
 *
 * 与 TiptapEditor（`@/components/editor`）通过其 defineExpose 的命令式 API 协作。
 */

import { ref } from 'vue'
import markdown from '@/helpers/markdown'
import ga from '@/helpers/analytics'
import { toast } from '@/helpers/toast'
import { BrowserOpenURL } from '@/wailsjs/runtime'
import { UploadImagesFromFrontend } from '@/wailsjs/go/facade/PostFacade'
import { OpenImageDialog } from '@/wailsjs/go/app/App'
import { domain } from '@/wailsjs/go/models'

/** TiptapEditor 暴露给父级的命令式 API（见 components/editor/index.vue defineExpose） */
export interface TiptapEditorExposed {
  editor: unknown
  insertImageByPath: (path: string) => void
  insertImageBytes: (file: File) => Promise<void>
  insertMore: () => void
  insertEmoji: (emoji: string) => void
  focus: () => void
}

export function useEditorHelper() {
  const tiptapEditor = ref<TiptapEditorExposed | null>(null)
  const previewHtml = ref('')

  const previewVisible = ref(false)
  const entering = ref(false)

  // ── 插入图片（选择文件） ───────────────────────────────
  const insertImage = async () => {
    ga('Post', 'Post - click-insert-image', '')
    try {
      const filePath = await OpenImageDialog()
      if (!filePath) return
      const name = filePath.split('/').pop() || filePath.split('\\').pop() || 'image'
      const paths = await UploadImagesFromFrontend([new domain.UploadedFile({ name, path: filePath })])
      paths.forEach((p) => tiptapEditor.value?.insertImageByPath(p))
    } catch (e) {
      console.error(e)
      toast.error('上传图片失败')
    }
  }

  // ── 插入更多分隔符 ────────────────────────────────────
  const insertMore = () => {
    tiptapEditor.value?.insertMore()
    ga('Post', 'Post - click-add-more', '')
  }

  // ── Emoji 插入 ────────────────────────────────────────
  const handleEmojiSelect = (emoji: string) => {
    tiptapEditor.value?.insertEmoji(emoji)
  }

  // ── Markdown 预览 ─────────────────────────────────────
  const previewPost = (content: string) => {
    previewVisible.value = true
    previewHtml.value = markdown.render(content)
    ga('Post', 'Post - click-preview-post', '')
  }

  // ── 快捷键处理 ────────────────────────────────────────
  const handleInputKeydown = (e: KeyboardEvent, content: string) => {
    entering.value = true
    if (e.ctrlKey && e.key === 'p') {
      e.preventDefault()
      previewPost(content)
    }
  }

  const handlePageMousemove = () => {
    entering.value = false
  }

  // ── GA 辅助 ───────────────────────────────────────────
  const handleInfoClick = () => {
    ga('Post', 'Post - click-post-info', '')
  }

  const handleEmojiClick = () => {
    ga('Post', 'Post - click-emoji-card', '')
  }

  // ── 外部链接 ──────────────────────────────────────────
  const openPage = (url: string) => {
    BrowserOpenURL(url)
  }

  return {
    // 编辑器实例引用（指向 TiptapEditor 组件）
    tiptapEditor,
    previewHtml,
    previewVisible,
    entering,
    insertImage,
    insertMore,
    handleEmojiSelect,
    previewPost,
    handleInputKeydown,
    handlePageMousemove,
    handleInfoClick,
    handleEmojiClick,
    openPage,
  }
}
