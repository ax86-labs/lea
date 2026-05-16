package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andev0x/ctxd/internal/parser/golang"
	"github.com/andev0x/ctxd/internal/parser/treesitter"
	graph "github.com/andev0x/ctxd/internal/graph/contracts"
	"github.com/andev0x/ctxd/internal/storage/sqlite"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index [path]",
	Short: "Index a repository",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		// Ensure .ctxd directory exists
		ctxdDir := filepath.Join(path, ".ctxd")
		if _, err := os.Stat(ctxdDir); os.IsNotExist(err) {
			if err := os.Mkdir(ctxdDir, 0755); err != nil {
				return fmt.Errorf("failed to create .ctxd directory: %w", err)
			}
		}

		dbPath := filepath.Join(ctxdDir, "graph.db")
		store, err := sqlite.NewStore(dbPath)
		if err != nil {
			return err
		}
		defer store.Close()

		goParser := golang.NewParser()
		tsParser := treesitter.NewParser()
		ctx := context.Background()

		err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				// Skip hidden directories (except current dir)
				if strings.HasPrefix(info.Name(), ".") && info.Name() != "." && info.Name() != ".ctxd" {
					return filepath.SkipDir
				}
				// Skip common non-source dirs
				if info.Name() == "vendor" || info.Name() == "node_modules" || info.Name() == "target" {
					return filepath.SkipDir
				}
				return nil
			}

			var nodes []*graph.Node
			var edges []*graph.Edge

			ext := filepath.Ext(filePath)
			if ext == ".go" {
				nodes, edges, _ = goParser.ParseFile(filePath)
				callEdges, _ := goParser.ExtractCalls(filePath)
				flowEdges, _ := goParser.ExtractControlFlow(filePath)
				edges = append(edges, callEdges...)
				edges = append(edges, flowEdges...)
			} else if ext == ".py" || ext == ".rs" || ext == ".ts" {
				nodes, edges, _ = tsParser.ParseFile(ctx, filePath)
			} else {
				return nil
			}

			fmt.Printf("Indexing %s...\n", filePath)
			for _, n := range nodes {
				store.SaveNode(ctx, n)
			}
			for _, e := range edges {
				store.SaveEdge(ctx, e)
			}

			return nil
		})

		if err != nil {
			return err
		}

		fmt.Println("Indexing complete.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)
}
