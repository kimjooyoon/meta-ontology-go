package semanticdelta

import (
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

func TestSemanticIRAdapterApplyIgnoresCandidateWithoutCommit(t *testing.T) {
	before := semanticContractFixture(t, "Pay order", false)
	after := semanticContractFixture(t, "Pay order", false)
	candidate := semantic.NewCandidateFact(
		semantic.MustIdentity("billing://entity/order"), semantic.WasDerivedFrom,
		semantic.MustIdentity("billing://entity/order"), "candidate only",
	)
	if err := after.AddCandidate(candidate); err != nil {
		t.Fatal(err)
	}
	commits := 0
	report, err := semanticIRContractAdapter().Apply(
		before,
		after,
		Scope{Prefixes: []string{"billing://"}},
		nil,
	)
	if err != nil || !report.Passes() {
		t.Fatalf("candidate-only Apply = report %#v, error %v", report, err)
	}
	if commits != 0 {
		t.Fatalf("candidate-only change reached commit callback %d time(s)", commits)
	}
}

func semanticIRContractAdapter() Adapter[semantic.IR] {
	return Adapter[semantic.IR]{
		Nodes: func(ir semantic.IR) ([]Node, error) {
			normalized, err := ir.Normalized()
			if err != nil {
				return nil, err
			}
			nodes := make([]Node, 0, len(normalized.Graph.Nodes()))
			for _, node := range normalized.Graph.Nodes() {
				nodes = append(nodes, Node{ID: node.ID.String(), Kind: node.Kind.String()})
			}
			return nodes, nil
		},
		Facts: func(ir semantic.IR) ([]Fact, error) {
			normalized, err := ir.Normalized()
			if err != nil {
				return nil, err
			}
			facts := make([]Fact, 0, len(normalized.Graph.DeterministicFacts()))
			for _, fact := range normalized.Graph.DeterministicFacts() {
				facts = append(facts, Fact{
					Subject: fact.Subject.String(), Predicate: fact.Predicate.String(), Object: fact.Object.String(),
				})
			}
			return facts, nil
		},
	}
}

func semanticContractFixture(t *testing.T, activityName string, withReceipt bool) semantic.IR {
	t.Helper()
	ir := semantic.NewIR("semantic-delta-contract", semantic.Namespace("billing"))
	activityID := semantic.MustIdentity("billing://activity/pay-order")
	orderID := semantic.MustIdentity("billing://entity/order")
	activity, err := semantic.NewActivity(activityID, ir.Namespace, activityName)
	if err != nil {
		t.Fatal(err)
	}
	order, err := semantic.NewEntity(orderID, ir.Namespace, "Order")
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range []semantic.Node{activity, order} {
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if err := ir.AddFact(semantic.NewUsedFact(activityID, orderID)); err != nil {
		t.Fatal(err)
	}
	if !withReceipt {
		return ir
	}
	receiptID := semantic.MustIdentity("billing://entity/receipt")
	receipt, err := semantic.NewEntity(receiptID, ir.Namespace, "Receipt")
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(receipt); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddFact(semantic.NewUsedFact(activityID, receiptID)); err != nil {
		t.Fatal(err)
	}
	return ir
}
