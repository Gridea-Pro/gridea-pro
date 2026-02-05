# Gridea Pro - Wails 版本

一个基于 Wails (Go + Vue 3) 的静态博客写作客户端。

## 🎉 特性

- ✅ 完整的文章管理（Markdown 编辑、图片上传）
- ✅ 标签和菜单管理
- ✅ 主题系统（支持自定义主题）
- ✅ 静态站点生成（Goldmark + EJS）
- ✅ 多平台发布（GitHub/Coding/SFTP/Netlify）
- ✅ 本地预览服务器
- ✅ Feed 生成（Atom + RSS）

## 🚀 快速开始

### 前置要求

1. **Go 1.22+**
   ```bash
   brew install go
   go env -w GOPROXY=https://goproxy.cn,direct
   ```

2. **Wails CLI**
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   wails doctor
   ```

3. **Node.js 18+** 和 **Less 编译器**
   ```bash
   brew install node
   npm install -g less
   ```

### 安装与运行

```bash
# 1. 克隆项目
git clone <repository>
cd "Gridea Pro"

# 2. 安装依赖
npm install
go mod tidy

# 3. 运行开发环境
wails dev

# 4. 构建生产版本
wails build
```

## 📖 文档

- [快速开始指南](./SETUP.md) - 详细的安装和使用说明
- [Wails 使用文档](./README_WAILS.md) - 架构说明和 API 文档
- [迁移总结](./MIGRATION_SUMMARY.md) - Electron → Wails 迁移详情
- [完成报告](./MIGRATION_COMPLETE.md) - 功能清单和性能数据

## 📊 性能对比

| 指标 | Electron | Wails | 提升 |
|------|----------|-------|------|
| 启动时间 | 2-3秒 | 0.5-1秒 | **2-3x** |
| 内存占用 | 150-200MB | 30-50MB | **3-4x** |
| 打包体积 | ~200MB | ~20-50MB | **4-10x** |

## 🛠 技术栈

- **桌面框架**: Wails 2.9.2
- **后端语言**: Go 1.22
- **前端框架**: Vue 3 + TypeScript
- **UI 组件**: Ant Design Vue
- **构建工具**: Vite
- **Markdown 渲染**: Goldmark
- **模板引擎**: EJS
- **样式处理**: Less + Tailwind CSS

## 📝 开发

```bash
# 开发模式（热重载）
wails dev

# 构建
wails build

# 测试 Go 代码
go test ./...

# 格式化 Go 代码
go fmt ./...
```

## 📄 许可证

MIT License

## 🙏 致谢

基于 [Gridea](https://github.com/getgridea/gridea) 项目重构。
