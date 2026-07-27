# 贡献指南

首先，感谢你考虑为 Gridea Pro 做贡献！每一份贡献都让这个项目变得更好。

## 目录

- [行为准则](#行为准则)
- [如何贡献](#如何贡献)
- [开发环境搭建](#开发环境搭建)
- [项目结构](#项目结构)
- [工作流程](#工作流程)
- [提交信息规范](#提交信息规范)
- [Pull Request 指南](#pull-request-指南)
- [主题开发](#主题开发)
- [报告 Bug](#报告-bug)
- [提出功能建议](#提出功能建议)
- [社区](#社区)

## 行为准则

本项目遵循 [Contributor Covenant 行为准则](https://www.contributor-covenant.org/zh-cn/version/2/1/code_of_conduct/)。参与贡献即表示你同意遵守此准则。如遇到不当行为，请通过 [Issues](https://github.com/Gridea-Pro/gridea-pro/issues) 反馈。

## 如何贡献

贡献不仅仅是写代码，还有很多方式可以参与：

- **报告 Bug** — 发现问题了？[提交 Bug 报告](https://github.com/Gridea-Pro/gridea-pro/issues/new?template=bug_report.yml)。
- **功能建议** — 有好想法？[提交功能建议](https://github.com/Gridea-Pro/gridea-pro/issues/new?template=feature_request.yml)。
- **制作主题** — 设计一款主题并分享给社区。
- **完善文档** — 修正错别字、改进说明、补充示例。
- **翻译** — 帮助 Gridea Pro 支持更多语言。
- **回答问题** — 在 [Discussions](https://github.com/Gridea-Pro/gridea-pro/discussions) 中帮助其他用户。
- **推广项目** — 给仓库点个 Star、写篇博客、推荐给朋友。

## 开发环境搭建

### 前置要求

- [Go](https://go.dev/dl/) 1.22+
- [Node.js](https://nodejs.org/) 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2
- [Git](https://git-scm.com/)

### 开始开发

```bash
# 1. 在 GitHub 上 Fork 仓库，然后克隆你的 Fork
git clone https://github.com/<你的用户名>/gridea-pro.git
cd gridea-pro

# 2. 安装前端依赖
cd frontend && npm install && cd ..

# 3. 以开发模式运行
wails dev
```

应用会启动并启用前端热重载。Go 代码的修改需要重新启动。

### 构建

```bash
# 为当前平台构建
wails build
```

## 项目结构

```
gridea-pro/
├── main.go              # 应用入口
├── app.go               # Wails 应用逻辑和绑定
├── frontend/            # Vue 3 + Vite + Tailwind CSS
│   ├── src/
│   │   ├── components/  # Vue 组件
│   │   ├── views/       # 页面视图
│   │   └── ...
│   └── package.json
├── build/               # 构建资源（图标、安装器配置）
├── themes/              # 内置主题
└── .github/
    └── workflows/       # CI/CD（GitHub Actions）
```

## 工作流程

我们使用 **GitHub Flow** — 简洁的分支工作流：

1. **从 `main` 创建分支：**
   ```bash
   git checkout -b feat/my-feature
   # 或
   git checkout -b fix/some-bug
   ```

2. **进行修改**，保持每次提交聚焦且清晰。

3. **推送分支：**
   ```bash
   git push origin feat/my-feature
   ```

4. **发起 Pull Request**，目标分支为 `main`。

### 分支命名规范

| 前缀 | 用途 | 示例 |
|------|------|------|
| `feat/` | 新功能 | `feat/dark-mode` |
| `fix/` | Bug 修复 | `fix/image-upload-crash` |
| `docs/` | 文档 | `docs/theme-api` |
| `refactor/` | 代码重构 | `refactor/renderer` |
| `chore/` | 工具、CI、依赖 | `chore/update-deps` |

## 提交信息规范

我们遵循 [Conventional Commits](https://www.conventionalcommits.org/zh-hans/) 规范：

```
<类型>(<范围>): <简短描述>

[可选的正文]
```

**类型：** `feat`、`fix`、`docs`、`style`、`refactor`、`perf`、`test`、`chore`、`ci`

**示例：**

```
feat(editor): 添加图片拖拽上传支持
fix(renderer): 修复 Jinja2 循环变量作用域问题
docs(readme): 更新安装说明
chore(ci): 在发布流程中添加 Linux ARM 构建
```

## Pull Request 指南

- **每个 PR 只做一件事。** 保持 PR 聚焦——单个功能或 Bug 修复。
- **说明做了什么、为什么。** 提供上下文，附上截图或 GIF 辅助说明。
- **关联相关 Issue。** 在 PR 描述中使用 `Closes #123` 或 `Fixes #456`。
- **确保能构建通过。** 提交前在本地运行 `wails build` 验证。
- **耐心等待。** 我们会尽快审核 PR。反馈意见是为了改进代码，而不是批评。

### PR 标题格式

与提交信息保持一致：

```
feat(editor): 添加图片拖拽上传支持
```

## 主题开发

Gridea Pro 支持三种模板引擎：**Jinja2 (Pongo2)**、**EJS** 和 **Go Templates**。制作主题是参与贡献的最佳方式之一。

### 快速开始

1. 查阅 [主题开发文档](https://gridea.pro/docs/themes) 了解完整 API 参考。
2. 以现有主题（如 `flavor`）作为起点。
3. 一个主题由模板、样式和 `theme.toml` 配置文件组成。

### 主题结构

```
my-theme/
├── theme.toml           # 主题元数据和配置
├── templates/
│   ├── index.html       # 首页
│   ├── post.html        # 文章页
│   ├── tag.html         # 标签归档页
│   └── ...
├── assets/
│   ├── styles/
│   └── scripts/
└── preview.png          # 主题预览图（800×600）
```

### 提交主题

主题准备好后，分享给社区：

1. 将主题发布为独立的 GitHub 仓库。
2. 在 [Show and Tell](https://github.com/Gridea-Pro/gridea-pro/discussions/categories/show-and-tell) 讨论区发布帖子。
3. 受欢迎的社区主题可能会被收录到官网展示。

## 报告 Bug

请使用 [Bug 报告模板](https://github.com/Gridea-Pro/gridea-pro/issues/new?template=bug_report.yml)。一份好的 Bug 报告应包含：

- Gridea Pro 版本和操作系统
- 复现步骤
- 期望行为 vs 实际行为
- 截图或错误日志（如有）

## 提出功能建议

请使用 [功能建议模板](https://github.com/Gridea-Pro/gridea-pro/issues/new?template=feature_request.yml)。有价值的功能建议应包含：

- 你想解决的问题
- 你建议的解决方案
- 你考虑过的替代方案

对于还在早期阶段的想法，建议先在 [Discussions](https://github.com/Gridea-Pro/gridea-pro/discussions) 发帖讨论——在提交正式请求之前，这是获取社区反馈的好方式。

## 社区

- 💬 [GitHub Discussions](https://github.com/Gridea-Pro/gridea-pro/discussions) — 提问、交流、分享想法
- 🐛 [Issue Tracker](https://github.com/Gridea-Pro/gridea-pro/issues) — Bug 报告和功能建议
- 🌐 [官方网站](https://gridea.pro) — 下载、文档和主题

---

## 提交前自查（CI 会强制检查）

PR 提交后会自动运行检查，**任一不通过都无法合并**，请提交前在本地自查：

- `gofmt -w` 已格式化（CI 拒绝未格式化的 Go 代码）
- `go vet ./...`、`go test ./backend/...` 通过
- 前端 `npm run lint` 无 error
- 新增/修改的 UI 文案已补全 **全部 12 个语言**（`frontend/src/locales/`），
  可运行 `python3 scripts/check_i18n.py frontend/src/locales` 自检；en.json 不要遗留中文
- 本地 `wails dev` 跑通并自测过改动

检查全部通过后，PR 进入维护者 review，批准后方可合并。

---

感谢你帮助 Gridea Pro 变得更好！🎉
