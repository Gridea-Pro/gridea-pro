package utils

import (
	"errors"
	"testing"
)

func TestSlugifyName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// 英文 / 数字
		{"english_basic", "Hello World", "hello-world"},
		{"english_punct", "Hello, World!", "hello-world"},
		{"mixed_case", "HelloWorld", "helloworld"},
		{"numbers", "Go 1.22", "go-1-22"},
		{"hyphen_passthrough", "abc-def", "abc-def"},

		// 中文 → 拼音
		{"chinese_simple", "测试", "ce-shi"},
		{"chinese_longer", "你好世界", "ni-hao-shi-jie"},
		{"chinese_mixed", "hello 世界", "hello-shi-jie"},

		// 非法 URL 字符被清理
		{"slash", "a/b/c", "a-b-c"},
		{"hash", "a#b", "a-b"},
		{"question", "a?b", "a-b"},
		{"space_only_inside", "foo  bar", "foo-bar"},

		// 空 / 无有效字符（期望空串，由调用方决定兜底）
		{"empty", "", ""},
		{"pure_emoji", "🌟✨", ""},
		{"pure_symbol", "!!!", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SlugifyName(tt.input)
			if got != tt.want {
				t.Errorf("SlugifyName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateSlug(t *testing.T) {
	cases := []struct {
		name  string
		input string
		ok    bool
	}{
		// 合法
		{"simple", "abc", true},
		{"digits_only", "123", true},
		{"alnum_mix", "go1-22", true},
		{"multi_hyphen", "a-b-c", true},
		{"long", "hello-world-foo-bar", true},

		// 不合法 —— 空串 / 边界
		{"empty", "", false},
		{"single_hyphen", "-", false},
		{"leading_hyphen", "-abc", false},
		{"trailing_hyphen", "abc-", false},
		{"double_hyphen", "a--b", false},

		// 不合法 —— 大小写混用（#99 核心回归）
		{"uppercase_only", "ABC", false},
		{"mixed_case", "AbC", false},
		{"one_upper", "aBc", false},

		// 不合法 —— URL 保留字符（#99 触发场景）
		{"hash", "c#", false},
		{"plus", "c++", false},
		{"slash", "a/b", false},
		{"question", "a?b", false},
		{"space", "a b", false},

		// 不合法 —— 非 ASCII
		{"chinese", "测试", false},
		{"accent", "café", false},
		{"emoji", "🔥", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSlug(tc.input)
			if tc.ok && err != nil {
				t.Errorf("ValidateSlug(%q) returned err %v, want ok", tc.input, err)
			}
			if !tc.ok {
				if err == nil {
					t.Errorf("ValidateSlug(%q) returned nil, want error", tc.input)
				}
				if err != nil && !errors.Is(err, ErrInvalidSlug) {
					t.Errorf("ValidateSlug(%q) err not ErrInvalidSlug: %v", tc.input, err)
				}
			}
		})
	}
}

// SlugifyName 产出的所有 slug 都必须满足 ValidateSlug —— 这是两个函数之间的约定。
func TestSlugifyName_OutputPassesValidateSlug(t *testing.T) {
	inputs := []string{
		"Hello World",
		"Go 1.22",
		"测试",
		"hello 世界",
		"a#b",
		"a/b/c",
		"abc-def",
	}
	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			out := SlugifyName(in)
			if out == "" {
				t.Skip("empty output, not subject to ValidateSlug")
			}
			if err := ValidateSlug(out); err != nil {
				t.Errorf("SlugifyName(%q) = %q failed ValidateSlug: %v", in, out, err)
			}
		})
	}
}
