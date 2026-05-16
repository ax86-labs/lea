package contracts

import (
	"context"
	graph "github.com/andev0x/ctxd/internal/graph/contracts"
)

type Parser interface {
	ParseFile(ctx context.Context, path string) ([]*graph.Node, []*graph.Edge, error)
	ExtractCalls(ctx context.Context, path string) ([]*graph.Edge, error)
}
