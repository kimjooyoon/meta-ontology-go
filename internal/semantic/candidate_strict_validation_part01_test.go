package semantic

import (
	"testing"
)

func TestDuplicateCandidateFactsArePermutationStable(t *testing.T) {
	ns := Namespace("candidate-duplicates")
	activity := mustActivity(t, MustIdentity("candidate-duplicates://activity/run"), ns, "Run")
	entity := mustEntity(t, MustIdentity("candidate-duplicates://entity/output"), ns, "Output")
	first := NewCandidateFact(activity.ID, Used, entity.ID, "z observation").WithSpan(Span{
		File: "z.gooo", Start: Position{Offset: 20}, End: Position{Offset: 24},
	})
	second := NewCandidateFact(activity.ID, Used, entity.ID, "a observation").WithSpan(Span{
		File: "a.gooo", Start: Position{Offset: 2}, End: Position{Offset: 6},
	})

	build := func(facts ...Fact) Graph {
		graph := NewGraph()
		for _, node := range []Node{activity, entity} {
			if err := graph.AddNode(node); err != nil {
				t.Fatal(err)
			}
		}
		for _, fact := range facts {
			if err := graph.AddCandidate(fact); err != nil {
				t.Fatal(err)
			}
		}
		return graph
	}

	left := build(first, second, first)
	right := build(second, first, second)
	if len(left.Candidates()) != 1 || len(right.Candidates()) != 1 {
		t.Fatalf("duplicate candidates were not collapsed: left=%d right=%d", len(left.Candidates()), len(right.Candidates()))
	}
	if left.Canonical() != right.Canonical() || left.StableHash() != right.StableHash() {
		t.Fatal("duplicate candidate canonical/hash changed with insertion permutation")
	}
	if left.Candidates()[0].Reason != "a observation" || left.Candidates()[0].Span.File != "a.gooo" {
		t.Fatalf("duplicate candidate merge was not deterministic: %#v", left.Candidates()[0])
	}
}
