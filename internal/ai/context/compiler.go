package context

import (
	"context"
	"fmt"
	"strings"

	graph "github.com/PizenLabs/lea/internal/graph/contracts"
	"github.com/PizenLabs/lea/internal/storage/contracts"
)

type Compiler struct {
	store contracts.Store
}

func NewCompiler(store contracts.Store) *Compiler {
	return &Compiler{store: store}
}

func (c *Compiler) Compile(ctx context.Context, symbolID string) (string, error) {
	node, err := c.store.GetNode(ctx, symbolID)
	if err != nil {
		return "", err
	}
	if node == nil {
		return "", fmt.Errorf("symbol not found: %s", symbolID)
	}

	var sb strings.Builder

	// Header
	fmt.Fprintf(&sb, "## %s\n\n", node.Name)
	fmt.Fprintf(&sb, "Type: %s\n", node.Type)
	fmt.Fprintf(&sb, "File: %s\n\n", node.File)

	// Outbound Dependencies (Uses/Calls)
	outNodes, outEdges, err := c.store.GetNeighbors(ctx, symbolID)
	if err == nil && len(outNodes) > 0 {
		sb.WriteString("### Dependencies\n")
		for i, n := range outNodes {
			e := outEdges[i]
			if e.Type == graph.EdgeCalls || e.Type == graph.EdgeUses || e.Type == graph.EdgeBelongsTo {
				fmt.Fprintf(&sb, "- [%s] %s (%s)\n", e.Type, n.Name, n.Type)
			}
		}
		sb.WriteString("\n")
	}

	// Inbound Dependencies (Called by/Used by)
	inNodes, inEdges, err := c.store.GetInboundEdges(ctx, symbolID)
	if err == nil && len(inNodes) > 0 {
		sb.WriteString("### Relationships\n")
		for i, n := range inNodes {
			e := inEdges[i]
			fmt.Fprintf(&sb, "- %s (%s) [%s]\n", n.Name, n.Type, e.Type)
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}
