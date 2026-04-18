import { ref, watchEffect } from 'vue'
import type { Editor } from '@tiptap/core'

import markdown from '@/helpers/markdown'
import ga from '@/helpers/analytics'
import { toast } from '@/helpers/toast'
import { BrowserOpenURL } from '@/wailsjs/runtime'
import { UploadImagesFromFrontend } from '@/wailsjs/go/facade/PostFacade'
import { OpenImageDialog } from '@/wailsjs/go/app/App'
import { domain } from '@/wailsjs/go/models'

export type TiptapEditorRef = {
  getEditor: () => Editor | null
  focusEditor: () => void
  toggleHeading: (level: 1 | 2 | 3 | 4 | 5 | 6) => void
  toggleBold: () => void
  toggleItalic: () => void
  toggleStrike: () => void
  toggleInlineCode: () => void
  toggleBulletList: () => void
  toggleOrderedList: () => void
  toggleTaskList: () => void
  toggleBlockquote: () => void
  toggleCodeBlock: () => void
} | null

export function useEditorHelper(content: () => string) {
  const tiptapMarkdownEditor = ref<TiptapEditorRef>(null)

  const previewVisible = ref(false)
  const entering = ref(false)
  const previewHtml = ref('')

  watchEffect(() => {
    if (!previewVisible.value) {
      return
    }

    previewHtml.value = markdown.render(content())
  })

  const getEditor = () => {
    return tiptapMarkdownEditor.value?.getEditor() ?? null
  }

  const insertMarkdownAtCursor = (rawMarkdown: string) => {
    const editor = getEditor()
    if (!editor) {
      return
    }

    editor
      .chain()
      .focus()
      .insertContent(rawMarkdown, { contentType: 'markdown' })
      .run()
  }

  const insertTextAtCursor = (text: string) => {
    const editor = getEditor()
    if (!editor) {
      return
    }

    editor.chain().focus().insertContent(text).run()
  }

  const insertImage = async () => {
    ga('Post', 'Post - click-insert-image', '')

    try {
      const filePath = await OpenImageDialog()
      if (!filePath) {
        return
      }

      const fileName = filePath.split('/').pop() || filePath.split('\\').pop() || 'image'
      const uploadedFile = new domain.UploadedFile({
        name: fileName,
        path: filePath,
      })

      await uploadImageFiles([uploadedFile])
    } catch (error) {
      console.error(error)
      toast.error('上传图片失败')
    }
  }

  const uploadImageFiles = async (files: domain.UploadedFile[]) => {
    try {
      const uploadedPaths = await UploadImagesFromFrontend(files)
      if (!uploadedPaths.length) {
        return
      }

      const markdownImages = uploadedPaths.map((path: string) => `![](${path})`).join('\n\n')
      insertMarkdownAtCursor(markdownImages)
    } catch (error) {
      console.error(error)
      toast.error('上传图片失败')
    }
  }

  const insertMore = () => {
    const editor = getEditor()
    if (!editor) {
      return
    }

    // 直接插入块节点，避免摘要分隔符落在段落中间时退化为普通文本。
    editor.chain().focus().insertContent({ type: 'grideaMore' }).run()
    ga('Post', 'Post - click-add-more', '')
  }

  const handleEmojiSelect = (emoji: string) => {
    insertTextAtCursor(emoji)
  }

  const previewPost = () => {
    previewVisible.value = true
    ga('Post', 'Post - click-preview-post', '')
  }

  const handleInputKeydown = (event: KeyboardEvent) => {
    entering.value = true

    if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'p') {
      event.preventDefault()
      previewPost()
    }
  }

  const handlePageMousemove = () => {
    entering.value = false
  }

  const openPage = (url: string) => {
    BrowserOpenURL(url)
  }

  return {
    tiptapMarkdownEditor,
    previewHtml,
    previewVisible,
    entering,
    insertImage,
    insertMore,
    handleEmojiSelect,
    previewPost,
    handleInputKeydown,
    handlePageMousemove,
    openPage,
  }
}
