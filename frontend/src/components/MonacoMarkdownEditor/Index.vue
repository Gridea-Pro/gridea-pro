<template>
  <div id="monaco-markdown-editor" style="max-width: 728px; min-height: calc(100vh - 176px); margin: 0 auto;" :style="{
    width: isPostPage ? '728px' : 'auto'
  }">
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, watch, onUnmounted } from 'vue'
import * as monaco from 'monaco-editor'
import type { editor as MonacoEditor } from 'monaco-editor'
import * as MonacoMarkdown from 'monaco-markdown'
import theme from './theme'
import { useThemeStore } from '@/stores/theme'

const props = defineProps<{
  isPostPage?: boolean
}>()

// Vue 3.4+ defineModel 简化 v-model 实现
const modelValue = defineModel<string>('value', { required: true })

const emit = defineEmits<{
  'keydown': [event: KeyboardEvent]
  'change': [value: string]
}>()

let editor: MonacoEditor.IStandaloneCodeEditor | null = null
const prevLineCount = ref(-1)
const themeStore = useThemeStore()

const setEditorHeight = () => {
  const lines = document.querySelectorAll('.view-line')
  if (lines && lines.length > 0 && lines[0]) {
    if (lines.length === 1 && !(lines[0] as HTMLElement).innerText.trim()) {
      lines[0].classList.add('input-holder')
    } else if (lines[0].classList.contains('input-holder')) {
      lines[0].classList.remove('input-holder')
    }
  }
}

const updateTheme = () => {
  const isDark = themeStore.isDark
  monaco.editor.setTheme(isDark ? 'vs-dark' : 'GrideaLight')
}

onMounted(() => {
  monaco.editor.defineTheme('GrideaLight', theme as monaco.editor.IStandaloneThemeData)

  editor = monaco.editor.create(document.getElementById('monaco-markdown-editor') as HTMLElement, {
    language: 'markdown-math',
    value: modelValue.value,
    fontSize: 15,
    theme: themeStore.isDark ? 'vs-dark' : 'GrideaLight',
    lineNumbers: 'off',
    minimap: {
      enabled: false,
    },
    wordWrap: 'on',
    cursorWidth: 2,
    cursorSmoothCaretAnimation: true,
    cursorBlinking: 'smooth',
    colorDecorators: true,
    extraEditorClassName: 'gridea-editor',
    folding: false,
    guides: {
      indentation: false,
    },
    renderLineHighlight: 'none' as const,
    scrollbar: {
      vertical: 'auto',
      horizontal: 'hidden',
      verticalScrollbarSize: 4,
    },
    lineHeight: 26.25,
    scrollBeyondLastLine: false,
    wordBasedSuggestions: 'off' as unknown as boolean,
    snippetSuggestions: 'none',
    lineDecorationsWidth: 0,
    occurrencesHighlight: 'off' as unknown as boolean,
    automaticLayout: true,
    fontFamily: 'PingFang SC,-apple-system,SF UI Text,Lucida Grande,STheiti,Microsoft YaHei,sans-serif',
  })

  const extension = new MonacoMarkdown.MonacoMarkdownExtension()
  extension.activate(editor)

  setTimeout(setEditorHeight, 0)

  editor.onDidChangeModelContent(() => {
    setTimeout(setEditorHeight, 0)
    if (isSettingValue.value) return
    const value = editor!.getValue()
    if (modelValue.value !== value) {
      modelValue.value = value
      emit('change', value)
    }
  })
  
  editor.onKeyDown((e: monaco.IKeyboardEvent) => {
    emit('keydown', e.browserEvent)
  })
})

onUnmounted(() => {
  if (editor) {
    editor.dispose()
    editor = null
  }
})

// 重复的 onUnmounted 已移除

const isSettingValue = ref(false)

watch(modelValue, (newValue) => {
  if (editor && newValue !== editor.getValue()) {
    isSettingValue.value = true
    editor.setValue(newValue)
    isSettingValue.value = false
  }
})

watch(() => themeStore.isDark, () => {
  updateTheme()
})

// Expose editor instance for parent component to use
defineExpose({
  editor
})
</script>

<style lang="less" scoped>
:deep(.context-view .monaco-scrollable-element) {
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, .1), 0 1px 2px 0 rgba(0, 0, 0, .06) !important;
  border-radius: 4px;
}

:deep(.monaco-menu .monaco-action-bar.vertical .action-item) {
  border: none;
}

:deep(.action-menu-item) {
  color: #718096 !important;
  &:hover {
    color: #744210 !important;
    background: #FFFFF0 !important;
  }
}

:deep(.decorationsOverviewRuler) {
  display: none !important;
}

:deep(.monaco-menu .monaco-action-bar.vertical .action-label.separator) {
  border-bottom-color: #E2E8F0 !important;
}
:deep(.input-holder) {
  &:before {
    content: '开始写作...';
    color: rgba(208, 211, 217, 0.6);
  }
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