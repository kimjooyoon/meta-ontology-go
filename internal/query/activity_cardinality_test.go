package query

import "testing"

func TestResolveActivityCardinalityStatesAndPreservesGraph(t *testing.T) {
	graph := New()
	for _, node := range []Node{
		{ID: ID("urn:activity:z"), Kind: ActivityNodeKind, Namespace: "two", Name: "Run"},
		{ID: ID("urn:activity:a"), Kind: ActivityNodeKind, Namespace: "one", Name: "Run"},
		{ID: ID("urn:entity:a"), Kind: EntityNodeKind, Namespace: "one", Name: "Record"},
	} {
		if err := graph.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	before := graph.StableHash()
	one, err := graph.ResolveActivityCardinality(ActivitySelector{Namespace: "one", Name: "Run"})
	if err != nil || one.Decision != ActivityResolutionClosed || one.Occurrences != 1 || one.Claim.ProofChoice != "COHERENCE" {
		t.Fatalf("one resolution = %#v, err=%v", one, err)
	}
	missing, err := graph.ResolveActivityCardinality(ActivitySelector{Name: "Missing"})
	if err != nil || missing.Decision != ActivityResolutionUnknown || missing.Occurrences != 0 || missing.Claim.UnknownClass != "DIRECT_MISSING" {
		t.Fatalf("missing resolution = %#v, err=%v", missing, err)
	}
	many, err := graph.ResolveActivityCardinality(ActivitySelector{Name: "Run"})
	if err != nil || many.Decision != ActivityResolutionRefuted || many.Occurrences != 2 || many.Claim.Reason != "AMBIGUOUS_ACTIVITY_BINDING" {
		t.Fatalf("many resolution = %#v, err=%v", many, err)
	}
	if many.Matches[0].ID != "urn:activity:a" || many.Matches[1].ID != "urn:activity:z" {
		t.Fatalf("matches are not stable: %#v", many.Matches)
	}
	if graph.StableHash() != before {
		t.Fatal("activity resolution mutated graph authority")
	}
}

func TestResolveActivityCardinalityRejectsEmptySelector(t *testing.T) {
	if _, err := New().ResolveActivityCardinality(ActivitySelector{}); !errorsIs(err, ErrInvalidActivitySelector) {
		t.Fatalf("empty selector error = %v", err)
	}
}

func errorsIs(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		next, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = next.Unwrap()
	}
	return false
}
