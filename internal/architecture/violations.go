package architecture

import (
	"context"

	graph "github.com/PizenLabs/lea/internal/graph/contracts"
	"github.com/PizenLabs/lea/internal/storage/contracts"
)

type Violation struct {
	FromID    string
	ToID      string
	FromLayer string
	ToLayer   string
	EdgeType  graph.EdgeType
	FromFile  string
	ToFile    string
	FromLine  int
	ToLine    int
}

func FindViolations(ctx context.Context, store contracts.Store, cfg *Config) ([]Violation, error) {
	nodes, err := store.ListNodes(ctx)
	if err != nil {
		return nil, err
	}
	edges, err := store.ListEdges(ctx)
	if err != nil {
		return nil, err
	}

	nodeMap := make(map[string]*graph.Node, len(nodes))
	for _, node := range nodes {
		nodeMap[node.ID] = node
	}

	matcher := NewMatcher(cfg)
	var violations []Violation

	for _, edge := range edges {
		if !isDependencyEdge(edge.Type) {
			continue
		}
		fromNode := nodeMap[edge.FromID]
		toNode := nodeMap[edge.ToID]
		if fromNode == nil || toNode == nil {
			if cfg.AllowUnknown() {
				continue
			}
			continue
		}

		fromLayer, okFrom := matcher.LayerForFile(fromNode.File)
		toLayer, okTo := matcher.LayerForFile(toNode.File)
		if !okFrom || !okTo {
			if cfg.AllowUnknown() {
				continue
			}
			continue
		}

		if cfg.Allowed(fromLayer, toLayer) {
			continue
		}

		violations = append(violations, Violation{
			FromID:    edge.FromID,
			ToID:      edge.ToID,
			FromLayer: fromLayer,
			ToLayer:   toLayer,
			EdgeType:  edge.Type,
			FromFile:  fromNode.File,
			ToFile:    toNode.File,
			FromLine:  fromNode.Line,
			ToLine:    toNode.Line,
		})
	}

	return violations, nil
}

func isDependencyEdge(edgeType graph.EdgeType) bool {
	switch edgeType {
	case graph.EdgeCalls, graph.EdgeUses, graph.EdgeDependsOn, graph.EdgeImports:
		return true
	default:
		return false
	}
}

func (c *Config) AllowUnknown() bool {
	if c.Settings.AllowUnknown == nil {
		return true
	}
	return *c.Settings.AllowUnknown
}

func (c *Config) Allowed(fromLayer, toLayer string) bool {
	if fromLayer == "" || toLayer == "" {
		return c.AllowUnknown()
	}
	if c.AllowSelf() && fromLayer == toLayer {
		return true
	}
	layer := c.layerByName(fromLayer)
	if layer == nil {
		return c.AllowUnknown()
	}
	if len(layer.Allow) == 0 {
		return c.DefaultAllowAll()
	}
	for _, allowed := range layer.Allow {
		if allowed == toLayer {
			return true
		}
	}
	return false
}

func (c *Config) AllowSelf() bool {
	if c.Settings.AllowSelf == nil {
		return true
	}
	return *c.Settings.AllowSelf
}

func (c *Config) DefaultAllowAll() bool {
	if c.Settings.DefaultAllowAll == nil {
		return true
	}
	return *c.Settings.DefaultAllowAll
}

func (c *Config) layerByName(name string) *Layer {
	for i := range c.Layers {
		if c.Layers[i].Name == name {
			return &c.Layers[i]
		}
	}
	return nil
}
