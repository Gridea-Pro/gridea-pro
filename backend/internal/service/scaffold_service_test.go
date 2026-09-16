package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// newAssetsWithTheme 构造一份最小的 default-files 资源树，内含一个主题。
func newAssetsWithTheme(files map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{}
	for name, content := range files {
		fsys["frontend/public/default-files/"+name] = &fstest.MapFile{Data: []byte(content)}
	}
	return fsys
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 %s 失败: %v", path, err)
	}
	return string(data)
}

// TestInitSite_TopsUpMissingThemeFiles 覆盖新版本给已安装主题新增模板的场景
// （比如这次新增的分类模板）：缺的要补上，用户改过的文件一个字都不能动。
func TestInitSite_TopsUpMissingThemeFiles(t *testing.T) {
	appDir := t.TempDir()

	// 第一次初始化：主题只有 tag 模板
	v1 := newAssetsWithTheme(map[string]string{
		"themes/simple/templates/tag.ejs": "v1 tag",
	})
	if err := NewScaffoldService(v1).InitSite(appDir); err != nil {
		t.Fatalf("首次 InitSite 失败: %v", err)
	}
	tagPath := filepath.Join(appDir, "themes", "simple", "templates", "tag.ejs")
	if readFile(t, tagPath) != "v1 tag" {
		t.Fatal("首次初始化没有复制主题模板")
	}

	// 用户改了自己的模板
	if err := os.WriteFile(tagPath, []byte("用户改过的内容"), 0644); err != nil {
		t.Fatalf("模拟用户修改失败: %v", err)
	}

	// 新版本：tag 模板内容变了，并且新增了 category 模板
	v2 := newAssetsWithTheme(map[string]string{
		"themes/simple/templates/tag.ejs":      "v2 tag",
		"themes/simple/templates/category.ejs": "v2 category",
	})
	if err := NewScaffoldService(v2).InitSite(appDir); err != nil {
		t.Fatalf("升级后 InitSite 失败: %v", err)
	}

	// 新增的模板要补进来，老用户才拿得到分类页
	categoryPath := filepath.Join(appDir, "themes", "simple", "templates", "category.ejs")
	if got := readFile(t, categoryPath); got != "v2 category" {
		t.Errorf("新版本新增的模板没有补齐，得到 %q", got)
	}
	// 用户的改动绝对不能被覆盖
	if got := readFile(t, tagPath); got != "用户改过的内容" {
		t.Errorf("用户修改过的模板被覆盖了，得到 %q", got)
	}
}

// TestInitSite_DeletedThemeStaysDeleted 用户整个删掉的主题不能被复活——
// 补缺逻辑只针对还在磁盘上的主题。
func TestInitSite_DeletedThemeStaysDeleted(t *testing.T) {
	appDir := t.TempDir()

	assets := newAssetsWithTheme(map[string]string{
		"themes/simple/templates/tag.ejs": "tag",
	})
	if err := NewScaffoldService(assets).InitSite(appDir); err != nil {
		t.Fatalf("首次 InitSite 失败: %v", err)
	}

	themeDir := filepath.Join(appDir, "themes", "simple")
	if err := os.RemoveAll(themeDir); err != nil {
		t.Fatalf("模拟用户删除主题失败: %v", err)
	}

	assetsV2 := newAssetsWithTheme(map[string]string{
		"themes/simple/templates/tag.ejs":      "tag",
		"themes/simple/templates/category.ejs": "category",
	})
	if err := NewScaffoldService(assetsV2).InitSite(appDir); err != nil {
		t.Fatalf("第二次 InitSite 失败: %v", err)
	}

	if _, err := os.Stat(themeDir); err == nil {
		t.Error("用户删除的主题被复活了")
	}
}

// TestInitSite_NewThemeStillGetsInstalled 新版本新增的整个主题仍然要装进来。
func TestInitSite_NewThemeStillGetsInstalled(t *testing.T) {
	appDir := t.TempDir()

	v1 := newAssetsWithTheme(map[string]string{
		"themes/simple/templates/tag.ejs": "tag",
	})
	if err := NewScaffoldService(v1).InitSite(appDir); err != nil {
		t.Fatalf("首次 InitSite 失败: %v", err)
	}

	v2 := newAssetsWithTheme(map[string]string{
		"themes/simple/templates/tag.ejs":  "tag",
		"themes/flavor/templates/tag.html": "flavor tag",
	})
	if err := NewScaffoldService(v2).InitSite(appDir); err != nil {
		t.Fatalf("升级后 InitSite 失败: %v", err)
	}

	if _, err := os.Stat(filepath.Join(appDir, "themes", "flavor", "templates", "tag.html")); err != nil {
		t.Errorf("新版本新增的主题没有安装: %v", err)
	}

	// manifest 要记上新主题，否则下次会被当成全新主题反复处理
	data, err := os.ReadFile(filepath.Join(appDir, ".scaffold.json"))
	if err != nil {
		t.Fatalf("读取 manifest 失败: %v", err)
	}
	var m scaffoldManifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("解析 manifest 失败: %v", err)
	}
	found := false
	for _, name := range m.Themes {
		if name == "flavor" {
			found = true
		}
	}
	if !found {
		t.Errorf("manifest 未记录新主题，实际记录: %v", m.Themes)
	}
}
