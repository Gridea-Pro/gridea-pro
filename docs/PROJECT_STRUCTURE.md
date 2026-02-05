# 项目结构文档

## 📁 当前项目结构

```
Gridea Pro/
├── src/                          # 前端源代码
│   ├── assets/                   # 静态资源
│   │   ├── fonts/               # 字体文件
│   │   ├── images/              # 图片资源
│   │   ├── styles/              # 样式文件
│   │   │   ├── tailwind.css    # Tailwind CSS 入口
│   │   │   ├── main.less       # 主样式文件
│   │   │   ├── var.less        # Less 变量
│   │   │   └── custom.less     # 自定义样式
│   │   └── locales.ts          # 国际化文件
│   ├── components/              # 可复用组件
│   │   ├── AppSystem/          # 系统设置组件
│   │   ├── ColorCard/          # 颜色选择卡片
│   │   ├── EmojiCard/          # Emoji 选择卡片
│   │   ├── FooterBox/          # 页脚组件
│   │   ├── MonacoMarkdownEditor/ # Monaco 编辑器
│   │   └── PostsCard/          # 文章卡片
│   ├── helpers/                 # 工具函数
│   │   ├── constants.ts        # 常量定义
│   │   ├── utils.ts            # 通用工具
│   │   ├── markdown.ts         # Markdown 处理
│   │   └── slug.ts             # URL slug 生成
│   ├── interfaces/              # TypeScript 接口
│   │   ├── post.ts             # 文章接口
│   │   ├── tag.ts              # 标签接口
│   │   ├── theme.ts            # 主题接口
│   │   ├── menu.ts             # 菜单接口
│   │   └── setting.ts          # 设置接口
│   ├── layouts/                 # 布局组件
│   │   └── MainLayout.vue      # 主布局
│   ├── pages/                   # 页面组件（简单页面）
│   │   └── Home.vue            # 首页
│   ├── router/                  # 路由配置
│   │   └── index.ts            # 路由定义
│   ├── stores/                  # Pinia 状态管理
│   │   └── site.ts             # 站点状态
│   ├── types/                   # 全局类型定义
│   │   └── global.d.ts         # 全局类型声明
│   ├── views/                   # 视图组件（复杂页面）
│   │   ├── article/            # 文章管理
│   │   ├── tags/               # 标签管理
│   │   ├── menu/               # 菜单管理
│   │   ├── theme/              # 主题设置
│   │   ├── setting/            # 系统设置
│   │   └── loading/            # 加载页面
│   ├── App.vue                  # 根组件
│   ├── main.ts                  # 入口文件
│   └── wails-bridge.ts         # Wails 桥接
├── wailsjs/                     # Wails 生成的 JS 绑定
│   ├── go/                      # Go 函数绑定
│   └── runtime/                 # Wails 运行时
├── public/                      # 公共资源
│   ├── app-icons/              # 应用图标
│   └── default-files/          # 默认文件模板
├── docs/                        # 项目文档
│   ├── OPTIMIZATION_SUMMARY.md # 优化总结
│   └── PROJECT_STRUCTURE.md    # 项目结构（本文件）
├── .vscode/                     # VS Code 配置
│   └── settings.json           # 编辑器设置
├── dist/                        # 构建输出目录
├── build/                       # Wails 构建目录
├── *.go                         # Go 后端代码
├── package.json                 # NPM 依赖
├── tsconfig.json               # TypeScript 配置
├── vite.config.ts              # Vite 配置
├── tailwind.config.js          # Tailwind 配置
├── wails.json                  # Wails 配置
└── go.mod                      # Go 模块定义
```

## ✅ 结构优化评估

### 优点
1. ✅ **清晰的分层结构**: components, views, layouts 分离合理
2. ✅ **类型定义集中**: interfaces 和 types 目录管理类型
3. ✅ **工具函数模块化**: helpers 目录组织良好
4. ✅ **状态管理独立**: stores 目录专门管理 Pinia store
5. ✅ **资源分类清晰**: assets 按类型分类（fonts, images, styles）

### 改进建议

#### 1. 组件命名规范
**现状**: 部分组件使用 `Index.vue` 作为文件名

**建议**: 
```
❌ components/ColorCard/Index.vue
✅ components/ColorCard/ColorCard.vue

或保持 Index.vue 但文件夹名使用 PascalCase（已经符合）
```

