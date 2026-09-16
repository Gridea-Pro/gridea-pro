package engine

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/render"
	"gridea-pro/backend/internal/template"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoThemesDir 定位仓库里随安装包分发的内置主题源目录。
func repoThemesDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", "..", "frontend", "public", "default-files", "themes"))
	if err != nil {
		t.Fatalf("解析主题目录失败: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("内置主题目录不存在: %v", err)
	}
	return dir
}

// stageTheme 把一个内置主题复制到临时 appDir 下，返回 appDir。
// 复制而不是直接用仓库目录，是为了能在副本上删模板来构造「主题缺模板」的场景。
func stageTheme(t *testing.T, themeName string) string {
	t.Helper()
	appDir := t.TempDir()
	src := filepath.Join(repoThemesDir(t), themeName)
	dst := filepath.Join(appDir, DirThemes, themeName)
	if err := os.CopyFS(dst, os.DirFS(src)); err != nil {
		t.Fatalf("复制主题 %s 失败: %v", themeName, err)
	}
	return appDir
}

func newRealPageRenderer(t *testing.T, appDir, themeName, buildDir string) *PageRenderer {
	t.Helper()
	renderer, err := render.NewRendererFactory(appDir, themeName).CreateRenderer()
	if err != nil {
		t.Fatalf("创建 %s 的渲染器失败: %v", themeName, err)
	}
	return &PageRenderer{
		appDir:   appDir,
		logger:   slog.Default(),
		manifest: NewRenderManifest(buildDir),
		renderer: renderer,
	}
}

func categoryTestData(themeName string) *template.TemplateData {
	post := template.PostView{
		Title:      "分类测试文章",
		FileName:   "cat-post",
		Link:       "/post/cat-post/",
		DateFormat: "2026-09-16",
		Published:  true,
		Stats:      template.PostStats{Text: "1 min read"},
		Categories: []template.CategoryView{
			{Name: "技术笔记", Slug: "tech", Link: "/category/tech/"},
		},
	}
	return &template.TemplateData{
		ThemeConfig: template.ThemeConfigView{
			SiteName:     "Test Site",
			PostPageSize: 10,
			ThemeName:    themeName,
		},
		// 内置主题的 head 片段会读 site.customConfig.*，缺了它 EJS 会直接抛错
		Site: template.SiteView{
			CustomConfig: map[string]interface{}{},
			Utils:        template.NewSiteUtils(),
		},
		Posts: []template.PostView{post},
		Categories: []template.CategoryView{
			{Name: "技术笔记", Slug: "tech", Link: "/category/tech/", Count: 1},
		},
	}
}

// builtinThemesWithCategoryTemplates 是随包分发、应当自带分类模板的主题。
// claudo-theme 不在列：它只有 links.html，连 index/post/tag 都没有。
var builtinThemesWithCategoryTemplates = []string{
	"simple", "notes", "fly", "amore",
	"amore-jinja2", "inotes", "letters-theme", "flavor-theme",
}

// TestBuiltinThemesRenderCategoryPages 用真实模板渲染分类落地页和分类总览页，
// 断言分类名真的出现在产物里——模板缺失/变量名写错都会在这里被抓住。
func TestBuiltinThemesRenderCategoryPages(t *testing.T) {
	for _, themeName := range builtinThemesWithCategoryTemplates {
		t.Run(themeName, func(t *testing.T) {
			appDir := stageTheme(t, themeName)
			buildDir := filepath.Join(appDir, DirOutput)
			r := newRealPageRenderer(t, appDir, themeName, buildDir)
			data := categoryTestData(themeName)

			if err := r.RenderCategoryPages(context.Background(), buildDir, data); err != nil {
				t.Fatalf("RenderCategoryPages 失败: %v", err)
			}
			landing := filepath.Join(buildDir, DefaultCategoryPath, "tech", FileIndexHTML)
			html, err := os.ReadFile(landing)
			if err != nil {
				t.Fatalf("分类落地页未生成: %v", err)
			}
			if !strings.Contains(string(html), "技术笔记") {
				t.Errorf("分类落地页没有输出分类名，产物:\n%s", truncate(string(html)))
			}

			if err := r.RenderCategories(context.Background(), buildDir, data); err != nil {
				t.Fatalf("RenderCategories 失败: %v", err)
			}
			overview := filepath.Join(buildDir, DefaultCategoriesPath, FileIndexHTML)
			overviewHTML, err := os.ReadFile(overview)
			if err != nil {
				t.Fatalf("分类总览页未生成: %v", err)
			}
			if !strings.Contains(string(overviewHTML), "技术笔记") {
				t.Errorf("分类总览页没有列出分类，产物:\n%s", truncate(string(overviewHTML)))
			}
			if !strings.Contains(string(overviewHTML), "/category/tech/") {
				t.Errorf("分类总览页没有链接到分类落地页，产物:\n%s", truncate(string(overviewHTML)))
			}

			// 主题自带分类模板时不应该触发回退警告
			if w := r.TakeWarnings(); len(w) > 0 {
				t.Errorf("内置主题不应触发回退，却产生了警告: %v", w)
			}
		})
	}
}

