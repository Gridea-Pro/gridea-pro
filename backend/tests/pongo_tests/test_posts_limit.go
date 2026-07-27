//go:build ignore

package main

import (
	"fmt"
	"github.com/flosch/pongo2/v6"
)

func main() {
	tplStr := `{% for i in loop %}{% if forloop.Counter0 < count|default:"4" %}X{% endif %}{% endfor %}`
	tpl, _ := pongo2.FromString(tplStr)

	ctx_12 := pongo2.Context{
		"loop":  make([]int, 8),
		"count": "12",
	}
	res2, _ := tpl.Execute(ctx_12)
	fmt.Println("Count '12' with 8 items in loop:", res2)
}
