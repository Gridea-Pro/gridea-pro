package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service/ai"
)

// 内置 Key 调用频率限制（仅对使用内置免费模型的用户生效）
const (
	builtInDailyLimit  = 20 // 每天最多 20 次
	builtInMinuteLimit = 5  // 每分钟最多 5 次
)

// AIService AI 功能服务
type AIService struct {
	repo        domain.AISettingRepository
	settingRepo domain.SettingRepository
	usageRepo   domain.AIUsageRepository
	usageMu     sync.Mutex
}

func NewAIService(repo domain.AISettingRepository, settingRepo domain.SettingRepository, usageRepo domain.AIUsageRepository) *AIService {
	return &AIService{repo: repo, settingRepo: settingRepo, usageRepo: usageRepo}
}

// httpClient 根据当前代理配置返回合适的 HTTP client
func (s *AIService) httpClient(ctx context.Context) *http.Client {
	if s.settingRepo != nil {
		setting, err := s.settingRepo.GetSetting(ctx)
		if err == nil && setting.ProxyEnabled && setting.ProxyURL != "" {
			return newHTTPClient(setting.ProxyURL)
		}
	}
	return &http.Client{Timeout: 30 * time.Second}
}

var customBaseURLProviders = map[string]struct{}{
	"custom-openai":    {},
	"custom-anthropic": {},
}

func isCustomBaseURLProvider(providerID string) bool {
	_, ok := customBaseURLProviders[strings.TrimSpace(providerID)]
	return ok
}

func newCustomBaseURLProvider(providerID, baseURL string) ai.Provider {
	switch strings.TrimSpace(providerID) {
	case "custom-openai":
		return ai.NewOpenAICompatibleProvider(baseURL)
	case "custom-anthropic":
		return ai.NewAnthropicCompatibleProvider(baseURL)
	default:
		return nil
	}
}

// resolveProvider 根据当前 AI 设置返回 (provider, model, apiKey, isBuiltIn, error)
func (s *AIService) resolveProvider(ctx context.Context) (ai.Provider, string, string, bool, error) {
	setting, _ := s.repo.GetAISetting(ctx)

	// 默认使用内置模型
	if setting.Mode == "" || setting.Mode == domain.AIModeBuiltIn {
		key := ai.DecryptBuiltInKey()
		if key == "" {
			return nil, "", "", true, errors.New("内置模型暂不可用")
		}
		return ai.NewBuiltInProvider(), ai.PickBuiltInModel(), key, true, nil
	}

	// 自定义模式：从 customs[activeProvider] 取配置
	activeProvider := strings.TrimSpace(setting.ActiveProvider)
	if activeProvider == "" {
		return nil, "", "", false, errors.New("请先在「偏好设置 → AI 配置」中选择模型厂商")
	}
	cfg, ok := setting.Customs[activeProvider]
	if !ok {
		return nil, "", "", false, errors.New("请先在「偏好设置 → AI 配置」中完成当前厂商的配置")
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, "", "", false, errors.New("请先在「偏好设置 → AI 配置」中填写 API Key")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, "", "", false, errors.New("请先在「偏好设置 → AI 配置」中选择模型")
	}
	if isCustomBaseURLProvider(activeProvider) {
		baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
		if baseURL == "" {
			return nil, "", "", false, errors.New("请先在「偏好设置 → AI 配置」中填写 Base URL")
		}
		return newCustomBaseURLProvider(activeProvider, baseURL), strings.TrimSpace(cfg.Model), strings.TrimSpace(cfg.APIKey), false, nil
	}
	provider, _, err := ai.NewProvider(activeProvider)
	if err != nil {
		return nil, "", "", false, err
	}
	return provider, strings.TrimSpace(cfg.Model), strings.TrimSpace(cfg.APIKey), false, nil
}

