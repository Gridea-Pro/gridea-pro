<template>
  <div class="monaco-editor-wrapper flex flex-col h-full w-full" :style="{
    maxWidth: props.isPostPage ? '728px' : 'none',
    margin: '0 auto',
    width: props.isPostPage ? '728px' : '100%',
    position: 'relative',
    overflow: 'visible'
  }">
    <div ref="elRef" class="monaco-editor-container w-full h-full" style="flex: 1;" />
    <!-- 模板级占位符，比 CSS 方案更可靠 -->
    <div v-if="isEmpty && props.placeholder" class="monaco-placeholder"
      :style="{ transform: `translateY(-${editorScrollTop}px)` }">
      {{ props.placeholder }}
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, shallowRef, onMounted, watch, onUnmounted, computed } from 'vue'
import * as monaco from 'monaco-editor'
import * as MonacoMarkdown from 'monaco-markdown'
import { useThemeStore } from '@/stores/theme'
import { readTokenHex, readTokenRaw, withAlpha } from '@/helpers/themeColor'

import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'

self.MonacoEnvironment = {
  getWorker(_, label) {
    return new EditorWorker()
  },
}

const MONACO_THEME_ID = 'gridea'

/**
 * 从语义 token 现算 Monaco 主题。
 *
 * Monaco 只接受色值字面量，无法消费 CSS 变量，因此每次主题变化都要重新
 * defineTheme——否则编辑区会停在旧配色上，与外壳出现色差断层。
 * 标点/结构符号统一走 --code-comment，正文靠 --foreground 突出。
 */
const applyMonacoTheme = () => {
  const bg = readTokenHex('background')
  const fg = readTokenHex('foreground', bg)
  const muted = readTokenRaw('code-comment', bg)
  const accent = readTokenHex('primary-text', bg)

  const punctuation = [
    'string.link', 'variable.source',
    'punctuation.definition.constant.markdown',
    'punctuation.definition.bold.markdown',
    'punctuation.definition.italic.markdown',
    'punctuation.definition.heading.markdown',
    'punctuation.definition.heading.begin.markdown',
    'punctuation.definition.heading.end.markdown',
    'punctuation.definition.heading.setext.markdown',
    'punctuation.definition.list_item.markdown',
    'markup.list.numbered.bullet.markdown',
    'punctuation.definition.bold.begin.markdown',
    'punctuation.definition.bold.end.markdown',
    'punctuation.definition.italic.begin.markdown',
    'punctuation.definition.italic.end.markdown',
    'punctuation.definition.variable.begin.markdown',
    'punctuation.definition.variable.end.markdown',
    'punctuation.definition.link.begin.markdown',
    'punctuation.definition.link.end.markdown',
  ].map((token) => ({ foreground: muted, token }))

  monaco.editor.defineTheme(MONACO_THEME_ID, {
    base: themeStore.isDark ? 'vs-dark' : 'vs',
    inherit: true,
    rules: [
      { foreground: muted, token: 'comment' },
      { foreground: readTokenRaw('code-string', bg), token: 'string' },
      { foreground: readTokenRaw('code-func', bg), token: 'variable' },
      { foreground: readTokenRaw('primary-text', bg), token: 'markup.list' },
      { foreground: readTokenRaw('primary-text', bg), token: 'markup.underline.link' },
      { foreground: readTokenRaw('code-number', bg), token: 'constant.numeric' },
      { foreground: readTokenRaw('code-string', bg), token: 'constant.language' },
      { foreground: readTokenRaw('code-keyword', bg), token: 'keyword' },
      { fontStyle: 'bold', token: 'markup.heading' },
      { fontStyle: 'bold', token: 'markup.bold' },
      { fontStyle: 'italic', token: 'markup.italic' },
      ...punctuation,
    ],
    colors: {
      'editor.foreground': fg,
      'editor.background': bg,
      'editor.selectionBackground': withAlpha(accent, 0.24),
      'editor.inactiveSelectionBackground': withAlpha(accent, 0.14),
      'editor.selectionHighlightBackground': withAlpha(accent, 0.16),
      'editor.wordHighlightBackground': withAlpha(accent, 0.14),
      'editor.wordHighlightStrongBackground': withAlpha(accent, 0.2),
      'editor.findMatchHighlightBackground': withAlpha(accent, 0.28),
      'editor.lineHighlightBackground': withAlpha(accent, 0.06),
      'editorCursor.foreground': accent,
      'editorWhitespace.foreground': withAlpha(readTokenHex('muted-foreground', bg), 0.4),
      'editorLineNumber.foreground': readTokenHex('muted-foreground', bg),
      'textLink.foreground': accent,
    },
  })
  monaco.editor.setTheme(MONACO_THEME_ID)
}

