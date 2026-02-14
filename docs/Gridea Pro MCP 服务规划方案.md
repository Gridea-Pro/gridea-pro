# Gridea Pro MCP 服务规划方案

> 创建日期：2026-02-13
> 状态：**规划讨论中**

---

## 一、架构背景

Gridea Pro 是一个基于 **Wails (Go + Vue 3)** 的桌面静态博客客户端，后端采用 `Domain → Repository → Service → Facade` 分层架构。

核心实体包括：

| 实体 | 说明 |
|------|------|
| **Post** | 文章（Markdown，含元数据、标签、分类） |
| **Memo** | 闪念（短文本笔记，支持标签） |
| **Tag** | 标签（name, slug, color） |
| **Category** | 分类（name, slug, description） |
| **Link** | 友链（name, url, description, avatar） |
| **Menu** | 导航菜单（name, link, openType） |
| **Theme** | 主题（模板引擎 + 自定义配置） |
| **Setting** | 部署设置（Git/Netlify/SFTP 等） |
| **Comment** | 评论系统 |

MCP 服务将作为 Go 后端的一个**独立模块**，复用现有的 Service 层，以 **stdio 方式** 启动 MCP Server 进程，供 Claude Desktop / Cursor / 其他 AI Agent 调用。

---

## 二、MCP 接口设计（Tools）

### 2.1 文章管理（Post）— *核心能力*

| Tool | 描述 | 典型 AI 使用场景 |
|------|------|----------------|
| `list_posts` | 获取所有文章列表（含元数据） | 了解博客全景、筛选文章 |
| `get_post` | 获取单篇文章内容（by fileName） | 阅读/分析文章内容 |
| `create_post` | 创建新文章（标题、内容、标签、分类等） | AI 辅助写作、批量生成 |
| `update_post` | 更新文章内容或元数据 | AI 润色、翻译、修改 |
| `delete_post` | 删除文章 | 批量清理 |

### 2.2 闪念管理（Memo）— *核心能力*

| Tool | 描述 | 典型 AI 使用场景 |
|------|------|----------------|
| `list_memos` | 获取所有闪念 | 浏览灵感/笔记 |
| `create_memo` | 创建闪念（内容自动提取标签） | 快速记录想法 |
| `update_memo` | 更新闪念 | 编辑/补充内容 |
| `delete_memo` | 删除闪念 | 清理 |
| `get_memo_stats` | 获取闪念统计（热力图数据等） | 分析写作频率 |

### 2.3 标签管理（Tag）

| Tool | 描述 |
|------|------|
| `list_tags` | 获取所有标签 |
| `create_tag` | 创建标签（name, slug, color） |
| `delete_tag` | 删除标签 |

### 2.4 分类管理（Category）

| Tool | 描述 |
|------|------|
| `list_categories` | 获取所有分类 |
| `create_category` | 创建分类 |
| `delete_category` | 删除分类 |

### 2.5 友链管理（Link）

| Tool | 描述 |
|------|------|
| `list_links` | 获取友链列表 |
| `save_link` | 新增/更新友链 |
| `delete_link` | 删除友链 |

### 2.6 菜单管理（Menu）

| Tool | 描述 |
|------|------|
| `list_menus` | 获取菜单列表 |
| `save_menu` | 新增/更新菜单 |
| `delete_menu` | 删除菜单 |

### 2.7 主题与站点配置（Theme & Setting）

| Tool | 描述 |
|------|------|
| `get_theme_config` | 获取当前主题配置（站点名、作者、分页等） |
| `update_theme_config` | 更新主题配置 |
| `list_themes` | 获取已安装主题列表 |
| `get_site_setting` | 获取部署设置（平台、仓库等） |

### 2.8 渲染与部署（Renderer & Deploy）— *高级操作*

| Tool | 描述 |
|------|------|
| `render_site` | 渲染/生成静态站点 |
| `deploy_site` | 部署到 Git/Netlify/SFTP |

### 2.9 评论管理（Comment）

