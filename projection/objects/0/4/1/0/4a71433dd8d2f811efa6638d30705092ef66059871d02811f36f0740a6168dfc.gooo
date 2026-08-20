package semantic

import (
	"errors"
	"testing"
)

func TestCandidateDirectionMismatchIsRejectedAtomically(t *testing.T) {
	ns := Namespace("deferred")
	activity := mustActivity(t, MustIdentity("deferred://activity/run"), ns, "Run")
	entity := mustEntity(t, MustIdentity("deferred://entity/input"), ns, "Input")
	output := mustEntity(t, MustIdentity("deferred://entity/output"), ns, "Output")
	agent := mustAgent(t, MustIdentity("deferred://agent/verifier"), ns, "Verifier")
	cases := []struct {
		name      string
		subject   ID
		predicate Relation
		object    ID
	}{
		{name: "used reversed", subject: entity.ID, predicate: Used, object: activity.ID},
		{name: "generated reversed", subject: activity.ID, predicate: WasGeneratedBy, object: output.ID},
		{name: "derived reversed", subject: activity.ID, predicate: WasDerivedFrom, object: output.ID},
		{name: "associated reversed", subject: agent.ID, predicate: WasAssociatedWith, object: activity.ID},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			graph := NewGraph()
			for _, node := range []Node{activity, entity, output, agent} {
				if err := graph.AddNode(node); err != nil {
					t.Fatal(err)
				}
			}
			fact := NewCandidateFact(test.subject, test.predicate, test.object, "untyped observation")
			beforeCanonical, beforeHash := graph.Canonical(), graph.StableHash()
			if err := graph.AddCandidate(fact); !errors.Is(err, ErrInvalidFact) {
				t.Fatalf("candidate direction error = %v, want ErrInvalidFact", err)
			}
			if graph.HasCandidate(fact.Key()) || graph.HasFact(fact.Key()) {
				t.Fatal("rejected candidate direction crossed the graph boundary")
			}
			if graph.Canonical() != beforeCanonical || graph.StableHash() != beforeHash {
				t.Fatal("rejected candidate direction mutated the graph")
			}
		})
	}
}
func TestMissingCandidatePromotionDoesNotInitializeOrMutateGraph(t *testing.T) {
	graph := Graph{}
	key := FactKey{
		Subject: MustIdentity("deferred://activity/missing"), Predicate: Used,
		Object: MustIdentity("deferred://entity/missing"),
	}
	beforeCanonical, beforeHash := graph.Canonical(), graph.StableHash()
	if _, err := graph.PromoteCandidate(key); !errors.Is(err, ErrCandidateNotFound) {
		t.Fatalf("missing candidate error = %v, want ErrCandidateNotFound", err)
	}
	if graph.Canonical() != beforeCanonical || graph.StableHash() != beforeHash {
		t.Fatal("missing candidate promotion changed canonical/hash")
	}
	if graph.nodes != nil || graph.names != nil || graph.facts != nil || graph.candidates != nil {
		t.Fatal("missing candidate promotion initialized graph storage")
	}
}
