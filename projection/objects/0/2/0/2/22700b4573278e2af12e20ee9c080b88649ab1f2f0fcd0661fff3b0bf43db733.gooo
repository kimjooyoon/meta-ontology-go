package semantic

import (
	"testing"
)

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
