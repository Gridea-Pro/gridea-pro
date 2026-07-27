<script setup lang="ts">
/**
 * 快捷键面板（参考 editor-vue ShortcutsPanel，按本编辑器实际键位修订）。
 * 遮罩弹窗 + 分组 2 列网格；按键符号按平台显示（mac: ⌘⌥⇧ / 其它: Ctrl Alt Shift）。
 */
import { computed, watch, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: [] }>()

// ESC 关闭：遮罩 div 不可聚焦、keydown 落不进来，必须用 window 级监听（开时挂、关时摘）
function onWindowKey(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}
watch(
  () => props.open,
  (o) => {
    if (o) window.addEventListener('keydown', onWindowKey)
    else window.removeEventListener('keydown', onWindowKey)
  },
)
onBeforeUnmount(() => window.removeEventListener('keydown', onWindowKey))

const { t } = useI18n()

const isMac = /mac/i.test(navigator.platform)
const MOD = isMac ? '⌘' : 'Ctrl'
const ALT = isMac ? '⌥' : 'Alt'
const SHIFT = isMac ? '⇧' : 'Shift'

interface Item {
  name: string
  keys: string[]
}
const groups = computed((): { title: string; items: Item[] }[] => [
  {
    title: t('editor.shortcuts.basic'),
    items: [
      { name: t('editor.shortcuts.undo'), keys: [MOD, 'Z'] },
      { name: t('editor.shortcuts.redo'), keys: [SHIFT, MOD, 'Z'] },
      { name: t('editor.shortcuts.save'), keys: [MOD, 'S'] },
      { name: t('editor.shortcuts.slash'), keys: ['/'] },
    ],
  },
  {
    title: t('editor.shortcuts.textStyle'),
    items: [
      { name: t('editor.bold'), keys: [MOD, 'B'] },
      { name: t('editor.italic'), keys: [MOD, 'I'] },
      { name: t('editor.shortcuts.underline'), keys: [MOD, 'U'] },
      { name: t('editor.strike'), keys: [SHIFT, MOD, 'S'] },
      { name: t('editor.code'), keys: [MOD, 'E'] },
      { name: t('editor.highlight'), keys: [SHIFT, MOD, 'H'] },
      { name: t('editor.sup'), keys: [MOD, '.'] },
      { name: t('editor.sub'), keys: [MOD, ','] },
    ],
  },
  {
    title: t('editor.shortcuts.paragraph'),
    items: [
      { name: t('editor.heading.paragraph'), keys: [MOD, ALT, '0'] },
      { name: t('editor.heading.level', { n: 1 }), keys: [MOD, ALT, '1'] },
      { name: t('editor.heading.level', { n: 2 }), keys: [MOD, ALT, '2'] },
      { name: t('editor.heading.level', { n: 3 }), keys: [MOD, ALT, '3'] },
      { name: t('editor.heading.level', { n: 4 }), keys: [MOD, ALT, '4'] },
      { name: t('editor.heading.level', { n: 5 }), keys: [MOD, ALT, '5'] },
      { name: t('editor.heading.level', { n: 6 }), keys: [MOD, ALT, '6'] },
    ],
  },
  {
    title: t('editor.shortcuts.align'),
    items: [
      { name: t('editor.shortcuts.alignLeft'), keys: [SHIFT, MOD, 'L'] },
      { name: t('editor.shortcuts.alignCenter'), keys: [SHIFT, MOD, 'E'] },
      { name: t('editor.shortcuts.alignRight'), keys: [SHIFT, MOD, 'R'] },
    ],
  },
  {
    title: t('editor.shortcuts.lists'),
    items: [
      { name: t('editor.bulletList'), keys: [SHIFT, MOD, '8'] },
      { name: t('editor.orderedList'), keys: [SHIFT, MOD, '7'] },
      { name: t('editor.taskList'), keys: [SHIFT, MOD, '9'] },
      { name: t('editor.blockquote'), keys: [SHIFT, MOD, 'B'] },
      { name: t('editor.codeBlock'), keys: [MOD, ALT, 'C'] },
    ],
  },
  {
    title: t('editor.shortcuts.insert'),
    items: [
      { name: t('editor.link'), keys: [MOD, 'K'] },
      { name: t('editor.shortcuts.aiAccept'), keys: ['Tab'] },
    ],
  },
])
</script>

<template>
  <!-- 不能 Teleport 到 body：--editor-* 变量定义在 .gridea-tiptap 作用域内，
       teleport 出去变量解析不到 → 背景变非法色（透明）。position:fixed 在子树内同样全屏覆盖。 -->
  <div v-if="open" class="shortcuts-overlay" @click.self="emit('close')">
      <div class="shortcuts-dialog">
        <div class="shortcuts-header">
          <h3>{{ t('editor.shortcuts.title') }}</h3>
          <button type="button" class="close-btn" @click="emit('close')">×</button>
        </div>
        <div class="shortcuts-content">
          <div v-for="(group, gi) in groups" :key="gi" class="shortcut-group">
            <div class="group-title">{{ group.title }}</div>
            <div class="group-items">
              <div v-for="(item, ii) in group.items" :key="ii" class="shortcut-item">
                <span class="shortcut-name">{{ item.name }}</span>
                <span class="shortcut-keys">
                  <kbd v-for="(k, ki) in item.keys" :key="ki" class="key">{{ k }}</kbd>
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
  </div>
</template>

<style scoped>
.shortcuts-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 60;
}
.shortcuts-dialog {
  /* 注意：宿主是 Tailwind v4，--popover 等是完整颜色值，包 hsl() 会变非法色导致透明。
     这里统一用编辑器自有 --editor-* 变量（完整颜色，深浅主题自适应）。 */
  background: var(--editor-popover);
  color: var(--editor-popover-fg);
  border: 1px solid var(--editor-border);
  border-radius: 12px;
  width: 100%;
  max-width: 620px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}
.shortcuts-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 20px;
  border-bottom: 1px solid var(--editor-border);
}
.shortcuts-header h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
}
.close-btn {
  background: none;
  border: none;
  font-size: 22px;
  line-height: 1;
  cursor: pointer;
  color: var(--editor-muted);
}
.close-btn:hover {
  color: var(--editor-fg);
}
.shortcuts-content {
  overflow-y: auto;
  padding: 16px 20px 20px;
}
.shortcut-group + .shortcut-group {
  margin-top: 16px;
}
.group-title {
  font-size: 12px;
  color: var(--editor-muted);
  margin-bottom: 8px;
}
.group-items {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px 24px;
}
.shortcut-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
}
.shortcut-keys {
  display: inline-flex;
  gap: 4px;
}
.key {
  background: var(--editor-secondary);
  border: 1px solid var(--editor-border);
  border-radius: 4px;
  padding: 1px 6px;
  font-size: 12px;
  color: var(--editor-muted);
  font-family: ui-monospace, monospace;
}
</style>
