package semantic

import (
	"testing"
)

func TestDeterministicFactsShadowCandidates(t *testing.T) {
	g := NewGraph()
	activity := MustIdentity("billing://activity/pay")
	entity := MustIdentity("billing://entity/order")
	if err := g.AddNode(mustActivity(t, activity, Namespace("billing"), "Pay")); err != nil {
		t.Fatal(err)
	}
	if err := g.AddNode(mustEntity(t, entity, Namespace("billing"), "Order")); err != nil {
		t.Fatal(err)
	}
	candidate := NewCandidateFact(activity, Used, entity, "observed registered semantic symbol")
	if err := g.AddCandidate(candidate); err != nil {
		t.Fatal(err)
	}
	if len(g.Candidates()) != 1 || len(g.Facts()) != 0 {
		t.Fatal("candidate was not kept separate from deterministic facts")
	}
	if err := g.AddFact(NewUsedFact(activity, entity)); err != nil {
		t.Fatal(err)
	}
	if len(g.Candidates()) != 0 || len(g.Facts()) != 1 {
		t.Fatal("deterministic fact did not shadow candidate")
	}
}
func TestActivityContractDerivesOnlyPROVCoreFacts(t *testing.T) {
	g := NewGraph()
	ns := Namespace("billing")
	activity := MustIdentity("billing://activity/pay-order")
	order := MustIdentity("billing://entity/order")
	method := MustIdentity("billing://entity/payment-method")
	payment := MustIdentity("billing://entity/payment")
	for _, node := range []Node{
		mustActivity(t, activity, ns, "Pay order"),
		mustEntity(t, order, ns, "Order"),
		mustEntity(t, method, ns, "Payment method"),
		mustEntity(t, payment, ns, "Payment"),
	} {
		if err := g.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.AddActivityContract(ActivityContract{
		Activity: activity,
		Inputs:   []ID{order, method},
		Outputs:  []ID{payment},
	}); err != nil {
		t.Fatal(err)
	}
	facts := g.Facts()
	if len(facts) != 3 {
		t.Fatalf("derived fact count = %d, want 3", len(facts))
	}
	if facts[0].Status != FactDeterministic || facts[1].Status != FactDeterministic || facts[2].Status != FactDeterministic {
		t.Fatal("activity contract produced a non-deterministic core fact")
	}
}
