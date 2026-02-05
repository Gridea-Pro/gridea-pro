# Gridea Pro - Wails 版本

本项目已从 Electron 架构重构为 Wails (Go + Vue 3) 架构，完整保留所有前端 Vue 代码和功能。

## 架构说明

### 前端（Frontend）
- **框架**: Vue 3 + TypeScript + Vite
- **UI 库**: Ant Design Vue 4.0
- **状态管理**: Pinia
- **样式**: Tailwind CSS + Less
- **路由**: Vue Router
- **编辑器**: Monaco Markdown
- **国际化**: Vue I18n

所有前端代码完全保留，无需任何修改。

### 后端（Backend）
- **语言**: Go 1.22+
- **框架**: Wails v2.9.2
- **数据存储**: JSON 文件（与原版兼容）
- **YAML 解析**: gopkg.in/yaml.v3

后端完全替换 Electron + Node.js，使用 Go 重写所有业务逻辑。

## 目录结构

```
.
├── backend/           # Go 后端代码
│   ├── app.go        # 主应用和事件处理
│   ├── models.go     # 数据模型
│   ├── posts.go      # 文章管理
│   ├── tags.go       # 标签管理
│   └── menus.go      # 菜单管理
├── src/              # Vue 前端代码（完整保留）
│   ├── views/        # 页面组件
│   ├── components/   # UI 组件
│   ├── stores/       # Pinia 状态管理
│   ├── router/       # 路由配置
│   ├── assets/       # 静态资源
│   └── wails-electron-bridge.ts  # Wails 兼容桥
├── wailsjs/          # Wails 自动生成的 JS 绑定（由 wails dev/build 生成）
├── main.go           # Go 主入口
├── wails.json        # Wails 项目配置
└── go.mod            # Go 依赖管理
```

## 开发环境要求

### 必需安装

1. **Go 1.22+**
   ```bash
   # macOS
   brew install go
   
   # 或从官网下载
   # https://go.dev/dl/
   ```

2. **Wails CLI**
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

3. **Node.js 18+** 和 **npm/pnpm**（前端开发）

### 可选工具

- **Go 依赖代理**（国内推荐）：
  ```bash
  go env -w GOPROXY=https://goproxy.cn,direct
  ```

## 安装和运行

### 1. 安装依赖

```bash
# 安装前端依赖
npm install
# 或
pnpm install

# 安装 Go 依赖
go mod tidy
```

### 2. 开发模式

```bash
# 启动 Wails 开发服务（自动启动前端 Vite DevServer）
wails dev

# 或使用 npm script
npm run wails:dev
```

开发模式下：
- 前端运行在 `http://127.0.0.1:5173`（Vite DevServer）
- 后端 Go 代码热重载
- 前端 Vue 代码热重载
- Wails 窗口自动打开并加载前端

### 3. 构建生产版本

```bash
# 构建完整应用
wails build

# 或使用 npm script
npm run wails:build
```

构建产物位于 `build/bin/` 目录：
- **macOS**: `Gridea Pro.app`
- **Windows**: `Gridea Pro.exe`
- **Linux**: `gridea-pro`

## 功能映射

### 已完成功能

| 原 Electron 功能 | Wails 实现 | 状态 |
|---|---|---|
| 文章管理（增删改查） | `backend/posts.go` | ✅ 完成 |
| 标签管理 | `backend/tags.go` | ✅ 完成 |
| 菜单管理 | `backend/menus.go` | ✅ 完成 |
| 设置管理 | `app.go` | ✅ 完成 |
| 站点目录选择 | `OpenFolderDialog()` | ✅ 完成 |
| 预览服务器 | Go `http.Server` | ✅ 完成 |
| 图片上传 | `posts.go` | ✅ 完成 |
| 打开外部链接 | `runtime.BrowserOpenURL` | ✅ 完成 |
| 事件通信 | `runtime.EventsOn/Emit` | ✅ 完成 |

### 待实现功能

| 功能 | 优先级 | 说明 |
|---|---|---|
| 主题管理 | 高 | 主题列表、切换、自定义配置 |
| 静态站点渲染 | 高 | Markdown → HTML + EJS 模板 + Less 样式 |
| 发布功能（Git） | 中 | 基于 go-git 实现 |
| 发布功能（SFTP） | 中 | SSH/SFTP 客户端 |
| 发布功能（Netlify） | 低 | HTTP API 调用 |
| 自动更新 | 低 | 基于 GitHub Releases |

## 事件系统

### 前端到后端

前端使用 `window.electronAPI` 发送事件（兼容层自动转换为 Wails 事件）：

```typescript
// 发送事件
window.electronAPI.send('app-site-reload')

// 监听回复
window.electronAPI.on('app-site-loaded', (event, data) => {
  console.log('站点数据加载完成', data)
})

// 调用后端方法（invoke）
const folders = await window.electronAPI.invoke('open-folder-dialog')
```

### 后端处理

在 `backend/app.go` 中注册事件处理器：

```go
runtime.EventsOn(ctx, "app-site-reload", func(...interface{}) {
    data := a.LoadSite()
    runtime.EventsEmit(ctx, "app-site-loaded", data)
})
```

### 兼容桥

`src/wails-electron-bridge.ts` 提供完整的 Electron API 兼容：

- `getLocale()` - 获取系统语言
- `on(channel, listener)` - 监听事件
- `once(channel, listener)` - 监听一次性事件
- `off(channel)` - 移除监听
- `removeAllListeners(channel)` - 移除所有监听
- `send(channel, ...args)` - 发送事件
- `invoke(channel, ...args)` - 调用方法并等待返回
- `openExternal(url)` - 打开外部链接

## 数据存储

与原 Electron 版本完全兼容，数据存储在：

- **配置目录**: `~/.gridea/`
  - `config.json` - 源文件夹配置
  - `output/` - 构建输出目录

- **站点目录**: `~/Documents/Gridea/` (默认)
  - `posts/` - Markdown 文章
  - `config/` - JSON 配置文件
    - `posts.json` - 文章索引
    - `tags.json` - 标签列表
    - `menus.json` - 菜单列表
    - `setting.json` - 远程发布配置
    - `theme.json` - 主题配置
  - `post-images/` - 文章图片
  - `images/` - 站点图片（头像等）
  - `themes/` - 主题模板
  - `static/` - 静态文件

## 开发技巧

### 查看 Wails 日志

```bash
# 开发模式下，控制台会显示：
# - Go 后端日志（log.Printf）
# - 前端控制台日志（通过 renderer-log 事件）
# - Wails 系统日志
```

### 调试前端

开发模式下，Wails 窗口内可以打开开发者工具（类似 Electron DevTools）。

### 调试后端

使用 Delve 调试器：

```bash
# 安装
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试运行
dlv debug
```

### 热重载

- **前端**: Vite 自动热重载
- **后