| Tool | 描述 |
|------|------|
| `list_comments` | 获取评论列表（分页） |
| `reply_comment` | 回复评论 |
| `delete_comment` | 删除评论 |

---

## 三、Resources（资源暴露）

MCP 协议允许 Server 暴露"资源"，AI 可以主动读取这些资源来获取上下文信息。

| Resource URI | 描述 |
|------|------|
| `gridea://site/info` | 站点基本信息（名称、域名、描述等） |
| `gridea://posts/summary` | 文章列表概要（标题 + 日期 + 标签，供 AI 快速了解博客全貌） |
| `gridea://memos/recent` | 最近的闪念列表 |

---

## 四、危险操作确认机制

> ✅ **已确认：需要对危险操作加确认机制。**

对以下高危操作实施"两步确认"模式：

| 危险操作 | 确认策略 |
|---------|---------|
| `delete_post` | 先返回文章标题和摘要，要求 AI 确认后真正删除 |
| `delete_memo` | 先返回闪念内容预览，要求确认 |
| `delete_tag` / `delete_category` | 返回关联的文章数量，提示影响范围 |
| `deploy_site` | 返回当前站点状态（未渲染的变更数量），要求确认 |
| `update_theme_config` | 返回变更 diff，要求确认 |

**实现方式**：Tool 接收一个 `confirm: bool` 参数。
- 首次调用（`confirm=false` 或不传）：返回预览信息 + `"需要确认，请带 confirm=true 重新调用"`
- 二次调用（`confirm=true`）：真正执行操作

---

## 五、技术实现方案

### 5.1 目录结构

```
backend/
  cmd/
    mcp/
      main.go               ← MCP 独立入口（或集成到主 binary 的子命令）
  internal/
    mcp/                     ← 新增 MCP 模块
      server.go              ← MCP Server 初始化 & stdio transport
      tools.go               ← Tool 定义与注册
      tool_post.go           ← 文章相关 Tool 处理器
      tool_memo.go           ← 闪念相关 Tool 处理器
      tool_tag.go            ← 标签相关 Tool 处理器
      tool_category.go       ← 分类相关 Tool 处理器
      tool_link.go           ← 友链相关 Tool 处理器
      tool_menu.go           ← 菜单相关 Tool 处理器
      tool_theme.go          ← 主题相关 Tool 处理器
      tool_renderer.go       ← 渲染 & 部署 Tool 处理器
      tool_comment.go        ← 评论相关 Tool 处理器
      resources.go           ← Resource 定义
      prompts.go             ← Prompt 模板定义
```

### 5.2 依赖关系

```
MCP Server (stdio)
  └── Tool Handlers
        └── Service 层（复用现有代码，零重复业务逻辑）
              └── Repository 层
                    └── 文件系统（JSON/Markdown 文件）
```

### 5.3 Go MCP SDK

推荐使用 [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go)，这是 Go 生态最成熟的 MCP SDK，支持：
- stdio / SSE transport
- Tool、Resource、Prompt 完整协议
- JSON Schema 参数校验

### 5.4 环境变量

| 变量 | 说明 | 默认值 |
|------|------|-------|
| `GRIDEA_SOURCE_DIR` | 站点数据目录 | `~/Documents/Gridea Pro` |

---

## 六、Q&A：关键设计决策

### Q1：MCP 启动方式——子命令 vs 独立二进制

#### 方案 A：子命令模式（`gridea-pro mcp`）

将 MCP 功能集成到 Gridea Pro 主程序中，通过子命令启动。

**用户配置：**
```json
{
  "mcpServers": {
    "gridea-pro": {
      "command": "/Applications/Gridea Pro.app/Contents/MacOS/Gridea Pro",
      "args": ["mcp"]
    }
  }
}
```

**优点：**
- 统一一个二进制，安装即可用，不需要额外下载
- 版本始终和 GUI 保持同步
- 共用同一套配置加载逻辑

