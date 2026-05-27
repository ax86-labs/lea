package contracts

import (
	"context"
	graph "github.com/PizenLabs/lea/internal/graph/contracts"
)

type Parser interface {
	ParseFile(ctx context.Context, path string) ([]*graph.Node, []*graph.Edge, error)
	ExtractCalls(ctx context.Context, path string) ([]*graph.Edge, error)
	ExtractControlFlow(ctx context.Context, path string) ([]*graph.Edge, error)
}