// TestCategoryPagesFallBackToTagTemplate 覆盖第三方主题没有 category 模板的情况：
// 必须回退到 tag 模板并产出页面，而不是静默地什么都不生成。
func TestCategoryPagesFallBackToTagTemplate(t *testing.T) {
	const themeName = "flavor-theme"
	appDir := stageTheme(t, themeName)
	tmplDir := filepath.Join(appDir, DirThemes, themeName, DirTemplates)
	if err := os.Remove(filepath.Join(tmplDir, "category.html")); err != nil {
		t.Fatalf("构造缺模板场景失败: %v", err)
	}
	if err := os.Remove(filepath.Join(tmplDir, "categories.html")); err != nil {
		t.Fatalf("构造缺模板场景失败: %v", err)
	}

	buildDir := filepath.Join(appDir, DirOutput)
	r := newRealPageRenderer(t, appDir, themeName, buildDir)
	data := categoryTestData(themeName)

	if err := r.RenderCategoryPages(context.Background(), buildDir, data); err != nil {
		t.Fatalf("RenderCategoryPages 失败: %v", err)
	}
	html, err := os.ReadFile(filepath.Join(buildDir, DefaultCategoryPath, "tech", FileIndexHTML))
	if err != nil {
		t.Fatalf("回退后分类落地页仍未生成: %v", err)
	}
	// tag 模板读的是 tag 变量，回退时必须把分类填进去，否则标题会是空的
	if !strings.Contains(string(html), "技术笔记") {
		t.Errorf("回退渲染没有输出分类名，产物:\n%s", truncate(string(html)))
	}

	if err := r.RenderCategories(context.Background(), buildDir, data); err != nil {
		t.Fatalf("RenderCategories 失败: %v", err)
	}
	overview, err := os.ReadFile(filepath.Join(buildDir, DefaultCategoriesPath, FileIndexHTML))
	if err != nil {
		t.Fatalf("回退后分类总览页仍未生成: %v", err)
	}
	if !strings.Contains(string(overview), "技术笔记") {
		t.Errorf("回退的总览页没有列出分类，产物:\n%s", truncate(string(overview)))
	}

	warnings := r.TakeWarnings()
	if len(warnings) != 2 {
		t.Fatalf("回退应产生 2 条用户可见警告（落地页 + 总览页），实际 %d 条: %v", len(warnings), warnings)
	}
	for _, w := range warnings {
		if !strings.Contains(w, "分类") {
			t.Errorf("警告文案应说明是分类模板缺失，实际: %s", w)
		}
	}
}

// TestCategoryPathIsConfigurable 确认分类路径跟随主题配置，
// 而不是写死的 "category"。
func TestCategoryPathIsConfigurable(t *testing.T) {
	const themeName = "flavor-theme"
	appDir := stageTheme(t, themeName)
	buildDir := filepath.Join(appDir, DirOutput)
	r := newRealPageRenderer(t, appDir, themeName, buildDir)

	data := categoryTestData(themeName)
	data.ThemeConfig.CategoryPath = "topics"
	data.ThemeConfig.CategoriesPath = "all-topics"

	if err := r.RenderCategoryPages(context.Background(), buildDir, data); err != nil {
		t.Fatalf("RenderCategoryPages 失败: %v", err)
	}
	if _, err := os.Stat(filepath.Join(buildDir, "topics", "tech", FileIndexHTML)); err != nil {
		t.Errorf("分类落地页没有落在配置的路径下: %v", err)
	}

	if err := r.RenderCategories(context.Background(), buildDir, data); err != nil {
		t.Fatalf("RenderCategories 失败: %v", err)
	}
	if _, err := os.Stat(filepath.Join(buildDir, "all-topics", FileIndexHTML)); err != nil {
		t.Errorf("分类总览页没有落在配置的路径下: %v", err)
	}
}

