<template>
  <div class="gridea-tiptap" :class="{ 'is-post': isPostPage }">
    <Toolbar
      :editor="editor"
      :mode="mode"
      @link="toggleLink"
      @image="pickImageFromDialog"
      @polish="polishSelection"
      @update:mode="setMode"
    />

    <div class="editor-body" :class="`mode-${mode}`">
      <div v-show="mode !== 'source'" class="rich-pane">
        <EditorContent :editor="editor" class="rich-content" @keydown="onKeydown" @focus.capture="emit('focus')" />
      </div>
      <div v-show="mode !== 'rich'" class="source-pane" @keydown="onKeydown" @focusin="emit('focus')">
        <SourceEditor v-model:value="model" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onBeforeUnmount, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import './styles/theme.css'
import './styles/editor.css'
import { buildExtensions } from './extensions'
import { getMarkdown, setMarkdown } from './markdown'
import Toolbar from './ui/Toolbar.vue'
import SourceEditor from './SourceEditor.vue'
import type { EditorMode } from './types'
import { toast } from '@/helpers/toast'
import { OpenImageDialog } from '@/wailsjs/go/app/App'
import { UploadImagesFromFrontend, SaveImageBytesFromFrontend } from '@/wailsjs/go/facade/PostFacade'
import { Polish } from '@/wailsjs/go/facade/AIFacade'
import { domain } from '@/wailsjs/go/models'

const props = withDefaults(defineProps<{ isPostPage?: boolean; placeholder?: string }>(), {
  isPostPage: false,
  placeholder: '',
})

const model = defineModel<string>('value', { required: true })
const emit = defineEmits<{ keydown: [e: KeyboardEvent]; focus: [] }>()

const mode = ref<EditorMode>('rich')
const { t } = useI18n()

// 防止「外部回填 → setContent → onUpdate → 回写 model」的回环
let applyingExternal = false

const editor = useEditor({
  content: model.value || '',
  // contentType 由 @tiptap/markdown 增强：以 Markdown 解析初始内容
  contentType: 'markdown',
  extensions: buildExtensions({
    content: model.value || '',
    placeholder: props.placeholder,
    upload: async (file: File) => uploadBytes(file),
  }),
  editorProps: {
    attributes: { class: 'markdown-body focus:outline-none' },
    handlePaste: (_view, event) => {
      const files = Array.from(event.clipboardData?.files || []).filter((f) => f.type.startsWith('image/'))
      if (files.length) {
        event.preventDefault()
        files.forEach((f) => void insertImageBytes(f))
        return true
      }
      return false
    },
    handleDrop: (_view, event) => {
      const dt = (event as DragEvent).dataTransfer
      const files = Array.from(dt?.files || []).filter((f) => f.type.startsWith('image/'))
      if (files.length) {
        event.preventDefault()
        files.forEach((f) => void insertImageBytes(f))
        return true
      }
      return false
    },
  },
  onUpdate: ({ editor }) => {
    if (applyingExternal) return
    model.value = getMarkdown(editor)
  },
})

// 外部（加载文章 / 源码模式编辑）回填 → 同步富文本
watch(model, (val) => {
  // 纯源码模式下富文本不可见，无需每次按键重解析（切回 rich 时由 setMode 一次性回灌）
  if (mode.value === 'source') return
  const e = editor.value
  if (!e) return
  if (val === getMarkdown(e)) return
  applyingExternal = true
  setMarkdown(e, val ?? '', false)
  applyingExternal = false
})

function setMode(next: EditorMode) {
  // 切到源码前，确保 model 是富文本最新内容
  const e = editor.value
  if (e && mode.value !== 'source') {
    const md = getMarkdown(e)
    if (md !== model.value) model.value = md
  }
  mode.value = next
  if (next !== 'source') {
    nextTick(() => {
      const ed = editor.value
      if (ed && model.value !== getMarkdown(ed)) {
        applyingExternal = true
        setMarkdown(ed, model.value ?? '', false)
        applyingExternal = false
      }
    })
  }
}

// ── 图片 ─────────────────────────────────────────────
function bytesToBase64(bytes: Uint8Array): string {
  let binary = ''
  const chunk = 0x8000
  for (let i = 0; i < bytes.length; i += chunk) {
    binary += String.fromCharCode(...bytes.subarray(i, i + chunk))
  }
  return btoa(binary)
}

