package commands

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	graph "github.com/ax86-labs/lea/internal/graph/contracts"
	"github.com/ax86-labs/lea/internal/parser/golang"
	"github.com/ax86-labs/lea/internal/parser/treesitter"
	"github.com/ax86-labs/lea/internal/storage/sqlite"
	"github.com/ax86-labs/lea/internal/workspace/ignore"
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
		matcher := ignore.NewMatcher(path)

		err = filepath.WalkDir(path, func(filePath string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if matcher.ShouldSkipDir(filePath, entry) {
					return filepath.SkipDir
				}
				return nil
			}
			if matcher.ShouldSkipFile(filePath, entry) {
				return nil
			}

			var nodes []*graph.Node
			var edges []*graph.Edge

			ext := strings.ToLower(filepath.Ext(filePath))
			switch ext {
			case ".go":
				var parseErr error
				nodes, edges, parseErr = goParser.ParseFile(ctx, filePath)
				if parseErr != nil {
					return fmt.Errorf("parse %s: %w", filePath, parseErr)
				}
				callEdges, callErr := goParser.ExtractCalls(ctx, filePath)
				if callErr != nil {
					return fmt.Errorf("extract calls %s: %w", filePath, callErr)
				}
				flowEdges, flowErr := goParser.ExtractControlFlow(ctx, filePath)
				if flowErr != nil {
					return fmt.Errorf("extract flow %s: %w", filePath, flowErr)
				}
				edges = append(edges, callEdges...)
				edges = append(edges, flowEdges...)
			case ".py", ".rs", ".ts":
				var parseErr error
				nodes, edges, parseErr = tsParser.ParseFile(ctx, filePath)
				if parseErr != nil {
					return fmt.Errorf("parse %s: %w", filePath, parseErr)
				}
			default:
				return nil
			}

			if len(nodes) == 0 && len(edges) == 0 {
				return nil
			}

			fmt.Printf("Indexing %s...\n", filePath)
			if err := store.SaveGraph(ctx, nodes, edges); err != nil {
				return fmt.Errorf("persist graph for %s: %w", filePath, err)
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
