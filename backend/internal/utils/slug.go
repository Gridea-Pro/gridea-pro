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

// slugPattern 允许的 slug 形态：字母（不限大小写）/ 数字 / 中间的连字符。
//
// 为什么允许大写：
//   - 「两个仅大小写不同的 slug 在大小写不敏感的文件系统上撞同一个目录」这个风险，
//     由 repository 的唯一性校验独立堵住（Name/Slug 比较走 EqualFold，见 uniqueness.go），
//     系统里不可能同时存在 `Tech` 与 `tech`。禁大写对这个故障没有额外防护作用。
//   - 反过来，禁大写会让历史数据（旧版本写入的、或经 MCP 由外部写入的带大写 slug）
//     连带无法被编辑——用户只想改描述或封面，却会被 slug 格式拦下。
//   - 自动生成仍统一产出小写（见 SlugifyName），大写只在用户显式指定时出现。
//
// 为什么禁中文 / 特殊字符：
//   - `#` 在 URL 里是 fragment 分隔符，`/tag/C#/` 浏览器解析为 `/tag/C` + `#/`
//   - `/`、`?`、`&`、空格、emoji 等同理会破坏 URL 路径或引入编码歧义
//   - 想用中文标签名，由 SlugifyName（拼音归一）自动生成 slug，别让用户手填
var slugPattern = regexp.MustCompile(`^[A-Za-z0-9]+(-[A-Za-z0-9]+)*$`)

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
// 合法规则：`^[A-Za-z0-9]+(-[A-Za-z0-9]+)*$`
//   - 只允许字母（不限大小写）、数字、中间的连字符
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
