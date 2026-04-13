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

import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'

self.MonacoEnvironment = {
  getWorker(_, label) {
    return new EditorWorker()
  },
}

// 重新定义自定义亮色主题，不再依赖外部废弃文件，保留原来精心设计的橙灰配色和选中高亮
monaco.editor.defineTheme('GrideaLight', {
  base: 'vs',
  inherit: true,
  rules: [
    { foreground: '999999', token: 'comment' },
    { foreground: 'e88501', token: 'string' },
    { foreground: '999999', token: 'string.link' },
    { foreground: '999999', token: 'variable.source' },
    { foreground: '4C51BF', token: 'variable' },
    { foreground: '2B6CB0', token: 'markup.list' },
    { foreground: '2B6CB0', token: 'markup.underline.link' },
    { foreground: '46a609', token: 'constant.numeric' },
    { foreground: '39946a', token: 'constant.language' },
    { foreground: 'b7791f', token: 'keyword' },
    { fontStyle: 'bold', token: 'markup.heading' },
    { fontStyle: 'bold', token: 'markup.bold' },
    { fontStyle: 'italic', token: 'markup.italic' },
    { foreground: '999999', token: 'punctuation.definition.constant.markdown' },
    { foreground: '999999', token: 'punctuation.definition.bold.markdown' },
    { foreground: '999999', token: 'punctuation.definition.italic.markdown' },
    { foreground: '999999', token: 'punctuation.definition.heading.markdown' },
    { foreground: '999999', token: 'punctuation.definition.heading.begin.markdown' },
    { foreground: '999999', token: 'punctuation.definition.heading.end.markdown' },
    { foreground: '999999', token: 'punctuation.definition.heading.setext.markdown' },
    { foreground: '999999', token: 'punctuation.definition.list_item.markdown' },
    { foreground: '999999', token: 'markup.list.numbered.bullet.markdown' },
    { foreground: '999999', token: 'punctuation.definition.bold.begin.markdown' },
    { foreground: '999999', token: 'punctuation.definition.bold.end.markdown' },
    { foreground: '999999', token: 'punctuation.definition.italic.begin.markdown' },
    { foreground: '999999', token: 'punctuation.definition.italic.end.markdown' },
    { foreground: '999999', token: 'punctuation.definition.variable.begin.markdown' },
    { foreground: '999999', token: 'punctuation.definition.variable.end.markdown' },
    { foreground: '999999', token: 'punctuation.definition.link.begin.markdown' },
    { foreground: '999999', token: 'punctuation.definition.link.end.markdown' },
  ],
  colors: {
    'editor.foreground': '#333333',
    'editor.background': '#FFFFFF',
    'editor.selectionBackground': '#FFEBB7',
    'editor.inactiveSelectionBackground': '#FFEBB7',
    'editor.selectionHighlightBackground': '#FFEBB7',
    'editor.wordHighlightBackground': '#FFEBB7',
    'editor.wordHighlightStrongBackground': '#FFEBB7',
    'editor.findMatchHighlightBackground': '#FFEBB7',
    'editor.lineHighlightBackground': '#ff9e74ff',
    'editorCursor.foreground': '#000000',
    'editorWhitespace.foreground': '#BFBFBF',
    'textLink.foreground': '#666',
  }
})

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
  const editorInstance = monaco.editor.create(elRef.value, {
    language: 'markdown', // 恢复标准 markdown 语言模式，兼容 monaco-markdown 插件高亮
    value: modelValue.value || '',
    fontSize: 16,
    theme: themeStore.isDark ? 'vs-dark' : 'GrideaLight',
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
    fontFamily:
      'ui-monospace, Menlo, Monaco, "Cascadia Code", "Segoe UI Mono", Consolas, "Courier New", monospace',
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

watch(
  () => themeStore.isDark,
  (isDark) => {
    monaco.editor.setTheme(isDark ? 'vs-dark' : 'GrideaLight')
  },
)

watch(isEmpty, (val) => {
  if (editorRef.value) {
    editorRef.value.updateOptions({ scrollBeyondLastLine: !val })
  }
})

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
  color: #b2b2b2;
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
  color: #718096 !important;

  &:hover {
    color: #744210 !important;
    background: #fffff0 !important;
  }
}

:deep(.decorationsOverviewRuler) {
  display: none !important;
}

:deep(.monaco-menu .monaco-action-bar.vertical .action-label.separator) {
  border-bottom-color: #e2e8f0 !important;
}

:deep(.monaco-editor-container) {
  background: transparent !important;
}

:deep(.monaco-editor) {
  .scrollbar {
    .slider {
      background: #eee;
    }
  }

  .scroll-decoration {
    box-shadow: #efefef 0 2px 2px -2px inset;
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

/* 覆盖亮色模式下的原生选中和 Monaco DOM 选中颜色为此前设定的橙黄色 */
:deep(.monaco-editor.vs) {
  .view-lines ::selection {
    background-color: #FFEBB7 !important;
  }

  .selected-text {
    background-color: #FFEBB7 !important;
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
