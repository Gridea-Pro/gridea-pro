# Markdown 往返保真（方言对齐）— Round 1 结果与说明

测试台：`frontend/scripts/md-roundtrip.mts`（`npm run test:md` / `test:md:posts`）。
走 `@tiptap/markdown` 的 `MarkdownManager.parse/serialize` —— 与生产
`editor.getMarkdown()` / `setContent(md,'markdown')` 同一条管线、同一套扩展钩子，结论忠实于生产；
不创建 EditorView（用 jsdom 提供 DOMParser），规避无头视图问题。

## 验收口径（硬门槛）
- **lossless**：词/URL 级 token 无丢失（对重排/重格式化不敏感）。
- **stable**：序列化收敛到不动点（`out2 === out3`）——至多重格式化一次后永久稳定，杜绝渐进式损坏。
- 另报：**1-pass-idempotent**（`out1===out2`，一次保存即稳定）与 **exact**（字节级一致）。

## 测试台门槛（交叉验证强化后）
- **lossless**：词/URL token 无丢失 **且无新增 HTML 实体**（escape/entity 丢失现在能被检出）。
- **stable**：收敛到不动点。
- 合成样例新增 **DIALECT** 精确断言（转义、实体、flanking、围栏内表格等给定 expect）。
- 阻断门槛：无 THREW/UNSTABLE/CONTENT-LOSS/ENTITY-ENCODED/DIALECT。

## Round 1 结果（强化测试台）
- 合成方言样例（32 条，含转义/实体/flanking/围栏边界）：**32/32** stable / lossless / 1-pass / exact，PASS。
- 真实语料 `/Users/eric/Documents/Gridea Pro/posts/*.md`（38 篇）：
  **stable 38/38 · lossless 38/38**（合并合成共 70/70）· 1-pass 37/38 · exact 多数；无阻断问题，PASS。

## 交叉验证（35 Agent）发现并已修复的 HIGH 问题
1. **反斜杠转义被静默丢弃**（`\$ \* \| \\` → 整体消失，跨文档内容丢失）→ 新增 `MarkdownEscape.ts`
   注册 'escape' handler 保留字面字符。
2. **正文 `< > &` 被序列化为 HTML 实体**（改写源文件）→ `canonicalize.ts` 在非代码区反解实体。
3. **`==` 缺 flanking**（`a == b and c == d` 误判高亮）→ 对齐 markdown-it-mark：开界后/闭界前非空白。
4. **`~x~` / `^x^` 允许内部空格**（误判上下标）→ 收紧为禁含空白，对齐 markdown-it-sub/-sup。
5. **canonicalize 不识别围栏代码块**（```内表格状文本被规整）→ 全程围栏感知，围栏内原样。
6. **表格单元格 `\|` 端到端丢失**（splitRow 误切、转义被丢）→ 修复 #1 + splitRow 多反引号感知 + 单元格 `|`→`\|`。

## 已实现的方言钩子（render + 往返均验证）
| 构造 | 扩展 | token | 序列化 | 对齐 Gridea 解析器 |
|---|---|---|---|---|
| 高亮 | `Highlight.ts` | `mark` | `==x==` | markdown-it-mark |
| 下标 | `Script.ts` | `sub` | `~x~` | markdown-it-sub |
| 上标 | `Script.ts` | `sup` | `^x^` | markdown-it-sup |
| 行内公式 | `Math.ts` | `inlineMath` | `$x$` | **@iktakahiro/markdown-it-katex 同款界定规则** |
| 块公式 | `Math.ts` | `blockMath` | `$$\nx\n$$` | 同上 |
| Emoji | `Emoji.ts` | `emoji` | `:name:` | markdown-it-emoji（仅命中已知短码） |
| 脚注引用/定义 | `Footnote.ts` | `footnoteRef`/`footnoteDef` | `[^id]` / `[^id]: text` | markdown-it-footnote（基础版） |
| 阅读更多 | `MoreBreak.ts` | `html`(拦截 more 注释) | `<!-- more -->` | 后端 more 正则 |
| 原始 HTML 透传 | `RawHtml.ts` | `html`(兜底) | 原样 | html:true |

**关键防误判**：行内公式严格套用 markdown-it-katex 规则（开界后非空白、闭界前非空白且闭界后非数字），
使货币 `$11 亿`、环境变量 `$PATH`、模板 `${x}` 不被误判为公式（曾会吞掉整段表格）。

## 允许的白名单规范化（确定性、幂等；首存可能一次性重排，之后稳定）
- 表格：单元格统一最小内边距 `| a | b |`、分隔行 `---`、单元格内字面 `|` 转义为 `\|`
  （见 `markdown/canonicalize.ts`，已并入生产 `getMarkdown`）。
- 列表标记 `*`/`+` → `-`；有序标记多空格 `1.  ` → `1. `；主题分割线 `***`/`___` → `---`。
- 嵌套标记顺序 `**[x](y)**` → `[**x**](y)`（渲染等价）；嵌套列表缩进规范化为 2 空格。

## 已知边界（均 lossless + stable，非阻断）
1. 1 篇松散有序列表（`1.  **粗体**：` + 4 空格续行段落）首存重排一次续行缩进，第 2 存即稳定。
2. marked 默认对裸 URL 自动链接（`http://x` → `[http://x](http://x)`），Gridea 预览未启用 linkify。
3. **反斜杠转义规范化**：`\$`→`$` 等（字面字符保留，反斜杠规范化掉）。`$`/`<` 等安全；
   但 `\*`/`\_`（强调符）去掉反斜杠后若成对，可能被重解释为强调——技术博客中罕见，登记为已知边界。
4. **单 `~` 含空格**（`~a b~`）：marked 视为删除线 `~~a b~~`，Gridea 视为字面。无损、首存后一致。
5. 行内原始 HTML 标签会被 marked 规整为等价 Markdown（逐字透传仅保证**块级** HTML）。
6. Emoji 选择器 UI、脚注富文本定义/回跳、公式 KaTeX 编辑面板属后续 UI 轮。
