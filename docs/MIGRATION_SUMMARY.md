# Gridea Pro Electron → Wails 迁移总结

## 迁移完成情况

### ✅ 已完成

#### 1. **后端架构迁移**
- [x] 从 Electron + Node.js 迁移到 Wails + Go
- [x] 创建 Go 项目结构（`main.go`、`backend/`、`go.mod`、`wails.json`）
- [x] 实现事件桥接系统（`runtime.EventsOn/Emit`）
- [x] 实现文件对话框（`OpenFolderDialog`）
- [x] 实现预览服务器（Go `http.Server`）

#### 2. **业务逻辑迁移**
- [x] **文章管理** (`backend/posts.go`)
  - 文章加载与解析（Markdown + YAML front-matter）
  - 文章保存（创建/更新）
  - 文章删除（包含图片清理）
  - 图片上传
  - 摘要提取（`<!-- more -->` 标记）
- [x] **标签管理** (`backend/tags.go`)
  - 标签加载
  - 标签保存
  - 标签删除
- [x] **菜单管理** (`backend/menus.go`)
  - 菜单加载
  - 菜单保存
  - 菜单删除
  - 菜单排序
- [x] **设置管理** (`backend/app.go`)
  - 设置保存
  - 站点重新加载

#### 3. **前端兼容性**
- [x] 创建 Wails-Electron 兼容桥（`src/wails-electron-bridge.ts`）
- [x] 保留所有 Vue 3 前端代码无需修改
- [x] 兼容所有 `window.electronAPI` 调用
- [x] 支持所有事件通信模式（`on/once/send/invoke`）

#### 4. **配置与构建**
- [x] 更新 `package.json`（移除 Electron 依赖）
- [x] 更新 `vite.config.ts`（移除 electron 插件）
- [x] 清理 Electron 残留文件（`electron/` 目录）
- [x] 添加 Wails 开发/构建脚本

#### 5. **文档**
- [x] 创建详细的使用文档（`README_WAILS.md`）
- [x] 创建迁移总结文档（本文档）

### 🚧 待实现（高优先级）

#### 1. **主题管理**
```go
// backend/themes.go - 需要实现
type ThemeManager struct {
    app *App
}

func (tm *ThemeManager) LoadThemes() ([]Theme, error)
func (tm *ThemeManager) LoadThemeConfig() (ThemeConfig, error)
func (tm *ThemeManager) SaveThemeConfig(config ThemeConfig) error
func (tm *ThemeManager) LoadThemeCustomConfig() (map[string]interface{}, error)
func (tm *ThemeManager) SaveThemeCustomConfig(config map[string]interface{}) error
func (tm *ThemeManager) UploadAvatar(file UploadedFile) error
func (tm *ThemeManager) UploadFavicon(file UploadedFile) error
```

**事件映射**：
- `theme-save` → `theme-saved`
- `theme-custom-config-save` → `theme-custom-config-saved`
- `avatar-upload` → `avatar-uploaded`
- `favicon-upload` → `favicon-uploaded`

#### 2. **静态站点渲染引擎**
```go
// backend/renderer.go - 需要实现
type Renderer struct {
    app         *App
    postsData   []PostRenderData
    tagsData    []TagRenderData
    menuData    []Menu
}

func (r *Renderer) RenderAll() error
func (r *Renderer) RenderPostList(archivePath string) error
func (r *Renderer) RenderPostDetail() error
func (r *Renderer) RenderTags() error
func (r *Renderer) RenderTagDetail() error
func (r *Renderer) RenderCustomPage() error
func (r *Renderer) BuildCSS() error
func (r *Renderer) BuildFeed() error
func (r *Renderer) CopyFiles() error
```

**技术选型**：
- Markdown 渲染：`github.com/yuin/goldmark` + 插件
- 模板引擎：
  - 选项 1: 保留 EJS → 通过 Node 子进程调用
  - 选项 2: 迁移到 Go `html/template`
- Less 编译：
  - 选项 1: 调用 `lessc` 命令行
  - 选项 2: 改用 Sass (`github.com/wellington/go-libsass`)
- Feed 生成：`github.com/gorilla/feeds`

**事件映射**：
- `html-render` → `html-rendered`
- `preview-site` → 打开浏览器预览

### 🔮 待实现（中优先级）

#### 3. **发布功能 - Git**
```go
// backend/deploy_git.go
func (d *Deployer) DeployToGit(config GitConfig) error
```

**依赖**：`github.com/go-git/go-git/v5`

**事件映射**：
- `publish-site` → `publish-site-result`

#### 4. **发布功能 - SFTP**
```go
// backend/deploy_sftp.go
func (d *Deployer) DeployToSFTP(config SFTPConfig) error
```

**依赖**：`github.com/pkg/sftp` + `golang.org/x/crypto/ssh`

#### 5. **发布功能 - Netlify**
```go
// backend/deploy_netlify.go
func (d *Deployer) DeployToNetlify(config NetlifyConfig) error
```

**依赖**：HTTP 客户端调用 Netlify API

### 📋 待实现（低优先级）

#### 6. **自动更新**
- 基于 GitHub Releases 检查更新
- 下载并安装更新包
- 可选：使用 `inconshreveable/go-update`

#### 7. **Analytics（可选）**
- 原版使用 `electron-google-analytics`
- Wails 版本可选择：
  - 使用 HTTP 客户端直接发送到 GA
  - 或完全移除（隐私考虑）

## 文件变更清单

