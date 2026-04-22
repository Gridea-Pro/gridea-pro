package utils

import "testing"

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
