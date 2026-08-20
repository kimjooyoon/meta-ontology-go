package semantic

import (
	"testing"
)

func TestNormalizationIsOrderIndependentAndIdempotent(t *testing.T) {
	billing := Namespace("billing")
	activityID := MustIdentity("billing://activity/pay-order")
	orderID := MustIdentity("billing://entity/order")
	paymentID := MustIdentity("billing://entity/payment")

	first := NewGraph()
	for _, node := range []Node{
		{ID: paymentID, Kind: Entity, Namespace: billing, Name: " Payment ", Aliases: []string{"Receipt", "Receipt", " Payment "}},
		{ID: activityID, Kind: Activity, Namespace: billing, Name: " PayOrder "},
		{ID: orderID, Kind: Entity, Namespace: billing, Name: "Order"},
	} {
		if err := first.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if err := first.AddFact(NewWasGeneratedByFact(paymentID, activityID)); err != nil {
		t.Fatal(err)
	}
	if err := first.AddFact(NewUsedFact(activityID, orderID)); err != nil {
		t.Fatal(err)
	}

	second := NewGraph()
	for _, node := range []Node{
		{ID: orderID, Kind: Entity, Namespace: billing, Name: "Order"},
		{ID: activityID, Kind: Activity, Namespace: billing, Name: "PayOrder"},
		{ID: paymentID, Kind: Entity, Namespace: billing, Name: "Payment", Aliases: []string{"Receipt"}},
	} {
		if err := second.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if err := second.AddFact(NewUsedFact(activityID, orderID)); err != nil {
		t.Fatal(err)
	}
	if err := second.AddFact(NewWasGeneratedByFact(paymentID, activityID)); err != nil {
		t.Fatal(err)
	}

	if first.StableHash() != second.StableHash() {
		t.Fatalf("stable hash changed with insertion order or display metadata:\n%s\n%s", first.StableHash(), second.StableHash())
	}
	normalized, err := first.Normalized()
	if err != nil {
		t.Fatal(err)
	}
	normalizedAgain, err := normalized.Normalized()
	if err != nil {
		t.Fatal(err)
	}
	if normalized.Canonical() != normalizedAgain.Canonical() {
		t.Fatal("normalization is not idempotent")
	}
	if got := normalized.Nodes()[2].Aliases; len(got) != 1 || got[0] != "Receipt" {
		t.Fatalf("aliases were not normalized: %#v", got)
	}
}
