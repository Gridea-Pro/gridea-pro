package deploy

import "testing"

func TestIsUserPageRepo(t *testing.T) {
	tests := []struct {
		repo string
		want bool
	}{
		{"alice/alice.github.io", true},
		{"alice/alice.gitee.io", true},
		{"alice/alice.coding.me", true},
		{"alice/my-blog", false},
		{"alice/gridea-site", false},
		// 带仓库主机前缀的完整路径
		{"github.com/alice/alice.github.io", true},
		{"github.com/alice/my-blog", false},
		{"github.com/alice/alice.github.io.git", true},
		// 边界
		{"", false},
		{"alice", false},
		{"/", false},
	}
	for _, tt := range tests {
		t.Run(tt.repo, func(t *testing.T) {
			if got := isUserPageRepo(tt.repo); got != tt.want {
				t.Errorf("isUserPageRepo(%q) = %v, want %v", tt.repo, got, tt.want)
			}
		})
	}
}

func TestDestructiveBranchesSet(t *testing.T) {
	for _, name := range []string{"main", "master", "develop", "release"} {
		if !destructiveBranches[name] {
			t.Errorf("expected %q to be in destructive set", name)
		}
	}
	for _, name := range []string{"gh-pages", "pages", "deploy"} {
		if destructiveBranches[name] {
			t.Errorf("%q should not be destructive", name)
		}
	}
}
