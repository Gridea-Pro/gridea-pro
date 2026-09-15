# 编辑器扩展/组件编写规范

保证多人/多轮生成的一致性。所有 `extensions/`、`ui/`、`nodeviews/` 文件遵循本规范。

## 通用
- TS strict、无 `any`（必要时 `unknown` + 收窄）；无分号、单引号、2 空格、`printWidth 100`（与项目 prettier 一致）。
- 导入顺序：`vue` → `@tiptap/*` → 项目内 `@/` → 相对路径。
- 颜色/间距只用 `--editor-*` 变量（见 `styles/theme.css`），不写死颜色、不用 `dark:` 工具类。
- UI 文案走 i18n：`const { t } = useI18n()`，key 在 `editor.*` 命名空间。

## 自定义扩展（节点/标记）
- 用 `Node.create`/`Mark.create`/`Extension.create`。
- **必须**实现 Markdown 往返：通过 `@tiptap/markdown` 的存储钩子（见 `markdown/serializer.ts` 的注册方式）声明该节点的 `serialize`/`parse`；自定义围栏/语法节点在此对齐 Gridea 方言（见 `docs/tiptap-editor-开发计划.md` §4）。
- 含 `renderHTML`/`parseHTML`，保证 HTML 兜底正确。
- 需要交互的节点用 `VueNodeViewRenderer(() => import('./nodeviews/XView.vue'))`，NodeView 内用 `NodeViewWrapper`/`NodeViewContent`。

## 移植参考
- React 源码在 `/tmp/ctzhian-tiptap-src/src/extension/`（节点的 `renderMarkdown/parseMarkdown` 逻辑可直接搬，UI 改 Vue）。
- 数据（斜杠命令清单、emoji 列表等）直接照搬，不删减。
- 正文 CSS 照搬 `@ctzhian/index.css`，把 `--mui-palette-*` 交给 `theme.css` 的别名（已映射）。

## NodeView 组件（`nodeviews/*.vue`）
```vue
<template>
  <NodeViewWrapper class="...">
    <!-- 可编辑文本用 <NodeViewContent /> -->
  </NodeViewWrapper>
</template>
<script setup lang="ts">
import { NodeViewWrapper, NodeViewContent, nodeViewProps } from '@tiptap/vue-3'
const props = defineProps(nodeViewProps)
</script>
```

## UI 组件（`ui/*.vue`）
- 用 Radix Vue 原语：`DropdownMenu*` / `Popover*` / `Dialog*`（可复用 `@/components/ui/*` 的封装）。
- 按钮统一用 `ToolbarButton.vue`（ghost icon + i18n title）。
- 菜单/弹层挂载到 `Teleport` 或 Radix `Portal`，定位用 `@floating-ui` 或 Radix 内建。