// reserveBuiltInQuota 原子地"检查并预占"一次内置 Key 配额（检查+自增在同一把锁内完成）。
// 这样并发请求不会都通过检查、再各自计数导致超配额。配额已满则返回错误且不占用。
// 错误信息使用 [DAILY_LIMIT] / [RATE_LIMIT] 前缀，供前端 i18n 匹配。
func (s *AIService) reserveBuiltInQuota(ctx context.Context) error {
	s.usageMu.Lock()
	defer s.usageMu.Unlock()

	// 读用量失败（IO/权限/损坏）时 fail-closed：宁可拒绝本次免费调用，也不放行不计数——
	// 否则读失败就等于无限额度。not-exist 由 repo 内部当新设备处理，不会走到这里。
	usage, err := s.usageRepo.GetAIUsage(ctx)
	if err != nil {
		return fmt.Errorf("[QUOTA_UNAVAILABLE] 无法读取免费额度用量，请稍后再试：%w", err)
	}
	now := time.Now()
	today := now.Format("2006-01-02")
	minute := now.Format("2006-01-02 15:04")

	if usage.Date != today {
		usage.Date = today
		usage.DailyCount = 0
	}
	if usage.Minute != minute {
		usage.Minute = minute
		usage.MinuteCount = 0
	}

	if usage.DailyCount >= builtInDailyLimit {
		return fmt.Errorf("[DAILY_LIMIT] 今日免费额度已用完（%d 次/天），请明日再试，或在「偏好设置 → AI 配置」中切换为自定义模型", builtInDailyLimit)
	}
	if usage.MinuteCount >= builtInMinuteLimit {
		return fmt.Errorf("[RATE_LIMIT] 调用过于频繁，请稍后再试（限制 %d 次/分钟）", builtInMinuteLimit)
	}
	// 预占：检查通过即自增计数并落盘，后续并发请求会看到已占用的额度。
	usage.DailyCount++
	usage.MinuteCount++
	// 落盘失败视为预占失败：否则计数没持久化，并发/重启后额度形同虚设。
	if err := s.usageRepo.SaveAIUsage(ctx, usage); err != nil {
		return fmt.Errorf("[QUOTA_UNAVAILABLE] 无法记录免费额度用量，请稍后再试：%w", err)
	}
	return nil
}

// refundBuiltInUsage 在预占配额后、真实 AI 调用失败（未实际消耗）时回滚一次计数。
func (s *AIService) refundBuiltInUsage(ctx context.Context) {
	s.usageMu.Lock()
	defer s.usageMu.Unlock()

	usage, _ := s.usageRepo.GetAIUsage(ctx)
	now := time.Now()
	today := now.Format("2006-01-02")
	minute := now.Format("2006-01-02 15:04")
	// 仅在同一天/同一分钟窗口内回滚才有意义（跨窗口计数已重置）。
	if usage.Date == today && usage.DailyCount > 0 {
		usage.DailyCount--
	}
	if usage.Minute == minute && usage.MinuteCount > 0 {
		usage.MinuteCount--
	}
	_ = s.usageRepo.SaveAIUsage(ctx, usage)
}

