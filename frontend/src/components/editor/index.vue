<template>
  <div class="gridea-tiptap" :class="[`mode-${mode}`, { 'is-post': isPostPage }]">
    <Toolbar
      :editor="editor"
      :mode="mode"
      @link="openLink"
      @image="imageOpen = true"
      @color="onColorSelect"
      @highlight="onHighlightSelect"
      @emoji="onEmojiSelect"
      @polish="polishSelection"
      @summary="openSummary"
      @update:mode="setMode"
    />

    <!-- 工具栏固定在顶部；标题/摘要/正文在下方区域滚动（贴合 editor-vue 母版布局） -->
    <div class="editor-scroll">
      <div v-if="$slots.header" class="editor-header-slot">
        <slot name="header" />
      </div>
      <div class="editor-body" :class="`mode-${mode}`">
        <div v-show="mode !== 'source'" ref="richPaneRef" class="rich-pane">
          <EditorBubbleMenu
          :editor="editor ?? null"
          @link="openLink"
          @polish="polishSelection"
          @image-edit="openImageEdit"
          @image-preview="openImagePreview"
        />
          <DragHandle :editor="editor ?? null" />
          <TableMenu :editor="editor ?? null" />
          <EditorContent :editor="editor" class="rich-content" @keydown="onKeydown" @focus.capture="emit('focus')" />
        </div>
        <div v-show="mode !== 'rich'" ref="sourcePaneRef" class="source-pane" @keydown="onKeydown" @focusin="emit('focus')">
          <SourceEditor v-model:value="model" />
        </div>
      </div>
    </div>

    <LinkDialog v-model:open="linkOpen" :url="linkUrl" :text="linkText" @save="onLinkSave" @remove="onLinkRemove" />
    <ImageDialog v-model:open="imageOpen" @insert-url="onImageInsertUrl" @pick-local="pickImageFromDialog" />
    <ImageEditDialog v-model:open="imageEditOpen" :src="imageEditSrc" :alt="imageEditAlt" @save="onImageEditSave" />
    <ImageLightbox :src="previewSrc" @close="previewSrc = null" />
    <SummaryDialog
      v-model:open="summaryOpen"
      :loading="summaryLoading"
      :text="summaryText"
      @regenerate="generateSummary"
      @insert="onSummaryInsert"
    />
    <MathPopover
      :open="mathEdit.open"
      :kind="mathEdit.kind"
      :latex="mathEdit.latex"
      :x="mathEdit.x"
      :y="mathEdit.y"
      @save="onMathSave"
      @cancel="onMathCancel"
    />

    <!-- 右下角快捷键入口 + 面板 -->
    <button
      type="button"
      class="shortcuts-fab"
      :title="t('editor.shortcuts.title')"
      @click="shortcutsOpen = true"
    >
      <IconKeyboard class="h-[18px] w-[18px]" />
    </button>
    <ShortcutsPanel :open="shortcutsOpen" @close="shortcutsOpen = false" />
  </div>
</template>