// TestRenderCategoriesSkipsWhenNoCategories 没有分类时不应产出空的总览页，
// 否则 sitemap 里会出现一个没有内容的页面。
func TestRenderCategoriesSkipsWhenNoCategories(t *testing.T) {
	const themeName = "flavor-theme"
	appDir := stageTheme(t, themeName)
	buildDir := filepath.Join(appDir, DirOutput)
	r := newRealPageRenderer(t, appDir, themeName, buildDir)

	data := categoryTestData(themeName)
	data.Categories = nil

	if err := r.RenderCategories(context.Background(), buildDir, data); err != nil {
		t.Fatalf("RenderCategories 失败: %v", err)
	}
	if _, err := os.Stat(filepath.Join(buildDir, DefaultCategoriesPath, FileIndexHTML)); err == nil {
		t.Error("没有分类时不应生成分类总览页")
	}
}

func truncate(s string) string {
	if len(s) > 600 {
		return s[:600] + "..."
	}
	return s
}

// TestSitemapIncludesCategories 确认分类页进了 sitemap——这是分类能被搜索引擎
// 收录的前提，也是它和标签待遇不对等的历史遗留点之一。
func TestSitemapIncludesCategories(t *testing.T) {
	buildDir := t.TempDir()
	g := NewSeoGenerator()
	g.SetManifest(NewRenderManifest(buildDir))

	data := categoryTestData("flavor-theme")
	data.ThemeConfig.Domain = "https://example.com"

	if err := g.RenderSitemap(buildDir, data); err != nil {
		t.Fatalf("RenderSitemap 失败: %v", err)
	}
	xml, err := os.ReadFile(filepath.Join(buildDir, "sitemap.xml"))
	if err != nil {
		t.Fatalf("sitemap 未生成: %v", err)
	}
	for _, want := range []string{
		"https://example.com/categories/",
		"https://example.com/category/tech/",
	} {
		if !strings.Contains(string(xml), want) {
			t.Errorf("sitemap 缺少 %s，产物:\n%s", want, truncate(string(xml)))
		}
	}
}

// TestSitemapOmitsCategoriesWhenNone 没有分类时不能把总览页写进 sitemap，
// 因为那个页面根本不会被渲染出来。
func TestSitemapOmitsCategoriesWhenNone(t *testing.T) {
	buildDir := t.TempDir()
	g := NewSeoGenerator()
	g.SetManifest(NewRenderManifest(buildDir))

	data := categoryTestData("flavor-theme")
	data.ThemeConfig.Domain = "https://example.com"
	data.Categories = nil

	if err := g.RenderSitemap(buildDir, data); err != nil {
		t.Fatalf("RenderSitemap 失败: %v", err)
	}
	xml, err := os.ReadFile(filepath.Join(buildDir, "sitemap.xml"))
	if err != nil {
		t.Fatalf("sitemap 未生成: %v", err)
	}
	if strings.Contains(string(xml), "/categories/") {
		t.Error("没有分类时 sitemap 不应包含分类总览页")
	}
}

// TestRenderPostUsesConfiguredCategoryPath 文章详情页里的分类链接必须跟随
// 配置的分类路径。RenderPost 会自己拼一个精简的 ThemeConfig 传给 ConvertPost，
// 漏传 CategoryPath 的话文章页里的分类链接会指向一个不存在的路径。
func TestRenderPostUsesConfiguredCategoryPath(t *testing.T) {
	const themeName = "flavor-theme"
	appDir := stageTheme(t, themeName)
	buildDir := filepath.Join(appDir, DirOutput)
	r := newRealPageRenderer(t, appDir, themeName, buildDir)
	r.dataBuilder = &TemplateDataBuilder{logger: slog.Default()}

	data := categoryTestData(themeName)
	data.ThemeConfig.CategoryPath = "topics"

	post := domain.Post{
		FileName:   "cat-post",
		Title:      "分类测试文章",
		Published:  true,
		Categories: []string{"技术笔记"},
	}
	if err := r.RenderPost(context.Background(), buildDir, post, data); err != nil {
		t.Fatalf("RenderPost 失败: %v", err)
	}

	html, err := os.ReadFile(filepath.Join(buildDir, DefaultPostPath, "cat-post", FileIndexHTML))
	if err != nil {
		t.Fatalf("文章页未生成: %v", err)
	}
	if !strings.Contains(string(html), "/topics/") {
		t.Errorf("文章页里的分类链接没有跟随配置路径，产物:\n%s", truncate(string(html)))
	}
	if strings.Contains(string(html), "/category/") {
		t.Errorf("文章页里仍有写死的默认分类路径，产物:\n%s", truncate(string(html)))
	}
}

