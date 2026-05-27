package main

import (
	"fmt"
	"os"

	"github.com/ax86-labs/lea/internal/cli/commands"
)

var Version = "dev"

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
