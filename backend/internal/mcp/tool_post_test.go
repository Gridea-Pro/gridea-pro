package mcp

import "testing"

// TestNormalizeFileName verifies that normalizeFileName strips any trailing
// ".md" extension(s) so callers can pass either "hello-world" or
// "hello-world.md" interchangeably. This is the fix for issues #117 and #119:
// without normalization, passing "hello-world.md" would produce a file named
// "hello-world.md.md" on disk and a cache entry keyed by "hello-world.md",
// causing update_post / get_post / delete_post to silently target the wrong
// file.
func TestNormalizeFileName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no extension", "hello-world", "hello-world"},
		{"single extension", "hello-world.md", "hello-world"},
		{"double extension (legacy bug state)", "hello-world.md.md", "hello-world"},
		{"triple extension", "hello.md.md.md", "hello"},
		{"empty string", "", ""},
		{"only extension", ".md", ""},
		{"trailing dot then md", "..md", "."},
		{"case-sensitive (only lowercase stripped)", "hello.MD", "hello.MD"},
		{"md in middle is preserved", "hello.md.world", "hello.md.world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeFileName(tt.in)
			if got != tt.want {
				t.Errorf("normalizeFileName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
