package comment

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ValineProvider LeanCloud/Valine 评论提供者
type ValineProvider struct {
	AppID      string
	AppKey     string
	MasterKey  string
	ServerURLs string
}

// NewValineProvider 创建 Valine Provider
func NewValineProvider(appID, appKey, masterKey, serverURLs string) *ValineProvider {
	if serverURLs == "" {
		// 默认 LeanCloud API 域名 (主要用于国际版，国内版通常需要自定义域名)
		serverURLs = "https://leancloud.cn"
	}
	return &ValineProvider{
		AppID:      appID,
		AppKey:     appKey,
		MasterKey:  masterKey,
		ServerURLs: serverURLs,
	}
}

// LeanCloud Comment 结构
type leanCloudComment struct {
	ObjectId  string `json:"objectId"`
	Nick      string `json:"nick"`
	Comment   string `json:"comment"`
	Mail      string `json:"mail"`
	Link      string `json:"link"`
	Pid       string `json:"pid"`
	Rid       string `json:"rid"`
	Pnick     string `json:"pnick"`
	Url       string `json:"url"` // Article URL path
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// LeanCloud Response 结构
type leanCloudResponse struct {
	Results []leanCloudComment `json:"results"`
}

func (p *ValineProvider) GetComments(ctx context.Context, articleID string) ([]domain.Comment, error) {
	// Valine 使用 url 作为文章标识
	//构建查询条件: {"url": articleID}
	where := fmt.Sprintf(`{"url":"%s"}`, articleID)
	params := url.Values{}
	params.Add("where", where)
	params.Add("order", "-createdAt")

	apiURL := fmt.Sprintf("%s/1.1/classes/Comment?%s", p.ServerURLs, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	p.setHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LeanCloud API error: %d %s", resp.StatusCode, string(body))
	}

	var result leanCloudResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	comments := make([]domain.Comment, 0, len(result.Results))
	for _, c := range result.Results {
		comments = append(comments, domain.Comment{
			ID:         c.ObjectId,
			Nickname:   c.Nick,
			URL:        c.Link,
			Content:    c.Comment,
			CreatedAt:  c.CreatedAt,
			ArticleID:  c.Url,
			ParentID:   c.Pid,
			ParentNick: c.Pnick,
			Email:      c.Mail,
			Avatar:     p.getGravatar(c.Mail),
		})
	}

	return comments, nil
}

func (p *ValineProvider) GetRecentComments(ctx context.Context, limit int) ([]domain.Comment, error) {
	params := url.Values{}
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("order", "-createdAt")

	apiURL := fmt.Sprintf("%s/1.1/classes/Comment?%s", p.ServerURLs, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	p.setHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LeanCloud API error: %d %s", resp.StatusCode, string(body))
	}

	var result leanCloudResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	comments := make([]domain.Comment, 0, len(result.Results))
	for _, c := range result.Results {
		comments = append(comments, domain.Comment{
			ID:         c.ObjectId,
			Nickname:   c.Nick,
			URL:        c.Link,
			Content:    c.Comment,
			CreatedAt:  c.CreatedAt,
			ArticleID:  c.Url,
			ParentID:   c.Pid,
			ParentNick: c.Pnick,
			Email:      c.Mail,
			Avatar:     p.getGravatar(c.Mail),
			// IsNew: true, // TODO: 根据本地记录判断是否新评论
		})
	}

	return comments, nil
}

func (p *ValineProvider) GetAdminComments(ctx context.Context, page, pageSize int) (*domain.PaginatedComments, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}

	params := url.Values{}
	params.Add("limit", fmt.Sprintf("%d", pageSize))
	params.Add("skip", fmt.Sprintf("%d", (page-1)*pageSize))
	params.Add("order", "-createdAt")
	params.Add("count", "1") // 请求返回总量

	apiURL := fmt.Sprintf("%s/1.1/classes/Comment?%s", p.ServerURLs, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	p.setHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LeanCloud API error: %d %s", resp.StatusCode, string(body))
	}

	// LeanCloud Response with Count
	type leanCloudCountResponse struct {
		Results []leanCloudComment `json:"results"`
		Count   int                `json:"count"`
	}

	var result leanCloudCountResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	comments := make([]domain.Comment, 0, len(result.Results))
	for _, c := range result.Results {
		comments = append(comments, domain.Comment{
			ID:         c.ObjectId,
			Nickname:   c.Nick,
			URL:        c.Link,
			Content:    c.Comment,
			CreatedAt:  c.CreatedAt,
			ArticleID:  c.Url,
			ParentID:   c.Pid,
			ParentNick: c.Pnick,
			Email:      c.Mail,
			Avatar:     p.getGravatar(c.Mail),
			// IsNew: true, // TODO: logic
		})
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = (result.Count + pageSize - 1) / pageSize
	}

	return &domain.PaginatedComments{
		Comments:   comments,
		Total:      result.Count,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (p *ValineProvider) PostComment(ctx context.Context, comment domain.Comment) error {
	// 构造 LeanCloud Comment 对象
	lcComment := map[string]interface{}{
		"nick":    comment.Nickname,
		"comment": comment.Content,
		"url":     comment.ArticleID,
	}

	// 如果是回复，需要获取父评论信息以正确设置 pid 和 rid
	if comment.ParentID != "" {
		parent, err := p.getCommentByID(ctx, comment.ParentID)
		if err != nil {
			return fmt.Errorf("failed to fetch parent comment: %w", err)
		}

		lcComment["pid"] = parent.ObjectId
		if parent.Rid != "" {
			lcComment["rid"] = parent.Rid
		} else {
			lcComment["rid"] = parent.ObjectId
		}
		// 还可以设置被回复人的昵称，有些主题可能需要
		lcComment["pnick"] = parent.Nick
	}

	// 模拟 UserAgent，包含标准头部以便 parser 识别，同时加入 Gridea Pro 标识
	// 格式参考: Mozilla/5.0 (Platform; Security; OS-or-CPU; Localization; rv: revision-version-number) Gecko/gecko-trail User-Agent-String
	lcComment["ua"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/26.2 Safari/605.1.15"

	if comment.Email != "" {
		lcComment["mail"] = comment.Email
	}
	if comment.URL != "" {
		lcComment["link"] = comment.URL
	}

	jsonData, err := json.Marshal(lcComment)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("%s/1.1/classes/Comment", p.ServerURLs)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	p.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LeanCloud API error: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

func (p *ValineProvider) getCommentByID(ctx context.Context, id string) (*leanCloudComment, error) {
	apiURL := fmt.Sprintf("%s/1.1/classes/Comment/%s", p.ServerURLs, id)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	p.setHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var comment leanCloudComment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return nil, err
	}
	return &comment, nil
}

func (p *ValineProvider) DeleteComment(ctx context.Context, commentID string) error {
	apiURL := fmt.Sprintf("%s/1.1/classes/Comment/%s", p.ServerURLs, commentID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", apiURL, nil)
	if err != nil {
		return err
	}

	p.setHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LeanCloud API error: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

func (p *ValineProvider) setHeaders(req *http.Request) {
	req.Header.Set("X-LC-Id", p.AppID)
	if p.MasterKey != "" {
		req.Header.Set("X-LC-Key", fmt.Sprintf("%s,master", p.MasterKey))
	} else {
		req.Header.Set("X-LC-Key", p.AppKey)
	}
}

func (p *ValineProvider) getGravatar(email string) string {
	if email == "" {
		return ""
	}
	email = strings.TrimSpace(strings.ToLower(email))
	hash := md5.Sum([]byte(email))
	return fmt.Sprintf("https://cravatar.cn/avatar/%s?d=mp&v=1.4.14", hex.EncodeToString(hash[:]))
}
