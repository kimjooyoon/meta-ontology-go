package bodyexecution

import (
	"context"
	"os"
	"testing"
)

func TestSequentialRecordUpdatesExecuteAndDeliverNativeValues(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/record-field-updates.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/record-field-updates-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateComposition(context.Background(), "updates.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	if prior.Steps[0].Generation.Report.RecordAssembly.FieldsPassed != 15 {
		t.Fatal("selection")
	}
	run, err := ExecuteComposition(context.Background(), "updates.gooo", source, prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 14 || run.FiniteTotal != 14 || !run.RuntimeReplayed || run.ModelCalls != 0 {
		t.Fatal("native sequential updates", err)
	}
	for _, trace := range run.Traces {
		if string(trace.Deliveries[0].Actual) != string(trace.Deliveries[1].Input) {
			t.Fatal("record delivery")
		}
	}
}
