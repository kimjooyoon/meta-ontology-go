package semantic

import (
	"testing"
)

func TestPromoteCandidateIsAtomicAndPreservesFact(t *testing.T) {
	activity := MustIdentity("billing://activity/pay")
	entity := MustIdentity("billing://entity/order")
	g := NewGraph()
	if err := g.AddNode(mustActivity(t, activity, Namespace("billing"), "Pay")); err != nil {
		t.Fatal(err)
	}
	if err := g.AddNode(mustEntity(t, entity, Namespace("billing"), "Order")); err != nil {
		t.Fatal(err)
	}
	candidate := NewCandidateFact(activity, Used, entity, "implementation evidence")
	if err := g.AddCandidate(candidate); err != nil {
		t.Fatal(err)
	}
	promoted, err := g.PromoteCandidate(candidate.Key())
	if err != nil {
		t.Fatal(err)
	}
	if promoted.Status != FactDeterministic || len(g.Candidates()) != 0 || len(g.Facts()) != 1 {
		t.Fatalf("candidate promotion state = %#v, facts=%d, candidates=%d", promoted, len(g.Facts()), len(g.Candidates()))
	}
	if promoted.Reason != candidate.Reason {
		t.Fatalf("promotion lost reason: %q", promoted.Reason)
	}
}
func TestSpanRejectsReversedOffsetsEvenWithoutFilename(t *testing.T) {
	span := Span{Start: Position{Offset: 8}, End: Position{Offset: 3}}
	if err := span.Validate(); err == nil {
		t.Fatal("reversed source span was accepted")
	}
}
