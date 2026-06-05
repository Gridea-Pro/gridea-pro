<template>
  <NodeViewWrapper
    class="mermaid-node-view"
    :class="{ 'is-selected': selected, 'is-editing': editing }"
  >
    <div v-if="!editing" class="mermaid-render" @dblclick="startEdit">
      <div v-if="error" class="mermaid-error">
        <AlertTriangle class="size-4 shrink-0" />
        <span>{{ t('editor.mermaid.error') }}: {{ error }}</span>
      </div>
      <div
        v-else-if="!code.trim()"
        class="mermaid-empty"
        :class="{ 'is-readonly': !editor?.isEditable }"
        @click="startEdit"
      >
        <Workflow class="size-4" />
        {{ t('editor.mermaid.empty') }}
      </div>
      <div v-else ref="svgHost" class="mermaid-svg" v-html="svg" />
      <button
        v-if="editor?.isEditable"
        type="button"
        class="mermaid-edit-btn"
        :title="t('editor.mermaid.edit')"
        @click="startEdit"
      >
        <Pencil class="size-3.5" />
      </button>
    </div>

    <div v-else class="mermaid-editor">
      <textarea
        ref="textarea"
        v-model="draft"
        class="mermaid-textarea"
        spellcheck="false"
        :placeholder="t('editor.mermaid.placeholder')"
        @keydown.escape.prevent.stop="cancelEdit"
        @keydown.meta.enter.prevent="saveEdit"
        @keydown.ctrl.enter.prevent="saveEdit"
      />
      <div class="mermaid-editor-actions">
        <button type="button" class="me-btn" @click="cancelEdit">{{ t('editor.mermaid.cancel') }}</button>
        <button type="button" class="me-btn me-primary" @click="saveEdit">{{ t('editor.mermaid.save') }}</button>
      </div>
    </div>
  </NodeViewWrapper>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { NodeViewWrapper, nodeViewProps } from '@tiptap/vue-3'
import mermaid from 'mermaid'
import { useThemeStore } from '@/stores/theme'
import { IconAlertTriangle as AlertTriangle, IconPencil as Pencil, IconSitemap as Workflow } from '@tabler/icons-vue'

const props = defineProps(nodeViewProps)
const { t } = useI18n()
const themeStore = useThemeStore()

const code = computed(() => (props.node.attrs.code as string) || '')
const svg = ref('')
const error = ref('')
const editing = ref(false)
const draft = ref('')
const svgHost = ref<HTMLElement | null>(null)
const textarea = ref<HTMLTextAreaElement | null>(null)

let seq = 0
let renderGen = 0
let mermaidInited = false

function initMermaid() {
  // securityLevel:'strict' — mermaid 用 DOMPurify 净化输出且禁用 htmlLabels，
  // 阻断 <script>/onerror/data: 等注入；务必在任何 parse/render 之前完成初始化。
  mermaid.initialize({
    startOnLoad: false,
    securityLevel: 'strict',
    theme: themeStore.isDark ? 'dark' : 'default',
  })
  mermaidInited = true
}

async function render() {
  // generation token：仅最新一次 render 可写回，避免并发渲染时旧结果覆盖新结果
  const gen = ++renderGen
  const src = code.value.trim()
  if (!src) {
    svg.value = ''
    error.value = ''
    return
  }
  if (!mermaidInited) initMermaid()
  const id = `mmd-${Date.now()}-${(seq += 1)}`
  try {
    await mermaid.parse(src)
    const { svg: out } = await mermaid.render(id, src)
    if (gen !== renderGen) return
    svg.value = out
    error.value = ''
  } catch (e) {
    // 渲染失败时 mermaid 可能在 body 残留临时容器，显式清理避免堆积
    document.getElementById(id)?.remove()
    document.getElementById(`d${id}`)?.remove()
    if (gen !== renderGen) return
    svg.value = ''
    error.value = (e as Error).message?.split('\n')[0] || String(e)
  }
}

function startEdit() {
  if (!props.editor?.isEditable) return
  draft.value = code.value
  editing.value = true
  nextTick(() => {
    textarea.value?.focus()
    textarea.value?.select()
  })
}
function cancelEdit() {
  editing.value = false
}
function saveEdit() {
  props.updateAttributes({ code: draft.value })
  editing.value = false
}

watch(code, () => render())
watch(() => themeStore.isDark, () => { mermaidInited = false; render() })
onMounted(() => {
  // 提前初始化，确保 securityLevel 在首次渲染前已生效（消除惰性时序隐患）
  if (!mermaidInited) initMermaid()
  render()
})
</script>

<style scoped>
.mermaid-node-view {
  margin: 0.5em 0;
}
.mermaid-render {
  position: relative;
  display: flex;
  justify-content: center;
  padding: 0.75em;
  background: var(--editor-secondary);
  border: 1px solid var(--editor-border);
  border-radius: var(--editor-radius);
  cursor: default;
}
.mermaid-node-view.is-selected .mermaid-render {
  outline: 2px solid var(--editor-accent);
}
.mermaid-svg :deep(svg) {
  max-width: 100%;
  height: auto;
}
.mermaid-empty {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--editor-muted);
  cursor: pointer;
  font-size: 13px;
}
.mermaid-empty.is-readonly {
  cursor: default;
}
.mermaid-error {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--editor-accent);
  font-size: 13px;
  font-family: var(--editor-mono);
}
.mermaid-edit-btn {
  position: absolute;
  top: 6px;
  right: 6px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border: 1px solid var(--editor-border);
  border-radius: 6px;
  background: var(--editor-bg);
  color: var(--editor-muted);
  opacity: 0;
  transition: opacity 0.12s;
  cursor: pointer;
}
.mermaid-render:hover .mermaid-edit-btn {
  opacity: 1;
}
.mermaid-edit-btn:hover {
  color: var(--editor-fg);
}
.mermaid-editor {
  border: 1px solid var(--editor-accent);
  border-radius: var(--editor-radius);
  overflow: hidden;
}
.mermaid-textarea {
  width: 100%;
  min-height: 120px;
  padding: 10px 12px;
  border: none;
  outline: none;
  resize: vertical;
  background: var(--editor-code-bg);
  color: var(--editor-fg);
  font-family: var(--editor-mono);
  font-size: 13px;
  line-height: 1.6;
}
.mermaid-editor-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 6px 8px;
  background: var(--editor-secondary);
}
.me-btn {
  padding: 4px 12px;
  border: 1px solid var(--editor-border);
  border-radius: 6px;
  background: var(--editor-bg);
  color: var(--editor-fg);
  font-size: 13px;
  cursor: pointer;
}
.me-primary {
  background: var(--editor-accent);
  color: var(--editor-accent-fg);
  border-color: var(--editor-accent);
}
</style>
