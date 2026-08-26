package semanticdelta

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func semanticContractFixture(t *testing.T, activityName string, withReceipt bool) semantic.IR {
	t.Helper()
	ir := semantic.NewIR("semantic-delta-contract", semantic.Namespace("billing"))
	activityID := semantic.MustIdentity("billing://activity/pay-order")
	orderID := semantic.MustIdentity("billing://entity/order")
	activity, err := semantic.NewActivity(activityID, ir.Namespace, activityName)
	if err != nil {
		t.Fatal(err)
	}
	order, err := semantic.NewEntity(orderID, ir.Namespace, "Order")
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range []semantic.Node{activity, order} {
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if err := ir.AddFact(semantic.NewUsedFact(activityID, orderID)); err != nil {
		t.Fatal(err)
	}
	if !withReceipt {
		return ir
	}
	receiptID := semantic.MustIdentity("billing://entity/receipt")
	receipt, err := semantic.NewEntity(receiptID, ir.Namespace, "Receipt")
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(receipt); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddFact(semantic.NewUsedFact(activityID, receiptID)); err != nil {
		t.Fatal(err)
	}
	return ir
}