// ─── Props / Model ────────────────────────────────────────────

interface Props {
  isPostPage?: boolean
  placeholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  isPostPage: false,
  placeholder: ''
})

// 恢复为 explicit 'value' 名称以确保最大兼容性，解决父组件类型报错
const modelValue = defineModel<string>('value', { required: true })

const emit = defineEmits<{
  'keydown': [event: KeyboardEvent]
}>()

// ─── 响应式状态 ───────────────────────────────────────────────

const elRef = ref<HTMLElement | null>(null)

// 使用 shallowRef 避免 Vue 对庞大 Monaco 实例进行深度代理，防止性能崩溃
const editorRef = shallowRef<monaco.editor.IStandaloneCodeEditor | null>(null)

// 控制 watch 更新时跳过 onDidChangeModelContent 回调，防止循环触发
const isSettingValue = ref(false)

// 控制 Placeholder 的显示（CSS 伪元素方案，替代脆弱的内部 DOM 操作）
const isEmpty = computed(() => !modelValue.value || modelValue.value.trim() === '')

const themeStore = useThemeStore()

const editorScrollTop = ref(0) // Track Monaco's internal scroll position

// ─── 初始化逻辑 ───────────────────────────────────────────────

const initEditor = () => {
  if (!elRef.value) return

  // 卸载旧实例防止内存泄漏
  if (editorRef.value) {
    editorRef.value.dispose()
  }

  console.log('[Monaco] Initializing with value length:', modelValue.value?.length || 0)
  applyMonacoTheme()
  const editorInstance = monaco.editor.create(elRef.value, {
    language: 'markdown', // 恢复标准 markdown 语言模式，兼容 monaco-markdown 插件高亮
    value: modelValue.value || '',
    fontSize: 16,
    theme: MONACO_THEME_ID,
    lineNumbers: 'off',
    minimap: { enabled: false },
    wordWrap: 'on',
    cursorWidth: 2,
    cursorStyle: 'line',
    smoothScrolling: true,
    fontLigatures: true,
    cursorSmoothCaretAnimation: 'off',
    cursorBlinking: 'smooth',
    colorDecorators: true,
    extraEditorClassName: 'gridea-editor',
    folding: false,
    guides: { indentation: false },
    renderLineHighlight: 'none' as const,
    scrollbar: {
      vertical: 'hidden',
      horizontal: 'hidden',
      verticalScrollbarSize: 0,
      horizontalScrollbarSize: 0,
      useShadows: false,
      handleMouseWheel: true,
    },
    overviewRulerBorder: false,
    overviewRulerLanes: 0,
    lineHeight: 28,
    letterSpacing: 0.2,
    scrollBeyondLastLine: !isEmpty.value,
    scrollBeyondLastColumn: 0,
    wordBasedSuggestions: 'off',
    snippetSuggestions: 'none',
    lineDecorationsWidth: 0,
    occurrencesHighlight: 'off',
    selectionHighlight: false,
    dragAndDrop: false,
    links: false,
    automaticLayout: true,
    padding: { top: 24, bottom: 64 },
    fontFamily: themeStore.editorFontFamily,
    unicodeHighlight: {
      ambiguousCharacters: false,
      invisibleCharacters: false,
    },
  })

  // 强制同步初始值，确保通过 create 注入失败时有兜底
  if (modelValue.value) {
    console.log('[Monaco] Force setting initial value, length:', modelValue.value.length)
    editorInstance.setValue(modelValue.value)
  }

  editorRef.value = editorInstance

  // 重新激活扩展以恢复语法高亮
  const extension = new MonacoMarkdown.MonacoMarkdownExtension()
  extension.activate(editorInstance as any)

  // 监听内容变化：同步 modelValue
  editorInstance.onDidChangeModelContent(() => {
    if (isSettingValue.value) return
    const value = editorInstance.getValue()
    if (modelValue.value !== value) {
      modelValue.value = value
    }
  })

  // 监听滚动事件，同步给 Placeholder + 锁死水平滚动
  editorInstance.onDidScrollChange((e) => {
    editorScrollTop.value = e.scrollTop
    // 如果产生了任何水平偏移，立即重置为 0
    if (e.scrollLeft > 0) {
      editorInstance.setScrollLeft(0)
    }
  })

  editorInstance.onKeyDown((e: monaco.IKeyboardEvent) => {
    emit('keydown', e.browserEvent)
  })

  // 快捷键拦截 Cmd/Ctrl + S
  editorInstance.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
    // 触发父组件的保存逻辑
    emit('keydown', new KeyboardEvent('keydown', { ctrlKey: true, key: 's' } as any))
  })
}

// ─── 生命周期 ────────────────────────────────────────────────────────────────

