/**
 * MonacoMarkdownEditor 组件类型定义
 */
import type { editor as MonacoEditor } from 'monaco-editor'

/**
 * MonacoMarkdownEditor 组件通过 defineExpose 暴露的实例接口
 */
export interface MonacoMarkdownEditorExposed {
    editor: MonacoEditor.IStandaloneCodeEditor | null
}
