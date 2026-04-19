<p align="center">
  <img src="build/appicon.png" alt="Gridea Pro" width="100">
</p>

<h1 align="center">Gridea Pro</h1>

<p align="center">
  下一代桌面静态博客写作客户端 —— 像用 Notion 一样写博客。
</p>
<p align="center">
  一个基于 Wails (Go + Vue 3) 的静态博客写作客户端，永久开源免费！
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-GPL%20v3.0-blue.svg" alt="License"></a>
  <a href="https://github.com/Gridea-Pro/gridea-pro/releases"><img src="https://img.shields.io/github/v/release/Gridea-Pro/gridea-pro?color=brightgreen" alt="Release"></a>
  <img src="https://img.shields.io/badge/Platform-macOS%20%7C%20Windows%20%7C%20Linux-brightgreen.svg" alt="Platform">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8.svg?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Vue-3.x-4FC08D.svg?logo=vue.js&logoColor=white" alt="Vue 3">
  <a href="README.md"><img src="https://img.shields.io/badge/English-README-orange.svg" alt="English"></a>
</p>

<p align="center">
  <a href="README-en.md">English</a> · <b>简体中文</b>
</p>

---

<p align="center">
  <b>预览站点</b>&nbsp;&nbsp;👉&nbsp;&nbsp;<a href="https://are.ink/">are.ink</a>
</p>

---

