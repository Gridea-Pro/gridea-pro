package engine

import (
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/template"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestPageRenderer 构造一个仅用于兜底测试的 PageRenderer：
// 不依赖 renderer 与 dataBuilder，仅需 manifest 能写文件。
func newTestPageRenderer(t *testing.T, buildDir string) *PageRenderer {
	t.Helper()
	return &PageRenderer{
		logger:   slog.Default(),
		manifest: NewRenderManifest(buildDir),
	}
}

func mkVisiblePost(title string) template.PostView {
	return template.PostView{
		Title:      title,
		FileName:   strings.ReplaceAll(strings.ToLower(title), " ", "-"),
		Link:       "/post/" + strings.ReplaceAll(strings.ToLower(title), " ", "-") + "/",
		DateFormat: "2026-04-22",
		Published:  true,
	}
}

func TestRenderSimpleIndex_Pagination(t *testing.T) {
	buildDir := t.TempDir()
	r := newTestPageRenderer(t, buildDir)

	// 25 篇文章，每页 10 篇 → 3 页
	posts := make([]template.PostView, 25)
	for i := range posts {
		posts[i] = mkVisiblePost(fmt.Sprintf("Post %d", i+1))
	}

	data := &template.TemplateData{
		ThemeConfig: template.ThemeConfigView{
			SiteName:         "Test",
			PostPageSize:     10,
			ThemeName:        "test",
		},
		Posts: posts,
	}

	if err := r.renderSimpleIndex(buildDir, data); err != nil {
		t.Fatalf("renderSimpleIndex failed: %v", err)
	}

	// 第 1 页落在 buildDir/index.html
	page1 := filepath.Join(buildDir, "index.html")
	if _, err := os.Stat(page1); err != nil {
		t.Fatalf("page 1 missing: %v", err)
	}
	// 第 2 页落在 buildDir/page/2/index.html
	page2 := filepath.Join(buildDir, "page", "2", "index.html")
	if _, err := os.Stat(page2); err != nil {
		t.Fatalf("page 2 missing: %v", err)
	}
	// 第 3 页落在 buildDir/page/3/index.html
	page3 := filepath.Join(buildDir, "page", "3", "index.html")
	if _, err := os.Stat(page3); err != nil {
		t.Fatalf("page 3 missing: %v", err)
	}

	// 降级 banner 应出现在每一页
	for _, p := range []string{page1, page2, page3} {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		content := string(data)
		if !strings.Contains(content, "降级视图") {
			t.Errorf("%s missing fallback banner", filepath.Base(filepath.Dir(p)))
		}
	}

	// 第 1 页只含前 10 篇标题（<a>Post N</a> 形式）
	p1Content, _ := os.ReadFile(page1)
	if !strings.Contains(string(p1Content), ">Post 1<") {
		t.Error("page 1 should contain Post 1")
	}
	if !strings.Contains(string(p1Content), ">Post 10<") {
		t.Error("page 1 should contain Post 10")
	}
	if strings.Contains(string(p1Content), ">Post 11<") {
		t.Error("page 1 should NOT contain Post 11")
	}
	// 分页导航
	if !strings.Contains(string(p1Content), "/page/2/") {
		t.Error("page 1 should link to /page/2/")
	}

	// 第 3 页含最后 5 篇
	p3Content, _ := os.ReadFile(page3)
	if !strings.Contains(string(p3Content), ">Post 21<") {
		t.Error("page 3 should contain Post 21")
	}
	if !strings.Contains(string(p3Content), ">Post 25<") {
		t.Error("page 3 should contain Post 25")
	}
	// 第 3 页无"下一页"
	if strings.Contains(string(p3Content), "/page/4/") {
		t.Error("page 3 should NOT link to /page/4/")
	}
}

func TestRenderSimpleIndex_HidesUnpublished(t *testing.T) {
	buildDir := t.TempDir()
	r := newTestPageRenderer(t, buildDir)

	posts := []template.PostView{
		{Title: "Visible", FileName: "visible", Link: "/post/visible/", Published: true},
		{Title: "Draft", FileName: "draft", Link: "/post/draft/", Published: false},
		{Title: "Hidden", FileName: "hidden", Link: "/post/hidden/", Published: true, HideInList: true},
	}
	data := &template.TemplateData{
		ThemeConfig: template.ThemeConfigView{SiteName: "Test", PostPageSize: 10},
		Posts:       posts,
	}

	if err := r.renderSimpleIndex(buildDir, data); err != nil {
		t.Fatalf("renderSimpleIndex failed: %v", err)
	}

	b, _ := os.ReadFile(filepath.Join(buildDir, "index.html"))
	content := string(b)
	if !strings.Contains(content, "Visible") {
		t.Error("visible post missing")
	}
	if strings.Contains(content, "Draft") {
		t.Error("draft post should not appear")
	}
	if strings.Contains(content, "Hidden") {
		t.Error("hide-in-list post should not appear on home page")
	}
}

func TestRenderSimpleIndex_EmptyPostsStillWritesIndex(t *testing.T) {
	buildDir := t.TempDir()
	r := newTestPageRenderer(t, buildDir)

	data := &template.TemplateData{
		ThemeConfig: template.ThemeConfigView{SiteName: "Test", PostPageSize: 10},
		Posts:       nil,
	}
	if err := r.renderSimpleIndex(buildDir, data); err != nil {
		t.Fatalf("renderSimpleIndex failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(buildDir, "index.html")); err != nil {
		t.Fatalf("even with no posts, index.html must exist: %v", err)
	}
}

func TestRenderSimplePost_SpecialPage(t *testing.T) {
	buildDir := t.TempDir()
	r := newTestPageRenderer(t, buildDir)

	data := &template.TemplateData{
		ThemeConfig: template.ThemeConfigView{SiteName: "Test"},
		SiteTitle:   "About | Test",
		Post: template.PostView{
			Title:      "About Me",
			Content:    "<p>Hello</p>",
			DateFormat: "2026-04-22",
		},
	}

	postDir := filepath.Join(buildDir, "about")
	_ = os.MkdirAll(postDir, 0o755)

	if err := r.renderSimplePost(postDir, data, true); err != nil {
		t.Fatalf("renderSimplePost failed: %v", err)
	}

	b, _ := os.ReadFile(filepath.Join(postDir, "index.html"))
	content := string(b)

	// 特殊页：不显示"返回首页"链接与日期元数据
	if strings.Contains(content, "返回首页") {
		t.Error("special page should not show back-to-home link")
	}
	if strings.Contains(content, "2026-04-22") {
		t.Error("special page should not show post-meta date")
	}
	if !strings.Contains(content, "About Me") {
		t.Error("title missing")
	}
	if !strings.Contains(content, "<p>Hello</p>") {
		t.Error("content missing")
	}
	// 降级 banner 必须保留
	if !strings.Contains(content, "降级视图") {
		t.Error("missing fallback banner")
	}
}

func TestRenderSimplePost_RegularPost(t *testing.T) {
	buildDir := t.TempDir()
	r := newTestPageRenderer(t, buildDir)

	data := &template.TemplateData{
		ThemeConfig: template.ThemeConfigView{SiteName: "Test"},
		SiteTitle:   "Post | Test",
		Post: template.PostView{
			Title:      "My Post",
			Content:    "<p>Body</p>",
			DateFormat: "2026-04-22",
		},
	}

	postDir := filepath.Join(buildDir, "post", "my-post")
	_ = os.MkdirAll(postDir, 0o755)

	if err := r.renderSimplePost(postDir, data, false); err != nil {
		t.Fatalf("renderSimplePost failed: %v", err)
	}

	b, _ := os.ReadFile(filepath.Join(postDir, "index.html"))
	content := string(b)

	if !strings.Contains(content, "返回首页") {
		t.Error("regular post should show back-to-home link")
	}
	if !strings.Contains(content, "2026-04-22") {
		t.Error("regular post should show post-meta date")
	}
	if !strings.Contains(content, "降级视图") {
		t.Error("missing fallback banner")
	}
}

// 让 linter 不抱怨未使用的 domain import；保留以防后续扩展
var _ = domain.Post{}
