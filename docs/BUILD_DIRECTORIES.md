# 构建目录说明

## 📁 构建相关目录解析

### 1. `/dist` 目录
**用途**: Vite 前端构建输出目录

**生成时机**: 
```bash
npm run build
```

**内容**:
- HTML、CSS、JavaScript 等前端静态资源
- 优化和压缩后的代码
- 静态资源（图片、字体等）

**是否需要**: ✅ **必需**
- 这是前端应用的最终产物
- Wails 打包时会将此目录嵌入到应用中

**Git 追踪**: ❌ 不应该提交到 Git（已在 .gitignore 中）

---

### 2. `/dist-electron` 目录
**用途**: Electron 构建输出（如果项目曾使用 Electron）

**当前状态**: ⚠️ **已废弃**
- 项目已从 Electron 迁移到 Wails
- 此目录不再使用

**建议**: 🗑️ **可以安全删除**
```bash
rm -rf dist-electron
```

**Git 追踪**: ❌ 不应该提交到 Git（已在 .gitignore 中）

---

### 3. `/build` 目录
**用途**: Wails 应用构建输出目录

**生成时机**:
```bash
wails build
```

**内容**:
- 打包好的桌面应用程序
- macOS: `.app` 文件
- Windows: `.exe` 文件
- Linux: 可执行二进制文件

**是否需要**: ✅ **必需**
- 包含最终的可分发应用程序
- 用于发布和分发

**Git 追踪**: ❌ 不应该提交到 Git（已在 .gitignore 中）

---

### 4. `/wailsjs` 目录
**用途**: Wails 自动生成的 JavaScript/TypeScript 绑定

**生成时机**:
```bash
wails dev
wails build
```

**内容**:
- Go 函数的 TypeScript/JavaScript 绑定
- 运行时类型定义
- 自动生成，不应手动修改

**是否需要**: ✅ **必需**（开发时）
- 前端调用后端 Go 函数时需要
- 每次 Wails 命令都会重新生成

**Git 追踪**: ❌ 不应该提交到 Git（已在 .gitignore 中）
- 因为是自动生成的，不同环境可能略有差异

---

### 5. `/node_modules` 目录
**用途**: NPM 依赖包目录

**生成时机**:
```bash
npm install
```

**是否需要**: ✅ **必需**
- 包含所有前端依赖

**Git 追踪**: ❌ 绝对不应该提交到 Git

---

## 🧹 清理建议

### 可以安全删除的目录
```bash
# 删除旧的 Electron 构建目录
rm -rf dist-electron

# 清理所有构建产物（需要时重新构建）
rm -rf dist build wailsjs

# 清理依赖（需要时重新安装）
rm -rf node_modules
```

### 重新构建流程
```bash
# 1. 安装依赖
npm install

# 2. 开发模式（会自动生成 wailsjs）
wails dev

# 或者

# 2. 构建生产版本
npm run build    # 构建前端 -> dist/
wails build      # 构建应用 -> build/
```

---

## 📊 目录大小和重要性

| 目录 | 典型大小 | 重要性 | Git 追踪 | 可删除 |
|------|---------|--------|----------|--------|
| `dist` | ~10-50MB | 高 | ❌ | ✅ 可重建 |
| `dist-electron` | ~50MB | 无（废弃） | ❌ | ✅ 建议删除 |
| `build` | ~50-200MB | 高 | ❌ | ✅ 可重建 |
| `wailsjs` | ~1MB | 高 | ❌ | ✅ 自动生成 |
| `node_modules` | ~500MB | 高 | ❌ | ✅ 可重建 |

---

## 🎯 最佳实践

### 开发阶段
```bash
# 只需要运行
wails dev

# 会自动：
# 1. 生成 wailsjs 绑定
# 2. 启动前端开发服务器
# 3. 启动 Go 后端
# 4. 启动应用窗口
```

### 构建阶段
```bash
# 构建前端
npm run build

# 构建整个应用
wails build

# 产物在 build/ 目录
```

### 清理阶段
```bash
# 清理所有构建产物
npm run clean  # 如果有配置

# 或手动清理
rm -rf dist build wailsjs node_modules
```

---

## ⚙️ package.json 建议脚本

建议在 `package.json` 中添加以下脚本：

```json
{
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    "clean": "rm -rf dist build wailsjs node_modules",
    "clean:build": "rm -rf dist build",
    "lint": "eslint src --ext .vue,.ts,.js",
    "lint:fix": "eslint src --ext .vue,.ts,.js --fix",
    "format": "prettier --write \"src/**/*.{vue,ts,js,json,css,less}\""
  }
}
```

---

## 📝 总结

### 需要保留的目录
- ✅ `dist` - 前端构建输出（可重建）
- ✅ `build` - 应用构建输出（可重建）
- ✅ `wailsjs` - Wails 绑定（自动生成）
- ✅ `node_modules` - NPM 依赖（可重建）

### 可以删除的目录
- 🗑️ `dist-electron` - 已废弃，不再使用

### 不应该提交到 Git 的目录
- ❌ 所有上述目录都不应该提交到 Git
- ✅ 已在 `.gitignore` 中正确配置
