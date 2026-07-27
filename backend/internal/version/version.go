// Package version 暴露 Gridea Pro 的产品名与版本号。
//
// Version 是唯一的版本号真源，CI 通过 ldflags 在编译期注入：
//
//	-X gridea-pro/backend/internal/version.Version=${VERSION#v}
//
// ⚠ 必须保持 var 形式 + 字符串字面量初始化。改成 const 或函数调用初始化
//
//	会让 Go 链接器的 -X 注入静默失效（常量被内联，函数调用覆盖注入值），
//	导致发版后应用内仍显示旧版本号、自更新反复弹窗。
package version

// Product 是产品名称，用于 <meta name="generator"> 等场景。
const Product = "Gridea Pro"

// Version 是当前版本号。本地开发为 "0.0.0-dev"，CI 发版时由 ldflags 覆盖。
var Version = "0.0.0-dev"

// Generator 返回 <meta name="generator"> 的 content 值。
func Generator() string {
	return Product + " " + Version
}
