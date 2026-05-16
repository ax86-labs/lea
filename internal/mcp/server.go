package mcp

import (
	"context"
	"fmt"

	aictx "github.com/andev0x/ctxd/internal/ai/context"
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

func repeat(s string, n int) string {
	res := ""
	for i := 0; i < n; i++ {
		res += s
	}
	return res
}
