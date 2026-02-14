package main

import (
	"fmt"
	"gridea-pro/backend/internal/mcp"
	"os"
)

func main() {
	if _, err := os.Stat(mcp.GetAppDir()); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: GRIDEA_SOURCE_DIR not found at %s\n", mcp.GetAppDir())
		fmt.Fprintln(os.Stderr, "Please set GRIDEA_SOURCE_DIR environment variable to your Gridea Pro data directory.")
		os.Exit(1)
	}

	server := mcp.NewServer()
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting MCP server: %v\n", err)
		os.Exit(1)
	}
}
