package commands

import (
	"context"
	"os"
	"path/filepath"

	"github.com/PizenLabs/lea/internal/storage/sqlite"
	"github.com/PizenLabs/lea/internal/watcher"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch [path]",
	Short: "Watch a repository for changes and update the index incrementally",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		ctxdDir := filepath.Join(path, ".ctxd")
		if err := os.MkdirAll(ctxdDir, 0755); err != nil {
			return err
		}

		dbPath := filepath.Join(ctxdDir, "graph.db")
		store, err := sqlite.NewStore(dbPath)
		if err != nil {
			return err
		}
		defer func() { _ = store.Close() }()

		w := watcher.NewWatcher(store, path)
		return w.Start(context.Background())
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
