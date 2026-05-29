// Package main provides the lea CLI entrypoint.
package main

import (
	"fmt"
	"os"

	"github.com/PizenLabs/lea/internal/cli/commands"
)

// Version is the current build version.
var Version = "dev"

func main() {
	commands.Version = Version
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
