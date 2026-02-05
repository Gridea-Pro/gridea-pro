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

// WalineProvider Waline 评论提供者
type WalineProvider struct {
	AppID      string
	AppKey     string
	MasterKey  string
	ServerURLs string
}

// NewWalineProvider 创建 Waline Provider
func NewWalineProvider(appID, appKey, masterKey, serverURLs string) *WalineProvider {
	return &WalineProvider{
		AppID:      appID,
		AppKey:     appKey,
		MasterKey:  masterKey,
		ServerURLs: serverURLs,
	}
}

// Waline Comment Response
type walineResponse struct {
	Errno  int             `json:"errno"`
	Errmsg interface{}     `json:"errmsg"`
	Data   json.RawMessage `json:"data"` // List or Object depending on context
}

// Waline List Response Data
type walineListData struct {
	Data  []walineComment `json:"data"`
	Count int             `json:"count"`
}

type walineComment struct {
	ObjectId  interface{} `json:"objectId"` // Can be string or number
	Nick      string      `json:"nick"`
	Comment   string      `json:"comment"`
	Mail      string      `json:"mail"`
	Link      string      `json:"link"`
	Pid       interface{} `json:"pid"` // Can be string or number
	Rid       interface{} `json:"rid"` // Root ID
	Url       string      `json:"url"`
	CreatedAt interface{} `json:"createdAt"` // Can be string or number (timestamp)
	// Waline specific
	Type   string `json:"type"`
	Status string `json:"status"`
}

// Helper to convert interface{} to string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%.0f", val)
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case map[string]interface{}:
		// Handle erroneous object return in errmsg or others
		b, _ := json.Marshal(val)
		return string(b)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// parseWalineData parses the data field which can be a list or an object
func parseWalineData(data json.RawMessage) ([]walineComment, int, error) {
	var listData walineListData
	if len(data) == 0 || string(data) == "null" {
		return []walineComment{}, 0, nil
	}

	// Try parsing as Object (paginated)
	if err := json.Unmarshal(data, &listData); err == nil {
		if listData.Data != nil {
			return listData.Data, listData.Count, nil
		}
	}

	// Try parsing as Array (direct list)
	var list []walineComment
	if err := json.Unmarshal(data, &list); err == nil {
		return list, len(list), nil
	}

	return nil, 0, fmt.Errorf("failed to parse waline data: %s", string(data))
}

// GetAdminComments implementation
func (p *WalineProvider) GetAdminComments(ctx context.Context, page, pageSize int) (*domain.PaginatedComments, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}

	serverURL := strings.TrimSuffix(p.ServerURLs, "/")
	apiURL := fmt.Sprintf("%s/api/comment?type=list&page=%d&pageSize=%d", serverURL, page, pageSize)

	// Use executeAuthRequest to handle potential Auth retry logic
	resp, err := p.executeAuthRequest(ctx, "GET", apiURL, nil)
	if err != nil {
		if strings.Contains(err.Error(), "Waline Logged Auth Failed (401)") {
			// Fallback logic for 401
			fmt.Printf("[Waline] Admin Auth failed (401), falling back to public recent comments.\n")
			recentComments, err := p.GetRecentComments(ctx, pageSize)
			if err != nil {
				return nil, fmt.Errorf("Waline 认证失败且无法获取公开评论: %v", err)
			}
			return &domain.PaginatedComments{
				Comments:   recentComments,
				Total:      len(recentComments),
				Page:       1,
				PageSize:   pageSize,
				TotalPages: 1,
			}, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	// resp is guaranteed to be non-401 by executeAuthRequest unless it returns the specific error

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Waline API error status: %d", resp.StatusCode)
	}

	var result walineResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Errno != 0 {
		return nil, fmt.Errorf("Waline API error: %s", toString(result.Errmsg))
	}

	commentsList, count, err := parseWalineData(result.Data)
	if err != nil {
		return nil, err
	}

	var comments []domain.Comment
	for _, c := range commentsList {
		// Fix protocol-relative URLs in content for Wails
		content := c.Comment
		content = strings.ReplaceAll(content, "src=\"//", "src=\"https://")
		content = strings.ReplaceAll(content, "src='//", "src='https://")

		comments = append(comments, domain.Comment{
			ID:        toString(c.ObjectId),
			Nickname:  c.Nick,
			URL:       c.Link,
			Content:   content,
			CreatedAt: toString(c.CreatedAt),
			ArticleID: c.Url,
			ParentID:  toString(c.Pid),
			Email:     c.Mail,
			Avatar:    p.getGravatar(c.Mail),
		})
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = (count + pageSize - 1) / pageSize
	}

	return &domain.PaginatedComments{
		Comments:   comments,
		Total:      count,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (p *WalineProvider) GetComments(ctx context.Context, articleID string) ([]domain.Comment, error) {
	serverURL := strings.TrimSuffix(p.ServerURLs, "/")
	apiURL := fmt.Sprintf("%s/api/comment?path=%s", serverURL, url.QueryEscape(articleID))

	fmt.Printf("[Waline] GetComments Request: %s\n", apiURL)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("[Waline] GetComments Request Failed: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[Waline] GetComments HTTP Error: %d\n", resp.StatusCode)
		return nil, fmt.Errorf("Waline API error status: %d", resp.StatusCode)
	}

	var result walineResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("[Waline] GetComments JSON Decode Error: %v\n", err)
		return nil, err
	}

	if result.Errno != 0 {
		fmt.Printf("[Waline] GetComments API Error: %s\n", toString(result.Errmsg))
		return nil, fmt.Errorf("Waline API error: %s", toString(result.Errmsg))
	}

	commentsList, _, err := parseWalineData(result.Data)
	if err != nil {
		fmt.Printf("[Waline] GetComments Data Parse Error: %v | Raw: %s\n", err, string(result.Data))
		// Don't fail hard, return empty if parsing fails but errno is 0?
		// Better to fail so we know.
		return nil, err
	}

	var comments []domain.Comment
	for _, c := range commentsList {
		// Fix protocol-relative URLs in content
		content := c.Comment
		content = strings.ReplaceAll(content, "src=\"//", "src=\"https://")
		content = strings.ReplaceAll(content, "src='//", "src='https://")

		comments = append(comments, domain.Comment{
			ID:        toString(c.ObjectId),
			Nickname:  c.Nick,
			URL:       c.Link,
			Content:   content,
			CreatedAt: toString(c.CreatedAt),
			ArticleID: c.Url,
			ParentID:  toString(c.Pid),
			Email:     c.Mail,
			Avatar:    p.getGravatar(c.Mail),
		})
	}
	return comments, nil
}

