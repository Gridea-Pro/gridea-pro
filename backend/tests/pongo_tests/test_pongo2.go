//go:build ignore

package main

import (
	"fmt"
	"github.com/flosch/pongo2/v6"
)

func main() {
	tplStr := `{% for i in loop %}{% if forloop.Counter0 < count|default:"4" %}{{ i }}{% endif %}{% endfor %}`
	tpl, err := pongo2.FromString(tplStr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 测试未定义
	ctx1 := pongo2.Context{
		"loop": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		// "count": "8",
	}
	res1, _ := tpl.Execute(ctx1)
	fmt.Println("Result without count:", res1)

	// 测试字符串
	ctx2 := pongo2.Context{
		"loop":  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		"count": "8",
	}
	res2, _ := tpl.Execute(ctx2)
	fmt.Println("Result with string count:", res2)

	// 测试数字
	ctx3 := pongo2.Context{
		"loop":  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		"count": 8,
	}
	res3, _ := tpl.Execute(ctx3)
	fmt.Println("Result with int count:", res3)
}
