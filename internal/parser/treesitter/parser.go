package treesitter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	graph "github.com/andev0x/ctxd/internal/graph/contracts"
	sitter "github.com/tree-sitter/go-tree-sitter"
	"github.com/tree-sitter/tree-sitter-python/bindings/go"
	"github.com/tree-sitter/tree-sitter-rust/bindings/go"
	"github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

type Parser struct {
	languages map[string]*sitter.Language
}

func NewParser() *Parser {
	return &Parser{
		languages: map[string]*sitter.Language{
			".py": sitter.NewLanguage(tree_sitter_python.Language()),
			".rs": sitter.NewLanguage(tree_sitter_rust.Language()),
			".ts": sitter.NewLanguage(tree_sitter_typescript.LanguageTypescript()),
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

	// This is a placeholder for actual Tree-sitter query logic
	// In a real implementation, we would use tree-sitter queries to extract symbols
	// For Phase 5 demonstration, we'll return a package/module node
	var nodes []*graph.Node
	var edges []*graph.Edge

	relPath, _ := filepath.Rel(".", path)
	nodes = append(nodes, &graph.Node{
		ID:   fmt.Sprintf("file:%s", relPath),
		Type: graph.NodeModule,
		Name: filepath.Base(path),
		File: relPath,
	})

	return nodes, edges, nil
}

func (p *Parser) ExtractCalls(ctx context.Context, path string) ([]*graph.Edge, error) {
	return nil, nil
}
