package semanticdelta

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

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