// TestFlavorCategoryCoverRendering 覆盖 flavor 主题上封面的两种状态：
// 配了封面走沉浸大图，没配则退成同版式的纯色头——降级不能是空白或塌掉的页面。
func TestFlavorCategoryCoverRendering(t *testing.T) {
	const themeName = "flavor-theme"
	const cover = "/post-images/cover-demo.webp"

	t.Run("落地页-有封面", func(t *testing.T) {
		appDir := stageTheme(t, themeName)
		buildDir := filepath.Join(appDir, DirOutput)
		r := newRealPageRenderer(t, appDir, themeName, buildDir)

		data := categoryTestData(themeName)
		data.Posts[0].Categories[0].Cover = cover

		if err := r.RenderCategoryPages(context.Background(), buildDir, data); err != nil {
			t.Fatalf("RenderCategoryPages 失败: %v", err)
		}
		html := readRendered(t, filepath.Join(buildDir, DefaultCategoryPath, "tech", FileIndexHTML))
		if !strings.Contains(html, cover) {
			t.Errorf("落地页没有输出封面图，产物:\n%s", truncate(html))
		}
		if !strings.Contains(html, "cat-hero__img") {
			t.Errorf("落地页没有走沉浸封面形态，产物:\n%s", truncate(html))
		}
		if strings.Contains(html, "cat-hero--bare") {
			t.Errorf("有封面时不该出现无封面的降级形态，产物:\n%s", truncate(html))
		}
	})

	t.Run("落地页-无封面降级", func(t *testing.T) {
		appDir := stageTheme(t, themeName)
		buildDir := filepath.Join(appDir, DirOutput)
		r := newRealPageRenderer(t, appDir, themeName, buildDir)

		data := categoryTestData(themeName) // 不设 Cover

		if err := r.RenderCategoryPages(context.Background(), buildDir, data); err != nil {
			t.Fatalf("RenderCategoryPages 失败: %v", err)
		}
		html := readRendered(t, filepath.Join(buildDir, DefaultCategoryPath, "tech", FileIndexHTML))
		if !strings.Contains(html, "cat-hero--bare") {
			t.Errorf("无封面时没有走降级形态，产物:\n%s", truncate(html))
		}
		if strings.Contains(html, "cat-hero__img") {
			t.Errorf("无封面时不该渲染 img 标签，产物:\n%s", truncate(html))
		}
		// 降级不等于丢信息：分类名仍然要在
		if !strings.Contains(html, "技术笔记") {
			t.Errorf("降级形态丢了分类名，产物:\n%s", truncate(html))
		}
	})

	t.Run("总览页-混合状态", func(t *testing.T) {
		appDir := stageTheme(t, themeName)
		buildDir := filepath.Join(appDir, DirOutput)
		r := newRealPageRenderer(t, appDir, themeName, buildDir)

		data := categoryTestData(themeName)
		data.Categories = []template.CategoryView{
			{Name: "技术笔记", Slug: "tech", Link: "/category/tech/", Count: 3, Cover: cover},
			{Name: "深度思考", Slug: "deep", Link: "/category/deep/", Count: 1},
		}

		if err := r.RenderCategories(context.Background(), buildDir, data); err != nil {
			t.Fatalf("RenderCategories 失败: %v", err)
		}
		html := readRendered(t, filepath.Join(buildDir, DefaultCategoriesPath, FileIndexHTML))
		if !strings.Contains(html, cover) {
			t.Errorf("总览页没有输出封面图，产物:\n%s", truncate(html))
		}
		if !strings.Contains(html, "cat-grid__fallback") {
			t.Errorf("没封面的分类应该走首字兜底，产物:\n%s", truncate(html))
		}
		// 首字必须按字符取，不能把中文截成半个字节
		if !strings.Contains(html, ">深<") {
			t.Errorf("首字兜底没有正确取到中文首字，产物:\n%s", truncate(html))
		}
	})
}