// slugPrompt 构建生成 Slug 的提示词
func slugPrompt(title string) string {
	return fmt.Sprintf(
		"Generate an SEO-friendly English URL slug from the blog title.\n\n"+
			"Goal: Both search engines and human readers should immediately understand "+
			"what the article is about just by looking at the slug. The slug must read "+
			"like a natural English phrase, not a word-for-word translation.\n\n"+
			"Process (think before writing):\n"+
			"1. Identify the SINGLE main idea of the title (one sentence in your head).\n"+
			"2. If the title has a subtitle (after —, ——, :, or 、), treat it as background "+
			"context only — DO NOT translate it word by word. Use it just to disambiguate the main idea.\n"+
			"3. Express that main idea as a short English phrase: subject + action + (optional context).\n"+
			"4. Trim to 4–8 words. NEVER exceed 8 words. Aim for 5–6.\n\n"+
			"Rules:\n"+
			"- HARD LIMIT: 8 words maximum. Count the words before outputting.\n"+
			"- Drop filler words: a, an, the, is, are, that, how, what, something, anything, everyone, every\n"+
			"- Keep short connectors only when they aid clarity: vs, with, for, to, in\n"+
			"- Brand/tech names must be exact and lowercased (e.g. macos, docker, nextjs, gpt-4, wechat, claude-code, gridea)\n"+
			"- Keep version numbers and years when present (e.g. gpt-4, 2026)\n"+
			"- All lowercase, hyphens as separators, no special characters, no trailing hyphen\n"+
			"- NEVER translate emotional/rhetorical phrases literally (e.g. 「每个想写点什么的人」「都值得」「让世界更美好」)\n\n"+
			"Examples:\n"+
			"- 我用 Claude Code 重构了整个项目的代码 → refactor-entire-project-with-claude-code\n"+
			"- Arc 和 Chrome 哪个更适合开发者日常使用？ → arc-vs-chrome-for-developers\n"+
			"- 独立开发者出海第一步：选对收款工具 → indie-developer-global-payment-tools\n"+
			"- The Best Markdown Editors for Developers in 2026 → best-markdown-editors-for-developers-2026\n"+
			"- 如何用 Docker 部署 Next.js 到生产环境 → deploy-nextjs-to-production-with-docker\n"+
			"- 从零搭建一个个人博客系统 → build-personal-blog-system-from-scratch\n"+
			"- 我为什么复活了 Gridea —— 每个想写点什么的人，都值得一个更简单的开始 → why-i-revived-gridea-for-simpler-writing\n"+
			"- ChatGPT 改变了我的工作方式：从效率工具到思考伙伴 → how-chatgpt-changed-my-workflow\n\n"+
			"Output ONLY the slug string, nothing else. No quotes, no explanation.\n\n"+
			"Title: %s",
		title,
	)
}

// sanitizeSlug 清理模型输出，只保留字母/数字/连字符
func sanitizeSlug(raw string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(raw)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	return strings.Trim(b.String(), "-")
}

// GenerateSlug 根据文章标题生成 SEO 友好的英文 Slug
func (s *AIService) GenerateSlug(ctx context.Context, title string) (string, error) {
	if strings.TrimSpace(title) == "" {
		return "", errors.New("文章标题不能为空")
	}

	provider, model, apiKey, isBuiltIn, err := s.resolveProvider(ctx)
	if err != nil {
		return "", err
	}

	// 仅对使用内置模型的用户做本地配额检查（原子预占，避免并发绕过）
	if isBuiltIn {
		if err := s.reserveBuiltInQuota(ctx); err != nil {
			return "", err
		}
	}

	req := ai.ChatRequest{
		Model:       model,
		Prompt:      slugPrompt(title),
		Temperature: 0.1,
		MaxTokens:   80,
	}
	raw, err := provider.Chat(ctx, req, apiKey, s.httpClient(ctx))
	if err != nil {
		// 真实 AI 调用失败（未实际消耗），回滚上面预占的配额。
		if isBuiltIn {
			s.refundBuiltInUsage(ctx)
		}
		return "", err
	}

	result := sanitizeSlug(raw)
	if result == "" {
		return "", errors.New("生成的 Slug 无效，请重试")
	}
	// 成功：配额已在 reserveBuiltInQuota 预占，无需再计数。
	return result, nil
}

// completePrompt 行内 AI 续写提示词（Fill-in-the-Middle 风格）
func completePrompt(prefix, suffix string) string {
	return fmt.Sprintf(
		"You are an inline writing assistant embedded in a Markdown blog editor.\n"+
			"Continue the text naturally at the cursor position marked by <CURSOR>.\n\n"+
			"Rules:\n"+
			"- Output ONLY the continuation that should be inserted at <CURSOR>. No explanation, no quotes, no code fences.\n"+
			"- Keep the SAME language as the surrounding text.\n"+
			"- Write a short, natural continuation (a clause or one sentence). Do not repeat the existing text.\n"+
			"- Match the existing tone, Markdown style and formatting.\n"+
			"- If the text before the cursor already ends a sentence, you may start a new one.\n\n"+
			"Text before cursor:\n%s<CURSOR>\n\nText after cursor:\n%s\n\nContinuation:",
		prefix, suffix,
	)
}