async function uploadBytes(file: File): Promise<string> {
  const buf = new Uint8Array(await file.arrayBuffer())
  return SaveImageBytesFromFrontend(file.name || 'image.png', bytesToBase64(buf))
}

function insertImageByPath(path: string) {
  const e = editor.value
  if (!e || !path) return
  e.chain().focus().setImage({ src: path }).run()
}

async function insertImageBytes(file: File) {
  try {
    const path = await uploadBytes(file)
    insertImageByPath(path)
  } catch (err) {
    console.error('[editor] insertImageBytes failed:', err)
    toast.error('上传图片失败')
  }
}

async function pickImageFromDialog() {
  try {
    const filePath = await OpenImageDialog()
    if (!filePath) return
    const name = filePath.split('/').pop() || filePath.split('\\').pop() || 'image'
    const paths = await UploadImagesFromFrontend([new domain.UploadedFile({ name, path: filePath })])
    paths.forEach((p) => insertImageByPath(p))
  } catch (err) {
    console.error('[editor] pickImage failed:', err)
    toast.error('上传图片失败')
  }
}

// ── 链接 ─────────────────────────────────────────────
function toggleLink() {
  const e = editor.value
  if (!e) return
  if (e.isActive('link')) {
    e.chain().focus().unsetLink().run()
    return
  }
  const prev = (e.getAttributes('link').href as string) || ''
  const url = window.prompt('链接地址', prev)
  if (url === null) return
  if (url === '') {
    e.chain().focus().unsetLink().run()
    return
  }
  e.chain().focus().setLink({ href: url }).run()
}

// ── AI 润色 ──────────────────────────────────────────
async function polishSelection() {
  const e = editor.value
  if (!e) return
  const { from, to } = e.state.selection
  if (from === to) {
    toast.warning('请先选中要润色的文本')
    return
  }
  const text = e.state.doc.textBetween(from, to, '\n')
  try {
    const polished = await Polish(text)
    if (polished) {
      // 润色结果是 Markdown：必须以 markdown 内容类型插入，否则会被当作 HTML 解析
      e.chain().focus().insertContentAt({ from, to }, polished, { contentType: 'markdown' }).run()
    }
  } catch (err: unknown) {
    console.error('[editor] polish failed:', err)
    const msg = String((err as { message?: string })?.message || err || '')
    if (msg.includes('[DAILY_LIMIT]')) toast.error(t('settings.ai.dailyLimitReached'))
    else if (msg.includes('[RATE_LIMIT]')) toast.error(t('settings.ai.rateLimited'))
    else if (msg.includes('[UPSTREAM_429]') || msg.includes('429')) toast.error(t('settings.ai.upstream429'))
    else toast.error('润色失败')
  }
}

// ── 暴露给父组件 ─────────────────────────────────────
function insertMore() {
  editor.value?.chain().focus().insertContent('\n\n<!-- more -->\n\n').run()
}
function insertEmoji(emoji: string) {
  editor.value?.chain().focus().insertContent(emoji).run()
}
function focus() {
  editor.value?.commands.focus()
}

function onKeydown(e: KeyboardEvent) {
  // 透传 Ctrl/Cmd+S 等给父级（沿用 Monaco 时代的保存行为）
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
    emit('keydown', e)
  }
}

defineExpose({
  editor,
  getMarkdown: () => (editor.value ? getMarkdown(editor.value) : ''),
  setMarkdown: (md: string) => {
    const e = editor.value
    if (!e) return
    applyingExternal = true
    setMarkdown(e, md, false)
    applyingExternal = false
  },
  insertImageByPath,
  insertImageBytes,
  insertMore,
  insertEmoji,
  focus,
  mode,
})

onBeforeUnmount(() => {
  editor.value?.destroy()
})
</script>

<style scoped>
.gridea-tiptap {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
}
.editor-body {
  flex: 1;
  display: flex;
  min-height: 0;
  overflow: hidden;
}
.editor-body.mode-split .rich-pane,
.editor-body.mode-split .source-pane {
  width: 50%;
  border-right: 1px solid var(--editor-border);
}
.rich-pane {
  flex: 1;
  min-width: 0;
  overflow: auto;
  padding: 8px 0 80px;
}
.source-pane {
  flex: 1;
  min-width: 0;
  overflow: auto;
  border-left: 1px solid var(--editor-border);
}
.editor-body.mode-rich .source-pane,
.editor-body.mode-source .rich-pane {
  display: none;
}
.rich-content {
  height: 100%;
}
</style>