**缺点：**
- 主程序体积增大（MCP SDK 依赖被打包进去）
- Wails 打包后的 `.app` 中执行子命令可能需要特殊处理（macOS .app bundle 路径问题）
- 启动时需要跳过 Wails GUI 初始化，增加判断逻辑

#### 方案 B：独立二进制（`gridea-mcp`）

单独编译一个轻量级的 MCP Server 二进制。

**用户配置：**
```json
{
  "mcpServers": {
    "gridea-pro": {
      "command": "/usr/local/bin/gridea-mcp",
      "env": {
        "GRIDEA_SOURCE_DIR": "/Users/eric/Documents/Gridea Pro"
      }
    }
  }
}
```

**优点：**
- 轻量，不包含 Wails/GUI 依赖，编译体积小
- 启动快，无 GUI 初始化开销
- 可独立安装（`go install` 或 Homebrew）
- 路径简单干净

**缺点：**
- 需要单独分发和更新
- 需要用户手动指定 `GRIDEA_SOURCE_DIR`

#### 🏆 推荐：方案 B（独立二进制）

理由：
1. MCP 的本质就是一个"无头"服务进程，不需要 GUI
2. 避免和 Wails 打包产生冲突
3. 可以通过 `go build -o gridea-mcp ./backend/cmd/mcp` 单独构建
4. 未来还可以发布到 Homebrew，方便其他用户安装

---

### Q2：危险操作确认机制

> 已在第四节详细说明，采用 `confirm` 参数的两步确认模式。

---

### Q3：MCP Prompts 是什么？怎么用？

#### 什么是 MCP Prompts？

MCP 协议中的 **Prompts** 是服务端预定义的"提示词模板"。它不是 Tool（不执行操作），而是告诉 AI **"在某种场景下，你应该怎么和用户交互、调用哪些 Tool、遵循什么规则"**。

可以理解为：**它是给 AI 的"操作手册"，AI 在合适的时机会自动加载并遵循。**

#### 举个具体例子

假设我们定义一个 Prompt 叫 `blog_writing_assistant`：

```json
{
  "name": "blog_writing_assistant",
  "description": "博客写作助手 - 帮助用户将想法转化为高质量博客文章",
  "arguments": [
    {
      "name": "topic",
      "description": "文章主题或方向",
      "required": true
    },
    {
      "name": "style",
      "description": "写作风格，如 '技术教程'、'随笔'、'评测'",
      "required": false
    }
  ]
}
```

当用户在 Claude Desktop 中选择这个 Prompt 并输入主题后，AI 会收到一段预设的系统提示词，例如：

```
你是 Gridea Pro 博客写作助手。请按以下步骤帮助用户创作博客文章：

1. 先调用 list_tags 和 list_categories 了解当前的标签和分类体系
2. 根据用户主题，拟定文章大纲
3. 征求用户意见后，撰写完整文章
4. 使用 create_post 创建文章，自动匹配合适的标签和分类
5. 询问用户是否需要立即 render_site 渲染站点
```

#### 可以做哪些 Prompts？

| Prompt | 描述 |
|--------|------|
| `blog_writing_assistant` | 博客写作助手：引导 AI 从构思到发布的完整流程 |
| `memo_to_post` | 闪念整理器：将多条闪念组织成一篇博客文章 |
| `content_review` | 内容审查：检查所有文章的 SEO、标签完整性、拼写等 |
| `site_health_check` | 站点健康检查：诊断问题（空标签、无分类文章、死链等） |
| `translate_post` | 文章翻译：将指定文章翻译成目标语言并创建新文章 |

#### 怎么使用？

在 Claude Desktop 中，用户在聊天输入框旁会看到一个 **"/"** 按钮或快捷方式，列出所有可用的 Prompts。选择一个 Prompt 后，AI 就会按照预设的引导流程和规则来执行任务。

**总结**：Prompts 让 MCP 从"工具箱"升级为"智能工作流"，为常见的博客操作场景提供开箱即用的最佳实践指引。

---

### Q4：SSE Transport 是什么？怎么用？

