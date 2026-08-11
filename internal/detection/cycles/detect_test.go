package cycles

import (
	"reflect"
	"strings"
	"testing"
)

func TestDetectsAllRequestedViolations(t *testing.T) {
	graph := Graph{
		Nodes: []Node{
			{ID: "billing://entity/order", Kind: Entity, Namespace: "billing", Name: "Order"},
			{ID: "billing://entity/duplicate", Kind: Entity, Namespace: "billing", Name: "Order"},
			{ID: "billing://activity/pay", Kind: Activity, Namespace: "billing", Name: "PayOrder"},
		},
		Edges: []Edge{
			{Subject: "billing://activity/pay", Predicate: Used, Object: "billing://entity/order"},
			{Subject: "billing://entity/order", Predicate: WasGeneratedBy, Object: "billing://activity/pay"},
			{Subject: "billing://entity/order", Predicate: Used, Object: "billing://entity/duplicate"},
			{Subject: "billing://activity/pay", Predicate: Used, Object: "billing://entity/missing"},
		},
	}

	diagnostics := Detect(graph)
	for _, code := range []Code{CycleDetected, IllegalRelationDirection, UnresolvedStableID, NamespaceCollision} {
		if !diagnostics.Has(code) {
			t.Fatalf("missing %q in %#v", code, diagnostics)
		}
	}
	if len(diagnostics) != 4 {
		t.Fatalf("unexpected diagnostic count: got %d, diagnostics=%v", len(diagnostics), diagnostics)
	}
	if err := Check(Graph{Nodes: graph.Nodes[:1]}); err != nil {
		t.Fatalf("valid graph returned an error: %v", err)
	}
}

func TestDiagnosticsAreIndependentOfInsertionOrder(t *testing.T) {
	first := Graph{
		Nodes: []Node{
			{ID: "urn:entity:order", Kind: Entity, Namespace: "billing", Name: "Order"},
			{ID: "urn:entity:payment", Kind: Entity, Namespace: "billing", Name: "Payment"},
			{ID: "urn:activity:pay", Kind: Activity, Namespace: "billing", Name: "Pay"},
		},
		Edges: []Edge{
			{Subject: "urn:entity:order", Predicate: Used, Object: "urn:entity:payment"},
			{Subject: "urn:activity:pay", Predicate: Used, Object: "urn:missing:entity"},
			{Subject: "urn:entity:order", Predicate: WasGeneratedBy, Object: "urn:activity:pay"},
		},
	}
	second := Graph{
		Nodes: []Node{first.Nodes[2], first.Nodes[1], first.Nodes[0]},
		Edges: []Edge{first.Edges[2], first.Edges[1], first.Edges[0]},
	}

	left, right := Detect(first), Detect(second)
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("diagnostics changed with insertion order:\nleft=%#v\nright=%#v", left, right)
	}
	if Detect(first).Error() != Detect(second).Error() {
		t.Fatal("formatted diagnostics changed with insertion order")
	}
}

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
