package treesitter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	graph "github.com/ax86-labs/lea/internal/graph/contracts"
	"github.com/ax86-labs/lea/internal/parser/treesitter/python"
	sitter "github.com/tree-sitter/go-tree-sitter"
	"github.com/tree-sitter/tree-sitter-python/bindings/go"
	"github.com/tree-sitter/tree-sitter-rust/bindings/go"
	"github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

type Parser struct {
	languages map[string]*sitter.Language
	queries   map[string]string
}

func NewParser() *Parser {
	return &Parser{
		languages: map[string]*sitter.Language{
			".py": sitter.NewLanguage(tree_sitter_python.Language()),
			".rs": sitter.NewLanguage(tree_sitter_rust.Language()),
			".ts": sitter.NewLanguage(tree_sitter_typescript.LanguageTypescript()),
		},
		queries: map[string]string{
			".py": python.SymbolsQuery,
		},
	}
}

func (p *Parser) ParseFile(ctx context.Context, path string) ([]*graph.Node, []*graph.Edge, error) {
	ext := filepath.Ext(path)
	lang, ok := p.languages[ext]
	if !ok {
		return nil, nil, fmt.Errorf("unsupported language: %s", ext)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	tree := parser.Parse(content, nil)
	if tree == nil {
		return nil, nil, fmt.Errorf("failed to parse %s", path)
	}

	var nodes []*graph.Node
	var edges []*graph.Edge

	relPath, _ := filepath.Rel(".", path)
	moduleID := fmt.Sprintf("file:%s", relPath)
	nodes = append(nodes, &graph.Node{
		ID:   moduleID,
		Type: graph.NodeModule,
		Name: filepath.Base(path),
		File: relPath,
	})

	queryStr, ok := p.queries[ext]
	if !ok {
		return nodes, edges, nil
	}

	fmt.Printf("Creating query for %s with lang %p and query:\n%s\n", ext, lang, queryStr)
	query, err := sitter.NewQuery(lang, queryStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create query: %w", err)
	}
	if query == nil {
		return nil, nil, fmt.Errorf("failed to create query: query is nil but error is nil")
	}

	cursor := sitter.NewQueryCursor()
	captures := cursor.Captures(query, tree.RootNode(), content)
	captureNames := query.CaptureNames()

	for {
		match, _ := captures.Next()
		if match == nil {
			break
		}

		for _, capture := range match.Captures {
			captureName := captureNames[capture.Index]
			if strings.HasSuffix(captureName, ".name") {
				continue
			}

			nodeName := ""
			// Find the .name capture in the same match
			for _, c := range match.Captures {
				cn := captureNames[c.Index]
				if cn == strings.Split(captureName, ".")[0]+".name" {
					nodeName = string(content[c.Node.StartByte():c.Node.EndByte()])
					break
				}
			}

			if nodeName == "" {
				continue
			}

			nodeType := graph.NodeFunction
			prefix := "func"
			if strings.HasPrefix(captureName, "class") {
				nodeType = graph.NodeStruct
				prefix = "type"
			}

			id := fmt.Sprintf("%s:%s:%s", prefix, relPath, nodeName)
			nodes = append(nodes, &graph.Node{
				ID:   id,
				Type: nodeType,
				Name: nodeName,
				File: relPath,
				Line: int(capture.Node.StartPosition().Row) + 1,
			})

			edges = append(edges, &graph.Edge{
				FromID: id,
				ToID:   moduleID,
				Type:   graph.EdgeBelongsTo,
			})
		}
	}

	return nodes, edges, nil
}

func (p *Parser) ExtractCalls(ctx context.Context, path string) ([]*graph.Edge, error) {
	return nil, nil
}

func (p *Parser) ExtractControlFlow(ctx context.Context, path string) ([]*graph.Edge, error) {
	return nil, nil
}
