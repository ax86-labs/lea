package commands

import (
	"path/filepath"

	"github.com/PizenLabs/lea/internal/storage/sqlite"
	"github.com/PizenLabs/lea/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Start the interactive TUI",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := filepath.Join(".ctxd", "graph.db")
		store, err := sqlite.NewStore(dbPath)
		if err != nil {
			return err
		}
		defer store.Close()

		return tui.Start(store)
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
