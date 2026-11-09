// Command pbjc copies enriched content to the macOS clipboard.
package main

import (
	"os"

	"github.com/the-inconvenience-store/pbj/internal/command"
)

var version = "dev"

func main() {
	cmd := command.Command{
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Version: version,
	}
	os.Exit(cmd.Run(os.Args[1:]))
}
