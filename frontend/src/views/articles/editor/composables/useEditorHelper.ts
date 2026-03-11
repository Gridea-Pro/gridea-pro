/**
 * 编辑器交互辅助 Composable
 *
 * 职责：
 * - 插入图片 (uploadInputRef 触发 + 文件上传 + Monaco 插入)
 * - 插入更多分隔符 (<!-- more -->)
 * - Emoji 插入
 * - Markdown 预览 (Prism 高亮)
 * - 文件选择回调
 * - 快捷键处理
 *
 * 从 ArticleUpdate.vue 精确迁移，零回归。
 */

import { ref, type ShallowRef } from 'vue'
import * as monaco from 'monaco-editor'
import Prism from 'prismjs'
import markdown from '@/helpers/markdown'
import ga from '@/helpers/analytics'
import { toast } from '@/helpers/toast'
import { BrowserOpenURL } from '@/wailsjs/runtime'
import { UploadImagesFromFrontend } from '@/wailsjs/go/facade/PostFacade'
import { OpenImageDialog } from '@/wailsjs/go/app/App'
import { domain } from '@/wailsjs/go/models'

/** Monaco 编辑器组件 ref 类型（editor 为 shallowRef 包装） */
export type MonacoEditorRef = {
    editor: ShallowRef<monaco.editor.IStandaloneCodeEditor | null>
} | null

export function useEditorHelper() {
    // ── DOM Refs ──────────────────────────────────────────

    const monacoMarkdownEditor = ref<MonacoEditorRef>(null)
    const previewHtml = ref('')

    // ── UI 状态 ───────────────────────────────────────────

    const previewVisible = ref(false)
    const entering = ref(false)

    // ── 获取 Monaco Editor 实例的安全方法 ──────────────────

    const getEditor = (): monaco.editor.IStandaloneCodeEditor | null => {
        // Vue ref 嵌套 ShallowRef 时会自动解包，editor 已是裸值
        const editorInstance = monacoMarkdownEditor.value?.editor as monaco.editor.IStandaloneCodeEditor | null
        if (!editorInstance) {
            console.error('Monaco editor is not ready')
            return null
        }
        return editorInstance
    }

    // ── 在编辑器光标处插入文本 ─────────────────────────────

    const insertTextAtCursor = (text: string) => {
        const editor = getEditor()
        if (!editor) return

        const position = editor.getPosition()
        if (!position) return

        editor.executeEdits('', [
            {
                range: monaco.Range.fromPositions(position),
                text,
                forceMoveMarkers: true,
            },
        ])
        editor.focus()
    }

    // ── 插入图片 ──────────────────────────────────────────

    const insertImage = async () => {
        ga('Post', 'Post - click-insert-image', '')
        try {
            const filePath = await OpenImageDialog()
            if (!filePath) return

            const fileName = filePath.split('/').pop() || filePath.split('\\').pop() || 'image'
            const uploadedFile = new domain.UploadedFile({
                name: fileName,
                path: filePath,
            })
            await uploadImageFiles([uploadedFile])
        } catch (e) {
            console.error(e)
            toast.error('上传图片失败')
        }
    }

    const uploadImageFiles = async (files: domain.UploadedFile[]) => {
        try {
            const data = await UploadImagesFromFrontend(files)
            const editor = getEditor()
            if (!editor) return

            for (const path of data) {
                const url = `![](${path})`
                const position = editor.getPosition()
                if (!position) return
                editor.executeEdits('', [
                    {
                        range: monaco.Range.fromPositions(position),
                        text: url,
                        forceMoveMarkers: true,
                    },
                ])
            }
            editor.focus()
        } catch (e) {
            console.error(e)
            toast.error('上传图片失败')
        }
    }

    // ── 插入更多分隔符 ────────────────────────────────────

    const insertMore = () => {
        insertTextAtCursor('\n<!-- more -->\n')
        ga('Post', 'Post - click-add-more', '')
    }

    // ── Emoji 插入 ────────────────────────────────────────

    const handleEmojiSelect = (emoji: string) => {
        insertTextAtCursor(emoji)
    }

    // ── Markdown 预览 ──────────────────────────────────────

    const previewPost = (content: string) => {
        console.log('Preview post clicked')
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
        // DOM Refs
        monacoMarkdownEditor,
        previewHtml,
        // UI 状态
        previewVisible,
        entering,
        // 编辑器操作
        insertImage,
        insertMore,
        handleEmojiSelect,
        // 预览
        previewPost,
        // 快捷键
        handleInputKeydown,
        handlePageMousemove,
        // GA
        handleInfoClick,
        handleEmojiClick,
        // 外部链接
        openPage,
    }
}
