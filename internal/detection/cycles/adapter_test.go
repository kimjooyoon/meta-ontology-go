package cycles

import (
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
	if diagnostics := DetectSemanticGraph(source); len(diagnostics) != 0 {
		t.Fatalf("valid semantic graph was rejected: %v", diagnostics)
	}
}

func TestSemanticGraphAdapterFindsMissingAndIllegalFacts(t *testing.T) {
	source := semantic.NewGraph()
	order := newSemanticNode(t, &source, semantic.Entity, "urn:billing:order", "billing", "Order")
	pay := newSemanticNode(t, &source, semantic.Activity, "urn:billing:pay", "billing", "PayOrder")
	if err := source.AddFact(semantic.NewFact(order.ID, semantic.Used, pay.ID)); err != nil {
		t.Fatal(err)
	}
	if err := source.AddFact(semantic.NewUsedFact(pay.ID, "urn:billing:missing")); err != nil {
		t.Fatal(err)
	}
	diagnostics := DetectSemanticGraph(source)
	if !diagnostics.Has(IllegalRelationDirection) || !diagnostics.Has(UnresolvedStableID) {
		t.Fatalf("adapter lost structural violations: %v", diagnostics)
	}
}

func TestHostContractDoesNotClaimFutureImplementation(t *testing.T) {
	contract := CurrentHostContract()
	if err := contract.Validate(); err != nil {
		t.Fatal(err)
	}
	if contract.GoooHosted.Status != StagePending {
		t.Fatalf("future stage was reported as implemented: %#v", contract.GoooHosted)
	}
	futureObserved := NewEvidence(GoooHostedStage, "gooo-hosted", nil)
	contract.GoooHosted = futureObserved
	if err := contract.Validate(); err == nil {
		t.Fatal("future implementation was accepted without a contract transition")
	}
}

func TestEvidenceIsComparableAcrossHosts(t *testing.T) {
	diagnostics := Diagnostics{{Code: UnresolvedStableID, Message: "missing"}}
	goEvidence := NewEvidence(GoHostedStage, "go", diagnostics)
	goooEvidence := NewEvidence(GoooHostedStage, "gooo", diagnostics)
	if !EquivalentEvidence(goEvidence, goooEvidence) {
		t.Fatal("equivalent diagnostics from two hosts did not compare equal")
	}
	if EquivalentEvidence(goEvidence, PendingEvidence(GoooHostedStage, "gooo")) {
		t.Fatal("pending future evidence was treated as equivalent")
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
