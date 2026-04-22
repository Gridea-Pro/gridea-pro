package utils

import (
	"errors"
	"regexp"
	"strings"
	"unicode"

	"github.com/gosimple/slug"
	"github.com/mozillazg/go-pinyin"
)

// ErrInvalidSlug 用户输入的 slug 不符合 URL-safe 规则。调用方用 errors.Is 识别。
var ErrInvalidSlug = errors.New("invalid slug")

// slugPattern 允许的 slug 形态：小写字母 / 数字 / 中间的连字符。
//
// 为什么只允许小写：
//   - macOS（APFS 默认）/ Windows 的文件系统大小写不敏感，
//     若 slug 大小写混用，磁盘上会和另一个仅大小写不同的 slug 撞同一个目录，
//     互相覆盖生成的 tag/category 页
//   - URL 层面大小写是敏感的，`Slug-A` 和 `slug-a` 会被当成两个不同地址，
//     造成 SEO 与用户期望不符
//
// 为什么禁中文 / 特殊字符：
//   - `#` 在 URL 里是 fragment 分隔符，`/tag/C#/` 浏览器解析为 `/tag/C` + `#/`
//   - `/`、`?`、`&`、空格、emoji 等同理会破坏 URL 路径或引入编码歧义
//   - 想用中文标签名，由 SlugifyName（拼音归一）自动生成 slug，别让用户手填
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

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

// ValidateSlug 校验用户输入的 slug 是否可直接作为 URL path 段 + 跨平台文件名。
// 不合法时返回 wrap 了 ErrInvalidSlug 的 error，消息面向用户可直接展示。
//
// 合法规则：`^[a-z0-9]+(-[a-z0-9]+)*$`
//   - 只允许小写字母、数字、中间的连字符
//   - 不允许空串 / 前后连字符 / 连续连字符
func ValidateSlug(s string) error {
	if s == "" {
		return ErrInvalidSlug
	}
	if !slugPattern.MatchString(s) {
		return ErrInvalidSlug
	}
	return nil
}
