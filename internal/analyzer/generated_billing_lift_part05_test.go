package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func declaredBillingContract(t *testing.T) semantic.IR {
	t.Helper()
	ir := semantic.NewIR("billing", semantic.Namespace("billing"))
	activity := mustBillingNode(t, semantic.Activity, "billing://activity/pay-order", "PayOrder")
	order := mustBillingNode(t, semantic.Entity, "billing://entity/order", "Order")
	method := mustBillingNode(t, semantic.Entity, "billing://entity/payment-method", "PaymentMethod")
	payment := mustBillingNode(t, semantic.Entity, "billing://entity/payment", "Payment")
	for _, node := range []semantic.Node{activity, order, method, payment} {
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if err := ir.AddActivityContract(semantic.ActivityContract{
		Activity: activity.ID, Inputs: []semantic.ID{order.ID, method.ID}, Outputs: []semantic.ID{payment.ID},
	}); err != nil {
		t.Fatal(err)
	}
	return ir
}
func mustBillingNode(t *testing.T, kind semantic.Kind, id, name string) semantic.Node {
	t.Helper()
	node, err := semantic.NewNode(kind, semantic.MustIdentity(id), semantic.Namespace("billing"), name)
	if err != nil {
		t.Fatal(err)
	}
	return node
}
func billingRef(name string) SymbolRef {
	return SymbolRef{PackagePath: "billing", PackageName: "billing", Name: name}
}
