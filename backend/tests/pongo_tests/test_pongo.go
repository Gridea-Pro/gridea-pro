//go:build ignore

package main

import (
	"fmt"
	"github.com/flosch/pongo2/v6"
)

func main() {
	// 测试字符串比较
	tplStr := `{% for i in loop %}{% if forloop.Counter0 < count %}{{ i }}{% endif %}{% endfor %}`
	tpl, err := pongo2.FromString(tplStr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	ctx := pongo2.Context{
		"loop":  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		"count": "8", // 测试字符串
	}

	res, err := tpl.Execute(ctx)
	if err != nil {
		fmt.Println("Exec Error:", err)
	} else {
		fmt.Println("Result:", res)
	}
}
