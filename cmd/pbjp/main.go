// Command pbjp extracts enriched content from the macOS clipboard.
package main

import (
	"os"

	"github.com/the-inconvenience-store/pbj/internal/command"
)

var version = "dev"

func main() {
	cmd := command.Command{Stdout: os.Stdout, Stderr: os.Stderr, Version: version}
	os.Exit(cmd.Paste(os.Args[1:]))
}