### 新增文件
```
main.go                          # Wails 主入口
go.mod                           # Go 依赖管理
wails.json                       # Wails 项目配置
backend/
  ├── app.go                     # 主应用 + 事件处理器
  ├── models.go                  # 数据模型定义
  ├── posts.go                   # 文章管理
  ├── tags.go                    # 标签管理
  └── menus.go                   # 菜单管理
src/
  └── wails-electron-bridge.ts  # 前端兼容桥
README_WAILS.md                  # Wails 使用文档
MIGRATION_SUMMARY.md             # 本文档
```

### 删除文件
```
electron/
  ├── main.ts                    # ❌ 已删除
  ├── preload.ts                 # ❌ 已删除
  └── analytics.ts               # ❌ 已删除
src/server/                      # ⚠️  暂时保留供参考，后续可删除
  ├── events/                    # 事件逻辑已迁移到 backend/app.go
  ├── app.ts                     # 应用逻辑已迁移到 backend/app.go
  ├── posts.ts                   # 已迁移到 backend/posts.go
  ├── tags.ts                    # 已迁移到 backend/tags.go
  ├── menus.ts                   # 已迁移到 backend/menus.go
  ├── renderer.ts                # 待迁移
  ├── theme.ts                   # 待迁移
  ├── setting.ts                 # 部分已迁移
  ├── deploy.ts                  # 待迁移
  └── plugins/                   # 待迁移
```

### 修改文件
```
package.json                     # 移除 Electron 依赖，添加 wails 脚本
vite.config.ts                   # 移除 vite-plugin-electron
src/main.ts                      # 第 1 行新增 import bridge
```

## 技术栈对比

| 层级 | Electron 版本 | Wails 版本 |
|---|---|---|
| **窗口** | Electron BrowserWindow | Wails Runtime Window |
| **进程间通信** | ipcMain/ipcRenderer | runtime.EventsOn/Emit |
| **后端语言** | Node.js + TypeScript | Go |
| **文件操作** | fs-extra | os/io/ioutil |
| **HTTP 服务** | Express | net/http |
| **Markdown 解析** | gray-matter | gopkg.in/yaml.v3 |
| **模板引擎** | EJS | （待选）Go template / 保留 EJS |
| **样式处理** | Less (Node) | （待选）lessc 命令 / Sass Go |
| **Git 操作** | isomorphic-git | go-git |
| **SSH/SFTP** | node-ssh, ssh2-sftp-client | golang.org/x/crypto/ssh |
| **打包工具** | electron-builder | wails build |
| **打包体积** | ~200MB (包含 Node.js) | ~20-50MB (原生编译) |

## 数据兼容性

### ✅ 完全兼容
- 所有 JSON 配置文件格式不变
- Markdown 文章格式不变
- 文件目录结构不变
- 数据可与原 Electron 版本互换使用

### 存储路径
- **配置**: `~/.gridea/config.json`
- **输出**: `~/.gridea/output/`
- **站点**: `~/Documents/Gridea/` (可自定义)

## 性能提升预期

| 指标 | Electron | Wails | 提升 |
|---|---|---|---|
| 启动时间 | ~2-3s | ~0.5-1s | **2-3x** |
| 内存占用 | ~150-200MB | ~30-50MB | **3-4x** |
| 打包体积 | ~200MB | ~20-50MB | **4-10x** |
| 文章解析速度 | Node.js | Go 原生 | **5-10x** |

## 下一步行动

### 立即可做
1. **测试 Wails 开发环境**
   ```bash
   # 安装 Go
   brew install go
   
   # 安装 Wails CLI
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   
   # 初次运行（会安装依赖）
   wails dev
   ```

2. **验证现有功能**
   - 文章列表加载
   - 文章创建/编辑/删除
   - 标签管理
   - 菜单管理
   - 图片上传

### 高优先级任务
1. **实现主题管理** (1-2 天)
   - 主题列表、切换
   - 基本配置保存
   - 头像/Favicon 上传

2. **实现渲染引擎** (3-5 天)
   - Markdown → HTML 渲染
   - 模板引擎集成
   - CSS 构建
   - Feed 生成
   - 文件复制

3. **实现发布功能** (2-3 天)
   - Git 推送
   - SFTP 上传
   - （可选）Netlify 部署

### 可选优化
- 添加单元测试（Go testing 包）
- 添加 CI/CD（GitHub Actions）
- 性能分析与优化
- 错误日志系统

## 注意事项

### 开发环境
- **Go 1.22+** 必需
- **Wails CLI** 必需
- **Xcode Command Line Tools** (macOS 必需)
- **网络**：首次 `go mod tidy` 需要访问 Go 代理

### 兼容性
- macOS 10.13+
- Windows 10/11
- Linux（主流发行版）

### 已知限制
1. **EJS 模板**: 当前未迁移，建议后续选择：
   - 保留 EJS（通过 Node 子进程）
   - 迁移到 Go `html/template`（推荐）

2. **Less 样式**: 当前未处理，建议：
   - 调用 `lessc` 命令编译
   - 迁移到 Sass Go 绑定

3. **Analytics**: Google Analytics 功能未迁移，可选实现

## 联系与支持

- **Wails 官方文档**: https://wails.io/docs/
- **Go 官方文档**: https://go.dev/doc/
- **Vue 3 文档**: https://vuejs.org/
- **问题反馈**: 提交 GitHub Issue

---

**迁移完成日期**: 2025-01-04  
**版本**: 1.0.0  
**状态**: 核心功能已完成，渲染与发布功能待实现
