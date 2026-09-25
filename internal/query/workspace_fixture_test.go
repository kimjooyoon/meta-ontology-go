package query

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func workspaceIR(t *testing.T, namespace, prefix string, candidate bool) semantic.IR {
	t.Helper()
	ir := semantic.NewIR(namespace, semantic.Namespace(namespace))
	entities := []struct{ id, name string }{
		{prefix + "entity/payment", "Payment"}, {prefix + "entity/order", "Order"},
		{prefix + "entity/base", "Base"},
	}
	for _, entity := range entities {
		node, err := semantic.NewEntity(semantic.MustIdentity(entity.id), semantic.Namespace(namespace), entity.name)
		if err != nil {
			t.Fatal(err)
		}
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	activity, err := semantic.NewActivity(semantic.MustIdentity(prefix+"activity/pay"), semantic.Namespace(namespace), "Pay")
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(activity); err != nil {
		t.Fatal(err)
	}
	add := func(fact semantic.Fact) {
		t.Helper()
		if err := ir.AddFact(fact); err != nil {
			t.Fatal(err)
		}
	}
	add(semantic.NewUsedFact(activity.ID, semantic.MustIdentity(prefix+"entity/order")))
	add(semantic.NewWasDerivedFromFact(semantic.MustIdentity(prefix+"entity/payment"), semantic.MustIdentity(prefix+"entity/order")))
	add(semantic.NewWasDerivedFromFact(semantic.MustIdentity(prefix+"entity/order"), semantic.MustIdentity(prefix+"entity/base")))
	if !candidate {
		return ir
	}
	external, err := semantic.NewEntity(semantic.MustIdentity(prefix+"entity/external"), semantic.Namespace(namespace), "External")
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(external); err != nil {
		t.Fatal(err)
	}
	observation := semantic.NewCandidateFact(
		semantic.MustIdentity(prefix+"entity/payment"), semantic.WasDerivedFrom,
		semantic.MustIdentity(prefix+"entity/external"), "unresolved cross-context observation",
	)
	if err := ir.AddCandidate(observation); err != nil {
		t.Fatal(err)
	}
	return ir
}
