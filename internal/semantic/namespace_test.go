package semantic

import "testing"

func TestNamespaceQualifiedNamesDoNotCollide(t *testing.T) {
	g := NewGraph()
	billing := Namespace("billing")
	settlement := Namespace("settlement")
	orderID := MustIdentity("billing://entity/order")
	settlementOrderID := MustIdentity("settlement://entity/order")
	for _, node := range []Node{
		mustEntity(t, orderID, billing, "Order"),
		mustEntity(t, settlementOrderID, settlement, "Order"),
	} {
		if err := g.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if _, ok := g.NodeByName(billing, "Order"); !ok {
		t.Fatal("billing::Order did not resolve")
	}
	if _, ok := g.NodeByName(settlement, "Order"); !ok {
		t.Fatal("settlement::Order did not resolve")
	}
	if _, ok := g.NodeByName(billing, "settlement::Order"); ok {
		t.Fatal("an unqualified namespace lookup crossed a namespace boundary")
	}
}

func TestSameNamespaceNamesCollideButExplicitCrossNamespaceFactsAreAllowed(t *testing.T) {
	g := NewGraph()
	billing := Namespace("billing")
	activity := mustActivity(t, MustIdentity("billing://activity/pay"), billing, "PayOrder")
	order := mustEntity(t, MustIdentity("billing://entity/order"), billing, "Order")
	fraud := mustEntity(t, MustIdentity("fraud://entity/check"), Namespace("fraud"), "Check")
	for _, node := range []Node{activity, order, fraud} {
		if err := g.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	duplicate, err := NewEntity(MustIdentity("billing://entity/other"), billing, "Order")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AddNode(duplicate); err == nil {
		t.Fatal("same-namespace duplicate name was accepted")
	}
	if err := g.AddFact(NewUsedFact(activity.ID, fraud.ID)); err != nil {
		t.Fatal(err)
	}
	if err := g.Validate(); err != nil {
		t.Fatalf("explicit cross-namespace relation should validate: %v", err)
	}
}