// TestCategoryCoverFlowsFromRepository 确认封面从分类数据一路流到模板视图，
// 中间任何一层漏传都会让主题拿不到图。
func TestCategoryCoverFlowsFromRepository(t *testing.T) {
	b := newTestBuilder()
	categoryByID := map[string]domain.Category{
		"cat-1": {ID: "cat-1", Name: "技术笔记", Slug: "tech", Cover: "/post-images/a.webp"},
	}
	post := domain.Post{
		FileName:    "hello",
		Categories:  []string{"技术笔记"},
		CategoryIDs: []string{"cat-1"},
	}
	view := b.convertPost(post, domain.ThemeConfig{}, categoryByID, nil, nil, nil)
	if len(view.Categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(view.Categories))
	}
	if view.Categories[0].Cover != "/post-images/a.webp" {
		t.Errorf("文章视图里的分类丢了封面，got %q", view.Categories[0].Cover)
	}
}

func readRendered(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("页面未生成: %v", err)
	}
	return string(data)
}

// fakeCategoryRepo 只实现 List，其余方法不会被 Build() 用到。
type fakeCategoryRepo struct{ items []domain.Category }

func (f *fakeCategoryRepo) List(context.Context) ([]domain.Category, error) { return f.items, nil }
func (f *fakeCategoryRepo) Create(context.Context, *domain.Category) error  { return nil }
func (f *fakeCategoryRepo) Update(context.Context, string, *domain.Category) error {
	return nil
}
func (f *fakeCategoryRepo) Delete(context.Context, string) error { return nil }
func (f *fakeCategoryRepo) GetByID(context.Context, string) (*domain.Category, error) {
	return nil, nil
}
func (f *fakeCategoryRepo) GetBySlug(context.Context, string) (*domain.Category, error) {
	return nil, nil
}
func (f *fakeCategoryRepo) SaveAll(context.Context, []domain.Category) error { return nil }

// TestBuildAllCategories 覆盖全站分类列表的构建：仓库顺序、计数口径、孤儿兜底。
// 这段逻辑是这次改动里分支最多的一处——顺序依赖仓库、计数要排除草稿和
// 「不在列表显示」的文章、仓库里已删除但文章仍引用的分类要兜底列出。
func TestBuildAllCategories(t *testing.T) {
	b := NewTemplateDataBuilder(nil, nil, nil, NewThemeConfigService(t.TempDir()))
	b.SetCategoryRepo(&fakeCategoryRepo{items: []domain.Category{
		// 刻意让仓库顺序与字母序相反，才能验证「保持用户排列」而不是被重排
		{ID: "c2", Name: "深度思考", Slug: "deep", Description: "慢下来", Cover: "/post-images/d.webp"},
		{ID: "c1", Name: "技术笔记", Slug: "tech", Description: "技术"},
		{ID: "c3", Name: "空分类", Slug: "empty"},
	}})

	posts := []domain.Post{
		{FileName: "p1", Title: "已发布", Published: true, CategoryIDs: []string{"c1"}, Categories: []string{"技术笔记"}},
		{FileName: "p2", Title: "也已发布", Published: true, CategoryIDs: []string{"c2"}, Categories: []string{"深度思考"}},
		{FileName: "p3", Title: "隐藏", Published: true, HideInList: true, CategoryIDs: []string{"c1"}, Categories: []string{"技术笔记"}},
		{FileName: "p4", Title: "草稿", Published: false, CategoryIDs: []string{"c1"}, Categories: []string{"技术笔记"}},
		// 引用了一个仓库里已不存在的分类：不能因此丢页面
		{FileName: "p5", Title: "孤儿", Published: true, CategoryIDs: []string{"gone"}},
	}

	data, err := b.Build(context.Background(), posts, domain.ThemeConfig{ThemeName: "x"})
	if err != nil {
		t.Fatalf("Build 失败: %v", err)
	}

	// 顺序按仓库；没有可见文章的分类（空分类）不列出；孤儿排在仓库分类之后
	var gotNames []string
	for _, c := range data.Categories {
		gotNames = append(gotNames, c.Name)
	}
	if len(gotNames) != 3 {
		t.Fatalf("期望 3 个分类（2 个仓库分类 + 1 个孤儿），实际 %d: %v", len(gotNames), gotNames)
	}
	if gotNames[0] != "深度思考" || gotNames[1] != "技术笔记" {
		t.Errorf("没有保持仓库顺序，实际: %v", gotNames)
	}
	for _, name := range gotNames {
		if name == "空分类" {
			t.Error("没有可见文章的分类不应出现在总览页")
		}
	}

	byName := map[string]template.CategoryView{}
	for _, c := range data.Categories {
		byName[c.Name] = c
	}
	// 计数必须与分类页实际列出的一致：隐藏文章和草稿都不算
	if got := byName["技术笔记"].Count; got != 1 {
		t.Errorf("计数应排除草稿与 HideInList，期望 1，实际 %d", got)
	}
	if got := byName["深度思考"].Description; got != "慢下来" {
		t.Errorf("描述没有传到视图，实际 %q", got)
	}
	if got := byName["深度思考"].Cover; got != "/post-images/d.webp" {
		t.Errorf("封面没有传到视图，实际 %q", got)
	}
	if got := byName["技术笔记"].Link; got != "/category/tech/" {
		t.Errorf("链接不对，实际 %q", got)
	}
}