// polishPrompt 文本润色提示词
func polishPrompt(text string) string {
	return fmt.Sprintf(
		"You are a writing assistant for a Markdown blog editor.\n"+
			"Polish the following text: improve clarity, grammar, flow and word choice.\n\n"+
			"Rules:\n"+
			"- Keep the ORIGINAL meaning and the SAME language.\n"+
			"- Preserve Markdown syntax (links, emphasis, code, lists, etc.).\n"+
			"- Do NOT add new information or commentary.\n"+
			"- Output ONLY the polished text, with no quotes, no explanation, no code fences.\n\n"+
			"Text:\n%s\n\nPolished:",
		text,
	)
}

// stripWrapping 去除模型输出常见的包裹（首尾引号 / ``` 代码围栏）
func stripWrapping(raw string) string {
	s := strings.TrimSpace(raw)
	// 去除整体代码围栏：需确有闭合 ``` 才剥离，避免误伤正文中的反引号
	if strings.HasPrefix(s, "```") {
		if nl := strings.IndexByte(s, '\n'); nl >= 0 {
			body := s[nl+1:]
			if end := strings.LastIndex(body, "```"); end >= 0 {
				s = strings.TrimSpace(body[:end])
			}
		}
	}
	// 去除整体包裹引号：仅当首尾为同种引号、且内部不再出现该引号（即确为整段包裹）
	if len(s) >= 2 {
		q := s[0]
		if (q == '"' || q == '\'') && s[len(s)-1] == q {
			if inner := s[1 : len(s)-1]; !strings.ContainsRune(inner, rune(q)) {
				s = inner
			}
		}
	}
	return s
}

// Complete 行内 AI 续写：给定光标前后文，返回应插入光标处的补全文本
func (s *AIService) Complete(ctx context.Context, prefix, suffix string) (string, error) {
	if strings.TrimSpace(prefix) == "" && strings.TrimSpace(suffix) == "" {
		return "", errors.New("上下文为空")
	}

	provider, model, apiKey, isBuiltIn, err := s.resolveProvider(ctx)
	if err != nil {
		return "", err
	}
	if isBuiltIn {
		if err := s.reserveBuiltInQuota(ctx); err != nil {
			return "", err
		}
	}

	req := ai.ChatRequest{
		Model:       model,
		Prompt:      completePrompt(prefix, suffix),
		Temperature: 0.3,
		MaxTokens:   160,
	}
	raw, err := provider.Chat(ctx, req, apiKey, s.httpClient(ctx))
	if err != nil {
		if isBuiltIn {
			s.refundBuiltInUsage(ctx)
		}
		return "", err
	}

	return stripWrapping(raw), nil
}

// Polish 文本润色：返回润色后的文本
func (s *AIService) Polish(ctx context.Context, text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", errors.New("待润色文本为空")
	}

	provider, model, apiKey, isBuiltIn, err := s.resolveProvider(ctx)
	if err != nil {
		return "", err
	}
	if isBuiltIn {
		if err := s.reserveBuiltInQuota(ctx); err != nil {
			return "", err
		}
	}

	// 输出 token 上限按输入长度放宽（粗略：rune 数 + 余量）
	maxTokens := len([]rune(text)) + 200
	if maxTokens > 2000 {
		maxTokens = 2000
	}

	req := ai.ChatRequest{
		Model:       model,
		Prompt:      polishPrompt(text),
		Temperature: 0.4,
		MaxTokens:   maxTokens,
	}
	raw, err := provider.Chat(ctx, req, apiKey, s.httpClient(ctx))
	if err != nil {
		if isBuiltIn {
			s.refundBuiltInUsage(ctx)
		}
		return "", err
	}

	result := stripWrapping(raw)
	if result == "" {
		return "", errors.New("润色结果为空，请重试")
	}

	return result, nil
}

// summaryPrompt 摘要生成提示词
func summaryPrompt(content string) string {
	return "你是一位中文博客编辑。请为下面的文章生成一段摘要，要求：\n" +
		"1. 100 字以内，单段纯文本；\n" +
		"2. 概括文章核心内容与价值，语气自然，适合作为博客列表页导语；\n" +
		"3. 不要使用 Markdown 语法、不要引号包裹、不要「摘要：」前缀，直接输出摘要正文。\n\n" +
		"文章内容：\n" + content
}

