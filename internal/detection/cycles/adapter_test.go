package cycles

import (
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestSemanticGraphAdapterPreservesValidFacts(t *testing.T) {
	source := semantic.NewGraph()
	order := newSemanticNode(t, &source, semantic.Entity, "urn:billing:order", "billing", "Order")
	pay := newSemanticNode(t, &source, semantic.Activity, "urn:billing:pay", "billing", "PayOrder")
	if err := source.AddFact(semantic.NewUsedFact(pay.ID, order.ID)); err != nil {
		t.Fatal(err)
	}
	beforeNodes, beforeFacts := source.SortedNodes(), source.AllFacts()
	if diagnostics := DetectSemanticGraph(source); len(diagnostics) != 0 {
		t.Fatalf("valid semantic graph was rejected: %v", diagnostics)
	}
	if !reflect.DeepEqual(beforeNodes, source.SortedNodes()) || !reflect.DeepEqual(beforeFacts, source.AllFacts()) {
		t.Fatal("semantic graph adapter mutated graph authority")
	}
}

func TestDetectorFindsMissingAndIllegalFacts(t *testing.T) {
	source := semantic.NewGraph()
	order := newSemanticNode(t, &source, semantic.Entity, "urn:billing:order", "billing", "Order")
	pay := newSemanticNode(t, &source, semantic.Activity, "urn:billing:pay", "billing", "PayOrder")
	if diagnostics := DetectSemanticGraph(source); len(diagnostics) != 0 {
		t.Fatalf("valid semantic graph was rejected: %v", diagnostics)
	}
	raw := Graph{
		Nodes: []Node{
			{ID: order.ID.String(), Kind: Entity, Namespace: "billing", Name: "Order"},
			{ID: pay.ID.String(), Kind: Activity, Namespace: "billing", Name: "PayOrder"},
		},
		Edges: []Edge{
			{Subject: order.ID.String(), Predicate: Used, Object: pay.ID.String()},
			{Subject: pay.ID.String(), Predicate: Used, Object: "urn:billing:missing"},
		},
	}
	diagnostics := Detect(raw)
	if !diagnostics.Has(IllegalRelationDirection) || !diagnostics.Has(UnresolvedStableID) {
		t.Fatalf("adapter lost structural violations: %v", diagnostics)
	}
}

func newSemanticNode(t *testing.T, graph *semantic.Graph, kind semantic.Kind, id, namespace, name string) semantic.Node {
	t.Helper()
	node, err := semantic.NewNodeFromStrings(kind, id, namespace, name)
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.AddNode(node); err != nil {
		t.Fatal(err)
	}
	return node
}
