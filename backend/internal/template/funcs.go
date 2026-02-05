package template

import (
	htmltemplate "html/template"
	"strings"
	"time"
)

// TemplateFuncs 返回模板可用的自定义函数
func TemplateFuncs() htmltemplate.FuncMap {
	return htmltemplate.FuncMap{
		// 安全 HTML 输出（不转义）
		"safeHTML": func(s string) htmltemplate.HTML {
			return htmltemplate.HTML(s)
		},

		// 安全 CSS 输出
		"safeCSS": func(s string) htmltemplate.CSS {
			return htmltemplate.CSS(s)
		},

		// 安全 JS 输出
		"safeJS": func(s string) htmltemplate.JS {
			return htmltemplate.JS(s)
		},

		// 安全 URL 输出
		"safeURL": func(s string) htmltemplate.URL {
			return htmltemplate.URL(s)
		},

		// or 函数 - 返回第一个非空值
		"or": func(values ...interface{}) interface{} {
			for _, v := range values {
				if v != nil && v != "" && v != false && v != 0 {
					return v
				}
			}
			if len(values) > 0 {
				return values[len(values)-1]
			}
			return nil
		},

		// 日期格式化
		"formatDate": func(t time.Time, format string) string {
			// 将常见格式转换为 Go 格式
			format = convertDateFormat(format)
			return t.Format(format)
		},

		// 字符串连接
		"join": func(sep string, items []string) string {
			return strings.Join(items, sep)
		},

		// 默认值
		"default": func(defaultVal, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},

		// 截断字符串
		"truncate": func(length int, s string) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},

		// 检查切片是否为空
		"empty": func(v interface{}) bool {
			if v == nil {
				return true
			}
			switch val := v.(type) {
			case string:
				return val == ""
			case []interface{}:
				return len(val) == 0
			case []string:
				return len(val) == 0
			}
			return false
		},

		// 不为空
		"notEmpty": func(v interface{}) bool {
			if v == nil {
				return false
			}
			switch val := v.(type) {
			case string:
				return val != ""
			case []interface{}:
				return len(val) > 0
			case []string:
				return len(val) > 0
			case bool:
				return val
			}
			return true
		},

		// 获取当前时间戳（用于缓存刷新）
		"now": func() int64 {
			return time.Now().Unix()
		},

		// URL 路径拼接
		"urlJoin": func(parts ...string) string {
			var result strings.Builder
			for i, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				if i > 0 {
					part = strings.TrimPrefix(part, "/")
				}
				part = strings.TrimSuffix(part, "/")
				if result.Len() > 0 && !strings.HasSuffix(result.String(), "/") {
					result.WriteString("/")
				}
				result.WriteString(part)
			}
			return result.String()
		},
	}
}

// convertDateFormat 将通用日期格式转换为 Go 格式
func convertDateFormat(format string) string {
	// 常见格式映射
	replacements := map[string]string{
		"YYYY": "2006",
		"YY":   "06",
		"MM":   "01",
		"DD":   "02",
		"HH":   "15",
		"mm":   "04",
		"ss":   "05",
		"M":    "1",
		"D":    "2",
	}

	result := format
	for from, to := range replacements {
		result = strings.ReplaceAll(result, from, to)
	}

	return result
}
