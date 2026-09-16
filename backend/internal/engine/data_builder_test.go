package engine

import (
	"log/slog"
	"testing"

	"gridea-pro/backend/internal/domain"
)

func newTestBuilder() *TemplateDataBuilder {
	return &TemplateDataBuilder{
		logger: slog.Default(),
	}
}

func TestConvertPost_TagIDsResolveViaTagByID(t *testing.T) {
	b := newTestBuilder()

	// 两个同名但不同 ID / Slug 的标签（数据层唯一性约束下不会发生，
	// 但 ID-first 逻辑必须能正确处理，不受 Name 重复影响）
	tagByID := map[string]domain.Tag{
		"id-a": {ID: "id-a", Name: "Go", Slug: "go-lang"},
		"id-b": {ID: "id-b", Name: "Rust", Slug: "rust"},
	}
	tagByName := map[string]domain.Tag{
		"Go":   tagByID["id-a"],
		"Rust": tagByID["id-b"],
	}

	post := domain.Post{
		FileName: "hello",
		Tags:     []string{"Go", "Rust"},
		TagIDs:   []string{"id-a", "id-b"},
	}
	config := domain.ThemeConfig{TagPath: "tag"}

	view := b.convertPost(post, config, nil, nil, tagByID, tagByName)

	if len(view.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(view.Tags))
	}
	if view.Tags[0].Slug != "go-lang" {
		t.Errorf("expected Slug go-lang from tagByID, got %q", view.Tags[0].Slug)
	}
	if view.Tags[0].Link != "/tag/go-lang/" {
		t.Errorf("expected link /tag/go-lang/, got %q", view.Tags[0].Link)
	}
	if view.Tags[1].Slug != "rust" {
		t.Errorf("expected Slug rust, got %q", view.Tags[1].Slug)
	}
}

func TestConvertPost_LegacyNoTagIDsFallsBackToName(t *testing.T) {
	b := newTestBuilder()

	tagByID := map[string]domain.Tag{
		"id-a": {ID: "id-a", Name: "Go", Slug: "go"},
	}
	tagByName := map[string]domain.Tag{
		"Go": tagByID["id-a"],
	}

	// 老文章：只有 Tags（Name），没有 TagIDs
	post := domain.Post{
		FileName: "old-post",
		Tags:     []string{"Go"},
		TagIDs:   nil,
	}
	config := domain.ThemeConfig{TagPath: "tag"}

	view := b.convertPost(post, config, nil, nil, tagByID, tagByName)

	if len(view.Tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(view.Tags))
	}
	if view.Tags[0].Name != "Go" {
		t.Errorf("expected Name Go, got %q", view.Tags[0].Name)
	}
	if view.Tags[0].Slug != "go" {
		t.Errorf("expected Slug go from tagByName, got %q", view.Tags[0].Slug)
	}
}

func TestConvertPost_TagIDMissingIsSkipped(t *testing.T) {
	b := newTestBuilder()

	tagByID := map[string]domain.Tag{
		"id-a": {ID: "id-a", Name: "Go", Slug: "go"},
		// id-b 被删除后仍被 post 引用
	}
	tagByName := map[string]domain.Tag{
		"Go": tagByID["id-a"],
	}

	post := domain.Post{
		FileName: "hello",
		Tags:     []string{"Go", "DeletedTag"},
		TagIDs:   []string{"id-a", "id-b"},
	}
	config := domain.ThemeConfig{TagPath: "tag"}

	view := b.convertPost(post, config, nil, nil, tagByID, tagByName)

	// id-b 命中失败 → 跳过；不应输出 NanoID 作为假标签
	if len(view.Tags) != 1 {
		t.Fatalf("expected 1 tag (missing id skipped), got %d: %+v", len(view.Tags), view.Tags)
	}
	if view.Tags[0].Slug != "go" {
		t.Errorf("expected Slug go, got %q", view.Tags[0].Slug)
	}
	if view.TagsString != "Go" {
		t.Errorf("expected TagsString=Go, got %q", view.TagsString)
	}
}

func TestConvertPost_DuplicateNameDifferentSlugResolvedByID(t *testing.T) {
	b := newTestBuilder()

	// 模拟数据层"脏数据"场景：同名标签指向不同 Slug
	// 按 Name 反查会覆盖，按 ID 反查各自独立
	tagByID := map[string]domain.Tag{
		"id-a": {ID: "id-a", Name: "Lang", Slug: "lang-a"},
		"id-b": {ID: "id-b", Name: "Lang", Slug: "lang-b"}, // 与 id-a 同名
	}
	// tagByName 只会保留一个（先添加的）
	tagByName := map[string]domain.Tag{
		"Lang": tagByID["id-a"],
	}

	post := domain.Post{
		FileName: "x",
		Tags:     []string{"Lang", "Lang"},
		TagIDs:   []string{"id-a", "id-b"},
	}
	config := domain.ThemeConfig{TagPath: "tag"}

	view := b.convertPost(post, config, nil, nil, tagByID, tagByName)

	// 重名在数据层约束下虽不应发生，但 ID-first 路径要能正确分流
	if len(view.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(view.Tags))
	}
	if view.Tags[0].Slug == view.Tags[1].Slug {
		t.Errorf("ID-first 路径应产生不同的 Slug，得到相同: %+v", view.Tags)
	}
	slugs := map[string]bool{view.Tags[0].Slug: true, view.Tags[1].Slug: true}
	if !slugs["lang-a"] || !slugs["lang-b"] {
		t.Errorf("expected slugs lang-a and lang-b, got %v", slugs)
	}
}

func TestConvertPost_CategoryLinkFollowsConfiguredPath(t *testing.T) {
	b := newTestBuilder()

	categoryByID := map[string]domain.Category{
		"cat-1": {ID: "cat-1", Name: "技术笔记", Slug: "tech"},
	}
	post := domain.Post{
		FileName:    "hello",
		Categories:  []string{"技术笔记"},
		CategoryIDs: []string{"cat-1"},
	}

	// 未配置时回退到默认的 category 前缀
	view := b.convertPost(post, domain.ThemeConfig{}, categoryByID, nil, nil, nil)
	if len(view.Categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(view.Categories))
	}
	if view.Categories[0].Link != "/category/tech/" {
		t.Errorf("expected default link /category/tech/, got %q", view.Categories[0].Link)
	}

	// 配置了自定义前缀时必须跟随配置，否则文章页里的分类链接会指向不存在的页面
	view = b.convertPost(post, domain.ThemeConfig{CategoryPath: "topics"}, categoryByID, nil, nil, nil)
	if view.Categories[0].Link != "/topics/tech/" {
		t.Errorf("expected configured link /topics/tech/, got %q", view.Categories[0].Link)
	}
}
