# Gridea Pro Wails - 快速开始

## 前置要求

### 必需安装

1. **Go 1.22+**

   ```bash
   # macOS
   brew install go

   # 设置国内代理（可选，但强烈推荐）
   go env -w GOPROXY=https://goproxy.cn,direct
   ```
2. **Wails CLI**

   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest

   # 验证安装
   wails doctor
   ```
3. **Node.js 18+** 和 **npm**

   ```bash
   # macOS
   brew install node
   ```
4. **Less 编译器**（用于主题样式编译）

   ```bash
   npm install -g less
   ```

### 可选安装

1. **EJS 模板引擎**（已在项目中安装）
   ```bash
   npm install ejs --save-dev
   ```

## 初始化项目

### 1. 安装前端依赖

```bash
npm install
```

### 2. 安装 Go 依赖

```bash
go mod tidy
```

如果遇到网络问题：

```bash
# 设置国内代理
go env -w GOPROXY=https://goproxy.cn,direct

# 重新下载
go mod tidy
```

### 3. 初始化 Wails

```bash
# 检查环境
wails doctor

# 如果有缺失依赖，按提示安装
```

## 开发运行

### 方式1：使用 Wails CLI（推荐）

```bash
wails dev
```

这会自动：

- 启动前端 Vite DevServer（http://127.0.0.1:5173）
- 编译并运行 Go 后端
- 打开 Wails 应用窗口
- 启用热重载（前端 + 后端）

### 方式2：使用 npm script

```bash
npm run wails:dev
```

### 开发模式特性

- **前端热重载**：修改 Vue 代码自动刷新
- **后端热重载**：修改 Go 代码自动重启
- **开发工具**：应用窗口内可打开 DevTools
- **日志输出**：终端显示前后端所有日志

## 生产构建

```bash
# 构建当前平台
wails build

# macOS
wails build -platform darwin/universal

# Windows（需在 Windows 系统运行）
wails build -platform windows/amd64

# Linux
wails build -platform linux/amd64
```

构建产物位于 `build/bin/` 目录。

## 项目结构

```
.
├── backend/              # Go 后端代码
│   ├── app.go           # 主应用 + 事件处理
│   ├── models.go        # 数据模型
│   ├── posts.go         # 文章管理
│   ├── tags.go          # 标签管理
│   ├── menus.go         # 菜单管理
│   ├── themes.go        # 主题管理
│   ├── renderer.go      # 静态站点渲染
│   └── deploy.go        # 发布功能
├── src/                 # Vue 前端代码
│   ├── views/           # 页面组件
│   ├── components/      # UI 组件
│   ├── stores/          # Pinia 状态
│   ├── router/          # 路由配置
│   └── wails-electron-bridge.ts  # 兼容桥
├── wailsjs/             # Wails 自动生成（运行时）
├── main.go              # Go 主入口
├── wails.json           # Wails 配置
├── go.mod               # Go 依赖
└── package.json         # npm 依赖
```

## 功能清单

### ✅ 已实现功能

- [X] 文章管理（增删改查、图片上传）
- [X] 标签管理
- [X] 菜单管理
- [X] 设置管理
- [X] 主题管理（列表、切换、自定义配置）
- [X] 头像/Favicon 上传
- [X] 站点目录选择
- [X] 预览服务器（localhost:4000）
- [X] 静态站点渲染
  - [X] Markdown → HTML（Goldmark）
  - [X] EJS 模板引擎
  - [X] Less 样式编译
  - [X] Feed 生成（Atom + RSS）
  - [X] 文件复制与分页
- [X] 发布功能
  - [X] Git 推送（GitHub/Coding）
  - [X] SFTP 上传
  - [X] Netlify 部署

### 事件 API

所有事件通过 `window.electronAPI` 调用（兼容桥）：

#### 站点管理

- `app-ready` → `app-site-loaded`
- `app-site-reload` → `app-site-loaded`
- `app-source-folder-setting` → `app-source-folder-set`

#### 文章管理

- `app-post-create` → `app-post-created`
- `app-post-delete` → `app-post-deleted`
- `app-post-list-delete` → `app-post-list-deleted`
- `image-upload` → `image-uploaded`

#### 标签管理

- `tag-save` → `tag-saved`
- `tag-delete` → `tag-deleted`

#### 菜单管理

- `menu-save` → `menu-saved`
- `menu-delete` → `menu-deleted`
- `menu-sort` → 自动保存

#### 主题管理

- `theme-save` → `theme-saved`
- `theme-custom-config-save` → `theme-custom-config-saved`
- `avatar-upload` → `avatar-uploaded`
- `favicon-upload` → `favicon-uploaded`

#### 渲染与发布

- `html-render` → `html-rendered`
- `publish-site` → `publish-site-result`
- `preview-site` → 打开浏览器

## 常见问题

### Q: wails dev 启动失败？

**A1**: 检查 Go 依赖是否完整

```bash
go mod tidy
```

**A2**: 检查 Wails 环境

```bash
wails doctor
```

**A3**: 清理缓存重试

```bash
rm -rf wailsjs
wails dev
```

### Q: 前端编译错误？

**A**: 确保安装了所有 npm 依赖

```bash
rm -rf node_modules package-lock.json
npm install
```

### Q: 渲染失败？

**A1**: 确认已选择主题

- 在应用的"主题"页面选择一个主题

**A2**: 确认 Less 编译器已安装

```bash
npm install -g less
lessc --version
```

**A3**: 确认 EJS 已安装

```bash
npm list ejs
```

### Q: 发布失败？

**A1**: Git 推送失败

- 检查 Token 权限
- 检查仓库地址格式
- 检查分支名称

**A2**: SFTP 失败

- 检查服务器地址和端口
- 检查用户名和密码/密钥
- 检查远程路径权限

**A3**: Netlify 失败

- 检查 Site ID 和 Access Token
- 检查网络连接

## 数据目录

### 配置目录

```
~/.gridea/
├── config.json        # 源文件夹配置
└── output/            # 构建输出目录
```

### 站点目录（默认）

```
~/Documents/Gridea/
├── posts/             # Markdown 文章
├── config/            # JSON 配置
│   ├── posts.json
│   ├── tags.json
│   ├── menus.json
│   ├── setti
```
