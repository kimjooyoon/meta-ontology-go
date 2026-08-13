package semantic

import "testing"

func TestCanonicalHashExcludesPresentationAndEvidence(t *testing.T) {
	ns := Namespace("billing")
	activityID := MustIdentity("billing://activity/pay")
	entityID := MustIdentity("billing://entity/order")

	left := NewGraph()
	leftNode := Node{
		ID: activityID, Kind: Activity, Namespace: ns, Name: "PayOrder",
		Aliases: []string{"Pay"}, Span: Span{File: "left.gooo", Start: Position{Offset: 1}, End: Position{Offset: 8}},
	}
	right := NewGraph()
	rightNode := Node{
		ID: activityID, Kind: Activity, Namespace: ns, Name: "CollectOrder",
		Aliases: []string{"Collect"}, Span: Span{File: "right.gooo", Start: Position{Offset: 20}, End: Position{Offset: 30}},
	}
	if err := left.AddNode(leftNode); err != nil {
		t.Fatal(err)
	}
	if err := right.AddNode(rightNode); err != nil {
		t.Fatal(err)
	}
	for _, graph := range []*Graph{&left, &right} {
		if err := graph.AddNode(Node{ID: entityID, Kind: Entity, Namespace: ns, Name: "Order"}); err != nil {
			t.Fatal(err)
		}
		fact := NewCandidateFact(activityID, Used, entityID, "  inferred from an implementation call  ").WithSpan(Span{
			File: "evidence.go", Start: Position{Offset: 40}, End: Position{Offset: 50},
		})
		if err := graph.AddCandidate(fact); err != nil {
			t.Fatal(err)
		}
	}

	if left.StableHash() != right.StableHash() {
		t.Fatalf("semantic hash changed with presentation/evidence metadata: %s != %s", left.StableHash(), right.StableHash())
	}
	if left.Canonical() == right.Canonical() {
		t.Fatal("full canonical form discarded presentation metadata")
	}
	if left.Hash() != left.StableHash() {
		t.Fatal("Hash is not the stable semantic hash")
	}
}

func TestGraphSnapshotsDoNotExposeMutableAliases(t *testing.T) {
	id := MustIdentity("billing://entity/order")
	g := NewGraph()
	if err := g.AddNode(Node{ID: id, Kind: Entity, Namespace: Namespace("billing"), Name: "Order", Aliases: []string{"Purchase"}}); err != nil {
		t.Fatal(err)
	}
	node, ok := g.Node(id)
	if !ok {
		t.Fatal("node was not stored")
	}
	node.Aliases[0] = "Mutated"
	if _, ok := g.NodeByName(Namespace("billing"), "Purchase"); !ok {
		t.Fatal("mutating a node snapshot changed graph lookup state")
	}
}

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
