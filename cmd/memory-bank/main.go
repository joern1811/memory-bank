package main

import (
	"os"

	"github.com/joern1811/memory-bank/internal/infra/cli"
)

func main() {
	// Check if we should run in server mode (no args or "server" command)
	if len(os.Args) == 1 {
		// No arguments provided, run as MCP server for backward compatibility
		os.Args = append(os.Args, "server")
	}

	// Execute CLI
	cli.Execute()
}