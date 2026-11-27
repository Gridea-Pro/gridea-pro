package comment

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gridea-pro/backend/internal/domain"
)

// TwikooProvider Twikoo 评论提供者。
//
// 说明：Twikoo 云函数（1.6.x / 2.x）统一使用大写事件名
// （GET_RECENT_COMMENTS / COMMENT_GET / COMMENT_SUBMIT），
// 使用小写事件名会被服务端拒绝（返回 code 1001「请更新 Twikoo 云函数至最新版本」）。
// 请求体按 Twikoo 2.x 官方前端约定携带 envId（旧版本会忽略该未知字段）。
type TwikooProvider struct {
	*BaseProvider
	config *TwikooConfig
}

// NewTwikooProvider 创建 Twikoo Provider
func NewTwikooProvider(config *TwikooConfig, proxyURL string, logger *slog.Logger) *TwikooProvider {
	return &TwikooProvider{
		BaseProvider: NewBaseProvider(15*time.Second, proxyURL, logger),
		config:       config,
	}
}

// twikooRequest Twikoo 请求体。各事件的字段合并在此结构体中，统一以
// omitempty 序列化，保证只发送当前事件有意义的字段。
type twikooRequest struct {
	Event string `json:"event"`
	// envId Twikoo 2.x 请求体需携带的环境标识，旧版本忽略此字段。
	EnvId string `json:"envId,omitempty"`
	// get-recent-comments params
	IncludeReply bool `json:"includeReply,omitempty"`
	PageSize     int  `json:"pageSize,omitempty"`
	// comment-get params
	Url      string `json:"url,omitempty"`
	Admin    bool   `json:"admin,omitempty"`
	Page     int    `json:"page,omitempty"`
	ParentId string `json:"parentId,omitempty"`
	// comment-submit params
	Nick    string `json:"nick,omitempty"`
	Mail    string `json:"mail,omitempty"`
	Link    string `json:"link,omitempty"`
	Comment string `json:"comment,omitempty"`
	Pid     string `json:"pid,omitempty"`
	Rid     string `json:"rid,omitempty"`
	Ua      string `json:"ua,omitempty"`
}

// twikooComment Twikoo 评论文档结构。
type twikooComment struct {
	ID       string          `json:"id"`
	Nick     string          `json:"nick"`
	Mail     string          `json:"mail"`
	Link     string          `json:"link"`
	Comment  string          `json:"comment"` // HTML 内容
	Url      string          `json:"url"`
	ParentId string          `json:"pid"`
	Rid      string          `json:"rid"`
	Created  int64           `json:"created"`
	Updated  int64           `json:"updated"`
	Avatar   string          `json:"avatar"`
	IsSpam   bool            `json:"isSpam"`
	Top      bool            `json:"top"`
	Role     string          `json:"role"`
	Replies  []twikooComment `json:"replies"`
}

// twikooResponse Twikoo 响应。
// 不同事件的返回结构并不一致——GET_RECENT_COMMENTS 直接返回 {data:[...]}（无 code），
// COMMENT_GET 返回 {code:0,data:[...]}，出错时返回 {code,message}——因此这里把 Code
// 视为「缺省即 0」，errMessage 同时兼容 2.x 的 message 与旧版的 msg 字段。
type twikooResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"` // Twikoo 2.x
	Msg     string          `json:"msg"`     // 旧版本
	Data    []twikooComment `json:"data"`
}

// errMessage 返回错误描述，优先 2.x 的 message 字段。
func (r *twikooResponse) errMessage() string {
	if r.Message != "" {
		return r.Message
	}
	return r.Msg
}

// getAPIUrl 返回 Twikoo 云函数地址。
// 约定 EnvID 为完整的 URL（如 https://twikoo.example.com），与前端一致。
func (p *TwikooProvider) getAPIUrl() string {
	return p.config.EnvID
}

// callAPI 向 Twikoo 云函数发起一次请求并解析响应。
// 成功判定：HTTP 2xx 且 code ∈ {0, 200}（缺省 code 视为 0），否则返回 ErrProviderError。
func (p *TwikooProvider) callAPI(ctx context.Context, payload twikooRequest) (*twikooResponse, error) {
	// Twikoo 2.x 要求请求体携带 envId（官方前端同样如此）。
	payload.EnvId = p.config.EnvID

	var result twikooResponse
	if err := p.DoJSON(ctx, "POST", p.getAPIUrl(), payload, &result, nil); err != nil {
		return nil, err
	}
	if result.Code != 0 && result.Code != 200 {
		return nil, fmt.Errorf("%w: %s", ErrProviderError, result.errMessage())
	}
	return &result, nil
}

// GetAdminComments 获取管理端评论列表。
// Twikoo 公共 API 未提供独立的管理端列表事件，这里退化为「最新评论」，
// 总数以本次返回条数近似（与原实现保持一致）。
func (p *TwikooProvider) GetAdminComments(ctx context.Context, page, pageSize int) ([]domain.Comment, int64, error) {
	comments, err := p.GetRecentComments(ctx, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return comments, int64(len(comments)), nil
}

// GetComments 获取指定文章（articleID 为文章路径）的评论。
func (p *TwikooProvider) GetComments(ctx context.Context, articleID string) ([]domain.Comment, error) {
	resp, err := p.callAPI(ctx, twikooRequest{
		Event: "COMMENT_GET",
		Url:   articleID,
		Admin: true, // 携带管理标识以获取全部评论（含待审核）
	})
	if err != nil {
		return nil, err
	}
	return p.convertComments(resp.Data), nil
}

// GetRecentComments 获取最新评论。
func (p *TwikooProvider) GetRecentComments(ctx context.Context, limit int) ([]domain.Comment, error) {
	resp, err := p.callAPI(ctx, twikooRequest{
		Event:        "GET_RECENT_COMMENTS",
		IncludeReply: true,
		PageSize:     limit,
	})
	if err != nil {
		return nil, err
	}
	return p.convertComments(resp.Data), nil
}

// convertComments 将 Twikoo 评论转换为领域模型，并递归展开嵌套回复。
func (p *TwikooProvider) convertComments(tComments []twikooComment) []domain.Comment {
	var comments []domain.Comment
	for _, c := range tComments {
		// Twikoo 的 created 为毫秒级 Unix 时间戳
		createdTime := time.UnixMilli(c.Created)

		comments = append(comments, domain.Comment{
			ID:        c.ID,
			Avatar:    c.Avatar,
			Nickname:  c.Nick,
			URL:       c.Link,
			Content:   c.Comment,
			CreatedAt: createdTime,
			ArticleID: c.Url,
			ParentID:  c.ParentId,
		})

		// 展开嵌套回复
		if len(c.Replies) > 0 {
			comments = append(comments, p.convertComments(c.Replies)...)
		}
	}
	return comments
}

// PostComment 发表评论。
func (p *TwikooProvider) PostComment(ctx context.Context, comment *domain.Comment) error {
	_, err := p.callAPI(ctx, twikooRequest{
		Event:   "COMMENT_SUBMIT",
		Nick:    comment.Nickname,
		Mail:    "", // 可选
		Link:    comment.URL,
		Comment: comment.Content,
		Url:     comment.ArticleID,
		Pid:     comment.ParentID,
		Rid:     comment.ParentID, // Root ID 简化处理
		Ua:      "Gridea Pro Desktop",
	})
	return err
}

// DeleteComment Twikoo 删除评论需要管理端 accessToken，公共接口暂不支持。
func (p *TwikooProvider) DeleteComment(ctx context.Context, commentID string) error {
	return fmt.Errorf("%w: Twikoo delete not supported yet", ErrNotImplemented)
}
