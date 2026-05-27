package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/PizenLabs/lea/internal/storage/sqlite"
	"github.com/spf13/cobra"
)

var impactCmd = &cobra.Command{
	Use:   "impact [symbol_id]",
	Short: "Find symbols that depend on this symbol",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		symbolID := args[0]

		dbPath := filepath.Join(".ctxd", "graph.db")
		store, err := sqlite.NewStore(dbPath)
		if err != nil {
			return err
		}
		defer store.Close()

		ctx := context.Background()
		nodes, edges, err := store.GetInboundEdges(ctx, symbolID)
		if err != nil {
			return err
		}

		if len(nodes) == 0 {
			fmt.Printf("No impact found for %s\n", symbolID)
			return nil
		}

		fmt.Printf("Impact of %s (symbols that depend on it):\n", symbolID)
		for i, n := range nodes {
			e := edges[i]
			fmt.Printf("- %s (%s) [%s] at %s:%d\n", n.Name, n.Type, e.Type, n.File, n.Line)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(impactCmd)
}
