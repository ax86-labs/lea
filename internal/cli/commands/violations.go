package commands

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/andev0x/ctxd/internal/architecture"
	"github.com/andev0x/ctxd/internal/storage/sqlite"
	"github.com/spf13/cobra"
)

var violationsConfigPath string

var violationsCmd = &cobra.Command{
	Use:   "violations",
	Short: "Detect architecture boundary violations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := architecture.LoadConfig(violationsConfigPath)
		if err != nil {
			return err
		}

		dbPath := filepath.Join(".ctxd", "graph.db")
		store, err := sqlite.NewStore(dbPath)
		if err != nil {
			return err
		}
		defer store.Close()

		ctx := context.Background()
		violations, err := architecture.FindViolations(ctx, store, cfg)
		if err != nil {
			return err
		}
		if len(violations) == 0 {
			fmt.Println("No architecture violations found.")
			return nil
		}

		sort.Slice(violations, func(i, j int) bool {
			if violations[i].FromLayer == violations[j].FromLayer {
				return violations[i].ToLayer < violations[j].ToLayer
			}
			return violations[i].FromLayer < violations[j].FromLayer
		})

		fmt.Printf("Architecture violations (%d):\n", len(violations))
		for _, v := range violations {
			fmt.Printf("- [%s] %s (%s) -> %s (%s)\n  at %s:%d -> %s:%d\n", v.EdgeType, v.FromID, v.FromLayer, v.ToID, v.ToLayer, v.FromFile, v.FromLine, v.ToFile, v.ToLine)
		}

		return nil
	},
}

func init() {
	violationsCmd.Flags().StringVar(&violationsConfigPath, "config", filepath.Join(".ctxd", "architecture.yaml"), "Path to architecture config")
	rootCmd.AddCommand(violationsCmd)
}
