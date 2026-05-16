package mcp

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	aictx "github.com/andev0x/ctxd/internal/ai/context"
	"github.com/andev0x/ctxd/internal/architecture"
	graph "github.com/andev0x/ctxd/internal/graph/contracts"
	"github.com/andev0x/ctxd/internal/storage/contracts"
	"github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
)

type Server struct {
	store    contracts.Store
	compiler *aictx.Compiler
}

func NewServer(store contracts.Store) *Server {
	return &Server{
		store:    store,
		compiler: aictx.NewCompiler(store),
	}
}

func (s *Server) Start() error {
	mcpServer := mcp_golang.NewServer(stdio.NewStdioServerTransport(), mcp_golang.WithName("ctxd"))

	// Tool: get_symbol_context
	err := mcpServer.RegisterTool("get_symbol_context", "Generates AI-optimized markdown context for a given symbol ID", func(ctx context.Context, args struct {
		SymbolID string `json:"symbol_id" jsonschema:"description=The unique ID of the symbol (e.g. func:path:name)"`
	}) (*mcp_golang.ToolResponse, error) {
		content, err := s.compiler.Compile(ctx, args.SymbolID)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(content)), nil
	})
	if err != nil {
		return err
	}

	// Tool: find_neighbors
	err = mcpServer.RegisterTool("find_neighbors", "Finds symbols directly related to a given symbol ID", func(ctx context.Context, args struct {
		SymbolID string `json:"symbol_id" jsonschema:"description=The unique ID of the symbol"`
	}) (*mcp_golang.ToolResponse, error) {
		nodes, edges, err := s.store.GetNeighbors(ctx, args.SymbolID)
		if err != nil {
			return nil, err
		}
		if len(nodes) == 0 {
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("No neighbors found for %s", args.SymbolID))), nil
		}

		res := fmt.Sprintf("Neighbors of %s:\n", args.SymbolID)
		for i, n := range nodes {
			e := edges[i]
			res += fmt.Sprintf("- [%s] %s (%s) at %s:%d\n", e.Type, n.Name, n.Type, n.File, n.Line)
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(res)), nil
	})
	if err != nil {
		return err
	}

	// Tool: trace_calls
	err = mcpServer.RegisterTool("trace_calls", "Traces the call graph starting from a given symbol ID", func(ctx context.Context, args struct {
		SymbolID string `json:"symbol_id" jsonschema:"description=The unique ID of the symbol"`
		Depth    int    `json:"depth" jsonschema:"description=Maximum depth to trace,default=3"`
	}) (*mcp_golang.ToolResponse, error) {
		if args.Depth <= 0 {
			args.Depth = 3
		}
		var res string
		err := s.trace(ctx, args.SymbolID, 0, args.Depth, make(map[string]bool), &res)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(res)), nil
	})
	if err != nil {
		return err
	}

	// Tool: trace_execution_path
	err = mcpServer.RegisterTool("trace_execution_path", "Returns control-flow ordered calls for a symbol", func(ctx context.Context, args struct {
		SymbolID string `json:"symbol_id" jsonschema:"description=The unique ID of the symbol"`
	}) (*mcp_golang.ToolResponse, error) {
		res, err := s.flow(ctx, args.SymbolID)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(res)), nil
	})
	if err != nil {
		return err
	}

	// Tool: find_architecture_violations
	err = mcpServer.RegisterTool("find_architecture_violations", "Finds architecture boundary violations using a config file", func(ctx context.Context, args struct {
		ConfigPath string `json:"config_path" jsonschema:"description=Path to architecture YAML config"`
	}) (*mcp_golang.ToolResponse, error) {
		configPath := args.ConfigPath
		if configPath == "" {
			configPath = filepath.Join(".ctxd", "architecture.yaml")
		}
		cfg, err := architecture.LoadConfig(configPath)
		if err != nil {
			return nil, err
		}
		violations, err := architecture.FindViolations(ctx, s.store, cfg)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(formatViolations(violations))), nil
	})
	if err != nil {
		return err
	}

	return mcpServer.Serve()
}

func (s *Server) trace(ctx context.Context, id string, depth, maxDepth int, visited map[string]bool, res *string) error {
	if depth > maxDepth || visited[id] {
		return nil
	}
	visited[id] = true

	nodes, edges, err := s.store.GetNeighbors(ctx, id)
	if err != nil {
		return err
	}

	for i, n := range nodes {
		e := edges[i]
		if e.Type == graph.EdgeCalls {
			*res += fmt.Sprintf("%s-> [%s] %s (%s)\n", repeat("  ", depth), e.Type, n.Name, n.Type)
			if err := s.trace(ctx, n.ID, depth+1, maxDepth, visited, res); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Server) flow(ctx context.Context, id string) (string, error) {
	nodes, edges, err := s.store.GetNeighbors(ctx, id)
	if err != nil {
		return "", err
	}

	type entry struct {
		edge *graph.Edge
		node *graph.Node
	}

	var entries []entry
	for i, n := range nodes {
		e := edges[i]
		if e.Type == graph.EdgeFlowsThrough {
			entries = append(entries, entry{edge: e, node: n})
		}
	}
	if len(entries) == 0 {
		return fmt.Sprintf("No control flow entries found for %s", id), nil
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].edge.Sequence < entries[j].edge.Sequence
	})

	res := fmt.Sprintf("Control flow for %s:\n", id)
	for _, entry := range entries {
		contextLabel := ""
		if entry.edge.Metadata != nil {
			if ctx, ok := entry.edge.Metadata["context"].(string); ok && ctx != "" {
				contextLabel = fmt.Sprintf(" [%s]", ctx)
			}
		}
		res += fmt.Sprintf("%d. %s (%s)%s\n", entry.edge.Sequence, entry.node.Name, entry.node.Type, contextLabel)
	}
	return res, nil
}

func formatViolations(violations []architecture.Violation) string {
	if len(violations) == 0 {
		return "No architecture violations found."
	}

	sort.Slice(violations, func(i, j int) bool {
		if violations[i].FromLayer == violations[j].FromLayer {
			return violations[i].ToLayer < violations[j].ToLayer
		}
		return violations[i].FromLayer < violations[j].FromLayer
	})

	res := fmt.Sprintf("Architecture violations (%d):\n", len(violations))
	for _, v := range violations {
		res += fmt.Sprintf("- [%s] %s (%s) -> %s (%s)\n  at %s:%d -> %s:%d\n", v.EdgeType, v.FromID, v.FromLayer, v.ToID, v.ToLayer, v.FromFile, v.FromLine, v.ToFile, v.ToLine)
	}
	return res
}

func repeat(s string, n int) string {
	res := ""
	for i := 0; i < n; i++ {
		res += s
	}
	return res
}
