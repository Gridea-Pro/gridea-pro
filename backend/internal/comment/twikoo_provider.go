package comment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"net/http"
	"time"
)

// TwikooProvider Twikoo 评论提供者
type TwikooProvider struct {
	EnvID string
}

// NewTwikooProvider 创建 Twikoo Provider
func NewTwikooProvider(envID string) *TwikooProvider {
	return &TwikooProvider{EnvID: envID}
}

// Twikoo API Request
type twikooRequest struct {
	Event string `json:"event"`
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

// Twikoo Comment Data Structure
type twikooComment struct {
	ID       string          `json:"id"`
	Nick     string          `json:"nick"`
	Mail     string          `json:"mail"`
	Link     string          `json:"link"`
	Comment  string          `json:"comment"` // HTML content
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

type twikooResponse struct {
	Code int             `json:"code"`
	Data []twikooComment `json:"data"` // For get list
	Msg  string          `json:"msg"`
}

func (p *TwikooProvider) getAPIUrl() string {
	// Twikoo 云函数通常是 EnvID 本身如果是 URL，或者特定的云厂商格式
	// 简单起见，这里假设 EnvID 是完整的云函数 URL，或者用户需要填完整的 URL
	// 实际 Twikoo 客户端逻辑较复杂，这里简化处理，假设用户填的是 Vercel/腾讯云 等 HTTP URL
	return p.EnvID
}

func (p *TwikooProvider) callAPI(ctx context.Context, payload twikooRequest) (*twikooResponse, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.getAPIUrl(), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result twikooResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 && result.Code != 200 { // Twikoo 有时返回 0 或 200 表示成功
		return nil, fmt.Errorf("Twikoo API error: %s", result.Msg)
	}

	return &result, nil
}

// GetAdminComments implementation
func (p *TwikooProvider) GetAdminComments(ctx context.Context, page, pageSize int) (*domain.PaginatedComments, error) {
	// TODO: Implement pagination for Twikoo
	comments, err := p.GetRecentComments(ctx, pageSize)
	if err != nil {
		return nil, err
	}
	return &domain.PaginatedComments{
		Comments:   comments,
		Total:      len(comments),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 1,
	}, nil
}

func (p *TwikooProvider) GetComments(ctx context.Context, articleID string) ([]domain.Comment, error) {
	payload := twikooRequest{
		Event: "comment-get",
		Url:   articleID,
		Admin: true, // 获取全部评论可能需要
	}

	resp, err := p.callAPI(ctx, payload)
	if err != nil {
		return nil, err
	}

	return p.convertComments(resp.Data), nil
}

func (p *TwikooProvider) GetRecentComments(ctx context.Context, limit int) ([]domain.Comment, error) {
	payload := twikooRequest{
		Event:        "get-recent-comments",
		IncludeReply: true,
		PageSize:     limit,
	}

	resp, err := p.callAPI(ctx, payload)
	if err != nil {
		return nil, err
	}

	return p.convertComments(resp.Data), nil
}

func (p *TwikooProvider) convertComments(tComments []twikooComment) []domain.Comment {
	var comments []domain.Comment
	for _, c := range tComments {
		createdAt := time.UnixMilli(c.Created).Format(time.RFC3339)

		comments = append(comments, domain.Comment{
			ID:        c.ID,
			Avatar:    c.Avatar,
			Nickname:  c.Nick,
			URL:       c.Link,
			Content:   c.Comment,
			CreatedAt: createdAt,
			ArticleID: c.Url,
			ParentID:  c.ParentId,
		})

		// 处理嵌套回复
		if len(c.Replies) > 0 {
			comments = append(comments, p.convertComments(c.Replies)...)
		}
	}
	return comments
}

func (p *TwikooProvider) PostComment(ctx context.Context, comment domain.Comment) error {
	payload := twikooRequest{
		Event:   "comment-submit",
		Nick:    comment.Nickname,
		Mail:    "", // 可选
		Link:    comment.URL,
		Comment: comment.Content,
		Url:     comment.ArticleID,
		Pid:     comment.ParentID,
		Rid:     comment.ParentID, // Root ID 简化处理
		Ua:      "Gridea Pro Desktop",
	}

	_, err := p.callAPI(ctx, payload)
	return err
}

func (p *TwikooProvider) DeleteComment(ctx context.Context, commentID string) error {
	// Twikoo 删除通常需要 accessToken，目前 API 暂不支持直接通过公共接口删除
	// 可能需要实现管理端专用接口
	return fmt.Errorf("Twikoo delete not supported yet")
}
