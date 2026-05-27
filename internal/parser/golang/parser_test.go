package golang

import (
	"context"
	"path/filepath"
	"testing"

	graph "github.com/ax86-labs/lea/internal/graph/contracts"
)

func TestParseFile(t *testing.T) {
	p := NewParser()
	ctx := context.Background()
	path := "../../../testdata/golang/simple.go"
	absPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}

	nodes, edges, err := p.ParseFile(ctx, absPath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Verify nodes
	expectedNodes := map[string]graph.NodeType{
		"pkg:../../../testdata/golang":      graph.NodePackage,
		"type:../../../testdata/golang:Calculator": graph.NodeStruct,
		"method:../../../testdata/golang:Calculator.Add": graph.NodeMethod,
		"func:../../../testdata/golang:Add": graph.NodeFunction,
		"func:../../../testdata/golang:Main": graph.NodeFunction,
	}

	// Adjust pkgPath expectation based on how it's calculated in ParseFile
	pkgPath := filepath.Dir(absPath)
	expectedNodes = map[string]graph.NodeType{
		"pkg:" + pkgPath:                         graph.NodePackage,
		"type:" + pkgPath + ":Calculator":        graph.NodeStruct,
		"method:" + pkgPath + ":Calculator.Add":  graph.NodeMethod,
		"func:" + pkgPath + ":Add":               graph.NodeFunction,
		"func:" + pkgPath + ":Main":              graph.NodeFunction,
	}

	if len(nodes) != len(expectedNodes) {
		t.Errorf("Expected %d nodes, got %d", len(expectedNodes), len(nodes))
		for _, n := range nodes {
			t.Logf("Found node: %s (%s)", n.ID, n.Type)
		}
	}

	for id, nodeType := range expectedNodes {
		found := false
		for _, n := range nodes {
			if n.ID == id {
				found = true
				if n.Type != nodeType {
					t.Errorf("Node %s has wrong type: expected %s, got %s", id, nodeType, n.Type)
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected node %s not found", id)
		}
	}

	// Verify edges
	// Expect BELONGS_TO edges
	for _, e := range edges {
		if e.Type != graph.EdgeBelongsTo {
			t.Errorf("Unexpected edge type: %s", e.Type)
		}
	}
}

func TestExtractCalls(t *testing.T) {
	p := NewParser()
	ctx := context.Background()
	path := "../../../testdata/golang/simple.go"
	absPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}

	edges, err := p.ExtractCalls(ctx, absPath)
	if err != nil {
		t.Fatalf("ExtractCalls failed: %v", err)
	}

	pkgPath := filepath.Dir(absPath)
	expectedEdges := []struct {
		from string
		to   string
	}{
		{from: "method:" + pkgPath + ":Calculator.Add", to: "unknown:fmt.Println"},
		{from: "func:" + pkgPath + ":Main", to: "func:" + pkgPath + ":Add"},
	}

	for _, ee := range expectedEdges {
		found := false
		for _, e := range edges {
			if e.FromID == ee.from && e.ToID == ee.to {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected edge from %s to %s not found", ee.from, ee.to)
		}
	}
}

func TestExtractControlFlow(t *testing.T) {
	p := NewParser()
	ctx := context.Background()
	path := "../../../testdata/golang/simple.go"
	absPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}

	edges, err := p.ExtractControlFlow(ctx, absPath)
	if err != nil {
		t.Fatalf("ExtractControlFlow failed: %v", err)
	}

	if len(edges) == 0 {
		t.Error("No control flow edges found")
	}

	// In Main, we expect calc.Add(5) then Add(1, 2) then fmt.Println(res)
	// Note: calc.Add(5) might be tricky if not resolved.
	// Current implementation: calc.Add is a SelectorExpr, becomes unknown:calc.Add

	pkgPath := filepath.Dir(absPath)
	mainID := "func:" + pkgPath + ":Main"

	foundAdd := false
	foundInternalAdd := false
	for _, e := range edges {
		if e.FromID == mainID {
			if e.ToID == "unknown:calc.Add" {
				foundAdd = true
			}
			if e.ToID == "func:" + pkgPath + ":Add" {
				foundInternalAdd = true
			}
		}
	}

	if !foundAdd {
		t.Errorf("Expected call to calc.Add in Main flow not found")
	}
	if !foundInternalAdd {
		t.Errorf("Expected call to Add in Main flow not found")
	}
}