#### 什么是 SSE (Server-Sent Events)？

MCP 协议支持两种主要的传输方式：
1.  **stdio（标准输入输出）**：Claude Desktop 作为一个父进程，直接启动 `gridea-mcp` 二进制文件，通过管道（Pipe）进行通信。
    *   **特点**：简单、零网络开销、随 Claude 启动/关闭。
    *   **适用**：本地桌面应用（目前的 Gridea Pro 场景）。
2.  **SSE（基于 HTTP）**：`gridea-mcp` 作为一个独立的 Web 服务器运行（例如监听 8080 端口），Claude 通过 HTTP 协议连接。
    *   **特点**：支持远程连接、支持多个客户端同时连接。
    *   **适用**：将 Gridea 部署在云服务器上，或者需要通过网页版 Agent 访问时。

#### 怎么用？

如果启用了 SSE，你需要这样启动：
```bash
# 启动服务并监听端口
./gridea-mcp --transport sse --port 8080
```

然后在 Claude 配置中使用 `url` 而不是 `command`：
```json
"gridea-pro": {
  "url": "http://localhost:8080/sse",
  "env": ...
}
```

> **建议**：对于 Gridea Pro 桌面版，**stdio 模式是最佳选择**，配置更简单且无需管理后台进程。除非你有远程访问本地 Gridea 数据的需求，否则暂时不需要 SSE。

---

## 七、使用方式（最终用户视角）

### 7.1 安装

```bash
# 方式一：从源码构建
cd /Volumes/Work/VibeCoding/Gridea\ Pro
go build -o gridea-mcp ./backend/cmd/mcp

# 方式二：go install（发布后）
go install github.com/Tespera/gridea-pro/backend/cmd/mcp@latest
```

### 7.2 配置（Claude Desktop）

编辑 `~/Library/Application Support/Claude/claude_desktop_config.json`：

```json
{
  "mcpServers": {
    "gridea-pro": {
      "command": "/path/to/gridea-mcp",
      "env": {
        "GRIDEA_SOURCE_DIR": "/Users/eric/Documents/Gridea Pro"
      }
    }
  }
}
```

> **关键点**：`command` 必须指向 `gridea-mcp` 这个二进制文件的**真实绝对路径**。如果不确定，可以在终端找到它然后输入 `pwd` 查看。

> **注意**：`command` 必须是可执行文件的**绝对路径**。

**配置示例：**

```json
{
  "mcpServers": {
    "gridea-pro": {
      "command": "/usr/local/bin/gridea-mcp",
      "env": {
        "GRIDEA_SOURCE_DIR": "/Users/eric/Documents/Gridea Pro"
      }
    }
  }
}
```

> **注意**：`GRIDEA_SOURCE_DIR` 环境变量必须指向 Gridea Pro 数据目录。

### 7.3 使用示例

配置完成后，重启 Claude Desktop，即可在对话中直接操作博客：

| 你对 AI 说…… | AI 会做…… |
|-------------|----------|
| "帮我写一篇关于 Go 并发的技术博客" | 调用 `list_tags` → 构思 → `create_post` |
| "把我最近 10 条闪念整理成一篇文章" | `list_memos` → 分析归类 → `create_post` |
| "给所有没有标签的文章自动打标签" | `list_posts` → 分析内容 → `update_post` |
| "帮我翻译这篇文章为英文" | `get_post` → 翻译 → `create_post` |
| "检查我的博客有什么问题" | Prompt: site_health_check → 全面诊断 |
| "渲染并部署我的站点" | `render_site` → `deploy_site`（带确认） |

---

## 八、待确定事项

- [x] 危险操作确认机制 → **采用 `confirm` 参数两步确认**
- [x] MCP 启动方式 → **独立二进制 (`gridea-mcp`)**
- [x] 首期 Tool 范围 → **全量实现**
- [x] 是否加入 Prompts → **需要**
- [x] 是否需要 SSE transport → **暂不实现，优先支持 stdio，后续视需求添加**
