package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ensureGitHubRepo 确保用户的 GitHub Pages 仓库（username.github.io）存在
// 已存在则跳过（无论内容是否为空），不存在则创建公开仓库并自动初始化 README
func ensureGitHubRepo(client *http.Client, token, username string) (map[string]string, error) {
	repoName := strings.ToLower(username) + ".github.io"
	result := map[string]string{"repository": repoName}

	// 1. 检查仓库是否存在
	checkURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", username, repoName)
	req, _ := http.NewRequest(http.MethodGet, checkURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Gridea-Pro")

	resp, err := client.Do(req)
	if err != nil {
		return result, fmt.Errorf("check repo failed: %w", err)
	}
	resp.Body.Close()

	// 已存在：直接返回，绝不覆盖
	if resp.StatusCode == http.StatusOK {
		return result, nil
	}

	// 其他非 404 错误：不创建，避免误操作
	if resp.StatusCode != http.StatusNotFound {
		return result, fmt.Errorf("unexpected status when checking repo: %d", resp.StatusCode)
	}

	// 2. 404：创建新仓库
	payload := map[string]interface{}{
		"name":        repoName,
		"description": "My blog powered by Gridea Pro",
		"private":     false,
		"auto_init":   true,
	}
	body, _ := json.Marshal(payload)
	createReq, _ := http.NewRequest(http.MethodPost, "https://api.github.com/user/repos", bytes.NewReader(body))
	createReq.Header.Set("Authorization", "Bearer "+token)
	createReq.Header.Set("Accept", "application/vnd.github+json")
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("User-Agent", "Gridea-Pro")

	createResp, err := client.Do(createReq)
	if err != nil {
		return result, fmt.Errorf("create repo failed: %w", err)
	}
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusCreated {
		return result, fmt.Errorf("create repo failed with status %d", createResp.StatusCode)
	}
	return result, nil
}
// ensureNetlifySite 确保用户有一个可用的 Netlify Site
// 优先查找现有 sites 中名称包含 "gridea" 的，找不到则创建一个
// 返回 netlifySiteId 和 domain
func ensureNetlifySite(client *http.Client, token, username string) (map[string]string, error) {
	result := make(map[string]string)

	// 1. 列出现有 sites
	listReq, _ := http.NewRequest(http.MethodGet, "https://api.netlify.com/api/v1/sites", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listReq.Header.Set("Accept", "application/json")

	listResp, err := client.Do(listReq)
	if err != nil {
		return result, fmt.Errorf("list sites failed: %w", err)
	}
	defer listResp.Body.Close()

	if listResp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("list sites failed with status %d", listResp.StatusCode)
	}

	var sites []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		URL    string `json:"url"`
		SSLURL string `json:"ssl_url"`
	}
	listBody, _ := io.ReadAll(listResp.Body)
	if err := json.Unmarshal(listBody, &sites); err != nil {
		return result, fmt.Errorf("parse sites failed: %w", err)
	}

	// 优先使用名称包含 "gridea" 的 site（用户之前可能用 Gridea 创建过）
	for _, s := range sites {
		if strings.Contains(strings.ToLower(s.Name), "gridea") {
			result["netlifySiteId"] = s.ID
			if s.SSLURL != "" {
				result["domain"] = s.SSLURL
			} else {
				result["domain"] = s.URL
			}
			return result, nil
		}
	}

	// 没有匹配的 → 创建新 site（Netlify 会自动生成随机名称）
	payload := map[string]interface{}{}
	body, _ := json.Marshal(payload)
	createReq, _ := http.NewRequest(http.MethodPost, "https://api.netlify.com/api/v1/sites", bytes.NewReader(body))
	createReq.Header.Set("Authorization", "Bearer "+token)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Accept", "application/json")

	createResp, err := client.Do(createReq)
	if err != nil {
		return result, fmt.Errorf("create site failed: %w", err)
	}
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusCreated && createResp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("create site failed with status %d", createResp.StatusCode)
	}

	var newSite struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		URL    string `json:"url"`
		SSLURL string `json:"ssl_url"`
	}
	createBody, _ := io.ReadAll(createResp.Body)
	if err := json.Unmarshal(createBody, &newSite); err != nil {
		return result, fmt.Errorf("parse new site failed: %w", err)
	}

	result["netlifySiteId"] = newSite.ID
	if newSite.SSLURL != "" {
		result["domain"] = newSite.SSLURL
	} else {
		result["domain"] = newSite.URL
	}
	return result, nil
}