// Summary 生成文章摘要：输入正文（markdown/纯文本），返回 100 字内摘要
func (s *AIService) Summary(ctx context.Context, content string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", errors.New("文章内容为空")
	}

	provider, model, apiKey, isBuiltIn, err := s.resolveProvider(ctx)
	if err != nil {
		return "", err
	}
	if isBuiltIn {
		if err := s.reserveBuiltInQuota(ctx); err != nil {
			return "", err
		}
	}

	// 输入过长截断，避免超 token（摘要看前文即可）
	runes := []rune(content)
	if len(runes) > 6000 {
		content = string(runes[:6000])
	}

	req := ai.ChatRequest{
		Model:       model,
		Prompt:      summaryPrompt(content),
		Temperature: 0.5,
		MaxTokens:   400,
	}
	raw, err := provider.Chat(ctx, req, apiKey, s.httpClient(ctx))
	if err != nil {
		if isBuiltIn {
			s.refundBuiltInUsage(ctx)
		}
		return "", err
	}

	result := stripWrapping(raw)
	if result == "" {
		return "", errors.New("摘要生成为空，请重试")
	}

	return result, nil
}

// TestConnection 测试自定义厂商的连接性（最小 chat 请求）
func (s *AIService) TestConnection(ctx context.Context, providerID, model, apiKey string) error {
	return s.TestConnectionWithBaseURL(ctx, providerID, model, apiKey, "")
}

// TestConnectionWithBaseURL 测试自定义厂商的连接性，自定义兼容厂商可传入 Base URL
func (s *AIService) TestConnectionWithBaseURL(ctx context.Context, providerID, model, apiKey, baseURL string) error {
	if strings.TrimSpace(providerID) == "" {
		return errors.New("请选择模型厂商")
	}
	if strings.TrimSpace(apiKey) == "" {
		return errors.New("请填写 API Key")
	}
	if strings.TrimSpace(model) == "" {
		return errors.New("请选择模型")
	}
	var provider ai.Provider
	if isCustomBaseURLProvider(providerID) {
		baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
		if baseURL == "" {
			return errors.New("请填写 Base URL")
		}
		provider = newCustomBaseURLProvider(providerID, baseURL)
	} else {
		p, _, err := ai.NewProvider(providerID)
		if err != nil {
			return err
		}
		provider = p
	}
	req := ai.ChatRequest{
		Model:       strings.TrimSpace(model),
		Prompt:      "hi",
		Temperature: 0.0,
		MaxTokens:   1,
	}
	_, err := provider.Chat(ctx, req, strings.TrimSpace(apiKey), s.httpClient(ctx))
	return err
}

// ListProviderModels 拉取指定厂商的真实模型列表
func (s *AIService) ListProviderModels(ctx context.Context, providerID, apiKey string) ([]string, error) {
	return s.ListProviderModelsWithBaseURL(ctx, providerID, apiKey, "")
}

// ListProviderModelsWithBaseURL 拉取指定厂商的真实模型列表，自定义兼容厂商可传入 Base URL
func (s *AIService) ListProviderModelsWithBaseURL(ctx context.Context, providerID, apiKey, baseURL string) ([]string, error) {
	if strings.TrimSpace(providerID) == "" {
		return nil, errors.New("请选择模型厂商")
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("请先填写 API Key")
	}
	var provider ai.Provider
	if isCustomBaseURLProvider(providerID) {
		baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
		if baseURL == "" {
			return nil, errors.New("请填写 Base URL")
		}
		provider = newCustomBaseURLProvider(providerID, baseURL)
	} else {
		p, _, err := ai.NewProvider(providerID)
		if err != nil {
			return nil, err
		}
		provider = p
	}
	return provider.ListModels(ctx, strings.TrimSpace(apiKey), s.httpClient(ctx))
}

// GetProviderRegistry 返回所有自定义厂商的元信息（前端下拉框使用）
func (s *AIService) GetProviderRegistry() []ai.ProviderInfo {
	return ai.AllProviders()
}

// GetBuiltInModels 返回内置免费模型清单（前端展示用）
func (s *AIService) GetBuiltInModels() []string {
	return ai.BuiltInModels()
}
