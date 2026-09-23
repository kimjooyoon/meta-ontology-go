package generator

import (
	"strings"
	"testing"
)

func TestGenerateEmitsRuntimeBindingExecutionBoundary(t *testing.T) {
	ir := billingIR()
	ir.Activities = append(ir.Activities, Activity{
		ID:   "billing://activity/record-payment",
		Name: "RecordPayment", GoName: "RecordPayment",
		Inputs:  []Port{{ID: "billing://entity/payment", Name: "payment", EntityID: "billing://entity/payment"}},
		Outputs: []Port{{ID: "billing://entity/payment", Name: "result", EntityID: "billing://entity/payment"}},
		Slots:   []Slot{{ID: "billing://slot/record-payment", Default: "return payment"}},
	})
	ir.RuntimeBindings = []RuntimeBinding{{
		Schema:           "gooo.runtime-binding/v1",
		ProducerActivity: "billing://activity/pay-order",
		ProducerPort:     "result",
		ConsumerActivity: "billing://activity/record-payment",
		ConsumerPort:     "input",
		Entity:           "billing://entity/payment",
	}}

	result, err := Generate(ir, nil)
	if err != nil {
		t.Fatal(err)
	}
	source := string(result.Source)
	for _, want := range []string{
		`goooRuntimeReflect "reflect"`,
		"type GoooRuntimeBinding struct",
		"func ExecuteGoooRuntimeBindings",
		"RUNTIME_EDGE_UNREACHABLE",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