func (p *WalineProvider) GetRecentComments(ctx context.Context, limit int) ([]domain.Comment, error) {
	serverURL := strings.TrimSuffix(p.ServerURLs, "/")
	apiURL := fmt.Sprintf("%s/api/comment?type=recent&count=%d", serverURL, limit)

	fmt.Printf("[Waline] GetRecentComments Request: %s\n", apiURL)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("[Waline] GetRecentComments Request Failed: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[Waline] GetRecentComments HTTP Error: %d\n", resp.StatusCode)
		return nil, fmt.Errorf("Waline API error status: %d", resp.StatusCode)
	}

	var result walineResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("[Waline] GetRecentComments JSON Decode Error: %v\n", err)
		return nil, err
	}

	if result.Errno != 0 {
		fmt.Printf("[Waline] GetRecentComments API Error: %s\n", toString(result.Errmsg))
		return nil, fmt.Errorf("Waline API error: %s", toString(result.Errmsg))
	}

	commentsList, _, err := parseWalineData(result.Data)
	if err != nil {
		fmt.Printf("[Waline] GetRecentComments Data Parse Error: %v | Raw: %s\n", err, string(result.Data))
		return nil, err
	}

	var comments []domain.Comment
	for _, c := range commentsList {
		// Fix protocol-relative URLs in content
		content := c.Comment
		content = strings.ReplaceAll(content, "src=\"//", "src=\"https://")
		content = strings.ReplaceAll(content, "src='//", "src='https://")

		comments = append(comments, domain.Comment{
			ID:        toString(c.ObjectId),
			Nickname:  c.Nick,
			URL:       c.Link,
			Content:   content,
			CreatedAt: toString(c.CreatedAt),
			ArticleID: c.Url,
			ParentID:  toString(c.Pid),
			Email:     c.Mail,
			Avatar:    p.getGravatar(c.Mail),
		})
	}
	return comments, nil
}

func (p *WalineProvider) PostComment(ctx context.Context, comment domain.Comment) error {
	payload := map[string]interface{}{
		"nick":    comment.Nickname,
		"comment": comment.Content,
		"url":     comment.ArticleID,
		"ua":      "Gridea Pro",
	}

	if comment.Email != "" {
		payload["mail"] = comment.Email
	}
	if comment.URL != "" {
		payload["link"] = comment.URL
	}

	if comment.ParentID != "" {
		payload["pid"] = comment.ParentID
		// Waline generally handles rid automatically or doesn't strictly require it if pid is present,
		// but passing it if known is good.
		// We lack `rid` in domain.Comment currently, so we rely on Waline backend logic.
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// If Admin Token is available, we use executeAuthRequest for robustness (though strictly optional for public post)
	// But PostComment typically doesn't REQUIRE admin token unless configured.
	// However, if we sent it before, we should send it now.
	// If MasterKey is empty, we just do normal request.
	// Move apiURL definition up
	serverURL := strings.TrimSuffix(p.ServerURLs, "/")
	apiURL := fmt.Sprintf("%s/api/comment", serverURL)

	if p.MasterKey != "" {
		resp, err := p.executeAuthRequest(ctx, "POST", apiURL, jsonData)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("Waline API error status: %d", resp.StatusCode)
		}

		var result walineResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil
		}
		if result.Errno != 0 {
			return fmt.Errorf("Waline API error: %s", toString(result.Errmsg))
		}
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Waline API error status: %d", resp.StatusCode)
	}

	var result walineResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}
	if result.Errno != 0 {
		return fmt.Errorf("Waline API error: %s", result.Errmsg)
	}

	return nil
}