<script setup lang="ts">
import { ref, onBeforeUnmount, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import { Extension } from '@tiptap/core'
import { IconKeyboard } from '@tabler/icons-vue'
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
import ShortcutsPanel from './ui/ShortcutsPanel.vue'
import CodeBlockView from './ui/CodeBlockView.vue'
import FootnoteRefView from './ui/FootnoteRefView.vue'
import FootnoteDefView from './ui/FootnoteDefView.vue'
import DetailsView from './ui/DetailsView.vue'
import ImageView from './ui/ImageView.vue'
import ImageEditDialog from './ui/ImageEditDialog.vue'
import ImageLightbox from './ui/ImageLightbox.vue'
import SummaryDialog from './ui/SummaryDialog.vue'
import MathPopover from './ui/MathPopover.vue'
import type { EditorMode } from './types'
import { toast } from '@/helpers/toast'
import { OpenImageDialog } from '@/wailsjs/go/app/App'
import { UploadImagesFromFrontend, SaveImageBytesFromFrontend } from '@/wailsjs/go/facade/PostFacade'
import { Polish, Complete, Summary } from '@/wailsjs/go/facade/AIFacade'
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
        image: ImageView,
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
    // ⌘K 打开链接弹窗（闭包直达 openLink，无需 DOM 事件中转）
    Extension.create({
      name: 'linkShortcut',
      addKeyboardShortcuts() {
        return {
          'Mod-k': () => {
            openLink()
            return true
          },
        }
      },
    }),
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

// ── 双栏同步滚动 ─────────────────────────────────────
// 富文本侧滚动者是 .rich-pane；源码侧实际滚动者是 CodeMirror 内部的 .cm-scroller
// （.cm-editor 高度 100%，外层 .source-pane 并不产生滚动）。按progress比例双向同步，
// 方向锁 + 超时解锁，防止 programmatic scrollTop 触发的回环。
const richPaneRef = ref<HTMLElement | null>(null)
const sourcePaneRef = ref<HTMLElement | null>(null)
let cmScroller: HTMLElement | null = null
let syncLock: 'rich' | 'source' | null = null
let syncUnlockTimer: ReturnType<typeof setTimeout> | null = null

function syncScrollTo(from: HTMLElement, to: HTMLElement) {
  const fromMax = from.scrollHeight - from.clientHeight
  const toMax = to.scrollHeight - to.clientHeight
  if (fromMax <= 0 || toMax <= 0) return
  to.scrollTop = (from.scrollTop / fromMax) * toMax
}
function scheduleSyncUnlock() {
  if (syncUnlockTimer) clearTimeout(syncUnlockTimer)
  syncUnlockTimer = setTimeout(() => (syncLock = null), 100)
}
function onRichScroll() {
  if (syncLock === 'source') return
  syncLock = 'rich'
  if (richPaneRef.value && cmScroller) syncScrollTo(richPaneRef.value, cmScroller)
  scheduleSyncUnlock()
}
function onSourceScroll() {
  if (syncLock === 'rich') return
  syncLock = 'source'
  if (richPaneRef.value && cmScroller) syncScrollTo(cmScroller, richPaneRef.value)
  scheduleSyncUnlock()
}
function detachScrollSync() {
  richPaneRef.value?.removeEventListener('scroll', onRichScroll)
  cmScroller?.removeEventListener('scroll', onSourceScroll)
  cmScroller = null
}
watch(mode, (m) => {
  detachScrollSync()
  if (m !== 'split') return
  nextTick(() => {
    cmScroller = (sourcePaneRef.value?.querySelector('.cm-scroller') as HTMLElement | null) ?? null
    richPaneRef.value?.addEventListener('scroll', onRichScroll, { passive: true })
    cmScroller?.addEventListener('scroll', onSourceScroll, { passive: true })
    // 进入双栏时先把源码栏对齐到富文本当前位置
    if (richPaneRef.value && cmScroller) syncScrollTo(richPaneRef.value, cmScroller)
  })
})

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

// ── 图片编辑 / 预览 ──────────────────────────────────
const imageEditOpen = ref(false)
const imageEditSrc = ref('')
const imageEditAlt = ref('')
const previewSrc = ref<string | null>(null)

function openImageEdit() {
  const e = editor.value
  if (!e || !e.isActive('image')) return
  const attrs = e.getAttributes('image')
  imageEditSrc.value = (attrs.src as string) || ''
  imageEditAlt.value = (attrs.alt as string) || ''
  imageEditOpen.value = true
}
function onImageEditSave(payload: { src: string; alt: string }) {
  editor.value?.chain().focus().updateAttributes('image', { src: payload.src, alt: payload.alt || null }).run()
}
function openImagePreview() {
  const e = editor.value
  if (!e || !e.isActive('image')) return
  previewSrc.value = (e.getAttributes('image').src as string) || null
}
function onColorSelect(color: string | null) {
  const e = editor.value
  if (!e) return
  if (color) e.chain().focus().setColor(color).run()
  else e.chain().focus().unsetColor().run()
}
function onHighlightSelect(color: string | null) {
  const e = editor.value
  if (!e) return
  if (color) e.chain().focus().setHighlight({ color }).run()
  else e.chain().focus().unsetHighlight().run()
}

// ── 快捷键面板 ───────────────────────────────────────
const shortcutsOpen = ref(false)

// ── AI 摘要 ──────────────────────────────────────────
const summaryOpen = ref(false)
const summaryLoading = ref(false)
const summaryText = ref('')

function openSummary() {
  if (!model.value?.trim()) {
    toast.warning(t('editor.summaryDialog.emptyContent'))
    return
  }
  summaryOpen.value = true
  void generateSummary()
}

async function generateSummary() {
  summaryLoading.value = true
  summaryText.value = ''
  try {
    summaryText.value = await Summary(model.value ?? '')
  } catch (err: unknown) {
    console.error('[editor] summary failed:', err)
    const msg = String((err as { message?: string })?.message || err || '')
    if (msg.includes('[DAILY_LIMIT]')) toast.error(t('settings.ai.dailyLimitReached'))
    else if (msg.includes('[RATE_LIMIT]')) toast.error(t('settings.ai.rateLimited'))
    else if (msg.includes('[UPSTREAM_429]') || msg.includes('429')) toast.error(t('settings.ai.upstream429'))
    else toast.error(t('editor.summaryDialog.failed'))
    summaryOpen.value = false
  } finally {
    summaryLoading.value = false
  }
}

function onSummaryInsert(text: string) {
  const e = editor.value
  if (!e || !text) return
  // 已有 <!-- more --> 则只插摘要段；否则补一个 more 分隔（Gridea 的发布摘要 = more 之前内容）
  let hasMore = false
  e.state.doc.descendants((n) => {
    if (n.type.name === 'moreBreak') hasMore = true
    return !hasMore
  })
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const content: any[] = [{ type: 'paragraph', content: [{ type: 'text', text }] }]
  if (!hasMore) content.push({ type: 'moreBreak' })
  e.chain().focus().insertContentAt(0, content).run()
}
function onEmojiSelect(emoji: string) {
  editor.value?.chain().focus().insertContent(emoji).run()
}

// 斜杠命令中的 UI 触发型动作（link/image）经编辑器 DOM 派发；监听绑在本编辑器 DOM 上（避免多实例 window 串扰）
function onEditorAction(ev: Event) {
  const action = (ev as CustomEvent<{ action?: string }>).detail?.action
  if (action === 'link') openLink()
  else if (action === 'image') imageOpen.value = true
  else if (action === 'image-preview') openImagePreview()
}
// 缓存挂载监听器的 DOM 元素：卸载时不能再读 editor.value.view —— @tiptap/vue-3 的 useEditor
// 会先 destroy 编辑器，之后 view getter 返回一个抛错的 Proxy，取 .dom 即触发 Runtime Error。
let actionDom: HTMLElement | null = null
watch(
  () => editor.value,
  (ed) => {
    actionDom?.removeEventListener('gridea-editor:action', onEditorAction)
    actionDom = (ed?.view?.dom as HTMLElement | undefined) ?? null
    actionDom?.addEventListener('gridea-editor:action', onEditorAction)
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
  // 对缓存的 DOM 解绑，绝不在此访问 editor.value.view（此刻已被 useEditor 先行 destroy）。
  actionDom?.removeEventListener('gridea-editor:action', onEditorAction)
  actionDom = null
  detachScrollSync()
  if (syncUnlockTimer) clearTimeout(syncUnlockTimer)
  // 不再手动 destroy：@tiptap/vue-3 的 useEditor 已注册自己的 onBeforeUnmount 负责销毁，
  // 重复 destroy + 访问已销毁 editor 正是本次 Runtime Error 的根因。
})
</script>

<style scoped>
.gridea-tiptap {
  position: relative; /* 供右下角快捷键入口/表格菜单等绝对定位 */
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
}
.shortcuts-fab {
  position: absolute;
  right: 14px;
  bottom: 12px;
  z-index: 7;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 1px solid var(--editor-border);
  background: var(--editor-bg);
  color: var(--editor-muted);
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: color 0.12s;
}
.shortcuts-fab:hover {
  color: var(--editor-fg);
}

/* 工具栏下方的内容容器 */
.editor-scroll {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
/* 单栏 rich：整体滚动，标题/摘要随正文一起滚走（贴合 editor-vue） */
.gridea-tiptap.mode-rich .editor-scroll {
  overflow-y: auto;
}
/* source / split：外层不滚，标题区固定在顶部，正文/分栏内部各自滚动
   （CodeMirror 源码编辑器自带滚动，需要有界高度才正常） */
.gridea-tiptap.mode-source .editor-scroll,
.gridea-tiptap.mode-split .editor-scroll {
  overflow: hidden;
}
.editor-header-slot {
  flex-shrink: 0;
}

.editor-body {
  display: flex;
  min-height: 0;
}
/* 单栏 rich：body 按内容高度，由 .editor-scroll 负责滚动 */
.gridea-tiptap.mode-rich .editor-body {
  flex: 0 0 auto;
}
.gridea-tiptap.mode-rich .rich-pane {
  overflow: visible;
}
/* source / split：body 撑满剩余高度，正文/分栏内部各自滚动 */
.gridea-tiptap.mode-source .editor-body,
.gridea-tiptap.mode-split .editor-body {
  flex: 1;
  overflow: hidden;
}
.gridea-tiptap.mode-source .source-pane {
  overflow: auto;
}
.editor-body.mode-split .rich-pane,
.editor-body.mode-split .source-pane {
  width: 50%;
  border-right: 1px solid var(--editor-border);
  overflow: auto;
}

.rich-pane {
  position: relative; /* 表格菜单等按钮相对内容区定位 */
  flex: 1;
  min-width: 0;
  padding: 8px 0 80px;
}
.source-pane {
  flex: 1;
  min-width: 0;
  border-left: 1px solid var(--editor-border);
}
.editor-body.mode-rich .source-pane,
.editor-body.mode-source .rich-pane {
  display: none;
}
/* 单栏正文随容器滚动（高度按内容）；双栏每栏填满各自高度 */
.gridea-tiptap.mode-split .rich-content {
  height: 100%;
}
</style>
