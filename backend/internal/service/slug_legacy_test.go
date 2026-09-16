package service

import (
	"context"
	"strings"
	"testing"

	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/repository"
)

// 历史数据里可能存在不符合当前 slug 规则的值（旧版本写入、或经 MCP 由外部写入）。
// 用户只改描述/封面时不该被这些历史值拦下——下面两个测试钉住这个行为。

func TestSaveCategory_LegacySlugUnchangedIsAllowed(t *testing.T) {
	appDir := t.TempDir()
	repo := repository.NewCategoryRepository(appDir)
	svc := NewCategoryService(repo, nil)
	ctx := context.Background()

	// 绕过 service 直接落一条历史数据：slug 含空格和感叹号，按当前规则非法
	legacy := domain.Category{Name: "历史分类", Slug: "legacy slug!", Description: "旧的"}
	if err := repo.Create(ctx, &legacy); err != nil {
		t.Fatalf("准备历史数据失败: %v", err)
	}

	// 只改描述，slug 原样带回 —— 必须能存下去
	updated := legacy
	updated.Description = "改过的描述"
	if err := svc.SaveCategory(ctx, updated, legacy.ID); err != nil {
		t.Fatalf("未改动 slug 的编辑被拦下了: %v", err)
	}

	got, err := repo.GetByID(ctx, legacy.ID)
	if err != nil || got == nil {
		t.Fatalf("读回分类失败: %v", err)
	}
	if got.Description != "改过的描述" {
		t.Errorf("描述没有被保存，得到 %q", got.Description)
	}
	if got.Slug != "legacy slug!" {
		t.Errorf("slug 不该被改动，得到 %q", got.Slug)
	}
}

func TestSaveCategory_ChangingSlugStillValidated(t *testing.T) {
	appDir := t.TempDir()
	repo := repository.NewCategoryRepository(appDir)
	svc := NewCategoryService(repo, nil)
	ctx := context.Background()

	legacy := domain.Category{Name: "历史分类", Slug: "legacy slug!"}
	if err := repo.Create(ctx, &legacy); err != nil {
		t.Fatalf("准备历史数据失败: %v", err)
	}

	// 一旦动了 slug，就必须按当前规则校验
	updated := legacy
	updated.Slug = "still bad!"
	err := svc.SaveCategory(ctx, updated, legacy.ID)
	if err == nil {
		t.Fatal("改成非法 slug 却被放行了")
	}
	if !strings.Contains(err.Error(), "不合法") {
		t.Errorf("错误信息应说明 slug 不合法，实际: %v", err)
	}

	// 大写是允许的：#99 真正要防的撞目录风险由大小写不敏感的唯一性校验独立堵住
	updated.Slug = "Gridea-Pro"
	if err := svc.SaveCategory(ctx, updated, legacy.ID); err != nil {
		t.Errorf("含大写的 slug 应被接受，却报错: %v", err)
	}
}

func TestSaveCategory_NewCategoryStillValidated(t *testing.T) {
	appDir := t.TempDir()
	svc := NewCategoryService(repository.NewCategoryRepository(appDir), nil)

	err := svc.SaveCategory(context.Background(), domain.Category{Name: "新的", Slug: "bad slug!"}, "")
	if err == nil {
		t.Fatal("新建时非法 slug 应被拒绝")
	}
}

func TestSaveTag_LegacySlugUnchangedIsAllowed(t *testing.T) {
	appDir := t.TempDir()
	repo := repository.NewTagRepository(appDir)
	svc := NewTagService(repo, nil)
	ctx := context.Background()

	legacy := domain.Tag{ID: "t1", Name: "历史标签", Slug: "legacy slug!", Color: "#111111"}
	if err := repo.Create(ctx, &legacy); err != nil {
		t.Fatalf("准备历史数据失败: %v", err)
	}

	// 只改颜色，slug 原样带回
	updated := legacy
	updated.Color = "#222222"
	if err := svc.SaveTag(ctx, updated, ""); err != nil {
		t.Fatalf("未改动 slug 的编辑被拦下了: %v", err)
	}

	tags, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("读回标签失败: %v", err)
	}
	var found *domain.Tag
	for i := range tags {
		if tags[i].ID == "t1" {
			found = &tags[i]
		}
	}
	if found == nil {
		t.Fatal("标签丢失")
	}
	if found.Color != "#222222" {
		t.Errorf("颜色没有被保存，得到 %q", found.Color)
	}
}
