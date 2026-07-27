//go:build ignore

package main

import (
	"fmt"
	"github.com/flosch/pongo2/v6"
)

func main() {
	tplStr := `{% for i in loop %}{% if forloop.Counter0 < count|default:"4" %}{{ forloop.Counter0 }} {% endif %}{% endfor %}`
	tpl, _ := pongo2.FromString(tplStr)

	ctx_8 := pongo2.Context{
		"loop":  make([]int, 20),
		"count": "8",
	}
	res, _ := tpl.Execute(ctx_8)
	fmt.Println("Count '8':", res)

	ctx_12 := pongo2.Context{
		"loop":  make([]int, 20),
		"count": "12",
	}
	res2, _ := tpl.Execute(ctx_12)
	fmt.Println("Count '12':", res2)
}
