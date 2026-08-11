package semantic

import (
	"errors"
	"testing"
)

func TestCandidateDirectionMismatchRemainsNonAuthoritative(t *testing.T) {
	ns := Namespace("deferred")
	activity := mustActivity(t, MustIdentity("deferred://activity/run"), ns, "Run")
	entity := mustEntity(t, MustIdentity("deferred://entity/input"), ns, "Input")
	graph := NewGraph()
	for _, node := range []Node{activity, entity} {
		if err := graph.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	fact := NewCandidateFact(entity.ID, Used, activity.ID, "untyped observation")
	if err := graph.AddCandidate(fact); err != nil {
		t.Fatal(err)
	}
	if !graph.HasCandidate(fact.Key()) || graph.HasFact(fact.Key()) {
		t.Fatal("candidate direction mismatch crossed the authoritative boundary")
	}
	if err := graph.Validate(); !errors.Is(err, ErrGraphInvalid) {
		t.Fatalf("invalid candidate direction was not fail-closed: %v", err)
	}
	if _, err := graph.PromoteCandidate(fact.Key()); !errors.Is(err, ErrInvalidFact) {
		t.Fatalf("invalid candidate promotion error = %v, want ErrInvalidFact", err)
	}
}

func TestQualifiedPROVCounterpartIsDeferredAndRejected(t *testing.T) {
	relation := Relation("wasGeneratedBy#qualified")
	if relation.Valid() {
		t.Fatal("qualified relation became a bare relation implicitly")
	}
	graph := NewGraph()
	fact := NewFact(
		MustIdentity("deferred://entity/output"), relation,
		MustIdentity("deferred://activity/run"),
	)
	if err := graph.AddFact(fact); !errors.Is(err, ErrUnknownRelation) {
		t.Fatalf("qualified relation error = %v, want ErrUnknownRelation", err)
	}
	if len(graph.AllFacts()) != 0 {
		t.Fatal("deferred qualified relation mutated graph")
	}
	t.Log("DEFERRED: semantic-ir/v1 has no qualified relation/event identity schema")
}

func TestIdentityReplacementIsNotImplicitlyEquivalent(t *testing.T) {
	ns := Namespace("deferred")
	oldID := MustIdentity("deferred://entity/old")
	newID := MustIdentity("deferred://entity/new")
	graph := NewGraph()
	if err := graph.AddNode(mustEntity(t, oldID, ns, "Record")); err != nil {
		t.Fatal(err)
	}
	before := graph.StableHash()
	if err := graph.AddNode(mustEntity(t, newID, ns, "RecordV2")); err != nil {
		t.Fatal(err)
	}
	if graph.StableHash() == before {
		t.Fatal("new ID was silently treated as an authorized rekey")
	}
	if _, ok := graph.Node(oldID); !ok {
		t.Fatal("old ID disappeared without a continuity contract")
	}
	if _, ok := graph.Node(newID); !ok {
		t.Fatal("new ID was not retained as a distinct declaration")
	}
	t.Log("DEFERRED: ID continuity/rekey authorization requires a future delta contract")
}
