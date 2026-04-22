package utils

import (
	"strings"
	"unicode"

	"github.com/gosimple/slug"
	"github.com/mozillazg/go-pinyin"
)

// SlugifyName 将人类可读的名称（可能含中文、标点、空格等）转成 URL-safe 的 slug。
//
// 处理流程：
//  1. 逐 rune 扫描：中文字符替换为「空格 + 拼音 + 空格」，其余字符原样保留
//  2. 将中间字符串交给 gosimple/slug.Make 做 ASCII 规范化
//     （小写、去特殊字符、连字符合并）
//
// 混排场景（"hello 世界"）也能正确得到 "hello-shi-jie" 而不是把英文逐字母拆散。
//
// 若结果为空串（例如纯 emoji / 符号名），返回空串由调用方决定兜底策略：
//   - 需要持久化时建议用 nanoid 保证唯一
//   - 仅做视图展示时建议用 url.PathEscape(name) 保证确定性 URL
func SlugifyName(name string) string {
	if name == "" {
		return ""
	}

	pinyinArgs := pinyin.NewArgs()

	var b strings.Builder
	b.Grow(len(name) * 2)
	for _, r := range name {
		if unicode.Is(unicode.Han, r) {
			rows := pinyin.Pinyin(string(r), pinyinArgs)
			if len(rows) > 0 && len(rows[0]) > 0 {
				b.WriteByte(' ')
				b.WriteString(rows[0][0])
				b.WriteByte(' ')
			}
		} else {
			b.WriteRune(r)
		}
	}

	return slug.Make(b.String())
}
