package generator

import (
	"bytes"
	"strings"
	"testing"
)

func billingIR() SemanticIR {
	return SemanticIR{
		Package: "billinggen",
		Entities: []Entity{
			{ID: "billing://entity/order", Name: "Order", GoName: "Order", Fields: []Field{{Name: "ID", GoName: "ID", GoType: "string"}}},
			{ID: "billing://entity/payment", Name: "Payment", GoName: "Payment", Fields: []Field{{Name: "ID", GoName: "ID", GoType: "string"}}},
		},
		Activities: []Activity{{
			ID: "billing://activity/pay-order", Name: "PayOrder", GoName: "PayOrder",
			Inputs:  []Port{{ID: "billing://entity/order", Name: "order", GoName: "order", EntityID: "billing://entity/order"}},
			Outputs: []Port{{ID: "billing://entity/payment", Name: "payment", GoName: "payment", EntityID: "billing://entity/payment"}},
		}},
	}
}

func TestGenerateIsDeterministicAndFormatted(t *testing.T) {
	a, err := Generate(billingIR(), nil)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Generate(billingIR(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a.Source, b.Source) || !strings.Contains(string(a.Source), `//gooo:generated:start id="billing://activity/pay-order"`) {
		t.Fatalf("unstable or incomplete output:\n%s", a.Source)
	}
	if len(a.SourceMap.Mappings) < 3 {
		t.Fatalf("expected entity/activity/slot source mappings: %#v", a.SourceMap)
	}
}

func TestGeneratePreservesHandwrittenSlot(t *testing.T) {
	first, err := Generate(billingIR(), nil)
	if err != nil {
		t.Fatal(err)
	}
	previous := strings.Replace(string(first.Source), "return Payment{}", "return Payment{ID: order.ID}", 1)
	second, err := Generate(billingIR(), []byte(previous))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(second.Source), "return Payment{ID: order.ID}") {
		t.Fatalf("handwritten slot was not preserved:\n%s", second.Source)
	}
}