#### 2. 类型定义整合
**现状**: 类型定义分散在 `interfaces/` 和 `types/` 两个目录

**建议**: 
```
方案 A: 合并到 types/ 目录
types/
  ├── global.d.ts
  ├── post.ts
  ├── tag.ts
  └── ...

方案 B: 保持当前结构，但明确职责
- interfaces/: 数据模型接口
- types/: 工具类型和全局声明
```

#### 3. 视图组件结构
**现状**: views 和 pages 并存，职责不够清晰

**建议**: 统一使用 views，移除 pages 目录
```
views/
  ├── home/
  │   └── Home.vue
  ├── article/
  │   ├── Articles.vue
  │   └── ArticleUpdate.vue
  └── ...
```

#### 4. 样式文件组织
**现状**: less 和 tailwind 混用

**建议**: 
- ✅ 保持当前结构（已经合理）
- 逐步迁移 less 变量到 Tailwind 主题配置
- 使用 CSS 模块或 scoped styles

## 📋 文件命名规范

### Vue 组件
- ✅ **PascalCase**: `ArticleUpdate.vue`, `MainLayout.vue`
- ✅ **单文件组件**: 每个组件一个文件
- ✅ **组件文件夹**: 复杂组件使用文件夹包含子组件

### TypeScript 文件
- ✅ **kebab-case**: `content-helper.ts`, `words-count.ts`
- ✅ **语义化命名**: 文件名清晰表达功能

### 样式文件
- ✅ **kebab-case**: `tailwind.css`, `main.less`
- ✅ **扩展名准确**: `.css`, `.less` 清晰区分

## 🎯 推荐的最佳实践

### 1. 组件组织
```typescript
// ✅ 推荐：按功能域组织
views/
  article/
    ├── Articles.vue          # 列表页
    ├── ArticleUpdate.vue     # 编辑页
    └── components/           # 页面专属组件
        └── ArticleCard.vue

// ✅ 推荐：可复用组件独立
components/
  ColorCard/
    ├── ColorCard.vue
    └── ColorCard.test.ts     # 单元测试（未来）
```

### 2. 类型导入
```typescript
// ✅ 推荐：使用 type 导入
import type { IPost } from '@/interfaces/post'

// ❌ 避免
import { IPost } from '@/interfaces/post'
```

### 3. 路径别名
```typescript
// ✅ 已配置：使用 @ 别名
import { useSiteStore } from '@/stores/site'

// ❌ 避免：相对路径
import { useSiteStore } from '../../../stores/site'
```

## 🔧 建议的调整

### 优先级 1（高）
1. ✅ **已完成**: Tailwind CSS 配置优化
2. ✅ **已完成**: TypeScript 配置现代化
3. ✅ **已完成**: 类型安全改进
4. ⏳ **待完成**: 统一视图组件结构（移除 pages/，全部使用 views/）

### 优先级 2（中）
1. ⏳ 考虑合并 interfaces/ 和 types/ 目录
2. ⏳ 为组件添加单元测试文件
3. ⏳ 创建 composables/ 目录存放可复用的组合式函数

### 优先级 3（低）
1. ⏳ 添加 .editorconfig 统一编辑器配置
2. ⏳ 添加 .prettierrc 统一代码格式
3. ⏳ 考虑使用 pnpm 替代 npm

## 📊 当前结构评分

| 维度 | 评分 | 说明 |
|------|------|------|
| 目录组织 | ⭐⭐⭐⭐⭐ | 结构清晰，分层合理 |
| 命名规范 | ⭐⭐⭐⭐☆ | 基本规范，少量可改进 |
| 模块化程度 | ⭐⭐⭐⭐⭐ | 高度模块化，职责清晰 |
| 可维护性 | ⭐⭐⭐⭐⭐ | 易于理解和维护 |
| 扩展性 | ⭐⭐⭐⭐⭐ | 便于添加新功能 |

## 总结

**当前项目结构整体上非常优秀**，遵循了 Vue 3 和现代前端项目的最佳实践。主要优点包括：

1. 清晰的目录分层
2. 合理的功能模块划分
3. 良好的命名规范
4. 完善的类型定义

建议的改进主要是细节优化，不影响当前的开发和维护工作。
