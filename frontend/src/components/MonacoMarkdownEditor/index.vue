<template>
  <div class="monaco-editor-wrapper" :style="{
    maxWidth: '728px',
    margin: '0 auto',
    width: props.isPostPage ? '728px' : 'auto',
    position: 'relative'
  }">
    <div ref="elRef" class="monaco-editor-container" :style="{
      minHeight: 'calc(100vh - 120px)',
      width: '100%'
    }" />
    <!-- 模板级占位符，比 CSS 方案更可靠 -->
    <div v-if="isEmpty && props.placeholder" class="monaco-placeholder">
      {{ props.placeholder }}
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, shallowRef, onMounted, watch, onUnmounted, computed } from 'vue'
import * as monaco from 'monaco-editor'
import * as MonacoMarkdown from 'monaco-markdown'
import theme from './theme'
import { useThemeStore } from '@/stores/theme'

// ─── 全局副作用 ───────────────────────────────────────────────

// 仅保留 editorWorker，Markdown 不需要其他语言 Worker
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'

if (!self.MonacoEnvironment) {
  self.MonacoEnvironment = {
    getWorker() {
      return new editorWorker()
    },
  }
}

// 自定义亮色主题，在模块顶层注册，避免多次挂载时重复注册
monaco.editor.defineTheme('GrideaLight', theme as monaco.editor.IStandaloneThemeData)

// ─── Props / Model ────────────────────────────────────────────

interface Props {
  isPostPage?: boolean
  placeholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  isPostPage: false,
  placeholder: '开始写作...'
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

// ─── 初始化逻辑 ───────────────────────────────────────────────

const initEditor = () => {
  if (!elRef.value) return

  // 卸载旧实例防止内存泄漏
  if (editorRef.value) {
    editorRef.value.dispose()
  }

  console.log('[Monaco] Initializing with value length:', modelValue.value?.length || 0)
  const editorInstance = monaco.editor.create(elRef.value, {
    language: 'markdown-math', // 恢复 markdown-math 语言模式
    value: modelValue.value || '',
    fontSize: 16,
    theme: themeStore.isDark ? 'vs-dark' : 'GrideaLight',
    lineNumbers: 'off',
    minimap: { enabled: false },
    wordWrap: 'on',
    cursorWidth: 2,
    cursorSmoothCaretAnimation: 'on',
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
    },
    lineHeight: 28,
    letterSpacing: 0.2,
    scrollBeyondLastLine: true,
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
      'Inter, PingFang SC, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif',
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
      // defineModel 内部会自动触发 update:value，无需手动 emit('change')
      modelValue.value = value
    }
  })

  editorInstance.onKeyDown((e: monaco.IKeyboardEvent) => {
    emit('keydown', e.browserEvent)
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
</style>