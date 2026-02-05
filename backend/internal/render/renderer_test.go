package render_test

import (
	"testing"

	"gridea-pro/backend/internal/render"
	"gridea-pro/backend/internal/template"
)

// TestGoTemplateRenderer 测试 Go Templates 渲染器
func TestGoTemplateRenderer(t *testing.T) {
	// 创建渲染器
	config := render.RenderConfig{
		AppDir:    "/path/to/app",
		ThemeName: "notes_go",
		CacheDir:  "/tmp/cache",
	}
	renderer := render.NewGoTemplateRenderer(config)

	// 准备数据
	data := &template.TemplateData{
		ThemeConfig: template.ThemeConfigView{
			SiteName:        "我的博客",
			SiteDescription: "技术分享",
		},
		Posts: []template.PostView{
			{
				Title:      "Go Templates 入门",
				Link:       "/post/go-templates/",
				DateFormat: "2024-01-20",
				Abstract:   "学习 Go Templates 的基础语法...",
				Tags: []template.TagView{
					{Name: "Go", Link: "/tag/go/"},
					{Name: "Templates", Link: "/tag/templates/"},
				},
			},
		},
	}

	// 渲染
	html, err := renderer.Render("index", data)
	if err != nil {
		t.Fatalf("渲染失败: %v", err)
	}

	if html == "" {
		t.Error("渲染结果为空")
	}

	t.Logf("渲染成功,引擎类型: %s", renderer.GetEngineType())
}

// TestEjsRenderer 测试 EJS 渲染器
func TestEjsRenderer(t *testing.T) {
	// 创建渲染器
	config := render.RenderConfig{
		AppDir:    "/path/to/app",
		ThemeName: "notes",
		CacheDir:  "/tmp/cache",
	}
	renderer := render.NewEjsRenderer(config)

	// 准备数据(与 Go Templates 完全相同)
	data := &template.TemplateData{
		ThemeConfig: template.ThemeConfigView{
			SiteName:        "我的博客",
			SiteDescription: "技术分享",
		},
		Posts: []template.PostView{
			{
				Title:      "EJS 入门",
				Link:       "/post/ejs-intro/",
				DateFormat: "2024-01-20",
				Abstract:   "学习 EJS 的基础语法...",
				Tags: []template.TagView{
					{Name: "EJS", Link: "/tag/ejs/"},
					{Name: "JavaScript", Link: "/tag/javascript/"},
				},
			},
		},
	}

	// 渲染
	html, err := renderer.Render("index", data)
	if err != nil {
		t.Fatalf("渲染失败: %v", err)
	}

	if html == "" {
		t.Error("渲染结果为空")
	}

	t.Logf("渲染成功,引擎类型: %s", renderer.GetEngineType())
}

// TestRendererFactory 测试渲染器工厂
func TestRendererFactory(t *testing.T) {
	// 创建工厂
	factory := render.NewRendererFactory("/path/to/app", "notes_go")

	// 检测引擎类型
	engineType, err := factory.GetEngineType()
	if err != nil {
		t.Fatalf("检测引擎类型失败: %v", err)
	}
	t.Logf("检测到引擎类型: %s", engineType)

	// 创建渲染器
	renderer, err := factory.CreateRenderer()
	if err != nil {
		t.Fatalf("创建渲染器失败: %v", err)
	}

	t.Logf("创建渲染器成功,类型: %s", renderer.GetEngineType())
}

// ExampleUsage 使用示例
func ExampleUsage() {
	// 1. 创建工厂
	factory := render.NewRendererFactory("/app/dir", "notes_go")

	// 2. 自动创建渲染器
	renderer, _ := factory.CreateRenderer()

	// 3. 准备数据
	data := &template.TemplateData{
		Posts: []template.PostView{
			{Title: "文章1", Link: "/post/1/"},
			{Title: "文章2", Link: "/post/2/"},
		},
	}

	// 4. 渲染(无需关心底层引擎)
	html, _ := renderer.Render("index", data)
	_ = html
}
