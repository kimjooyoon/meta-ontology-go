package cycles

import (
	"strings"
	"testing"
)

func TestNamespaceBoundaryAndAliases(t *testing.T) {
	valid := Graph{Nodes: []Node{
		{ID: "urn:billing:payment", Kind: Entity, Namespace: "billing", Name: "Payment"},
		{ID: "urn:fraud:payment", Kind: Entity, Namespace: "fraud", Name: "Payment"},
	}}
	if diagnostics := Detect(valid); len(diagnostics) != 0 {
		t.Fatalf("same display name crossed namespace boundary: %v", diagnostics)
	}

	collision := Graph{Nodes: []Node{
		{ID: "urn:billing:payment", Kind: Entity, Namespace: "billing", Name: "Payment", Aliases: []string{"Charge"}},
		{ID: "urn:billing:charge", Kind: Entity, Namespace: "billing", Name: "Charge"},
	}}
	diagnostics := Detect(collision)
	if !diagnostics.Has(NamespaceCollision) {
		t.Fatal("alias/name collision was not reported")
	}
}
func TestReportsInvalidStableIDs(t *testing.T) {
	graph := Graph{
		Nodes: []Node{{ID: "not-an-absolute-id", Kind: Entity, Namespace: "billing", Name: "Order"}},
		Edges: []Edge{{Subject: "", Predicate: Used, Object: "urn:entity:order"}},
	}
	diagnostics := Detect(graph)
	if !diagnostics.Has(InvalidStableID) {
		t.Fatalf("invalid IDs were accepted: %v", diagnostics)
	}
	if !strings.Contains(diagnostics.Error(), string(InvalidStableID)) {
		t.Fatalf("diagnostic error omitted code: %v", diagnostics)
	}
}
func TestCyclePathIsClosedSimpleAndBounded(t *testing.T) {
	graph := Graph{}
	for i := 0; i < 6; i++ {
		graph.Nodes = append(graph.Nodes, Node{
			ID: "urn:activity:" + string(rune('a'+i)), Kind: Activity,
			Namespace: "billing", Name: "Activity" + string(rune('A'+i)),
		})
	}
	for _, source := range graph.Nodes {
		for _, target := range graph.Nodes {
			if source.ID != target.ID {
				graph.Edges = append(graph.Edges, Edge{
					Subject: source.ID, Predicate: WasInformedBy, Object: target.ID,
				})
			}
		}
	}

	diagnostics := Detect(graph)
	if len(diagnostics) != 1 || diagnostics[0].Code != CycleDetected {
		t.Fatalf("dense SCC produced unexpected diagnostics: %v", diagnostics)
	}
	path := diagnostics[0].Cycle
	if len(path) < 2 || len(path) > len(graph.Nodes)+1 || path[0] != path[len(path)-1] {
		t.Fatalf("cycle path was not closed and bounded: %v", path)
	}
	seen := make(map[ID]struct{}, len(path)-1)
	for _, id := range path[:len(path)-1] {
		if _, exists := seen[id]; exists {
			t.Fatalf("cycle path repeated an intermediate node: %v", path)
		}
		seen[id] = struct{}{}
	}
}