onMounted(() => {
  initEditor()
})

onUnmounted(() => {
  if (editorRef.value) {
    editorRef.value.dispose()
    editorRef.value = null
  }
})

// ─── Watch ───────────────────────────────────────────────────────────────────

// 当父组件从外部更新 modelValue 时（如格式化、加载新文章），
// 使用 executeEdits 而非 setValue，保留撤销历史和光标位置
watch(modelValue, (newValue) => {
  const editor = editorRef.value
  if (!editor) {
    console.log('[Monaco] Editor not ready, skipping watch update')
    return
  }
  const currentVal = editor.getValue()
  console.log('[Monaco] modelValue changed, newValue length:', newValue?.length || 0, 'current editor value length:', currentVal.length)
  if (newValue === currentVal) return

  console.log('[Monaco] External update, new length:', newValue?.length || 0)
  isSettingValue.value = true
  const model = editor.getModel()
  if (model) {
    const fullRange = model.getFullModelRange()
    console.log('[Monaco] Applying edits to range:', fullRange)
    editor.executeEdits('external-update', [
      {
        range: fullRange,
        text: newValue || '',
        forceMoveMarkers: true,
      },
    ])
    // 在撤销栈中推入停止点，让此次外部更新作为独立的撤销单元
    editor.pushUndoStop()
  }
  isSettingValue.value = false
})

// 外观与强调色同样影响编辑区配色，不能只跟随明暗
watch(
  () => [themeStore.isDark, themeStore.surface, themeStore.accent],
  () => {
    // 等 CSS 变量在 <html> 上落定后再取值
    requestAnimationFrame(applyMonacoTheme)
  },
)

watch(isEmpty, (val) => {
  if (editorRef.value) {
    editorRef.value.updateOptions({ scrollBeyondLastLine: !val })
  }
})

watch(
  () => themeStore.editorFontFamily,
  (fontFamily) => {
    if (editorRef.value) {
      editorRef.value.updateOptions({ fontFamily })
    }
  },
)

// ─── 暴露给父组件 ─────────────────────────────────────────────────────────────

// 暴露 shallowRef，父组件可通过 watch 响应式监听编辑器实例变化
defineExpose({
  editor: editorRef,
})
</script>

<style lang="less" scoped>
.monaco-editor-wrapper {
  position: relative;
}

.monaco-placeholder {
  position: absolute;
  top: 24px; // 匹配编辑器 padding-top
  left: 5px; // 留一点边距
  color: var(--muted-foreground);
  font-size: 16px;
  line-height: 28px;
  pointer-events: none;
  z-index: 5;
  user-select: none;
}

:deep(.monaco-editor) {
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

:deep(.monaco-menu .monaco-action-bar.vertical .action-item) {
  border: none;
}

:deep(.action-menu-item) {
  color: var(--secondary-foreground) !important;

  &:hover {
    color: var(--primary-text) !important;
    background: var(--accent) !important;
  }
}

:deep(.decorationsOverviewRuler) {
  display: none !important;
}

:deep(.monaco-menu .monaco-action-bar.vertical .action-label.separator) {
  border-bottom-color: var(--border) !important;
}

:deep(.monaco-editor-container) {
  background: transparent !important;
}

:deep(.monaco-editor) {
  .scrollbar {
    .slider {
      background: color-mix(in srgb, var(--muted-foreground) 30%, transparent);
    }
  }

  .scroll-decoration {
    box-shadow: color-mix(in srgb, var(--foreground) 8%, transparent) 0 2px 2px -2px inset;
  }
}

/* 覆盖原生或默认的系统/主题 IME 输入框背景及文本颜色 */
:deep(.monaco-editor .ime-input) {
  background-color: transparent !important;
  color: transparent !important;
}

:deep(.monaco-editor .ime-input::selection) {
  background-color: transparent !important;
}

/* 增加光标与文字的距离，提升输入体验 */
:deep(.monaco-editor .cursor) {
  margin-left: 2px !important;
}

/* 原生选中与 Monaco DOM 选中统一走主题选区色 */
:deep(.monaco-editor) {
  .view-lines ::selection {
    background-color: var(--editor-selection) !important;
  }

  .selected-text {
    background-color: var(--editor-selection) !important;
  }
}
</style>

<style>
/* 全局强制修复 Monaco Command Palette (F1) 的居中和被裁切问题 */
.quick-input-widget {
  position: fixed !important;
  top: 15vh !important;
  left: 50% !important;
  transform: translateX(-50%) !important;
  margin-left: 0 !important;
  width: 600px !important;
  max-width: 90vw !important;
}

/* 防止可能的列表动画冲突导致错位 */
.quick-input-widget .monaco-list-row {
  transform: none !important;
}
</style>
