package engine

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"

	"gridea-pro/backend/internal/template"
)

// SearchIndexBuilder 负责生成搜索索引数据
type SearchIndexBuilder struct {
	logger   *slog.Logger
	manifest *RenderManifest
}

// NewSearchIndexBuilder 创建 SearchIndexBuilder
func NewSearchIndexBuilder() *SearchIndexBuilder {
	return &SearchIndexBuilder{
		logger: slog.Default(),
	}
}

// SetManifest 设置渲染产物跟踪器
func (b *SearchIndexBuilder) SetManifest(m *RenderManifest) {
	b.manifest = m
}

// RenderSearchJSON 生成搜索数据 /api/search.json
func (b *SearchIndexBuilder) RenderSearchJSON(buildDir string, data *template.TemplateData) error {
	var entries []searchEntry
	for _, post := range data.Posts {
		if post.HideInList {
			continue
		}
		// 将 HTML 内容转为纯文本用于搜索
		plainContent := stripHTMLForSearch(string(post.Content))
		// 限制内容长度，3000 字足以覆盖绝大多数搜索场景
		if len([]rune(plainContent)) > 3000 {
			plainContent = string([]rune(plainContent)[:3000])
		}

		// 收集标签名称
		var tagNames []string
		for _, t := range post.Tags {
			tagNames = append(tagNames, t.Name)
		}

		entries = append(entries, searchEntry{
			Title:   post.Title,
			Link:    post.Link,
			Date:    post.DateFormat,
			Tags:    tagNames,
			Content: plainContent,
		})
	}

	jsonData, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("序列化搜索数据失败: %w", err)
	}

	apiDir := filepath.Join(buildDir, "api")
	if err := os.MkdirAll(apiDir, 0755); err != nil {
		return err
	}

	b.logger.Info(fmt.Sprintf("✅ 搜索数据生成成功 (%d 篇文章)", len(entries)))
	return b.manifest.WriteFile(filepath.Join(apiDir, "search.json"), jsonData, 0644)
}

// stripHTMLForSearch 移除 HTML 标签，返回纯文本（用于搜索索引）。
func stripHTMLForSearch(s string) string {
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return s
	}

	var buf strings.Builder
	var extractText func(*html.Node)
	extractText = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
			buf.WriteByte(' ')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && (c.Data == "script" || c.Data == "style") {
				continue
			}
			extractText(c)
		}
	}
	extractText(doc)

	return strings.Join(strings.Fields(buf.String()), " ")
}