// TestCategoryDescriptionReachesTemplate 分类描述必须真的出现在最终 HTML 里。
// 刻意做成端到端断言：视图结构体少字段、或模板引用了不存在的键，pongo2 都只会
// 静默输出空字符串，只有比对产物才抓得住。
func TestCategoryDescriptionReachesTemplate(t *testing.T) {
	const themeName = "flavor-theme"
	const desc = "只说透一件事"

	appDir := stageTheme(t, themeName)
	buildDir := filepath.Join(appDir, DirOutput)
	r := newRealPageRenderer(t, appDir, themeName, buildDir)

	data := categoryTestData(themeName)
	data.Posts[0].Categories[0].Description = desc
	data.Categories[0].Description = desc

	if err := r.RenderCategoryPages(context.Background(), buildDir, data); err != nil {
		t.Fatalf("RenderCategoryPages 失败: %v", err)
	}
	landing := readRendered(t, filepath.Join(buildDir, DefaultCategoryPath, "tech", FileIndexHTML))
	if !strings.Contains(landing, desc) {
		t.Errorf("落地页没有输出分类描述，产物:\n%s", truncate(landing))
	}

	if err := r.RenderCategories(context.Background(), buildDir, data); err != nil {
		t.Fatalf("RenderCategories 失败: %v", err)
	}
	overview := readRendered(t, filepath.Join(buildDir, DefaultCategoriesPath, FileIndexHTML))
	if !strings.Contains(overview, desc) {
		t.Errorf("总览页没有输出分类描述，产物:\n%s", truncate(overview))
	}
}

// TestHasTemplateCoversJinja2Extensions 模板探测的扩展名必须覆盖 Jinja2 惯用后缀，
// 否则用 .jinja2/.j2 命名的主题会被误判成「缺模板」，回退网整体失效。
func TestHasTemplateCoversJinja2Extensions(t *testing.T) {
	appDir := t.TempDir()
	tmplDir := filepath.Join(appDir, DirThemes, "custom", DirTemplates)
	if err := os.MkdirAll(tmplDir, 0755); err != nil {
		t.Fatalf("准备目录失败: %v", err)
	}
	r := &PageRenderer{appDir: appDir, logger: slog.Default()}

	for _, ext := range []string{".html", ".ejs", ".gohtml", ".jinja2", ".j2"} {
		name := "tmpl" + strings.TrimPrefix(ext, ".")
		if err := os.WriteFile(filepath.Join(tmplDir, name+ext), []byte("x"), 0644); err != nil {
			t.Fatalf("写模板失败: %v", err)
		}
		if !r.hasTemplate("custom", name) {
			t.Errorf("扩展名 %s 的模板没有被识别", ext)
		}
	}
	if r.hasTemplate("custom", "nonexistent") {
		t.Error("不存在的模板不应被识别为存在")
	}
}

// TestPathSegmentRejectsTraversal 路径前缀会被 Join 进 buildDir 再 MkdirAll，
// 含分隔符或 .. 的值必须被挡掉，否则产物会写到 output 之外。
func TestPathSegmentRejectsTraversal(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", DefaultCategoryPath},
		{"topics", "topics"},
		{"../../tmp/pwn", DefaultCategoryPath},
		{"a/b", DefaultCategoryPath},
		{`a\b`, DefaultCategoryPath},
		{"..", DefaultCategoryPath},
	}
	for _, c := range cases {
		if got := categoryPathOf(c.in); got != c.want {
			t.Errorf("categoryPathOf(%q) = %q，期望 %q", c.in, got, c.want)
		}
	}
	if got := categoriesPathOf("../evil"); got != DefaultCategoriesPath {
		t.Errorf("categoriesPathOf 没有挡住路径穿越，得到 %q", got)
	}
}
