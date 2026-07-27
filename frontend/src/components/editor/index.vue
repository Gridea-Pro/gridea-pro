<template>
  <div class="gridea-tiptap" :class="{ 'is-post': isPostPage }">
    <Toolbar
      :editor="editor"
      :mode="mode"
      @link="openLink"
      @image="imageOpen = true"
      @color="onColorSelect"
      @emoji="onEmojiSelect"
      @polish="polishSelection"
      @update:mode="setMode"
    />

    <div class="editor-body" :class="`mode-${mode}`">
      <div v-show="mode !== 'source'" class="rich-pane">
        <EditorBubbleMenu :editor="editor ?? null" @link="openLink" @polish="polishSelection" />
        <DragHandle :editor="editor ?? null" />
        <TableMenu :editor="editor ?? null" />
        <EditorContent :editor="editor" class="rich-content" @keydown="onKeydown" @focus.capture="emit('focus')" />
      </div>
      <div v-show="mode !== 'rich'" class="source-pane" @keydown="onKeydown" @focusin="emit('focus')">
        <SourceEditor v-model:value="model" />
      </div>
    </div>

    <LinkDialog v-model:open="linkOpen" :url="linkUrl" :text="linkText" @save="onLinkSave" @remove="onLinkRemove" />
    <ImageDialog v-model:open="imageOpen" @insert-url="onImageInsertUrl" @pick-local="pickImageFromDialog" />
    <MathPopover
      :open="mathEdit.open"
      :kind="mathEdit.kind"
      :latex="mathEdit.latex"
      :x="mathEdit.x"
      :y="mathEdit.y"
      @save="onMathSave"
      @cancel="onMathCancel"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onBeforeUnmount, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import './styles/theme.css'
import './styles/editor.css'
import { buildExtensions } from './extensions'
import { SlashCommand } from './extensions/slash/SlashCommand'
import { getMarkdown, setMarkdown } from './markdown'
import Toolbar from './ui/Toolbar.vue'
import SourceEditor from './SourceEditor.vue'
import EditorBubbleMenu from './ui/BubbleMenu.vue'
import DragHandle from './ui/DragHandle.vue'
import TableMenu from './ui/TableMenu.vue'
import LinkDialog from './ui/LinkDialog.vue'
import ImageDialog from './ui/ImageDialog.vue'
import MermaidView from './ui/MermaidView.vue'
import CodeBlockView from './ui/CodeBlockView.vue'
import FootnoteRefView from './ui/FootnoteRefView.vue'
import FootnoteDefView from './ui/FootnoteDefView.vue'
import DetailsView from './ui/DetailsView.vue'
import MathPopover from './ui/MathPopover.vue'
import type { EditorMode } from './types'
import { toast } from '@/helpers/toast'
import { OpenImageDialog } from '@/wailsjs/go/app/App'
import { UploadImagesFromFrontend, SaveImageBytesFromFrontend } from '@/wailsjs/go/facade/PostFacade'
import { Polish, Complete } from '@/wailsjs/go/facade/AIFacade'
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
  extensions: [
    ...buildExtensions({
      content: model.value || '',
      placeholder: props.placeholder,
      upload: async (file: File) => uploadBytes(file),
      // NodeView 与 AI 续写仅在富文本环境注入（测试台不传，节点退回 renderHTML、AI 续写惰性）
      nodeViews: {
        mermaid: MermaidView,
        codeBlock: CodeBlockView,
        footnoteRef: FootnoteRefView,
        footnoteDef: FootnoteDefView,
        details: DetailsView,
      },
      aiComplete: async (prefix: string, suffix: string) => {
        try {
          return await Complete(prefix, suffix)
        } catch {
          return ''
        }
      },
      // 点击公式 → 打开 LaTeX 编辑弹层
      onMathEdit: openMathEdit,
    }),
    // UI 耦合扩展在此加入（不进 buildExtensions，避免污染纯 Markdown 测试台）
    SlashCommand,
  ],
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

// ── 链接 / 图片弹窗 + 颜色 / Emoji ─────────────────────
const linkOpen = ref(false)
const linkUrl = ref('')
const linkText = ref('')
const imageOpen = ref(false)

