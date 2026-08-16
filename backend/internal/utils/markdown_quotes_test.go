package utils

import "testing"

// TestSmartQuotesCJK 中文引号必须成对。
// goldmark 的 Typographer 判定闭引号要求引号右侧是空白或标点，中文里
// `力"这` 两侧都是汉字，判定落空——开引号被换成弯引号、闭引号仍是直引号。
func TestSmartQuotesCJK(t *testing.T) {
	cases := []struct{ name, md, want string }{
		{"中文引号成对", `"决策消耗意志力"这个说法`, "“决策消耗意志力”这个说法"},
		{"中文引号在句中", `五处"这样改行不行"，算五次。`, "五处“这样改行不行”，算五次。"},
		{"一段内多组引号", `"一"和"二"`, "“一”和“二”"},
	}
	for _, c := range cases {
		if got := body(c.md); got != c.want {
			t.Errorf("[%s] 得到 %q，期望 %q", c.name, got, c.want)
		}
	}
}

// TestSmartQuotesLatin 接管双引号后，英文场景不能回退。
func TestSmartQuotesLatin(t *testing.T) {
	if got, want := body(`He said "hello" to me.`), "He said “hello” to me."; got != want {
		t.Errorf("英文双引号：得到 %q，期望 %q", got, want)
	}
	// 单引号仍归 Typographer——英文撇号（it's）与闭单引号无法靠开闭交替区分。
	// Typographer 输出的是 HTML 实体，不是字面字符。
	if got, want := body("it's fine"), "it&rsquo;s fine"; got != want {
		t.Errorf("英文撇号：得到 %q，期望 %q", got, want)
	}
}

// TestSmartQuotesNotAcrossBlocks 引号状态不跨块，
// 否则一个未闭合的引号会让后面所有段落的引号方向全部颠倒。
func TestSmartQuotesNotAcrossBlocks(t *testing.T) {
	if got, want := body("\"第一段\n\n\"第二段"), "“第一段"; got != want {
		t.Errorf("跨段落状态泄漏：得到 %q，期望 %q", got, want)
	}
}
