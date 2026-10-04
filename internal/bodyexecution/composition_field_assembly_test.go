package bodyexecution

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestRecordFieldAssemblyIsCompiledAndDeliveredInNativeComposition(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/record-field-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/record-field-assembly-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateComposition(context.Background(), "fields.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	if prior.Steps[0].Generation.Report.RecordAssembly.FieldsPassed != 15 {
		t.Fatal("field selection missing")
	}
	for range 2 {
		run, err := ExecuteComposition(context.Background(), "fields.gooo", source, prior, suite, nativeTool())
		if err != nil || run.FinitePassed != 14 || run.FiniteTotal != 14 || !run.RuntimeReplayed || run.ModelCalls != 0 {
			t.Fatalf("record field execution: %v %+v", err, run)
		}
		for _, trace := range run.Traces {
			if string(trace.Deliveries[0].Actual) != string(trace.Deliveries[1].Input) || len(trace.Deliveries[0].ActualFields) != 3 || len(trace.Deliveries[1].InputFields) != 3 {
				t.Fatal("actual record producer delivery differs")
			}
		}
	}
	bad := []byte(strings.Replace(string(source), "alternative \"input0.title\"", "alternative \"input1\"", 1))
	result, err := GenerateComposition(context.Background(), "bad.gooo", bad, suite, "missing-model.json")
	if err == nil || len(result.Steps) != 0 || result.Model != nil {
		t.Fatal("field mismatch reached model construction", err)
	}
}