function openLink() {
  const e = editor.value
  if (!e) return
  const { from, to } = e.state.selection
  linkUrl.value = (e.getAttributes('link').href as string) || ''
  linkText.value = from !== to ? e.state.doc.textBetween(from, to, '') : ''
  linkOpen.value = true
}
function onLinkSave(payload: { url: string; text: string }) {
  const e = editor.value
  if (!e) return
  const { url, text } = payload
  const { from, to } = e.state.selection
  if (from !== to) {
    if (text && text !== e.state.doc.textBetween(from, to, '')) {
      e.chain().focus().insertContentAt({ from, to }, text)
        .setTextSelection({ from, to: from + text.length }).setLink({ href: url }).run()
    } else {
      e.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
    }
  } else {
    // 无选区：结构化插入文本并套链接 mark（避免 label/url 中的 markdown 特殊字符被再解析）
    const label = text || url
    e.chain().focus().insertContent(label)
      .setTextSelection({ from, to: from + label.length }).setLink({ href: url }).run()
  }
}
function onLinkRemove() {
  editor.value?.chain().focus().extendMarkRange('link').unsetLink().run()
}
function onImageInsertUrl(src: string) {
  if (src) insertImageByPath(src)
}
function onColorSelect(hex: string) {
  const e = editor.value
  if (!e) return
  if (hex) e.chain().focus().setColor(hex).run()
  else e.chain().focus().unsetColor().run()
}
function onEmojiSelect(emoji: string) {
  editor.value?.chain().focus().insertContent(emoji).run()
}

// 斜杠命令中的 UI 触发型动作（link/image）经编辑器 DOM 派发；监听绑在本编辑器 DOM 上（避免多实例 window 串扰）
function onEditorAction(ev: Event) {
  const action = (ev as CustomEvent<{ action?: string }>).detail?.action
  if (action === 'link') openLink()
  else if (action === 'image') imageOpen.value = true
}
watch(
  () => editor.value,
  (ed, prev) => {
    ;(prev?.view?.dom as HTMLElement | undefined)?.removeEventListener('gridea-editor:action', onEditorAction)
    ;(ed?.view?.dom as HTMLElement | undefined)?.addEventListener('gridea-editor:action', onEditorAction)
  },
  { immediate: true },
)

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

// ── 数学公式编辑弹层 ─────────────────────────────────
const mathEdit = ref<{
  open: boolean
  kind: 'inline' | 'block'
  pos: number
  latex: string
  x: number
  y: number
}>({ open: false, kind: 'inline', pos: -1, latex: '', x: 0, y: 0 })

function openMathEdit(p: { kind: 'inline' | 'block'; pos: number; latex: string }) {
  const e = editor.value
  if (!e) return
  let x = window.innerWidth / 2 - 160
  let y = 120
  try {
    const coords = e.view.coordsAtPos(p.pos)
    x = Math.min(Math.max(8, coords.left), window.innerWidth - 340)
    // y 也要夹住，避免视口底部的公式弹层落到折叠线以下不可见
    y = Math.min(coords.bottom + 6, window.innerHeight - 300)
  } catch {
    /* 坐标取不到时退回默认位置 */
  }
  mathEdit.value = { open: true, kind: p.kind, pos: p.pos, latex: p.latex, x, y }
}
function onMathSave(latex: string) {
  const e = editor.value
  const pos = mathEdit.value.pos
  if (e && pos >= 0) {
    e.chain()
      .command(({ tr }) => {
        // 弹层打开期间文档可能被改动 → 校验 pos 仍指向公式节点，避免写错节点/静默抛错
        const node = tr.doc.nodeAt(pos)
        if (!node || (node.type.name !== 'inlineMath' && node.type.name !== 'blockMath')) return false
        if (!latex.trim()) {
          // 清空即删除，避免遗留空公式孤儿节点
          tr.delete(pos, pos + node.nodeSize)
        } else {
          tr.setNodeAttribute(pos, 'latex', latex)
        }
        return true
      })
      .run()
  }
  mathEdit.value.open = false
}
function onMathCancel() {
  mathEdit.value.open = false
}

// ── 暴露给父组件 ─────────────────────────────────────
function insertMore() {
  editor.value?.chain().focus().insertContent({ type: 'moreBreak' }).run()
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
  ;(editor.value?.view?.dom as HTMLElement | undefined)?.removeEventListener('gridea-editor:action', onEditorAction)
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
