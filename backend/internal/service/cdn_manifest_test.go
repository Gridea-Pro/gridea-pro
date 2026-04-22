package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCdnManifest_EmptyLoadReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	m := loadCdnManifest(dir)
	if m == nil {
		t.Fatal("loadCdnManifest returned nil")
	}
	if len(m.Entries) != 0 {
		t.Errorf("expected empty entries, got %v", m.Entries)
	}
}

func TestCdnManifest_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	m := loadCdnManifest(dir)
	m.record("post-images/a.png", "sha-a")
	m.record("post-images/b.png", "sha-b")
	if err := m.save(dir); err != nil {
		t.Fatalf("save: %v", err)
	}

	// 再读一次应该拿到同样的内容
	m2 := loadCdnManifest(dir)
	if sha, ok := m2.hit("post-images/a.png"); !ok || sha != "sha-a" {
		t.Errorf("expected a.png sha-a, got %v %v", sha, ok)
	}
	if sha, ok := m2.hit("post-images/b.png"); !ok || sha != "sha-b" {
		t.Errorf("expected b.png sha-b, got %v %v", sha, ok)
	}
	if _, ok := m2.hit("missing.png"); ok {
		t.Error("expected miss on unknown path")
	}
}

func TestCdnManifest_CorruptFileTreatedAsEmpty(t *testing.T) {
	dir := t.TempDir()
	// 写一个无法解析的 JSON
	_ = os.WriteFile(filepath.Join(dir, ".cdn-manifest.json"), []byte("{{{not json"), 0o644)

	m := loadCdnManifest(dir)
	if m == nil {
		t.Fatal("loadCdnManifest returned nil")
	}
	if len(m.Entries) != 0 {
		t.Errorf("corrupt manifest should fall back to empty, got %v", m.Entries)
	}
	// 并且 record / save 仍应工作
	m.record("x", "y")
	if err := m.save(dir); err != nil {
		t.Fatalf("save after corrupt reload: %v", err)
	}
}

func TestCdnManifest_RecordUpdatesExisting(t *testing.T) {
	dir := t.TempDir()
	m := loadCdnManifest(dir)
	m.record("a.png", "old-sha")
	m.record("a.png", "new-sha")
	sha, ok := m.hit("a.png")
	if !ok || sha != "new-sha" {
		t.Errorf("expected new-sha, got %v %v", sha, ok)
	}
}
