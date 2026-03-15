<p align="center">
  <img src="build/appicon.png" alt="Gridea Pro" width="80">
</p>

<h1 align="center">Gridea Pro</h1>

<p align="center">
  下一代桌面静态博客写作客户端 —— 像用 Notion 一样写博客。
</p>
<p align="center">
  一个基于 Wails (Go + Vue 3) 的静态博客写作客户端，开源免费！
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-GPL%20v3.0-blue.svg" alt="License"></a>
  <img src="https://img.shields.io/badge/Platform-macOS%20%7C%20Windows%20%7C%20Linux-brightgreen.svg" alt="Platform">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8.svg?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Vue-3.x-4FC08D.svg?logo=vue.js&logoColor=white" alt="Vue 3">
  <a href="README.md"><img src="https://img.shields.io/badge/English-README-orange.svg" alt="English"></a>
</p>

---

> **Gridea Pro** 基于 [Gridea](https://github.com/getgridea/gridea)（10k+ Stars）完全升级重构，感谢原作者 [@EryouHao](https://github.com/EryouHao) 创造了  [Gridea](https://github.com/getgridea/gridea) 这个优秀的项目，帮助了无数人建立自己的博客。原版 Gridea 已停止更新约四年，Gridea Pro 以全新技术栈延续其愿景：**让每个人都能零门槛拥有自己的博客。**

---

## 软件截图
### 文章管理页面

![文章管理页面](screenshots/article-list.png)

### 写作页面

![编辑页面](screenshots/editor.png)

### 闪念页面

![闪念页面](screenshots/memos.png)

### 评论管理页面

![评论管理页面](screenshots/comments.png)

### 菜单管理页面

![菜单管理页面](screenshots/menus.png)

### 分类管理页面

![分类管理页面](screenshots/categories.png)

### 标签管理页面

![标签管理页面](screenshots/tags.png)

### 友链管理页面

![友链管理页面](screenshots/links.png)


## 为什么选择 Gridea Pro？

- **零命令行**：下载安装即可使用，所有操作通过可视化界面完成，无需操作终端命令行
- **一站式工作流**：写作、排版、主题配置、部署全部在一个桌面应用内闭环
- **轻量高性能**：基于 Go + Wails 构建，启动只需 0.5-1 秒，内存占用 30-50MB（Electron 方案通常 150-200MB）
- **国际化支持**：软件支持 11 种国际主流语言

## 对比

| 特性 | Gridea Pro | Hugo | Hexo |
|------|:----------:|:----:|:----:|
| 安装方式 | 下载桌面应用 | CLI 安装 | CLI + Node.js |
| 上手门槛 | 图形界面，开箱即用 | 需要终端操作经验 | 需要 Node.js 和终端经验 |
| 写作环境 | 内置 Monaco 编辑器 | 自行选择外部编辑器 | 自行选择外部编辑器 |
| 主题切换 | 应用内可视化切换 | 修改配置文件 | 修改配置文件 |
| 部署方式 | 一键部署（GUI） | 手动或 CI/CD | 手动或 CI/CD |
| 模板引擎 | Jinja2 / EJS / Go Templates | Go Templates | EJS / Nunjucks |
| 构建速度 | 快（百篇级博客） | 极快 | 中等 |
| 生态规模 | 成长中 | 成熟 | 成熟 |
| 适合人群 | 希望专注写作的用户 | 熟悉命令行的开发者 | 熟悉前端的开发者 |

> Hugo 和 Hexo 是优秀的静态站点生成器，适合需要高度定制和大规模站点的场景。Gridea Pro 专注于为"只想写博客"的用户提供最短路径。

## 功能

### 编辑器

- 内置 Monaco Editor（VS Code 同款引擎），语法高亮、智能补全
- 数学公式（KaTeX）、脚注、任务列表、Emoji、目录生成
- 代码块语法高亮，支持主流编程语言
- 实时预览

### 内容管理

- 文章管理：标签、分类、置顶、草稿、自定义 URL
- Memos（短笔记 / 灵感速记）
- 客户端全文搜索
- CJK 精确字数统计与阅读时间估算

### 主题系统

- 三模板引擎可选：Jinja2（Pongo2）、EJS、Go Templates
- 应用内可视化主题切换与参数配置
- 主题 config.json 声明式配置面板
- 支持暗色模式、响应式布局

### 部署与 SEO

- 一键部署：GitHub Pages、Vercel、Netlify、Gitee、Coding、SFTP
- 自动生成 sitemap.xml、robots.txt、RSS/Atom Feed
- Open Graph、Twitter Card 等社交分享 Meta 标签
- 自定义域名支持

### 社交与集成

- 评论系统集成（Gitalk、Disqus 等）
- MCP（AI 集成）支持
- 11 种界面语言

## 快速开始

### 下载安装

从 [Releases](https://github.com/Gridea-Pro/gridea-pro/releases) 页面下载对应平台的安装包：

| 平台 | 格式 |
|------|------|
| macOS | `.dmg` |
| Windows | `.exe` 安装包 |
| Linux | `.AppImage` / `.deb` |

下载后双击安装，打开即可开始写作。

### 从源码构建

确保已安装 Go 1.21+、Node.js 18+、[Wails v2](https://wails.io/)。

```bash
# 克隆仓库
git clone https://github.com/Gridea-Pro/gridea-pro.git
cd gridea-pro

# 安装前端依赖
cd frontend && npm install && cd ..

# 开发模式运行
wails dev

# 构建生产版本
wails build
```

## 主题开发

Gridea Pro 支持三种模板引擎开发主题，推荐使用 Jinja2（Pongo2）：

```
my-theme/
├── config.json          # 主题配置声明
├── templates/
│   ├── index.html       # 首页
│   ├── post.html        # 文章页
│   ├── tag.html         # 标签页
│   ├── archives.html    # 归档页
│   └── partials/        # 可复用组件
└── assets/
    ├── styles/          # CSS
    └── scripts/         # JS
```

详细开发文档请参阅 [主题开发指南](https://github.com/Gridea-Pro/gridea-pro/wiki/Theme-Development)。

## 路线图

- [ ] 编辑器优化，替换 tiptap 编辑器
- [ ] 主题市场（在线浏览与一键安装）
- [ ] 图片管理与图床集成
- [ ] 文章版本历史
- [ ] 多博客实例管理
- [ ] 插件系统

## 参与贡献

欢迎提交 Issue 和 Pull Request。参与贡献前请阅读 [CONTRIBUTING.md](CONTRIBUTING.md)。

```bash
# Fork 并克隆仓库
git clone https://github.com/<your-username>/gridea-pro.git

# 创建功能分支
git checkout -b feature/your-feature

# 开发完成后提交 PR
```

## 致谢

- [Gridea](https://github.com/getgridea/gridea) — 原版项目，感谢 [@EryouHao](https://github.com/EryouHao) 的开创性工作
- [Wails](https://wails.io/) — Go 桌面应用框架
- [Vue 3](https://vuejs.org/) — 前端框架
- [Monaco Editor](https://microsoft.github.io/monaco-editor/) — 代码编辑器引擎
- [Pongo2](https://github.com/flosch/pongo2) — Go 实现的 Jinja2 模板引擎
- [KaTeX](https://katex.org/) — 数学公式渲染
- [Tailwind CSS](https://tailwindcss.com/) — 原子化 CSS 框架

## 开源协议

[GPL-3.0](LICENSE) &copy; Gridea Pro
