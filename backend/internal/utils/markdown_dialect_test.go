package utils

import (
	"strings"
	"testing"
)

// body 取出单段 HTML 的正文部分，便于逐字比对。
func body(md string) string {
	out := strings.TrimSpace(ToHTMLUnsafe(md))
	out = strings.TrimPrefix(out, "<p>")
	if i := strings.Index(out, "</p>"); i >= 0 {
		out = out[:i]
	}
	return out
}

// TestDialectMarkAndScript 校验 ==高亮== / ^上标^ / ~下标~ 三种方言。
// 这三种编辑器侧一直支持，后端此前没有对应扩展：高亮发布后变成字面 ==，
// 下标还会被 GFM 删除线扩展吃掉渲染成 <del>。
func TestDialectMarkAndScript(t *testing.T) {
	cases := []struct{ name, md, want string }{
		{"高亮", "这是 ==高亮== 文本", "这是 <mark>高亮</mark> 文本"},
		{"高亮内嵌加粗", "==**粗体高亮**==", "<mark><strong>粗体高亮</strong></mark>"},
		{"高亮允许内部空格", "==多 词 高亮==", "<mark>多 词 高亮</mark>"},
		{"上标", "x^2^ 结束", "x<sup>2</sup> 结束"},
		{"下标", "H~2~O", "H<sub>2</sub>O"},
	}
	for _, c := range cases {
		if got := body(c.md); got != c.want {
			t.Errorf("[%s] 得到 %q，期望 %q", c.name, got, c.want)
		}
	}
}

// TestDialectBoundaries 校验方言不越界——这几条都是实现时真踩过的坑。
func TestDialectBoundaries(t *testing.T) {
	cases := []struct{ name, md, want string }{
		// GFM 删除线是长度 2 的 ~~，不能被长度 1 的下标抢走
		{"删除线不受影响", "~~删除~~", "<del>删除</del>"},
		{"下标与删除线并存", "H~2~O 和 ~~删除~~", "H<sub>2</sub>O 和 <del>删除</del>"},
		// 定界符长度必须精确匹配。首字符被拒后 goldmark 会从下一个字符重试，
		// 那里恰好只剩两个 =，曾把 ===x=== 误判成高亮。
		{"三等号不是高亮", "===不是高亮===", "===不是高亮==="},
		// 上下标内容禁含空白，否则正文里孤立的 ^ / ~ 会被误判
		{"孤立脱字符不误判", "温度 25^C 到 30^C", "温度 25^C 到 30^C"},
		{"上标含空格不匹配", "^a b^", "^a b^"},
	}
	for _, c := range cases {
		if got := body(c.md); got != c.want {
			t.Errorf("[%s] 得到 %q，期望 %q", c.name, got, c.want)
		}
	}
}