**Gridea Pro** 基于 [Gridea](https://github.com/getgridea/gridea)（10k+ Stars）完全重写，使用 Go + Wails + Vue 3 全新构建。原版 Gridea 已停止维护约四年，Gridea Pro 延续其核心愿景：**让每个人都能零门槛拥有自己的博客。**

感谢原作者 [@EryouHao](https://github.com/EryouHao) 创造了  [Gridea](https://github.com/getgridea/gridea) 这个优秀的项目，帮助了无数人建立自己的博客。

---

## Gridea Pro 官方交流群

扫描下方二维码加入微信交流群，反馈问题、交流使用心得：

<p align="center">
  <img src="images/wechatgroup.jpg" alt="微信交流群" width="260">
</p>

> 二维码会定期更新，如已过期请提交 Issue 获取最新二维码。

---

## 软件截图

<table>
  <tr>
    <td align="center"><b>文章管理</b></td>
    <td align="center"><b>写作编辑器</b></td>
  </tr>
  <tr>
    <td><img src="images/post.png" alt="文章管理"></td>
    <td><img src="images/editor.png" alt="写作编辑器"></td>
  </tr>
  <tr>
    <td align="center"><b>闪念速记</b></td>
    <td align="center"><b>评论管理</b></td>
  </tr>
  <tr>
    <td><img src="images/memos.png" alt="闪念速记"></td>
    <td><img src="images/comments.png" alt="评论管理"></td>
  </tr>
  <tr>
    <td align="center"><b>菜单管理</b></td>
    <td align="center"><b>分类管理</b></td>
  </tr>
  <tr>
    <td><img src="images/menus.png" alt="菜单管理"></td>
    <td><img src="images/categories.png" alt="分类管理"></td>
  </tr>
  <tr>
    <td align="center"><b>标签管理</b></td>
    <td align="center"><b>主题管理</b></td>
  </tr>
  <tr>
    <td><img src="images/tags.png" alt="标签管理"></td>
    <td><img src="images/themes.png" alt="主题管理"></td>
  </tr>
</table>

---

## 为什么选 Gridea Pro？

Hugo、Hexo、Jekyll 都很优秀，但它们面向开发者，且上手门槛不低——你需要装 Node.js、配置命令行、手动写部署脚本。Gridea Pro 走另一条路：**下载即用，所有操作都在 GUI 里完成，同时集成 AI 能力，顺应时代潮流。**

| | Gridea Pro | Hugo | Hexo |
|---|:---:|:---:|:---:|
| 安装方式 | 下载桌面应用 | CLI | CLI + Node.js |
| 上手门槛 | 开箱即用 | 需要终端经验 | 需要 Node.js + 终端 |
| 写作环境 | 内置 Monaco 编辑器 | 自选外部编辑器 | 自选外部编辑器 |
| 主题切换 | 应用内可视化切换 | 修改配置文件 | 修改配置文件 |
| 一键部署 | ✅ GUI 操作 | ❌ 手动 / CI | ❌ 手动 / CI |
| AI 集成 | ✅ MCP + 内置模型 | ❌ | ❌ |
| 内存占用 | ~30–50 MB | — | — |
| 模板引擎 | Jinja2 / EJS / Go | Go Templates | EJS / Nunjucks |

> Hugo 和 Hexo 是适合开发者的强大工具，Gridea Pro 专为"只想好好写博客"的人设计。

---

## 核心功能

### ✍️ 写作与编辑

- **Monaco Editor**（VS Code 同款引擎）：语法高亮、智能补全、Vim/Emacs 键位支持
- Markdown 扩展：数学公式（KaTeX）、脚注、任务列表、Emoji、自动目录
- 代码块语法高亮，支持主流编程语言
- 实时预览，所见即所得
- CJK 精确字数统计与阅读时间估算

### 📋 内容管理

- 文章：标签、分类、置顶、草稿、自定义 URL Slug、特色图片
- **闪念（Memos）**：灵感速记，支持 `#标签` 语法、图片附件、热力图统计
- 友情链接、导航菜单管理
- 评论管理：支持回复与删除（需评论系统配合）
- 客户端全文搜索

### 🎨 主题系统

- **9 款内置主题**，点击即切换：`amore`、`flavor`、`claudo`、`letters`、`inotes`、`fly`、`simple`、`notes` 等
- 三种模板引擎：**Jinja2（Pongo2）**、**EJS**、**Go Templates**
- 主题参数可视化配置（由 `config.json` 声明生成表单），无需手动改文件
- 支持暗色模式、响应式布局

### 🚀 部署

- 一键部署，支持 6 大平台：GitHub Pages、Vercel、Netlify、Gitee、Coding、SFTP/FTP
- 内置纯 Go 实现的 Git 引擎，**不依赖系统 Git**，同步更稳定
- CDN 媒体文件自动上传：部署时将图片等资源同步到 GitHub 仓库，支持自定义保存路径
- 自定义域名（CNAME）支持

### 🔍 SEO

- 自动生成 `sitemap.xml`（含图片元数据）、`robots.txt`、`RSS/Atom Feed`
- Open Graph、Twitter Card 等社交分享 Meta 标签
- JSON-LD 结构化数据
- Google Analytics、百度统计、Google Search Console 验证码
- 自定义 `<head>` 代码注入

### 💬 评论系统集成

内置 7 种评论系统，勾选即启用，无需手动引入代码：

<table>
  <tr>
    <td align="center"><b>Gitalk</b></td>
    <td align="center"><b>Giscus</b></td>
    <td align="center"><b>Disqus</b></td>
    <td align="center"><b>Valine</b></td>
    <td align="center"><b>Waline</b></td>
    <td align="center"><b>Twikoo</b></td>
    <td align="center"><b>Cusdis</b></td>
  </tr>
</table>

### 🤖 AI 集成（MCP）

Gridea Pro 实现了 [MCP（Model Context Protocol）](https://modelcontextprotocol.io/) 协议，让 AI 助手（Claude、Cursor 等）可以直接操作你的博客：

**25+ MCP 工具**，覆盖博客管理全流程：

| 类别 | 工具 |
|------|------|
| 文章 | 列表、查看、创建、更新、删除 |
| 闪念 | 列表、创建、更新、删除、热力图统计 |
| 标签 / 分类 | 完整 CRUD |
| 菜单 / 友链 | 完整 CRUD |
| 评论 | 列表、回复、删除 |
| 主题 | 列表主题、查看 / 更新主题配置 |
| 站点 | 查看 / 更新全局设置 |
| 渲染 & 部署 | 触发渲染、触发部署（需显式开启） |

**内置 5 个工作流提示词**：写作助手、闪念整理成文、内容审查、站点健康检查、文章翻译。

**MCP 配置示例：**

```json
{
  "mcpServers": {
    "gridea-pro": {
      "command": "/path/to/gridea-pro",
      "args": ["--mcp"],
      "env": {
        "GRIDEA_SITE_DIR": "/path/to/your/site",
        "DEPLOY_ENABLED": "false"
      }
    }
  }
}
```

> `DEPLOY_ENABLED=true` 时，AI 可直接触发部署；默认关闭，需手动确认。

**内置 AI 模型**：无需配置 API Key 即可使用内置免费模型（每日限额 20 次）；也支持接入 13 种自定义模型服务商：OpenAI、Anthropic、DeepSeek、Gemini、Kimi、Qwen、GLM 等。

### 📱 PWA 支持

- 一键开启 Progressive Web App
- 可配置：应用名称、图标、主题色、屏幕方向等
- 用户可将博客"安装"到手机/桌面，离线访问

### 🌍 国际化

软件界面支持 **11 种语言**：

`简体中文` · `繁体中文` · `English` · `日本語` · `한국어` · `Deutsch` · `Español` · `Français` · `Italiano` · `Português (BR)` · `Русский`

---

## 快速开始

### 下载安装

从 [Releases](https://github.com/Gridea-Pro/gridea-pro/releases) 页面下载对应平台的安装包：

| 平台 | 安装包 |
|------|--------|
| macOS | `.dmg` |
| Windows | `.exe` |
| Linux | `.AppImage` / `.deb` / `.rpm` |

下载后双击安装，打开即可开始写作。

> **AppImage 注意事项**：部分 Linux 发行版（如新版 Fedora）默认未预装 FUSE。如遇到 "AppImages require FUSE" 报错，请手动安装：`sudo dnf install fuse-libs`（Fedora）或 `sudo apt install libfuse2`（Ubuntu）。

**从旧版 Gridea 迁移**：在应用内将「站点目录」指向原来的 Gridea 数据目录，启动后自动迁移，无需手动操作。

### 从源码构建

**前置要求**：Go 1.22+、Node.js 18+、[Wails v2](https://wails.io/)

```bash
git clone https://github.com/Gridea-Pro/gridea-pro.git
cd gridea-pro

cd frontend && npm install && cd ..

# 开发模式（热重载）
wails dev

# 构建生产版本
wails build
```

---

## 主题开发

Gridea Pro 支持三种模板引擎，推荐使用 Jinja2（Pongo2）。主题目录结构：

```
my-theme/
├── config.json          # 主题配置声明（自动生成可视化配置面板）
├── templates/
│   ├── index.html       # 首页
│   ├── post.html        # 文章详情页
│   ├── tag.html         # 标签页
│   ├── archives.html    # 归档页
│   └── partials/        # 可复用组件
└── assets/
    ├── styles/
    └── scripts/
```

详细开发文档请参阅 [主题开发指南](https://github.com/Gridea-Pro/gridea-pro/wiki/Theme-Development)。

---


## 参与贡献

欢迎提交 Issue 和 Pull Request。参与贡献前请阅读 [CONTRIBUTING.md](CONTRIBUTING.md)。

```bash
# Fork 并克隆仓库
git clone https://github.com/<your-username>/gridea-pro.git

# 创建功能分支
git checkout -b feature/your-feature

# 开发完成后提交 PR
```

---

## 致谢

- [Gridea](https://github.com/getgridea/gridea) — 原版项目，感谢 [@EryouHao](https://github.com/EryouHao) 的创造
- [Wails](https://wails.io/) — Go 桌面应用框架
- [Vue 3](https://vuejs.org/) — 前端框架
- [Monaco Editor](https://microsoft.github.io/monaco-editor/) — 编辑器引擎
- [Pongo2](https://github.com/flosch/pongo2) — Go 实现的 Jinja2 模板引擎
- [KaTeX](https://katex.org/) — 数学公式渲染
- [Tailwind CSS](https://tailwindcss.com/) — CSS 框架

---

## 开源协议

[GPL-3.0](LICENSE) &copy; Gridea Pro
