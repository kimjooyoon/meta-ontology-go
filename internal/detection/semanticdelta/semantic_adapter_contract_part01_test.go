package semanticdelta

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
)

func TestSemanticIRAdapterContractIgnoresPresentationAndCandidates(t *testing.T) {
	before := semanticContractFixture(t, "Pay order", false)
	after := semanticContractFixture(t, "Collect order", true)
	candidate := semantic.NewCandidateFact(
		semantic.MustIdentity("billing://entity/order"), semantic.WasDerivedFrom,
		semantic.MustIdentity("billing://entity/receipt"), "candidate only",
	)
	if err := after.AddCandidate(candidate); err != nil {
		t.Fatal(err)
	}

	delta, err := semanticIRContractAdapter().Diff(before, after)
	if err != nil {
		t.Fatal(err)
	}
	want := Delta{
		AddedNodes: []Node{{ID: "billing://entity/receipt", Kind: "Entity"}},
		AddedFacts: []Fact{{
			Subject: "billing://activity/pay-order", Predicate: "used", Object: "billing://entity/receipt",
		}},
	}
	if !reflect.DeepEqual(delta, want) {
		t.Fatalf("semantic IR delta = %#v, want %#v", delta, want)
	}
}
func TestSemanticIRAdapterClassifiesPresentationOnlyEditAsSyntaxOnly(t *testing.T) {
	before := semanticContractFixture(t, "Pay order", false)
	after := semanticContractFixture(t, "Collect order", false)

	delta, err := semanticIRContractAdapter().Diff(before, after)
	if err != nil {
		t.Fatal(err)
	}
	if !delta.IsEmpty() {
		t.Fatalf("display-only edit produced semantic delta: %#v", delta)
	}
}
func TestSemanticIRAdapterContractRejectsInvalidProjection(t *testing.T) {
	adapter := Adapter[semantic.IR]{
		Nodes: func(semantic.IR) ([]Node, error) {
			return []Node{{ID: "billing://entity/order"}}, nil
		},
		Facts: func(semantic.IR) ([]Fact, error) { return nil, nil },
	}
	if _, err := adapter.Snapshot(semantic.IR{}); err == nil {
		t.Fatal("adapter accepted a node without a semantic kind")
	}
}
