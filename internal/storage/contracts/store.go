package contracts

import (
	"context"
	graph "github.com/andev0x/ctxd/internal/graph/contracts"
)

type Store interface {
	SaveNode(ctx context.Context, node *graph.Node) error
	SaveEdge(ctx context.Context, edge *graph.Edge) error
	GetNode(ctx context.Context, id string) (*graph.Node, error)
	ListNodes(ctx context.Context) ([]*graph.Node, error)
	GetNeighbors(ctx context.Context, id string) ([]*graph.Node, []*graph.Edge, error)
	GetInboundEdges(ctx context.Context, id string) ([]*graph.Node, []*graph.Edge, error)
	DeleteNode(ctx context.Context, id string) error
	DeleteByFile(ctx context.Context, file string) error
	DeleteEdgesFrom(ctx context.Context, id string) error
	Close() error
}
