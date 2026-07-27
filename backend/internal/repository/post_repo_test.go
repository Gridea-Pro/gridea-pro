package repository

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gridea-pro/backend/internal/domain"
)

// TestNormalizeFileName 验证 repo 入口归一：去掉末尾 .md（含历史遗留的多重 .md），
// 大小写敏感（只剥小写 .md，与 save 补的小写 .md 对齐），中间的 .md 保留。
// 这是 #122 的核心——保证 repo 内部 FileName 始终是裸名。
func TestNormalizeFileName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"裸名", "hello-world", "hello-world"},
		{"单个 .md", "hello-world.md", "hello-world"},
		{"双 .md（历史 bug 状态）", "hello-world.md.md", "hello-world"},
		{"三 .md", "a.md.md.md", "a"},
		{"空串", "", ""},
		{"仅 .md", ".md", ""},
		{"大写不剥", "hello.MD", "hello.MD"},
		{"中间 .md 保留", "a.md.b", "a.md.b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeFileName(tt.in); got != tt.want {
				t.Errorf("normalizeFileName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCopyFileSamePathDoesNotTruncate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cover.png")
	original := []byte("not a real png, but it must remain intact")
	if err := os.WriteFile(path, original, 0644); err != nil {
		t.Fatal(err)
	}

	// 同文件、不同写法（Windows \→/）：走 sameCleanPath 或 os.SameFile
	if err := CopyFile(path, filepath.ToSlash(path)); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("CopyFile truncated same-path copy: got %d bytes, want %d bytes", len(got), len(original))
	}

	// 语义相同、字符串不同的路径（经 /../ 规范化后等价）：
	// 确保 sameCleanPath 的 filepath.Abs 规范化分支在 Linux CI 中也被覆盖。
	altPath := filepath.Dir(path) + string(os.PathSeparator) + ".." +
		string(os.PathSeparator) + filepath.Base(filepath.Dir(path)) +
		string(os.PathSeparator) + filepath.Base(path)
	if err := CopyFile(path, altPath); err != nil {
		t.Fatal(err)
	}

	got2, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got2, original) {
		t.Fatalf("CopyFile truncated same-file (..-normalized) copy: got %d bytes, want %d bytes", len(got2), len(original))
	}
}

func TestPostRepositoryUpdateKeepsExistingFeatureImage(t *testing.T) {
	ctx := context.Background()
	appDir := t.TempDir()
	repo := NewPostRepository(appDir, nil)

	src := filepath.Join(t.TempDir(), "source-cover.png")
	original := []byte("stable cover image bytes")
	if err := os.WriteFile(src, original, 0644); err != nil {
		t.Fatal(err)
	}

	post := &domain.Post{
		ID:        "issue127",
		Title:     "Issue 127",
		CreatedAt: time.Now(),
		Published: true,
		Content:   "body",
		FileName:  "issue-127",
		FeatureImage: domain.FileInfo{
			Name: "source-cover.png",
			Path: src,
		},
	}
	if err := repo.Create(ctx, post); err != nil {
		t.Fatal(err)
	}

	coverPath := filepath.Join(appDir, "post-images", "issue-127.png")
	post.FeatureImage = domain.FileInfo{
		Name: "issue-127.png",
		Path: filepath.ToSlash(coverPath),
	}
	if err := repo.Update(ctx, post); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(coverPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("existing feature image changed on update: got %d bytes, want %d bytes", len(got), len(original))
	}
}
