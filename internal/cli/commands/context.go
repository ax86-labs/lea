package commands

import (
	"context"
	"fmt"
	"path/filepath"

	aictx "github.com/ax86-labs/lea/internal/ai/context"
	"github.com/ax86-labs/lea/internal/storage/sqlite"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context [symbol_id]",
	Short: "Generate AI-optimized context for a symbol",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		symbolID := args[0]

		dbPath := filepath.Join(".ctxd", "graph.db")
		store, err := sqlite.NewStore(dbPath)
		if err != nil {
			return err
		}
		defer store.Close()

		compiler := aictx.NewCompiler(store)
		ctx := context.Background()

		output, err := compiler.Compile(ctx, symbolID)
		if err != nil {
			return err
		}

		fmt.Println(output)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(contextCmd)
}
