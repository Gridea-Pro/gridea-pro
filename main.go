package main

import (
	"embed"
	"gridea-pro/backend/pkg/boot"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	boot.Run(assets)
}
