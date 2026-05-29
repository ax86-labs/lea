package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the current build version for the CLI.
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the lea version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
