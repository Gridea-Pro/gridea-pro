# AGENTS.md

本文件用于记录 AI 助手（如 GitHub Copilot、Cursor 等）在项目中需要遵循的约定和示例。

## Commit 信息规范

遵循 [Conventional Commits](https://www.conventionalcommits.org/zh-hans/) 规范，**使用中文撰写 commit 信息**：

```
<类型>(<范围>): <简短描述> (#PR号)

[可选的正文：问题描述、原因分析、解决方案]

[Closes #关联Issue号]
```

**注意**：`Closes #XX` 仅在解决 Issue 时才需要添加。如果不是解决 Issue，则不必附上此标签。

### 类型

| 类型 | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `refactor` | 代码重构（不改变功能） |
| `docs` | 文档更新 |
| `style` | 代码格式调整（不影响逻辑） |
| `perf` | 性能优化 |
| `test` | 测试相关 |
| `chore` | 构建、CI、依赖等杂项 |

### 范围

范围用于指定改动的模块或文件，例如：
- `window` - 窗口相关
- `editor` - 编辑器相关
- `theme` - 主题相关
- `i18n` - 国际化相关
- `oauth` - OAuth 认证相关

### 示例

```
fix(window): 移除 ShowPreferences 中多余的 WindowCenter 调用 (#17)

点击设置后窗口会自动回到屏幕中央，而非保持在之前的位置。
这是因为 ShowPreferences() 每次调用都会执行 runtime.WindowCenter()，
将窗口强制居中，覆盖了用户手动调整的位置。

移除多余的 WindowCenter 调用即可解决问题。

Closes #10
```

---

## Pull Request 创建示例

以下是创建 PR 时的描述模板和示例：

**重要：** 在 PR 描述的末尾必须添加 `Closes #XX` 标签，这样当 PR 被合并时，GitHub 会自动关闭关联的 Issue。

### PR 描述模板

```markdown
### 问题描述

[简要描述问题或需求背景]

### 解决方案

1. **[方案要点一]** - [具体说明]
2. **[方案要点二]** - [具体说明]
3. **[方案要点三]** - [具体说明]

### 改动范围

- `path/to/file1.go` - [改动说明]
- `path/to/file2.ts` - [改动说明]

### 测试

[测试方法和验证结果]
关联 Issue: #XX
分支已推送至 origin/fix/issue-XX。

Closes #XX
```

### PR 描述示例

以下是一个真实的 PR 描述示例：

```markdown
### 问题描述

旧版 Gridea 的主题使用 `.less` 文件编写样式，但 Gridea Pro 之前依赖外部的 `lessc` 命令（需手动安装 Node.js 和 lessc）。如果用户未安装 lessc，LESS 编译会静默失败，导致浏览器访问时 CSS 返回 404。

### 解决方案

1. **移除外部依赖** - 改用纯 Go 实现的 [lessgo](https://github.com/typomedia/lessgo) 库（基于 goja 运行 JavaScript 版本的 Less.js），无需用户安装任何外部工具
2. **修复 Import 路径解析** - lessgo 库本身存在 `@import` 语句的路径解析 bug（Less.js 在 goja 运行时无法自动添加 `.less` 扩展名），通过添加 `inlineLess` 函数在编译前将所有 `@import` 语句展开为内联内容解决
3. **兼容多种主题结构** - `BundleCSS` 现在支持多种模板引擎的 head 文件位置（`partials`、`includes`、`_blocks` 下的 `head.html` 或 `head.ejs`）

### 改动范围

- `backend/internal/engine/asset_manager.go` - 核心修复
- `go.mod` / `go.sum` - 新增 lessgo 依赖

### 测试

已在本地使用旧版 Gridea 主题（simple）验证 LESS 编译功能正常工作。
关联 Issue: #14。

Closes #14
```

---

## Issue 处理流程

处理 Issue 时需遵循以下标准流程：

### 1. 切换到 main 分支

```bash
git checkout main
```

### 2. 拉取上游最新代码

```bash
git pull upstream main
# 如果没有配置 upstream，先添加：
git remote add upstream https://github.com/Gridea-Pro/gridea-pro.git
```

### 3. 从 main 分支创建新分支

**重要：必须从 main 分支创建新分支**

```bash
git checkout -b <类型>/<issue号或描述>
```

**例外**：如果用户明确指出在 main 上工作（如"直接在 main 修改"、"不需要开新分支"等），则直接在 main 分支进行修改，无需创建新分支。

分支命名规范（参考 CONTRIBUTING.md）：

| 前缀 | 用途 | 示例 |
|------|------|------|
| `feat/` | 新功能 | `feat/dark-mode` |
| `fix/` | Bug 修复 | `fix/issue-24` |
| `docs/` | 文档 | `docs/api-reference` |
| `refactor/` | 代码重构 | `refactor/renderer` |
| `chore/` | 工具、CI、依赖 | `chore/update-deps` |

### 4. 处理 Issue

1. 阅读并理解 Issue 描述
2. 探索相关代码
3. 实现修复或功能
4. 运行 lint 和 format 确保代码规范
5. 提交代码（遵循 Commit 信息规范）
6. 推送分支并创建 PR

---

## 其他约定

（待补充）