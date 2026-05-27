package sqlite

import (
	"context"
	"testing"

	graph "github.com/ax86-labs/lea/internal/graph/contracts"
)

func TestStore(t *testing.T) {
	ctx := context.Background()
	// Use in-memory database for testing
	s, err := NewStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	// Test SaveNode
	node := &graph.Node{
		ID:   "test-node",
		Type: graph.NodeFunction,
		Name: "TestNode",
		File: "test.go",
		Line: 10,
		Metadata: map[string]interface{}{
			"foo": "bar",
		},
	}
	if err := s.SaveNode(ctx, node); err != nil {
		t.Fatalf("SaveNode failed: %v", err)
	}

	// Test GetNode
	gotNode, err := s.GetNode(ctx, "test-node")
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}
	if gotNode == nil {
		t.Fatal("Node not found")
	}
	if gotNode.Name != "TestNode" {
		t.Errorf("Expected Name TestNode, got %s", gotNode.Name)
	}
	if gotNode.Metadata["foo"] != "bar" {
		t.Errorf("Expected metadata foo=bar, got %v", gotNode.Metadata["foo"])
	}

	// Test SaveEdge
	node2 := &graph.Node{
		ID:   "test-node-2",
		Type: graph.NodeFunction,
		Name: "TestNode2",
		File: "test.go",
		Line: 20,
	}
	if err := s.SaveNode(ctx, node2); err != nil {
		t.Fatalf("SaveNode failed: %v", err)
	}

	edge := &graph.Edge{
		FromID: "test-node",
		ToID:   "test-node-2",
		Type:   graph.EdgeCalls,
	}
	if err := s.SaveEdge(ctx, edge); err != nil {
		t.Fatalf("SaveEdge failed: %v", err)
	}

	// Test GetNeighbors
	nodes, edges, err := s.GetNeighbors(ctx, "test-node")
	if err != nil {
		t.Fatalf("GetNeighbors failed: %v", err)
	}
	if len(edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(edges))
	}
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodes))
	}
	if nodes[0].ID != "test-node-2" {
		t.Errorf("Expected neighbor ID test-node-2, got %s", nodes[0].ID)
	}

	// Test DeleteByFile
	if err := s.DeleteByFile(ctx, "test.go"); err != nil {
		t.Fatalf("DeleteByFile failed: %v", err)
	}
	nodes, err = s.ListNodes(ctx)
	if err != nil {
		t.Fatalf("ListNodes failed: %v", err)
	}
	if len(nodes) != 0 {
		t.Errorf("Expected 0 nodes after DeleteByFile, got %d", len(nodes))
	}
}