func (p *WalineProvider) DeleteComment(ctx context.Context, commentID string) error {
	if p.MasterKey == "" {
		return fmt.Errorf("Waline delete requires MasterKey (Token)")
	}

	serverURL := strings.TrimSuffix(p.ServerURLs, "/")
	apiURL := fmt.Sprintf("%s/api/comment/%s", serverURL, commentID)

	fmt.Printf("[Waline] DeleteComment Request: %s\n", apiURL)

	resp, err := p.executeAuthRequest(ctx, "DELETE", apiURL, nil)
	if err != nil {
		// Log error but return friendly message
		fmt.Printf("[Waline] DeleteComment Request Failed or Auth Failed: %v\n", err)
		if strings.Contains(err.Error(), "Waline Logged Auth Failed (401)") {
			return fmt.Errorf("Waline 删除失败 (401): Master Key 无效。已尝试多种验证格式均失败。请检查 Token 设置。")
		}
		return err
	}
	defer resp.Body.Close()

	// executeAuthRequest ensures we don't get 401 here if it succeeded eventually.
	// We still check for other errors.

	if resp.StatusCode == http.StatusUnauthorized {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("[Waline] DeleteComment 401 Unauthorized. Body: %s\n", string(body))
		return fmt.Errorf("Waline 删除失败 (401): Master Key 无效或权限不足。请检查是否填写了正确的 Token。API返回: %s", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("[Waline] DeleteComment Error Status: %d. Body: %s\n", resp.StatusCode, string(body))
		return fmt.Errorf("Waline API error: %d %s", resp.StatusCode, string(body))
	}

	var result walineResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("[Waline] DeleteComment JSON Decode Error: %v\n", err)
		// If decode fails, but status 200, maybe it's fine or empty body
		return nil
	}
	if result.Errno != 0 {
		fmt.Printf("[Waline] DeleteComment API Error: %s\n", toString(result.Errmsg))
		return fmt.Errorf("Waline API error: %s", toString(result.Errmsg))
	}

	return nil
}

// executeAuthRequest calls the API with multiple Authorization header formats until one succeeds (non-401) or all fail.
func (p *WalineProvider) executeAuthRequest(ctx context.Context, method, apiURL string, body []byte) (*http.Response, error) {
	// Standardize key
	key := strings.TrimSpace(p.MasterKey)
	if key == "" {
		return nil, fmt.Errorf("master key is empty")
	}

	// Prepare candidates.
	// 1. Try "Bearer " + key (Most common for Admin, even Static Tokens often accept this in modern Waline)
	// 2. Try raw key (Older Waline or specific config)
	// 3. Try "Token " + key (Alternative standard)
	candidates := []string{}

	if strings.HasPrefix(key, "Bearer ") {
		candidates = append(candidates, key)
	} else {
		// Priority 1: Bearer (Standard)
		candidates = append(candidates, "Bearer "+key)
		// Priority 2: Raw (If Bearer fails)
		candidates = append(candidates, key)
		// Priority 3: Token prefix
		candidates = append(candidates, "Token "+key)
	}

	var lastResp *http.Response
	var lastErr error

	for _, token := range candidates {
		var reqBody io.Reader
		if body != nil {
			reqBody = bytes.NewBuffer(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, apiURL, reqBody)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)

		fmt.Printf("[Waline] Trying Auth: %s... (truncated)\n", token[:min(len(token), 15)])

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			lastErr = err
			// Network error, return immediate failure as auth won't fix it
			return nil, err
		}

		if resp.StatusCode != http.StatusUnauthorized {
			// Success or other error (500, 403, 200), but Auth passed!
			return resp, nil
		}

		// If 401, close body and try next
		resp.Body.Close()
		lastResp = resp
	}

	// If we are here, all attempts failed with 401 (or candidates empty).
	// Return the last 401 response if exists, to allow caller to read body or status.
	if lastResp != nil {
		// We need to re-issue the request one last time to get a readable body?
		// Or just return a specific error.
		// The caller expects *http.Response.
		// We can't return the closed response.
		// Let's re-request with the FIRST candidate to mock the "fail" state properly
		// or just return an error saying "Auth failed after retries".
		return nil, fmt.Errorf("Waline Logged Auth Failed (401) after multiple attempts")
	}

	return nil, lastErr
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (p *WalineProvider) getGravatar(email string) string {
	if email == "" {
		return ""
	}
	email = strings.TrimSpace(strings.ToLower(email))
	hash := md5.Sum([]byte(email))
	return fmt.Sprintf("https://cravatar.cn/avatar/%s?d=mp&v=1.4.14", hex.EncodeToString(hash[:]))
}
