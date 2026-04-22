package repository

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"gridea-pro/backend/internal/domain"
)

// newTestAppDir 返回一个临时目录并确保 config/ 子目录存在。
func newTestAppDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "config"), 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	return dir
}

// ─── Tag 唯一性检查 ───────────────────────────────────────────────────────────

func TestTagRepository_Create_RejectsDuplicateName(t *testing.T) {
	ctx := context.Background()
	appDir := newTestAppDir(t)
	repo := NewTagRepository(appDir)

	if err := repo.Create(ctx, &domain.Tag{Name: "Go", Slug: "go"}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	err := repo.Create(ctx, &domain.Tag{Name: "Go", Slug: "golang"})
	if err == nil {
		t.Fatal("expected duplicate name error, got nil")
	}
	if !errors.Is(err, ErrDuplicateName) {
		t.Errorf("expected ErrDuplicateName, got %v", err)
	}
}

func TestTagRepository_Create_RejectsDuplicateSlug(t *testing.T) {
	ctx := context.Background()
	appDir := newTestAppDir(t)
	repo := NewTagRepository(appDir)

	if err := repo.Create(ctx, &domain.Tag{Name: "Go", Slug: "go"}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	err := repo.Create(ctx, &domain.Tag{Name: "Golang", Slug: "go"})
	if err == nil {
		t.Fatal("expected duplicate slug error, got nil")
	}
	if !errors.Is(err, ErrDuplicateSlug) {
		t.Errorf("expected ErrDuplicateSlug, got %v", err)
	}
}

func TestTagRepository_Update_AllowsSelf(t *testing.T) {
	ctx := context.Background()
	appDir := newTestAppDir(t)
	repo := NewTagRepository(appDir)

	tag := &domain.Tag{Name: "Go", Slug: "go"}
	if err := repo.Create(ctx, tag); err != nil {
		t.Fatalf("create: %v", err)
	}
	// 更新自己的 color，不改变 Name / Slug：不应触发冲突
	tag.Color = "#ff0000"
	if err := repo.Update(ctx, tag); err != nil {
		t.Errorf("self-update should succeed, got %v", err)
	}
}

// Issue #99：仅大小写不同的 Name / Slug 在 macOS/Windows 默认大小写不敏感的
// 文件系统上会互相覆盖页面，必须判重。
func TestTagRepository_Create_RejectsCaseInsensitiveDuplicateName(t *testing.T) {
	ctx := context.Background()
	appDir := newTestAppDir(t)
	repo := NewTagRepository(appDir)

	if err := repo.Create(ctx, &domain.Tag{Name: "Go", Slug: "go"}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	err := repo.Create(ctx, &domain.Tag{Name: "GO", Slug: "golang"})
	if !errors.Is(err, ErrDuplicateName) {
		t.Errorf("GO vs Go should collide, got %v", err)
	}
}

func TestTagRepository_Create_RejectsCaseInsensitiveDuplicateSlug(t *testing.T) {
	ctx := context.Background()
	appDir := newTestAppDir(t)
	repo := NewTagRepository(appDir)

	if err := repo.Create(ctx, &domain.Tag{Name: "Go", Slug: "go"}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	err := repo.Create(ctx, &domain.Tag{Name: "Golang", Slug: "GO"})
	if !errors.Is(err, ErrDuplicateSlug) {
		t.Errorf("GO vs go should collide, got %v", err)
	}
}

func TestTagRepository_Update_RejectsCollisionWithOther(t *testing.T) {
	ctx := context.Background()
	appDir := newTestAppDir(t)
	repo := NewTagRepository(appDir)

	go1 := &domain.Tag{Name: "Go", Slug: "go"}
	rust := &domain.Tag{Name: "Rust", Slug: "rust"}
	if err := repo.Create(ctx, go1); err != nil {
		t.Fatalf("create go: %v", err)
	}
	if err := repo.Create(ctx, rust); err != nil {
		t.Fatalf("create rust: %v", err)
	}
	// 把 rust 改成 Name = "Go"（与 go1 冲突）
	rust.Name = "Go"
	err := repo.Update(ctx, rust)
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
	if !errors.Is(err, ErrDuplicateName) {
		t.Errorf("expected ErrDuplicateName, got %v", err)
	}
}

// ─── Category 唯一性检查 ──────────────────────────────────────────────────────

func TestCategoryRepository_Create_RejectsDuplicateName(t *testing.T) {
	ctx := context.Background()
	appDir := newTestAppDir(t)
	repo := NewCategoryRepository(appDir)

	if err := repo.Create(ctx, &domain.Category{Name: "Tech", Slug: "tech"}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	err := repo.Create(ctx, &domain.Category{Name: "Tech", Slug: "technology"})
	if err == nil {
		t.Fatal("expected duplicate name error")
	}
	if !errors.Is(err, ErrDuplicateName) {
		t.Errorf("expected ErrDuplicateName, got %v", err)
	}
}

func TestCategoryRepository_Create_RejectsDuplicateSlug(t *testing.T) {
	ctx := context.Background()
	appDir := newTestAppDir(t)
	repo := NewCategoryRepository(appDir)

	if err := repo.Create(ctx, &domain.Category{Name: "Tech", Slug: "tech"}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	err := repo.Create(ctx, &domain.Category{Name: "Technology", Slug: "tech"})
	if err == nil {
		t.Fatal("expected duplicate slug error")
	}
	if !errors.Is(err, ErrDuplicateSlug) {
		t.Errorf("expected ErrDuplicateSlug, got %v", err)
	}
}

func TestCategoryRepository_Create_RejectsCaseInsensitiveDuplicate(t *testing.T) {
	ctx := context.Background()
	appDir := newTestAppDir(t)
	repo := NewCategoryRepository(appDir)

	if err := repo.Create(ctx, &domain.Category{Name: "Tech", Slug: "tech"}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	if err := repo.Create(ctx, &domain.Category{Name: "TECH", Slug: "technology"}); !errors.Is(err, ErrDuplicateName) {
		t.Errorf("TECH vs Tech name should collide, got %v", err)
	}
	if err := repo.Create(ctx, &domain.Category{Name: "Other", Slug: "TECH"}); !errors.Is(err, ErrDuplicateSlug) {
		t.Errorf("TECH vs tech slug should collide, got %v", err)
	}
}

// ─── 启动期审计 ──────────────────────────────────────────────────────────────

func TestAuditTagUniqueness_FindsHistoricalDuplicates(t *testing.T) {
	ctx := context.Background()
	appDir := newTestAppDir(t)

	// 手工写入含重名的 tags.json，模拟用户手动编辑场景
	content := `{"tags":[
		{"id":"a","name":"Go","slug":"go"},
		{"id":"b","name":"Go","slug":"golang"},
		{"id":"c","name":"Rust","slug":"go"}
	]}`
	if err := os.WriteFile(filepath.Join(appDir, "config", "tags.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write tags.json: %v", err)
	}

	repo := NewTagRepository(appDir)
	conflicts := AuditTagUniqueness(ctx, repo)

	if len(conflicts) < 2 {
		t.Fatalf("expected at least 2 conflicts (name + slug), got %d: %v", len(conflicts), conflicts)
	}
	// 应同时检出 Name "Go" 重复和 Slug "go" 重复
	foundName := false
	foundSlug := false
	for _, c := range conflicts {
		if contains(c, "Go") && contains(c, "标签名") {
			foundName = true
		}
		if contains(c, "go") && contains(c, "Slug") {
			foundSlug = true
		}
	}
	if !foundName {
		t.Errorf("expected Name conflict in %v", conflicts)
	}
	if !foundSlug {
		t.Errorf("expected Slug conflict in %v", conflicts)
	}
}

// Issue #99：审计也必须识别仅大小写不同的历史重复。
func TestAuditTagUniqueness_FindsCaseInsensitiveDuplicates(t *testing.T) {
	ctx := context.Background()
	appDir := newTestAppDir(t)

	content := `{"tags":[
		{"id":"a","name":"Go","slug":"go"},
		{"id":"b","name":"GO","slug":"golang"},
		{"id":"c","name":"Rust","slug":"GO"}
	]}`
	if err := os.WriteFile(filepath.Join(appDir, "config", "tags.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write tags.json: %v", err)
	}

	repo := NewTagRepository(appDir)
	conflicts := AuditTagUniqueness(ctx, repo)
	if len(conflicts) < 2 {
		t.Fatalf("expected case-insensitive Name + Slug conflicts, got %d: %v", len(conflicts), conflicts)
	}
}

func TestAuditCategoryUniqueness_NoConflicts(t *testing.T) {
	ctx := context.Background()
	appDir := newTestAppDir(t)
	repo := NewCategoryRepository(appDir)

	if err := repo.Create(ctx, &domain.Category{Name: "A", Slug: "a"}); err != nil {
		t.Fatalf("create a: %v", err)
	}
	if err := repo.Create(ctx, &domain.Category{Name: "B", Slug: "b"}); err != nil {
		t.Fatalf("create b: %v", err)
	}

	conflicts := AuditCategoryUniqueness(ctx, repo)
	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts, got %v", conflicts)
	}
}

// 简易 contains 避免引入额外包
func contains(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) && findIndex(haystack, needle) >= 0)
}

func findIndex(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
